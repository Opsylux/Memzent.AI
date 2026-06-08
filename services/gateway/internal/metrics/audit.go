package metrics

import (
	"sync"
	"time"
)

// AuditEvent represents a single request/event in the system
type AuditEvent struct {
	Timestamp  time.Time         `json:"timestamp"`
	OrgID      string            `json:"org_id"`
	Type       string            `json:"type"`        // e.g. "AUTH", "SEMANTIC_HIT", "GATEWAY", "REGISTRY"
	User       string            `json:"user"`        // e.g. "admin", "user@memzent.io", etc.
	Detail     string            `json:"detail"`      // e.g. "Session Active", "Intent: 'analyze gateway latency' - L1.5 Resolved"
	Status     string            `json:"status"`      // e.g. "success", "error", "warning"
	Latency    float64           `json:"latency_ms"`  // processing latency
	CacheLayer string            `json:"cache_layer,omitempty"` // L1, L1b, L2, L5 (which layer resolved)
	Entities   map[string]string `json:"entities,omitempty"` // extracted entities from prompt (E1)
}

// AuditBuffer is a thread-safe ring buffer for the latest audit events
type AuditBuffer struct {
	events []AuditEvent
	mu     sync.RWMutex
	max    int
	cursor int
}

var (
	GlobalAuditBuffer *AuditBuffer
)

func init() {
	GlobalAuditBuffer = NewAuditBuffer(100) // Keep last 100 events
	
	// Add initial "System Started" event
	GlobalAuditBuffer.Add(AuditEvent{
		Timestamp: time.Now(),
		OrgID:     "system",
		Type:      "GATEWAY",
		User:      "system",
		Detail:    "Memzent Gateway Infrastructure Initialized",
		Status:    "success",
	})
}

func NewAuditBuffer(max int) *AuditBuffer {
	return &AuditBuffer{
		events: make([]AuditEvent, 0, max),
		max:    max,
	}
}

func (b *AuditBuffer) Add(event AuditEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.events) < b.max {
		b.events = append(b.events, event)
	} else {
		b.events[b.cursor] = event
		b.cursor = (b.cursor + 1) % b.max
	}
}

func (b *AuditBuffer) GetLatest(orgID string, limit int) []AuditEvent {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]AuditEvent, 0)
	count := 0
	
	// Iterate backwards from cursor to get newest first
	size := len(b.events)
	for i := 0; i < size && count < limit; i++ {
		idx := (b.cursor - 1 - i + size) % size
		event := b.events[idx]
		
		// Org scoping
		if orgID == "" || event.OrgID == "system" || event.OrgID == orgID {
			result = append(result, event)
			count++
		}
	}
	
	return result
}
