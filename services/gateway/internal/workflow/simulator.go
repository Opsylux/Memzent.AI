package workflow

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/lib/pq"
)

const (
	defaultSimulationFrequencyThreshold = 100
	defaultSimulationSampleSize         = 100
	defaultSimulationAccuracyThreshold  = 0.95
	defaultSimulationLookback           = 30 * 24 * time.Hour
)

// SimulationResult captures the outcome of replaying historical executions against a candidate.
type SimulationResult struct {
	TotalReplayed  int     `json:"total_replayed"`
	Successes      int     `json:"successes"`
	Failures       int     `json:"failures"`
	Accuracy       float64 `json:"accuracy"`
	Recommendation string  `json:"recommendation"`
}

// Simulator replays historical workflow executions to validate discovered candidates before review.
type Simulator struct {
	db                 *sql.DB
	registry           *Registry
	frequencyThreshold int
	defaultSampleSize  int
	promotionAccuracy  float64
	lookback           time.Duration
}

// NewSimulator creates a workflow replay simulator backed by Postgres.
func NewSimulator(db *sql.DB) *Simulator {
	return &Simulator{
		db:                 db,
		registry:           NewRegistry(db),
		frequencyThreshold: defaultSimulationFrequencyThreshold,
		defaultSampleSize:  defaultSimulationSampleSize,
		promotionAccuracy:  defaultSimulationAccuracyThreshold,
		lookback:           defaultSimulationLookback,
	}
}

// RunSimulation replays recent executions for the same org/pattern and scores whether the candidate matches them.
func (s *Simulator) RunSimulation(ctx context.Context, candidate *Candidate, sampleSize int) (*SimulationResult, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("workflow simulator is not initialized")
	}
	if candidate == nil {
		return nil, errors.New("workflow candidate is required")
	}
	if candidate.OrgID == "" {
		return nil, fmt.Errorf("workflow candidate %q is missing org_id", candidate.ID)
	}
	if candidate.Pattern == "" {
		return nil, fmt.Errorf("workflow candidate %q is missing pattern", candidate.ID)
	}
	if sampleSize <= 0 {
		sampleSize = s.defaultSampleSize
	}

	cutoff := time.Now().Add(-s.lookback)
	rows, err := s.db.QueryContext(ctx, `
		SELECT wc.tool_ids, we.success
		FROM workflow_executions we
		JOIN workflow_candidates wc ON wc.id = we.workflow_id
		WHERE we.org_id = $1
		  AND wc.pattern = $2
		  AND we.created_at >= $3
		ORDER BY we.created_at DESC
		LIMIT $4
	`, candidate.OrgID, candidate.Pattern, cutoff, sampleSize)
	if err != nil {
		return nil, fmt.Errorf("query workflow executions for candidate %s: %w", candidate.ID, err)
	}
	defer rows.Close()

	result := &SimulationResult{}
	for rows.Next() {
		var executedToolIDs []string
		var executionSuccess bool
		if err := rows.Scan(pq.Array(&executedToolIDs), &executionSuccess); err != nil {
			return nil, fmt.Errorf("scan workflow execution for candidate %s: %w", candidate.ID, err)
		}

		result.TotalReplayed++
		if executionSuccess && slices.Equal(candidate.ToolIDs, executedToolIDs) {
			result.Successes++
		} else {
			result.Failures++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workflow executions for candidate %s: %w", candidate.ID, err)
	}

	if result.TotalReplayed > 0 {
		result.Accuracy = float64(result.Successes) / float64(result.TotalReplayed)
	}
	result.Recommendation = s.recommendationFor(result.Accuracy, result.TotalReplayed)

	slog.Info("Workflow simulation completed",
		"candidate_id", candidate.ID,
		"org_id", candidate.OrgID,
		"pattern", candidate.Pattern,
		"total_replayed", result.TotalReplayed,
		"successes", result.Successes,
		"failures", result.Failures,
		"accuracy", result.Accuracy,
		"recommendation", result.Recommendation,
	)

	return result, nil
}

