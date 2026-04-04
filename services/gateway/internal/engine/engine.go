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
	cache         *cch.AuraCache
	router        *rtr.RouterClient
	rbac          *auth.RBACClient
	providers     map[string]lp.Provider // keyed by provider name e.g. "ollama"
	defaultProvider string               // key used when no X-Aura-Provider header is set
	mcp           *mc.MCPClient
	toolThreshold float64
	cacheTTL      time.Duration
	rateLimiters  sync.Map

	TotalRequests atomic.Uint64
	CacheHits     atomic.Uint64
}

// NewAuraEngine initializes the engine with its required dependencies.
// providers is a map of name->Provider; defaultProvider is the key used when no provider is specified.
func NewAuraEngine(c *cch.AuraCache, r *rtr.RouterClient, auth *auth.RBACClient, providers map[string]lp.Provider, defaultProvider string, m *mc.MCPClient, threshold float64, ttl time.Duration) *AuraEngine {
	return &AuraEngine{
		cache:           c,
		router:          r,
		rbac:            auth,
		providers:       providers,
		defaultProvider: defaultProvider,
		mcp:             m,
		toolThreshold:   threshold,
		cacheTTL:        ttl,
	}
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

	// E. Tool Execution (MCP)
	var toolResults []string
	if e.mcp != nil && len(tools) > 0 {
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
		contextPrompt = fmt.Sprintf("User Prompt: %s\n\nExecuted Tool Data:\n%v", contextPrompt, toolResults)
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

	// G. Populate Cache for future requests (skipped if SkipCache was requested)
	if e.cache != nil && !req.SkipCache {
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
