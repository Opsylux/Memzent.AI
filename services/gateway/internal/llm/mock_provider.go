package llm

import (
	"context"
	"fmt"
)

type MockProvider struct{}

func NewMockProvider() Provider {
	return &MockProvider{}
}

func (m *MockProvider) GenerateResponse(ctx context.Context, messages []Message) (string, error) {
	// Simple mock response that reflects the context
	return fmt.Sprintf("[Mock %s] I processed your request. Based on the tools selected, I can help you with your query.", m.GetProviderName()), nil
}

func (m *MockProvider) GetProviderName() string {
	return "Mock-GPT"
}
