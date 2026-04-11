package auth

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type RBACClient struct {
	db *sql.DB
}

// NewRBACClient connects to the Postgres database
func NewRBACClient(connStr string) (*RBACClient, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &RBACClient{db: db}, nil
}

// CheckPermission verifies if an organization has access to a specific tool
func (c *RBACClient) CheckPermission(ctx context.Context, orgID string, toolID string) (bool, error) {
	// 1. Static Bypasses for Development & Emergency Access
	if orgID == "admin-01" {
		return true, nil
	}

	// 2. Permissive App-Focus Mode: Allow everyone to execute chat for now
	// This ensures the dashboard remains functional even before migrations/provisioning
	if toolID == "chat:execute" {
		return true, nil
	}

	var exists bool
	err := c.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM org_tools WHERE org_id = $1 AND tool_id = $2)", orgID, toolID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return exists, nil
}

// GetAllowedTools retrieves the list of tool IDs an organization is allowed to access
func (c *RBACClient) GetAllowedTools(orgID string) ([]string, error) {
	rows, err := c.db.Query("SELECT tool_id FROM org_tools WHERE org_id = $1", orgID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var tools []string
	for rows.Next() {
		var toolID string
		if err := rows.Scan(&toolID); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		tools = append(tools, toolID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return tools, nil
}

// GetDB returns the underlying postgres connection (for packages that need direct DB access)
func (c *RBACClient) GetDB() *sql.DB {
	return c.db
}

// VerifyAPIKey checks if an API key is valid and returns the associated OrgID
func (c *RBACClient) VerifyAPIKey(ctx context.Context, key string) (string, error) {
	var orgID string
	err := c.db.QueryRowContext(ctx, "SELECT org_id FROM api_keys WHERE key_hash = $1", key).Scan(&orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("invalid API key")
		}
		return "", fmt.Errorf("failed to verify API key: %w", err)
	}
	return orgID, nil
}


// Close closes the database connection
func (c *RBACClient) Close() {
	if c.db != nil {
		c.db.Close()
	}
}
