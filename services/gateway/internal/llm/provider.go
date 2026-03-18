package llm

import (
	"context"
)

// Message represents a single turn in a conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Provider defines the interface for interacting with various LLM backends
type Provider interface {
	// GenerateResponse sends a structured prompt (pruned tools + user query) to the LLM
	GenerateResponse(ctx context.Context, messages []Message) (string, error)

	// GetProviderName returns the name of the provider (e.g., "anthropic", "openai")
	GetProviderName() string
}
