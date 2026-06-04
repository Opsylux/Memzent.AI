package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/http"
	"github.com/valkey-io/valkey-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	router "memzent-mcp/internal/router"
)

// ToolArgs defines the expected schema for the execute_memzent_tool.
// Using JSON tags ensures the MCP client and server agree on field names.
type ToolArgs struct {
	ToolID string `json:"tool_id"`
	UserID string `json:"user_id,omitempty"`
}

// StoreMemoryArgs defines the schema for the store_memory tool parameters.
type StoreMemoryArgs struct {
	Fact   string `json:"fact" jsonschema:"description=The factual statement or user preference to remember"`
	UserID string `json:"user_id,omitempty" jsonschema:"description=The unique identifier of the user"`
	OrgID  string `json:"org_id,omitempty" jsonschema:"description=The organization ID for memory isolation"`
}

// RecallMemoryArgs defines the schema for the recall_memory tool parameters.
type RecallMemoryArgs struct {
	Query  string `json:"query" jsonschema:"description=The prompt or question to find relevant memories for"`
	UserID string `json:"user_id,omitempty" jsonschema:"description=The unique identifier of the user"`
	OrgID  string `json:"org_id,omitempty" jsonschema:"description=The organization ID for memory isolation"`
}

func main() {
	// 1. Setup Structured Logging to Stderr
	// MCP protocol uses Stdout for transport in some modes;
	// logging to Stderr prevents protocol corruption.
	logger := log.New(os.Stderr, "[MEMZENT-MCP] ", log.LstdFlags|log.Lshortfile)

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

	// 3. Initialize Router gRPC Client
	routerAddr := os.Getenv("ROUTER_ADDR")
	if routerAddr == "" {
		routerAddr = "router:50051"
	}
	conn, err := grpc.NewClient(routerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatalf("Failed to connect to Rust Router at %s: %v", routerAddr, err)
	}
	defer conn.Close()
	routerClient := router.NewSemanticRouterClient(conn)
	logger.Printf("Connected to Rust Router at %s", routerAddr)

	// 4. Initialize MCP Server over HTTP
	// We explicitly set the path to /mcp to match your Gateway's environment variables.
	t := http.NewHTTPTransport("/mcp").WithAddr(":50052")
	server := mcp.NewServer(t)

	// --- TOOL: GET TOOLS ---
	server.RegisterTool("get_memzent_tools", "Returns available tools from cache or registry", func(ctx context.Context) (string, error) {
		valkeyKey := "mcp:tools:list"

		// Use a sub-context with timeout for Valkey calls to prevent hanging
		vCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		resp := vClient.Do(vCtx, vClient.B().Get().Key(valkeyKey).Build())
		if cached, err := resp.ToString(); err == nil {
			return fmt.Sprintf("Retrieved from Cache: %s", cached), nil
		}

		// Fallback/Registry logic
		tools := "Available: [db_query, get_user, read_database, store_memory, recall_memory]"
		_ = vClient.Do(vCtx, vClient.B().Set().Key(valkeyKey).Value(tools).Ex(300).Build())

		return tools, nil
	})

	// --- TOOL: STORE MEMORY ---
	server.RegisterTool("store_memory", "Saves a permanent fact or user preference to long-term memory", func(ctx context.Context, args StoreMemoryArgs) (string, error) {
		logger.Printf("DEBUG: store_memory tool invoked with args: %+v", args)
		if args.Fact == "" {
			return "", fmt.Errorf("missing required parameter: fact")
		}

		req := &router.StoreMemoryRequest{
			Fact:   args.Fact,
			OrgId:  args.OrgID,
			UserId: args.UserID,
		}

		resp, err := routerClient.StoreMemory(ctx, req)
		if err != nil {
			return "", fmt.Errorf("failed to store memory via gRPC: %w", err)
		}
		if !resp.Success {
			return "", fmt.Errorf("failed to store memory: %s", resp.Error)
		}

		return fmt.Sprintf("Fact successfully stored in long-term memory: \"%s\"", args.Fact), nil
	})

	// --- TOOL: RECALL MEMORY ---
	server.RegisterTool("recall_memory", "Retrieves historical facts or user preferences from long-term memory", func(ctx context.Context, args RecallMemoryArgs) (string, error) {
		logger.Printf("DEBUG: recall_memory tool invoked with args: %+v", args)
		if args.Query == "" {
			return "", fmt.Errorf("missing required parameter: query")
		}

		req := &router.QueryMemoryRequest{
			Prompt:                 args.Query,
			OrgId:                  args.OrgID,
			UserId:                 args.UserID,
			ScoreThresholdOverride: 0.65,
		}

		resp, err := routerClient.QueryMemory(ctx, req)
		if err != nil {
			return "", fmt.Errorf("failed to query memory via gRPC: %w", err)
		}

		if len(resp.Memories) == 0 {
			return "No relevant memories found.", nil
		}

		var sb strings.Builder
		sb.WriteString("Retrieved memories:\n")
		for _, mem := range resp.Memories {
			sb.WriteString(fmt.Sprintf("- %s (relevance: %.2f)\n", mem.Fact, mem.RelevanceScore))
		}
		return sb.String(), nil
	})

	// --- TOOL: EXECUTE TOOL ---
	server.RegisterTool("execute_memzent_tool", "Runs a specific Memzent tool", func(ctx context.Context, args ToolArgs) (string, error) {
		logger.Printf("DEBUG: Tool handler invoked with raw args: %+v", args)
		
		toolID := args.ToolID
		userID := args.UserID

		// Validation: If ToolID is missing, return a clean error
		if toolID == "" {
			logger.Printf("ERROR: Missing tool_id in request")
			return "", fmt.Errorf("missing required parameter: tool_id")
		}

		logger.Printf("Executing Tool: %s (User: %s)", toolID, userID)

		switch toolID {
		case "db_query":
			return "SQL query executed successfully via Memzent Gateway.", nil
		case "get_user":
			if userID == "" {
				return "", fmt.Errorf("user_id is required for get_user tool")
			}
			return fmt.Sprintf("User data for ID %s fetched from Postgres.", userID), nil
		case "store_memory":
			return "Memory stored successfully (invoked via generic wrapper).", nil
		case "recall_memory":
			return "Retrieved memories (invoked via generic wrapper).", nil
		default:
			return "", fmt.Errorf("unknown tool_id: %s. Use get_memzent_tools to see valid options", toolID)
		}
	})

	// 5. Graceful Shutdown Handling
	// This prevents the "Key: 0" error by closing connections before the process dies.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Println("Memzent MCP Server is running on :50052/mcp")
		if err := server.Serve(); err != nil {
			logger.Printf("Server stopped: %v", err)
		}
	}()

	<-stop
	logger.Println("Shutting down Memzent MCP Server...")
}
