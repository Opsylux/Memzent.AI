package engine

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"memzent-gateway/internal/auth"
	"memzent-gateway/internal/billing"
	cch "memzent-gateway/internal/cache"
	"memzent-gateway/internal/connectors"
	lp "memzent-gateway/internal/llm"
	mc "memzent-gateway/internal/mcp"
	rtr "memzent-gateway/internal/router"
	"memzent-gateway/internal/metrics"
	"memzent-gateway/internal/tools"

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

// MemzentEngine orchestrates the flow between Cache, RBAC, Router, MCP, and LLM
type MemzentEngine struct {
	cache               *cch.MemzentCache
	router              *rtr.RouterClient
	rbac                *auth.RBACClient
	ledger              *billing.Ledger
	costCalc            *billing.CostCalculator
	providers           map[string]lp.Provider // keyed by provider name e.g. "ollama"
	defaultProvider     string                 // key used when no X-Memzent-Provider header is set
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

func NewMemzentEngine(
	cache *cch.MemzentCache,
	rtrClient *rtr.RouterClient,
	rbacClient *auth.RBACClient,
	ledger *billing.Ledger,
	costCalc *billing.CostCalculator,
	mcp *mc.MCPClient,
	reg *tools.Registry,
	connReg *connectors.ConnectorRegistry,
	providers map[string]lp.Provider,
	defaultProvider string,
	threshold float64,
	ttl time.Duration,
	audit *metrics.PersistentAuditLogger,
) *MemzentEngine {
	return &MemzentEngine{
		cache:             cache,
		router:            rtrClient,
		rbac:              rbacClient,
		ledger:            ledger,
		costCalc:          costCalc,
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

func (e *MemzentEngine) ActiveProviderNames() []string {
	providers := make([]string, 0, len(e.providers))
	for _, provider := range e.providers {
		providers = append(providers, provider.GetProviderName())
	}
	return providers
}

func (e *MemzentEngine) GetProviderMetadata() []lp.ProviderMetadata {
	metadata := make([]lp.ProviderMetadata, 0, len(e.providers))
	for _, provider := range e.providers {
		metadata = append(metadata, provider.GetMetadata())
	}
	return metadata
}

func (e *MemzentEngine) GetStats(orgID string) (uint64, uint64) {
	var reqs, hits uint64
	if counter, ok := e.orgRequests.Load(orgID); ok {
		reqs = counter.(*atomic.Uint64).Load()
	}
	if counter, ok := e.orgHits.Load(orgID); ok {
		hits = counter.(*atomic.Uint64).Load()
	}
	return reqs, hits
}

func (e *MemzentEngine) DefaultProviderName() string {
	if p, ok := e.providers[e.defaultProvider]; ok {
		return p.GetProviderName()
	}
	return "unknown"
}

func (e *MemzentEngine) ProviderCount() int {
	return len(e.providers)
}

func (e *MemzentEngine) Process(ctx context.Context, req *PromptRequest) (*PromptResponse, error) {
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

	// A.1 Billing Pre-Check (Bypass check for internal dashboard sessions / JWT users)
	authMethod, _ := ctx.Value("auth_method").(string)
	if e.ledger != nil && authMethod != "jwt" {
		hasBalance, err := e.ledger.HasSufficientBalance(ctx, orgID)
		if err != nil {
			slog.Error("Billing ledger check failed", "error", err, "org_id", orgID)
		} else if !hasBalance {
			slog.Warn("Organization out of tokens", "org_id", orgID)
			return nil, fmt.Errorf("payment required: token balance depleted")
		}
	}

	// B. Stage 1-2 Cache Lookup (Org-Isolated)
	if e.cache != nil && !req.SkipCache {
		cacheKey := fmt.Sprintf("org:%s:p:%s", orgID, req.Prompt)
		cachedResp, err := e.cache.GetSemanticResult(ctx, cacheKey)
		if err != nil || cachedResp == "" {
			// Valkey cache miss or restart/crash - fallback to persistent DB cache
			cachedResp, _ = e.getPersistentCache(ctx, cacheKey)
			if cachedResp != "" {
				slog.Info("🎯 Stage 1 Cache HIT (Durable DB Fallback)", "org_id", orgID, "key", cacheKey)
				// Backfill Valkey asynchronously
				go func() {
					_ = e.cache.SetResult(context.Background(), cacheKey, cachedResp, e.cacheTTL)
				}()
			}
		}

		if cachedResp != "" {
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
			e.chargeCacheHit(ctx, orgID, req.Provider, req.Model, req.Prompt)
			return &PromptResponse{Text: cachedResp, Cached: true}, nil
		}

		// Stage 1.5: Canonical Match (Normalized & Org-Isolated)
		_, cHash := NormalizePrompt(req.Prompt)
		canonKey := fmt.Sprintf("org:%s:c:%s", orgID, cHash)
		cachedCanon, err := e.cache.GetSemanticResult(ctx, canonKey)
		if err != nil || cachedCanon == "" {
			// Fallback to persistent DB cache
			cachedCanon, _ = e.getPersistentCache(ctx, canonKey)
			if cachedCanon != "" {
				slog.Info("🎯 Stage 1.5 Cache HIT (Durable DB Fallback)", "org_id", orgID, "canonical", cHash)
				// Backfill Valkey asynchronously
				go func() {
					_ = e.cache.SetResult(context.Background(), canonKey, cachedCanon, e.cacheTTL)
				}()
			}
		}

		if cachedCanon != "" {
			e.CacheHits.Add(1)
			hitCounter, _ := e.orgHits.LoadOrStore(orgID, &atomic.Uint64{})
			hitCounter.(*atomic.Uint64).Add(1)
			slog.Info("🎯 Stage 1.5 Cache HIT (Org-Isolated)", "org_id", orgID, "canonical", cHash)
			_ = e.cache.SetResult(ctx, cacheKey, cachedCanon, e.cacheTTL)
			e.chargeCacheHit(ctx, orgID, req.Provider, req.Model, req.Prompt)
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
		if err != nil || cachedResp == "" {
			// Fallback to persistent DB cache
			cachedResp, _ = e.getPersistentCache(ctx, simKey)
			if cachedResp != "" {
				slog.Info("🎯 Stage 2 Cache HIT (Durable DB Fallback)", "org_id", orgID, "similar_hash", similarPromptHash)
				// Backfill Valkey asynchronously
				go func() {
					_ = e.cache.SetResult(context.Background(), simKey, cachedResp, e.cacheTTL)
				}()
			}
		}

		if cachedResp != "" {
			e.CacheHits.Add(1)
			hitCounter, _ := e.orgHits.LoadOrStore(orgID, &atomic.Uint64{})
			hitCounter.(*atomic.Uint64).Add(1)
			slog.Info("🎯 Stage 2 Cache HIT (Org-Isolated)", "org_id", orgID, "similar_hash", similarPromptHash)
			e.chargeCacheHit(ctx, orgID, req.Provider, req.Model, req.Prompt)
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

	aiResp, tokenUsage, err := selectedProvider.Generate(ctx, contextPrompt, llmTools, req.Model)
	if err != nil {
		slog.Error("LLM generation failed", "error", err, "provider", selectedProvider.GetProviderName())
		return nil, err
	}

	if e.ledger != nil && e.costCalc != nil && tokenUsage != nil {
		cost := e.costCalc.CalculateCost(selectedProvider.GetMetadata().Name, req.Model, tokenUsage.PromptTokens, tokenUsage.CompletionTokens)
		if cost > 0 {
			go func() {
				_ = e.ledger.Deduct(context.Background(), orgID, cost, "llm_usage", fmt.Sprintf("Generation via %s", selectedProvider.GetProviderName()))
			}()
		}
	}

	// G. Populate Cache for future requests (Org-Isolated & Force Refresh)
	if e.cache != nil {
		// Layer 1: Literal Match
		cacheKey := fmt.Sprintf("org:%s:p:%s", orgID, req.Prompt)
		_ = e.cache.SetResult(ctx, cacheKey, aiResp, e.cacheTTL)
		e.setPersistentCache(ctx, orgID, cacheKey, aiResp, e.cacheTTL)

		// Layer 2: Canonical Match
		_, cHash := NormalizePrompt(req.Prompt)
		canonKey := fmt.Sprintf("org:%s:c:%s", orgID, cHash)
		_ = e.cache.SetResult(ctx, canonKey, aiResp, e.cacheTTL)
		e.setPersistentCache(ctx, orgID, canonKey, aiResp, e.cacheTTL)

		// Layer 3: Semantic Match
		if currentPromptHash != "" {
			simKey := fmt.Sprintf("org:%s:s:%s", orgID, currentPromptHash)
			_ = e.cache.SetResult(ctx, simKey, aiResp, e.cacheTTL)
			e.setPersistentCache(ctx, orgID, simKey, aiResp, e.cacheTTL)
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

func (e *MemzentEngine) chargeCacheHit(ctx context.Context, orgID, provider, model, prompt string) {
	if e.ledger != nil && e.costCalc != nil {
		// Rough estimate: 1 token = ~4 chars
		estimatedTokens := len(prompt) / 4
		
		providerName := provider
		if providerName == "" {
			providerName = e.defaultProvider
		}

		cost := e.costCalc.CalculateCacheDiscount(providerName, model, estimatedTokens)
		if cost > 0 {
			go func() {
				// Async deduction to not block latency
				_ = e.ledger.Deduct(context.Background(), orgID, cost, "cache_hit", "Semantic Cache Hit Discount")
			}()
		}
	}
}

func (e *MemzentEngine) getPersistentCache(ctx context.Context, cacheKey string) (string, error) {
	if e.rbac == nil || e.rbac.GetDB() == nil {
		return "", nil
	}

	var response string
	query := "SELECT response FROM persistent_cache WHERE cache_key = $1 AND expires_at > NOW()"
	err := e.rbac.GetDB().QueryRowContext(ctx, query, cacheKey).Scan(&response)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return response, nil
}

func (e *MemzentEngine) setPersistentCache(ctx context.Context, orgID, cacheKey, response string, ttl time.Duration) {
	if e.rbac == nil || e.rbac.GetDB() == nil {
		return
	}

	expiresAt := time.Now().Add(ttl)
	query := `
		INSERT INTO persistent_cache (org_id, cache_key, response, expires_at)
		VALUES ($1::uuid, $2, $3, $4)
		ON CONFLICT (cache_key) 
		DO UPDATE SET response = EXCLUDED.response, expires_at = EXCLUDED.expires_at, updated_at = NOW()
	`
	go func() {
		// Run in background so we never block prompt execution
		backgroundCtx := context.Background()
		_, err := e.rbac.GetDB().ExecContext(backgroundCtx, query, orgID, cacheKey, response, expiresAt)
		if err != nil {
			slog.Error("Failed to write to persistent database cache", "error", err, "key", cacheKey)
		}
	}()
}
