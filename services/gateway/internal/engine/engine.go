package engine

import (
	"context"
	"fmt"
	"time"

	"aura-gateway/internal/auth"
	"aura-gateway/internal/llm"
	"aura-gateway/internal/mcp"
	"aura-gateway/internal/router"
	"log"

	"github.com/valkey-io/valkey-go"
)

type PromptRequest struct {
	UserID string
	Prompt string
}

type PromptResponse struct {
	Text   string `json:"text"`
	Cached bool   `json:"cached"`
	Tools  []any  `json:"tools,omitempty"`
}

type AuraEngine struct {
	cache  valkey.Client
	router *router.RouterClient
	rbac   *auth.RBACClient
	llm    llm.Provider
	mcp    *mcp.MCPClient
}

func NewAuraEngine(c valkey.Client, r *router.RouterClient, auth *auth.RBACClient, p llm.Provider, m *mcp.MCPClient) *AuraEngine {
	return &AuraEngine{cache: c, router: r, rbac: auth, llm: p, mcp: m}
}

func (e *AuraEngine) Process(ctx context.Context, req *PromptRequest) (*PromptResponse, error) {
	// A. Semantic Cache Check (Valkey)
	// We use the prompt as a key for now. In 2026, we'd use a vector hash.
	resp, err := e.cache.Do(ctx, e.cache.B().Get().Key(req.Prompt).Build()).ToString()
	if err == nil && resp != "" {
		return &PromptResponse{Text: resp, Cached: true}, nil
	}

	// B. RBAC Check (Postgres)
	// Ensure the user has "execute" permissions before routing
	if e.rbac != nil {
		allowed, _ := e.rbac.CheckPermission(ctx, req.UserID, "chat:execute")
		if !allowed {
			return nil, fmt.Errorf("unauthorized: insufficient scope")
		}
	}

	// C. Semantic Routing (Rust gRPC)
	// Ask the Rust router: "Which tools do I need for this prompt?"
	var allowedTools []string
	if e.rbac != nil {
		var err error
		allowedTools, err = e.rbac.GetAllowedTools(req.UserID)
		if err != nil {
			log.Printf("RBAC allowed tools error: %v", err)
		}
	}

	tools, err := e.router.GetBestTools(ctx, req.Prompt, req.UserID, allowedTools)
	if err != nil {
		log.Printf("Router fallback: %v", err)
	}

	// D. Tool Execution (MCP Integration)
	var toolResults []string
	if e.mcp != nil && len(tools) > 0 {
		for _, t := range tools {
			// Threshold check: only execute if relevance > 0.7
			if t.RelevanceScore > 0.7 {
				log.Printf("Executing tool: %s (Score: %.2f)", t.Id, t.RelevanceScore)

				// Initialize MCP client if needed (stateful)
				_ = e.mcp.Initialize(ctx)

				// Prepare arguments
				args := map[string]string{
					"tool_id": t.Id,
					"user_id": req.UserID,
				}

				resp, err := e.mcp.CallTool(ctx, "execute_aura_tool", args)
				if err != nil {
					log.Printf("MCP tool execution error for %s: %v", t.Id, err)
					continue
				}

				for _, content := range resp.Content {
					if content.Type == "text" {
						toolResults = append(toolResults, content.Text)
					}
				}
			}
		}
	}

	// E. LLM Synthesis
	// Pass the prompt, selected tools, AND executed results to the LLM
	contextPrompt := req.Prompt
	if len(toolResults) > 0 {
		contextPrompt = fmt.Sprintf("User Prompt: %s\n\nExecuted Tool Data:\n%s", req.Prompt, toolResults)
	}

	// Convert tools for LLM interface
	var llmTools []any
	for _, t := range tools {
		llmTools = append(llmTools, t)
	}

	aiResp, err := e.llm.Generate(ctx, contextPrompt, llmTools)
	if err != nil {
		return nil, err
	}

	// E. Populate Cache for next time (1 hour TTL)
	e.cache.Do(ctx, e.cache.B().Set().Key(req.Prompt).Value(aiResp).Ex(1*time.Hour).Build())

	return &PromptResponse{
		Text:   aiResp,
		Cached: false,
		Tools:  llmTools,
	}, nil
}
