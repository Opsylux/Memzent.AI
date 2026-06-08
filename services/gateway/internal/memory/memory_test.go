package memory

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	"google.golang.org/grpc"
	"memzent-gateway/internal/llm"
	"memzent-gateway/internal/router"
)

type mockSemanticRouterClient struct {
	router.SemanticRouterClient
	QueryMemoryFn func(ctx context.Context, in *router.QueryMemoryRequest, opts ...grpc.CallOption) (*router.QueryMemoryResponse, error)
	StoreMemoryFn func(ctx context.Context, in *router.StoreMemoryRequest, opts ...grpc.CallOption) (*router.StoreMemoryResponse, error)
}

func (m *mockSemanticRouterClient) QueryMemory(ctx context.Context, in *router.QueryMemoryRequest, opts ...grpc.CallOption) (*router.QueryMemoryResponse, error) {
	if m.QueryMemoryFn != nil {
		return m.QueryMemoryFn(ctx, in, opts...)
	}
	return nil, nil
}

func (m *mockSemanticRouterClient) StoreMemory(ctx context.Context, in *router.StoreMemoryRequest, opts ...grpc.CallOption) (*router.StoreMemoryResponse, error) {
	if m.StoreMemoryFn != nil {
		return m.StoreMemoryFn(ctx, in, opts...)
	}
	return nil, nil
}

type mockLLMProvider struct {
	GenerateFn func(ctx context.Context, messages []llm.Message, tools []any, model string) (string, *llm.TokenUsage, error)
}

func (m *mockLLMProvider) Generate(ctx context.Context, messages []llm.Message, tools []any, model string) (string, *llm.TokenUsage, error) {
	if m.GenerateFn != nil {
		return m.GenerateFn(ctx, messages, tools, model)
	}
	return "", nil, nil
}
func (m *mockLLMProvider) GetProviderName() string           { return "Mock" }
func (m *mockLLMProvider) GetMetadata() llm.ProviderMetadata { return llm.ProviderMetadata{} }

func setUnexportedField(target interface{}, fieldName string, value interface{}) {
	rv := reflect.ValueOf(target).Elem()
	field := rv.FieldByName(fieldName)
	ptr := unsafe.Pointer(field.UnsafeAddr())
	reflect.NewAt(field.Type(), ptr).Elem().Set(reflect.ValueOf(value))
}

