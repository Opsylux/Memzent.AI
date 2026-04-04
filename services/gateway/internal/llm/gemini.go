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

func (g *GeminiProvider) Generate(ctx context.Context, prompt string, tools []any, model string) (string, error) {
	// Resolve model: per-request override takes priority over configured default
	activeModel := g.Model
	if model != "" {
		activeModel = model
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", activeModel, g.APIKey)

	// 1. Build System Instruction with Tool Context
	system := "You are Aura, an enterprise-grade AI Gateway. "
	if len(tools) > 0 {
		system += fmt.Sprintf("The user has requested actions that triggered the following tools/data: %v. Use this data to provide a precise, professional response.", tools)
	}

	// 2. Prepare Gemini Request Body
	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"systemInstruction": map[string]interface{}{
			"parts": []map[string]string{
				{"text": system},
			},
		},
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
		return "", fmt.Errorf("gemini error (%d): %s", resp.StatusCode, errResp.Error.Message)
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned no content")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}
