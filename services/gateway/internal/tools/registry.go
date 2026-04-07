package tools

import (
	"context"
	"database/sql"
	"encoding/json"
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
	OrgID           *string                `json:"org_id,omitempty"` // Null for system-wide tools
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	ConnectorType   ToolConnectorType      `json:"connector_type"`
	Endpoint        string                 `json:"endpoint"`
	Config          map[string]interface{} `json:"config"` // Tool-specific dynamic config
	InputSchema     map[string]interface{} `json:"input_schema"`
	OutputSchema    map[string]interface{} `json:"output_schema"`
	TimeoutSeconds  int                    `json:"timeout_seconds"`
	Enabled         bool                   `json:"enabled"`
	RequiresAuth    bool                   `json:"requires_auth"`
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
		INSERT INTO tools (id, org_id, name, description, connector_type, endpoint, 
		                   config, input_schema, output_schema, timeout_seconds, 
		                   enabled, requires_auth, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		ON CONFLICT (id) DO UPDATE SET
			name = $3,
			description = $4,
			connector_type = $5,
			endpoint = $6,
			config = $7,
			input_schema = $8,
			output_schema = $9,
			timeout_seconds = $10,
			enabled = $11,
			requires_auth = $12,
			updated_at = NOW()
	`

	configBuf, _ := json.Marshal(tool.Config)
	inputBuf, _ := json.Marshal(tool.InputSchema)
	outputBuf, _ := json.Marshal(tool.OutputSchema)

	_, err := r.db.ExecContext(ctx, query,
		tool.ID, tool.OrgID, tool.Name, tool.Description, tool.ConnectorType, tool.Endpoint,
		configBuf, inputBuf, outputBuf, tool.TimeoutSeconds,
		tool.Enabled, tool.RequiresAuth,
	)
	return err
}

// GetTool retrieves a single tool by ID
func (r *Registry) GetTool(ctx context.Context, toolID string) (*Tool, error) {
	tool := &Tool{}
	var configData, inputData, outputData []byte
	query := `
		SELECT id, org_id, name, description, connector_type, endpoint,
		       config, input_schema, output_schema, timeout_seconds,
		       enabled, requires_auth, created_at, updated_at
		FROM tools WHERE id = $1 AND enabled = true
	`

	err := r.db.QueryRowContext(ctx, query, toolID).Scan(
		&tool.ID, &tool.OrgID, &tool.Name, &tool.Description, &tool.ConnectorType, &tool.Endpoint,
		&configData, &inputData, &outputData, &tool.TimeoutSeconds,
		&tool.Enabled, &tool.RequiresAuth, &tool.CreatedAt, &tool.UpdatedAt,
	)

	json.Unmarshal(configData, &tool.Config)
	json.Unmarshal(inputData, &tool.InputSchema)
	json.Unmarshal(outputData, &tool.OutputSchema)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return tool, err
}

// ListTools retrieves all tools available for an organization (scoped to org or system)
func (r *Registry) ListTools(ctx context.Context, orgID string) ([]*Tool, error) {
	query := `
		SELECT id, org_id, name, description, connector_type, endpoint,
		       config, input_schema, output_schema, timeout_seconds,
		       enabled, requires_auth, created_at, updated_at
		FROM tools WHERE enabled = true AND (org_id = $1 OR org_id IS NULL)
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []*Tool
	for rows.Next() {
		tool := &Tool{}
		var configData, inputData, outputData []byte
		err := rows.Scan(
			&tool.ID, &tool.OrgID, &tool.Name, &tool.Description, &tool.ConnectorType, &tool.Endpoint,
			&configData, &inputData, &outputData, &tool.TimeoutSeconds,
			&tool.Enabled, &tool.RequiresAuth, &tool.CreatedAt, &tool.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(configData, &tool.Config)
		json.Unmarshal(inputData, &tool.InputSchema)
		json.Unmarshal(outputData, &tool.OutputSchema)
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
