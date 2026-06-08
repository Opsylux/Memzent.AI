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

// ToolSequence represents a detected tool execution pattern.
type ToolSequence struct {
	OrgID       string   `json:"org_id"`
	Tools       []string `json:"tools"`        // ordered tool IDs
	Pattern     string   `json:"pattern"`      // human-readable "toolA → toolB → toolC"
	Frequency   int64    `json:"frequency"`
	SuccessRate float64  `json:"success_rate"` // % of executions that succeeded
	LastSeen    time.Time `json:"last_seen"`
	TotalSuccesses int64 `json:"total_successes"`
	TotalFailures  int64 `json:"total_failures"`
}

// WorkflowMinerOutput holds O3's periodic analysis results.
type WorkflowMinerOutput struct {
	DetectedSequences []ToolSequence // All sequences above threshold
	PromotionReady    []ToolSequence // Sequences meeting promotion criteria (freq + success rate)
}

// WorkflowMiner is O3: detects repeated tool execution sequences across requests.
// Counts frequencies and outputs workflow candidates when threshold is met.
type WorkflowMiner struct {
	mu        sync.RWMutex
	sequences map[string]*ToolSequence // key: "org_id:toolA→toolB→toolC"
	output    chan WorkflowMinerOutput

	// Config
	minFrequency   int64   // minimum executions to report (default: 100)
	minSuccessRate float64 // minimum success rate for promotion (default: 0.90)
}

// NewWorkflowMiner creates an O3 workflow miner.
func NewWorkflowMiner(minFrequency int64, minSuccessRate float64) *WorkflowMiner {
	if minFrequency <= 0 {
		minFrequency = 100
	}
	if minSuccessRate <= 0 {
		minSuccessRate = 0.90
	}
	return &WorkflowMiner{
		sequences:      make(map[string]*ToolSequence),
		output:         make(chan WorkflowMinerOutput, 1),
		minFrequency:   minFrequency,
		minSuccessRate: minSuccessRate,
	}
}

func (m *WorkflowMiner) Name() string { return "O3:WorkflowMiner" }

func (m *WorkflowMiner) Process(_ context.Context, event offline.OfflineEvent) {
	// Only care about requests that actually used tools
	if len(event.ToolsUsed) < 2 {
		return // single tool or no tools — not a sequence
	}

	pattern := strings.Join(event.ToolsUsed, " → ")
	compositeKey := event.OrgID + ":" + pattern

	m.mu.Lock()
	defer m.mu.Unlock()

	seq, exists := m.sequences[compositeKey]
	if !exists {
		seq = &ToolSequence{
			OrgID:   event.OrgID,
			Tools:   event.ToolsUsed,
			Pattern: pattern,
		}
		m.sequences[compositeKey] = seq
	}

	seq.Frequency++
	seq.LastSeen = event.Timestamp
	if event.Success {
		seq.TotalSuccesses++
	} else {
		seq.TotalFailures++
	}
	total := seq.TotalSuccesses + seq.TotalFailures
	if total > 0 {
		seq.SuccessRate = float64(seq.TotalSuccesses) / float64(total)
	}
}

func (m *WorkflowMiner) Flush(_ context.Context) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var detected []ToolSequence
	var promotionReady []ToolSequence

	for _, seq := range m.sequences {
		if seq.Frequency >= m.minFrequency {
			detected = append(detected, *seq)
			if seq.SuccessRate >= m.minSuccessRate {
				promotionReady = append(promotionReady, *seq)
			}
		}
	}

	if len(detected) == 0 {
		return
	}

	// Sort by frequency descending
	sort.Slice(detected, func(i, j int) bool { return detected[i].Frequency > detected[j].Frequency })
	sort.Slice(promotionReady, func(i, j int) bool { return promotionReady[i].Frequency > promotionReady[j].Frequency })

	output := WorkflowMinerOutput{
		DetectedSequences: detected,
		PromotionReady:    promotionReady,
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

	slog.Info("📊 O3 WorkflowMiner flush",
		"detected_sequences", len(detected),
		"promotion_ready", len(promotionReady),
	)
}

// Output returns the channel for consuming miner results.
func (m *WorkflowMiner) Output() <-chan WorkflowMinerOutput {
	return m.output
}

// GetDetectedSequences returns current detected tool sequences (for API/dashboard).
func (m *WorkflowMiner) GetDetectedSequences() []ToolSequence {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var all []ToolSequence
	for _, seq := range m.sequences {
		if seq.Frequency >= m.minFrequency {
			all = append(all, *seq)
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Frequency > all[j].Frequency })
	return all
}
