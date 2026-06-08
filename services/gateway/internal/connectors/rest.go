package connectors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// RESTConnector executes tools via HTTP REST APIs
type RESTConnector struct {
	endpoint string
	client   *http.Client
}

// NewRESTConnector creates a REST connector for an HTTP endpoint
func NewRESTConnector(endpoint string) *RESTConnector {
	return &RESTConnector{
		endpoint: endpoint,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Execute sends a request to the REST endpoint and returns the response
func (c *RESTConnector) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error) {
	start := time.Now()

	// Create context with timeout if specified
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.Timeout)*time.Second)
		defer cancel()
	}

	// Build request body
	payload, err := json.Marshal(req.Inputs)
	if err != nil {
		return &ExecutionResponse{
			ToolID:   req.ToolID,
			Status:   "error",
			Error:    fmt.Sprintf("failed to marshal inputs: %v", err),
			Duration: int(time.Since(start).Milliseconds()),
		}, nil
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return &ExecutionResponse{
			ToolID:   req.ToolID,
			Status:   "error",
			Error:    fmt.Sprintf("failed to create request: %v", err),
			Duration: int(time.Since(start).Milliseconds()),
		}, nil
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Memzent-User-ID", req.UserID)
	httpReq.Header.Set("X-Memzent-Tool-ID", req.ToolID)

	// Execute request
	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &ExecutionResponse{
				ToolID:   req.ToolID,
				Status:   "timeout",
				Error:    "REST call exceeded timeout",
				Duration: int(time.Since(start).Milliseconds()),
			}, nil
		}
		return &ExecutionResponse{
			ToolID:   req.ToolID,
			Status:   "error",
			Error:    fmt.Sprintf("REST call failed: %v", err),
			Duration: int(time.Since(start).Milliseconds()),
		}, nil
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return &ExecutionResponse{
			ToolID:   req.ToolID,
			Status:   "error",
			Error:    fmt.Sprintf("failed to read response: %v", err),
			Duration: int(time.Since(start).Milliseconds()),
		}, nil
	}

	// Check HTTP status
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return &ExecutionResponse{
			ToolID:   req.ToolID,
			Status:   "error",
			Error:    fmt.Sprintf("HTTP %d: %s", httpResp.StatusCode, string(body)),
			Duration: int(time.Since(start).Milliseconds()),
		}, nil
	}

	// Parse response JSON
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		// If not JSON, return raw response
		data = string(body)
	}

	return &ExecutionResponse{
		ToolID:   req.ToolID,
		Status:   "success",
		Data:     data,
		Duration: int(time.Since(start).Milliseconds()),
	}, nil
}

// Validate checks if the REST request is valid
func (c *RESTConnector) Validate(req *ExecutionRequest) error {
	if req.ToolID == "" {
		return fmt.Errorf("tool_id is required")
	}
	if req.Inputs == nil {
		req.Inputs = make(map[string]interface{})
	}
	return nil
}

// HealthCheck verifies the REST endpoint is reachable
func (c *RESTConnector) HealthCheck(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint, nil)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("health check returned HTTP %d", resp.StatusCode)
	}

	slog.Info("REST connector health check passed", "endpoint", c.endpoint)
	return nil
}

// Type returns the connector type
func (c *RESTConnector) Type() ConnectorType {
	return TypeREST
}
