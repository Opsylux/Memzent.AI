package engine

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"aura-gateway/internal/auth"
	cch "aura-gateway/internal/cache"
	"aura-gateway/internal/connectors"
	lp "aura-gateway/internal/llm"
	mc "aura-gateway/internal/mcp"
	rtr "aura-gateway/internal/router"

	"golang.org/x/time/rate"
)

// PromptRequest defines the incoming user payload
type PromptRequest struct {
	UserID    string `json:"user_id"`
	Prompt    string `json:"prompt"`
	Provider  string `json:"provider,omitempty"`  // e.g. "ollama", "openai", "anthropic", "gemini"
	Model     string `json:"model,omitempty"`     // optional per-request model override
	SkipCache bool   `json:"skip_cache,omitempty"` // set by X-Skip-Cache header
}

// PromptResponse defines the gateway's response to the client
type PromptResponse struct {
	Text      string `json:"text"`
	Cached    bool   `json:"cached"`
	Provider  string `json:"provider,omitempty"`
	Tools     []any  `json:"tools,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// AuraEngine orchestrates the flow between Cache, RBAC, Router, MCP, and LLM
type AuraEngine struct {
	cache               *cch.AuraCache
	router              *rtr.RouterClient
	rbac                *auth.RBACClient
	providers           map[string]lp.Provider // keyed by provider name e.g. "ollama"
	defaultProvider     string                 // key used when no X-Aura-Provider header is set
	mcp                 *mc.MCPClient
	connectorRegistry   *connectors.ConnectorRegistry // Phase 3: Multi-connector framework
	toolThreshold       float64
	cacheTTL            time.Duration
	rateLimiters        sync.Map

	TotalRequests atomic.Uint64
	CacheHits     atomic.Uint64
}

// NewAuraEngine initializes the engine with its required dependencies.
// providers is a map of name->Provider; defaultProvider is the key used when no provider is specified.
func NewAuraEngine(c *cch.AuraCache, r *rtr.RouterClient, auth *auth.RBACClient, providers map[string]lp.Provider, defaultProvider string, m *mc.MCPClient, connReg *connectors.ConnectorRegistry, threshold float64, ttl time.Duration) *AuraEngine {
	return &AuraEngine{
		cache:             c,
		router:            r,
		rbac:              auth,
		providers:         providers,
		defaultProvider:   defaultProvider,
		mcp:               m,
		connectorRegistry: connReg,
		toolThreshold:     threshold,
		cacheTTL:          ttl,
	}
}

func (e *AuraEngine) ActiveProviderNames() []string {
	providers := make([]string, 0, len(e.providers))
	for _, provider := range e.providers {
		providers = append(providers, provider.GetProviderName())
	}
	return providers
}

func (e *AuraEngine) DefaultProviderName() string {
	if p, ok := e.providers[e.defaultProvider]; ok {
		return p.GetProviderName()
	}
	return "unknown"
}

func (e *AuraEngine) ProviderCount() int {
	return len(e.providers)
}

func (e *AuraEngine) Process(ctx context.Context, req *PromptRequest) (*PromptResponse, error) {
	e.TotalRequests.Add(1)

	// A. Rate Limiting (10 requests per minute per user)
	limiter, _ := e.rateLimiters.LoadOrStore(req.UserID, rate.NewLimiter(rate.Every(time.Minute/10), 10))
	if !limiter.(*rate.Limiter).Allow() {
		return nil, fmt.Errorf("rate limit exceeded: please wait a moment")
	}

	// B. Stage 1-2 Cache Lookup — skipped if client requests a fresh response
	if e.cache != nil && !req.SkipCache {
		cachedResp, err := e.cache.GetSemanticResult(ctx, req.Prompt)
		if err == nil && cachedResp != "" {
			e.CacheHits.Add(1)
			slog.Info("🎯 Stage 1 Cache HIT (Literal)", "prompt", req.Prompt)
			return &PromptResponse{Text: cachedResp, Cached: true}, nil
		}

		// Stage 1.5: Canonical Match (Normalized)
		// Mask IDs, lowercase, and stabilize to catch write011 vs write111
		canonical, cHash := NormalizePrompt(req.Prompt)
		cachedCanon, err := e.cache.GetSemanticResult(ctx, cHash)
		if err == nil && cachedCanon != "" {
			e.CacheHits.Add(1)
			slog.Info("🎯 Stage 1.5 Cache HIT (Canonical)", "original", req.Prompt, "canonical", canonical)
			// Map the original literal string to this hit for faster Stage 1 next time
			_ = e.cache.SetResult(ctx, req.Prompt, cachedCanon, e.cacheTTL)
			return &PromptResponse{Text: cachedCanon, Cached: true}, nil
		}
	}


	// C. RBAC Check
	var allowedTools []string
	if e.rbac != nil {
		allowed, err := e.rbac.CheckPermission(ctx, req.UserID, "chat:execute")
		if err != nil {
			slog.Error("RBAC check failed", "error", err, "user_id", req.UserID)
		}
		if !allowed {
			return nil, fmt.Errorf("unauthorized: insufficient scope")
		}
		// Get tools specifically allowed for this user
		allowedTools, _ = e.rbac.GetAllowedTools(req.UserID)
	}

	// D. Semantic Routing (includes Vector Search & Prompt Compression via Rust)
	tools, compressedPrompt, similarPromptHash, currentPromptHash, err := e.router.GetBestTools(ctx, req.Prompt, req.UserID, allowedTools)
	if err != nil {
		slog.Warn("Router fallback engaged", "error", err)
	}

	// NEW: Stage 2 Cache Check (Fuzzy Vector Semantic Match) — also skipped on SkipCache
	if similarPromptHash != "" && e.cache != nil && !req.SkipCache {
		cachedResp, err := e.cache.GetSemanticResult(ctx, similarPromptHash)
		if err == nil && cachedResp != "" {
			slog.Info("🎯 Stage 2 Cache HIT (Vector)", "original", req.Prompt, "similar_hash", similarPromptHash)

			// Repopulate Literal (Stage 1) and Canonical (Stage 1.5) for next time
			_ = e.cache.SetResult(ctx, req.Prompt, cachedResp, e.cacheTTL)
			_, cHash := NormalizePrompt(req.Prompt)
			_ = e.cache.SetResult(ctx, cHash, cachedResp, e.cacheTTL)

			return &PromptResponse{Text: cachedResp, Cached: true}, nil
		}
	}

	// E. Tool Execution (Multi-Connector: Phase 3)
	// Supports MCP, REST, SQL, GraphQL, gRPC, Webhooks (connectors.ConnectorRegistry)
	var toolResults []string
	if e.connectorRegistry != nil && len(tools) > 0 {
		for _, t := range tools {
			if t.RelevanceScore > float32(e.toolThreshold) {
				slog.Info("Executing tool via connector", "tool_id", t.Id, "score", t.RelevanceScore)

				// HYBRID STRATEGY: 
				// 1. Try Core (Internal) first for zero-latency
				// 2. Fall back to MCP for external/registry tools
				var connector connectors.Connector
				if core, ok := e.connectorRegistry.Get(connectors.TypeCore); ok {
					// Check if this specific tool is registered in Core
					if cc, ok := core.(*connectors.CoreConnector); ok && cc.HasTool(t.Id) {
						connector = core
					}
				}

				if connector == nil {
					// Default to MCP if not a core tool
					var ok bool
					connector, ok = e.connectorRegistry.Get(connectors.TypeMCP)
					if !ok {
						slog.Warn("No connector available for tool", "tool_id", t.Id)
						continue
					}
				}

				// Build execution request
				execReq := &connectors.ExecutionRequest{
					ToolID:  t.Id,
					UserID:  req.UserID,
					Inputs:  make(map[string]interface{}),
					Timeout: 15,
				}

				// Execute via connector
				toolCtx, cancel := context.WithTimeout(ctx, 16*time.Second)
				execResp, err := connector.Execute(toolCtx, execReq)
				cancel()

				if err != nil {
					slog.Error("Connector execution error", "tool_id", t.Id, "error", err)
					continue
				}

				if execResp.Status != "success" {
					slog.Warn("Tool execution failed", "tool_id", t.Id, "status", execResp.Status, "error", execResp.Error)
					continue
				}

				// Extract results from connector response
				if execResp.Data != nil {
					switch v := execResp.Data.(type) {
					case []string:
						toolResults = append(toolResults, v...)
					case string:
						toolResults = append(toolResults, v)
					case []interface{}:
						for _, item := range v {
							toolResults = append(toolResults, fmt.Sprintf("%v", item))
						}
					default:
						toolResults = append(toolResults, fmt.Sprintf("%v", v))
					}
				}
			}
		}
	} else if e.mcp != nil && len(tools) > 0 {
		// Fall back to direct MCP for backward compatibility (no connector registry)
		slog.Info("Using legacy MCP execution (connector registry not available)")
		for _, t := range tools {
			if t.RelevanceScore > float32(e.toolThreshold) {
				slog.Info("Executing tool", "tool_id", t.Id, "score", t.RelevanceScore)

				// Use a sub-context for the tool call to prevent hanging
				toolCtx, cancel := context.WithTimeout(ctx, 15*time.Second)

				type ToolArgs struct {
					ToolID string `json:"tool_id"`
					UserID string `json:"user_id,omitempty"`
				}

				args := ToolArgs{
					ToolID:  t.Id,
					UserID: req.UserID,
				}

				resp, err := e.mcp.CallTool(toolCtx, "execute_aura_tool", args)
				cancel() // Release context immediately

				if err != nil {
					slog.Error("MCP tool execution error", "tool_id", t.Id, "error", err)
					continue
				}

				// Extract text content safely
				if resp != nil {
					for _, content := range resp.Content {
						if content.TextContent != nil {
							toolResults = append(toolResults, content.TextContent.Text)
						}
					}
				}
			}
		}
	}


	// F. Build LLM context from compressed prompt + tool results
	// Use the compressed prompt from the Rust layer to save costs and latency.
	contextPrompt := compressedPrompt
	if contextPrompt == "" {
		contextPrompt = req.Prompt // Fallback
	}
	if len(toolResults) > 0 {
		contextPrompt = fmt.Sprintf("%s\n\n### SUPPLEMENTARY TOOL CONTEXT\n%v\n--- END TOOL CONTEXT ---", contextPrompt, toolResults)
	}

	// Mapping *router.Tool to any slice for the prompt response payload
	var llmTools []any
	for _, t := range tools {
		llmTools = append(llmTools, t)
	}

	// G. Provider Selection

	providerKey := req.Provider
	if providerKey == "" {
		providerKey = e.defaultProvider
	}
	selectedProvider, ok := e.providers[providerKey]
	if !ok {
		slog.Warn("Unknown provider requested, falling back to default", "requested", providerKey, "default", e.defaultProvider)
		selectedProvider = e.providers[e.defaultProvider]
	}

	slog.Info("🤖 LLM Provider selected", "provider", selectedProvider.GetProviderName(), "model_override", req.Model, "skip_cache", req.SkipCache)

	aiResp, err := selectedProvider.Generate(ctx, contextPrompt, llmTools, req.Model)
	if err != nil {
		slog.Error("LLM generation failed", "error", err, "provider", selectedProvider.GetProviderName())
		return nil, err
	}

	// G. Populate Cache for future requests (Force Refresh Pattern)
	if e.cache != nil {
		// Layer 1: Literal Match
		_ = e.cache.SetResult(ctx, req.Prompt, aiResp, e.cacheTTL)

		// Layer 2: Canonical Match (Masking IDs/Numbers)
		_, cHash := NormalizePrompt(req.Prompt)
		_ = e.cache.SetResult(ctx, cHash, aiResp, e.cacheTTL)

		// Layer 3: Semantic Match (Representative Hash from Router)
		if currentPromptHash != "" && currentPromptHash != cHash {
			_ = e.cache.SetResult(ctx, currentPromptHash, aiResp, e.cacheTTL)
		}
	}

	return &PromptResponse{
		Text:     aiResp,
		Cached:   false,
		Provider: selectedProvider.GetProviderName(),
		Tools:    llmTools,
	}, nil
}
