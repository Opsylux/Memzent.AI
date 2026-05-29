package auth

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

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

// rotationGracePeriod is the window during which both the old and new key hash
// are accepted after a rotation event, allowing agents to swap keys without downtime.
const rotationGracePeriod = 15 * time.Minute

// VerifyAPIKey checks if an API key is valid and returns the associated OrgID, UserID, Scopes, and Role.
// It enforces:
//   - Expiry TTL (expires_at)
//   - Dual-hash acceptance during rotation grace window (prev_key_hash)
//   - last_used_at update on every successful auth
//   - Automatic clearing of prev_key_hash once the grace window has passed
func (c *RBACClient) VerifyAPIKey(ctx context.Context, rawKey string) (string, string, []string, string, error) {
	if len(rawKey) < 16 {
		return "", "", nil, "", fmt.Errorf("invalid API key format")
	}
	prefix := rawKey[:16]

	var orgID, userID, keyID, storedHash, role string
	var scopes []string
	var expiresAt sql.NullTime
	var prevKeyHash sql.NullString
	var rotatedAt sql.NullTime

	err := c.db.QueryRowContext(ctx,
		`SELECT id, org_id, user_id, key_hash, scopes, role, expires_at, prev_key_hash, rotated_at
		 FROM api_keys WHERE key_prefix = $1`,
		prefix,
	).Scan(&keyID, &orgID, &userID, &storedHash, pq.Array(&scopes), &role,
		&expiresAt, &prevKeyHash, &rotatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", nil, "", fmt.Errorf("invalid API key")
		}
		return "", "", nil, "", fmt.Errorf("failed to verify API key: %w", err)
	}

	// 1. Enforce expiry TTL
	if expiresAt.Valid && time.Now().UTC().After(expiresAt.Time) {
		return "", "", nil, "", fmt.Errorf("API key has expired")
	}

	// 2. Verify bcrypt hash — try primary hash first, then prev_key_hash (rotation grace)
	primaryErr := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(rawKey))
	if primaryErr != nil {
		// Try the previous hash if we're within the rotation grace window
		if prevKeyHash.Valid && rotatedAt.Valid &&
			time.Since(rotatedAt.Time) < rotationGracePeriod {
			prevErr := bcrypt.CompareHashAndPassword([]byte(prevKeyHash.String), []byte(rawKey))
			if prevErr != nil {
				return "", "", nil, "", fmt.Errorf("invalid API key")
			}
			// Accepted on old hash — don't clear yet, grace window still active
		} else {
			// Either no prev hash or grace window expired
			return "", "", nil, "", fmt.Errorf("invalid API key")
		}
	} else if prevKeyHash.Valid && rotatedAt.Valid &&
		time.Since(rotatedAt.Time) >= rotationGracePeriod {
		// New hash accepted and grace window has passed — clear prev_key_hash async
		go c.clearPrevKeyHash(keyID)
	}

	// 3. Update last_used_at asynchronously — don't block the auth path
	go c.updateLastUsed(keyID)

	return orgID, userID, scopes, role, nil
}

// updateLastUsed stamps the key's last_used_at without blocking the auth hot path.
func (c *RBACClient) updateLastUsed(keyID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, _ = c.db.ExecContext(ctx,
		"UPDATE api_keys SET last_used_at = now() WHERE id = $1", keyID)
}

// clearPrevKeyHash removes the rotation overlap hash once the grace window has elapsed.
func (c *RBACClient) clearPrevKeyHash(keyID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, _ = c.db.ExecContext(ctx,
		"UPDATE api_keys SET prev_key_hash = NULL WHERE id = $1", keyID)
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
