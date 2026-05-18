package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type OpenAIProvider struct {
	APIKey string
	Model  string
}

func NewOpenAIProvider(apiKey, model string) Provider {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIProvider{APIKey: apiKey, Model: model}
}

func (o *OpenAIProvider) GetProviderName() string { return "OpenAI (" + o.Model + ")" }

func (o *OpenAIProvider) GetMetadata() ProviderMetadata {
	return ProviderMetadata{
		Name:            "openai",
		DefaultModel:    o.Model,
		SupportedModels: []string{o.Model, "gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"},
	}
}

func (o *OpenAIProvider) Generate(ctx context.Context, messages []Message, tools []any, model string) (string, *TokenUsage, error) {
	url := "https://api.openai.com/v1/chat/completions"

	// Resolve model: per-request override takes priority over configured default
	activeModel := o.Model
	if model != "" {
		activeModel = model
	}

	system := BuildSystemPrompt(tools)
	
	apiMessages := []Message{{Role: "system", Content: system}}
	apiMessages = append(apiMessages, messages...)

	// 2. Prepare OpenAI Request Body
	reqBody := map[string]interface{}{
		"model": activeModel,
		"messages": apiMessages,
		"temperature": 0.5,
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
	req.Header.Set("Authorization", "Bearer "+o.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return "", nil, fmt.Errorf("openai error (%d): %s", resp.StatusCode, errResp.Error.Message)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", nil, err
	}

	if len(result.Choices) == 0 {
		return "", nil, fmt.Errorf("openai returned no choices")
	}

	usage := &TokenUsage{
		PromptTokens:     result.Usage.PromptTokens,
		CompletionTokens: result.Usage.CompletionTokens,
		TotalTokens:      result.Usage.TotalTokens,
	}

	return result.Choices[0].Message.Content, usage, nil
}
