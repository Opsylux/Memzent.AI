package offline

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mockMiner implements Miner interface for testing.
type mockMiner struct {
	name      string
	processed []OfflineEvent
	mu        sync.Mutex
	flushes   atomic.Int32
}

func (m *mockMiner) Name() string { return m.name }
func (m *mockMiner) Process(_ context.Context, event OfflineEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processed = append(m.processed, event)
}
func (m *mockMiner) Flush(_ context.Context) {
	m.flushes.Add(1)
}
func (m *mockMiner) getProcessed() []OfflineEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]OfflineEvent, len(m.processed))
	copy(cp, m.processed)
	return cp
}

func TestPlane_EmitAndProcess(t *testing.T) {
	miner := &mockMiner{name: "test-miner"}
	plane := NewPlane(PlaneConfig{BufferSize: 100, WorkerCount: 2, FlushInterval: time.Hour}, miner)

	ctx := context.Background()
	plane.Start(ctx)

	// Emit events
	for i := 0; i < 10; i++ {
		plane.Emit(OfflineEvent{
			OrgID:      "org1",
			CacheLayer: "L5",
			PromptHash: "hash_" + string(rune('0'+i)),
			Timestamp:  time.Now(),
		})
	}

	// Give workers time to process
	time.Sleep(100 * time.Millisecond)
	plane.Stop()

	processed := miner.getProcessed()
	if len(processed) != 10 {
		t.Errorf("expected 10 processed events, got %d", len(processed))
	}
	if plane.EventsEmitted.Load() != 10 {
		t.Errorf("EventsEmitted = %d, want 10", plane.EventsEmitted.Load())
	}
	if plane.EventsProcessed.Load() != 10 {
		t.Errorf("EventsProcessed = %d, want 10", plane.EventsProcessed.Load())
	}
	if plane.EventsDropped.Load() != 0 {
		t.Errorf("EventsDropped = %d, want 0", plane.EventsDropped.Load())
	}
}

func TestPlane_NonBlockingDrop(t *testing.T) {
	miner := &mockMiner{name: "slow-miner"}
	// Tiny buffer to force drops
	plane := NewPlane(PlaneConfig{BufferSize: 2, WorkerCount: 0, FlushInterval: time.Hour}, miner)

	// Don't start workers — events will fill the buffer and drop
	plane.running.Store(true)

	for i := 0; i < 10; i++ {
		plane.Emit(OfflineEvent{OrgID: "org1", Timestamp: time.Now()})
	}

	if plane.EventsEmitted.Load() != 10 {
		t.Errorf("EventsEmitted = %d, want 10", plane.EventsEmitted.Load())
	}
	// Buffer is 2, so 8 should be dropped
	if plane.EventsDropped.Load() != 8 {
		t.Errorf("EventsDropped = %d, want 8", plane.EventsDropped.Load())
	}
}

func TestPlane_EmitWhenStopped(t *testing.T) {
	plane := NewPlane(DefaultConfig())
	// Never started — Emit should be a no-op
	plane.Emit(OfflineEvent{OrgID: "org1"})
	if plane.EventsEmitted.Load() != 0 {
		t.Errorf("should not emit when stopped, got %d", plane.EventsEmitted.Load())
	}
}

func TestPlane_FlushCalledOnStop(t *testing.T) {
	miner := &mockMiner{name: "flush-test"}
	plane := NewPlane(PlaneConfig{BufferSize: 100, WorkerCount: 1, FlushInterval: time.Hour}, miner)

	ctx := context.Background()
	plane.Start(ctx)
	time.Sleep(10 * time.Millisecond)
	plane.Stop()

	if miner.flushes.Load() < 1 {
		t.Error("Flush should have been called at least once on Stop")
	}
}

func TestPlane_Stats(t *testing.T) {
	plane := NewPlane(PlaneConfig{BufferSize: 100, WorkerCount: 1, FlushInterval: time.Hour})
	ctx := context.Background()
	plane.Start(ctx)

	plane.Emit(OfflineEvent{OrgID: "org1", Timestamp: time.Now()})
	time.Sleep(50 * time.Millisecond)

	stats := plane.Stats()
	if stats["emitted"] != 1 {
		t.Errorf("stats emitted = %d, want 1", stats["emitted"])
	}
	plane.Stop()
}