func TestMemoryManager_RetrieveSemanticContext(t *testing.T) {
	ctx := context.Background()

	t.Run("Nil Router", func(t *testing.T) {
		mm := NewMemoryManager(nil, nil, "")
		res, err := mm.RetrieveSemanticContext(ctx, "prompt", "org1", "user1", 0.7)
		if err != nil || res != "" {
			t.Errorf("expected empty context and no error, got %q (err: %v)", res, err)
		}
	})

	t.Run("Router Query Error", func(t *testing.T) {
		mockRouter := &mockSemanticRouterClient{
			QueryMemoryFn: func(ctx context.Context, in *router.QueryMemoryRequest, opts ...grpc.CallOption) (*router.QueryMemoryResponse, error) {
				return nil, fmt.Errorf("qdrant timeout")
			},
		}
		rc := &router.RouterClient{}
		setUnexportedField(rc, "client", mockRouter)

		mm := NewMemoryManager(rc, nil, "")
		_, err := mm.RetrieveSemanticContext(ctx, "prompt", "org1", "user1", 0.7)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("No Memories Found", func(t *testing.T) {
		mockRouter := &mockSemanticRouterClient{
			QueryMemoryFn: func(ctx context.Context, in *router.QueryMemoryRequest, opts ...grpc.CallOption) (*router.QueryMemoryResponse, error) {
				return &router.QueryMemoryResponse{Memories: []*router.MemoryHit{}}, nil
			},
		}
		rc := &router.RouterClient{}
		setUnexportedField(rc, "client", mockRouter)

		mm := NewMemoryManager(rc, nil, "")
		res, err := mm.RetrieveSemanticContext(ctx, "prompt", "org1", "user1", 0.7)
		if err != nil || res != "" {
			t.Errorf("expected empty string and no error, got %q (err: %v)", res, err)
		}
	})

	t.Run("Success Formatting", func(t *testing.T) {
		mockRouter := &mockSemanticRouterClient{
			QueryMemoryFn: func(ctx context.Context, in *router.QueryMemoryRequest, opts ...grpc.CallOption) (*router.QueryMemoryResponse, error) {
				return &router.QueryMemoryResponse{
					Memories: []*router.MemoryHit{
						{Fact: "User runs Valkey standalone", RelevanceScore: 0.85},
						{Fact: "User prefers Go", RelevanceScore: 0.75},
					},
				}, nil
			},
		}
		rc := &router.RouterClient{}
		setUnexportedField(rc, "client", mockRouter)

		mm := NewMemoryManager(rc, nil, "")
		res, err := mm.RetrieveSemanticContext(ctx, "prompt", "org1", "user1", 0.7)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(res, "SYSTEM CONTEXT MEMORY") {
			t.Errorf("expected branding block wrapper")
		}
		if !strings.Contains(res, "User runs Valkey standalone (relevance: 0.85)") {
			t.Errorf("missing fact 1 formatting in output: %q", res)
		}
	})
}

func TestMemoryManager_ExtractAndStoreFacts(t *testing.T) {
	ctx := context.Background()

	t.Run("No Active Providers", func(t *testing.T) {
		mockRouter := &mockSemanticRouterClient{}
		rc := &router.RouterClient{}
		setUnexportedField(rc, "client", mockRouter)

		mm := NewMemoryManager(rc, make(map[string]llm.Provider), "default-provider")
		// Should return immediately without panics
		mm.ExtractAndStoreFacts(ctx, "org1", "user1", "hi", "hello")
	})

	t.Run("Extraction Success", func(t *testing.T) {
		done := make(chan bool, 2)

		mockRouter := &mockSemanticRouterClient{
			StoreMemoryFn: func(ctx context.Context, in *router.StoreMemoryRequest, opts ...grpc.CallOption) (*router.StoreMemoryResponse, error) {
				if in.Fact != "User runs Qdrant on port 6334" && in.Fact != "User prefers Postgres" {
					t.Errorf("unexpected fact in store call: %q", in.Fact)
				}
				done <- true
				return &router.StoreMemoryResponse{Success: true}, nil
			},
		}
		rc := &router.RouterClient{}
		setUnexportedField(rc, "client", mockRouter)

		mockProvider := &mockLLMProvider{
			GenerateFn: func(ctx context.Context, messages []llm.Message, tools []any, model string) (string, *llm.TokenUsage, error) {
				// Return JSON wrapped with conversational noise to test indexing brackets
				response := "Sure, here are the facts extracted: ```json\n" +
					`{"facts": ["User runs Qdrant on port 6334", "User prefers Postgres", "shrt"]}` +
					"\n```\nHope that helps!"
				return response, nil, nil
			},
		}

		providers := map[string]llm.Provider{
			"default": mockProvider,
		}

		mm := NewMemoryManager(rc, providers, "default")
		mm.ExtractAndStoreFacts(ctx, "org1", "user1", "Qdrant port is 6334 and I prefer Postgres", "Understood.")

		// Block and wait for background async storage
		for i := 0; i < 2; i++ {
			select {
			case <-done:
				// Synced successfully
			case <-time.After(1 * time.Second):
				t.Fatalf("timed out waiting for async fact extraction")
			}
		}
	})

	t.Run("Provider Fallback and LLM Failure", func(t *testing.T) {
		done := make(chan bool, 1)

		mockRouter := &mockSemanticRouterClient{}
		rc := &router.RouterClient{}
		setUnexportedField(rc, "client", mockRouter)

		mockProvider := &mockLLMProvider{
			GenerateFn: func(ctx context.Context, messages []llm.Message, tools []any, model string) (string, *llm.TokenUsage, error) {
				done <- true
				return "", nil, fmt.Errorf("llm offline")
			},
		}

		providers := map[string]llm.Provider{
			"some-provider": mockProvider,
		}

		// Set default provider to a non-existent key to trigger fallback logic
		mm := NewMemoryManager(rc, providers, "missing-default")
		mm.ExtractAndStoreFacts(ctx, "org1", "user1", "prompt", "resp")

		select {
		case <-done:
			// Generates fallback call and safely terminates
		case <-time.After(1 * time.Second):
			t.Fatalf("timeout waiting for fallback provider generation")
		}
	})

	t.Run("LLM Parse Error", func(t *testing.T) {
		done := make(chan bool, 1)

		mockRouter := &mockSemanticRouterClient{}
		rc := &router.RouterClient{}
		setUnexportedField(rc, "client", mockRouter)

		mockProvider := &mockLLMProvider{
			GenerateFn: func(ctx context.Context, messages []llm.Message, tools []any, model string) (string, *llm.TokenUsage, error) {
				done <- true
				return "No json markers here", nil, nil
			},
		}

		providers := map[string]llm.Provider{
			"default": mockProvider,
		}

		mm := NewMemoryManager(rc, providers, "default")
		mm.ExtractAndStoreFacts(ctx, "org1", "user1", "prompt", "resp")

		select {
		case <-done:
			// Terminates safely on parse error
		case <-time.After(1 * time.Second):
			t.Fatalf("timeout")
		}
	})
}
