package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMigrationRunner_Run_NonExistentDir(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnResult(sqlmock.NewResult(0, 0))

	runner := NewMigrationRunner(db, "/non-existent-directory-path-abcd")
	err = runner.Run(context.Background())
	if err != nil {
		t.Errorf("expected no error when dir is missing, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestMigrationRunner_Run_CreateTableFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnError(fmt.Errorf("db connection failure"))

	runner := NewMigrationRunner(db, "/some-dir")
	err = runner.Run(context.Background())
	if err == nil {
		t.Errorf("expected error when table creation fails, got nil")
	}
}

func TestMigrationRunner_Run_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	tempDir := t.TempDir()
	
	err = os.WriteFile(filepath.Join(tempDir, "001_init.sql"), []byte("CREATE TABLE users (id int);"), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "002_update.sql"), []byte("ALTER TABLE users ADD COLUMN name text;"), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// 1. Table creation mock
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnResult(sqlmock.NewResult(0, 0))

	// 2. Migration 1: 001_init.sql - Pending
	mock.ExpectQuery("SELECT applied_at FROM schema_migrations WHERE version = \\$1").
		WithArgs("001_init.sql").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE users \\(id int\\);").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO schema_migrations \\(version\\) VALUES \\(\\$1\\)").
		WithArgs("001_init.sql").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// 3. Migration 2: 002_update.sql - Pending
	mock.ExpectQuery("SELECT applied_at FROM schema_migrations WHERE version = \\$1").
		WithArgs("002_update.sql").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectBegin()
	mock.ExpectExec("ALTER TABLE users ADD COLUMN name text;").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO schema_migrations \\(version\\) VALUES \\(\\$1\\)").
		WithArgs("002_update.sql").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	runner := NewMigrationRunner(db, tempDir)
	err = runner.Run(context.Background())
	if err != nil {
		t.Errorf("expected successful migration run, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestMigrationRunner_Run_SkipAppliedAndRollbackOnFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	tempDir := t.TempDir()

	// 001_init.sql: Will be skipped (already applied)
	err = os.WriteFile(filepath.Join(tempDir, "001_init.sql"), []byte("CREATE TABLE users (id int);"), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	// 002_fail.sql: Will fail during execution (rollback test)
	err = os.WriteFile(filepath.Join(tempDir, "002_fail.sql"), []byte("INVALID SQL syntax;"), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// 1. Table creation mock
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnResult(sqlmock.NewResult(0, 0))

	// 2. Migration 1: 001_init.sql - Already Applied
	mock.ExpectQuery("SELECT applied_at FROM schema_migrations WHERE version = \\$1").
		WithArgs("001_init.sql").
		WillReturnRows(sqlmock.NewRows([]string{"applied_at"}).AddRow("2026-06-03 00:00:00"))

	// 3. Migration 2: 002_fail.sql - Pending but Fails
	mock.ExpectQuery("SELECT applied_at FROM schema_migrations WHERE version = \\$1").
		WithArgs("002_fail.sql").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectBegin()
	mock.ExpectExec("INVALID SQL syntax;").
		WillReturnError(fmt.Errorf("syntax error near INVALID"))
	mock.ExpectRollback()

	runner := NewMigrationRunner(db, tempDir)
	err = runner.Run(context.Background())
	if err == nil {
		t.Errorf("expected error on failed migration exec, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestMigrationRunner_Run_StatusCheckError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	tempDir := t.TempDir()
	err = os.WriteFile(filepath.Join(tempDir, "001_init.sql"), []byte("CREATE TABLE users;"), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery("SELECT applied_at FROM schema_migrations WHERE version = \\$1").
		WithArgs("001_init.sql").
		WillReturnError(fmt.Errorf("database query error"))

	runner := NewMigrationRunner(db, tempDir)
	err = runner.Run(context.Background())
	if err == nil {
		t.Errorf("expected error on database query error, got nil")
	}
}
