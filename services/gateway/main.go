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
	"aura-gateway/internal/cache"
	"aura-gateway/internal/config"
	"aura-gateway/internal/engine"
	"aura-gateway/internal/llm"
	"aura-gateway/internal/mcp"
	"aura-gateway/internal/metrics"
	"aura-gateway/internal/router"

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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Aura-Provider, X-Aura-Model, X-Skip-Cache")

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

	// 8. Initialize Engine
	auraEngine := engine.NewAuraEngine(vCache, rClient, rbacClient, providers, defaultProvider, mcpClient, cfg.ToolRelevanceThreshold, cfg.LLMCacheTTL)

	mux := http.NewServeMux()

	// Metrics API
	mux.Handle("/metrics", promhttp.Handler())

	// Health Checks (Enterprise standard)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		status := map[string]string{"status": "healthy", "service": "aura-gateway"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		status := map[string]string{"status": "ready", "time": time.Now().Format(time.RFC3339)}
		
		if err := vCache.Ping(r.Context()); err != nil {
			status["status"] = "not_ready"
			status["valkey"] = "unreachable"
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			status["valkey"] = "online"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	// Chat API (v1)
	mux.HandleFunc("/v1/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var req engine.PromptRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid Request Body", http.StatusBadRequest)
			return
		}

		if req.Prompt == "" {
			http.Error(w, "Missing 'prompt' field", http.StatusBadRequest)
			return
		}
		if req.UserID == "" {
			req.UserID = "anonymous"
		}

		// Headers override body for cache/provider/model control
		if p := r.Header.Get("X-Aura-Provider"); p != "" && req.Provider == "" {
			req.Provider = p
		}
		if m := r.Header.Get("X-Aura-Model"); m != "" && req.Model == "" {
			req.Model = m
		}
		if r.Header.Get("X-Skip-Cache") == "true" || r.Header.Get("X-Skip-Cache") == "1" {
			req.SkipCache = true
		}

		resp, err := auraEngine.Process(r.Context(), &req)
		if err != nil {
			slog.Error("Engine Processing Error", "error", err, "user_id", req.UserID)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if resp.Cached {
			w.Header().Set("X-Cache", "HIT")
		} else {
			w.Header().Set("X-Cache", "MISS")
		}
		w.Header().Set("X-Aura-Provider", resp.Provider)
		json.NewEncoder(w).Encode(resp)
	})

	// Tools API (v1)
	mux.HandleFunc("/v1/tools", func(w http.ResponseWriter, r *http.Request) {
		allTools := []map[string]any{
			{"id": "aura_search", "name": "Neural Semantic Search", "provider": "Aura-Core", "status": "online"},
		}

		if mcpClient != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			mcpTools, err := mcpClient.ListTools(ctx)
			if err == nil {
				for _, t := range mcpTools {
					allTools = append(allTools, map[string]any{
						"id":       t.Name,
						"name":     t.Name,
						"provider": "Aura-MCP",
						"status":   "online",
						"desc":     t.Description,
					})
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(allTools)
	})

	// Stats API (v1)
	startupTime := time.Now()
	mux.HandleFunc("/v1/stats", func(w http.ResponseWriter, r *http.Request) {
		uptime := time.Since(startupTime)
		
		stats := map[string]any{
			"total_requests": auraEngine.TotalRequests.Load(),
			"cache_hits":     auraEngine.CacheHits.Load(),
			"uptime_seconds": int(uptime.Seconds()),
			"status":         "online",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	srv := &http.Server{
		Addr:    cfg.Port,
		Handler: auth.JWTMiddleware(cfg.JWTSecret)(metricsMiddleware(commonMiddleware(mux))),
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
