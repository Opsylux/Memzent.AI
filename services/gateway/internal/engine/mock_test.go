package engine

import (
	"context"
	"sync"
	"time"

	"memzent-gateway/internal/router"
)

type mockCache struct {
	mu   sync.RWMutex
	data map[string]string
}

func newMockCache(seed map[string]string) *mockCache {
	if seed == nil {
		seed = make(map[string]string)
	}
	return &mockCache{data: seed}
}

func (m *mockCache) GetSemanticResult(_ context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data[key], nil
}

func (m *mockCache) SetResult(_ context.Context, key, value string, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

type mockRouter struct{}

func (m *mockRouter) GetBestTools(_ context.Context, _, _ string, _ []string, _ bool) ([]*router.Tool, string, string, string, map[string]string, error) {
	return nil, "", "", "", nil, nil
}

func (m *mockRouter) RegisterTool(_ context.Context, _, _, _, _ string) (bool, error) {
	return true, nil
}

func (m *mockRouter) PlanToolChain(_ context.Context, _, _ string, _ []string) ([]*router.ToolStep, float32, error) {
	return nil, 0, nil
}

func (m *mockRouter) StoreMemory(_ context.Context, _, _, _ string) (bool, error) {
	return true, nil
}

func (m *mockRouter) QueryMemory(_ context.Context, _, _, _ string, _ float32) ([]*router.MemoryHit, error) {
	return nil, nil
}

func (m *mockRouter) Close() {}
