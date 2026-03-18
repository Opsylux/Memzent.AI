package mcp

import (
	"context"
	"os"

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
	_, err := c.client.Initialize(ctx)
	return err
}

func (c *MCPClient) ListTools(ctx context.Context) ([]mcp.ToolRetType, error) {
	resp, err := c.client.ListTools(ctx, nil)
	if err != nil {
		return nil, err
	}
	return resp.Tools, nil
}

func (c *MCPClient) CallTool(ctx context.Context, name string, arguments any) (*mcp.ToolResponse, error) {
	return c.client.CallTool(ctx, name, arguments)
}
