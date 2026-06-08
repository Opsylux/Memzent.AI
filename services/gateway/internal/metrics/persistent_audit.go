package metrics

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"strings"
	"time"
)

// PersistentAuditLogger handles storage of audit events in Postgres
type PersistentAuditLogger struct {
	db *sql.DB
}

func NewPersistentAuditLogger(db *sql.DB) *PersistentAuditLogger {
	return &PersistentAuditLogger{db: db}
}

// Log persists an event to the database and also adds it to the in-memory buffer for the real-time UI
func (l *PersistentAuditLogger) Log(ctx context.Context, event AuditEvent, metadata map[string]interface{}) {
	// 1. Add to in-memory buffer (Real-time compatibility)
	GlobalAuditBuffer.Add(event)

	// 2. Persist to Postgres (Enterprise Compliance)
	if l.db == nil {
		return
	}

	metaBuf, _ := json.Marshal(metadata)

	const systemOrgID = "00000000-0000-0000-0000-000000000000"
	targetOrgID := event.OrgID
	if targetOrgID == "system" || targetOrgID == "" {
		targetOrgID = systemOrgID
	}

	query := `
		INSERT INTO audit_logs (org_id, user_id, action, metadata, created_at)
		VALUES ($1::uuid, NULLIF($2, ''), $3, $4, $5)
	`
	_, err := l.db.ExecContext(ctx, query,
		targetOrgID,
		event.User,
		event.Type+":"+event.Detail, // Normalize action type
		string(metaBuf),
		event.Timestamp,
	)

	if err != nil {
		slog.Error("Failed to persist audit log", "error", err, "org_id", event.OrgID)
	}
}

// StartRetentionJob runs a background task to delete old logs
func (l *PersistentAuditLogger) StartRetentionJob(ctx context.Context, retentionDays int) {
	if l.db == nil {
		return
	}

	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			select {
			case <-ticker.C:
				l.Cleanup(ctx, retentionDays)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Initial cleanup on startup
	go l.Cleanup(ctx, retentionDays)
}

func (l *PersistentAuditLogger) Cleanup(ctx context.Context, days int) {
	if l.db == nil {
		return
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	slog.Info("Running Audit Retention Job", "cutoff", cutoff)

	query := `DELETE FROM audit_logs WHERE created_at < $1`
	res, err := l.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		slog.Error("Audit Retention Job failed", "error", err)
		return
	}

	rows, _ := res.RowsAffected()
	slog.Info("Audit Retention completed", "rows_deleted", rows)
}

// GetLatest returns at most 'limit' audit events from Postgres. If orgID is empty, it returns for all orgs.
func (l *PersistentAuditLogger) GetLatest(orgID string, limit int) ([]AuditEvent, error) {
	if l.db == nil {
		return []AuditEvent{}, nil
	}

	// 1. Build query dynamically to handle "all orgs" and keep indexes active
	var rows *sql.Rows
	var err error

	// 2. Pre-allocate slice capacity to the limit to avoid re-allocations
	events := make([]AuditEvent, 0, limit)

	if orgID == "" {
		query := `
            SELECT org_id, user_id, action, created_at
            FROM audit_logs
            ORDER BY created_at DESC
            LIMIT $1`
		rows, err = l.db.Query(query, limit)
	} else {
		query := `
            SELECT org_id, user_id, action, created_at
            FROM audit_logs
            WHERE org_id = $1
            ORDER BY created_at DESC
            LIMIT $2`
		rows, err = l.db.Query(query, orgID, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ev AuditEvent
		var action string
		if err := rows.Scan(&ev.OrgID, &ev.User, &action, &ev.Timestamp); err != nil {
			return nil, err // Don't silently fail
		}

		// Optimization: Standard string split is fine, but ensure DB schema
		// eventually reflects these as separate columns for best performance.
		parts := strings.SplitN(action, ":", 2)
		if len(parts) == 2 {
			ev.Type, ev.Detail = parts[0], parts[1]
		} else {
			ev.Type, ev.Detail = "SYSTEM", action
		}

		ev.Status = "success"
		events = append(events, ev)
	}

	// Check for errors encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// GetCacheStats aggregates the total cache hits and global requests for an org
func (l *PersistentAuditLogger) GetCacheStats(orgID string) (uint64, uint64) {
	if l.db == nil {
		return 0, 0
	}

	// Strict org isolation — only count this org's audit entries
	query := `
		SELECT 
			COUNT(*)::bigint as total_requests,
			COALESCE(SUM(CASE WHEN action LIKE 'CACHE:%' THEN 1 ELSE 0 END), 0)::bigint as cache_hits
		FROM audit_logs
		WHERE org_id::text = $1
	`

	var total, hits sql.NullInt64
	err := l.db.QueryRow(query, orgID).Scan(&total, &hits)
	if err != nil {
		slog.Error("Failed to fetch persistent cache stats", "error", err)
		return 0, 0
	}

	return uint64(total.Int64), uint64(hits.Int64)
}
