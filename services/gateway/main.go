package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aura-gateway/internal/auth"
	"aura-gateway/internal/billing"
	"aura-gateway/internal/cache"
	"aura-gateway/internal/config"
	"aura-gateway/internal/connectors"
	"aura-gateway/internal/engine"
	"aura-gateway/internal/llm"
	"aura-gateway/internal/mcp"
	"aura-gateway/internal/metrics"
	"aura-gateway/internal/router"
	"aura-gateway/internal/tools"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Aura-Provider, X-Aura-Model, X-Skip-Cache, X-Org-ID")

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

	slog.Info("Starting Aura Gateway", "port", cfg.Port, "env", cfg.Environment)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 3. Initialize Cache
	vCache, err := cache.NewAuraCache(ctx, cfg.ValkeyURL)
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
	}

	// 5.1 Initialize JWKS Provider (Dynamic Auth discovery)
	var jwksProvider *auth.JWKSProvider
	if cfg.JWKSURL != "" {
		jwksProvider = auth.NewJWKSProvider(cfg.JWKSURL, cfg.SupabaseKey)
		slog.Info("JWKS Provider initialized", "url", cfg.JWKSURL)
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
		return "Mock Database Trace: Successfully indexed 1,241 cluster metrics via Aura Core (Native Connector).", nil
	})

	// Register: aura_search (Native Implementation)
	coreConnector.RegisterTool("aura_search", func(ctx context.Context, userID string, inputs map[string]interface{}) (string, error) {
		slog.Info("Executing CORE tool: aura_search", "user_id", userID)
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

	// 8. Initialize Engine
	auraEngine := engine.NewAuraEngine(vCache, rClient, rbacClient, providers, defaultProvider, mcpClient, toolRegistry, connRegistry, cfg.ToolRelevanceThreshold, cfg.LLMCacheTTL)

	// 8.1 Initialize Stripe Handler (SaaS Billing)
	stripeHandler := billing.NewStripeHandler(
		rbacClient.GetDB(),
		os.Getenv("STRIPE_WEBHOOK_SECRET"),
		os.Getenv("STRIPE_PRO_ID"),
		os.Getenv("STRIPE_BIZ_ID"),
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

	// Chat API
	mux.Handle("/v1/chat", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		orgID, _ := r.Context().Value("org_id").(string)
		
		req.UserID = userID
		if orgID != "" {
			req.UserID = orgID 
		}

		// Headers override
		if p := r.Header.Get("X-Aura-Provider"); p != "" { req.Provider = p }
		if m := r.Header.Get("X-Aura-Model"); m != "" { req.Model = m }

		resp, err := auraEngine.Process(r.Context(), &req)
		if err != nil {
			slog.Error("Engine Processing Error", "error", err, "user", req.UserID)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if resp.Cached { w.Header().Set("X-Cache", "HIT") }
		json.NewEncoder(w).Encode(resp)
	})))

	// Tools API
	mux.Handle("/v1/tools", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgID, _ := r.Context().Value("org_id").(string)
		
		if r.Method == http.MethodPost {
			tools.HandleRegisterTool(toolRegistry)(w, r)
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
				if t.Description != nil { desc = *t.Description }
				allTools = append(allTools, tools.ToolWithProvider{
					ID: t.Name, Name: t.Name, Description: desc, Provider: "Aura-MCP", Status: "online",
				})
			}
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(allTools)
	})))

	mux.Handle("/v1/tools/sync", middleware(http.HandlerFunc(tools.HandleSyncTools(toolRegistry))))
	mux.Handle("/v1/tools/register", middleware(http.HandlerFunc(tools.HandleRegisterTool(toolRegistry))))

	// Audit API
	mux.Handle("/v1/audit", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgID, _ := r.Context().Value("org_id").(string)
		events := metrics.GlobalAuditBuffer.GetLatest(orgID, 20)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	})))

	// Stats API
	startupTime := time.Now()
	mux.Handle("/v1/stats", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgID, _ := r.Context().Value("org_id").(string)
		stats := map[string]any{
			"total_requests": auraEngine.TotalRequests.Load(),
			"cache_hits":     auraEngine.CacheHits.Load(),
			"uptime_seconds": int(time.Since(startupTime).Seconds()),
			"status":         "online",
			"org_id":         orgID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})))

	mux.Handle("/v1/billing/checkout", middleware(http.HandlerFunc(stripeHandler.CreateCheckoutSession)))

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
	slog.Info("Shutting down Aura Gateway...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Graceful shutdown failed", "error", err)
	}

	slog.Info("Aura Gateway stopped")
}
