package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"crypto/rand"
	"encoding/hex"

	"memzent-gateway/internal/auth"
	"memzent-gateway/internal/billing"
	"memzent-gateway/internal/cache"
	"memzent-gateway/internal/config"
	"memzent-gateway/internal/connectors"
	"memzent-gateway/internal/db"
	"memzent-gateway/internal/engine"
	"memzent-gateway/internal/llm"
	"memzent-gateway/internal/mcp"
	"memzent-gateway/internal/memory"
	"memzent-gateway/internal/metrics"
	"memzent-gateway/internal/router"
	"memzent-gateway/internal/tools"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "memzent-gateway/docs"
)

// Replace with your module path

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		duration := time.Since(start).Seconds()

		metrics.HttpRequestsTotal.WithLabelValues(r.URL.Path, r.Method, fmt.Sprintf("%d", rw.status)).Inc()
		metrics.RequestDurationSeconds.WithLabelValues(r.URL.Path, r.Method).Observe(duration)
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *statusResponseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Memzent-Provider, X-Memzent-Model, X-Skip-Cache, X-Org-ID")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// 1. Initialize Config
	cfg := config.LoadConfig()

	// 2. Initialize Structured Logging
	var handler slog.Handler
	if cfg.Environment == "production" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	slog.SetDefault(slog.New(handler))

	slog.Info("Starting Memzent Gateway", "port", cfg.Port, "env", cfg.Environment)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 3. Initialize Cache
	vCache, err := cache.NewMemzentCache(ctx, cfg.ValkeyURL)
	if err != nil {
		slog.Error("Failed to connect to Valkey", "error", err)
		os.Exit(1)
	}
	defer vCache.Close()
	slog.Info("Connected to Valkey")

	// 4. Initialize Router Client
	rClient, err := router.NewRouterClient(ctx, cfg.RouterURL)
	if err != nil {
		slog.Error("Failed to connect to Rust Router", "error", err)
		os.Exit(1)
	}
	defer rClient.Close()
	slog.Info("Connected to Rust Router")

	// 5. Initialize RBAC Client
	rbacClient, err := auth.NewRBACClient(cfg.PostgresURL)
	if err != nil {
		slog.Warn("Postgres RBAC unavailable, starting with limited permissions", "error", err)
	} else {
		defer rbacClient.Close()
		slog.Info("Connected to Postgres RBAC")

		// 5.0. Run Automated Migrations (Infrastructure Hardening)
		migrationRunner := db.NewMigrationRunner(rbacClient.GetDB(), "migrations")
		if err := migrationRunner.Run(ctx); err != nil {
			slog.Error("Database migration failed", "error", err)
		}
	}

	// 5.1 Initialize JWKS Provider (Dynamic Auth discovery)
	var jwksProvider *auth.JWKSProvider
	if cfg.JWKSURL != "" {
		jwksProvider = auth.NewJWKSProvider(cfg.JWKSURL, cfg.SupabaseKey)
		slog.Info("JWKS Provider initialized", "url", cfg.JWKSURL)

		// Seed the known Supabase EC public key (ES256 / P-256) so that
		// JWT verification works immediately, even when the JWKS endpoint
		// returns 401 due to network policy.
		//
		// The literal JWK can be overridden via SUPABASE_STATIC_JWK env var.
		// Default is the key retrieved from the Supabase project dashboard.
		staticJWK := os.Getenv("SUPABASE_STATIC_JWK")
		if staticJWK == "" {
			staticJWK = `{"x":"UTI4xiBSLxHQs5oiBAsa-kpdkrkU0c-ZLQ05RajACOw","y":"b0Fgsaxo33a4HCdADuLLJu1XFXqTDRwXQYEkQVEvOGQ","alg":"ES256","crv":"P-256","kid":"ab27c078-c304-4414-87f2-9ca0622a565e","kty":"EC"}`
		}
		if kid, ecKey, keyErr := auth.ParseECJWKLiteral(staticJWK); keyErr != nil {
			slog.Warn("JWKS: failed to parse static JWK, remote fetch only", "error", keyErr)
		} else {
			jwksProvider.SeedKey(kid, ecKey)
		}
	}

	// 5.5. Initialize Tool Registry
	var toolRegistry *tools.Registry
	if rbacClient != nil {
		toolRegistry = tools.NewRegistry(rbacClient.GetDB())
	}

	// 6. Initialize MCP Client
	mcpClient, err := mcp.NewMCPClient()
	if err != nil {
		slog.Warn("MCP Client unavailable", "error", err)
	} else {
		// Eagerly initialize the stateful mcp-golang client handshake
		initCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if initErr := mcpClient.Initialize(initCtx); initErr != nil {
			slog.Warn("MCP Handshake failed", "error", initErr)
		} else {
			slog.Info("Connected to Internal MCP Network")
		}
	}

	// 7. Initialize LLM Provider Registry
	// All enabled providers are registered. Ollama is always present as the default.
	providers := make(map[string]llm.Provider)
	defaultProvider := "ollama"

	providers["ollama"] = llm.NewOllamaProvider(cfg.OllamaURL, cfg.OllamaModel)
	slog.Info("Provider registered: Ollama", "model", cfg.OllamaModel)

	if cfg.OpenAIAPIKey != "" {
		providers["openai"] = llm.NewOpenAIProvider(cfg.OpenAIAPIKey, cfg.OpenAIModel)
		slog.Info("Provider registered: OpenAI", "model", cfg.OpenAIModel)
	}
	if cfg.AnthropicAPIKey != "" {
		providers["anthropic"] = llm.NewAnthropicProvider(cfg.AnthropicAPIKey, "")
		slog.Info("Provider registered: Anthropic")
	}
	if cfg.GeminiAPIKey != "" {
		providers["gemini"] = llm.NewGeminiProvider(cfg.GeminiAPIKey, "")
		slog.Info("Provider registered: Gemini")
	}

	// 7.5. Initialize Connector Registry (Phase 3: Multi-Connector Framework)
	connRegistry := connectors.NewConnectorRegistry()

	// 7.4 Initialize Core Connector (Hybrid Approach: Native Go Tools)
	coreConnector := connectors.NewCoreConnector()

	// Register: read_database (Native Implementation)
	coreConnector.RegisterTool("read_database", func(ctx context.Context, userID string, inputs map[string]interface{}) (string, error) {
		slog.Info("Executing CORE tool: read_database", "user_id", userID)
		return "Mock Database Trace: Successfully indexed 1,241 cluster metrics via Memzent Core (Native Connector).", nil
	})

	// Register: memzent_search (Native Implementation)
	coreConnector.RegisterTool("memzent_search", func(ctx context.Context, userID string, inputs map[string]interface{}) (string, error) {
		slog.Info("Executing CORE tool: memzent_search", "user_id", userID)
		return "Semantic Search Results: No direct matches found in local index. Proceeding with neural expansion.", nil
	})

	connRegistry.Register(connectors.TypeCore, coreConnector)
	slog.Info("Connector registered: CORE (Native)")

	// Register MCP Connector (Phase 1 backward compatibility)
	if mcpClient != nil {
		mcpConnector := connectors.NewMCPConnector(mcpClient)
		connRegistry.Register(connectors.TypeMCP, mcpConnector)
		slog.Info("Connector registered: MCP")
	}

	// Register REST Connector (Phase 3)
	// REST connector is stateless; instance can be shared
	restConnector := connectors.NewRESTConnector("")
	connRegistry.Register(connectors.TypeREST, restConnector)
	slog.Info("Connector registered: REST (Phase 3)")

	// Register SQL Connector (Phase 3)
	// Note: SQL connector will be instantiated per-tool with connection string from tool registry
	if rbacClient != nil {
		sqlConnector := connectors.NewSQLConnector(cfg.PostgresURL)
		if err := sqlConnector.Connect(ctx); err != nil {
			slog.Warn("SQL Connector initialization failed", "error", err)
		} else {
			connRegistry.Register(connectors.TypeSQL, sqlConnector)
			slog.Info("Connector registered: SQL (Phase 3)")
		}
	}

	// 7.6 Initialize Persistent Audit Logger (RC1)
	var auditLogger *metrics.PersistentAuditLogger
	if rbacClient != nil {
		auditLogger = metrics.NewPersistentAuditLogger(rbacClient.GetDB())
		// Supported: Safe to run because prepare_threshold=0 disables SQLPrepare driver-wide, bypassing PgBouncer limits
		auditLogger.StartRetentionJob(ctx, 30) // 30-day retention
		slog.Info("Audit Logger initialized with Postgres persistence", "retention_days", 30)
	}

	// 7.7 Initialize Billing System
	var billingLedger *billing.Ledger
	costCalc := billing.NewCostCalculator()
	if rbacClient != nil {
		billingLedger = billing.NewLedger(rbacClient.GetDB())
		slog.Info("Billing Ledger initialized with Postgres")
	}

	// 7.8 Initialize memory and telemetry services
	var sessionMgr *memory.SessionManager
	var memoryMgr *memory.MemoryManager
	var telemetry *metrics.TelemetryAggregator

	if rbacClient != nil {
		sessionMgr = memory.NewSessionManager(rbacClient.GetDB())
		memoryMgr = memory.NewMemoryManager(rClient, providers, defaultProvider)
		telemetry = metrics.NewTelemetryAggregator(rbacClient.GetDB())
		slog.Info("Memory & Context Telemetry Services initialized with Postgres")
	}

	// 8. Initialize Engine
	memzentEngine := engine.NewMemzentEngine(
		vCache,
		rClient,
		rbacClient,
		billingLedger,
		costCalc,
		mcpClient,
		toolRegistry,
		connRegistry,
		providers,
		defaultProvider,
		cfg.ToolRelevanceThreshold,
		cfg.LLMCacheTTL,
		auditLogger,
		sessionMgr,
		memoryMgr,
		telemetry,
	)

	// 8.0a Start rate limiter TTL eviction — prevents unbounded memory growth
	// in long-running multi-tenant deployments (one entry per orgID:userID pair).
	memzentEngine.StartRateLimiterEviction(ctx)

	// Start background model discovery for all registered LLM providers
	memzentEngine.StartModelDiscovery(ctx)


	// 8.0 Start Tool Registry Refresh Loop (Phase 2: Dynamic Tool Sync)
	// Every 30 seconds, check Postgres for drifted tools and push them to Qdrant.
	if toolRegistry != nil {
		registrySyncCallback := func(syncCtx context.Context, driftedTools []*tools.Tool) {
			for _, t := range driftedTools {
				orgID := ""
				if t.OrgID != nil {
					orgID = *t.OrgID
				}
				ok, err := rClient.RegisterTool(syncCtx, t.ID, t.Name, t.Description, orgID)
				if err != nil || !ok {
					slog.Error("[RegistrySync] Qdrant vectorization failed", "tool_id", t.ID, "error", err)
				} else {
					slog.Info("[RegistrySync] Tool synced to Qdrant", "tool_id", t.ID, "name", t.Name)
				}
			}
		}
		go toolRegistry.StartRefreshLoop(ctx, 30*time.Second, registrySyncCallback)
		slog.Info("Tool Registry background sync started", "interval", "30s")
	}

	// 8.0.5 Pre-warm the memory cache from PostgreSQL persistent B-Tree store in the background
	go memzentEngine.WarmCache(ctx)


	// 8.1 Initialize Stripe Handler (SaaS Billing)
	stripeHandler := billing.NewStripeHandler(
		rbacClient.GetDB(),
		billingLedger,
		os.Getenv("STRIPE_WEBHOOK_SECRET"),
		os.Getenv("STRIPE_PRO_PRODUCT_ID"),
		os.Getenv("STRIPE_BIZ_PRODUCT_ID"),
		auditLogger,
	)

	mux := http.NewServeMux()

	// Metrics API
	mux.Handle("/metrics", promhttp.Handler())

	// Health Checks
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		status := map[string]string{"status": "ready"}
		if err := vCache.Ping(r.Context()); err != nil {
			status["status"] = "not_ready"
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	// --- Authenticated v1 API ---
	middleware := auth.UnifiedAuthMiddleware(cfg.JWTSecret, jwksProvider, rbacClient)

	// Scope verification wrapper
	requireScope := func(scope string, next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !auth.HasScope(r.Context(), scope) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Forbidden: API key lacks required scope '%s'", scope)})
				return
			}
			next(w, r)
		}
	}

	// Chat API
	mux.Handle("/v1/chat", middleware(requireScope("chat:execute", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		var req engine.PromptRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid Request Body", http.StatusBadRequest)
			return
		}

		// Extract identity from Middleware context
		userID, _ := r.Context().Value("user_id").(string)
		_, _ = r.Context().Value("org_id").(string) // Ensure it exists in context, but not needed here directly

		req.UserID = userID
		// Note: The engine now uses orgID from context for RBAC and scoping.
		// req.UserID is kept as the physical user ID for individual tracking.

		// Headers override
		if p := r.Header.Get("X-Memzent-Provider"); p != "" {
			req.Provider = p
		}
		if m := r.Header.Get("X-Memzent-Model"); m != "" {
			req.Model = m
		}

		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			b := make([]byte, 16)
			rand.Read(b)
			reqID = hex.EncodeToString(b)
		}
		
		w.Header().Set("X-Request-ID", reqID)

		resp, err := memzentEngine.Process(r.Context(), &req)
		if err != nil {
			slog.Error("Engine Processing Error", "error", err, "user", req.UserID)
			w.Header().Set("Content-Type", "application/json")
			errMsg := err.Error()

			if strings.Contains(errMsg, "rate limit exceeded") {
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": errMsg})
				return
			}
			if strings.Contains(errMsg, "payment required") {
				w.WriteHeader(http.StatusPaymentRequired)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": errMsg})
				return
			}
			if strings.Contains(errMsg, "unauthorized") {
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": errMsg})
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Internal Server Error: " + errMsg})
			return
		}

		// Support Server-Sent Events (SSE) streaming if requested
		if req.Stream || r.Header.Get("Accept") == "text/event-stream" {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			w.Header().Set("Transfer-Encoding", "chunked")
			if resp.Cached {
				w.Header().Set("X-Cache", "HIT")
			}

			flusher, ok := w.(http.Flusher)
			if !ok {
				http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
				return
			}

			// Split the final generated response into words to simulate smooth premium SSE stream
			words := strings.Fields(resp.Text)
			for idx, word := range words {
				select {
				case <-r.Context().Done():
					return
				default:
					chunk := word
					if idx < len(words)-1 {
						chunk += " "
					}
					data, _ := json.Marshal(map[string]any{
						"text":       chunk,
						"cached":     resp.Cached,
						"provider":   resp.Provider,
						"request_id": reqID,
					})
					fmt.Fprintf(w, "data: %s\n\n", data)
					flusher.Flush()
					time.Sleep(20 * time.Millisecond) // Premium smooth typewriter delay
				}
			}
			// Write termination packet
			fmt.Fprint(w, "data: [DONE]\n\n")
			flusher.Flush()
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if resp.Cached {
			w.Header().Set("X-Cache", "HIT")
		}
		resp.RequestID = reqID
		_ = json.NewEncoder(w).Encode(resp)
	})))

	// Tools API
	mux.Handle("/v1/tools", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgID, _ := r.Context().Value("org_id").(string)

		if r.Method == http.MethodPost {
			if !auth.HasScope(r.Context(), "tools:write") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{"error": "Forbidden: token lacks tools:write scope"})
				return
			}
			tools.HandleRegisterTool(toolRegistry, rClient, auditLogger)(w, r)
			return
		}

		if !auth.HasScope(r.Context(), "tools:read") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "Forbidden: token lacks tools:read scope"})
			return
		}

		var allTools []tools.ToolWithProvider
		if toolRegistry != nil {
			dbTools, _ := toolRegistry.ListTools(r.Context(), orgID)
			for _, t := range dbTools {
				allTools = append(allTools, tools.ToolToAPI(t))
			}
		}

		if mcpClient != nil {
			mcpTools, _ := mcpClient.ListTools(r.Context())
			for _, t := range mcpTools {
				desc := ""
				if t.Description != nil {
					desc = *t.Description
				}
				allTools = append(allTools, tools.ToolWithProvider{
					ID: t.Name, Name: t.Name, Description: desc, Provider: "Memzent-MCP", Status: "online",
				})
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(allTools)
	})))

	mux.Handle("/v1/tools/sync", middleware(requireScope("tools:write", tools.HandleSyncTools(toolRegistry, rClient, auditLogger))))
	mux.Handle("/v1/tools/register", middleware(requireScope("tools:write", tools.HandleRegisterTool(toolRegistry, rClient, auditLogger))))
	mux.Handle("/v1/tools/status", middleware(requireScope("tools:read", tools.HandleRegistryStatus(toolRegistry))))

	// Audit API
	mux.Handle("/v1/audit", middleware(requireScope("audit:read", func(w http.ResponseWriter, r *http.Request) {
		orgID, _ := r.Context().Value("org_id").(string)

		// Security: GetLatest automatically scopes to the provided orgID
		events, err := auditLogger.GetLatest(orgID, 50)
		if err != nil {
			slog.Error("Failed to fetch persistent audit logs", "error", err)
			// Fallback to in-memory if db fails (optional)
			events = metrics.GlobalAuditBuffer.GetLatest(orgID, 20)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	})))

	// Stats API
	startupTime := time.Now()
	mux.Handle("/v1/stats", middleware(requireScope("audit:read", func(w http.ResponseWriter, r *http.Request) {
		orgID, _ := r.Context().Value("org_id").(string)
		
		var reqs, hits uint64
		if auditLogger != nil {
			// Pull durable stats from Postgres
			reqs, hits = auditLogger.GetCacheStats(orgID)
		} else {
			// Fallback to ephemeral in-memory stats
			reqs, hits = memzentEngine.GetStats(orgID)
		}

		var tokenBalance float64
		if billingLedger != nil {
			tokenBalance, _ = billingLedger.GetBalance(r.Context(), orgID)
		}

		stats := map[string]any{
			"total_requests": reqs,
			"cache_hits":     hits,
			"token_balance":  tokenBalance,
			"uptime_seconds": int(time.Since(startupTime).Seconds()),
			"status":         "online",
			"org_id":         orgID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})))

	mux.Handle("/v1/billing/checkout", middleware(http.HandlerFunc(stripeHandler.CreateCheckoutSession)))

	// Sessions API
	mux.Handle("/v1/sessions", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgID, _ := r.Context().Value("org_id").(string)
		userID, _ := r.Context().Value("user_id").(string)

		if r.Method == http.MethodPost {
			var body struct {
				Title string `json:"title"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)

			sessionID, err := sessionMgr.CreateSession(r.Context(), orgID, userID, body.Title)
			if err != nil {
				slog.Error("Failed to create session", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"session_id": sessionID})
			return
		}

		if r.Method == http.MethodGet {
			sessions, err := sessionMgr.ListSessions(r.Context(), orgID)
			if err != nil {
				slog.Error("Failed to list sessions", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(sessions)
			return
		}

		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})))

	mux.Handle("/v1/sessions/{id}/messages", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		sessionID := r.PathValue("id")
		if sessionID == "" {
			http.Error(w, "Missing session ID", http.StatusBadRequest)
			return
		}

		messages, err := sessionMgr.GetSessionMessages(r.Context(), sessionID, 50)
		if err != nil {
			slog.Error("Failed to fetch session messages", "session_id", sessionID, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	})))

	mux.Handle("/v1/sessions/{id}", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		sessionID := r.PathValue("id")
		if sessionID == "" {
			http.Error(w, "Missing session ID", http.StatusBadRequest)
			return
		}

		err := sessionMgr.DeleteSession(r.Context(), sessionID)
		if err != nil {
			slog.Error("Failed to delete session", "session_id", sessionID, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
	})))

	// Context Analytics API
	mux.Handle("/v1/analytics/context", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		orgID, _ := r.Context().Value("org_id").(string)

		analyticsRes, err := telemetry.GetContextAnalytics(r.Context(), orgID)
		if err != nil {
			slog.Error("Failed to aggregate context analytics", "org_id", orgID, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(analyticsRes)
	})))

	// Providers API (Model Discovery)
	mux.Handle("/v1/providers", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metadata := memzentEngine.GetProviderMetadata()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metadata)
	})))

	// /generate-token is a dev-only convenience endpoint for issuing admin JWTs.
	// It is disabled in production. Set ENABLE_DEV_TOKEN=true to enable locally.
	if os.Getenv("ENABLE_DEV_TOKEN") == "true" {
		mux.HandleFunc("/generate-token", func(w http.ResponseWriter, r *http.Request) {
			secret := cfg.JWTSecret

			token, err := auth.GenerateJWT("admin-01", "admin", secret, time.Hour*24)
			if err != nil {
				slog.Error("Failed to generate JWT", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			resp := map[string]string{"token": token}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		})
		slog.Warn("⚠️  /generate-token endpoint is ENABLED — disable in production (unset ENABLE_DEV_TOKEN)")
	}

	// 8.5. Register SaaS Webhooks (Exclude from JWT Middleware)

	// Separate mux for endpoints that skip JWT middleware (or handle it manually)
	publicMux := http.NewServeMux()
	publicMux.HandleFunc("POST /v1/webhooks/stripe", stripeHandler.HandleWebhook)

	// Checkout session handler is already registered on 'mux' above

	publicMux.Handle("/", auth.UnifiedAuthMiddleware(cfg.JWTSecret, jwksProvider, rbacClient)(metricsMiddleware(commonMiddleware(mux))))

	srv := &http.Server{
		Addr:    cfg.Port,
		Handler: publicMux,
	}

	// 9. Start Server
	go func() {
		slog.Info("Server listening", "addr", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failure", "error", err)
			os.Exit(1)
		}
	}()

	// 10. Graceful Shutdown
	<-ctx.Done()
	slog.Info("Shutting down Memzent Gateway...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Graceful shutdown failed", "error", err)
	}

	slog.Info("Memzent Gateway stopped")
}
