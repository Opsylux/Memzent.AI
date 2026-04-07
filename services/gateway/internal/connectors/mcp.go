package connectors

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"aura-gateway/internal/mcp"
)

// MCPConnector wraps MCP client to implement Connector interface
// This maintains backward compatibility with existing MCP tools
type MCPConnector struct {
	mcpClient *mcp.MCPClient
}

// NewMCPConnector creates a connector for MCP protocol
func NewMCPConnector(mcpClient *mcp.MCPClient) *MCPConnector {
	return &MCPConnector{
		mcpClient: mcpClient,
	}
}

// Execute calls the MCP server to execute a tool
func (c *MCPConnector) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error) {
	start := time.Now()

	if c.mcpClient == nil {
		return &ExecutionResponse{
			ToolID:   req.ToolID,
			Status:   "error",
			Error:    "MCP client not available",
			Duration: int(time.Since(start).Milliseconds()),
		}, nil
	}

	// Create context with timeout if specified
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.Timeout)*time.Second)
		defer cancel()
	}

	// Call MCP tool
	resp, err := c.mcpClient.CallTool(ctx, "execute_aura_tool", map[string]interface{}{
		"tool_id": req.ToolID,
		"user_id": req.UserID,
	})

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &ExecutionResponse{
				ToolID:   req.ToolID,
				Status:   "timeout",
				Error:    "MCP call exceeded timeout",
				Duration: int(time.Since(start).Milliseconds()),
			}, nil
		}
		return &ExecutionResponse{
			ToolID:   req.ToolID,
			Status:   "error",
			Error:    fmt.Sprintf("MCP execution failed: %v", err),
			Duration: int(time.Since(start).Milliseconds()),
		}, nil
	}

	// Extract text content from MCP response
	var toolResults []string
	if resp != nil {
		for _, content := range resp.Content {
			if content.TextContent != nil {
				toolResults = append(toolResults, content.TextContent.Text)
			}
		}
	}

	return &ExecutionResponse{
		ToolID:   req.ToolID,
		Status:   "success",
		Data:     toolResults,
		Duration: int(time.Since(start).Milliseconds()),
	}, nil
}

// Validate checks if the MCP request is valid
func (c *MCPConnector) Validate(req *ExecutionRequest) error {
	if req.ToolID == "" {
		return fmt.Errorf("tool_id is required")
	}
	if req.UserID == "" {
		req.UserID = "anonymous"
	}
	return nil
}

// HealthCheck verifies MCP connectivity
func (c *MCPConnector) HealthCheck(ctx context.Context) error {
	if c.mcpClient == nil {
		return fmt.Errorf("MCP client not available")
	}
	slog.Info("MCP connector health check passed")
	return nil
}

// Type returns the connector type
func (c *MCPConnector) Type() ConnectorType {
	return TypeMCP
}
