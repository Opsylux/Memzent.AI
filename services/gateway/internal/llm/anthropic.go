package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type AnthropicProvider struct {
	APIKey string
	Model  string

	mu              sync.RWMutex
	supportedModels []string
}

func NewAnthropicProvider(apiKey, model string) Provider {
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	return &AnthropicProvider{APIKey: apiKey, Model: model}
}

func (a *AnthropicProvider) GetProviderName() string { return "Anthropic" }

func (a *AnthropicProvider) GetMetadata() ProviderMetadata {
	a.mu.RLock()
	defer a.mu.RUnlock()

	models := a.supportedModels
	if len(models) == 0 {
		models = []string{"claude-sonnet-4-20250514", "claude-opus-4-20250514", "claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022"}
	}
	return ProviderMetadata{
		Name:            "anthropic",
		DefaultModel:    a.Model,
		SupportedModels: models,
	}
}

func (a *AnthropicProvider) DiscoverModels(ctx context.Context) ([]string, error) {
	if a.APIKey == "" {
		return []string{"claude-sonnet-4-20250514", "claude-opus-4-20250514", "claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022"}, nil
	}

	url := "https://api.anthropic.com/v1/models"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", a.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Anthropic models API may not be available on all plans — fall back gracefully
		return []string{"claude-sonnet-4-20250514", "claude-opus-4-20250514", "claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022"}, nil
	}

	var res struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range res.Data {
		if strings.Contains(m.ID, "claude") {
			models = append(models, m.ID)
		}
	}

	if len(models) == 0 {
		models = []string{"claude-sonnet-4-20250514", "claude-opus-4-20250514", "claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022"}
	}

	a.mu.Lock()
	a.supportedModels = models
	a.mu.Unlock()

	return models, nil
}

func (a *AnthropicProvider) Generate(ctx context.Context, messages []Message, tools []any, model string) (string, *TokenUsage, error) {
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
		"messages":   messages,
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
