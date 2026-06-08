package miners

import (
	"context"
	"log/slog"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"memzent-gateway/internal/offline"
)

// SpeculativeEntry represents a pre-warm candidate identified by O2.
type SpeculativeEntry struct {
	OrgID         string `json:"org_id"`
	PromptHash    string `json:"prompt_hash"`
	CanonicalHash string `json:"canonical_hash"`
	EntityKey     string `json:"entity_key"`
	MissCount     int64  `json:"miss_count"`
	LastMissAt    time.Time `json:"last_miss_at"`
}

// CacheMinerOutput holds O2's periodic analysis results.
type CacheMinerOutput struct {
	TopMisses      []SpeculativeEntry // Highest-frequency cache misses
	PreWarmTargets []SpeculativeEntry // Candidates eligible for speculative pre-warming
}

// CacheMiner is O2: tracks cache misses, identifies top-N miss patterns,
// and generates pre-warm candidates tagged speculative:true.
type CacheMiner struct {
	mu     sync.RWMutex
	misses map[string]*SpeculativeEntry // key: "org_id:prompt_hash"
	output chan CacheMinerOutput

	// Metrics
	SpeculativeHits   atomic.Uint64
	SpeculativeMisses atomic.Uint64

	// Config
	minMisses    int64 // minimum miss count to qualify for pre-warming
	topN         int   // how many top misses to report
}

// NewCacheMiner creates an O2 cache miner.
func NewCacheMiner(minMisses int64, topN int) *CacheMiner {
	if minMisses <= 0 {
		minMisses = 10
	}
	if topN <= 0 {
		topN = 100
	}
	return &CacheMiner{
		misses:    make(map[string]*SpeculativeEntry),
		output:    make(chan CacheMinerOutput, 1),
		minMisses: minMisses,
		topN:      topN,
	}
}

func (m *CacheMiner) Name() string { return "O2:CacheMiner" }

func (m *CacheMiner) Process(_ context.Context, event offline.OfflineEvent) {
	// Only care about L5 (cache misses that went to LLM)
	if event.CacheLayer != "L5" {
		return
	}
	if event.PromptHash == "" {
		return
	}

	compositeKey := event.OrgID + ":" + event.PromptHash

	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.misses[compositeKey]
	if !exists {
		entry = &SpeculativeEntry{
			OrgID:         event.OrgID,
			PromptHash:    event.PromptHash,
			CanonicalHash: event.CanonicalHash,
			EntityKey:     buildEntityKey(event.Entities),
		}
		m.misses[compositeKey] = entry
	}
	entry.MissCount++
	entry.LastMissAt = event.Timestamp
}

func (m *CacheMiner) Flush(_ context.Context) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var candidates []SpeculativeEntry
	for _, entry := range m.misses {
		if entry.MissCount >= m.minMisses {
			candidates = append(candidates, *entry)
		}
	}

	if len(candidates) == 0 {
		return
	}

	// Sort by miss count descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].MissCount > candidates[j].MissCount
	})

	// Cap to topN
	topMisses := candidates
	if len(topMisses) > m.topN {
		topMisses = topMisses[:m.topN]
	}

	output := CacheMinerOutput{
		TopMisses:      topMisses,
		PreWarmTargets: topMisses, // all top misses are pre-warm candidates
	}

	select {
	case m.output <- output:
	default:
		select {
		case <-m.output:
		default:
		}
		m.output <- output
	}

	slog.Info("📊 O2 CacheMiner flush", "total_misses_tracked", len(m.misses), "pre_warm_candidates", len(topMisses))
}

// Output returns the channel for consuming miner results.
func (m *CacheMiner) Output() <-chan CacheMinerOutput {
	return m.output
}

// GetTopMisses returns current top cache miss patterns (for API/dashboard).
func (m *CacheMiner) GetTopMisses() []SpeculativeEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var all []SpeculativeEntry
	for _, e := range m.misses {
		if e.MissCount >= m.minMisses {
			all = append(all, *e)
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].MissCount > all[j].MissCount })
	if len(all) > m.topN {
		all = all[:m.topN]
	}
	return all
}

// PredictionAccuracy returns the speculative hit/miss ratio.
// Returns -1 if no speculative entries have been tracked.
func (m *CacheMiner) PredictionAccuracy() float64 {
	hits := m.SpeculativeHits.Load()
	misses := m.SpeculativeMisses.Load()
	total := hits + misses
	if total == 0 {
		return -1
	}
	return float64(hits) / float64(total)
}

// RecordSpeculativeHit increments the speculative hit counter.
func (m *CacheMiner) RecordSpeculativeHit() {
	m.SpeculativeHits.Add(1)
}

// RecordSpeculativeMiss increments the speculative miss counter (expired without use).
func (m *CacheMiner) RecordSpeculativeMiss() {
	m.SpeculativeMisses.Add(1)
}
