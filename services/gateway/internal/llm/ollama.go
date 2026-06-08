package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// OllamaProvider interfaces with a local instance of the open-source Ollama engine
type OllamaProvider struct {
	BaseURL string
	Model   string

	mu              sync.RWMutex
	supportedModels []string
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
	o.mu.RLock()
	defer o.mu.RUnlock()

	models := o.supportedModels
	if len(models) == 0 {
		models = []string{o.Model, "llama3", "mistral", "phi3"}
	}
	return ProviderMetadata{
		Name:            "ollama",
		DefaultModel:    o.Model,
		SupportedModels: models,
	}
}

func (o *OllamaProvider) DiscoverModels(ctx context.Context) ([]string, error) {
	url := fmt.Sprintf("%s/api/tags", o.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama models list error: status %d", resp.StatusCode)
	}

	var res struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range res.Models {
		models = append(models, m.Name)
	}

	if len(models) == 0 {
		models = []string{o.Model, "llama3", "mistral", "phi3"}
	}

	o.mu.Lock()
	o.supportedModels = models
	o.mu.Unlock()

	return models, nil
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

	reqBody := map[string]interface{}{
		"model":    activeModel,
		"messages": apiMessages,
		"stream":   false,
		// Keep the model loaded in VRAM indefinitely between requests.
		// Without this Ollama defaults to a 5-minute TTL and evicts the model,
		// causing a 2.5-2.7s cold-load penalty on the next request.
		// -1 means "never unload until Ollama is restarted".
		"keep_alive": -1,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal ollama request: %w", err)
	}

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

// GenerateStream uses Ollama's NDJSON streaming API and invokes onToken per delta.
func (o *OllamaProvider) GenerateStream(ctx context.Context, messages []Message, tools []any, model string, onToken func(string) error) (string, *TokenUsage, error) {
	url := fmt.Sprintf("%s/api/chat", o.BaseURL)

	activeModel := o.Model
	if model != "" {
		activeModel = model
	}

	system := BuildSystemPrompt(tools)
	apiMessages := []map[string]string{{"role": "system", "content": system}}
	for _, m := range messages {
		apiMessages = append(apiMessages, map[string]string{"role": m.Role, "content": m.Content})
	}

	reqBody := map[string]interface{}{
		"model":      activeModel,
		"messages":   apiMessages,
		"stream":     true,
		"keep_alive": -1,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, err
	}

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

	var full strings.Builder
	usage := &TokenUsage{}
	scanner := bufio.NewScanner(resp.Body)
	// Ollama NDJSON lines can exceed the default 64 KB scanner token limit.
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var chunk struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Done              bool `json:"done"`
			PromptEvalCount   int  `json:"prompt_eval_count"`
			EvalCount         int  `json:"eval_count"`
		}
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			continue
		}
		if chunk.Message.Content != "" {
			full.WriteString(chunk.Message.Content)
			if onToken != nil {
				if err := onToken(chunk.Message.Content); err != nil {
					return full.String(), usage, err
				}
			}
		}
		if chunk.Done {
			usage.PromptTokens = chunk.PromptEvalCount
			usage.CompletionTokens = chunk.EvalCount
			usage.TotalTokens = chunk.PromptEvalCount + chunk.EvalCount
			break
		}
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		return full.String(), usage, err
	}

	text := full.String()
	if text == "" {
		return "", nil, fmt.Errorf("ollama stream returned empty response")
	}
	return text, usage, nil
}
