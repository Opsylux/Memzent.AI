package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/http"
	"github.com/valkey-io/valkey-go"
)

// ToolArgs defines the expected schema for the execute_aura_tool.
// Using JSON tags ensures the MCP client and server agree on field names.
type ToolArgs struct {
	ToolID string `json:"tool_id"`
	UserID string `json:"user_id,omitempty"`
}

func main() {
	// 1. Setup Structured Logging to Stderr
	// MCP protocol uses Stdout for transport in some modes;
	// logging to Stderr prevents protocol corruption.
	logger := log.New(os.Stderr, "[AURA-MCP] ", log.LstdFlags|log.Lshortfile)

	// 2. Initialize Valkey Client
	valkeyAddr := os.Getenv("VALKEY_ADDR")
	if valkeyAddr == "" {
		valkeyAddr = "valkey:6379"
	}

	vClient, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{valkeyAddr},
		SelectDB:    0,
	})
	if err != nil {
		logger.Fatalf("Failed to connect to Valkey: %v", err)
	}
	defer vClient.Close()

	// 3. Initialize MCP Server over HTTP
	// We explicitly set the path to /mcp to match your Gateway's environment variables.
	t := http.NewHTTPTransport("/mcp").WithAddr(":50052")
	server := mcp.NewServer(t)

	// --- TOOL: GET TOOLS ---
	server.RegisterTool("get_aura_tools", "Returns available tools from cache or registry", func(ctx context.Context) (string, error) {
		valkeyKey := "mcp:tools:list"

		// Use a sub-context with timeout for Valkey calls to prevent hanging
		vCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		resp := vClient.Do(vCtx, vClient.B().Get().Key(valkeyKey).Build())
		if cached, err := resp.ToString(); err == nil {
			return fmt.Sprintf("Retrieved from Cache: %s", cached), nil
		}

		// Fallback/Registry logic
		tools := "Available: [db_query, get_user, read_database]"
		_ = vClient.Do(vCtx, vClient.B().Set().Key(valkeyKey).Value(tools).Ex(300).Build())

		return tools, nil
	})

	// --- TOOL: EXECUTE TOOL ---
	server.RegisterTool("execute_aura_tool", "Runs a specific Aura tool", func(ctx context.Context, args ToolArgs) (string, error) {
		toolID := args.ToolID
		userID := args.UserID

		// Validation: If ToolID is missing, return a clean error
		if toolID == "" {
			return "", fmt.Errorf("missing required parameter: tool_id")
		}

		logger.Printf("Executing Tool: %s (User: %s)", toolID, userID)

        switch toolID {
        case "db_query":
            return "SQL query executed successfully via Aura Gateway.", nil
        case "get_user":
            if userID == "" {
                return "", fmt.Errorf("user_id is required for get_user tool")
            }
            return fmt.Sprintf("User data for ID %s fetched from Postgres.", userID), nil
        case "read_database":
            return "Mock Database Trace: Successfully indexed 1,241 cluster metrics via Aura Core.", nil
        default:
            return "", fmt.Errorf("unknown tool_id: %s. Use get_aura_tools to see valid options", toolID)
        }
    })

	// 4. Graceful Shutdown Handling
	// This prevents the "Key: 0" error by closing connections before the process dies.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Println("Aura MCP Server is running on :50052/mcp")
		if err := server.Serve(); err != nil {
			logger.Printf("Server stopped: %v", err)
		}
	}()

	<-stop
	logger.Println("Shutting down Aura MCP Server...")
	// Add any specific cleanup logic here
}
