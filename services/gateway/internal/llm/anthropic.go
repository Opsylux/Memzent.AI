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

func (a *AnthropicProvider) Generate(ctx context.Context, prompt string, tools []any) (string, error) {
	url := "https://api.anthropic.com/v1/messages"

	// 1. Build System Message with Tool Context
	system := "You are Aura, an AI Gateway. "
	if len(tools) > 0 {
		system += fmt.Sprintf("Selected tools: %v. Use this context to answer.", tools)
	}

	// 2. Prepare Request
	reqBody := map[string]interface{}{
		"model":      a.Model,
		"system":     system,
		"max_tokens": 1024,
		"messages": []Message{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Content []struct{ Text string }  `json:"content"`
		Error   struct{ Message string } `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic error: %s", result.Error.Message)
	}

	return result.Content[0].Text, nil
}
