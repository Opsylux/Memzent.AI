package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"sync"
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

// SyncCallback is called after a successful Refresh with the list of tools that need re-vectorization.
type SyncCallback func(ctx context.Context, tools []*Tool)

// Registry manages tool storage and retrieval from Postgres.
// It also drives the Phase 2 incremental refresh loop that keeps
// the Qdrant vector store in sync with Postgres.
type Registry struct {
	db           *sql.DB
	lastRefresh  time.Time
	mu           sync.RWMutex
}

// NewRegistry creates a new tool registry backed by Postgres.
func NewRegistry(db *sql.DB) *Registry {
	return &Registry{db: db}
}

// Refresh polls Postgres for tools that have drifted from the Qdrant vector store
// (i.e., last_synced_at IS NULL  or last_synced_at < updated_at).
// It invokes onSync for any tools it finds, then marks them as synced in Postgres.
// Returns the number of tools that required syncing.
func (r *Registry) Refresh(ctx context.Context, onSync SyncCallback) (int, error) {
	query := `
		SELECT id, org_id, name, description, connector_type, endpoint,
		       config, input_schema, output_schema, timeout_seconds,
		       enabled, requires_auth, created_at, updated_at
		FROM tools
		WHERE enabled = true
		  AND (last_synced_at IS NULL OR last_synced_at < updated_at)
		ORDER BY updated_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var driftedTools []*Tool
	for rows.Next() {
		tool := &Tool{}
		var configData, inputData, outputData []byte
		err := rows.Scan(
			&tool.ID, &tool.OrgID, &tool.Name, &tool.Description, &tool.ConnectorType, &tool.Endpoint,
			&configData, &inputData, &outputData, &tool.TimeoutSeconds,
			&tool.Enabled, &tool.RequiresAuth, &tool.CreatedAt, &tool.UpdatedAt,
		)
		if err != nil {
			return 0, err
		}
		json.Unmarshal(configData, &tool.Config)
		json.Unmarshal(inputData, &tool.InputSchema)
		json.Unmarshal(outputData, &tool.OutputSchema)
		driftedTools = append(driftedTools, tool)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	if len(driftedTools) == 0 {
		r.mu.Lock()
		r.lastRefresh = time.Now()
		r.mu.Unlock()
		return 0, nil
	}

	// Invoke the caller-provided callback (e.g., gRPC to Rust for vectorization)
	if onSync != nil {
		onSync(ctx, driftedTools)
	}

	// Mark all drifted tools as synced
	for _, t := range driftedTools {
		_, _ = r.db.ExecContext(ctx,
			`UPDATE tools SET last_synced_at = NOW() WHERE id = $1`,
			t.ID,
		)
	}

	r.mu.Lock()
	r.lastRefresh = time.Now()
	r.mu.Unlock()

	return len(driftedTools), nil
}

// LastRefreshTime returns the time of the most recent successful refresh.
func (r *Registry) LastRefreshTime() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastRefresh
}

// StartRefreshLoop runs Refresh on a fixed interval until ctx is cancelled.
// The onSync callback is forwarded directly to Refresh on every cycle.
func (r *Registry) StartRefreshLoop(ctx context.Context, interval time.Duration, onSync SyncCallback) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("[ToolRegistry] Refresh loop started", "interval", interval)

	// Run an eager first pass immediately
	if n, err := r.Refresh(ctx, onSync); err != nil {
		slog.Warn("[ToolRegistry] Initial refresh failed", "error", err)
	} else if n > 0 {
		slog.Info("[ToolRegistry] Initial sync complete", "tools_vectorized", n)
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("[ToolRegistry] Refresh loop stopping")
			return
		case <-ticker.C:
			if n, err := r.Refresh(ctx, onSync); err != nil {
				slog.Warn("[ToolRegistry] Periodic refresh failed", "error", err)
			} else if n > 0 {
				slog.Info("[ToolRegistry] Periodic sync complete", "tools_vectorized", n)
			}
		}
	}
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

// UpdateTool updates mutable fields of an existing tool
func (r *Registry) UpdateTool(ctx context.Context, tool *Tool) error {
	query := `
		UPDATE tools SET
			name = $2,
			description = $3,
			connector_type = $4,
			endpoint = $5,
			config = $6,
			input_schema = $7,
			output_schema = $8,
			timeout_seconds = $9,
			enabled = $10,
			requires_auth = $11,
			updated_at = NOW()
		WHERE id = $1
	`
	configBuf, _ := json.Marshal(tool.Config)
	inputBuf, _ := json.Marshal(tool.InputSchema)
	outputBuf, _ := json.Marshal(tool.OutputSchema)

	_, err := r.db.ExecContext(ctx, query,
		tool.ID, tool.Name, tool.Description, tool.ConnectorType, tool.Endpoint,
		configBuf, inputBuf, outputBuf, tool.TimeoutSeconds,
		tool.Enabled, tool.RequiresAuth,
	)
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
		var inputData, outputData []byte
		err := rows.Scan(
			&tool.ID, &tool.Name, &tool.Description, &tool.ConnectorType, &tool.Endpoint,
			&inputData, &outputData, &tool.TimeoutSeconds,
			&tool.Enabled, &tool.RequiresAuth, &tool.CreatedAt, &tool.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(inputData, &tool.InputSchema)
		json.Unmarshal(outputData, &tool.OutputSchema)
		tools = append(tools, tool)
	}

	return tools, rows.Err()
}
