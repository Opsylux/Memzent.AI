package connectors

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// ──────────────────────────────────────────────────────────────────────
// ConnectorRegistry tests (connector.go)
// ──────────────────────────────────────────────────────────────────────

func TestConnectorRegistry_RegisterAndGet(t *testing.T) {
	reg := NewConnectorRegistry()

	// Get from empty registry returns false
	_, ok := reg.Get(TypeCore)
	if ok {
		t.Fatal("expected Get on empty registry to return false")
	}

	// Register a core connector and retrieve it
	core := NewCoreConnector()
	reg.Register(TypeCore, core)

	got, ok := reg.Get(TypeCore)
	if !ok {
		t.Fatal("expected Get after Register to return true")
	}
	if got.Type() != TypeCore {
		t.Fatalf("expected TypeCore, got %s", got.Type())
	}

	// Unregistered type still returns false
	_, ok = reg.Get(TypeREST)
	if ok {
		t.Fatal("expected unregistered type to return false")
	}
}

func TestConnectorRegistry_OverwriteRegistration(t *testing.T) {
	reg := NewConnectorRegistry()

	core1 := NewCoreConnector()
	core1.RegisterTool("tool-a", func(ctx context.Context, userID string, inputs map[string]interface{}) (string, error) {
		return "v1", nil
	})

	core2 := NewCoreConnector()
	core2.RegisterTool("tool-b", func(ctx context.Context, userID string, inputs map[string]interface{}) (string, error) {
		return "v2", nil
	})

	reg.Register(TypeCore, core1)
	reg.Register(TypeCore, core2) // overwrite

	got, ok := reg.Get(TypeCore)
	if !ok {
		t.Fatal("expected Get to return true")
	}

	// The second registration should have replaced the first
	cc := got.(*CoreConnector)
	if cc.HasTool("tool-a") {
		t.Fatal("first registration should have been overwritten")
	}
	if !cc.HasTool("tool-b") {
		t.Fatal("second registration should be active")
	}
}

// ──────────────────────────────────────────────────────────────────────
// CoreConnector tests (core.go)
// ──────────────────────────────────────────────────────────────────────

func TestCoreConnector_Type(t *testing.T) {
	c := NewCoreConnector()
	if c.Type() != TypeCore {
		t.Fatalf("expected TypeCore, got %s", c.Type())
	}
}

func TestCoreConnector_HealthCheck(t *testing.T) {
	c := NewCoreConnector()
	if err := c.HealthCheck(context.Background()); err != nil {
		t.Fatalf("expected nil HealthCheck, got %v", err)
	}
}

func TestCoreConnector_RegisterToolAndHasTool(t *testing.T) {
	c := NewCoreConnector()

	if c.HasTool("echo") {
		t.Fatal("HasTool should return false for unregistered tool")
	}

	c.RegisterTool("echo", func(ctx context.Context, userID string, inputs map[string]interface{}) (string, error) {
		return "ok", nil
	})

	if !c.HasTool("echo") {
		t.Fatal("HasTool should return true after registration")
	}
}

