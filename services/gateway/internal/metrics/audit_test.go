package metrics

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// AuditBuffer — ring buffer behaviour
// ---------------------------------------------------------------------------

func TestNewAuditBuffer_Empty(t *testing.T) {
	buf := NewAuditBuffer(10)
	got := buf.GetLatest("", 10)
	if len(got) != 0 {
		t.Errorf("Expected empty buffer, got %d events", len(got))
	}
}

func TestAuditBuffer_AddAndRetrieve(t *testing.T) {
	buf := NewAuditBuffer(10)
	buf.Add(AuditEvent{OrgID: "org1", Type: "GATEWAY", Detail: "test", Status: "success", Timestamp: time.Now()})

	got := buf.GetLatest("org1", 10)
	if len(got) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(got))
	}
	if got[0].OrgID != "org1" {
		t.Errorf("Expected OrgID org1, got %q", got[0].OrgID)
	}
}

func TestAuditBuffer_OrgScoping(t *testing.T) {
	buf := NewAuditBuffer(20)
	for i := 0; i < 5; i++ {
		buf.Add(AuditEvent{OrgID: "orgA", Type: "GATEWAY", Timestamp: time.Now()})
		buf.Add(AuditEvent{OrgID: "orgB", Type: "GATEWAY", Timestamp: time.Now()})
	}

	gotA := buf.GetLatest("orgA", 20)
	for _, ev := range gotA {
		if ev.OrgID != "orgA" && ev.OrgID != "system" {
			t.Errorf("Scoped query returned event for wrong org: %q", ev.OrgID)
		}
	}
}

func TestAuditBuffer_SystemEventsVisibleToAll(t *testing.T) {
	buf := NewAuditBuffer(10)
	buf.Add(AuditEvent{OrgID: "system", Type: "GATEWAY", Detail: "boot", Timestamp: time.Now()})

	// Should be visible when filtering by any specific org
	got := buf.GetLatest("orgX", 10)
	if len(got) == 0 {
		t.Error("System events should be visible to all orgs")
	}
}

func TestAuditBuffer_LimitRespected(t *testing.T) {
	buf := NewAuditBuffer(50)
	for i := 0; i < 20; i++ {
		buf.Add(AuditEvent{OrgID: "", Type: "GATEWAY", Timestamp: time.Now()})
	}
	got := buf.GetLatest("", 5)
	if len(got) > 5 {
		t.Errorf("Limit not respected: got %d events, want ≤5", len(got))
	}
}

func TestAuditBuffer_RingOverwrite(t *testing.T) {
	const max = 5
	buf := NewAuditBuffer(max)

	// Fill beyond capacity
	for i := 0; i < max+3; i++ {
		buf.Add(AuditEvent{
			OrgID:     "org1",
			Type:      "GATEWAY",
			Detail:    fmt.Sprintf("event-%d", i),
			Timestamp: time.Now(),
		})
	}

	// Buffer should contain exactly max events (ring wraps)
	buf.mu.RLock()
	count := len(buf.events)
	buf.mu.RUnlock()

	if count != max {
		t.Errorf("Ring buffer should have exactly %d events, got %d", max, count)
	}
}

func TestAuditBuffer_ConcurrentSafety(t *testing.T) {
	buf := NewAuditBuffer(100)
	var wg sync.WaitGroup

	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				buf.Add(AuditEvent{
					OrgID:     fmt.Sprintf("org-%d", id),
					Type:      "GATEWAY",
					Timestamp: time.Now(),
				})
			}
		}(g)
	}

	// Concurrent reads
	for r := 0; r < 5; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = buf.GetLatest("", 20)
		}()
	}

	wg.Wait() // Must not race or deadlock
}

// ---------------------------------------------------------------------------
// PersistentAuditLogger — nil DB guard (no real DB needed)
// ---------------------------------------------------------------------------

func TestPersistentAuditLogger_NilDB_DoesNotPanic(t *testing.T) {
	logger := NewPersistentAuditLogger(nil)

	// Log with nil DB should silently no-op after in-memory add
	logger.Log(nil, AuditEvent{OrgID: "org1", Type: "TEST", Timestamp: time.Now()}, nil)

	// GetLatest with nil DB should return empty, not panic
	events, err := logger.GetLatest("org1", 10)
	if err != nil {
		t.Errorf("GetLatest with nil DB should return nil error, got: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("GetLatest with nil DB should return empty slice, got %d events", len(events))
	}
}

func TestPersistentAuditLogger_NilDB_GetCacheStats(t *testing.T) {
	logger := NewPersistentAuditLogger(nil)
	total, hits := logger.GetCacheStats("org1")
	if total != 0 || hits != 0 {
		t.Errorf("GetCacheStats with nil DB should return (0,0), got (%d,%d)", total, hits)
	}
}

func TestPersistentAuditLogger_NilDB_RetentionJobSafe(t *testing.T) {
	logger := NewPersistentAuditLogger(nil)
	// Cleanup with nil DB should no-op safely without panicking
	logger.Cleanup(nil, 30)
}

// ---------------------------------------------------------------------------
// AuditEvent — action string normalisation
// ---------------------------------------------------------------------------

func TestAuditEvent_ActionFormat(t *testing.T) {
	// Verify the action column format used in GetLatest reconstruction
	raw := "CACHE:Stage 1 HIT: how do I monitor cost"
	parts := splitAction(raw)
	if parts[0] != "CACHE" {
		t.Errorf("Expected type CACHE, got %q", parts[0])
	}
	if parts[1] != "Stage 1 HIT: how do I monitor cost" {
		t.Errorf("Detail mismatch: %q", parts[1])
	}
}

func TestAuditEvent_ActionNoColon(t *testing.T) {
	// When action has no colon, fallback: Type=SYSTEM, Detail=action
	raw := "SYSTEM_BOOT"
	parts := splitAction(raw)
	if len(parts) != 1 {
		t.Errorf("Expected 1 part for colon-free action, got %d", len(parts))
	}
}

// splitAction mirrors the SplitN logic in GetLatest for direct testing.
func splitAction(action string) []string {
	const sep = ":"
	idx := -1
	for i, ch := range action {
		if string(ch) == sep {
			idx = i
			break
		}
	}
	if idx == -1 {
		return []string{action}
	}
	return []string{action[:idx], action[idx+1:]}
}
