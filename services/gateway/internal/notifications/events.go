package notifications

import "time"

// Event types that can trigger webhook notifications
const (
	EventCacheHit      = "cache_hit"
	EventToolExecution = "tool_execution"
	EventRateLimit     = "rate_limit"
	EventKeyRotated    = "key_rotated"
	EventToolRegistered = "tool_registered"
	EventSessionCreated = "session_created"
)

// AllEventTypes is the canonical list of subscribable events
var AllEventTypes = []string{
	EventCacheHit,
	EventToolExecution,
	EventRateLimit,
	EventKeyRotated,
	EventToolRegistered,
	EventSessionCreated,
}

// Event is the envelope delivered to webhook endpoints
type Event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	OrgID     string    `json:"org_id"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data"`
}

// Typed payloads for each event

type CacheHitData struct {
	Query    string  `json:"query"`
	Score    float64 `json:"score"`
	LatencyMs int64  `json:"latency_ms"`
	Model    string  `json:"model,omitempty"`
}

type ToolExecutionData struct {
	ToolID     string `json:"tool_id"`
	ToolName   string `json:"tool_name"`
	DurationMs int64  `json:"duration_ms"`
	Status     string `json:"status"` // "success", "error", "timeout"
	Error      string `json:"error,omitempty"`
}

type RateLimitData struct {
	UserID  string `json:"user_id"`
	Limit   int    `json:"limit"`
	Window  string `json:"window"` // "1m"
	Scope   string `json:"scope"`  // "org" or "user"
}

type KeyRotatedData struct {
	KeyID     string    `json:"key_id"`
	RotatedAt time.Time `json:"rotated_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type ToolRegisteredData struct {
	ToolID        string `json:"tool_id"`
	ToolName      string `json:"tool_name"`
	ConnectorType string `json:"connector_type"`
}

type SessionCreatedData struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id,omitempty"`
}
