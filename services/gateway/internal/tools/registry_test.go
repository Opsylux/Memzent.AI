package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRegistry_RegisterTool(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	registry := NewRegistry(db)
	ctx := context.Background()

	orgID := "org1"
	tool := &Tool{
		ID:             "tool-01",
		OrgID:          &orgID,
		Name:           "Test Tool",
		Description:    "Does tests",
		ConnectorType:  ConnectorMCP,
		Endpoint:       "http://test-endpoint",
		Config:         map[string]interface{}{"key": "val"},
		InputSchema:    map[string]interface{}{"type": "object"},
		OutputSchema:   map[string]interface{}{"type": "string"},
		TimeoutSeconds: 15,
		Enabled:        true,
		RequiresAuth:   true,
	}

	configBuf, _ := json.Marshal(tool.Config)
	inputBuf, _ := json.Marshal(tool.InputSchema)
	outputBuf, _ := json.Marshal(tool.OutputSchema)

	mock.ExpectExec("INSERT INTO tools").
		WithArgs(
			tool.ID, tool.OrgID, tool.Name, tool.Description, tool.ConnectorType, tool.Endpoint,
			configBuf, inputBuf, outputBuf, tool.TimeoutSeconds,
			tool.Enabled, tool.RequiresAuth,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = registry.RegisterTool(ctx, tool)
	if err != nil {
		t.Errorf("unexpected error on register: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRegistry_GetTool(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	registry := NewRegistry(db)
	ctx := context.Background()

	configData := []byte(`{"key": "val"}`)
	inputData := []byte(`{"type": "object"}`)
	outputData := []byte(`{"type": "string"}`)
	now := time.Now()

	mock.ExpectQuery("SELECT id, org_id, name, description, connector_type, endpoint").
		WithArgs("tool-01").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "org_id", "name", "description", "connector_type", "endpoint",
			"config", "input_schema", "output_schema", "timeout_seconds",
			"enabled", "requires_auth", "created_at", "updated_at",
		}).AddRow(
			"tool-01", "org1", "Test Tool", "Does tests", "mcp", "http://test-endpoint",
			configData, inputData, outputData, 15,
			true, true, now, now,
		))

	tool, err := registry.GetTool(ctx, "tool-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tool == nil {
		t.Fatalf("expected tool, got nil")
	}
	if tool.ID != "tool-01" || tool.Config["key"] != "val" {
		t.Errorf("unexpected tool contents: %+v", tool)
	}

	// ErrNoRows
	mock.ExpectQuery("SELECT id, org_id, name, description, connector_type, endpoint").
		WithArgs("missing-tool").
		WillReturnError(sql.ErrNoRows)

	tool, err = registry.GetTool(ctx, "missing-tool")
	if err != nil {
		t.Errorf("expected no error on ErrNoRows, got: %v", err)
	}
	if tool != nil {
		t.Errorf("expected nil tool, got: %+v", tool)
	}
}

func TestRegistry_ListTools(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	registry := NewRegistry(db)
	ctx := context.Background()

	now := time.Now()
	mock.ExpectQuery("SELECT id, org_id, name, description, connector_type, endpoint").
		WithArgs("org1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "org_id", "name", "description", "connector_type", "endpoint",
			"config", "input_schema", "output_schema", "timeout_seconds",
			"enabled", "requires_auth", "created_at", "updated_at",
		}).AddRow(
			"tool-01", "org1", "Test Tool", "Description", "mcp", "endpoint",
			[]byte("{}"), []byte("{}"), []byte("{}"), 15,
			true, false, now, now,
		))

	tools, err := registry.ListTools(ctx, "org1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tools) != 1 || tools[0].ID != "tool-01" {
		t.Errorf("unexpected tools list: %v", tools)
	}
}

func TestRegistry_DisableTool(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	registry := NewRegistry(db)
	ctx := context.Background()

	mock.ExpectExec("UPDATE tools SET enabled = false").
		WithArgs(sqlmock.AnyArg(), "tool-01").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = registry.DisableTool(ctx, "tool-01")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRegistry_ListByConnectorType(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	registry := NewRegistry(db)
	ctx := context.Background()

	now := time.Now()
	mock.ExpectQuery("SELECT id, name, description, connector_type, endpoint").
		WithArgs("mcp").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "connector_type", "endpoint",
			"input_schema", "output_schema", "timeout_seconds",
			"enabled", "requires_auth", "created_at", "updated_at",
		}).AddRow(
			"tool-01", "Test Tool", "Description", "mcp", "endpoint",
			[]byte("{}"), []byte("{}"), 15,
			true, false, now, now,
		))

	tools, err := registry.ListByConnectorType(ctx, ConnectorMCP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tools) != 1 || tools[0].ID != "tool-01" {
		t.Errorf("unexpected tools: %v", tools)
	}
}

func TestRegistry_RefreshAndLoop(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	registry := NewRegistry(db)
	ctx := context.Background()

	now := time.Now()

	t.Run("No drifted tools", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, org_id, name, description, connector_type, endpoint").
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "org_id", "name", "description", "connector_type", "endpoint",
				"config", "input_schema", "output_schema", "timeout_seconds",
				"enabled", "requires_auth", "created_at", "updated_at",
			}))

		n, err := registry.Refresh(ctx, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if n != 0 {
			t.Errorf("expected 0 drifted tools, got %d", n)
		}
		if registry.LastRefreshTime().IsZero() {
			t.Errorf("expected LastRefreshTime to be set")
		}
	})

	t.Run("drifted tools found and synced", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, org_id, name, description, connector_type, endpoint").
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "org_id", "name", "description", "connector_type", "endpoint",
				"config", "input_schema", "output_schema", "timeout_seconds",
				"enabled", "requires_auth", "created_at", "updated_at",
			}).AddRow(
				"tool-01", "org1", "Name", "Desc", "mcp", "endpoint",
				[]byte("{}"), []byte("{}"), []byte("{}"), 15,
				true, false, now, now,
			))

		mock.ExpectExec("UPDATE tools SET last_synced_at = NOW\\(\\)").
			WithArgs("tool-01").
			WillReturnResult(sqlmock.NewResult(1, 1))

		syncCalled := false
		onSync := func(c context.Context, tools []*Tool) {
			if len(tools) == 1 && tools[0].ID == "tool-01" {
				syncCalled = true
			}
		}

		n, err := registry.Refresh(ctx, onSync)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if n != 1 {
			t.Errorf("expected 1 tool synced, got %d", n)
		}
		if !syncCalled {
			t.Errorf("onSync callback was not called")
		}
	})

	t.Run("StartRefreshLoop eager exit", func(t *testing.T) {
		// Mock query for the eager first refresh inside loop
		mock.ExpectQuery("SELECT id, org_id, name, description, connector_type, endpoint").
			WillReturnError(fmt.Errorf("db error"))

		cancelCtx, cancel := context.WithCancel(ctx)
		// Cancel immediately so the loop exits right after eager refresh
		cancel()

		registry.StartRefreshLoop(cancelCtx, 50*time.Millisecond, nil)
	})
}
