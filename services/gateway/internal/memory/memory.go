package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"memzent-gateway/internal/llm"
	"memzent-gateway/internal/router"
)

// MemoryManager orchestrates long-term semantic facts extraction and retrieval
type MemoryManager struct {
	routerClient router.SemanticRouter
	providers    map[string]llm.Provider
	defProvider  string
}

// NewMemoryManager instantiates a long-term memory orchestrator
func NewMemoryManager(rc router.SemanticRouter, providers map[string]llm.Provider, defaultProvider string) *MemoryManager {
	return &MemoryManager{
		routerClient: rc,
		providers:    providers,
		defProvider:  defaultProvider,
	}
}

// RetrieveSemanticContext searches semantic memory and returns formatted systemic guidelines
func (mm *MemoryManager) RetrieveSemanticContext(ctx context.Context, prompt, orgID, userID string, threshold float32) (string, error) {
	if mm.routerClient == nil {
		return "", nil
	}

	hits, err := mm.routerClient.QueryMemory(ctx, prompt, orgID, userID, threshold)
	if err != nil {
		slog.Error("Semantic memory query failed", "error", err, "org_id", orgID)
		return "", err
	}

	if len(hits) == 0 {
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString("\n\n### SYSTEM CONTEXT MEMORY\n")
	sb.WriteString("The following historical facts about this user/organization have been retrieved from long-term memory. Maintain perfect consistency with this context:\n")
	for _, hit := range hits {
		sb.WriteString(fmt.Sprintf("- %s (relevance: %.2f)\n", hit.Fact, hit.RelevanceScore))
	}
	sb.WriteString("--- END SYSTEM CONTEXT MEMORY ---\n")

	slog.Info("🎯 Retrieved long-term memories", "count", len(hits), "org_id", orgID)
	return sb.String(), nil
}

// FactExtractionResult maps the structured output of the fact-extraction prompt
type FactExtractionResult struct {
	Facts []string `json:"facts"`
}

// ExtractAndStoreFacts parses dialogue in the background, extracts distinct truths, and stores them in Qdrant
func (mm *MemoryManager) ExtractAndStoreFacts(ctx context.Context, orgID, userID, userPrompt, assistantResponse string) {
	if mm.routerClient == nil {
		return
	}

	// Resolve the extraction provider
	provider, ok := mm.providers[mm.defProvider]
	if !ok {
		// Fallback to any active provider if default is missing
		for _, p := range mm.providers {
			provider = p
			break
		}
	}

	if provider == nil {
		slog.Warn("Skipping fact extraction: no active LLM provider found")
		return
	}

	go func() {
		// Run out-of-band so we never block prompt execution
		bgCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()

		extractionPrompt := fmt.Sprintf(`Analyze the following user-assistant exchange.
Extract any distinct, permanent, and factual information declared by the user about their system configuration, database choices, tech stack, workspace parameters, or long-term preferences.

### GUIDELINES:
- A "fact" is a clear historical truth, user choice, or explicit architectural rule (e.g. "User runs Qdrant on port 6334", "User uses Next.js 15", "User prefers dark mode").
- Ignore casual dialogue, greetings, generic questions, temporal state mentions, or conversational noise.
- Respond with a structured JSON object containing a string array under the key "facts".
- If no new permanent facts are stated, return an empty array {"facts": []}.

### EXCHANGED CONVERSATION:
User: "%s"
Assistant: "%s"

Output strict JSON:`, userPrompt, assistantResponse)

		messages := []llm.Message{
			{Role: "user", Content: extractionPrompt},
		}

		response, _, err := provider.Generate(bgCtx, messages, nil, "")
		if err != nil {
			slog.Error("Fact extraction LLM call failed", "error", err)
			return
		}

		// Strip markdown code fences and trailing commentary that LLMs often add
		cleanResponse := response
		if idx := strings.Index(cleanResponse, "```json"); idx != -1 {
			cleanResponse = cleanResponse[idx+7:]
		} else if idx := strings.Index(cleanResponse, "```"); idx != -1 {
			cleanResponse = cleanResponse[idx+3:]
		}
		if idx := strings.Index(cleanResponse, "```"); idx != -1 {
			cleanResponse = cleanResponse[:idx]
		}
		cleanResponse = strings.TrimSpace(cleanResponse)

		// Locate JSON boundaries in response
		startIdx := strings.Index(cleanResponse, "{")
		endIdx := strings.LastIndex(cleanResponse, "}")
		if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
			slog.Debug("No JSON brackets found in fact extraction response", "response", response)
			return
		}

		cleanJSON := cleanResponse[startIdx : endIdx+1]
		var result FactExtractionResult
		if err := json.Unmarshal([]byte(cleanJSON), &result); err != nil {
			slog.Warn("Failed to unmarshal structured facts result", "error", err, "raw", cleanJSON)
			return
		}

		extractedCount := 0
		for _, fact := range result.Facts {
			trimmedFact := strings.TrimSpace(fact)
			if len(trimmedFact) < 5 {
				continue
			}

			// Store memory in Qdrant via the Rust router
			ok, err := mm.routerClient.StoreMemory(bgCtx, trimmedFact, orgID, userID)
			if err != nil || !ok {
				slog.Error("Failed to store vectorized memory fact in Qdrant", "fact", trimmedFact, "error", err)
			} else {
				slog.Info("💾 Saved long-term semantic memory fact", "fact", trimmedFact, "org_id", orgID)
				extractedCount++
			}
		}

		if extractedCount > 0 {
			slog.Info("🔥 Fact extraction complete", "saved_facts", extractedCount, "org_id", orgID)
		}
	}()
}
