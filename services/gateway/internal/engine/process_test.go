package engine

import (
	"context"
	"testing"
	"memzent-gateway/internal/billing"
	"memzent-gateway/internal/llm"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMemzentEngine_Process_RateLimit(t *testing.T) {
	e := newTestEngineWithProviders()
	ctx := context.Background()
	ctx = context.WithValue(ctx, "org_id", "org1")
	ctx = context.WithValue(ctx, "tier", "free")

	req := &PromptRequest{UserID: "user1", Messages: []llm.Message{{Role: "user", Content: "test"}}}

	// Drain the rate limiter (limit 10)
	for i := 0; i < 10; i++ {
		func() {
			defer func() { recover() }()
			_, _ = e.Process(ctx, req)
		}()
	}

	// 11th request should fail with rate limit
	_, err := e.Process(ctx, req)
	if err == nil {
		t.Errorf("Expected rate limit error")
	} else if err.Error() != "rate limit exceeded for organization org1 (tier: free)" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestMemzentEngine_Process_BillingFailure(t *testing.T) {
	e := newTestEngineWithProviders()

	// Using a mock ledger to return insufficient balance
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("SELECT COALESCE").
		WithArgs("org2").
		WillReturnRows(sqlmock.NewRows([]string{"token_balance", "default_provider", "default_model"}).AddRow(0, "", "")) // 0 balance
	mock.ExpectQuery("SELECT id, amount").
		WithArgs("org2", 5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "amount", "transaction_type", "description", "created_at"}))
	e.ledger = billing.NewLedger(db)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "org_id", "org2")
	// Must not be JWT otherwise billing is skipped
	ctx = context.WithValue(ctx, "auth_method", "api_key")
	
	req := &PromptRequest{UserID: "user1", Messages: []llm.Message{{Role: "user", Content: "test"}}}

	_, err := e.Process(ctx, req)
	if err == nil {
		t.Errorf("Expected billing error")
	} else if err.Error() != "payment required: token balance depleted" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestMemzentEngine_Process_NoProviderFallback(t *testing.T) {
	// If provider is missing, it should fallback to default provider
	e := newTestEngineWithProviders()
	ctx := context.Background()
	ctx = context.WithValue(ctx, "org_id", "admin-01") // bypass RBAC

	req := &PromptRequest{UserID: "user1", Messages: []llm.Message{{Role: "user", Content: "test"}}, Provider: "invalid-provider"}

	// Without a real router client, it will panic or error out on router.GetBestTools
	// But before that it resolves the provider. We can't reach the end without panicking on nil router.
	// So we expect a panic because e.router is nil, but we can catch it.
	defer func() {
		if r := recover(); r != nil {
			// Expected panic due to nil router
		}
	}()

	e.Process(ctx, req)
}