// SimulateDiscovered replays discovered workflow candidates and updates their lifecycle state.
func (s *Simulator) SimulateDiscovered(ctx context.Context) error {
	if s == nil || s.db == nil {
		return errors.New("workflow simulator is not initialized")
	}

	candidates, err := s.listDiscoveredCandidates(ctx)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return nil
	}

	var joinedErr error
	for i := range candidates {
		if err := ctx.Err(); err != nil {
			return err
		}

		candidate := &candidates[i]
		result, err := s.RunSimulation(ctx, candidate, s.defaultSampleSize)
		if err != nil {
			joinedErr = errors.Join(joinedErr, err)
			slog.Error("Workflow simulation failed", "candidate_id", candidate.ID, "org_id", candidate.OrgID, "error", err)
			continue
		}

		newStatus := StatusSimulated
		if result.Accuracy >= s.promotionAccuracy {
			newStatus = StatusPendingReview
		}

		if err := s.registry.UpdateSimulationResults(ctx, candidate.ID, result.Accuracy, result.TotalReplayed, newStatus); err != nil {
			joinedErr = errors.Join(joinedErr, err)
			slog.Error("Workflow simulation update failed", "candidate_id", candidate.ID, "org_id", candidate.OrgID, "error", err)
			continue
		}

		slog.Info("Workflow candidate simulation state updated",
			"candidate_id", candidate.ID,
			"org_id", candidate.OrgID,
			"status", newStatus,
			"accuracy", result.Accuracy,
			"replay_count", result.TotalReplayed,
		)
	}

	return joinedErr
}

// StartSimulationLoop runs the discovered-candidate simulation job on a fixed interval.
func (s *Simulator) StartSimulationLoop(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour
	}

	go func() {
		runCycle := func() {
			if err := s.SimulateDiscovered(ctx); err != nil && !errors.Is(err, context.Canceled) {
				slog.Error("Workflow simulation cycle failed", "error", err)
			}
		}

		runCycle()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				runCycle()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *Simulator) listDiscoveredCandidates(ctx context.Context) ([]Candidate, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, org_id, pattern, frequency, tool_ids, entity_schema, status,
		       replay_accuracy, replay_count, last_hit_at, hit_count_7d, accuracy_7d,
		       tokens_saved, created_at, updated_at, reviewed_by, reviewed_at,
		       promoted_at, demoted_at, demotion_reason
		FROM workflow_candidates
		WHERE status = $1 AND frequency >= $2
		ORDER BY frequency DESC, updated_at DESC
	`, StatusDiscovered, s.frequencyThreshold)
	if err != nil {
		return nil, fmt.Errorf("list discovered workflow candidates: %w", err)
	}
	defer rows.Close()

	var candidates []Candidate
	for rows.Next() {
		var c Candidate
		var entitySchemaRaw []byte
		if err := rows.Scan(
			&c.ID, &c.OrgID, &c.Pattern, &c.Frequency,
			pq.Array(&c.ToolIDs), &entitySchemaRaw, &c.Status,
			&c.ReplayAccuracy, &c.ReplayCount, &c.LastHitAt,
			&c.HitCount7d, &c.Accuracy7d, &c.TokensSaved,
			&c.CreatedAt, &c.UpdatedAt, &c.ReviewedBy, &c.ReviewedAt,
			&c.PromotedAt, &c.DemotedAt, &c.DemotionReason,
		); err != nil {
			return nil, fmt.Errorf("scan discovered workflow candidate: %w", err)
		}
		if len(entitySchemaRaw) > 0 {
			if err := json.Unmarshal(entitySchemaRaw, &c.EntitySchema); err != nil {
				return nil, fmt.Errorf("unmarshal entity schema for workflow candidate %s: %w", c.ID, err)
			}
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate discovered workflow candidates: %w", err)
	}
	return candidates, nil
}

func (s *Simulator) recommendationFor(accuracy float64, totalReplayed int) string {
	if totalReplayed == 0 {
		return "insufficient_history"
	}
	if accuracy >= s.promotionAccuracy {
		return "promote_to_pending_review"
	}
	if accuracy >= 0.80 {
		return "keep_simulated"
	}
	return "reject_candidate"
}
