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
	"aura-gateway/internal/metrics"
	"aura-gateway/internal/tools"

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
	registry            *tools.Registry               // Registry for user-provisioned tools
	connectorRegistry   *connectors.ConnectorRegistry // Core/Native connectors
	toolThreshold       float64
	cacheTTL            time.Duration
	rateLimiters        sync.Map
	auditLogger         *metrics.PersistentAuditLogger

	TotalRequests atomic.Uint64
	CacheHits     atomic.Uint64
	orgRequests   sync.Map // Tracks requests per org (map[string]*atomic.Uint64)
	orgHits       sync.Map // Tracks cache hits per org (map[string]*atomic.Uint64)
}

func NewAuraEngine(
	cache *cch.AuraCache,
	rtrClient *rtr.RouterClient,
	rbacClient *auth.RBACClient,
	mcp *mc.MCPClient,
	reg *tools.Registry,
	connReg *connectors.ConnectorRegistry,
	providers map[string]lp.Provider,
	defaultProvider string,
	threshold float64,
	ttl time.Duration,
	audit *metrics.PersistentAuditLogger,
) *AuraEngine {
	return &AuraEngine{
		cache:             cache,
		router:            rtrClient,
		rbac:              rbacClient,
		mcp:               mcp,
		registry:          reg,
		connectorRegistry: connReg,
		providers:         providers,
		defaultProvider:   defaultProvider,
		toolThreshold:     threshold,
		cacheTTL:          ttl,
		auditLogger:       audit,
	}
}

func (e *AuraEngine) ActiveProviderNames() []string {
	providers := make([]string, 0, len(e.providers))
	for _, provider := range e.providers {
		providers = append(providers, provider.GetProviderName())
	}
	return providers
}