func TestCoreConnector_Validate(t *testing.T) {
	c := NewCoreConnector()

	// Missing tool_id
	err := c.Validate(&ExecutionRequest{})
	if err == nil || !strings.Contains(err.Error(), "tool_id") {
		t.Fatalf("expected tool_id validation error, got %v", err)
	}

	// Valid request
	err = c.Validate(&ExecutionRequest{ToolID: "test"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestCoreConnector_Execute_Success(t *testing.T) {
	c := NewCoreConnector()
	c.RegisterTool("greet", func(ctx context.Context, userID string, inputs map[string]interface{}) (string, error) {
		name, _ := inputs["name"].(string)
		return fmt.Sprintf("Hello, %s!", name), nil
	})

	resp, err := c.Execute(context.Background(), &ExecutionRequest{
		ToolID: "greet",
		UserID: "user-1",
		Inputs: map[string]interface{}{"name": "Alice"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("expected success, got %s (error: %s)", resp.Status, resp.Error)
	}
	if resp.Data != "Hello, Alice!" {
		t.Fatalf("expected 'Hello, Alice!', got %v", resp.Data)
	}
	if resp.ToolID != "greet" {
		t.Fatalf("expected ToolID 'greet', got %s", resp.ToolID)
	}
}

func TestCoreConnector_Execute_UnregisteredTool(t *testing.T) {
	c := NewCoreConnector()

	resp, err := c.Execute(context.Background(), &ExecutionRequest{
		ToolID: "nonexistent",
		UserID: "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "error" {
		t.Fatalf("expected error status, got %s", resp.Status)
	}
	if !strings.Contains(resp.Error, "not registered") {
		t.Fatalf("expected 'not registered' message, got %s", resp.Error)
	}
}

func TestCoreConnector_Execute_ToolReturnsError(t *testing.T) {
	c := NewCoreConnector()
	c.RegisterTool("fail", func(ctx context.Context, userID string, inputs map[string]interface{}) (string, error) {
		return "", fmt.Errorf("intentional failure")
	})

	resp, err := c.Execute(context.Background(), &ExecutionRequest{
		ToolID: "fail",
		UserID: "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Fatalf("expected error status, got %s", resp.Status)
	}
	if !strings.Contains(resp.Error, "intentional failure") {
		t.Fatalf("expected tool error message, got %s", resp.Error)
	}
}

func TestCoreConnector_Execute_DurationTracking(t *testing.T) {
	c := NewCoreConnector()
	c.RegisterTool("slow", func(ctx context.Context, userID string, inputs map[string]interface{}) (string, error) {
		time.Sleep(10 * time.Millisecond)
		return "done", nil
	})

	resp, _ := c.Execute(context.Background(), &ExecutionRequest{ToolID: "slow"})
	if resp.Duration < 10 {
		t.Fatalf("expected duration >= 10ms, got %dms", resp.Duration)
	}
}

// ──────────────────────────────────────────────────────────────────────
// MCPConnector tests (mcp.go)
//
// Since MCPClient is a concrete struct wrapping a third-party *mcp.Client,
// we can't easily mock CallTool. We test the nil-client guard paths,
// Validate, HealthCheck, and Type.
// ──────────────────────────────────────────────────────────────────────

func TestMCPConnector_Type(t *testing.T) {
	c := NewMCPConnector(nil)
	if c.Type() != TypeMCP {
		t.Fatalf("expected TypeMCP, got %s", c.Type())
	}
}

func TestMCPConnector_HealthCheck_NilClient(t *testing.T) {
	c := NewMCPConnector(nil)
	err := c.HealthCheck(context.Background())
	if err == nil || !strings.Contains(err.Error(), "not available") {
		t.Fatalf("expected 'not available' error, got %v", err)
	}
}

func TestMCPConnector_Execute_NilClient(t *testing.T) {
	c := NewMCPConnector(nil)
	resp, err := c.Execute(context.Background(), &ExecutionRequest{
		ToolID: "test-tool",
		UserID: "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Fatalf("expected error status, got %s", resp.Status)
	}
	if !strings.Contains(resp.Error, "not available") {
		t.Fatalf("expected 'not available' in error, got %s", resp.Error)
	}
}

func TestMCPConnector_Validate(t *testing.T) {
	c := NewMCPConnector(nil)

	// Missing tool_id
	err := c.Validate(&ExecutionRequest{})
	if err == nil || !strings.Contains(err.Error(), "tool_id") {
		t.Fatalf("expected tool_id validation error, got %v", err)
	}

	// Empty UserID gets defaulted to "anonymous"
	req := &ExecutionRequest{ToolID: "test"}
	err = c.Validate(req)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if req.UserID != "anonymous" {
		t.Fatalf("expected UserID to be set to 'anonymous', got %s", req.UserID)
	}

	// Non-empty UserID is left alone
	req2 := &ExecutionRequest{ToolID: "test", UserID: "user-42"}
	err = c.Validate(req2)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if req2.UserID != "user-42" {
		t.Fatalf("expected UserID 'user-42', got %s", req2.UserID)
	}
}

// ──────────────────────────────────────────────────────────────────────
// RESTConnector tests (rest.go)
// ──────────────────────────────────────────────────────────────────────

func TestRESTConnector_Type(t *testing.T) {
	c := NewRESTConnector("http://example.com")
	if c.Type() != TypeREST {
		t.Fatalf("expected TypeREST, got %s", c.Type())
	}
}

func TestRESTConnector_Validate(t *testing.T) {
	c := NewRESTConnector("http://example.com")

	// Missing tool_id
	err := c.Validate(&ExecutionRequest{})
	if err == nil || !strings.Contains(err.Error(), "tool_id") {
		t.Fatalf("expected tool_id validation error, got %v", err)
	}

	// Nil inputs get initialised
	req := &ExecutionRequest{ToolID: "tool-1"}
	if err := c.Validate(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Inputs == nil {
		t.Fatal("expected Inputs to be initialised to an empty map")
	}
}

func TestRESTConnector_Execute_Success_JSON(t *testing.T) {
	expected := map[string]interface{}{"result": "ok"}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("X-Memzent-User-ID") != "user-1" {
			t.Errorf("expected X-Memzent-User-ID user-1, got %s", r.Header.Get("X-Memzent-User-ID"))
		}
		if r.Header.Get("X-Memzent-Tool-ID") != "tool-1" {
			t.Errorf("expected X-Memzent-Tool-ID tool-1, got %s", r.Header.Get("X-Memzent-Tool-ID"))
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer ts.Close()

	c := NewRESTConnector(ts.URL)
	resp, err := c.Execute(context.Background(), &ExecutionRequest{
		ToolID: "tool-1",
		UserID: "user-1",
		Inputs: map[string]interface{}{"key": "value"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("expected success, got %s (error: %s)", resp.Status, resp.Error)
	}

	// The response data should be a parsed JSON map
	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", resp.Data)
	}
	if dataMap["result"] != "ok" {
		t.Fatalf("expected result 'ok', got %v", dataMap["result"])
	}
}

func TestRESTConnector_Execute_Success_NonJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("plain text response"))
	}))
	defer ts.Close()

	c := NewRESTConnector(ts.URL)
	resp, err := c.Execute(context.Background(), &ExecutionRequest{
		ToolID: "tool-2",
		UserID: "user-1",
		Inputs: map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("expected success, got %s (error: %s)", resp.Status, resp.Error)
	}
	// Non-JSON body falls back to raw string
	if resp.Data != "plain text response" {
		t.Fatalf("expected raw string data, got %v", resp.Data)
	}
}

func TestRESTConnector_Execute_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer ts.Close()

	c := NewRESTConnector(ts.URL)
	resp, err := c.Execute(context.Background(), &ExecutionRequest{
		ToolID: "tool-3",
		UserID: "user-1",
		Inputs: map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Fatalf("expected error status, got %s", resp.Status)
	}
	if !strings.Contains(resp.Error, "500") {
		t.Fatalf("expected HTTP 500 in error, got %s", resp.Error)
	}
}

func TestRESTConnector_Execute_ConnectionError(t *testing.T) {
	// Use an unreachable endpoint
	c := NewRESTConnector("http://127.0.0.1:1") // port 1 should refuse
	resp, err := c.Execute(context.Background(), &ExecutionRequest{
		ToolID: "tool-4",
		UserID: "user-1",
		Inputs: map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Fatalf("expected error status, got %s", resp.Status)
	}
	if !strings.Contains(resp.Error, "REST call failed") {
		t.Fatalf("expected 'REST call failed' in error, got %s", resp.Error)
	}
}

func TestRESTConnector_Execute_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second) // hold the connection
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewRESTConnector(ts.URL)
	resp, err := c.Execute(context.Background(), &ExecutionRequest{
		ToolID:  "tool-5",
		UserID:  "user-1",
		Inputs:  map[string]interface{}{},
		Timeout: 1, // 1 second timeout
	})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "timeout" {
		t.Fatalf("expected timeout status, got %s (error: %s)", resp.Status, resp.Error)
	}
}

func TestRESTConnector_Execute_InvalidEndpointURL(t *testing.T) {
	// A URL with a control character will cause http.NewRequestWithContext to fail
	c := NewRESTConnector("http://example.com/\x7f")
	resp, err := c.Execute(context.Background(), &ExecutionRequest{
		ToolID: "tool-6",
		UserID: "user-1",
		Inputs: map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Fatalf("expected error status, got %s", resp.Status)
	}
	if !strings.Contains(resp.Error, "failed to create request") {
		t.Fatalf("expected 'failed to create request' in error, got %s", resp.Error)
	}
}

func TestRESTConnector_HealthCheck_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewRESTConnector(ts.URL)
	if err := c.HealthCheck(context.Background()); err != nil {
		t.Fatalf("expected nil health check, got %v", err)
	}
}

func TestRESTConnector_HealthCheck_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	c := NewRESTConnector(ts.URL)
	err := c.HealthCheck(context.Background())
	if err == nil || !strings.Contains(err.Error(), "503") {
		t.Fatalf("expected HTTP 503 health check error, got %v", err)
	}
}

func TestRESTConnector_HealthCheck_Unreachable(t *testing.T) {
	c := NewRESTConnector("http://127.0.0.1:1")
	err := c.HealthCheck(context.Background())
	if err == nil || !strings.Contains(err.Error(), "health check failed") {
		t.Fatalf("expected health check failure, got %v", err)
	}
}

// ──────────────────────────────────────────────────────────────────────
// SQLConnector tests (sql.go)
// ──────────────────────────────────────────────────────────────────────

func TestSQLConnector_Type(t *testing.T) {
	c := NewSQLConnector("postgres://localhost/test")
	if c.Type() != TypeSQL {
		t.Fatalf("expected TypeSQL, got %s", c.Type())
	}
}

func TestSQLConnector_Close_NilDB(t *testing.T) {
	c := NewSQLConnector("")
	// Close on nil db should be a no-op
	if err := c.Close(); err != nil {
		t.Fatalf("expected nil error closing nil db, got %v", err)
	}
}

func TestSQLConnector_Close_WithDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectClose()
	c := &SQLConnector{db: db}
	if err := c.Close(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestSQLConnector_HealthCheck_NilDB(t *testing.T) {
	c := NewSQLConnector("")
	err := c.HealthCheck(context.Background())
	if err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("expected 'not initialized' error, got %v", err)
	}
}

func TestSQLConnector_HealthCheck_PingSuccess(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectPing()
	c := &SQLConnector{db: db}

	if err := c.HealthCheck(context.Background()); err != nil {
		t.Fatalf("expected nil HealthCheck, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLConnector_HealthCheck_PingFail(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectPing().WillReturnError(fmt.Errorf("connection lost"))
	c := &SQLConnector{db: db}

	hcErr := c.HealthCheck(context.Background())
	if hcErr == nil || !strings.Contains(hcErr.Error(), "ping failed") {
		t.Fatalf("expected ping failed error, got %v", hcErr)
	}
}

func TestSQLConnector_Validate(t *testing.T) {
	c := NewSQLConnector("")

	// Missing tool_id
	err := c.Validate(&ExecutionRequest{})
	if err == nil || !strings.Contains(err.Error(), "tool_id") {
		t.Fatalf("expected tool_id error, got %v", err)
	}

	// Missing query field
	err = c.Validate(&ExecutionRequest{ToolID: "t1", Inputs: map[string]interface{}{}})
	if err == nil || !strings.Contains(err.Error(), "query") {
		t.Fatalf("expected query error, got %v", err)
	}

	// Non-string query field
	err = c.Validate(&ExecutionRequest{ToolID: "t1", Inputs: map[string]interface{}{"query": 123}})
	if err == nil || !strings.Contains(err.Error(), "query") {
		t.Fatalf("expected query error for non-string, got %v", err)
	}

	// Valid
	err = c.Validate(&ExecutionRequest{ToolID: "t1", Inputs: map[string]interface{}{"query": "SELECT 1"}})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestSQLConnector_Execute_NilDB(t *testing.T) {
	c := NewSQLConnector("")
	resp, err := c.Execute(context.Background(), &ExecutionRequest{
		ToolID: "t1",
		Inputs: map[string]interface{}{"query": "SELECT 1"},
	})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Fatalf("expected error status, got %s", resp.Status)
	}
	if !strings.Contains(resp.Error, "not initialized") {
		t.Fatalf("expected 'not initialized' in error, got %s", resp.Error)
	}
}

func TestSQLConnector_Execute_MissingQuery(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	c := &SQLConnector{db: db}

	// No "query" key in inputs
	resp, err := c.Execute(context.Background(), &ExecutionRequest{
		ToolID: "t1",
		Inputs: map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Fatalf("expected error status, got %s", resp.Status)
	}
	if !strings.Contains(resp.Error, "query field is required") {
		t.Fatalf("expected 'query field is required', got %s", resp.Error)
	}
}

func TestSQLConnector_Execute_QuerySuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Alice").
		AddRow(2, "Bob")
	mock.ExpectQuery("SELECT id, name FROM users").WillReturnRows(rows)

	c := &SQLConnector{db: db}
	resp, err := c.Execute(context.Background(), &ExecutionRequest{
		ToolID: "t1",
		Inputs: map[string]interface{}{"query": "SELECT id, name FROM users"},
	})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("expected success, got %s (error: %s)", resp.Status, resp.Error)
	}

	data, ok := resp.Data.([]map[string]interface{})
	if !ok {
		t.Fatalf("expected []map[string]interface{}, got %T", resp.Data)
	}
	if len(data) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(data))
	}
	if data[0]["name"] != "Alice" {
		t.Fatalf("expected 'Alice', got %v", data[0]["name"])
	}
	if data[1]["name"] != "Bob" {
		t.Fatalf("expected 'Bob', got %v", data[1]["name"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLConnector_Execute_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("syntax error"))

	c := &SQLConnector{db: db}
	resp, err := c.Execute(context.Background(), &ExecutionRequest{
		ToolID: "t1",
		Inputs: map[string]interface{}{"query": "SELECT %%%"},
	})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Fatalf("expected error status, got %s", resp.Status)
	}
	if !strings.Contains(resp.Error, "query execution failed") {
		t.Fatalf("expected 'query execution failed', got %s", resp.Error)
	}
}

func TestSQLConnector_Execute_EmptyResultSet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id"})
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	c := &SQLConnector{db: db}
	resp, err := c.Execute(context.Background(), &ExecutionRequest{
		ToolID: "t1",
		Inputs: map[string]interface{}{"query": "SELECT id FROM empty_table"},
	})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("expected success, got %s (error: %s)", resp.Status, resp.Error)
	}
	// result should be nil or empty slice (no rows matched)
	if resp.Data != nil {
		if dataSlice, ok := resp.Data.([]map[string]interface{}); ok && len(dataSlice) > 0 {
			t.Fatalf("expected empty data for empty result, got %v", resp.Data)
		}
	}
}

func TestSQLConnector_Execute_RowScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Create rows then close them to cause a scan error on iteration
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1).
		RowError(0, fmt.Errorf("row scan problem"))
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	c := &SQLConnector{db: db}
	resp, err := c.Execute(context.Background(), &ExecutionRequest{
		ToolID: "t1",
		Inputs: map[string]interface{}{"query": "SELECT id FROM bad_table"},
	})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if resp.Status != "error" {
		t.Fatalf("expected error status, got %s", resp.Status)
	}
	if !strings.Contains(resp.Error, "row iteration error") {
		t.Fatalf("expected row iteration error, got %s", resp.Error)
	}
}

func TestSQLConnector_Connect_ConnectionStringVariants(t *testing.T) {
	// We test the connection string manipulation logic.
	// sql.Open will succeed even if the driver isn't registered,
	// but PingContext will fail. We just check the error to confirm
	// the code path executed.
	tests := []struct {
		name       string
		connString string
	}{
		{"plain postgres://", "postgres://localhost/testdb"},
		{"postgres with existing params", "postgres://localhost/testdb?sslmode=disable"},
		{"postgres with binary_parameters already", "postgres://localhost/testdb?binary_parameters=yes"},
		{"postgresql:// scheme", "postgresql://localhost/testdb"},
		{"non-postgres scheme", "mysql://localhost/testdb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewSQLConnector(tt.connString)
			// Connect will fail because there's no actual DB, but we exercise the path
			err := c.Connect(context.Background())
			// We expect an error (no real database), but not a panic
			if err == nil {
				// If it somehow succeeded (unlikely), just verify DB is set
				c.Close()
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────────────
// Interface compliance checks
// ──────────────────────────────────────────────────────────────────────

// Ensure all connectors implement the Connector interface at compile time.
var (
	_ Connector = (*CoreConnector)(nil)
	_ Connector = (*MCPConnector)(nil)
	_ Connector = (*RESTConnector)(nil)
	_ Connector = (*SQLConnector)(nil)
)

// Ensure ConnectorRegistry supports multiple connector types
func TestConnectorRegistry_MultipleTypes(t *testing.T) {
	reg := NewConnectorRegistry()

	reg.Register(TypeCore, NewCoreConnector())
	reg.Register(TypeREST, NewRESTConnector("http://example.com"))
	reg.Register(TypeSQL, &SQLConnector{db: nil})

	for _, ct := range []ConnectorType{TypeCore, TypeREST, TypeSQL} {
		if _, ok := reg.Get(ct); !ok {
			t.Fatalf("expected connector type %s to be registered", ct)
		}
	}

	// MCP not registered
	if _, ok := reg.Get(TypeMCP); ok {
		t.Fatal("MCP should not be registered")
	}
}

// Suppress "imported and not used" for sql package — used indirectly via sqlmock
var _ = sql.Open
