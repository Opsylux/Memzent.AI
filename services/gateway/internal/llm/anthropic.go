package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type AnthropicProvider struct {
	APIKey string
	Model  string
}

func NewAnthropicProvider(apiKey, model string) Provider {
	if model == "" {
		model = "claude-3-5-sonnet-20240620"
	}
	return &AnthropicProvider{APIKey: apiKey, Model: model}
}

func (a *AnthropicProvider) GetProviderName() string { return "Anthropic" }

func (a *AnthropicProvider) GetMetadata() ProviderMetadata {
	return ProviderMetadata{
		Name:            "anthropic",
		DefaultModel:    "claude-3-5-sonnet-20240620",
		SupportedModels: []string{"claude-3-5-sonnet-20240620", "claude-3-opus-20240229", "claude-3-haiku-20240307"},
	}
}

func (a *AnthropicProvider) Generate(ctx context.Context, prompt string, tools []any, model string) (string, *TokenUsage, error) {
	url := "https://api.anthropic.com/v1/messages"

	// Resolve model: per-request override takes priority over configured default
	activeModel := a.Model
	if model != "" {
		activeModel = model
	}

	system := BuildSystemPrompt(tools)

	// 2. Prepare Request
	reqBody := map[string]interface{}{
		"model":      activeModel,
		"system":     system,
		"max_tokens": 1024,
		"messages": []Message{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Content []struct{ Text string }  `json:"content"`
		Error   struct{ Message string } `json:"error"`
		Usage   struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("anthropic error: %s", result.Error.Message)
	}

	usage := &TokenUsage{
		PromptTokens:     result.Usage.InputTokens,
		CompletionTokens: result.Usage.OutputTokens,
		TotalTokens:      result.Usage.InputTokens + result.Usage.OutputTokens,
	}

	return result.Content[0].Text, usage, nil
}
