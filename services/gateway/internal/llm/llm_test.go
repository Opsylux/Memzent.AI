package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type mockRoundTripper struct {
	roundTrip func(*http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTrip(req)
}

func newJSONResponse(statusCode int, bodyObj interface{}) *http.Response {
	bodyBytes, _ := json.Marshal(bodyObj)
	return &http.Response{
		StatusCode: statusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(bodyBytes)),
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	promptWithoutTools := BuildSystemPrompt(nil)
	if !strings.Contains(promptWithoutTools, "memzent.ai") {
		t.Errorf("expected system prompt to contain product branding")
	}
	if !strings.Contains(promptWithoutTools, "Provide a helpful, direct, and concise response") {
		t.Errorf("expected fallback instruction in system prompt")
	}

	promptWithTools := BuildSystemPrompt([]any{"dummy_tool"})
	if !strings.Contains(promptWithTools, "supplemented with data from semantic tools") {
		t.Errorf("expected tool instruction in system prompt")
	}
}

func TestAnthropicProvider(t *testing.T) {
	oldTransport := http.DefaultClient.Transport
	defer func() { http.DefaultClient.Transport = oldTransport }()

	provider := NewAnthropicProvider("test-key", "")
	if provider.GetProviderName() != "Anthropic" {
		t.Errorf("expected provider name Anthropic, got %q", provider.GetProviderName())
	}

	meta := provider.GetMetadata()
	if meta.Name != "anthropic" || meta.DefaultModel != "claude-3-5-sonnet-20240620" {
		t.Errorf("unexpected metadata: %+v", meta)
	}

	providerCustom := NewAnthropicProvider("test-key", "claude-custom")
	if providerCustom.(*AnthropicProvider).Model != "claude-custom" {
		t.Errorf("expected custom model claude-custom")
	}

	t.Run("Generate Success", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				if req.URL.String() != "https://api.anthropic.com/v1/messages" {
					return nil, fmt.Errorf("unexpected URL: %s", req.URL)
				}
				if req.Header.Get("x-api-key") != "test-key" {
					return nil, fmt.Errorf("missing/invalid API Key header")
				}
				if req.Header.Get("anthropic-version") != "2023-06-01" {
					return nil, fmt.Errorf("missing anthropic version header")
				}

				respBody := map[string]interface{}{
					"content": []map[string]interface{}{
						{"text": "Hello human"},
					},
					"usage": map[string]interface{}{
						"input_tokens":  10,
						"output_tokens": 15,
					},
				}

				return newJSONResponse(http.StatusOK, respBody), nil
			},
		}

		ctx := context.Background()
		msg, usage, err := provider.Generate(ctx, []Message{{Role: "user", Content: "hi"}}, nil, "")
		if err != nil {
			t.Fatalf("unexpected generate error: %v", err)
		}
		if msg != "Hello human" {
			t.Errorf("expected Hello human, got %q", msg)
		}
		if usage.PromptTokens != 10 || usage.CompletionTokens != 15 || usage.TotalTokens != 25 {
			t.Errorf("unexpected usage counts: %+v", usage)
		}
	})

	t.Run("Generate Status Error", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				respBody := map[string]interface{}{
					"error": map[string]interface{}{
						"message": "invalid signature key",
					},
				}
				return newJSONResponse(http.StatusBadRequest, respBody), nil
			},
		}

		ctx := context.Background()
		_, _, err := provider.Generate(ctx, []Message{{Role: "user", Content: "hi"}}, nil, "claude-override")
		if err == nil {
			t.Fatalf("expected error on bad request, got nil")
		}
		if !strings.Contains(err.Error(), "invalid signature key") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("Generate Decode Error", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("bad_json{")),
				}, nil
			},
		}

		ctx := context.Background()
		_, _, err := provider.Generate(ctx, []Message{{Role: "user", Content: "hi"}}, nil, "")
		if err == nil {
			t.Fatalf("expected JSON decode error, got nil")
		}
	})
}

