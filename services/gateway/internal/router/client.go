// service/gateway/internal/router/client.go
package router

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RouterClient wraps the gRPC connection to the Rust Semantic Router
type RouterClient struct {
	client SemanticRouterClient
	conn   *grpc.ClientConn
}

// NewRouterClient initializes a gRPC connection to the Rust service
func NewRouterClient(ctx context.Context, addr string) (*RouterClient, error) {
	// retryPolicy defines the gRPC retry strategy for enterprise resilience
	retryPolicy := `{
        "methodConfig": [{
            "name": [{"service": "router.SemanticRouter"}],
            "retryPolicy": {
                "maxAttempts": 3,
                "initialBackoff": "0.1s",
                "maxBackoff": "1s",
                "backoffMultiplier": 2,
                "retryableStatusCodes": ["UNAVAILABLE"]
            }
        }]
    }`

	// ✅ UPGRADED: Using the non-blocking grpc.NewClient
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy),
	)
	if err != nil {
		return nil, fmt.Errorf("could not connect to router at %s: %w", addr, err)
	}

	return &RouterClient{
		client: NewSemanticRouterClient(conn),
		conn:   conn,
	}, nil
}

// GetBestTools calls the Rust Router to find relevant tools for a prompt
// ✅ PRESERVED: Returning tools, compressed prompt, similar hash, and current hash!
func (rc *RouterClient) GetBestTools(ctx context.Context, prompt string, userID string, allowedToolIDs []string) ([]*Tool, string, string, string, error) {
	req := &ToolRequest{
		Prompt:                 prompt,
		UserId:                 userID,
		AllowedToolIds:         allowedToolIDs,
		ScoreThresholdOverride: 0.65,
	}

	// Call the gRPC method defined in your .proto file
	resp, err := rc.client.SelectTools(ctx, req)
	if err != nil {
		return nil, "", "", "", fmt.Errorf("gRPC SelectTools failed: %w", err)
	}

	return resp.Tools, resp.CompressedPrompt, resp.SimilarPromptHash, resp.CurrentPromptHash, nil
}

// RegisterTool notifies the Rust router about a new tool to vectorize it in Qdrant
func (rc *RouterClient) RegisterTool(ctx context.Context, id, name, description, orgID string) (bool, error) {
	req := &RegisterToolRequest{
		Id:          id,
		Name:        name,
		Description: description,
		OrgId:       orgID,
	}

	resp, err := rc.client.RegisterTool(ctx, req)
	if err != nil {
		return false, fmt.Errorf("gRPC RegisterTool failed: %w", err)
	}

	if !resp.Success {
		return false, fmt.Errorf("router failed to register tool: %s", resp.Error)
	}

	return true, nil
}

// PlanToolChain plans a sequence of sequential tools for complex user intents
func (rc *RouterClient) PlanToolChain(ctx context.Context, prompt string, userID string, allowedToolIDs []string) ([]*ToolStep, float32, error) {
	req := &ToolChainRequest{
		Prompt:                 prompt,
		UserId:                 userID,
		AllowedToolIds:         allowedToolIDs,
		ScoreThresholdOverride: 0.65,
	}

	resp, err := rc.client.PlanToolChain(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("gRPC PlanToolChain failed: %w", err)
	}

	return resp.Steps, resp.ConfidenceScore, nil
}


// Close cleans up the gRPC connection
func (rc *RouterClient) Close() {
	if rc.conn != nil {
		rc.conn.Close()
	}
}
