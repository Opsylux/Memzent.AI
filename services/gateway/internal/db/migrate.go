package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MigrationRunner handles sequential execution of SQL migrations
type MigrationRunner struct {
	db            *sql.DB
	migrationsDir string
}

func NewMigrationRunner(db *sql.DB, dir string) *MigrationRunner {
	return &MigrationRunner{db: db, migrationsDir: dir}
}

// Run executes all pending migrations in alphabetical order
func (r *MigrationRunner) Run(ctx context.Context) error {
	slog.Info("Checking for pending migrations...", "dir", r.migrationsDir)

	// 1. Create migrations tracking table
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT now()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	// 2. Load and sort migration files
	files, err := os.ReadDir(r.migrationsDir)
	if err != nil {
		slog.Warn("Migrations directory not found, skipping sync", "path", r.migrationsDir)
		return nil
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)

	// 3. Apply migrations inside a single transaction per file
	for _, filename := range sqlFiles {
		var appliedAt string
		err := r.db.QueryRowContext(ctx, "SELECT applied_at FROM schema_migrations WHERE version = $1", filename).Scan(&appliedAt)
		if err == nil {
			// Already applied
			continue
		}
		if err != sql.ErrNoRows {
			return fmt.Errorf("failed to check migration status for %s: %w", filename, err)
		}

		slog.Info("Applying migration", "version", filename)
		content, err := os.ReadFile(filepath.Join(r.migrationsDir, filename))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		// Execute migration
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", filename); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return err
		}
		slog.Info("Successfully applied migration", "version", filename)
	}

	slog.Info("Database synchronization complete")
	return nil
}
