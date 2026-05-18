package engine

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"memzent-gateway/internal/auth"
	"memzent-gateway/internal/billing"
	"memzent-gateway/internal/llm"

	"github.com/DATA-DOG/go-sqlmock"
)

// Mock Provider for testing
type mockProvider struct {
	name     string
	metadata llm.ProviderMetadata
	res      string
	err      error
}

func (m *mockProvider) GetProviderName() string { return m.name }
func (m *mockProvider) GetMetadata() llm.ProviderMetadata { return m.metadata }
func (m *mockProvider) Generate(ctx context.Context, messages []llm.Message, tools []any, model string) (string, *llm.TokenUsage, error) {
	if m.err != nil {
		return "", nil, m.err
	}
	return m.res, &llm.TokenUsage{PromptTokens: 10, CompletionTokens: 20, TotalTokens: 30}, nil
}

func newTestEngineWithProviders() *MemzentEngine {
	providers := map[string]llm.Provider{
		"mock1": &mockProvider{name: "mock1", metadata: llm.ProviderMetadata{Name: "mock1", DefaultModel: "default1"}, res: "mock response"},
		"mock2": &mockProvider{name: "mock2", metadata: llm.ProviderMetadata{Name: "mock2", DefaultModel: "default2"}, res: "mock response"},
	}
	return &MemzentEngine{
		providers:       providers,
		defaultProvider: "mock1",
	}
}

func TestEngine_ActiveProviderNames(t *testing.T) {
	e := newTestEngineWithProviders()
	names := e.ActiveProviderNames()
	if len(names) != 2 {
		t.Errorf("expected 2 provider names, got %d", len(names))
	}
}

func TestEngine_GetProviderMetadata(t *testing.T) {
	e := newTestEngineWithProviders()
	meta := e.GetProviderMetadata()
	if len(meta) != 2 {
		t.Errorf("expected 2 provider metadata entries, got %d", len(meta))
	}
}

func TestEngine_DefaultProviderName(t *testing.T) {
	e := newTestEngineWithProviders()
	if e.DefaultProviderName() != "mock1" {
		t.Errorf("expected mock1, got %s", e.DefaultProviderName())
	}

	e.defaultProvider = "unknown"
	if e.DefaultProviderName() != "unknown" {
		t.Errorf("expected unknown, got %s", e.DefaultProviderName())
	}
}

func TestEngine_ProviderCount(t *testing.T) {
	e := newTestEngineWithProviders()
	if e.ProviderCount() != 2 {
		t.Errorf("expected 2, got %d", e.ProviderCount())
	}
}

func TestEngine_GetStats(t *testing.T) {
	e := newTestEngineWithProviders()
	// Set initial stats via process simulation
	reqCounter, _ := e.orgRequests.LoadOrStore("org1", new(atomic.Uint64))
	reqCounter.(*atomic.Uint64).Store(42)
	hitCounter, _ := e.orgHits.LoadOrStore("org1", new(atomic.Uint64))
	hitCounter.(*atomic.Uint64).Store(10)

	reqs, hits := e.GetStats("org1")
	if reqs != 42 || hits != 10 {
		t.Errorf("expected 42 reqs, 10 hits, got %d reqs, %d hits", reqs, hits)
	}

	reqs, hits = e.GetStats("org2")
	if reqs != 0 || hits != 0 {
		t.Errorf("expected 0 reqs, 0 hits for new org, got %d reqs, %d hits", reqs, hits)
	}
}

// ---------------------------------------------------------------------------
// DB Persistent Cache Tests
// ---------------------------------------------------------------------------

func TestEngine_getPersistentCache(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	rbacClient := auth.NewRBACClientForTest(db)
	e := &MemzentEngine{
		rbac: rbacClient,
	}

	ctx := context.Background()

	// Hit
	mock.ExpectQuery("SELECT response FROM persistent_cache").
		WithArgs("test-key").
		WillReturnRows(sqlmock.NewRows([]string{"response"}).AddRow("cached response"))

	resp, err := e.getPersistentCache(ctx, "test-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp != "cached response" {
		t.Errorf("expected 'cached response', got '%s'", resp)
	}

	// Miss (ErrNoRows)
	mock.ExpectQuery("SELECT response FROM persistent_cache").
		WithArgs("miss-key").
		WillReturnError(sql.ErrNoRows)

	resp, err = e.getPersistentCache(ctx, "miss-key")
	if err != nil {
		t.Errorf("unexpected error for miss: %v", err)
	}
	if resp != "" {
		t.Errorf("expected empty response for miss, got '%s'", resp)
	}

	// Query error
	mock.ExpectQuery("SELECT response FROM persistent_cache").
		WithArgs("err-key").
		WillReturnError(fmt.Errorf("db err"))

	_, err = e.getPersistentCache(ctx, "err-key")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestEngine_setPersistentCache(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	rbacClient := auth.NewRBACClientForTest(db)
	e := &MemzentEngine{
		rbac: rbacClient,
	}

	ctx := context.Background()

	mock.ExpectExec("INSERT INTO persistent_cache").
		WithArgs("org1", "test-key", "response text", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	e.setPersistentCache(ctx, "org1", "test-key", "response text", 5*time.Minute)

	// Sleep slightly to allow background goroutine to execute
	time.Sleep(10 * time.Millisecond)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestEngine_WarmCache(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	rbacClient := auth.NewRBACClientForTest(db)
	e := &MemzentEngine{
		rbac: rbacClient,
		// Skipping cache field so it skips gracefully (avoiding panic)
		// Or we can just let it skip. Actually let's test the skip behavior.
	}

	ctx := context.Background()

	// Should skip if cache is nil
	e.WarmCache(ctx)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// ---------------------------------------------------------------------------
// Billing Cache Hit Charge
// ---------------------------------------------------------------------------

func TestEngine_chargeCacheHit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	ledger := billing.NewLedger(db)
	costCalc := billing.NewCostCalculator()
	costCalc.Rates["mock1:default"] = 0.1
	costCalc.OutputRates["mock1:default"] = 0.1
	
	e := &MemzentEngine{
		ledger:          ledger,
		costCalc:        costCalc,
		defaultProvider: "mock1",
	}

	ctx := context.Background()

	// 100 char prompt = 25 tokens. 25 * 0.1 / 1000 = 0.0025 * 0.1 (cache discount) = 0.00025
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE organizations").
		WithArgs(sqlmock.AnyArg(), "org1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO billing_ledger").
		WithArgs("org1", sqlmock.AnyArg(), "cache_hit", "Semantic Cache Hit Discount").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	prompt := string(make([]byte, 100)) // 100 chars
	e.chargeCacheHit(ctx, "org1", "mock1", "default", prompt)

	// Sleep slightly to allow background goroutine to execute
	time.Sleep(10 * time.Millisecond)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
