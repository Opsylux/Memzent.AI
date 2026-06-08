package miners

import (
	"context"
	"testing"
	"time"

	"memzent-gateway/internal/offline"
)

func TestRequestMiner_BasicFrequency(t *testing.T) {
	m := NewRequestMiner(3) // threshold of 3

	ctx := context.Background()
	now := time.Now()

	// Emit same entity pattern 5 times
	for i := 0; i < 5; i++ {
		m.Process(ctx, offline.OfflineEvent{
			OrgID:      "org1",
			Entities:   map[string]string{"action": "transfer", "amount": "100"},
			CacheLayer: "L5",
			Timestamp:  now,
		})
	}

	// Emit different pattern 2 times (below threshold)
	for i := 0; i < 2; i++ {
		m.Process(ctx, offline.OfflineEvent{
			OrgID:      "org1",
			Entities:   map[string]string{"action": "balance", "customer": "Raj"},
			CacheLayer: "L1",
			Timestamp:  now,
		})
	}

	hot := m.GetHotPatterns()
	if len(hot) != 1 {
		t.Fatalf("expected 1 hot pattern, got %d", len(hot))
	}
	if hot[0].Frequency != 5 {
		t.Errorf("frequency = %d, want 5", hot[0].Frequency)
	}
	if hot[0].Entities["action"] != "transfer" {
		t.Errorf("expected action=transfer, got %s", hot[0].Entities["action"])
	}
}

func TestRequestMiner_CacheMissDetection(t *testing.T) {
	m := NewRequestMiner(2)
	ctx := context.Background()

	// Pattern that always misses (L5)
	for i := 0; i < 5; i++ {
		m.Process(ctx, offline.OfflineEvent{
			OrgID:      "org1",
			Entities:   map[string]string{"action": "lookup", "id": "123"},
			CacheLayer: "L5",
			Timestamp:  time.Now(),
		})
	}

	m.Flush(ctx)

	select {
	case output := <-m.Output():
		if len(output.CacheMissCandidates) != 1 {
			t.Fatalf("expected 1 cache miss candidate, got %d", len(output.CacheMissCandidates))
		}
		if !output.CacheMissCandidates[0].CacheMiss {
			t.Error("expected CacheMiss=true")
		}
	default:
		t.Fatal("expected output from miner")
	}
}

func TestRequestMiner_OrgIsolation(t *testing.T) {
	m := NewRequestMiner(2)
	ctx := context.Background()

	// Same entities, different orgs
	for i := 0; i < 3; i++ {
		m.Process(ctx, offline.OfflineEvent{
			OrgID:    "org1",
			Entities: map[string]string{"action": "transfer"},
			Timestamp: time.Now(),
		})
	}
	for i := 0; i < 3; i++ {
		m.Process(ctx, offline.OfflineEvent{
			OrgID:    "org2",
			Entities: map[string]string{"action": "transfer"},
			Timestamp: time.Now(),
		})
	}

	hot := m.GetHotPatterns()
	if len(hot) != 2 {
		t.Fatalf("expected 2 hot patterns (one per org), got %d", len(hot))
	}
}

func TestCacheMiner_TracksL5Misses(t *testing.T) {
	m := NewCacheMiner(3, 10)
	ctx := context.Background()

	// Same prompt hitting L5 repeatedly
	for i := 0; i < 5; i++ {
		m.Process(ctx, offline.OfflineEvent{
			OrgID:      "org1",
			PromptHash: "abc123",
			CacheLayer: "L5",
			Timestamp:  time.Now(),
		})
	}

	// L1 hit (should be ignored)
	m.Process(ctx, offline.OfflineEvent{
		OrgID:      "org1",
		PromptHash: "abc123",
		CacheLayer: "L1",
		Timestamp:  time.Now(),
	})

	misses := m.GetTopMisses()
	if len(misses) != 1 {
		t.Fatalf("expected 1 miss entry, got %d", len(misses))
	}
	if misses[0].MissCount != 5 {
		t.Errorf("miss count = %d, want 5", misses[0].MissCount)
	}
}

func TestCacheMiner_PredictionAccuracy(t *testing.T) {
	m := NewCacheMiner(5, 10)

	// No data
	if acc := m.PredictionAccuracy(); acc != -1 {
		t.Errorf("expected -1 with no data, got %f", acc)
	}

	// 3 hits, 1 miss
	m.RecordSpeculativeHit()
	m.RecordSpeculativeHit()
	m.RecordSpeculativeHit()
	m.RecordSpeculativeMiss()

	acc := m.PredictionAccuracy()
	if acc < 0.74 || acc > 0.76 {
		t.Errorf("expected ~0.75 accuracy, got %f", acc)
	}
}

func TestWorkflowMiner_DetectsSequences(t *testing.T) {
	m := NewWorkflowMiner(3, 0.90)
	ctx := context.Background()

	// Repeated tool sequence
	for i := 0; i < 5; i++ {
		m.Process(ctx, offline.OfflineEvent{
			OrgID:     "org1",
			ToolsUsed: []string{"search_customer", "get_ledger", "calc_balance"},
			Success:   true,
			Timestamp: time.Now(),
		})
	}

	// Single tool (should be ignored - need 2+ for a sequence)
	m.Process(ctx, offline.OfflineEvent{
		OrgID:     "org1",
		ToolsUsed: []string{"single_tool"},
		Success:   true,
		Timestamp: time.Now(),
	})

	sequences := m.GetDetectedSequences()
	if len(sequences) != 1 {
		t.Fatalf("expected 1 detected sequence, got %d", len(sequences))
	}
	if sequences[0].Frequency != 5 {
		t.Errorf("frequency = %d, want 5", sequences[0].Frequency)
	}
	if sequences[0].Pattern != "search_customer → get_ledger → calc_balance" {
		t.Errorf("unexpected pattern: %s", sequences[0].Pattern)
	}
	if sequences[0].SuccessRate != 1.0 {
		t.Errorf("success rate = %f, want 1.0", sequences[0].SuccessRate)
	}
}

func TestWorkflowMiner_PromotionCriteria(t *testing.T) {
	m := NewWorkflowMiner(3, 0.90)
	ctx := context.Background()

	// High success rate sequence
	for i := 0; i < 10; i++ {
		m.Process(ctx, offline.OfflineEvent{
			OrgID:     "org1",
			ToolsUsed: []string{"tool_a", "tool_b"},
			Success:   true,
			Timestamp: time.Now(),
		})
	}

	// Low success rate sequence
	for i := 0; i < 5; i++ {
		m.Process(ctx, offline.OfflineEvent{
			OrgID:     "org1",
			ToolsUsed: []string{"bad_a", "bad_b"},
			Success:   i < 2, // 2 success, 3 failure = 40% rate
			Timestamp: time.Now(),
		})
	}

	m.Flush(ctx)

	select {
	case output := <-m.Output():
		if len(output.DetectedSequences) != 2 {
			t.Fatalf("expected 2 detected, got %d", len(output.DetectedSequences))
		}
		if len(output.PromotionReady) != 1 {
			t.Fatalf("expected 1 promotion-ready, got %d", len(output.PromotionReady))
		}
		if output.PromotionReady[0].Pattern != "tool_a → tool_b" {
			t.Errorf("wrong promoted pattern: %s", output.PromotionReady[0].Pattern)
		}
	default:
		t.Fatal("expected output from miner")
	}
}
