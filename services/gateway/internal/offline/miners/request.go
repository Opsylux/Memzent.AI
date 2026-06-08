package miners

import (
	"context"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"memzent-gateway/internal/offline"
)

// EntityPattern tracks a recurring entity pattern and its frequency.
type EntityPattern struct {
	OrgID     string            `json:"org_id"`
	Entities  map[string]string `json:"entities"`
	Key       string            `json:"key"`       // deterministic sorted entity key
	Frequency int64             `json:"frequency"`
	LastSeen  time.Time         `json:"last_seen"`
	CacheMiss bool              `json:"cache_miss"` // true if this pattern tends to miss cache
}

// RequestMinerOutput holds the O1 miner's periodic output.
type RequestMinerOutput struct {
	HotPatterns    []EntityPattern // High-frequency entity patterns
	CacheMissCandidates []EntityPattern // Patterns that frequently miss cache → L1b pre-warm candidates
}

// RequestMiner is O1: aggregates request frequencies by entity pattern per org.
// Detects hot prompts that always miss cache and outputs L1b pre-warm candidates.
type RequestMiner struct {
	mu       sync.RWMutex
	patterns map[string]*EntityPattern // key: "org_id:entity_key" → pattern
	output   chan RequestMinerOutput   // consumers can read miner output

	// Config
	minFrequency int64 // minimum hits to consider a pattern "hot" (default: 50)
}

// NewRequestMiner creates an O1 request miner.
func NewRequestMiner(minFrequency int64) *RequestMiner {
	if minFrequency <= 0 {
		minFrequency = 50
	}
	return &RequestMiner{
		patterns:     make(map[string]*EntityPattern),
		output:       make(chan RequestMinerOutput, 1),
		minFrequency: minFrequency,
	}
}

func (m *RequestMiner) Name() string { return "O1:RequestMiner" }

func (m *RequestMiner) Process(_ context.Context, event offline.OfflineEvent) {
	if len(event.Entities) == 0 {
		return // nothing to mine
	}

	entityKey := buildEntityKey(event.Entities)
	compositeKey := event.OrgID + ":" + entityKey
	isCacheMiss := event.CacheLayer == "L5"

	m.mu.Lock()
	defer m.mu.Unlock()

	p, exists := m.patterns[compositeKey]
	if !exists {
		p = &EntityPattern{
			OrgID:    event.OrgID,
			Entities: event.Entities,
			Key:      entityKey,
		}
		m.patterns[compositeKey] = p
	}
	p.Frequency++
	p.LastSeen = event.Timestamp
	if isCacheMiss {
		p.CacheMiss = true
	}
}

func (m *RequestMiner) Flush(_ context.Context) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var hot []EntityPattern
	var missCandidates []EntityPattern

	for _, p := range m.patterns {
		if p.Frequency >= m.minFrequency {
			hot = append(hot, *p)
			if p.CacheMiss {
				missCandidates = append(missCandidates, *p)
			}
		}
	}

	if len(hot) == 0 {
		return
	}

	// Sort by frequency descending
	sort.Slice(hot, func(i, j int) bool { return hot[i].Frequency > hot[j].Frequency })
	sort.Slice(missCandidates, func(i, j int) bool { return missCandidates[i].Frequency > missCandidates[j].Frequency })

	output := RequestMinerOutput{
		HotPatterns:         hot,
		CacheMissCandidates: missCandidates,
	}

	// Non-blocking output send
	select {
	case m.output <- output:
	default:
		// Previous output not consumed yet — overwrite
		select {
		case <-m.output:
		default:
		}
		m.output <- output
	}

	slog.Info("📊 O1 RequestMiner flush", "hot_patterns", len(hot), "miss_candidates", len(missCandidates))
}

// Output returns the channel for consuming miner results.
func (m *RequestMiner) Output() <-chan RequestMinerOutput {
	return m.output
}

// GetHotPatterns returns current hot patterns (for API/dashboard).
func (m *RequestMiner) GetHotPatterns() []EntityPattern {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var hot []EntityPattern
	for _, p := range m.patterns {
		if p.Frequency >= m.minFrequency {
			hot = append(hot, *p)
		}
	}
	sort.Slice(hot, func(i, j int) bool { return hot[i].Frequency > hot[j].Frequency })
	return hot
}

// buildEntityKey creates a deterministic string from entity map (sorted keys).
func buildEntityKey(entities map[string]string) string {
	keys := make([]string, 0, len(entities))
	for k := range entities {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte(':')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(entities[k])
	}
	return sb.String()
}
