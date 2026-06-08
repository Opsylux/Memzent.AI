package workflow

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/lib/pq"
)

// Status constants for workflow lifecycle.
const (
	StatusDiscovered    = "discovered"
	StatusSimulated     = "simulated"
	StatusPendingReview = "pending_review"
	StatusApproved      = "approved"
	StatusActive        = "active"
	StatusStale         = "stale"
	StatusDemoted       = "demoted"
)

// Candidate represents a workflow candidate in the registry.
type Candidate struct {
	ID             string            `json:"id"`
	OrgID          string            `json:"org_id"`
	Pattern        string            `json:"pattern"`
	Frequency      int               `json:"frequency"`
	ToolIDs        []string          `json:"tool_ids"`
	EntitySchema   map[string]string `json:"entity_schema,omitempty"`
	Status         string            `json:"status"`
	ReplayAccuracy *float64          `json:"replay_accuracy,omitempty"`
	ReplayCount    *int              `json:"replay_count,omitempty"`
	LastHitAt      *time.Time        `json:"last_hit_at,omitempty"`
	HitCount7d     int               `json:"hit_count_7d"`
	Accuracy7d     float64           `json:"accuracy_7d"`
	TokensSaved    int64             `json:"tokens_saved"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	ReviewedBy     *string           `json:"reviewed_by,omitempty"`
	ReviewedAt     *time.Time        `json:"reviewed_at,omitempty"`
	PromotedAt     *time.Time        `json:"promoted_at,omitempty"`
	DemotedAt      *time.Time        `json:"demoted_at,omitempty"`
	DemotionReason *string           `json:"demotion_reason,omitempty"`
}

// Registry manages the workflow candidate lifecycle in Postgres.
type Registry struct {
	db          *sql.DB
	mu          sync.RWMutex
	activeCache map[string][]Candidate // org_id → active workflows (in-memory hot cache)
}

// NewRegistry creates a new workflow registry backed by Postgres.
func NewRegistry(db *sql.DB) *Registry {
	return &Registry{
		db:          db,
		activeCache: make(map[string][]Candidate),
	}
}

// UpsertCandidate inserts or updates a workflow candidate from the O3 miner.
func (r *Registry) UpsertCandidate(ctx context.Context, orgID, pattern string, frequency int, toolIDs []string, entitySchema map[string]string) (*Candidate, error) {
	schemaJSON, _ := json.Marshal(entitySchema)

	var c Candidate
	var entitySchemaRaw []byte
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO workflow_candidates (org_id, pattern, frequency, tool_ids, entity_schema, status)
		VALUES ($1, $2, $3, $4, $5, 'discovered')
		ON CONFLICT ON CONSTRAINT idx_workflow_candidates_org_pattern
		DO UPDATE SET frequency = EXCLUDED.frequency, updated_at = now()
		RETURNING id, org_id, pattern, frequency, tool_ids, entity_schema, status, 
		          hit_count_7d, accuracy_7d, tokens_saved, created_at, updated_at
	`, orgID, pattern, frequency, pq.Array(toolIDs), schemaJSON).Scan(
		&c.ID, &c.OrgID, &c.Pattern, &c.Frequency,
		pq.Array(&c.ToolIDs), &entitySchemaRaw, &c.Status,
		&c.HitCount7d, &c.Accuracy7d, &c.TokensSaved, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert workflow candidate: %w", err)
	}
	if len(entitySchemaRaw) > 0 {
		_ = json.Unmarshal(entitySchemaRaw, &c.EntitySchema)
	}
	return &c, nil
}

