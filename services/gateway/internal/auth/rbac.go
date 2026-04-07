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

	// For demonstration purposes, we ensure the table exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_tools (
			user_id VARCHAR(50) NOT NULL,
			tool_id VARCHAR(50) NOT NULL,
			PRIMARY KEY (user_id, tool_id)
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure user_tools table exists: %w", err)
	}

	// Insert mock data if we are testing with 'solo-user' or 'admin-01'
	db.Exec(`INSERT INTO user_tools (user_id, tool_id) VALUES ('solo-user', 'tool_123') ON CONFLICT DO NOTHING`)
	db.Exec(`INSERT INTO user_tools (user_id, tool_id) VALUES ('solo-user', 'read_database') ON CONFLICT DO NOTHING`)
	db.Exec(`INSERT INTO user_tools (user_id, tool_id) VALUES ('admin-01', 'aura_search') ON CONFLICT DO NOTHING`)
	db.Exec(`INSERT INTO user_tools (user_id, tool_id) VALUES ('admin-01', 'read_database') ON CONFLICT DO NOTHING`)

	return &RBACClient{db: db}, nil
}

// CheckPermission verifies if a user has a specific scope/permission
func (c *RBACClient) CheckPermission(ctx context.Context, userID string, scope string) (bool, error) {
	// For testing, we'll allow 'solo-user' and 'admin-01' to execute anything
	if userID == "solo-user" || userID == "admin-01" {
		return true, nil
	}

	// In a real scenario, we'd query a permissions table
	// var exists bool
	// err := c.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM user_permissions WHERE user_id = $1 AND scope = $2)", userID, scope).Scan(&exists)
	// return exists, err

	return false, nil
}

// GetAllowedTools retrieves the list of tool IDs a user is allowed to access
func (c *RBACClient) GetAllowedTools(userID string) ([]string, error) {
	rows, err := c.db.Query("SELECT tool_id FROM user_tools WHERE user_id = $1", userID)
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


// Close closes the database connection
func (c *RBACClient) Close() {
	if c.db != nil {
		c.db.Close()
	}
}
