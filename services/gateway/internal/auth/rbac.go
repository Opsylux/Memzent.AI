package auth

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type RBACClient struct {
	db *sql.DB
}

// NewRBACClient connects to the Postgres database
func NewRBACClient(connStr string) (*RBACClient, error) {
	// Auto-append binary_parameters=yes to support PgBouncer transaction pooling (e.g. Supabase poolers)
	if strings.HasPrefix(connStr, "postgres://") || strings.HasPrefix(connStr, "postgresql://") {
		if strings.Contains(connStr, "?") {
			if !strings.Contains(connStr, "binary_parameters=") {
				connStr += "&binary_parameters=yes"
			}
		} else {
			connStr += "?binary_parameters=yes"
		}
	}

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

// VerifyAPIKey checks if an API key is valid and returns the associated OrgID, UserID, Scopes, and Role.
// It uses the first 8 characters (prefix) for lookup and bcrypt for verification.
func (c *RBACClient) VerifyAPIKey(ctx context.Context, rawKey string) (string, string, []string, string, error) {
	if len(rawKey) < 8 {
		return "", "", nil, "", fmt.Errorf("invalid API key format")
	}
	prefix := rawKey[:8]

	var orgID, userID, storedHash, role string
	var scopes []string
	// Lookup key by prefix - now including user_id, scopes and role from migration 014
	err := c.db.QueryRowContext(ctx, "SELECT org_id, user_id, key_hash, scopes, role FROM api_keys WHERE key_prefix = $1", prefix).Scan(&orgID, &userID, &storedHash, pq.Array(&scopes), &role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", nil, "", fmt.Errorf("invalid API key")
		}
		return "", "", nil, "", fmt.Errorf("failed to verify API key: %w", err)
	}

	// Compare bcrypt hash. 
	// Note: We use the full rawKey for the check.
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(rawKey))
	if err != nil {
		return "", "", nil, "", fmt.Errorf("invalid API key")
	}

	return orgID, userID, scopes, role, nil
}


func (c *RBACClient) GetMemberRole(ctx context.Context, orgID, userID string) (string, error) {
	var role string
	// We use the 'members' table established in migration 004
	err := c.db.QueryRowContext(ctx, "SELECT role FROM members WHERE org_id = $1 AND user_id = $2", orgID, userID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			// If no specific role is found in DB, default to guest or use JWT role as fallback?
			// For security, we return 'guest' to deny admin actions unless explicitly in DB.
			return "guest", nil
		}
		return "", err
	}
	return role, nil
}

// Close closes the database connection
func (c *RBACClient) Close() {
	if c.db != nil {
		c.db.Close()
	}
}
