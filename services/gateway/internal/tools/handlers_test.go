package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/grpc"
	"memzent-gateway/internal/router"
)

type mockSemanticRouterClient struct {
	router.SemanticRouterClient
	RegisterToolFn func(ctx context.Context, in *router.RegisterToolRequest, opts ...grpc.CallOption) (*router.RegisterToolResponse, error)
}

func (m *mockSemanticRouterClient) RegisterTool(ctx context.Context, in *router.RegisterToolRequest, opts ...grpc.CallOption) (*router.RegisterToolResponse, error) {
	if m.RegisterToolFn != nil {
		return m.RegisterToolFn(ctx, in, opts...)
	}
	return &router.RegisterToolResponse{Success: true}, nil
}

func setUnexportedField(target interface{}, fieldName string, value interface{}) {
	rv := reflect.ValueOf(target).Elem()
	field := rv.FieldByName(fieldName)
	ptr := unsafe.Pointer(field.UnsafeAddr())
	reflect.NewAt(field.Type(), ptr).Elem().Set(reflect.ValueOf(value))
}

func TestHandleRegisterTool_Forbidden(t *testing.T) {
	db, _, _ := sqlmock.New()
	registry := NewRegistry(db)
	handler := HandleRegisterTool(registry, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/v1/tools", nil)
	// Inject non-admin role
	ctx := context.WithValue(req.Context(), "user_role", "viewer")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d", rec.Code)
	}
}

func TestHandleRegisterTool_InvalidBody(t *testing.T) {
	db, _, _ := sqlmock.New()
	registry := NewRegistry(db)
	handler := HandleRegisterTool(registry, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/v1/tools", strings.NewReader("bad_json{"))
	ctx := context.WithValue(req.Context(), "user_role", "admin")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rec.Code)
	}
}

func TestHandleRegisterTool_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	registry := NewRegistry(db)

	// Set up mock router client using reflection
	mockRouter := &mockSemanticRouterClient{
		RegisterToolFn: func(ctx context.Context, in *router.RegisterToolRequest, opts ...grpc.CallOption) (*router.RegisterToolResponse, error) {
			if in.Id != "test-tool" || in.Name != "Test Tool" {
				return nil, fmt.Errorf("unexpected args")
			}
			return &router.RegisterToolResponse{Success: true}, nil
		},
	}
	rc := &router.RouterClient{}
	setUnexportedField(rc, "client", mockRouter)

	handler := HandleRegisterTool(registry, rc, nil)

	// Prepare payload
	payload := RegisterRequest{
		ID:            "test-tool",
		Name:          "Test Tool",
		Description:   "Description",
		ConnectorType: "mcp",
		Endpoint:      "endpoint",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/tools", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), "user_role", "admin")
	ctx = context.WithValue(ctx, "org_id", "org1")
	req = req.WithContext(ctx)

	// Registry insert mock
	mock.ExpectExec("INSERT INTO tools").
		WillReturnResult(sqlmock.NewResult(1, 1))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Wait briefly for the async goroutine calling routerClient.RegisterTool to run
	time.Sleep(10 * time.Millisecond)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestHandleDisableTool(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	registry := NewRegistry(db)
	handler := HandleDisableTool(registry)

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/tools/test-tool", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405 Method Not Allowed, got %d", rec.Code)
		}
	})

	t.Run("Forbidden", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/v1/tools/test-tool", nil)
		ctx := context.WithValue(req.Context(), "user_role", "viewer")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden, got %d", rec.Code)
		}
	})

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/v1/tools/test-tool", nil)
		ctx := context.WithValue(req.Context(), "user_role", "admin")
		req = req.WithContext(ctx)

		mock.ExpectExec("UPDATE tools SET enabled = false").
			WithArgs(sqlmock.AnyArg(), "test-tool").
			WillReturnResult(sqlmock.NewResult(1, 1))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", rec.Code)
		}

		var res map[string]string
		json.Unmarshal(rec.Body.Bytes(), &res)
		if res["status"] != "disabled" || res["tool_id"] != "test-tool" {
			t.Errorf("unexpected body: %s", rec.Body.String())
		}
	})
}

func TestHandleSyncTools(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	registry := NewRegistry(db)

	mockRouter := &mockSemanticRouterClient{
		RegisterToolFn: func(ctx context.Context, in *router.RegisterToolRequest, opts ...grpc.CallOption) (*router.RegisterToolResponse, error) {
			return &router.RegisterToolResponse{Success: true}, nil
		},
	}
	rc := &router.RouterClient{}
	setUnexportedField(rc, "client", mockRouter)

	handler := HandleSyncTools(registry, rc, nil)

	req := httptest.NewRequest(http.MethodPost, "/v1/tools/sync", nil)
	ctx := context.WithValue(req.Context(), "user_role", "admin")
	req = req.WithContext(ctx)

	// Mock refresh query returning one drifted tool
	now := time.Now()
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

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["status"] != "success" || resp["tools_synced"].(float64) != 1 {
		t.Errorf("unexpected body payload: %+v", resp)
	}
}

func TestHandleRegistryStatus(t *testing.T) {
	db, _, _ := sqlmock.New()
	registry := NewRegistry(db)
	handler := HandleRegistryStatus(registry)

	req := httptest.NewRequest(http.MethodGet, "/v1/tools/status", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rec.Code)
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["status"] != "healthy" {
		t.Errorf("unexpected status: %+v", resp)
	}
}

func TestToolToAPI(t *testing.T) {
	tool := &Tool{
		ID:             "tool-01",
		Name:           "Tool",
		Description:    "Desc",
		ConnectorType:  ConnectorMCP,
		TimeoutSeconds: 30,
		InputSchema:    map[string]interface{}{"in": "schema"},
		OutputSchema:   map[string]interface{}{"out": "schema"},
	}

	apiTool := ToolToAPI(tool)
	if apiTool.ID != "tool-01" || apiTool.Provider != "Memzent-MCP" || apiTool.Status != "online" {
		t.Errorf("unexpected converted API format: %+v", apiTool)
	}
}
