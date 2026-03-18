package llm

import (
	"context"
	"fmt"
)

type MockProvider struct{}

func NewMockProvider() Provider { return &MockProvider{} }

func (m *MockProvider) GetProviderName() string { return "Mock-Aura" }

func (m *MockProvider) Generate(ctx context.Context, prompt string, tools []any) (string, error) {
	return fmt.Sprintf("[Mock] Prompt: %s | Tools: %d active", prompt, len(tools)), nil
}
