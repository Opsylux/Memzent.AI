package engine

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"aura-gateway/internal/auth"
	cch "aura-gateway/internal/cache"
	lp "aura-gateway/internal/llm"
	mc "aura-gateway/internal/mcp"
	rtr "aura-gateway/internal/router"

	"golang.org/x/time/rate"
)

// PromptRequest defines the incoming user payload
type PromptRequest struct {
	UserID string `json:"user_id"`
	Prompt string `json:"prompt"`
}

// PromptResponse defines the gateway's response to the client
type PromptResponse struct {
	Text      string `json:"text"`
	Cached    bool   `json:"cached"`
	Tools     []any  `json:"tools,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// AuraEngine orchestrates the flow between Cache, RBAC, Router, MCP, and LLM
type AuraEngine struct {
	cache         *cch.AuraCache
	router        *rtr.RouterClient
	rbac          *auth.RBACClient
	llm           lp.Provider
	mcp           *mc.MCPClient
	toolThreshold float64
	cacheTTL      time.Duration
	rateLimiters  sync.Map

	TotalRequests atomic.Uint64
	CacheHits     atomic.Uint64
}

// NewAuraEngine initializes the engine with its required dependencies
func NewAuraEngine(c *cch.AuraCache, r *rtr.RouterClient, auth *auth.RBACClient, p lp.Provider, m *mc.MCPClient, threshold float64, ttl time.Duration) *AuraEngine {
	return &AuraEngine{
		cache:         c,
		router:        r,
		rbac:          auth,
		llm:           p,
		mcp:           m,
		toolThreshold: threshold,
		cacheTTL:      ttl,
	}
}

func (e *AuraEngine) Process(ctx context.Context, req *PromptRequest) (*PromptResponse, error) {
	e.TotalRequests.Add(1)

	// A. Rate Limiting (10 requests per minute per user)
	limiter, _ := e.rateLimiters.LoadOrStore(req.UserID, rate.NewLimiter(rate.Every(time.Minute/10), 10))
	if !limiter.(*rate.Limiter).Allow() {
		return nil, fmt.Errorf("rate limit exceeded: please wait a moment")
	}

	// B. Stage 1: Exact Match (Hash)
	if e.cache != nil {
		cachedResp, err := e.cache.GetSemanticResult(ctx, req.Prompt)
		if err == nil && cachedResp != "" {
			e.CacheHits.Add(1)
			slog.Info("🎯 Stage 1 Cache HIT (Literal)", "prompt", req.Prompt)
			return &PromptResponse{Text: cachedResp, Cached: true}, nil
		}

		// Stage 1.5: Canonical Match (Normalized)
		// Mask IDs, lowercase, and stabilize to catch write011 vs write111
		canonical, cHash := NormalizePrompt(req.Prompt)
		cachedCanon, err := e.cache.GetSemanticResult(ctx, cHash)
		if err == nil && cachedCanon != "" {
			e.CacheHits.Add(1)
			slog.Info("🎯 Stage 1.5 Cache HIT (Canonical)", "original", req.Prompt, "canonical", canonical)
			// Map the original literal string to this hit for faster Stage 1 next time
			_ = e.cache.SetResult(ctx, req.Prompt, cachedCanon, e.cacheTTL)
			return &PromptResponse{Text: cachedCanon, Cached: true}, nil
		}
	}


	// C. RBAC Check
	var allowedTools []string
	if e.rbac != nil {
		allowed, err := e.rbac.CheckPermission(ctx, req.UserID, "chat:execute")
		if err != nil {
			slog.Error("RBAC check failed", "error", err, "user_id", req.UserID)
		}
		if !allowed {
			return nil, fmt.Errorf("unauthorized: insufficient scope")
		}
		// Get tools specifically allowed for this user
		allowedTools, _ = e.rbac.GetAllowedTools(req.UserID)
	}

	// D. Semantic Routing (includes Vector Search & Prompt Compression via Rust)
	tools, compressedPrompt, similarPromptHash, currentPromptHash, err := e.router.GetBestTools(ctx, req.Prompt, req.UserID, allowedTools)
	if err != nil {
		slog.Warn("Router fallback engaged", "error", err)
	}

	// NEW: Stage 2 Cache Check (Fuzzy Vector Semantic Match)
	// If the Rust router found a highly similar prompt hash, we check Valkey using it.
	if similarPromptHash != "" && e.cache != nil {
		cachedResp, err := e.cache.GetSemanticResult(ctx, similarPromptHash)
		if err == nil && cachedResp != "" {
			slog.Info("🎯 Stage 2 Cache HIT (Vector)", "original", req.Prompt, "similar_hash", similarPromptHash)

			// Repopulate Literal (Stage 1) and Canonical (Stage 1.5) for next time
			_ = e.cache.SetResult(ctx, req.Prompt, cachedResp, e.cacheTTL)
			_, cHash := NormalizePrompt(req.Prompt)
			_ = e.cache.SetResult(ctx, cHash, cachedResp, e.cacheTTL)

			return &PromptResponse{Text: cachedResp, Cached: true}, nil

		}
	}

	// E. Tool Execution (MCP)
	var toolResults []string
	if e.mcp != nil && len(tools) > 0 {
		for _, t := range tools {
			if t.RelevanceScore > float32(e.toolThreshold) {
				slog.Info("Executing tool", "tool_id", t.Id, "score", t.RelevanceScore)

				// Use a sub-context for the tool call to prevent hanging
				toolCtx, cancel := context.WithTimeout(ctx, 15*time.Second)

				args := map[string]string{
					"tool_id": t.Id,
					"user_id": req.UserID,
				}

				resp, err := e.mcp.CallTool(toolCtx, "execute_aura_tool", args)
				cancel() // Release context immediately

				if err != nil {
					slog.Error("MCP tool execution error", "tool_id", t.Id, "error", err)
					continue
				}

				// Extract text content safely
				if resp != nil {
					for _, content := range resp.Content {
						if content.TextContent != nil {
							toolResults = append(toolResults, content.TextContent.Text)
						}
					}
				}
			}
		}
	}

	// F. LLM Synthesis
	// Use the compressed prompt from the Rust layer to save costs and latency.
	contextPrompt := compressedPrompt
	if contextPrompt == "" {
		contextPrompt = req.Prompt // Fallback
	}

	if len(toolResults) > 0 {
		contextPrompt = fmt.Sprintf("User Prompt: %s\n\nExecuted Tool Data:\n%v", contextPrompt, toolResults)
	}

	// Mapping *router.Tool to any slice for the prompt response payload
	var llmTools []any
	for _, t := range tools {
		llmTools = append(llmTools, t)
	}

	aiResp, err := e.llm.Generate(ctx, contextPrompt, llmTools)
	if err != nil {
		slog.Error("LLM generation failed", "error", err)
		return nil, err
	}

	// G. Populate Cache for future requests
	if e.cache != nil {
		// Layer 1: Literal Match
		_ = e.cache.SetResult(ctx, req.Prompt, aiResp, e.cacheTTL)

		// Layer 2: Canonical Match (Masking IDs/Numbers)
		_, cHash := NormalizePrompt(req.Prompt)
		_ = e.cache.SetResult(ctx, cHash, aiResp, e.cacheTTL)

		// Layer 3: Semantic Match (Representative Hash from Router)
		if currentPromptHash != "" && currentPromptHash != cHash {
			_ = e.cache.SetResult(ctx, currentPromptHash, aiResp, e.cacheTTL)
		}
	}


	return &PromptResponse{
		Text:   aiResp,
		Cached: false,
		Tools:  llmTools,
	}, nil
}
