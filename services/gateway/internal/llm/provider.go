package llm

import "context"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Provider interface {
	// The Engine calls this. It handles turning tools into a system prompt internally.
	Generate(ctx context.Context, prompt string, tools []any) (string, error)
	GetProviderName() string
}
