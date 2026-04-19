package llm

import "context"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Provider interface {
	// Generate produces an LLM response. Model may be empty to use the provider default.
	Generate(ctx context.Context, prompt string, tools []any, model string) (string, error)
	GetProviderName() string
	GetMetadata() ProviderMetadata
}

type ProviderMetadata struct {
	Name          string   `json:"name"`
	DefaultModel  string   `json:"default_model"`
	SupportedModels []string `json:"supported_models"`
}
