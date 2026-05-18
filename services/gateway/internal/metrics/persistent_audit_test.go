package metrics

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPersistentAuditLogger_Log(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	logger := NewPersistentAuditLogger(db)
	ctx := context.Background()

	event := AuditEvent{
		OrgID:     "123e4567-e89b-12d3-a456-426614174000",
		Type:      "TEST",
		User:      "testuser",
		Detail:    "test action",
		Timestamp: time.Now(),
	}

	mock.ExpectExec("INSERT INTO audit_logs").WithArgs(
		event.OrgID,
		event.User,
		"TEST:test action",
		sqlmock.AnyArg(),
		event.Timestamp,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	logger.Log(ctx, event, map[string]interface{}{"foo": "bar"})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestPersistentAuditLogger_LogSystemOrg(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	logger := NewPersistentAuditLogger(db)
	ctx := context.Background()

	event := AuditEvent{
		OrgID:     "system",
		Type:      "TEST",
		User:      "testuser",
		Detail:    "test action",
		Timestamp: time.Now(),
	}

	mock.ExpectExec("INSERT INTO audit_logs").WithArgs(
		"00000000-0000-0000-0000-000000000000",
		event.User,
		"TEST:test action",
		sqlmock.AnyArg(),
		event.Timestamp,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	logger.Log(ctx, event, nil)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestPersistentAuditLogger_GetLatest(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	logger := NewPersistentAuditLogger(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"org_id", "user_id", "action", "created_at"}).
		AddRow("123e4567-e89b-12d3-a456-426614174000", "user1", "CACHE:hit", now).
		AddRow("123e4567-e89b-12d3-a456-426614174000", "user2", "SYSTEM_BOOT", now)

	mock.ExpectQuery("SELECT org_id, user_id, action, created_at FROM audit_logs").
		WithArgs("123e4567-e89b-12d3-a456-426614174000", 10).
		WillReturnRows(rows)

	events, err := logger.GetLatest("123e4567-e89b-12d3-a456-426614174000", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
	if events[0].Type != "CACHE" || events[0].Detail != "hit" {
		t.Errorf("expected CACHE:hit, got %s:%s", events[0].Type, events[0].Detail)
	}
	if events[1].Type != "SYSTEM" || events[1].Detail != "SYSTEM_BOOT" {
		t.Errorf("expected SYSTEM:SYSTEM_BOOT, got %s:%s", events[1].Type, events[1].Detail)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestPersistentAuditLogger_GetLatest_AllOrgs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	logger := NewPersistentAuditLogger(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"org_id", "user_id", "action", "created_at"}).
		AddRow("123e4567-e89b-12d3-a456-426614174000", "user1", "CACHE:hit", now)

	mock.ExpectQuery("SELECT org_id, user_id, action, created_at FROM audit_logs ORDER BY created_at DESC LIMIT \\$1").
		WithArgs(5).
		WillReturnRows(rows)

	events, err := logger.GetLatest("", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestPersistentAuditLogger_GetCacheStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	logger := NewPersistentAuditLogger(db)

	row := sqlmock.NewRows([]string{"total_requests", "cache_hits"}).AddRow(100, 42)
	
	// Use a simpler regex to match the query since it spans multiple lines
	mock.ExpectQuery("SELECT.+COUNT\\(\\*\\)::bigint as total_requests.+SUM\\(CASE WHEN action LIKE 'CACHE:%'.+").
		WithArgs("org123").
		WillReturnRows(row)

	total, hits := logger.GetCacheStats("org123")
	if total != 100 || hits != 42 {
		t.Errorf("expected 100 total, 42 hits, got %d, %d", total, hits)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestPersistentAuditLogger_Cleanup(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	logger := NewPersistentAuditLogger(db)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM audit_logs WHERE created_at < \\$1").
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 50)) // 50 rows deleted

	logger.Cleanup(ctx, 30)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestPersistentAuditLogger_GetLatest_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	logger := NewPersistentAuditLogger(db)

	mock.ExpectQuery("SELECT org_id, user_id, action, created_at FROM audit_logs").
		WithArgs("org1", 10).
		WillReturnError(fmt.Errorf("db error"))

	events, err := logger.GetLatest("org1", 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if events != nil {
		t.Errorf("expected nil events on error, got %v", events)
	}
}
