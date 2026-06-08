package offline

import "time"

// OfflineEvent is the self-contained context capsule emitted from engine.Process()
// after every request resolution. It contains NO raw prompt text — only hashes and
// extracted structured data. See DESIGN_INTENT_CANONICALIZATION.md §7.2.
type OfflineEvent struct {
	OrgID         string            `json:"org_id"`
	UserID        string            `json:"user_id"`
	PromptHash    string            `json:"prompt_hash"`    // SHA-256 of raw prompt — non-reversible
	CanonicalHash string            `json:"canonical_hash"` // SHA-256 of normalized prompt
	Entities      map[string]string `json:"entities"`       // extracted entity map (may be nil)
	EntitySource  string            `json:"entity_source"`  // "regex" | "llm" | "none"
	ToolIDs       []string          `json:"tool_ids"`       // ordered list of tool IDs from workflow or routing
	ToolsUsed     []string          `json:"tools_used"`     // ordered list of tool IDs that were invoked
	WorkflowID    string            `json:"workflow_id"`    // matched workflow ID (if any)
	CacheLayer    string            `json:"cache_layer"`    // L1, L1b, L2, L5, workflow
	LatencyMs     int64             `json:"latency_ms"`
	TokensUsed    int               `json:"tokens_used"`
	Provider      string            `json:"provider"`
	Model         string            `json:"model"`
	Success       bool              `json:"success"`
	Timestamp     time.Time         `json:"timestamp"`
}
