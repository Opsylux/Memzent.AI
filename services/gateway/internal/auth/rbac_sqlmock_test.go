package auth

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"golang.org/x/crypto/bcrypt"
)

func TestRBACClient_CheckPermission(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	client := &RBACClient{db: db}
	ctx := context.Background()

	// Static bypasses
	allowed, err := client.CheckPermission(ctx, "admin-01", "some_tool")
	if err != nil || !allowed {
		t.Errorf("admin-01 bypass failed")
	}

	allowed, err = client.CheckPermission(ctx, "some_org", "chat:execute")
	if err != nil || !allowed {
		t.Errorf("chat:execute bypass failed")
	}

	// Normal check - allowed
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("org1", "tool1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	allowed, err = client.CheckPermission(ctx, "org1", "tool1")
	if err != nil || !allowed {
		t.Errorf("CheckPermission expected true, got false")
	}

	// Normal check - denied
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("org1", "tool2").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	allowed, err = client.CheckPermission(ctx, "org1", "tool2")
	if err != nil || allowed {
		t.Errorf("CheckPermission expected false, got true")
	}

	// Query error
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("org1", "tool3").
		WillReturnError(fmt.Errorf("db error"))

	_, err = client.CheckPermission(ctx, "org1", "tool3")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func TestRBACClient_GetAllowedTools(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	client := &RBACClient{db: db}

	mock.ExpectQuery("SELECT tool_id FROM org_tools").
		WithArgs("org1").
		WillReturnRows(sqlmock.NewRows([]string{"tool_id"}).AddRow("tool1").AddRow("tool2"))

	tools, err := client.GetAllowedTools("org1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tools) != 2 || tools[0] != "tool1" || tools[1] != "tool2" {
		t.Errorf("unexpected tools result: %v", tools)
	}

	// Query error
	mock.ExpectQuery("SELECT tool_id FROM org_tools").
		WithArgs("org2").
		WillReturnError(fmt.Errorf("db error"))

	_, err = client.GetAllowedTools("org2")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestRBACClient_GetMemberRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	client := &RBACClient{db: db}
	ctx := context.Background()

	mock.ExpectQuery("SELECT role FROM members").
		WithArgs("org1", "user1").
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("admin"))

	role, err := client.GetMemberRole(ctx, "org1", "user1")
	if err != nil || role != "admin" {
		t.Errorf("expected admin, got %v (err: %v)", role, err)
	}

	// ErrNoRows
	mock.ExpectQuery("SELECT role FROM members").
		WithArgs("org1", "user2").
		WillReturnError(sql.ErrNoRows)

	role, err = client.GetMemberRole(ctx, "org1", "user2")
	if err != nil || role != "guest" {
		t.Errorf("expected guest on ErrNoRows, got %v (err: %v)", role, err)
	}

	// Query error
	mock.ExpectQuery("SELECT role FROM members").
		WithArgs("org1", "user3").
		WillReturnError(fmt.Errorf("db error"))

	_, err = client.GetMemberRole(ctx, "org1", "user3")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestRBACClient_VerifyAPIKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	client := &RBACClient{db: db}
	ctx := context.Background()

	rawKey := "12345678-real-key"
	hashedKey, _ := bcrypt.GenerateFromPassword([]byte(rawKey), bcrypt.MinCost)

	// Valid key
	mock.ExpectQuery("SELECT org_id, user_id, key_hash, scopes, role FROM api_keys").
		WithArgs("12345678-real-ke").
		WillReturnRows(sqlmock.NewRows([]string{"org_id", "user_id", "key_hash", "scopes", "role"}).
			AddRow("org1", "user1", string(hashedKey), "{read,write}", "admin"))

	orgID, userID, scopes, role, err := client.VerifyAPIKey(ctx, rawKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if orgID != "org1" || userID != "user1" || role != "admin" || len(scopes) != 2 {
		t.Errorf("unexpected results: %v, %v, %v, %v", orgID, userID, scopes, role)
	}

	// Invalid hash
	mock.ExpectQuery("SELECT org_id, user_id, key_hash, scopes, role FROM api_keys").
		WithArgs("12345678-real-ke").
		WillReturnRows(sqlmock.NewRows([]string{"org_id", "user_id", "key_hash", "scopes", "role"}).
			AddRow("org1", "user1", "invalid_hash", "{read,write}", "admin"))

	_, _, _, _, err = client.VerifyAPIKey(ctx, rawKey)
	if err == nil {
		t.Errorf("Expected error for invalid hash, got nil")
	}

	// ErrNoRows
	mock.ExpectQuery("SELECT org_id, user_id, key_hash, scopes, role FROM api_keys").
		WithArgs("12345678-real-ke").
		WillReturnError(sql.ErrNoRows)

	_, _, _, _, err = client.VerifyAPIKey(ctx, rawKey)
	if err == nil {
		t.Errorf("Expected error for ErrNoRows, got nil")
	}
}

func TestRBACClient_GetDBAndClose(t *testing.T) {
	db, _, _ := sqlmock.New()
	client := &RBACClient{db: db}

	if client.GetDB() != db {
		t.Errorf("GetDB returned wrong db")
	}

	client.Close()
	// db is closed now
}