// ListCandidates returns workflow candidates for an org, optionally filtered by status.
func (r *Registry) ListCandidates(ctx context.Context, orgID string, status string) ([]Candidate, error) {
	var rows *sql.Rows
	var err error

	if status != "" {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, org_id, pattern, frequency, tool_ids, entity_schema, status,
			       replay_accuracy, replay_count, last_hit_at, hit_count_7d, accuracy_7d,
			       tokens_saved, created_at, updated_at, reviewed_by, reviewed_at,
			       promoted_at, demoted_at, demotion_reason
			FROM workflow_candidates
			WHERE org_id = $1 AND status = $2
			ORDER BY frequency DESC
		`, orgID, status)
	} else {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, org_id, pattern, frequency, tool_ids, entity_schema, status,
			       replay_accuracy, replay_count, last_hit_at, hit_count_7d, accuracy_7d,
			       tokens_saved, created_at, updated_at, reviewed_by, reviewed_at,
			       promoted_at, demoted_at, demotion_reason
			FROM workflow_candidates
			WHERE org_id = $1
			ORDER BY frequency DESC
		`, orgID)
	}
	if err != nil {
		return nil, fmt.Errorf("list workflow candidates: %w", err)
	}
	defer rows.Close()

	var candidates []Candidate
	for rows.Next() {
		var c Candidate
		var entitySchemaRaw []byte
		err := rows.Scan(
			&c.ID, &c.OrgID, &c.Pattern, &c.Frequency,
			pq.Array(&c.ToolIDs), &entitySchemaRaw, &c.Status,
			&c.ReplayAccuracy, &c.ReplayCount, &c.LastHitAt,
			&c.HitCount7d, &c.Accuracy7d, &c.TokensSaved,
			&c.CreatedAt, &c.UpdatedAt, &c.ReviewedBy, &c.ReviewedAt,
			&c.PromotedAt, &c.DemotedAt, &c.DemotionReason,
		)
		if err != nil {
			return nil, fmt.Errorf("scan workflow candidate: %w", err)
		}
		if len(entitySchemaRaw) > 0 {
			_ = json.Unmarshal(entitySchemaRaw, &c.EntitySchema)
		}
		candidates = append(candidates, c)
	}
	return candidates, nil
}

// Approve transitions a candidate to 'approved' status.
func (r *Registry) Approve(ctx context.Context, id, reviewedBy string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE workflow_candidates
		SET status = 'approved', reviewed_by = $2, reviewed_at = now(), updated_at = now()
		WHERE id = $1 AND status IN ('pending_review', 'simulated', 'discovered')
	`, id, reviewedBy)
	if err != nil {
		return fmt.Errorf("approve workflow: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("workflow %s not found or not in reviewable state", id)
	}
	r.invalidateCache()
	return nil
}

// Activate transitions an approved workflow to 'active' (starts serving traffic).
func (r *Registry) Activate(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE workflow_candidates
		SET status = 'active', promoted_at = now(), updated_at = now()
		WHERE id = $1 AND status = 'approved'
	`, id)
	if err != nil {
		return fmt.Errorf("activate workflow: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("workflow %s not found or not approved", id)
	}
	r.invalidateCache()
	return nil
}

// Reject transitions a candidate to 'demoted' with reason 'manual'.
func (r *Registry) Reject(ctx context.Context, id, reviewedBy string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE workflow_candidates
		SET status = 'demoted', demotion_reason = 'manual', demoted_at = now(),
		    reviewed_by = $2, reviewed_at = now(), updated_at = now()
		WHERE id = $1 AND status NOT IN ('demoted')
	`, id, reviewedBy)
	if err != nil {
		return fmt.Errorf("reject workflow: %w", err)
	}
	r.invalidateCache()
	return nil
}

// UpdateSimulationResults stores replay metrics and advances the candidate lifecycle after simulation.
func (r *Registry) UpdateSimulationResults(ctx context.Context, candidateID string, accuracy float64, replayCount int, newStatus string) error {
	if candidateID == "" {
		return fmt.Errorf("candidate id is required")
	}
	if accuracy < 0 || accuracy > 1 {
		return fmt.Errorf("simulation accuracy must be between 0 and 1")
	}
	if replayCount < 0 {
		return fmt.Errorf("simulation replay count cannot be negative")
	}
	if !isValidWorkflowStatus(newStatus) {
		return fmt.Errorf("invalid workflow status %q", newStatus)
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE workflow_candidates
		SET replay_accuracy = $2,
		    replay_count = $3,
		    status = $4,
		    updated_at = now()
		WHERE id = $1 AND status IN ('discovered', 'simulated')
	`, candidateID, accuracy, replayCount, newStatus)
	if err != nil {
		return fmt.Errorf("update workflow simulation results: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("workflow %s not found", candidateID)
	}
	if newStatus == StatusPendingReview {
		slog.Info("Workflow auto-promoted after simulation", "candidate_id", candidateID, "accuracy", accuracy, "replay_count", replayCount)
	}
	r.invalidateCache()
	return nil
}

