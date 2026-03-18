package main

import (
	"context"
	"fmt"
	"os"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport"
	"github.com/valkey-io/valkey-go"
)

func main() {

	// 1. Initialize Valkey Client (Go Native)
	valkeyAddr := os.Getenv("VALKEY_URL")
	if valkeyAddr == "" {
		valkeyAddr = "valkey:6379"
	}

	vClient, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{valkeyAddr}})
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect to Valkey: %v\n", err)
		os.Exit(1)
	}
	defer vClient.Close()

	// 2. Initialize MCP Server over Stdio
	server := mcp.NewServer(transport.NewStdioTransport())

	// 3. Tool: Get Tools
	server.RegisterTool("get_aura_tools", "Returns available tools", func(ctx context.Context) (string, error) {
		valkeyKey := "mcp:tools:list"

		// Check Cache
		resp := vClient.Do(ctx, vClient.B().Get().Key(valkeyKey).Build())
		if cached, err := resp.ToString(); err == nil {
			return fmt.Sprintf("Retrieved from Cache: %s", cached), nil
		}

		// Mock Tool List
		tools := "Available: [db_query, get_user]"

		// Set Cache with 5m TTL
		_ = vClient.Do(ctx, vClient.B().Set().Key(valkeyKey).Value(tools).Ex(300).Build())

		return tools, nil
	})

	// 4. Tool: Execute Tool
	type ToolArgs struct {
		ToolID string `json:"tool_id"`
		UserID string `json:"user_id,omitempty"`
	}

	server.RegisterTool("execute_aura_tool", "Runs a specific Aura tool", func(ctx context.Context, args ToolArgs) (string, error) {
		switch args.ToolID {
		case "db_query":
			return "✅ SQL query executed successfully via Aura Gateway.", nil
		case "get_user":
			return fmt.Sprintf("👤 User data for ID %s fetched from Postgres.", args.UserID), nil
		default:
			return "", fmt.Errorf("tool %s not found", args.ToolID)
		}
	})

	// 5. Start the MCP Server
	if err := server.Serve(); err != nil {
		fmt.Fprintf(os.Stderr, "MCP Server Error: %v\n", err)
		os.Exit(1)
	}
}
