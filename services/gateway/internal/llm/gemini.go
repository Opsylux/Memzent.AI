package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type GeminiProvider struct {
	APIKey string
	Model  string
}

func NewGeminiProvider(apiKey, model string) Provider {
	if model == "" {
		model = "gemini-2.5-flash"
	}
	return &GeminiProvider{APIKey: apiKey, Model: model}
}

func (g *GeminiProvider) GetProviderName() string { return "Gemini" }

func (g *GeminiProvider) GetMetadata() ProviderMetadata {
	return ProviderMetadata{
		Name:            "gemini",
		DefaultModel:    "gemini-1.5-flash",
		SupportedModels: []string{"gemini-1.5-flash", "gemini-1.5-pro", "gemini-1.0-pro"},
	}
}

func (g *GeminiProvider) Generate(ctx context.Context, messages []Message, tools []any, model string) (string, *TokenUsage, error) {
	// Resolve model: per-request override takes priority over configured default
	activeModel := g.Model
	if model != "" {
		activeModel = model
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", activeModel, g.APIKey)

	system := BuildSystemPrompt(tools)

	var apiContents []map[string]interface{}
	for _, m := range messages {
		role := m.Role
		if role == "assistant" {
			role = "model" // Gemini uses "model" instead of "assistant"
		}
		apiContents = append(apiContents, map[string]interface{}{
			"role": role,
			"parts": []map[string]string{
				{"text": m.Content},
			},
		})
	}

	// 2. Prepare Gemini Request Body
	reqBody := map[string]interface{}{
		"contents": apiContents,
		"systemInstruction": map[string]interface{}{
			"parts": []map[string]string{
				{"text": system},
			},
		},
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
		return "", nil, fmt.Errorf("gemini error (%d): %s", resp.StatusCode, errResp.Error.Message)
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", nil, err
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", nil, fmt.Errorf("gemini returned no content")
	}

	usage := &TokenUsage{
		PromptTokens:     result.UsageMetadata.PromptTokenCount,
		CompletionTokens: result.UsageMetadata.CandidatesTokenCount,
		TotalTokens:      result.UsageMetadata.TotalTokenCount,
	}

	return result.Candidates[0].Content.Parts[0].Text, usage, nil
}
