package engine

import (
	"context"
	"testing"
	"time"

	"memzent-gateway/internal/auth"
	"memzent-gateway/internal/billing"
	"memzent-gateway/internal/llm"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMemzentEngine_Process_RateLimit(t *testing.T) {
	e := newTestEngineWithProviders()
	e.router = &mockRouter{}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "org_id", "org1")
	ctx = context.WithValue(ctx, "tier", "free")

	req := &PromptRequest{UserID: "user1", Messages: []llm.Message{{Role: "user", Content: "test"}}}

	// Rate limiting requires Valkey (distributed via cache.RateLimit).
	// Without a cache, rate limiting is skipped and the request proceeds.
	_, err := e.Process(ctx, req)
	if err != nil && err.Error() == "rate limit exceeded for organization org1 (tier: free)" {
		t.Errorf("Rate limit should not fire without Valkey cache configured")
	}
}

func TestMemzentEngine_Process_BillingFailure(t *testing.T) {
	e := newTestEngineWithProviders()

	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("SELECT COALESCE").
		WithArgs("org2").
		WillReturnRows(sqlmock.NewRows([]string{"token_balance", "default_provider", "default_model"}).AddRow(0, "", ""))
	mock.ExpectQuery("SELECT id, amount").
		WithArgs("org2", 5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "amount", "transaction_type", "description", "created_at"}))
	e.ledger = billing.NewLedger(db)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "org_id", "org2")
	ctx = context.WithValue(ctx, "auth_method", "api_key")

	req := &PromptRequest{UserID: "user1", Messages: []llm.Message{{Role: "user", Content: "test"}}}

	_, err := e.Process(ctx, req)
	if err == nil {
		t.Errorf("Expected billing error")
	} else if err.Error() != "payment required: token balance depleted" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestMemzentEngine_Process_CacheHit(t *testing.T) {
	e := newTestEngineWithProviders()
	e.cacheTTL = time.Hour

	prompt := "what is memzent"
	cacheKey := e.buildCacheKey("org1", "p", "default1", prompt)
	e.cache = newMockCache(map[string]string{cacheKey: "cached answer"})

	ctx := context.Background()
	ctx = context.WithValue(ctx, "org_id", "org1")
	req := &PromptRequest{UserID: "user1", Messages: []llm.Message{{Role: "user", Content: prompt}}}

	resp, err := e.Process(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Cached {
		t.Error("expected cached response")
	}
	if resp.Text != "cached answer" {
		t.Errorf("expected cached answer, got %q", resp.Text)
	}
}

func TestMemzentEngine_Process_RBACDeny(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("denied-org", "chat:execute").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	e := newTestEngineWithProviders()
	e.cache = newMockCache(nil)
	e.router = &mockRouter{}
	e.rbac = auth.NewRBACClientForTestWithEnv(db, "production")

	ctx := context.Background()
	ctx = context.WithValue(ctx, "org_id", "denied-org")
	req := &PromptRequest{UserID: "user1", Messages: []llm.Message{{Role: "user", Content: "hello"}}}

	_, err = e.Process(ctx, req)
	if err == nil {
		t.Fatal("expected RBAC deny error")
	}
	if err.Error() != "unauthorized: insufficient scope" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMemzentEngine_Process_CacheMissWithMockRouter(t *testing.T) {
	e := newTestEngineWithProviders()
	e.cache = newMockCache(nil)
	e.router = &mockRouter{}
	e.cacheTTL = time.Hour

	ctx := context.Background()
	ctx = context.WithValue(ctx, "org_id", "org1")
	req := &PromptRequest{UserID: "user1", Messages: []llm.Message{{Role: "user", Content: "generate this"}}}

	resp, err := e.Process(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Cached {
		t.Error("expected cache miss path")
	}
	if resp.Text != "mock response" {
		t.Errorf("expected mock LLM response, got %q", resp.Text)
	}

	resp2, err := e.Process(ctx, req)
	if err != nil {
		t.Fatalf("second request error: %v", err)
	}
	if !resp2.Cached {
		t.Error("expected cache hit on second request")
	}
	if resp2.Text != "mock response" {
		t.Errorf("expected cached mock response, got %q", resp2.Text)
	}
}
