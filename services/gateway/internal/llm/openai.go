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

func (o *OpenAIProvider) GetProviderName() string { return "OpenAI" }

func (o *OpenAIProvider) Generate(ctx context.Context, prompt string, tools []any) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	// 1. Build System Message with Tool Context
	system := "You are Aura, an enterprise-grade AI Gateway. "
	if len(tools) > 0 {
		system += fmt.Sprintf("The user has requested actions that triggered the following tools/data: %v. Use this data to provide a precise, professional response.", tools)
	}

	// 2. Prepare OpenAI Request Body
	reqBody := map[string]interface{}{
		"model": o.Model,
		"messages": []Message{
			{Role: "system", Content: system},
			{Role: "user", Content: prompt},
		},
		"temperature": 0.5,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return "", fmt.Errorf("openai error (%d): %s", resp.StatusCode, errResp.Error.Message)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("openai returned no choices")
	}

	return result.Choices[0].Message.Content, nil
}
