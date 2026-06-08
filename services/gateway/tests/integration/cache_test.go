//go:build integration

package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"memzent-gateway/internal/cache"
	"memzent-gateway/internal/engine"
	"memzent-gateway/internal/llm"
	"memzent-gateway/internal/router"
)

type stubRouter struct{}

func (stubRouter) GetBestTools(context.Context, string, string, []string) ([]*router.Tool, string, string, string, error) {
	return nil, "", "", "", nil
}
func (stubRouter) RegisterTool(context.Context, string, string, string, string) (bool, error) {
	return true, nil
}
func (stubRouter) PlanToolChain(context.Context, string, string, []string) ([]*router.ToolStep, float32, error) {
	return nil, 0, nil
}
func (stubRouter) StoreMemory(context.Context, string, string, string) (bool, error) { return true, nil }
func (stubRouter) QueryMemory(context.Context, string, string, string, float32) ([]*router.MemoryHit, error) {
	return nil, nil
}
func (stubRouter) Close() {}

type stubProvider struct{}

func (stubProvider) GetProviderName() string { return "stub" }
func (stubProvider) GetMetadata() llm.ProviderMetadata {
	return llm.ProviderMetadata{Name: "stub", DefaultModel: "stub-model"}
}
func (stubProvider) Generate(context.Context, []llm.Message, []any, string) (string, *llm.TokenUsage, error) {
	return "integration-response", &llm.TokenUsage{PromptTokens: 1, CompletionTokens: 2, TotalTokens: 3}, nil
}

func valkeyAddr() string {
	if v := os.Getenv("VALKEY_ADDR"); v != "" {
		return v
	}
	return "localhost:6379"
}

func TestValkeyCacheRoundTrip(t *testing.T) {
	ctx := context.Background()
	c, err := cache.NewMemzentCache(ctx, valkeyAddr())
	if err != nil {
		t.Skipf("valkey unavailable at %s: %v", valkeyAddr(), err)
	}
	defer c.Close()

	if err := c.Ping(ctx); err != nil {
		t.Skipf("valkey ping failed: %v", err)
	}

	key := "integration:test:" + time.Now().Format("20060102150405")
	if err := c.SetResult(ctx, key, "hello-valkey", time.Minute); err != nil {
		t.Fatalf("set: %v", err)
	}
	got, err := c.GetSemanticResult(ctx, key)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != "hello-valkey" {
		t.Fatalf("got %q, want hello-valkey", got)
	}
}

func TestEngineCacheMissThenHit_Valkey(t *testing.T) {
	ctx := context.Background()
	c, err := cache.NewMemzentCache(ctx, valkeyAddr())
	if err != nil {
		t.Skipf("valkey unavailable: %v", err)
	}
	defer c.Close()
	if err := c.Ping(ctx); err != nil {
		t.Skipf("valkey ping failed: %v", err)
	}

	providers := map[string]llm.Provider{"stub": stubProvider{}}
	e := engine.NewMemzentEngine(
		c,
		stubRouter{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		providers,
		"stub",
		0.65,
		time.Minute,
		nil,
		nil,
		nil,
		nil,
	)

	prompt := "integration cache miss then hit " + time.Now().Format(time.RFC3339Nano)
	req := &engine.PromptRequest{
		UserID:   "integration-user",
		Messages: []llm.Message{{Role: "user", Content: prompt}},
	}

	runCtx := context.WithValue(ctx, "org_id", "integration-org")

	resp, err := e.Process(runCtx, req, nil)
	if err != nil {
		t.Fatalf("first process: %v", err)
	}
	if resp.Cached {
		t.Fatal("expected cache miss on first request")
	}
	if resp.Text != "integration-response" {
		t.Fatalf("got %q, want integration-response", resp.Text)
	}

	resp2, err := e.Process(runCtx, req, nil)
	if err != nil {
		t.Fatalf("second process: %v", err)
	}
	if !resp2.Cached {
		t.Fatal("expected cache hit on second request")
	}
	if resp2.Text != "integration-response" {
		t.Fatalf("cached got %q", resp2.Text)
	}
}
