package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"memzent-gateway/internal/auth"
	"memzent-gateway/internal/billing"
	"memzent-gateway/internal/cache"
	"memzent-gateway/internal/config"
	"memzent-gateway/internal/connectors"
	"memzent-gateway/internal/db"
	"memzent-gateway/internal/engine"
	"memzent-gateway/internal/featureflags"
	"memzent-gateway/internal/llm"
	"memzent-gateway/internal/mcp"
	"memzent-gateway/internal/memory"
	"memzent-gateway/internal/metrics"
	"memzent-gateway/internal/notifications"
	"memzent-gateway/internal/offline"
	"memzent-gateway/internal/offline/miners"
	"memzent-gateway/internal/prewarmer"
	"memzent-gateway/internal/router"
	"memzent-gateway/internal/tools"
	"memzent-gateway/internal/workflow"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	prometheusModel "github.com/prometheus/client_model/go"

	_ "memzent-gateway/docs"
)

// getCounterValue reads the current value of a prometheus Counter.
func getCounterValue(c prometheus.Counter) float64 {
	var m prometheusModel.Metric
	if err := c.Write(&m); err != nil {
		return 0
	}
	if m.Counter != nil {
		return *m.Counter.Value
	}
	return 0
}

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

func corsMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	allowAll := len(allowedOrigins) == 1 && allowedOrigins[0] == "*"
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if _, ok := originSet[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-Memzent-Provider, X-Memzent-Model, X-Skip-Cache, X-Org-ID, X-Request-ID")
			w.Header().Set("Access-Control-Expose-Headers", "X-Cache, X-Request-ID, X-RateLimit-Remaining")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
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
	rbacClient, err := auth.NewRBACClient(cfg.PostgresURL, cfg.DevAdminBypass)
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

	// Register: store_memory (Native Implementation)
	coreConnector.RegisterTool("store_memory", func(ctx context.Context, userID string, inputs map[string]interface{}) (string, error) {
		slog.Info("Executing CORE tool: store_memory", "user_id", userID)
		fact, ok := inputs["fact"].(string)
		if !ok || fact == "" {
			return "", fmt.Errorf("missing required parameter: fact")
		}
		orgID, _ := ctx.Value("org_id").(string)
		success, err := rClient.StoreMemory(ctx, fact, orgID, userID)
		if err != nil {
			return "", fmt.Errorf("failed to store memory: %w", err)
		}
		if !success {
			return "", fmt.Errorf("failed to store memory (router rejected)")
		}
		return fmt.Sprintf("Memory stored successfully: \"%s\"", fact), nil
	})

	// Register: recall_memory (Native Implementation)
	coreConnector.RegisterTool("recall_memory", func(ctx context.Context, userID string, inputs map[string]interface{}) (string, error) {
		slog.Info("Executing CORE tool: recall_memory", "user_id", userID)
		query, ok := inputs["query"].(string)
		if !ok || query == "" {
			return "", fmt.Errorf("missing required parameter: query")
		}
		orgID, _ := ctx.Value("org_id").(string)
		hits, err := rClient.QueryMemory(ctx, query, orgID, userID, 0.65)
		if err != nil {
			return "", fmt.Errorf("failed to query memory: %w", err)
		}
		if len(hits) == 0 {
			return "No relevant memories found in long-term storage.", nil
		}
		var sb strings.Builder
		sb.WriteString("Retrieved memories:\n")
		for _, hit := range hits {
			sb.WriteString(fmt.Sprintf("- %s (relevance: %.2f)\n", hit.Fact, hit.RelevanceScore))
		}
		return sb.String(), nil
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

	// 7b. Initialize Webhook Notification Dispatcher
	var webhookRegistry *notifications.Registry
	var webhookDispatcher *notifications.Dispatcher
	if rbacClient != nil {
		webhookRegistry = notifications.NewRegistry(rbacClient.GetDB())
		webhookDispatcher = notifications.NewDispatcher(webhookRegistry, 4)
		slog.Info("Webhook Notification Pipeline initialized")
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

	// Attach webhook event emitter to engine
	if webhookDispatcher != nil {
		memzentEngine.SetEventEmitter(webhookDispatcher)
	}

	// Load feature flags from environment
	flags := featureflags.Load()
	slog.Info("🚩 Feature flags loaded",
		"l1b_cache", flags.L1bCache,
		"offline_plane", flags.OfflinePlane,
		"workflow_engine", flags.WorkflowEngine,
		"entity_metrics", flags.EntityMetrics,
	)

	// 8.0b Initialize Offline Learning Plane (E3)
	var requestMiner *miners.RequestMiner
	var cacheMiner *miners.CacheMiner
	var workflowMiner *miners.WorkflowMiner
	var streamPlane *offline.StreamPlane
	var channelPlane *offline.Plane
	if flags.OfflinePlane {
		requestMiner = miners.NewRequestMiner(50)
		cacheMiner = miners.NewCacheMiner(10, 100)
		workflowMiner = miners.NewWorkflowMiner(100, 0.90)

		if flags.OfflineStreams && vCache != nil {
			// Valkey Streams mode — crash-durable, multi-instance
			streamCfg := offline.DefaultStreamConfig("")
			streamPlane = offline.NewStreamPlane(vCache, streamCfg, requestMiner, cacheMiner, workflowMiner)
			streamPlane.Start(ctx)
			memzentEngine.SetOfflinePlane(streamPlane)
			slog.Info("🌊 Offline Learning Plane started (Valkey Streams mode)")
		} else {
			// In-memory channel mode (single instance, fast)
			channelPlane = offline.NewPlane(
				offline.DefaultConfig(),
				requestMiner, cacheMiner, workflowMiner,
			)
			channelPlane.Start(ctx)
			memzentEngine.SetOfflinePlane(channelPlane)
			slog.Info("🧠 Offline Learning Plane started (channel mode)")
		}
	}

	// 8.0c Initialize Workflow Registry (E4)
	var workflowRegistry *workflow.Registry
	if flags.WorkflowEngine && rbacClient != nil && rbacClient.GetDB() != nil {
		workflowRegistry = workflow.NewRegistry(rbacClient.GetDB())
		workflowRegistry.StartDemotionLoop(ctx, 1*time.Hour)
		workflow.NewSimulator(rbacClient.GetDB()).StartSimulationLoop(ctx, 1*time.Hour)
		memzentEngine.SetWorkflowRegistry(workflow.NewEngineAdapter(workflowRegistry))
		slog.Info("📋 Workflow Registry initialized with hourly demotion and simulation checks")
	}

	// 8.0d O3 → Registry Bridge: auto-populate workflow candidates from miner output
	if workflowMiner != nil && workflowRegistry != nil {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case output := <-workflowMiner.Output():
					for _, seq := range output.PromotionReady {
						_, err := workflowRegistry.UpsertCandidate(
							ctx, seq.OrgID, seq.Pattern,
							int(seq.Frequency), seq.Tools, nil,
						)
						if err != nil {
							slog.Warn("O3→Registry bridge: failed to upsert candidate",
								"pattern", seq.Pattern, "org_id", seq.OrgID, "error", err)
						} else {
							slog.Info("🔗 O3→Registry: workflow candidate upserted",
								"pattern", seq.Pattern, "frequency", seq.Frequency,
								"success_rate", seq.SuccessRate)
						}
					}
				}
			}
		}()
		slog.Info("🔗 O3→Registry bridge started")
	}

	// 8.0e Speculative Pre-Warmer: O2 → Valkey (L1b pre-warming)
	var preWarm *prewarmer.PreWarmer
	if cacheMiner != nil && vCache != nil && flags.L1bCache {
		// Response lookup callback — queries the persistent Postgres cache
		var responseLookup prewarmer.ResponseLookup
		if rbacClient != nil && rbacClient.GetDB() != nil {
			pgDB := rbacClient.GetDB()
			responseLookup = func(lookupCtx context.Context, key string) (string, error) {
				var resp string
				err := pgDB.QueryRowContext(lookupCtx,
					"SELECT response FROM cache_entries WHERE cache_key = $1 AND expires_at > now() LIMIT 1", key,
				).Scan(&resp)
				if err != nil {
					return "", err
				}
				return resp, nil
			}
		}
		preWarm = prewarmer.New(vCache, cacheMiner, prewarmer.DefaultConfig(), responseLookup)
		preWarm.Start(ctx)
		slog.Info("🔥 Speculative Pre-Warmer started (O2 → L1b)")
	}

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
		if strings.EqualFold(r.Header.Get("X-Skip-Cache"), "true") {
			req.SkipCache = true
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

	// Tool CRUD by ID: GET, PUT, DELETE /v1/tools/{toolId}
	mux.Handle("/v1/tools/", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if !auth.HasScope(r.Context(), "tools:read") {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			tools.HandleGetTool(toolRegistry)(w, r)
		case http.MethodPut:
			if !auth.HasScope(r.Context(), "tools:write") {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			tools.HandleUpdateTool(toolRegistry, rClient, auditLogger)(w, r)
		case http.MethodDelete:
			if !auth.HasScope(r.Context(), "tools:write") {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			tools.HandleDisableTool(toolRegistry)(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})))

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

	// Webhooks API (Phase 7: Notification Pipeline)
	if webhookRegistry != nil {
		mux.Handle("/v1/webhooks", middleware(requireScope("tools:write", notifications.HandleWebhooks(webhookRegistry))))
		mux.Handle("/v1/webhooks/", middleware(requireScope("tools:write", notifications.HandleWebhookByID(webhookRegistry))))
		mux.Handle("/v1/webhooks/event-types", middleware(requireScope("tools:read", notifications.HandleEventTypes())))
	}

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

		// Build list of active provider names
		activeProviders := make([]string, 0, len(providers))
		for name := range providers {
			activeProviders = append(activeProviders, name)
		}

		stats := map[string]any{
			"total_requests":   reqs,
			"cache_hits":       hits,
			"token_balance":    tokenBalance,
			"uptime_seconds":   int(time.Since(startupTime).Seconds()),
			"status":           "online",
			"org_id":           orgID,
			"active_providers": activeProviders,
			"default_provider": defaultProvider,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})))

	mux.Handle("/v1/billing/checkout", middleware(http.HandlerFunc(stripeHandler.CreateCheckoutSession)))

	// Budget & Spend API (for planning tools / external integrations)
	mux.Handle("/v1/billing/budget", middleware(requireScope("audit:read", func(w http.ResponseWriter, r *http.Request) {
		orgID, _ := r.Context().Value("org_id").(string)

		if billingLedger == nil {
			http.Error(w, "Billing not configured", http.StatusServiceUnavailable)
			return
		}

		status, err := billingLedger.GetBudgetStatus(r.Context(), orgID)
		if err != nil {
			slog.Error("Failed to get budget status", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Include spend limits
		if spendLimits, err := billingLedger.CheckSpendLimits(r.Context(), orgID); err == nil && spendLimits != nil {
			status.SpendLimit = spendLimits.DailyLimit
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})))

	mux.Handle("/v1/billing/spend-timeseries", middleware(requireScope("audit:read", func(w http.ResponseWriter, r *http.Request) {
		orgID, _ := r.Context().Value("org_id").(string)

		if billingLedger == nil {
			http.Error(w, "Billing not configured", http.StatusServiceUnavailable)
			return
		}

		days := 30
		if d := r.URL.Query().Get("days"); d != "" {
			if parsed, err := fmt.Sscanf(d, "%d", &days); parsed == 0 || err != nil {
				days = 30
			}
		}

		data, err := billingLedger.GetSpendTimeseries(r.Context(), orgID, days)
		if err != nil {
			slog.Error("Failed to get spend timeseries", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})))

	mux.Handle("/v1/billing/spend-limits", middleware(requireScope("audit:read", func(w http.ResponseWriter, r *http.Request) {
		orgID, _ := r.Context().Value("org_id").(string)

		if billingLedger == nil {
			http.Error(w, "Billing not configured", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			status, err := billingLedger.CheckSpendLimits(r.Context(), orgID)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if status == nil {
				status = &billing.SpendLimitStatus{}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(status)

		case http.MethodPut:
			var req struct {
				DailyLimit        *float64 `json:"daily_limit"`
				MonthlyLimit      *float64 `json:"monthly_limit"`
				DailyTokenLimit   *int64   `json:"daily_token_limit"`
				MonthlyTokenLimit *int64   `json:"monthly_token_limit"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			if err := billingLedger.SetSpendLimits(r.Context(), orgID, req.DailyLimit, req.MonthlyLimit, req.DailyTokenLimit, req.MonthlyTokenLimit); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "updated"})

		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})))

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
			json.NewEncoder(w).Encode(map[string]string{"id": sessionID, "session_id": sessionID})
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

		// Offline Learning Plane Stats API (E3)
		mux.Handle("/v1/offline/stats", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
			if channelPlane == nil && streamPlane == nil {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{"enabled": false})
				return
			}
			var planeStats map[string]uint64
			mode := "channel"
			if streamPlane != nil {
				planeStats = streamPlane.Stats()
				mode = "valkey_streams"
			} else {
				planeStats = channelPlane.Stats()
			}
			stats := map[string]interface{}{
				"enabled":             true,
				"mode":                mode,
				"plane":               planeStats,
				"hot_patterns":        requestMiner.GetHotPatterns(),
				"cache_misses":        cacheMiner.GetTopMisses(),
				"workflow_sequences":  workflowMiner.GetDetectedSequences(),
				"prediction_accuracy": cacheMiner.PredictionAccuracy(),
			}
			if preWarm != nil {
				stats["pre_warmer"] = preWarm.Stats()
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(stats)
		})))

		// Workflow Registry API (E4)
		mux.Handle("/v1/workflows", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if workflowRegistry == nil {
				http.Error(w, `{"error":"workflow registry not initialized"}`, http.StatusServiceUnavailable)
				return
			}
			orgID, _ := r.Context().Value("org_id").(string)

			switch r.Method {
			case http.MethodGet:
				status := r.URL.Query().Get("status")
				candidates, err := workflowRegistry.ListCandidates(r.Context(), orgID, status)
				if err != nil {
					slog.Error("Failed to list workflows", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
					return
				}
				if candidates == nil {
					candidates = []workflow.Candidate{}
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(candidates)
			default:
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
		})))

		mux.Handle("/v1/workflows/approve", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if workflowRegistry == nil {
				http.Error(w, `{"error":"workflow registry not initialized"}`, http.StatusServiceUnavailable)
				return
			}
			if r.Method != http.MethodPut {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
			var body struct {
				ID string `json:"id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == "" {
				http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
				return
			}
			userID, _ := r.Context().Value("user_id").(string)
			if err := workflowRegistry.Approve(r.Context(), body.ID, userID); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			// Auto-activate after approval
			_ = workflowRegistry.Activate(r.Context(), body.ID)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "approved_and_activated"})
		})))

		mux.Handle("/v1/workflows/reject", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if workflowRegistry == nil {
				http.Error(w, `{"error":"workflow registry not initialized"}`, http.StatusServiceUnavailable)
				return
			}
			if r.Method != http.MethodPut {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
			var body struct {
				ID string `json:"id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == "" {
				http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
				return
			}
			userID, _ := r.Context().Value("user_id").(string)
			if err := workflowRegistry.Reject(r.Context(), body.ID, userID); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "rejected"})
		})))

		// Entity Quality Metrics API (E5)
		mux.Handle("/v1/metrics/entities", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
			regexSuccess := metrics.EntityRegexSuccess
			regexFailure := metrics.EntityRegexFailure
			mismatch := metrics.EntityMismatchTotal
			llmUsage := metrics.EntityLLMUsage
			gpuAvoided := metrics.GPUAvoidanceTotal
			gpuInvoked := metrics.GPUInvocationTotal

			// Read counters via prometheus
			rsVal := getCounterValue(regexSuccess)
			rfVal := getCounterValue(regexFailure)
			mmVal := getCounterValue(mismatch)
			llmVal := getCounterValue(llmUsage)
			gpuAvVal := getCounterValue(gpuAvoided)
			gpuInvVal := getCounterValue(gpuInvoked)

			total := rsVal + rfVal
			var regexSuccessRate float64
			if total > 0 {
				regexSuccessRate = rsVal / total
			}
			var gpuAvoidanceRate float64
			gpuTotal := gpuAvVal + gpuInvVal
			if gpuTotal > 0 {
				gpuAvoidanceRate = gpuAvVal / gpuTotal
			}

			stats := map[string]interface{}{
				"regex_success":      rsVal,
				"regex_failure":      rfVal,
				"regex_success_rate": regexSuccessRate,
				"entity_mismatch":    mmVal,
				"llm_entity_usage":   llmVal,
				"gpu_avoidance_rate": gpuAvoidanceRate,
				"gpu_avoided":        gpuAvVal,
				"gpu_invoked":        gpuInvVal,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(stats)
		})))

		// Feature flags status endpoint (E1-E5 layer toggles)
		mux.Handle("/v1/features", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
			f := featureflags.Get()
			response := map[string]interface{}{
				"l1b_cache":       f.L1bCache,
				"offline_plane":   f.OfflinePlane,
				"offline_streams": f.OfflineStreams,
				"workflow_engine": f.WorkflowEngine,
				"entity_metrics":  f.EntityMetrics,
				"env_vars": map[string]string{
					"MEMZENT_L1B_ENABLED":            "controls L1b entity-keyed hot path cache",
					"MEMZENT_OFFLINE_ENABLED":        "controls offline learning plane (O1/O2/O3 miners)",
					"MEMZENT_OFFLINE_STREAMS":        "use Valkey Streams instead of in-memory channels (requires Valkey)",
					"MEMZENT_WORKFLOW_ENABLED":       "controls workflow registry + engine shortcut",
					"MEMZENT_ENTITY_METRICS_ENABLED": "controls entity quality + GPU avoidance counters",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		})))
		json.NewEncoder(w).Encode(metadata)
	})))

	// Full model listing endpoint — returns all discovered models per provider
	mux.Handle("/v1/models", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metadata := memzentEngine.GetProviderMetadata()
		type ModelEntry struct {
			Provider string   `json:"provider"`
			Models   []string `json:"models"`
			Default  string   `json:"default_model"`
		}
		result := make([]ModelEntry, 0, len(metadata))
		for _, m := range metadata {
			result = append(result, ModelEntry{
				Provider: m.Name,
				Models:   m.SupportedModels,
				Default:  m.DefaultModel,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})))

	// Similarity Threshold API: per-org configurable semantic precision
	mux.Handle("/v1/settings/threshold", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgID, _ := r.Context().Value("org_id").(string)
		if orgID == "" {
			http.Error(w, `{"error":"org_id required"}`, http.StatusBadRequest)
			return
		}

		switch r.Method {
		case "GET":
			var threshold float64
			err := rbacClient.GetDB().QueryRowContext(r.Context(),
				"SELECT COALESCE(similarity_threshold, 0.88) FROM organizations WHERE id = $1", orgID).Scan(&threshold)
			if err != nil {
				threshold = 0.88
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]float64{"similarity_threshold": threshold})

		case "PUT", "PATCH":
			var body struct {
				Threshold float64 `json:"similarity_threshold"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
				return
			}
			if body.Threshold < 0.5 || body.Threshold > 1.0 {
				http.Error(w, `{"error":"threshold must be between 0.5 and 1.0"}`, http.StatusBadRequest)
				return
			}
			_, err := rbacClient.GetDB().ExecContext(r.Context(),
				"UPDATE organizations SET similarity_threshold = $1 WHERE id = $2", body.Threshold, orgID)
			if err != nil {
				http.Error(w, `{"error":"failed to update threshold"}`, http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "similarity_threshold": body.Threshold})

		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})))

	// Cache Flush API: invalidate stale canonical/semantic cache entries for an org
	mux.Handle("/v1/cache/flush", middleware(requireScope("tools:write", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		orgID, _ := r.Context().Value("org_id").(string)
		if orgID == "" {
			http.Error(w, `{"error":"org_id required"}`, http.StatusBadRequest)
			return
		}

		if vCache == nil {
			http.Error(w, `{"error":"cache not initialized"}`, http.StatusServiceUnavailable)
			return
		}

		// Flush all cache keys for this org (literal, canonical, semantic)
		pattern := fmt.Sprintf("org:%s:*", orgID)
		deleted, err := vCache.FlushByPattern(r.Context(), pattern)
		if err != nil {
			slog.Error("Cache flush failed", "error", err, "org_id", orgID)
			http.Error(w, `{"error":"cache flush failed"}`, http.StatusInternalServerError)
			return
		}

		// Purge persistent DB cache for this org
		var dbDeleted int64
		if rbacClient != nil && rbacClient.GetDB() != nil {
			result, dbErr := rbacClient.GetDB().ExecContext(r.Context(),
				"DELETE FROM persistent_cache WHERE org_id = $1", orgID)
			if dbErr != nil {
				slog.Warn("Persistent cache flush failed (non-fatal)", "error", dbErr, "org_id", orgID)
			} else if result != nil {
				dbDeleted, _ = result.RowsAffected()
			}
		}

		// Flush Qdrant prompts_collection for this org (semantic Stage 2 cache)
		if rClient != nil {
			if flushErr := rClient.FlushPromptCache(r.Context(), orgID); flushErr != nil {
				slog.Warn("Qdrant prompt cache flush failed (non-fatal)", "error", flushErr, "org_id", orgID)
			}
		}

		slog.Info("🗑️ Cache flushed", "org_id", orgID, "valkey_keys_deleted", deleted, "db_rows_deleted", dbDeleted)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":             true,
			"valkey_keys_deleted": deleted,
			"db_rows_deleted":     dbDeleted,
			"org_id":              orgID,
		})
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

	publicMux.Handle("/", auth.UnifiedAuthMiddleware(cfg.JWTSecret, jwksProvider, rbacClient)(metricsMiddleware(corsMiddleware(cfg.AllowedOrigins)(mux))))

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

	// Stop offline learning plane (drain events, flush miners)
	if channelPlane != nil {
		channelPlane.Stop()
	}
	if streamPlane != nil {
		streamPlane.Stop()
	}
	if preWarm != nil {
		preWarm.Stop()
	}

	// Stop webhook dispatcher (drain queue)
	if webhookDispatcher != nil {
		webhookDispatcher.Stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Graceful shutdown failed", "error", err)
	}

	slog.Info("Memzent Gateway stopped")
}
