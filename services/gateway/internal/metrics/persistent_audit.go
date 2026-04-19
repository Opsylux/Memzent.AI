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
		metaBuf,
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

	query := `
		SELECT org_id, user_id, action, created_at
		FROM audit_logs
		WHERE ($1 = '' OR org_id::text = $1)
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := l.db.Query(query, orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]AuditEvent, 0)
	for rows.Next() {
		var ev AuditEvent
		var action string
		if err := rows.Scan(&ev.OrgID, &ev.User, &action, &ev.Timestamp); err != nil {
			continue
		}

		// Split action back into Type and Detail if possible
		parts := strings.SplitN(action, ":", 2)
		if len(parts) == 2 {
			ev.Type = parts[0]
			ev.Detail = parts[1]
		} else {
			ev.Type = "SYSTEM"
			ev.Detail = action
		}

		ev.Status = "success" // Default for fetched logs
		events = append(events, ev)
	}

	return events, nil
}

// GetCacheStats aggregates the total cache hits and global requests for an org
func (l *PersistentAuditLogger) GetCacheStats(orgID string) (uint64, uint64) {
	if l.db == nil {
		return 0, 0
	}

	query := `
		SELECT 
			COUNT(*) as total_requests,
			SUM(CASE WHEN action LIKE 'CACHE:%' THEN 1 ELSE 0 END) as cache_hits
		FROM audit_logs
		WHERE ($1 = '' OR org_id::text = $1 OR org_id::text = '00000000-0000-0000-0000-000000000000')
	`
	
	var total, hits sql.NullInt64
	err := l.db.QueryRow(query, orgID).Scan(&total, &hits)
	if err != nil {
		slog.Error("Failed to fetch persistent cache stats", "error", err)
		return 0, 0
	}

	return uint64(total.Int64), uint64(hits.Int64)
}
