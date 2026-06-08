package connectors

import (
	"context"
)

// ConnectorType identifies the tool execution protocol
type ConnectorType string

const (
	TypeCore   ConnectorType = "core"
	TypeMCP    ConnectorType = "mcp"
	TypeREST   ConnectorType = "rest"
	TypeSQL    ConnectorType = "sql"
	TypeGraphQL ConnectorType = "graphql"
	TypeGRPC   ConnectorType = "grpc"
	TypeWebhook ConnectorType = "webhook"
)

// ExecutionRequest is the input to a connector
type ExecutionRequest struct {
	ToolID  string                 `json:"tool_id"`
	UserID  string                 `json:"user_id"`
	Inputs  map[string]interface{} `json:"inputs"`
	Timeout int                    `json:"timeout_seconds"`
}

// ExecutionResponse is the output from a connector
type ExecutionResponse struct {
	ToolID   string      `json:"tool_id"`
	Status   string      `json:"status"` // "success", "error", "timeout"
	Data     interface{} `json:"data,omitempty"`
	Error    string      `json:"error,omitempty"`
	Duration int         `json:"duration_ms"`
}

// Connector defines the interface for tool execution protocols
type Connector interface {
	// Execute runs the tool with given inputs, returns results or error
	Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error)

	// Validate checks if the request is valid for this connector
	Validate(req *ExecutionRequest) error

	// HealthCheck verifies the connector can reach its backend
	HealthCheck(ctx context.Context) error

	// Type returns the connector type identifier
	Type() ConnectorType
}

// Factory creates connectors based on type and endpoint
type Factory interface {
	// CreateConnector builds a connector for the given tool endpoint
	CreateConnector(connectorType ConnectorType, endpoint string) (Connector, error)
}

// ConnectorRegistry manages available connectors
type ConnectorRegistry struct {
	connectors map[ConnectorType]Connector
}

// NewConnectorRegistry creates an empty registry
func NewConnectorRegistry() *ConnectorRegistry {
	return &ConnectorRegistry{
		connectors: make(map[ConnectorType]Connector),
	}
}

// Register adds a connector to the registry
func (r *ConnectorRegistry) Register(t ConnectorType, c Connector) {
	r.connectors[t] = c
}

// Get retrieves a connector by type
func (r *ConnectorRegistry) Get(t ConnectorType) (Connector, bool) {
	c, ok := r.connectors[t]
	return c, ok
}
