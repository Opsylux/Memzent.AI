package workflow

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSimulator_RunSimulation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	simulator := NewSimulator(db)
	candidate := &Candidate{
		ID:      "candidate-01",
		OrgID:   "org-01",
		Pattern: "tool-a → tool-b",
		ToolIDs: []string{"tool-a", "tool-b"},
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT wc.tool_ids, we.success
		FROM workflow_executions we
		JOIN workflow_candidates wc ON wc.id = we.workflow_id
		WHERE we.org_id = $1
		  AND wc.pattern = $2
		  AND we.created_at >= $3
		ORDER BY we.created_at DESC
		LIMIT $4
	`)).
		WithArgs(candidate.OrgID, candidate.Pattern, sqlmock.AnyArg(), 3).
		WillReturnRows(sqlmock.NewRows([]string{"tool_ids", "success"}).
			AddRow(`{"tool-a","tool-b"}`, true).
			AddRow(`{"tool-a","tool-b"}`, false).
			AddRow(`{"tool-a","tool-c"}`, true))

	result, err := simulator.RunSimulation(context.Background(), candidate, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalReplayed != 3 || result.Successes != 1 || result.Failures != 2 {
		t.Fatalf("unexpected simulation counts: %+v", result)
	}
	if result.Accuracy != 1.0/3.0 {
		t.Fatalf("unexpected accuracy: %v", result.Accuracy)
	}
	if result.Recommendation != "reject_candidate" {
		t.Fatalf("unexpected recommendation: %s", result.Recommendation)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSimulator_RunSimulationNoHistory(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	simulator := NewSimulator(db)
	candidate := &Candidate{
		ID:      "candidate-02",
		OrgID:   "org-02",
		Pattern: "tool-a → tool-b",
		ToolIDs: []string{"tool-a", "tool-b"},
	}

	mock.ExpectQuery("SELECT wc.tool_ids, we.success").
		WithArgs(candidate.OrgID, candidate.Pattern, sqlmock.AnyArg(), 5).
		WillReturnRows(sqlmock.NewRows([]string{"tool_ids", "success"}))

	result, err := simulator.RunSimulation(context.Background(), candidate, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalReplayed != 0 || result.Accuracy != 0 {
		t.Fatalf("unexpected empty result: %+v", result)
	}
	if result.Recommendation != "insufficient_history" {
		t.Fatalf("unexpected recommendation: %s", result.Recommendation)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSimulator_SimulateDiscovered(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	simulator := NewSimulator(db)
	now := time.Now()

	mock.ExpectQuery("SELECT id, org_id, pattern, frequency, tool_ids, entity_schema, status").
		WithArgs(StatusDiscovered, defaultSimulationFrequencyThreshold).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "org_id", "pattern", "frequency", "tool_ids", "entity_schema", "status",
			"replay_accuracy", "replay_count", "last_hit_at", "hit_count_7d", "accuracy_7d",
			"tokens_saved", "created_at", "updated_at", "reviewed_by", "reviewed_at",
			"promoted_at", "demoted_at", "demotion_reason",
		}).AddRow(
			"candidate-01", "org-01", "tool-a → tool-b", 150, `{"tool-a","tool-b"}`, []byte(`{"account_id":"string"}`), StatusDiscovered,
			nil, nil, nil, 0, 1.0, 0, now, now, nil, nil, nil, nil, nil,
		))

	mock.ExpectQuery("SELECT wc.tool_ids, we.success").
		WithArgs("org-01", "tool-a → tool-b", sqlmock.AnyArg(), defaultSimulationSampleSize).
		WillReturnRows(sqlmock.NewRows([]string{"tool_ids", "success"}).
			AddRow(`{"tool-a","tool-b"}`, true).
			AddRow(`{"tool-a","tool-b"}`, true))

	mock.ExpectExec("UPDATE workflow_candidates").
		WithArgs("candidate-01", 1.0, 2, StatusPendingReview).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := simulator.SimulateDiscovered(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
