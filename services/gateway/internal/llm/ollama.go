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

func (o *OllamaProvider) Generate(ctx context.Context, prompt string, tools []any, model string) (string, error) {
	url := fmt.Sprintf("%s/api/chat", o.BaseURL)

	// Resolve model: per-request override takes priority over configured default
	activeModel := o.Model
	if model != "" {
		activeModel = model
	}

	// 1. Build System Context map for Tool Execution
	system := "You are Aura, an enterprise-grade open-source AI Gateway. "
	if len(tools) > 0 {
		system += fmt.Sprintf("\nThe user request triggered the executing of semantic tools. Here is the raw context retrieved: %v\n\nPlease output a concise, professional synthesis of this data.", tools)
	} else {
		system += "\nProvide a helpful, concise response to the user's prompt."
	}

	// 2. Prepare Ollama Request Body
	reqBody := map[string]interface{}{
		"model": activeModel,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": prompt},
		},
		"stream": false, // We want the full block before returning to the Go engine
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ollama request: %w", err)
	}

	// 3. Dispatch the external API call
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to Ollama at %s: %w", o.BaseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama API error: received status %d", resp.StatusCode)
	}

	// 4. Decode Response (Ollama format)
	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode ollama JSON: %w", err)
	}

	if result.Message.Content == "" {
		return "", fmt.Errorf("ollama returned empty response content")
	}

	return result.Message.Content, nil
}