func (e *AuraEngine) GetStats(orgID string) (uint64, uint64) {
	var reqs, hits uint64
	if counter, ok := e.orgRequests.Load(orgID); ok {
		reqs = counter.(*atomic.Uint64).Load()
	}
	if counter, ok := e.orgHits.Load(orgID); ok {
		hits = counter.(*atomic.Uint64).Load()
	}
	return reqs, hits
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

	// A. Rate Limiting (Based on Tier extracted from JWT)
	tier, _ := ctx.Value("tier").(string)
	orgID, _ := ctx.Value("org_id").(string)
	if orgID == "" {
		orgID = "default"
	}

	// Track per-org requests
	reqCounter, _ := e.orgRequests.LoadOrStore(orgID, &atomic.Uint64{})
	reqCounter.(*atomic.Uint64).Add(1)

	// Dynamic Rate Limiting Based on Tier
	limit := 10.0 // Free default
	if tier == "pro" {
		limit = 100.0
	} else if tier == "business" {
		limit = 1000.0
	}

	limitKey := fmt.Sprintf("rl:%s:%s", orgID, req.UserID)
	limiter, _ := e.rateLimiters.LoadOrStore(limitKey, rate.NewLimiter(rate.Limit(limit/60), int(limit)))
	if !limiter.(*rate.Limiter).Allow() {
		return nil, fmt.Errorf("rate limit exceeded for organization %s (tier: %s)", orgID, tier)
	}

	// B. Stage 1-2 Cache Lookup (Org-Isolated)
	if e.cache != nil && !req.SkipCache {
		cacheKey := fmt.Sprintf("org:%s:p:%s", orgID, req.Prompt)
		cachedResp, err := e.cache.GetSemanticResult(ctx, cacheKey)
		if err == nil && cachedResp != "" {
			e.CacheHits.Add(1)
			hitCounter, _ := e.orgHits.LoadOrStore(orgID, &atomic.Uint64{})
			hitCounter.(*atomic.Uint64).Add(1)
			if e.auditLogger != nil {
				e.auditLogger.Log(ctx, metrics.AuditEvent{
					Timestamp: time.Now(),
					OrgID:     orgID,
					Type:      "CACHE",
					User:      req.UserID,
					Detail:    "Stage 1 HIT: " + req.Prompt,
					Status:    "success",
					Latency:   0,
				}, map[string]interface{}{"prompt": req.Prompt, "stage": 1})
			}
			return &PromptResponse{Text: cachedResp, Cached: true}, nil
		}

		// Stage 1.5: Canonical Match (Normalized & Org-Isolated)
		_, cHash := NormalizePrompt(req.Prompt)
		canonKey := fmt.Sprintf("org:%s:c:%s", orgID, cHash)
		cachedCanon, err := e.cache.GetSemanticResult(ctx, canonKey)
		if err == nil && cachedCanon != "" {
			e.CacheHits.Add(1)
			hitCounter, _ := e.orgHits.LoadOrStore(orgID, &atomic.Uint64{})
			hitCounter.(*atomic.Uint64).Add(1)
			slog.Info("🎯 Stage 1.5 Cache HIT (Org-Isolated)", "org_id", orgID, "canonical", cHash)
			_ = e.cache.SetResult(ctx, cacheKey, cachedCanon, e.cacheTTL)
			return &PromptResponse{Text: cachedCanon, Cached: true}, nil
		}
	}


	// C. RBAC Check (Organization Scoped)
	var allowedTools []string
	if e.rbac != nil {
		// Use orgID from context for permission checks
		allowed, err := e.rbac.CheckPermission(ctx, orgID, "chat:execute")
		if err != nil {
			slog.Error("RBAC check failed", "error", err, "org_id", orgID)
		}
		if !allowed {
			slog.Warn("Unauthorized engine access attempted", "org_id", orgID, "user_id", req.UserID)
			return nil, fmt.Errorf("unauthorized: insufficient scope")
		}
		// Get tools specifically allowed for this organization
		allowedTools, _ = e.rbac.GetAllowedTools(orgID)
	}

	// D. Semantic Routing (includes Vector Search & Prompt Compression via Rust)
	tools, compressedPrompt, similarPromptHash, currentPromptHash, err := e.router.GetBestTools(ctx, req.Prompt, orgID, allowedTools)
	if err != nil {
		slog.Warn("Router fallback engaged", "error", err)
	}

	// NEW: Stage 2 Cache Check (Fuzzy Vector Semantic Match) — Org-Isolated
	if similarPromptHash != "" && e.cache != nil && !req.SkipCache {
		simKey := fmt.Sprintf("org:%s:s:%s", orgID, similarPromptHash)
		cachedResp, err := e.cache.GetSemanticResult(ctx, simKey)
		if err == nil && cachedResp != "" {
			e.CacheHits.Add(1)
			hitCounter, _ := e.orgHits.LoadOrStore(orgID, &atomic.Uint64{})
			hitCounter.(*atomic.Uint64).Add(1)
			slog.Info("🎯 Stage 2 Cache HIT (Org-Isolated)", "org_id", orgID, "similar_hash", similarPromptHash)
			return &PromptResponse{Text: cachedResp, Cached: true}, nil
		}
	}

	// E. Tool Execution (Multi-Connector: Universal Provisioning)
	var toolResults []string
	if len(tools) > 0 {
		for _, t := range tools {
			if t.RelevanceScore > float32(e.toolThreshold) {
				slog.Info("Resolving tool for execution", "tool_id", t.Id, "score", t.RelevanceScore)

				// 1. Fetch full metadata from Registry (to get OrgID and Config)
				toolMetadata, err := e.registry.GetTool(ctx, t.Id)
				if err != nil {
					slog.Error("Failed to fetch tool metadata", "tool_id", t.Id, "error", err)
					continue
				}

				if toolMetadata == nil {
					slog.Warn("Tool not found in registry", "tool_id", t.Id)
					continue
				}

				var connector connectors.Connector
				
				// 2. Connector Selection & Factory
				switch toolMetadata.ConnectorType {
				case "rest":
					connector = connectors.NewRESTConnector(toolMetadata.Endpoint)
				case "sql":
					connector = connectors.NewSQLConnector(toolMetadata.Endpoint)
				case "mcp":
					if reg, ok := e.connectorRegistry.Get(connectors.TypeMCP); ok {
						connector = reg
					}
				default:
					// Try Core Registry
					if reg, ok := e.connectorRegistry.Get(connectors.TypeCore); ok {
						if cc, ok := reg.(*connectors.CoreConnector); ok && cc.HasTool(t.Id) {
							connector = reg
						}
					}
				}

				if connector == nil {
					slog.Warn("No connector available for tool", "tool_id", t.Id, "type", toolMetadata.ConnectorType)
					continue
				}

				// 3. Execution Phase
				execReq := &connectors.ExecutionRequest{
					ToolID:  t.Id,
					UserID:  req.UserID,
					Inputs:  make(map[string]interface{}), // Future: LLM intent injection
					Timeout: toolMetadata.TimeoutSeconds,
				}

				toolCtx, cancel := context.WithTimeout(ctx, time.Duration(toolMetadata.TimeoutSeconds+1)*time.Second)
				execResp, err := connector.Execute(toolCtx, execReq)
				cancel()

				if err != nil {
					slog.Error("Tool execution failed", "tool_id", t.Id, "error", err)
					continue
				}

				if execResp.Status == "success" && execResp.Data != nil {
					toolResults = append(toolResults, fmt.Sprintf("%v", execResp.Data))
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

	// G. Populate Cache for future requests (Org-Isolated & Force Refresh)
	if e.cache != nil {
		// Layer 1: Literal Match
		cacheKey := fmt.Sprintf("org:%s:p:%s", orgID, req.Prompt)
		_ = e.cache.SetResult(ctx, cacheKey, aiResp, e.cacheTTL)

		// Layer 2: Canonical Match
		_, cHash := NormalizePrompt(req.Prompt)
		canonKey := fmt.Sprintf("org:%s:c:%s", orgID, cHash)
		_ = e.cache.SetResult(ctx, canonKey, aiResp, e.cacheTTL)

		// Layer 3: Semantic Match
		if currentPromptHash != "" {
			simKey := fmt.Sprintf("org:%s:s:%s", orgID, currentPromptHash)
			_ = e.cache.SetResult(ctx, simKey, aiResp, e.cacheTTL)
		}
	}

	// H. Log Final Result
	if e.auditLogger != nil {
		e.auditLogger.Log(ctx, metrics.AuditEvent{
			Timestamp: time.Now(),
			OrgID:     orgID,
			Type:      "GENERATION",
			User:      req.UserID,
			Detail:    fmt.Sprintf("Provider: %s", selectedProvider.GetProviderName()),
			Status:    "success",
		}, map[string]interface{}{"provider": selectedProvider.GetProviderName(), "tools_count": len(llmTools)})
	}

	return &PromptResponse{
		Text:     aiResp,
		Cached:   false,
		Provider: selectedProvider.GetProviderName(),
		Tools:    llmTools,
	}, nil
}
