package connectors

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CoreToolFunc defines the signature for a native tool function
type CoreToolFunc func(ctx context.Context, userID string, inputs map[string]interface{}) (string, error)

// CoreConnector handles internal tools executed directly in Go
type CoreConnector struct {
	mu    sync.RWMutex
	tools map[string]CoreToolFunc
}

// NewCoreConnector creates a connector for native execution
func NewCoreConnector() *CoreConnector {
	return &CoreConnector{
		tools: make(map[string]CoreToolFunc),
	}
}

// RegisterTool adds a native tool signature to the core connector
func (c *CoreConnector) RegisterTool(toolID string, f CoreToolFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tools[toolID] = f
}

// HasTool checks if a specific tool is handled by this core connector
func (c *CoreConnector) HasTool(toolID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.tools[toolID]
	return ok
}

// Execute runs the core tool natively
func (c *CoreConnector) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error) {
	start := time.Now()

	c.mu.RLock()
	f, ok := c.tools[req.ToolID]
	c.mu.RUnlock()

	if !ok {
		return &ExecutionResponse{
			ToolID:   req.ToolID,
			Status:   "error",
			Error:    fmt.Sprintf("tool not registered in core connector: %s", req.ToolID),
			Duration: int(time.Since(start).Milliseconds()),
		}, nil
	}

	result, err := f(ctx, req.UserID, req.Inputs)
	if err != nil {
		return &ExecutionResponse{
			ToolID:   req.ToolID,
			Status:   "error",
			Error:    err.Error(),
			Duration: int(time.Since(start).Milliseconds()),
		}, nil
	}

	return &ExecutionResponse{
		ToolID:   req.ToolID,
		Status:   "success",
		Data:     result,
		Duration: int(time.Since(start).Milliseconds()),
	}, nil
}

// Validate checks if the core request is valid
func (c *CoreConnector) Validate(req *ExecutionRequest) error {
	if req.ToolID == "" {
		return fmt.Errorf("tool_id is required")
	}
	return nil
}

// HealthCheck is always online for core
func (c *CoreConnector) HealthCheck(ctx context.Context) error {
	return nil
}

// Type returns the connector type
func (c *CoreConnector) Type() ConnectorType {
	return TypeCore
}
