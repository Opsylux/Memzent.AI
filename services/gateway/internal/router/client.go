// service/gateway/internal/router/client.go
package router

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// SemanticRouterInterface defines the contract for the semantic router.
// Implement this interface to mock the router in tests.
type SemanticRouterInterface interface {
	GetBestTools(ctx context.Context, prompt string, orgID string, allowedToolIDs []string) ([]*Tool, string, string, string, error)
	RegisterTool(ctx context.Context, id, name, description, orgID string) (bool, error)
	PlanToolChain(ctx context.Context, prompt string, orgID string, allowedToolIDs []string) ([]*ToolStep, float32, error)
	StoreMemory(ctx context.Context, fact, orgID, userID string) (bool, error)
	QueryMemory(ctx context.Context, prompt, orgID, userID string, threshold float32) ([]*MemoryHit, error)
	Close()
}

// RouterClient wraps the gRPC connection to the Rust Semantic Router
type RouterClient struct {
	client SemanticRouterClient
	conn   *grpc.ClientConn
}

// Compile-time assertion: *RouterClient satisfies SemanticRouterInterface.
var _ SemanticRouterInterface = (*RouterClient)(nil)

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

// GetBestTools calls the Rust Router to find relevant tools for a prompt.
// orgID is passed in the OrgId field so the Rust router can scope the
// prompts_collection cache lookup to this organisation only.
func (rc *RouterClient) GetBestTools(ctx context.Context, prompt string, orgID string, allowedToolIDs []string) ([]*Tool, string, string, string, error) {
	req := &ToolRequest{
		Prompt:                 prompt,
		UserId:                 orgID, // kept for backward compat with existing Qdrant payloads
		OrgId:                  orgID, // new field — drives org-isolated cache lookup
		AllowedToolIds:         allowedToolIDs,
		ScoreThresholdOverride: 0.65,
	}

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

// PlanToolChain plans a sequence of sequential tools for complex user intents.
// orgID is passed through so the Rust router can apply org-scoped RBAC filtering.
func (rc *RouterClient) PlanToolChain(ctx context.Context, prompt string, orgID string, allowedToolIDs []string) ([]*ToolStep, float32, error) {
	req := &ToolChainRequest{
		Prompt:                 prompt,
		UserId:                 orgID, // kept for backward compat
		OrgId:                  orgID, // new field — org-scoped tool filtering
		AllowedToolIds:         allowedToolIDs,
		ScoreThresholdOverride: 0.65,
	}

	resp, err := rc.client.PlanToolChain(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("gRPC PlanToolChain failed: %w", err)
	}

	return resp.Steps, resp.ConfidenceScore, nil
}

// StoreMemory sends a new semantic memory fact to the Rust Router for vector DB insertion
func (rc *RouterClient) StoreMemory(ctx context.Context, fact, orgID, userID string) (bool, error) {
	req := &StoreMemoryRequest{
		Fact:   fact,
		OrgId:  orgID,
		UserId: userID,
	}

	resp, err := rc.client.StoreMemory(ctx, req)
	if err != nil {
		return false, fmt.Errorf("gRPC StoreMemory failed: %w", err)
	}

	return resp.Success, nil
}

// QueryMemory queries Qdrant via the Rust Router for relevant user/org memory facts
func (rc *RouterClient) QueryMemory(ctx context.Context, prompt, orgID, userID string, threshold float32) ([]*MemoryHit, error) {
	req := &QueryMemoryRequest{
		Prompt:                 prompt,
		OrgId:                  orgID,
		UserId:                 userID,
		ScoreThresholdOverride: threshold,
	}

	resp, err := rc.client.QueryMemory(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC QueryMemory failed: %w", err)
	}

	return resp.Memories, nil
}


// FlushPromptCache deletes all cached prompt vectors for an org from Qdrant.
// Uses a direct gRPC call — the message types are defined here until proto is regenerated.
func (rc *RouterClient) FlushPromptCache(ctx context.Context, orgID string) error {
	req := &FlushPromptCacheRequest{OrgId: orgID}
	resp, err := rc.client.FlushPromptCache(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC FlushPromptCache failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("flush failed: %s", resp.Error)
	}
	return nil
}

// Close cleans up the gRPC connection
func (rc *RouterClient) Close() {
	if rc.conn != nil {
		rc.conn.Close()
	}
}
