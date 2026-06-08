package engine

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"memzent-gateway/internal/auth"
	"memzent-gateway/internal/billing"
	cch "memzent-gateway/internal/cache"
	"memzent-gateway/internal/connectors"
	lp "memzent-gateway/internal/llm"
	mc "memzent-gateway/internal/mcp"
	"memzent-gateway/internal/memory"
	"memzent-gateway/internal/metrics"
	"memzent-gateway/internal/offline"
	rtr "memzent-gateway/internal/router"
	toolspkg "memzent-gateway/internal/tools"

	"golang.org/x/time/rate"
)

// PromptRequest defines the incoming user payload
type PromptRequest struct {
	UserID    string       `json:"user_id"`
	SessionID string       `json:"session_id,omitempty"`
	Messages  []lp.Message `json:"messages"`
	Provider  string       `json:"provider,omitempty"`   // e.g. "ollama", "openai", "anthropic", "gemini"
	Model     string       `json:"model,omitempty"`      // optional per-request model override
	SkipCache bool         `json:"skip_cache,omitempty"` // set by X-Skip-Cache header
	Stream    bool         `json:"stream,omitempty"`
}

// PromptResponse defines the gateway's response to the client
type PromptResponse struct {
	Text      string            `json:"text"`
	Cached    bool              `json:"cached"`
	Provider  string            `json:"provider,omitempty"`
	Tools     []any             `json:"tools,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
	SessionID string            `json:"session_id,omitempty"`
	Entities  map[string]string `json:"entities,omitempty"` // extracted entities from the prompt (E1)
}

// rateLimiterEntry is retained for backward compatibility but the primary
// rate limiting is now distributed via Valkey (see cache.RateLimit).
type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// MemzentEngine orchestrates the flow between Cache, RBAC, Router, MCP, and LLM
type MemzentEngine struct {
	cache             *cch.MemzentCache
	router            rtr.SemanticRouterInterface
	rbac              *auth.RBACClient
	ledger            *billing.Ledger
	costCalc          *billing.CostCalculator
	providers         map[string]lp.Provider // keyed by provider name e.g. "ollama"
	defaultProvider   string                 // key used when no X-Memzent-Provider header is set
	mcp               *mc.MCPClient
	registry          *toolspkg.Registry            // Registry for user-provisioned tools
	connectorRegistry *connectors.ConnectorRegistry // Core/Native connectors
	toolThreshold     float64
	cacheTTL          time.Duration
	rateLimiters      sync.Map // Deprecated: retained for graceful fallback; primary rate limiting is via Valkey
	auditLogger       *metrics.PersistentAuditLogger

	// Memory & Telemetry extensions
	sessionMgr *memory.SessionManager
	memoryMgr  *memory.MemoryManager
	telemetry  *metrics.TelemetryAggregator

	// Webhook event emitter (Phase 7)
	eventEmitter EventEmitter

	// Offline Learning Plane (E3)
	offlinePlane *offline.Plane

	// Workflow Registry (E4)
	workflowRegistry WorkflowRegistry

	TotalRequests atomic.Uint64
	CacheHits     atomic.Uint64
	orgRequests   sync.Map // Tracks requests per org (map[string]*atomic.Uint64)
	orgHits       sync.Map // Tracks cache hits per org (map[string]*atomic.Uint64)
}

// EventEmitter abstracts webhook event dispatch so engine doesn't import notifications directly
type EventEmitter interface {
	Emit(ctx context.Context, orgID string, eventType string, data any)
}

// SetEventEmitter attaches a webhook dispatcher to the engine
func (e *MemzentEngine) SetEventEmitter(emitter EventEmitter) {
	e.eventEmitter = emitter
}

// SetOfflinePlane attaches the offline learning plane to the engine.
func (e *MemzentEngine) SetOfflinePlane(plane *offline.Plane) {
	e.offlinePlane = plane
}

// WorkflowRegistry abstracts workflow lookup so engine doesn't import workflow package directly.
type WorkflowRegistry interface {
	MatchWorkflow(ctx context.Context, orgID string, toolPattern string) (matched bool, toolIDs []string, workflowID string, err error)
	RecordExecution(ctx context.Context, workflowID, orgID, promptHash string, entities map[string]string, success bool, latencyMs int) error
}

// SetWorkflowRegistry attaches the workflow registry to the engine.
func (e *MemzentEngine) SetWorkflowRegistry(registry WorkflowRegistry) {
	e.workflowRegistry = registry
}

// emitOffline sends an event to the offline learning plane (non-blocking).
func (e *MemzentEngine) emitOffline(event offline.OfflineEvent) {
	if e.offlinePlane != nil {
		e.offlinePlane.Emit(event)
	}
}

func NewMemzentEngine(
	cache *cch.MemzentCache,
	rtrClient rtr.SemanticRouterInterface,
	rbacClient *auth.RBACClient,
	ledger *billing.Ledger,
	costCalc *billing.CostCalculator,
	mcp *mc.MCPClient,
	reg *toolspkg.Registry,
	connReg *connectors.ConnectorRegistry,
	providers map[string]lp.Provider,
	defaultProvider string,
	threshold float64,
	ttl time.Duration,
	audit *metrics.PersistentAuditLogger,
	sessionMgr *memory.SessionManager,
	memoryMgr *memory.MemoryManager,
	telemetry *metrics.TelemetryAggregator,
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
		sessionMgr:        sessionMgr,
		memoryMgr:         memoryMgr,
		telemetry:         telemetry,
	}
}

// StartRateLimiterEviction runs a background goroutine that removes stale rate limiter
// entries from the sync.Map every 10 minutes. Without this the map grows unbounded —
// one entry per unique orgID:userID pair, never freed.
func (e *MemzentEngine) StartRateLimiterEviction(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cutoff := time.Now().Add(-30 * time.Minute)
				e.rateLimiters.Range(func(key, value any) bool {
					if entry, ok := value.(*rateLimiterEntry); ok {
						if entry.lastSeen.Before(cutoff) {
							e.rateLimiters.Delete(key)
						}
					}
					return true
				})
			}
		}
	}()
}

// StartModelDiscovery kicks off the background model discovery loop for providers that support it.
func (e *MemzentEngine) StartModelDiscovery(ctx context.Context) {
	discover := func() {
		for name, p := range e.providers {
			if discoverer, ok := p.(lp.ModelDiscoverer); ok {
				slog.Info("Running model discovery", "provider", name)
				models, err := discoverer.DiscoverModels(ctx)
				if err != nil {
					slog.Warn("Model discovery failed", "provider", name, "error", err)
				} else {
					slog.Info("Model discovery succeeded", "provider", name, "models", models)
				}
			}
		}
	}

	go func() {
		// Run once on startup asynchronously to prevent blocking server launch
		discover()

		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				discover()
			}
		}
	}()
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

func (e *MemzentEngine) fitToolParameters(ctx context.Context, provider lp.Provider, queryPrompt string, tool *toolspkg.Tool) (map[string]interface{}, error) {
	if tool.InputSchema == nil || len(tool.InputSchema) == 0 {
		return make(map[string]interface{}), nil
	}

	schemaBytes, err := json.Marshal(tool.InputSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input schema: %w", err)
	}

	extractionPrompt := fmt.Sprintf(`Analyze the user prompt and extract parameters that match the following JSON Schema for the tool "%s" (%s).
JSON Schema:
%s

User Prompt:
"%s"

Extract the parameters and output a JSON object containing ONLY the keys and values defined in the schema. Output raw JSON ONLY. No markdown block wrappers, no other explanation.`, tool.Name, tool.Description, string(schemaBytes), queryPrompt)

	messages := []lp.Message{
		{Role: "user", Content: extractionPrompt},
	}

	response, _, err := provider.Generate(ctx, messages, nil, "")
	if err != nil {
		return nil, fmt.Errorf("lightweight parameter fitting failed: %w", err)
	}

	// Locate JSON boundaries in response to handle LLM wraps
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")
	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		return make(map[string]interface{}), nil
	}

	cleanJSON := response[startIdx : endIdx+1]
	var parsedParams map[string]interface{}
	if err := json.Unmarshal([]byte(cleanJSON), &parsedParams); err != nil {
		slog.Warn("Failed to unmarshal parsed parameters", "error", err, "raw", cleanJSON)
		return make(map[string]interface{}), nil
	}

	return parsedParams, nil
}

func (e *MemzentEngine) Process(ctx context.Context, req *PromptRequest) (*PromptResponse, error) {
	e.TotalRequests.Add(1)

	if req == nil {
		return nil, fmt.Errorf("invalid request")
	}

	var queryPrompt string
	if len(req.Messages) > 0 {
		queryPrompt = req.Messages[len(req.Messages)-1].Content
	}
	if queryPrompt == "" {
		return nil, fmt.Errorf("no messages provided")
	}

	processStart := time.Now()

	// A. Rate Limiting (Based on Tier extracted from JWT)
	tier, _ := ctx.Value("tier").(string)
	orgID, _ := ctx.Value("org_id").(string)
	if orgID == "" {
		orgID = "default"
	}

	// Track per-org requests
	reqCounter, _ := e.orgRequests.LoadOrStore(orgID, &atomic.Uint64{})
	reqCounter.(*atomic.Uint64).Add(1)

	// Fetch organization settings, balance, and transaction history details
	var settings *billing.OrgSettings
	var settingsErr error
	var recentTxs []billing.Transaction
	if e.ledger != nil {
		settings, settingsErr = e.ledger.GetOrgSettings(ctx, orgID)
		if settingsErr != nil {
			slog.Error("Failed to fetch organization settings", "org_id", orgID, "error", settingsErr)
		} else if orgID != "default" && orgID != "" && isBillingQuery(queryPrompt) {
			var txErr error
			recentTxs, txErr = e.ledger.GetRecentTransactions(ctx, orgID, 5)
			if txErr != nil {
				slog.Error("Failed to fetch recent transactions", "org_id", orgID, "error", txErr)
			}
		}
	}

	// Dynamic Rate Limiting Based on Tier (Distributed via Valkey)
	// Org-level aggregate limit
	orgLimit := int64(10) // Free default
	if tier == "pro" {
		orgLimit = 100
	} else if tier == "business" {
		orgLimit = 1000
	}

	// Pay-as-you-go boost: if they have a positive token balance, promote rate limit from free (10) to pro (100)
	if orgLimit < 100 && settings != nil && settings.TokenBalance > 0 {
		orgLimit = 100
	}

	// Per-user limit within org: viewers get 20% of org limit, agents get 50%, admins get full
	userRole, _ := ctx.Value("user_role").(string)
	userLimit := orgLimit // Default: admin/owner gets full org limit
	switch userRole {
	case "viewer":
		userLimit = max(orgLimit/5, 5) // Minimum 5 RPM for viewers
	case "member", "agent":
		userLimit = max(orgLimit/2, 10) // Members/agents get 50%
	case "admin", "owner", "":
		userLimit = orgLimit // Full access
	}

	// Check org-level rate limit first
	orgLimitKey := fmt.Sprintf("rl:%s", orgID)
	if e.cache != nil {
		allowed, rlErr := e.cache.RateLimit(ctx, orgLimitKey, orgLimit)
		if rlErr != nil {
			slog.Warn("Distributed org rate limit check failed, falling back to allow", "error", rlErr)
		} else if !allowed {
			e.emitEvent(ctx, orgID, "rate_limit", map[string]any{"user_id": req.UserID, "limit": orgLimit, "window": "1m", "scope": "org"})
			return nil, fmt.Errorf("rate limit exceeded for organization %s (tier: %s)", orgID, tier)
		}
	}

	// Check per-user rate limit within org
	limitKey := fmt.Sprintf("rl:%s:%s", orgID, req.UserID)
	if e.cache != nil {
		allowed, rlErr := e.cache.RateLimit(ctx, limitKey, userLimit)
		if rlErr != nil {
			slog.Warn("Distributed user rate limit check failed, falling back to allow", "error", rlErr)
		} else if !allowed {
			e.emitEvent(ctx, orgID, "rate_limit", map[string]any{"user_id": req.UserID, "limit": userLimit, "window": "1m", "scope": "user"})
			return nil, fmt.Errorf("rate limit exceeded for user %s (role: %s, limit: %d RPM)", req.UserID, userRole, userLimit)
		}
	}

	// Permission check: viewers cannot execute prompts (read-only access)
	if userRole == "viewer" {
		return nil, fmt.Errorf("permission denied: viewer role cannot execute prompts (user: %s)", req.UserID)
	}

	// A.1 Billing Pre-Check (Bypass check for internal dashboard sessions / JWT users)
	authMethod, _ := ctx.Value("auth_method").(string)
	if e.ledger != nil && authMethod != "jwt" {
		if settingsErr != nil {
			slog.Error("Billing ledger settings fetch failed, blocking transaction", "error", settingsErr, "org_id", orgID)
			return nil, fmt.Errorf("internal server error: failed to verify organization profile")
		} else if settings != nil && settings.TokenBalance <= 0 && orgID != "default" && orgID != "" {
			slog.Warn("Organization out of tokens", "org_id", orgID)
			return nil, fmt.Errorf("payment required: token balance depleted")
		}

		// A.2 Spend Limit Check (daily/monthly dollar + token caps)
		if spendStatus, err := e.ledger.CheckSpendLimits(ctx, orgID); err == nil && spendStatus != nil {
			if spendStatus.DailyExceeded {
				slog.Warn("Daily spend limit exceeded", "org_id", orgID, "spent", spendStatus.DailySpend, "limit", *spendStatus.DailyLimit)
				return nil, fmt.Errorf("daily spend limit reached ($%.2f of $%.2f). Resets at midnight UTC", spendStatus.DailySpend, *spendStatus.DailyLimit)
			}
			if spendStatus.MonthlyExceeded {
				slog.Warn("Monthly spend limit exceeded", "org_id", orgID, "spent", spendStatus.MonthlySpend, "limit", *spendStatus.MonthlyLimit)
				return nil, fmt.Errorf("monthly spend limit reached ($%.2f of $%.2f). Resets on the 1st", spendStatus.MonthlySpend, *spendStatus.MonthlyLimit)
			}
			if spendStatus.DailyTokensExceeded {
				slog.Warn("Daily token limit exceeded", "org_id", orgID, "used", spendStatus.DailyTokensUsed, "limit", *spendStatus.DailyTokenLimit)
				return nil, fmt.Errorf("daily token limit reached (%d of %d). Resets at midnight UTC", spendStatus.DailyTokensUsed, *spendStatus.DailyTokenLimit)
			}
			if spendStatus.MonthlyTokensExceeded {
				slog.Warn("Monthly token limit exceeded", "org_id", orgID, "used", spendStatus.MonthlyTokensUsed, "limit", *spendStatus.MonthlyTokenLimit)
				return nil, fmt.Errorf("monthly token limit reached (%d of %d). Resets on the 1st", spendStatus.MonthlyTokensUsed, *spendStatus.MonthlyTokenLimit)
			}
		}
	}

	// Resolve selected provider and model for cache key partitioning
	providerKey := req.Provider
	if providerKey == "" {
		if settings != nil && settings.DefaultProvider != "" {
			providerKey = settings.DefaultProvider
		} else {
			providerKey = e.defaultProvider
		}
	}
	selectedProvider, ok := e.providers[providerKey]
	if !ok {
		selectedProvider = e.providers[e.defaultProvider]
	}
	resolvedModel := req.Model
	if resolvedModel == "" {
		if settings != nil && settings.DefaultModel != "" {
			resolvedModel = settings.DefaultModel
		} else if selectedProvider != nil {
			resolvedModel = selectedProvider.GetMetadata().DefaultModel
		}
	}
	if resolvedModel == "" {
		resolvedModel = "default"
	}

	// B. Stage 1-2 Cache Lookup (Org-Isolated & Model-Scoped)
	if e.cache != nil && !req.SkipCache {
		cacheKey := e.buildCacheKey(orgID, "p", resolvedModel, queryPrompt)
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
			e.emitEvent(ctx, orgID, "cache_hit", map[string]any{"query": queryPrompt, "score": 1.0, "latency_ms": time.Since(processStart).Milliseconds(), "model": resolvedModel})
			if e.auditLogger != nil {
				e.auditLogger.Log(ctx, metrics.AuditEvent{
					Timestamp:  time.Now(),
					OrgID:      orgID,
					Type:       "CACHE",
					User:       req.UserID,
					Detail:     "Stage 1 HIT: " + queryPrompt,
					Status:     "success",
					CacheLayer: "L1",
				}, map[string]interface{}{"prompt": queryPrompt, "stage": 1})
			}
			e.chargeCacheHit(ctx, orgID, req.Provider, req.Model, queryPrompt)

			// Append user message and cached response to chat session history
			if req.SessionID != "" && e.sessionMgr != nil {
				_ = e.sessionMgr.AppendMessage(ctx, req.SessionID, "user", queryPrompt)
				_ = e.sessionMgr.AppendMessage(ctx, req.SessionID, "assistant", cachedResp)
			}
			e.emitOffline(offline.OfflineEvent{
				OrgID: orgID, UserID: req.UserID,
				PromptHash: sha256Hex(queryPrompt),
				CacheLayer: "L1", LatencyMs: time.Since(processStart).Milliseconds(),
				Provider: req.Provider, Model: resolvedModel, Success: true, Timestamp: time.Now(),
			})
			return &PromptResponse{Text: cachedResp, Cached: true, SessionID: req.SessionID}, nil
		}

		// Stage 1.5: Canonical Match (Normalized, Org-Isolated & Model-Scoped)
		_, cHash := NormalizePrompt(queryPrompt)
		canonKey := e.buildCacheKey(orgID, "c", resolvedModel, cHash)
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
			// NOTE: Do NOT backfill the literal cache from canonical matches.
			// If the canonical normalization is imprecise, backfilling would poison
			// the literal cache with incorrect responses for future exact-match lookups.
			e.chargeCacheHit(ctx, orgID, req.Provider, req.Model, queryPrompt)

			// Append user message and cached response to chat session history
			if req.SessionID != "" && e.sessionMgr != nil {
				_ = e.sessionMgr.AppendMessage(ctx, req.SessionID, "user", queryPrompt)
				_ = e.sessionMgr.AppendMessage(ctx, req.SessionID, "assistant", cachedCanon)
			}
			return &PromptResponse{Text: cachedCanon, Cached: true, SessionID: req.SessionID}, nil
		}

		// Stage 1b: Entity-Keyed Hot Path Cache (L1b)
		// Fast entity extraction via regex, then deterministic key lookup in Valkey.
		// Only fires if we have a non-trivial prompt (entities can be extracted client-side).
		l1bEntities := extractEntitiesLocal(queryPrompt)
		if len(l1bEntities) > 0 {
			entityKey := e.buildEntityCacheKey(orgID, resolvedModel, l1bEntities)
			if entityKey != "" {
				cachedEntity, err := e.cache.GetSemanticResult(ctx, entityKey)
				if err != nil || cachedEntity == "" {
					cachedEntity, _ = e.getPersistentCache(ctx, entityKey)
					if cachedEntity != "" {
						go func() {
							_ = e.cache.SetResult(context.Background(), entityKey, cachedEntity, e.cacheTTL)
						}()
					}
				}

				if cachedEntity != "" {
					e.CacheHits.Add(1)
					hitCounter, _ := e.orgHits.LoadOrStore(orgID, &atomic.Uint64{})
					hitCounter.(*atomic.Uint64).Add(1)
					slog.Info("🎯 Stage 1b Cache HIT (Entity-Keyed)", "org_id", orgID, "entities", l1bEntities)
						e.emitEvent(ctx, orgID, "cache_hit", map[string]any{"query": queryPrompt, "score": 1.0, "latency_ms": time.Since(processStart).Milliseconds(), "model": resolvedModel, "layer": "L1b"})
						if e.auditLogger != nil {
							e.auditLogger.Log(ctx, metrics.AuditEvent{
								Timestamp:  time.Now(),
								OrgID:      orgID,
								Type:       "CACHE",
								User:       req.UserID,
								Detail:     "Stage 1b HIT (Entity-Keyed): " + queryPrompt,
								Status:     "success",
								CacheLayer: "L1b",
								Entities:   l1bEntities,
							}, map[string]interface{}{"prompt": queryPrompt, "stage": "1b", "entities": l1bEntities})
						}
						e.chargeCacheHit(ctx, orgID, req.Provider, req.Model, queryPrompt)

					if req.SessionID != "" && e.sessionMgr != nil {
						_ = e.sessionMgr.AppendMessage(ctx, req.SessionID, "user", queryPrompt)
						_ = e.sessionMgr.AppendMessage(ctx, req.SessionID, "assistant", cachedEntity)
					}
					e.emitOffline(offline.OfflineEvent{
						OrgID: orgID, UserID: req.UserID,
						PromptHash: sha256Hex(queryPrompt), Entities: l1bEntities,
						EntitySource: "regex", CacheLayer: "L1b",
						LatencyMs: time.Since(processStart).Milliseconds(),
						Provider: req.Provider, Model: resolvedModel, Success: true, Timestamp: time.Now(),
					})
					return &PromptResponse{Text: cachedEntity, Cached: true, SessionID: req.SessionID, Entities: l1bEntities}, nil
				}
			}
		}
	}

	// 1. Short-term Memory: Load previous messages if SessionID is provided
	var historyMessages []lp.Message
	if req.SessionID != "" && e.sessionMgr != nil {
		// Save new user message to session in DB
		err := e.sessionMgr.AppendMessage(ctx, req.SessionID, "user", queryPrompt)
		if err != nil {
			slog.Error("Failed to append user message to session history", "session_id", req.SessionID, "error", err)
		}

		historyMessages, err = e.sessionMgr.GetSessionMessages(ctx, req.SessionID, 20)
		if err != nil {
			slog.Error("Failed to fetch session history messages", "session_id", req.SessionID, "error", err)
		}
	}

	var messagesToLLM []lp.Message
	if len(historyMessages) > 0 {
		messagesToLLM = historyMessages
	} else {
		messagesToLLM = make([]lp.Message, len(req.Messages))
		copy(messagesToLLM, req.Messages)
	}

	// 2. Long-term Memory: Retrieve related facts from Qdrant via memories_collection
	var memoryContext string
	if e.memoryMgr != nil {
		var err error
		memoryContext, err = e.memoryMgr.RetrieveSemanticContext(ctx, queryPrompt, orgID, req.UserID, 0.65)
		if err != nil {
			slog.Error("Failed to retrieve semantic guidelines", "error", err)
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
	tools, compressedPrompt, similarPromptHash, currentPromptHash, extractedEntities, err := e.router.GetBestTools(ctx, queryPrompt, orgID, allowedTools, req.SkipCache)
	if err != nil {
		slog.Warn("Router fallback engaged", "error", err)
	}
	if len(extractedEntities) > 0 {
		slog.Info("🏷️ Entities extracted", "org_id", orgID, "entities", extractedEntities)
	}
	entitySource := "none"
	if len(extractedEntities) > 0 {
		entitySource = "regex" // from Rust router's regex extractors
	}

	// NEW: Stage 2 Cache Check (Fuzzy Vector Semantic Match) — Org-Isolated & Model-Scoped
	if similarPromptHash != "" && e.cache != nil && !req.SkipCache {
		simKey := e.buildCacheKey(orgID, "s", resolvedModel, similarPromptHash)
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
			e.chargeCacheHit(ctx, orgID, req.Provider, req.Model, queryPrompt)

			if req.SessionID != "" && e.sessionMgr != nil {
				_ = e.sessionMgr.AppendMessage(ctx, req.SessionID, "assistant", cachedResp)
			}
			return &PromptResponse{Text: cachedResp, Cached: true, SessionID: req.SessionID}, nil
		}
	}

	// E. Tool Execution (Multi-Connector: Universal Provisioning & Chaining support)
	var toolResults []string
	useChaining := false
	if len(tools) > 1 && (strings.Contains(strings.ToLower(queryPrompt), "then") ||
		strings.Contains(strings.ToLower(queryPrompt), "after") ||
		strings.Contains(strings.ToLower(queryPrompt), "sequence") ||
		strings.Contains(strings.ToLower(queryPrompt), "chain") ||
		strings.Contains(strings.ToLower(queryPrompt), "first")) {
		useChaining = true
	}

	if useChaining && e.router != nil {
		steps, confidence, err := e.router.PlanToolChain(ctx, queryPrompt, orgID, allowedTools)
		if err == nil && len(steps) > 1 && confidence > 0.5 {
			slog.Info("⛓️ Sequential tool chaining activated", "steps_count", len(steps), "confidence", confidence)

			var lastOutput string
			for _, step := range steps {
				slog.Info("Executing chain step", "order", step.StepOrder, "tool_name", step.ToolName)

				var toolMetadata *toolspkg.Tool
				allTools, err := e.registry.ListTools(ctx, orgID)
				if err == nil {
					for _, item := range allTools {
						if item.Name == step.ToolName || item.ID == step.ToolName {
							toolMetadata = item
							break
						}
					}
				}

				if toolMetadata == nil {
					slog.Warn("Chain tool not found in registry", "tool_name", step.ToolName)
					continue
				}

				stepPrompt := queryPrompt
				if lastOutput != "" {
					stepPrompt = fmt.Sprintf("%s\n\nPrevious step output context: %s", queryPrompt, lastOutput)
				}

				inputs, err := e.fitToolParameters(ctx, selectedProvider, stepPrompt, toolMetadata)
				if err != nil {
					slog.Error("Failed to fit parameters for chain step", "tool_id", toolMetadata.ID, "error", err)
					inputs = make(map[string]interface{})
				}

				var connector connectors.Connector
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
					if reg, ok := e.connectorRegistry.Get(connectors.TypeCore); ok {
						if cc, ok := reg.(*connectors.CoreConnector); ok && cc.HasTool(toolMetadata.ID) {
							connector = reg
						}
					}
				}

				if connector == nil {
					slog.Warn("No connector available for chain tool", "tool_id", toolMetadata.ID)
					continue
				}

				execReq := &connectors.ExecutionRequest{
					ToolID:  toolMetadata.ID,
					UserID:  req.UserID,
					Inputs:  inputs,
					Timeout: toolMetadata.TimeoutSeconds,
				}

				startTime := time.Now()
				toolCtx, cancel := context.WithTimeout(ctx, time.Duration(toolMetadata.TimeoutSeconds+1)*time.Second)
					execResp, err := connectors.ExecuteWithRetry(toolCtx, connector, execReq, connectors.DefaultRetryConfig())
				cancel()
				duration := time.Since(startTime)

				status := "success"
				errMsg := ""
				if err != nil {
					status = "failure"
					errMsg = err.Error()
					slog.Error("Chain step execution failed", "tool_id", toolMetadata.ID, "error", err)
					} else if execResp != nil && execResp.Status == "failure" {
					status = "failure"
					errMsg = fmt.Sprintf("%v", execResp.Data)
				}

				if e.telemetry != nil {
					e.telemetry.LogToolExecution(ctx, orgID, toolMetadata.ID, req.SessionID, int(duration.Milliseconds()), status, errMsg)
				}

				if status == "success" && execResp.Data != nil {
					lastOutput = fmt.Sprintf("%v", execResp.Data)
					toolResults = append(toolResults, fmt.Sprintf("Step %d (%s): %s", step.StepOrder, toolMetadata.Name, lastOutput))
				}
			}
			useChaining = len(toolResults) > 0
		}
	}

	if !useChaining && len(tools) > 0 {
		for _, t := range tools {
			if t.RelevanceScore > float32(e.toolThreshold) {
				slog.Info("Resolving tool for execution", "tool_id", t.Id, "score", t.RelevanceScore)

				toolMetadata, err := e.registry.GetTool(ctx, t.Id)
				if err != nil {
					slog.Error("Failed to fetch tool metadata", "tool_id", t.Id, "error", err)
					continue
				}

				if toolMetadata == nil {
					slog.Warn("Tool not found in registry", "tool_id", t.Id)
					continue
				}

				inputs, err := e.fitToolParameters(ctx, selectedProvider, queryPrompt, toolMetadata)
				if err != nil {
					slog.Error("Failed to dynamically fit parameters", "tool_id", t.Id, "error", err)
					inputs = make(map[string]interface{})
				}

				var connector connectors.Connector
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

				execReq := &connectors.ExecutionRequest{
					ToolID:  t.Id,
					UserID:  req.UserID,
					Inputs:  inputs,
					Timeout: toolMetadata.TimeoutSeconds,
				}

				startTime := time.Now()
				toolCtx, cancel := context.WithTimeout(ctx, time.Duration(toolMetadata.TimeoutSeconds+1)*time.Second)
					execResp, err := connectors.ExecuteWithRetry(toolCtx, connector, execReq, connectors.DefaultRetryConfig())
				cancel()
				duration := time.Since(startTime)

				status := "success"
				errMsg := ""
				if err != nil {
					status = "failure"
					errMsg = err.Error()
					slog.Error("Tool execution failed", "tool_id", t.Id, "error", err)
					} else if execResp != nil && execResp.Status == "failure" {
					status = "failure"
					errMsg = fmt.Sprintf("%v", execResp.Data)
				}

				if e.telemetry != nil {
					e.telemetry.LogToolExecution(ctx, orgID, t.Id, req.SessionID, int(duration.Milliseconds()), status, errMsg)
				}

				if status == "success" && execResp.Data != nil {
					toolResults = append(toolResults, fmt.Sprintf("%v", execResp.Data))
				}
			}
		}
	}

	// F. Build LLM context from compressed prompt + memory context + tool results
	lastMsgIdx := len(messagesToLLM) - 1
	if lastMsgIdx >= 0 {
		currentContent := messagesToLLM[lastMsgIdx].Content
		if compressedPrompt != "" {
			currentContent = compressedPrompt
		}
		if memoryContext != "" {
			currentContent = fmt.Sprintf("%s\n\n%s", currentContent, memoryContext)
		}
		if len(toolResults) > 0 {
			currentContent = fmt.Sprintf("%s\n\n### SUPPLEMENTARY TOOL CONTEXT\n%v\n--- END TOOL CONTEXT ---", currentContent, toolResults)
		}
		if settings != nil && isBillingQuery(queryPrompt) {
			balanceVal := settings.TokenBalance
			if orgID == "default" || orgID == "" {
				balanceVal = 999999.0
			}
			billingContext := fmt.Sprintf("### ACTUAL ORGANIZATIONAL BILLING CONTEXT\n"+
				"Instructions for LLM: Use this billing context ONLY if the user is explicitly asking about their billing, token balance, transactions, or charges. If the prompt is about something else, ignore this context completely.\n"+
				"- Current Token Balance: $%.4f\n", balanceVal)
			if len(recentTxs) > 0 {
				billingContext += "- Recent Ledger Transactions:\n"
				for _, tx := range recentTxs {
					billingContext += fmt.Sprintf("  * Timestamp: %s, Action: %s, Amount: $%.4f, Description: %s\n",
						tx.CreatedAt.Format(time.RFC3339), tx.TransactionType, tx.Amount, tx.Description)
				}
			} else {
				billingContext += "- Recent Ledger Transactions: None\n"
			}
			billingContext += "--- END BILLING CONTEXT ---"
			currentContent = fmt.Sprintf("%s\n\n%s", currentContent, billingContext)
		}
		messagesToLLM[lastMsgIdx].Content = currentContent
	}

	var llmTools []any
	for _, t := range tools {
		llmTools = append(llmTools, t)
	}

	slog.Info("🤖 LLM Provider selected", "provider", selectedProvider.GetProviderName(), "model_override", req.Model, "skip_cache", req.SkipCache)

	aiResp, tokenUsage, err := selectedProvider.Generate(ctx, messagesToLLM, llmTools, req.Model)
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

	// G. Populate Cache for future requests (Org-Isolated, Model-Scoped & Force Refresh)
	// Skip all cache writes when SkipCache=true — forced-fresh requests must not
	// populate Valkey or the persistent DB, otherwise subsequent requests for the
	// same prompt would return cached=true even though the intent was a fresh hit.
	if e.cache != nil && !req.SkipCache {
		cacheKey := e.buildCacheKey(orgID, "p", resolvedModel, queryPrompt)
		_ = e.cache.SetResult(ctx, cacheKey, aiResp, e.cacheTTL)
		e.setPersistentCache(ctx, orgID, cacheKey, aiResp, e.cacheTTL)

		_, cHash := NormalizePrompt(queryPrompt)
		canonKey := e.buildCacheKey(orgID, "c", resolvedModel, cHash)
		_ = e.cache.SetResult(ctx, canonKey, aiResp, e.cacheTTL)
		e.setPersistentCache(ctx, orgID, canonKey, aiResp, e.cacheTTL)

		if currentPromptHash != "" {
			simKey := e.buildCacheKey(orgID, "s", resolvedModel, currentPromptHash)
			_ = e.cache.SetResult(ctx, simKey, aiResp, e.cacheTTL)
			e.setPersistentCache(ctx, orgID, simKey, aiResp, e.cacheTTL)
		}

		// L1b Dual-Write: entity-keyed cache entry (only if entities were extracted)
		if len(extractedEntities) > 0 {
			entityKey := e.buildEntityCacheKey(orgID, resolvedModel, extractedEntities)
			if entityKey != "" {
				_ = e.cache.SetResult(ctx, entityKey, aiResp, e.cacheTTL)
				e.setPersistentCache(ctx, orgID, entityKey, aiResp, e.cacheTTL)
				slog.Debug("💾 L1b entity-keyed cache written", "key", entityKey)
			}
		}
	}

	// Post-Generation: Out-of-band facts extraction
	if e.memoryMgr != nil {
		e.memoryMgr.ExtractAndStoreFacts(ctx, orgID, req.UserID, queryPrompt, aiResp)
	}

	// Post-Generation: Save assistant response to session history
	if req.SessionID != "" && e.sessionMgr != nil {
		err := e.sessionMgr.AppendMessage(ctx, req.SessionID, "assistant", aiResp)
		if err != nil {
			slog.Error("Failed to append assistant response message to session history", "session_id", req.SessionID, "error", err)
		}
	}

	// H. Log Final Result
	if e.auditLogger != nil {
		e.auditLogger.Log(ctx, metrics.AuditEvent{
			Timestamp:  time.Now(),
			OrgID:      orgID,
			Type:       "GENERATION",
			User:       req.UserID,
			Detail:     fmt.Sprintf("Provider: %s", selectedProvider.GetProviderName()),
			Status:     "success",
			CacheLayer: "L5",
			Entities:   extractedEntities,
		}, map[string]interface{}{"provider": selectedProvider.GetProviderName(), "tools_count": len(llmTools)})
	}

	// Emit offline learning event for L5 resolution
	var toolNames []string
	for _, t := range llmTools {
		if tm, ok := t.(map[string]any); ok {
			if name, ok := tm["name"].(string); ok {
				toolNames = append(toolNames, name)
			}
		}
	}
	_, offlineCHash := NormalizePrompt(queryPrompt)
	e.emitOffline(offline.OfflineEvent{
		OrgID: orgID, UserID: req.UserID,
		PromptHash: sha256Hex(queryPrompt), CanonicalHash: offlineCHash,
		Entities: extractedEntities, EntitySource: entitySource,
		ToolsUsed: toolNames, CacheLayer: "L5",
		LatencyMs: time.Since(processStart).Milliseconds(),
		Provider: selectedProvider.GetProviderName(), Model: resolvedModel,
		Success: true, Timestamp: time.Now(),
	})

	return &PromptResponse{
		Text:      aiResp,
		Cached:    false,
		Provider:  selectedProvider.GetProviderName(),
		Tools:     llmTools,
		SessionID: req.SessionID,
		Entities:  extractedEntities,
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

func (e *MemzentEngine) buildCacheKey(orgID, keyType, model, value string) string {
	return fmt.Sprintf("org:%s:m:%s:%s:%s", orgID, model, keyType, value)
}

// buildEntityCacheKey creates a deterministic L1b cache key from extracted entities.
// Format: org:{orgID}:m:{model}:e:{action}:{sorted_key=value pairs}
// Returns empty string if entities are empty or have no action.
func (e *MemzentEngine) buildEntityCacheKey(orgID, model string, entities map[string]string) string {
	if len(entities) == 0 {
		return ""
	}

	// Sort entity keys for deterministic ordering
	keys := make([]string, 0, len(entities))
	for k := range entities {
		keys = append(keys, k)
	}
	// Use sort to ensure deterministic key
	sortStrings(keys)

	// Build key parts
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+entities[k])
	}

	entityStr := strings.Join(parts, ":")
	return fmt.Sprintf("org:%s:m:%s:e:%s", orgID, model, entityStr)
}

// sortStrings sorts a slice of strings in place (simple insertion sort for small slices)
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// WarmCache queries PostgreSQL persistent cache for active entries and pre-warms Valkey in the background
func (e *MemzentEngine) WarmCache(ctx context.Context) {
	if e.cache == nil || e.rbac == nil || e.rbac.GetDB() == nil {
		slog.Info("Cache warming skipped: Valkey or Postgres not initialized")
		return
	}

	slog.Info("🔥 Pre-warming memory cache from PostgreSQL persistent B-Tree...")

	query := `
		SELECT cache_key, response, expires_at 
		FROM persistent_cache 
		WHERE expires_at > NOW() 
		ORDER BY updated_at DESC 
		LIMIT 1000
	`
	rows, err := e.rbac.GetDB().QueryContext(ctx, query)
	if err != nil {
		slog.Error("Failed to query persistent cache for pre-warming", "error", err)
		return
	}
	defer rows.Close()

	warmedCount := 0
	for rows.Next() {
		var cacheKey, response string
		var expiresAt time.Time
		if err := rows.Scan(&cacheKey, &response, &expiresAt); err != nil {
			slog.Error("Failed to scan persistent cache row for pre-warming", "error", err)
			continue
		}

		remainingTTL := expiresAt.Sub(time.Now())
		if remainingTTL > 0 {
			// Write directly into Valkey
			if err := e.cache.SetResult(ctx, cacheKey, response, remainingTTL); err == nil {
				warmedCount++
			}
		}
	}

	slog.Info("🔥 Memory cache pre-warming completed successfully!", "records_warmed", warmedCount)
}

func isBillingQuery(prompt string) bool {
	p := strings.ToLower(prompt)
	keywords := []string{
		"balance", "billing", "token", "transaction", "ledger", "charge",
		"payment", "cost", "invoice", "spend", "spent", "audit", "money",
		"usd", "account status", "pricing", "usage", "cache hit", "cache discount",
		"rate limit", "tier", "credit", "budget", "fee", "fees",
	}
	for _, kw := range keywords {
		if strings.Contains(p, kw) {
			return true
		}
	}
	return false
}

// emitEvent fires a webhook notification if an emitter is configured (non-blocking)
func (e *MemzentEngine) emitEvent(ctx context.Context, orgID, eventType string, data any) {
	if e.eventEmitter != nil {
		go e.eventEmitter.Emit(ctx, orgID, eventType, data)
	}
}

// sha256Hex returns the hex-encoded SHA-256 hash of a string.
func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
