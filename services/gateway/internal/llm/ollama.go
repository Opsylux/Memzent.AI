package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// OllamaProvider interfaces with a local instance of the open-source Ollama engine
type OllamaProvider struct {
	BaseURL string
	Model   string
}

func NewOllamaProvider(baseURL, model string) Provider {
	if model == "" {
		model = "llama3.2"
	}
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaProvider{BaseURL: baseURL, Model: model}
}

func (o *OllamaProvider) GetProviderName() string { return "Ollama (" + o.Model + ")" }

func (o *OllamaProvider) GetMetadata() ProviderMetadata {
	return ProviderMetadata{
		Name:            "ollama",
		DefaultModel:    o.Model,
		SupportedModels: []string{o.Model, "llama3", "mistral", "phi3"},
	}
}

func (o *OllamaProvider) Generate(ctx context.Context, messages []Message, tools []any, model string) (string, *TokenUsage, error) {
	url := fmt.Sprintf("%s/api/chat", o.BaseURL)

	// Resolve model: per-request override takes priority over configured default
	activeModel := o.Model
	if model != "" {
		activeModel = model
	}

	system := BuildSystemPrompt(tools)

	apiMessages := []map[string]string{
		{"role": "system", "content": system},
	}
	for _, m := range messages {
		apiMessages = append(apiMessages, map[string]string{"role": m.Role, "content": m.Content})
	}

	// 2. Prepare Ollama Request Body
	reqBody := map[string]interface{}{
		"model": activeModel,
		"messages": apiMessages,
		"stream": false, // We want the full block before returning to the Go engine
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal ollama request: %w", err)
	}

	// 3. Dispatch the external API call
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to connect to Ollama at %s: %w", o.BaseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("ollama API error: received status %d", resp.StatusCode)
	}

	// 4. Decode Response (Ollama format)
	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		PromptEvalCount int `json:"prompt_eval_count"`
		EvalCount       int `json:"eval_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", nil, fmt.Errorf("failed to decode ollama JSON: %w", err)
	}

	if result.Message.Content == "" {
		return "", nil, fmt.Errorf("ollama returned empty response content")
	}

	usage := &TokenUsage{
		PromptTokens:     result.PromptEvalCount,
		CompletionTokens: result.EvalCount,
		TotalTokens:      result.PromptEvalCount + result.EvalCount,
	}

	return result.Message.Content, usage, nil
}
