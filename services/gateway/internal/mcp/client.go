package mcp

import (
	"context"
	"fmt"
	"os"
	"time"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/http"
)

type MCPClient struct {
	client *mcp.Client
}

func NewMCPClient() (*MCPClient, error) {
	mcpAddr := os.Getenv("MCP_SERVER_URL")
	if mcpAddr == "" {
		mcpAddr = "http://aura-mcp-server:50052/mcp"
	}

	transport := http.NewHTTPClientTransport(mcpAddr)
	client := mcp.NewClient(transport)

	// We initialize the client immediately or on the first call
	// For simplicity, we'll try to initialize it now
	// But note that it might fail if the server isn't up yet
	return &MCPClient{client: client}, nil
}

func (c *MCPClient) Initialize(ctx context.Context) error {
    if c.client == nil {
        return fmt.Errorf("mcp client not configured")
    }

    // Retry logic: MCP server might be starting up
    maxRetries := 5
    retryDelay := 1 * time.Second

    var lastErr error

    for attempt := 1; attempt <= maxRetries; attempt++ {
        initCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
        _, err := c.client.Initialize(initCtx)
        cancel()

        if err == nil {
            fmt.Fprintf(os.Stderr, "[MCPClient] SUCCESS: Initialized on attempt %d\n", attempt)
            return nil
        }

        lastErr = err
        fmt.Fprintf(os.Stderr, "[MCPClient] RETRY %d/%d: Initialization failed - %v\n", attempt, maxRetries, err)

        if attempt < maxRetries {
            time.Sleep(retryDelay)
        }
    }

    return fmt.Errorf("mcp client initialization failed after %d retries: %w", maxRetries, lastErr)
}

func (c *MCPClient) ListTools(ctx context.Context) ([]mcp.ToolRetType, error) {
	resp, err := c.client.ListTools(ctx, nil)
	if err != nil {
		return nil, err
	}
	return resp.Tools, nil
}

func (c *MCPClient) CallTool(ctx context.Context, name string, arguments any) (*mcp.ToolResponse, error) {
    // 1. Ensure the client isn't nil
    if c.client == nil {
        return nil, fmt.Errorf("mcp client not configured")
    }

    // 2. Debug: Log the exact arguments being passed
    fmt.Fprintf(os.Stderr, "[MCPClient] DEBUG: Calling tool '%s' with arguments: %+v (type: %T)\n", name, arguments, arguments)

    // 3. Call tool with comprehensive error handling
    resp, err := c.client.CallTool(ctx, name, arguments)
    if err != nil {
        // Log more details about the error for debugging
        fmt.Fprintf(os.Stderr, "[MCPClient] ERROR: Tool call failed - Tool: %s, Error: %v, Error Type: %T\n", name, err, err)
        return nil, fmt.Errorf("mcp call tool [%s] failed: %w", name, err)
    }

    // 4. Verify response is not nil
    if resp == nil {
        return nil, fmt.Errorf("mcp call tool [%s] returned nil response", name)
    }

    fmt.Fprintf(os.Stderr, "[MCPClient] SUCCESS: Tool call returned %d content items\n", len(resp.Content))
    return resp, nil
}