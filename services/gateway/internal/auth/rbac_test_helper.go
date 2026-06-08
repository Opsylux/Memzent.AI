package auth

import "database/sql"

// NewRBACClientForTest allows injecting a mocked database for cross-package tests.
// This is used by internal/engine tests to mock RBAC database calls.
func NewRBACClientForTest(db *sql.DB) *RBACClient {
	return &RBACClient{db: db, devAdminBypass: false}
}
