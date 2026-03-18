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
	"aura-gateway/internal/mcp"
	"aura-gateway/internal/router"

	"github.com/valkey-io/valkey-go"
)

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
	fmt.Println("✅ Aura Gateway: Connected to Valkey (Native Go)")

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

	// 2.5 Initialize Postgres RBAC
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

	// 3. Health Check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		err := vClient.Do(ctx, vClient.B().Ping().Build()).Error()
		if err != nil {
			http.Error(w, "Valkey Unreachable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "healthy"}`)
	})

	// 4. Aura Chat & Routing Endpoint
	http.HandleFunc("/v1/chat", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("prompt")
		if query == "" {
			http.Error(w, "Missing 'prompt' parameter", http.StatusBadRequest)
			return
		}

		// A. Semantic Cache Check
		resp := vClient.Do(ctx, vClient.B().Get().Key(query).Build())
		if cached, err := resp.ToString(); err == nil {
			w.Header().Set("X-Cache", "HIT")
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(cached))
			return
		}

		// B. Check Postgres for Allowed Tools
		userID := "solo-user"
		var allowedTools []string
		if rbacClient != nil {
			var dbErr error
			allowedTools, dbErr = rbacClient.GetAllowedTools(userID)
			if dbErr != nil {
				log.Printf("RBAC Error for %s: %v", userID, dbErr)
			}
		}

		// C. Ask Rust Router for Tool Selection
		gCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		tools, err := rClient.GetBestTools(gCtx, query, userID, allowedTools)
		if err != nil {
			log.Printf("Router Error: %v", err)
			http.Error(w, "Semantic Router error", http.StatusBadGateway)
			return
		}

		// D. Simulate Execution & Output Compression
		var compressedOutput string
		if len(tools) > 0 {
			// Mocking raw tool execution output
			rawOutput := "Executing tool...\nConnecting to DB...\nError: Connection timeout on port 5432\nRetrying...\nFailed."
			compressedOutput = mcp.CompressToolOutput(rawOutput, query)
		}

		// E. Prepare Response and Cache it
		responseJSON, _ := json.Marshal(map[string]interface{}{
			"prompt":           query,
			"tools":            tools,
			"execution_result": compressedOutput,
			"ts":               time.Now().Unix(),
		})

		// Set in Valkey with a 1-hour TTL
		_ = vClient.Do(ctx, vClient.B().Set().Key(query).Value(string(responseJSON)).Ex(3600).Build())

		w.Header().Set("X-Cache", "MISS")
		w.Header().Set("Content-Type", "application/json")
		w.Write(responseJSON)
	})

	port := ":8080"
	fmt.Printf("🚀 Aura Gateway active on %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
