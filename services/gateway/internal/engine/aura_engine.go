package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"aura-gateway/internal/auth"
	"aura-gateway/internal/llm"
	"aura-gateway/internal/router"

	"github.com/valkey-io/valkey-go"
)

type engineImpl struct {
	vClient    valkey.Client
	rClient    *router.RouterClient
	rbacClient *auth.RBACClient
	llmProv    llm.Provider
}

// NewAuraEngine creates a new instance of the core orchestration engine
func NewAuraEngine(v valkey.Client, r *router.RouterClient, rbac *auth.RBACClient, lp llm.Provider) AuraEngine {
	return &engineImpl{
		vClient:    v,
		rClient:    r,
		rbacClient: rbac,
		llmProv:    lp,
	}
}

func (e *engineImpl) Process(ctx context.Context, req *PromptRequest) (*PromptResponse, error) {
	// 1. Semantic Cache Check
	resp := e.vClient.Do(ctx, e.vClient.B().Get().Key(req.Prompt).Build())
	if cached, err := resp.ToString(); err == nil {
		var cachedResp PromptResponse
		if err := json.Unmarshal([]byte(cached), &cachedResp); err == nil {
			cachedResp.Cached = true
			return &cachedResp, nil
		}
	}

	// 2. Check Postgres for Allowed Tools (RBAC)
	var allowedTools []string
	if e.rbacClient != nil {
		var err error
		allowedTools, err = e.rbacClient.GetAllowedTools(req.UserID)
		if err != nil {
			log.Printf("⚠️ RBAC Error for %s: %v", req.UserID, err)
		}
	}

	// 3. Ask Rust Router for Tool Selection
	routerCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	tools, err := e.rClient.GetBestTools(routerCtx, req.Prompt, req.UserID, allowedTools)
	if err != nil {
		return nil, fmt.Errorf("semantic router error: %w", err)
	}

	// 4. LLM Response Synthesis (Passing selected tools for context)
	// For now, we simulate the messages
	messages := []llm.Message{
		{Role: "system", Content: fmt.Sprintf("You are Aura, an AI assistant. You have access to the following relevant tools: %v", tools)},
		{Role: "user", Content: req.Prompt},
	}

	llmOutput, err := e.llmProv.GenerateResponse(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("llm error: %w", err)
	}

	// 5. Build Final Response
	finalResp := &PromptResponse{
		Prompt:        req.Prompt,
		SelectedTools: tools,
		EngineOutput:  llmOutput,
		TokenSavings:  450, // Mocked for now
		Cached:        false,
	}

	// 6. Cache the result for 1 hour
	respJSON, _ := json.Marshal(finalResp)
	_ = e.vClient.Do(ctx, e.vClient.B().Set().Key(req.Prompt).Value(string(respJSON)).Ex(3600).Build())

	return finalResp, nil
}