// RecordExecution logs a workflow execution for accuracy tracking.
func (r *Registry) RecordExecution(ctx context.Context, workflowID, orgID, promptHash string, entities map[string]string, success bool, latencyMs int) error {
	entitiesJSON, _ := json.Marshal(entities)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO workflow_executions (workflow_id, org_id, prompt_hash, entities, success, latency_ms)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, workflowID, orgID, promptHash, entitiesJSON, success, latencyMs)
	if err != nil {
		return fmt.Errorf("record workflow execution: %w", err)
	}

	// Update hit tracking on the candidate
	_, err = r.db.ExecContext(ctx, `
		UPDATE workflow_candidates
		SET last_hit_at = now(), hit_count_7d = hit_count_7d + 1, updated_at = now()
		WHERE id = $1
	`, workflowID)
	return err
}

// GetActiveWorkflows returns all active workflows for an org (cached in memory).
func (r *Registry) GetActiveWorkflows(ctx context.Context, orgID string) ([]Candidate, error) {
	r.mu.RLock()
	if cached, ok := r.activeCache[orgID]; ok {
		r.mu.RUnlock()
		return cached, nil
	}
	r.mu.RUnlock()

	candidates, err := r.ListCandidates(ctx, orgID, StatusActive)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.activeCache[orgID] = candidates
	r.mu.Unlock()
	return candidates, nil
}

// RunDemotionCheck scans for stale/underperforming workflows and demotes them.
// - Active with <10 hits/day for 7 days → stale
// - Stale for 7 more days or accuracy <85% → demoted
func (r *Registry) RunDemotionCheck(ctx context.Context) (staled, demoted int, err error) {
	// Mark active workflows with low hit count as stale
	result, err := r.db.ExecContext(ctx, `
		UPDATE workflow_candidates
		SET status = 'stale', updated_at = now()
		WHERE status = 'active'
		AND (last_hit_at IS NULL OR last_hit_at < now() - interval '7 days')
		AND hit_count_7d < 70
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("stale check: %w", err)
	}
	s, _ := result.RowsAffected()
	staled = int(s)

	// Demote stale workflows that haven't recovered or have low accuracy
	result, err = r.db.ExecContext(ctx, `
		UPDATE workflow_candidates
		SET status = 'demoted', demoted_at = now(), updated_at = now(),
		    demotion_reason = CASE
		        WHEN accuracy_7d < 0.85 THEN 'accuracy_drop'
		        ELSE 'frequency_drop'
		    END
		WHERE status = 'stale'
		AND (
		    (last_hit_at IS NULL OR last_hit_at < now() - interval '14 days')
		    OR accuracy_7d < 0.85
		)
	`)
	if err != nil {
		return staled, 0, fmt.Errorf("demotion check: %w", err)
	}
	d, _ := result.RowsAffected()
	demoted = int(d)

	if staled > 0 || demoted > 0 {
		r.invalidateCache()
		slog.Info("🔄 Workflow demotion check", "staled", staled, "demoted", demoted)
	}
	return staled, demoted, nil
}

// StartDemotionLoop runs demotion checks periodically in the background.
func (r *Registry) StartDemotionLoop(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s, d, err := r.RunDemotionCheck(ctx)
				if err != nil {
					slog.Error("Workflow demotion check failed", "error", err)
				} else if s > 0 || d > 0 {
					slog.Info("Workflow demotion cycle", "staled", s, "demoted", d)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// MatchWorkflow checks if an active workflow matches the given tool sequence.
func (r *Registry) MatchWorkflow(ctx context.Context, orgID string, toolPattern string) (*Candidate, error) {
	workflows, err := r.GetActiveWorkflows(ctx, orgID)
	if err != nil {
		return nil, err
	}
	for i := range workflows {
		if workflows[i].Pattern == toolPattern {
			return &workflows[i], nil
		}
	}
	return nil, nil
}

func (r *Registry) invalidateCache() {
	r.mu.Lock()
	r.activeCache = make(map[string][]Candidate)
	r.mu.Unlock()
}

func isValidWorkflowStatus(status string) bool {
	switch status {
	case StatusDiscovered, StatusSimulated, StatusPendingReview, StatusApproved, StatusActive, StatusStale, StatusDemoted:
		return true
	default:
		return false
	}
}
