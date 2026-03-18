package engine

import (
	"aura-gateway/internal/router"
	"context"
)

// PromptRequest represents the incoming user request
type PromptRequest struct {
	UserID string
	Prompt string
}

// PromptResponse represents the final synthesized response
type PromptResponse struct {
	Prompt        string         `json:"prompt"`
	SelectedTools []*router.Tool `json:"tools"`
	EngineOutput  string         `json:"engine_output"`
	TokenSavings  int64          `json:"token_savings"`
	Cached        bool           `json:"cached"`
}

// AuraEngine defines the core orchestration logic
type AuraEngine interface {
	// Process handles the end-to-end flow: Cache -> Router -> RBAC -> LLM
	Process(ctx context.Context, req *PromptRequest) (*PromptResponse, error)
}