func TestOpenAIProvider(t *testing.T) {
	oldTransport := http.DefaultClient.Transport
	defer func() { http.DefaultClient.Transport = oldTransport }()

	provider := NewOpenAIProvider("openai-key", "")
	if provider.GetProviderName() != "OpenAI (gpt-4o-mini)" {
		t.Errorf("unexpected name: %q", provider.GetProviderName())
	}

	t.Run("DiscoverModels Success", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				if req.URL.String() != "https://api.openai.com/v1/models" {
					return nil, fmt.Errorf("unexpected URL: %s", req.URL)
				}
				if req.Header.Get("Authorization") != "Bearer openai-key" {
					return nil, fmt.Errorf("missing Authorization header")
				}

				respBody := map[string]interface{}{
					"data": []map[string]interface{}{
						{"id": "gpt-4o"},
						{"id": "o1-mini"},
						{"id": "text-davinci-003"},
					},
				}
				return newJSONResponse(http.StatusOK, respBody), nil
			},
		}

		ctx := context.Background()
		models, err := provider.(*OpenAIProvider).DiscoverModels(ctx)
		if err != nil {
			t.Fatalf("unexpected discover error: %v", err)
		}
		if len(models) != 2 || models[0] != "gpt-4o" || models[1] != "o1-mini" {
			t.Errorf("unexpected filtered models: %v", models)
		}

		meta := provider.GetMetadata()
		if len(meta.SupportedModels) != 2 {
			t.Errorf("metadata did not cache discovered models: %+v", meta)
		}
	})

	t.Run("DiscoverModels Status Error", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return newJSONResponse(http.StatusUnauthorized, nil), nil
			},
		}

		ctx := context.Background()
		p := NewOpenAIProvider("bad-key", "")
		_, err := p.(*OpenAIProvider).DiscoverModels(ctx)
		if err == nil {
			t.Fatalf("expected error on unauthorized, got nil")
		}
	})

	t.Run("Generate Success", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				if req.URL.String() != "https://api.openai.com/v1/chat/completions" {
					return nil, fmt.Errorf("unexpected URL: %s", req.URL)
				}

				respBody := map[string]interface{}{
					"choices": []map[string]interface{}{
						{
							"message": map[string]interface{}{
								"content": "OpenAI response",
							},
						},
					},
					"usage": map[string]interface{}{
						"prompt_tokens":     20,
						"completion_tokens": 30,
						"total_tokens":      50,
					},
				}

				return newJSONResponse(http.StatusOK, respBody), nil
			},
		}

		ctx := context.Background()
		msg, usage, err := provider.Generate(ctx, []Message{{Role: "user", Content: "hello"}}, nil, "")
		if err != nil {
			t.Fatalf("unexpected generate error: %v", err)
		}
		if msg != "OpenAI response" {
			t.Errorf("expected OpenAI response, got %q", msg)
		}
		if usage.TotalTokens != 50 {
			t.Errorf("unexpected usage counts: %+v", usage)
		}
	})

	t.Run("Generate Status Error", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				errResp := map[string]interface{}{
					"error": map[string]interface{}{
						"message": "quota exceeded",
					},
				}
				return newJSONResponse(http.StatusPaymentRequired, errResp), nil
			},
		}

		ctx := context.Background()
		_, _, err := provider.Generate(ctx, []Message{{Role: "user", Content: "hello"}}, nil, "")
		if err == nil {
			t.Fatalf("expected error on quota error, got nil")
		}
		if !strings.Contains(err.Error(), "quota exceeded") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("Generate No Choices Error", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				respBody := map[string]interface{}{
					"choices": []map[string]interface{}{},
				}
				return newJSONResponse(http.StatusOK, respBody), nil
			},
		}

		ctx := context.Background()
		_, _, err := provider.Generate(ctx, []Message{{Role: "user", Content: "hello"}}, nil, "")
		if err == nil {
			t.Fatalf("expected error on empty choices, got nil")
		}
		if !strings.Contains(err.Error(), "returned no choices") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestOllamaProvider(t *testing.T) {
	oldTransport := http.DefaultClient.Transport
	defer func() { http.DefaultClient.Transport = oldTransport }()

	provider := NewOllamaProvider("http://ollama-local:11434", "llama-test")
	if provider.GetProviderName() != "Ollama (llama-test)" {
		t.Errorf("unexpected name: %q", provider.GetProviderName())
	}

	providerDefault := NewOllamaProvider("", "")
	if providerDefault.GetMetadata().DefaultModel != "llama3.2" {
		t.Errorf("unexpected default model")
	}

	t.Run("DiscoverModels Success", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				if req.URL.String() != "http://ollama-local:11434/api/tags" {
					return nil, fmt.Errorf("unexpected URL: %s", req.URL)
				}

				respBody := map[string]interface{}{
					"models": []map[string]interface{}{
						{"name": "llama-test:latest"},
						{"name": "mistral:latest"},
					},
				}
				return newJSONResponse(http.StatusOK, respBody), nil
			},
		}

		ctx := context.Background()
		models, err := provider.(*OllamaProvider).DiscoverModels(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(models) != 2 || models[0] != "llama-test:latest" {
			t.Errorf("unexpected models: %v", models)
		}
	})

	t.Run("DiscoverModels Status Error", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return newJSONResponse(http.StatusInternalServerError, nil), nil
			},
		}

		ctx := context.Background()
		_, err := provider.(*OllamaProvider).DiscoverModels(ctx)
		if err == nil {
			t.Fatalf("expected error on internal server error, got nil")
		}
	})

	t.Run("Generate Success", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				if req.URL.String() != "http://ollama-local:11434/api/chat" {
					return nil, fmt.Errorf("unexpected URL: %s", req.URL)
				}

				respBody := map[string]interface{}{
					"message": map[string]interface{}{
						"content": "Ollama response content",
					},
					"prompt_eval_count": 5,
					"eval_count":        12,
				}

				return newJSONResponse(http.StatusOK, respBody), nil
			},
		}

		ctx := context.Background()
		msg, usage, err := provider.Generate(ctx, []Message{{Role: "user", Content: "say hi"}}, nil, "")
		if err != nil {
			t.Fatalf("unexpected generate error: %v", err)
		}
		if msg != "Ollama response content" {
			t.Errorf("expected Ollama response content, got %q", msg)
		}
		if usage.PromptTokens != 5 || usage.CompletionTokens != 12 || usage.TotalTokens != 17 {
			t.Errorf("unexpected usage counts: %+v", usage)
		}
	})

	t.Run("Generate Status Error", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return newJSONResponse(http.StatusNotFound, nil), nil
			},
		}

		ctx := context.Background()
		_, _, err := provider.Generate(ctx, []Message{{Role: "user", Content: "say hi"}}, nil, "")
		if err == nil {
			t.Fatalf("expected error on 404, got nil")
		}
	})

	t.Run("Generate Empty Content Error", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				respBody := map[string]interface{}{
					"message": map[string]interface{}{
						"content": "",
					},
				}
				return newJSONResponse(http.StatusOK, respBody), nil
			},
		}

		ctx := context.Background()
		_, _, err := provider.Generate(ctx, []Message{{Role: "user", Content: "say hi"}}, nil, "")
		if err == nil {
			t.Fatalf("expected error on empty content, got nil")
		}
	})
}

