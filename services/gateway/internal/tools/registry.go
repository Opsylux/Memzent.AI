package tools

import (
	"context"
	"database/sql"
	"time"
)

// ToolConnectorType defines supported tool execution protocols
type ToolConnectorType string

const (
	ConnectorMCP    ToolConnectorType = "mcp"       // Model Context Protocol (current)
	ConnectorREST   ToolConnectorType = "rest"      // REST API (to be implemented)
	ConnectorSQL    ToolConnectorType = "sql"       // Direct SQL (to be implemented)
	ConnectorGraphQL ToolConnectorType = "graphql"  // GraphQL (to be implemented)
	ConnectorGRPC   ToolConnectorType = "grpc"      // gRPC service (to be implemented)
	ConnectorWebhook ToolConnectorType = "webhook"  // Async webhook (to be implemented)
)

// Tool represents a registered AI tool/agent capability
type Tool struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	ConnectorType   ToolConnectorType      `json:"connector_type"`
	Endpoint        string                 `json:"endpoint"`        // API URL, DB connection, MCP tool name, etc.
	InputSchema     map[string]interface{} `json:"input_schema"`    // JSON schema for tool inputs
	OutputSchema    map[string]interface{} `json:"output_schema"`   // JSON schema for tool outputs
	TimeoutSeconds  int                    `json:"timeout_seconds"` // Max execution time
	Enabled         bool                   `json:"enabled"`
	RequiresAuth    bool                   `json:"requires_auth"`   // Needs user RBAC check
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// Registry manages tool storage and retrieval from Postgres
type Registry struct {
	db *sql.DB
}

// NewRegistry creates a new tool registry backed by Postgres
func NewRegistry(db *sql.DB) *Registry {
	return &Registry{db: db}
}

// RegisterTool stores a new tool in the registry
func (r *Registry) RegisterTool(ctx context.Context, tool *Tool) error {
	query := `
		INSERT INTO tools (id, name, description, connector_type, endpoint, 
		                   input_schema, output_schema, timeout_seconds, 
		                   enabled, requires_auth, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			name = $2,
			description = $3,
			connector_type = $4,
			endpoint = $5,
			input_schema = $6,
			output_schema = $7,
			timeout_seconds = $8,
			enabled = $9,
			requires_auth = $10,
			updated_at = $12
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		tool.ID, tool.Name, tool.Description, tool.ConnectorType, tool.Endpoint,
		tool.InputSchema, tool.OutputSchema, tool.TimeoutSeconds,
		tool.Enabled, tool.RequiresAuth, now, now,
	)
	return err
}

// GetTool retrieves a single tool by ID
func (r *Registry) GetTool(ctx context.Context, toolID string) (*Tool, error) {
	tool := &Tool{}
	query := `
		SELECT id, name, description, connector_type, endpoint,
		       input_schema, output_schema, timeout_seconds,
		       enabled, requires_auth, created_at, updated_at
		FROM tools WHERE id = $1 AND enabled = true
	`

	err := r.db.QueryRowContext(ctx, query, toolID).Scan(
		&tool.ID, &tool.Name, &tool.Description, &tool.ConnectorType, &tool.Endpoint,
		&tool.InputSchema, &tool.OutputSchema, &tool.TimeoutSeconds,
		&tool.Enabled, &tool.RequiresAuth, &tool.CreatedAt, &tool.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return tool, err
}

// ListTools retrieves all enabled tools
func (r *Registry) ListTools(ctx context.Context) ([]*Tool, error) {
	query := `
		SELECT id, name, description, connector_type, endpoint,
		       input_schema, output_schema, timeout_seconds,
		       enabled, requires_auth, created_at, updated_at
		FROM tools WHERE enabled = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []*Tool
	for rows.Next() {
		tool := &Tool{}
		err := rows.Scan(
			&tool.ID, &tool.Name, &tool.Description, &tool.ConnectorType, &tool.Endpoint,
			&tool.InputSchema, &tool.OutputSchema, &tool.TimeoutSeconds,
			&tool.Enabled, &tool.RequiresAuth, &tool.CreatedAt, &tool.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tools = append(tools, tool)
	}

	return tools, rows.Err()
}

// DisableTool soft-deletes a tool (sets enabled = false)
func (r *Registry) DisableTool(ctx context.Context, toolID string) error {
	query := `UPDATE tools SET enabled = false, updated_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), toolID)
	return err
}

// ListByConnectorType returns all tools of a specific connector type
func (r *Registry) ListByConnectorType(ctx context.Context, connectorType ToolConnectorType) ([]*Tool, error) {
	query := `
		SELECT id, name, description, connector_type, endpoint,
		       input_schema, output_schema, timeout_seconds,
		       enabled, requires_auth, created_at, updated_at
		FROM tools WHERE connector_type = $1 AND enabled = true
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, string(connectorType))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []*Tool
	for rows.Next() {
		tool := &Tool{}
		err := rows.Scan(
			&tool.ID, &tool.Name, &tool.Description, &tool.ConnectorType, &tool.Endpoint,
			&tool.InputSchema, &tool.OutputSchema, &tool.TimeoutSeconds,
			&tool.Enabled, &tool.RequiresAuth, &tool.CreatedAt, &tool.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tools = append(tools, tool)
	}

	return tools, rows.Err()
}
