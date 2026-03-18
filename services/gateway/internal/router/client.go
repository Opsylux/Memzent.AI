package router

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RouterClient wraps the gRPC connection to the Rust Semantic Router
type RouterClient struct {
	client SemanticRouterClient
	conn   *grpc.ClientConn
}

// NewRouterClient initializes a gRPC connection to the Rust service
func NewRouterClient(addr string) (*RouterClient, error) {
	// 2026 Best Practice: Use a context with timeout for the initial dial
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect to the Rust service over the internal Docker network
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), // Wait until the connection is ready
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
func (rc *RouterClient) GetBestTools(ctx context.Context, prompt string, userID string, allowedToolIDs []string) ([]*Tool, error) {
	req := &ToolRequest{
		Prompt:         prompt,
		UserId:         userID,
		AllowedToolIds: allowedToolIDs,
	}

	// Call the gRPC method defined in your .proto file
	resp, err := rc.client.SelectTools(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC SelectTools failed: %w", err)
	}

	return resp.Tools, nil
}

// Close cleans up the gRPC connection
func (rc *RouterClient) Close() {
	if rc.conn != nil {
		rc.conn.Close()
	}
}