func TestGeminiProvider(t *testing.T) {
	oldTransport := http.DefaultClient.Transport
	defer func() { http.DefaultClient.Transport = oldTransport }()

	provider := NewGeminiProvider("gemini-key", "gemini-model")
	if provider.GetProviderName() != "Gemini" {
		t.Errorf("unexpected name: %q", provider.GetProviderName())
	}

	providerDefault := NewGeminiProvider("gemini-key", "")
	if providerDefault.GetMetadata().DefaultModel != "gemini-2.5-flash" {
		t.Errorf("unexpected default model")
	}

	t.Run("DiscoverModels Default Fallback (No Key)", func(t *testing.T) {
		p := NewGeminiProvider("", "")
		ctx := context.Background()
		models, err := p.(*GeminiProvider).DiscoverModels(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(models) != 3 || models[0] != "gemini-1.5-flash" {
			t.Errorf("unexpected fallback models: %v", models)
		}
	})

	t.Run("DiscoverModels Success", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				if !strings.Contains(req.URL.String(), "generativelanguage.googleapis.com/v1beta/models") {
					return nil, fmt.Errorf("unexpected URL: %s", req.URL)
				}

				respBody := map[string]interface{}{
					"models": []map[string]interface{}{
						{"name": "models/gemini-pro"},
						{"name": "models/gemini-ultra"},
						{"name": "models/bison-001"},
					},
				}
				return newJSONResponse(http.StatusOK, respBody), nil
			},
		}

		ctx := context.Background()
		models, err := provider.(*GeminiProvider).DiscoverModels(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(models) != 2 || models[0] != "gemini-pro" || models[1] != "gemini-ultra" {
			t.Errorf("unexpected models: %v", models)
		}
	})

	t.Run("DiscoverModels Status Error", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return newJSONResponse(http.StatusForbidden, nil), nil
			},
		}

		ctx := context.Background()
		_, err := provider.(*GeminiProvider).DiscoverModels(ctx)
		if err == nil {
			t.Fatalf("expected error on 403, got nil")
		}
	})

	t.Run("Generate Success", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				if !strings.Contains(req.URL.String(), "generateContent") {
					return nil, fmt.Errorf("unexpected URL: %s", req.URL)
				}

				respBody := map[string]interface{}{
					"candidates": []map[string]interface{}{
						{
							"content": map[string]interface{}{
								"parts": []map[string]interface{}{
									{"text": "Gemini response text"},
								},
							},
						},
					},
					"usageMetadata": map[string]interface{}{
						"promptTokenCount":     15,
						"candidatesTokenCount": 25,
						"totalTokenCount":      40,
					},
				}

				return newJSONResponse(http.StatusOK, respBody), nil
			},
		}

		ctx := context.Background()
		msg, usage, err := provider.Generate(ctx, []Message{{Role: "user", Content: "hello"}}, nil, "")
		if err != nil {
			t.Fatalf("unexpected generate error: %v", err)
		}
		if msg != "Gemini response text" {
			t.Errorf("expected Gemini response text, got %q", msg)
		}
		if usage.PromptTokens != 15 || usage.CompletionTokens != 25 || usage.TotalTokens != 40 {
			t.Errorf("unexpected usage counts: %+v", usage)
		}
	})

	t.Run("Generate Status Error", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				errResp := map[string]interface{}{
					"error": map[string]interface{}{
						"message": "invalid key",
					},
				}
				return newJSONResponse(http.StatusBadRequest, errResp), nil
			},
		}

		ctx := context.Background()
		_, _, err := provider.Generate(ctx, []Message{{Role: "user", Content: "hello"}}, nil, "")
		if err == nil {
			t.Fatalf("expected error on bad request, got nil")
		}
		if !strings.Contains(err.Error(), "invalid key") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("Generate Empty Content Error", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				respBody := map[string]interface{}{
					"candidates": []map[string]interface{}{},
				}
				return newJSONResponse(http.StatusOK, respBody), nil
			},
		}

		ctx := context.Background()
		_, _, err := provider.Generate(ctx, []Message{{Role: "user", Content: "hello"}}, nil, "")
		if err == nil {
			t.Fatalf("expected error on empty content, got nil")
		}
	})
}
