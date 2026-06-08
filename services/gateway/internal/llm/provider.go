package llm

import "context"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type Provider interface {
	// Generate produces an LLM response. Model may be empty to use the provider default.
	Generate(ctx context.Context, messages []Message, tools []any, model string) (string, *TokenUsage, error)
	GetProviderName() string
	GetMetadata() ProviderMetadata
}

type ModelDiscoverer interface {
	DiscoverModels(ctx context.Context) ([]string, error)
}

type ProviderMetadata struct {
	Name          string   `json:"name"`
	DefaultModel  string   `json:"default_model"`
	SupportedModels []string `json:"supported_models"`
}

// BuildSystemPrompt generates a concise system prompt focused on answering the user's question.
// Internal architecture, API docs, and code samples are NOT included — they caused the LLM
// to hallucinate Memzent internals instead of answering user queries directly.
func BuildSystemPrompt(tools []any) string {
	system := `You are a helpful AI assistant powered by Memzent. Answer the user's question directly, accurately, and concisely. Do not describe your own infrastructure, API endpoints, or internal architecture unless the user explicitly asks about Memzent's API or platform.`

	if len(tools) > 0 {
		system += "\n\nYou have been provided with context from semantic tools. Use this context ONLY if it is directly relevant to the user's question. If the tool data is irrelevant (e.g. database metrics for a math question), ignore it completely and answer the user's question normally."
	}
	return system
}
