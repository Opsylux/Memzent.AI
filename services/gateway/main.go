package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"aura-gateway/internal/auth"
	"aura-gateway/internal/engine"
	"aura-gateway/internal/llm"
	"aura-gateway/internal/mcp"
	"aura-gateway/internal/router"

	"github.com/valkey-io/valkey-go"
)

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Development friendly
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	ctx := context.Background()

	// 1. Initialize Valkey (Cache)
	valkeyAddr := os.Getenv("VALKEY_URL")
	if valkeyAddr == "" {
		valkeyAddr = "valkey:6379"
	}
	vClient, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{valkeyAddr}})
	if err != nil {
		log.Fatalf("❌ Failed to connect to Valkey: %v", err)
	}
	defer vClient.Close()
	fmt.Println("✅ Aura Gateway: Connected to Valkey")

	// 2. Initialize Rust Router Client (gRPC)
	routerAddr := os.Getenv("ROUTER_URL")
	if routerAddr == "" {
		routerAddr = "router:50051"
	}
	rClient, err := router.NewRouterClient(routerAddr)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Rust Router: %v", err)
	}
	defer rClient.Close()
	fmt.Println("✅ Aura Gateway: Connected to Rust Router")

	// 3. Initialize Postgres RBAC
	pgAddr := os.Getenv("POSTGRES_URL")
	if pgAddr == "" {
		pgAddr = "postgres://user:password@postgres:5432/aura_db?sslmode=disable"
	}
	rbacClient, err := auth.NewRBACClient(pgAddr)
	if err != nil {
		log.Printf("⚠️ Failed to connect to Postgres RBAC: %v. Starting without it.", err)
	} else {
		defer rbacClient.Close()
		fmt.Println("✅ Aura Gateway: Connected to Postgres RBAC")
	}

	// 4. Initialize MCP Client
	mcpClient, err := mcp.NewMCPClient()
	if err != nil {
		log.Printf("⚠️ Failed to initialize MCP Client: %v", err)
	}

	// 5. Initialize LLM Provider & Engine
	var llmProvider llm.Provider
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey != "" {
		llmProvider = llm.NewAnthropicProvider(anthropicKey, "")
		fmt.Println("✅ Aura Gateway: Using Anthropic LLM Provider")
	} else {
		llmProvider = llm.NewMockProvider()
		fmt.Println("⚠️ Aura Gateway: Using Mock LLM Provider (No API Key found)")
	}
	auraEngine := engine.NewAuraEngine(vClient, rClient, rbacClient, llmProvider, mcpClient)

	// 5. Handlers
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		err := vClient.Do(ctx, vClient.B().Ping().Build()).Error()
		if err != nil {
			http.Error(w, "Valkey Unreachable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "healthy"}`)
	})

	http.HandleFunc("/v1/chat", func(w http.ResponseWriter, r *http.Request) {
		prompt := r.URL.Query().Get("prompt")
		if prompt == "" {
			http.Error(w, "Missing 'prompt' parameter", http.StatusBadRequest)
			return
		}

		// Use the Orchestration Engine
		resp, err := auraEngine.Process(ctx, &engine.PromptRequest{
			UserID: "solo-user",
			Prompt: prompt,
		})

		if err != nil {
			log.Printf("Engine Error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if resp.Cached {
			w.Header().Set("X-Cache", "HIT")
		} else {
			w.Header().Set("X-Cache", "MISS")
		}
		json.NewEncoder(w).Encode(resp)
	})

	http.HandleFunc("/v1/tools", func(w http.ResponseWriter, r *http.Request) {
		// This endpoint is used by the Dashboard to list available tools

		// 1. Start with high-level known system capabilities
		allTools := []map[string]any{
			{"id": "aura_search", "name": "Neural Semantic Search", "provider": "Aura-Core", "status": "online"},
		}

		// 2. Try to fetch dynamic tools from MCP server
		if mcpClient != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()

			// Ensure initialized
			_ = mcpClient.Initialize(ctx)

			mcpTools, err := mcpClient.ListTools(ctx)
			if err == nil {
				for _, t := range mcpTools {
					allTools = append(allTools, map[string]any{
						"id":       t.Name, // Using Name as ID for UI simplicity
						"name":     t.Name,
						"provider": "Aura-MCP",
						"status":   "online",
						"desc":     t.Description,
					})
				}
			} else {
				log.Printf("⚠️ Dashboard: Failed to list MCP tools: %v", err)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(allTools)
	})

	port := ":8080"
	fmt.Printf("🚀 Aura Gateway active on %s\n", port)
	if err := http.ListenAndServe(port, commonMiddleware(http.DefaultServeMux)); err != nil {
		log.Fatal(err)
	}
}
