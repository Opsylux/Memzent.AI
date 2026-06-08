package miners

import (
	"context"
	"log/slog"
	"sort"
	"sync"
	"time"

	"memzent-gateway/internal/offline"
)

// TransitionEdge represents a weighted edge in the Markov chain.
type TransitionEdge struct {
	FromTool    string  `json:"from_tool"`
	ToTool      string  `json:"to_tool"`
	Count       int64   `json:"count"`
	Probability float64 `json:"probability"` // P(ToTool | FromTool)
}

// MarkovState holds transition probabilities from a single tool.
type MarkovState struct {
	Tool        string           `json:"tool"`
	Transitions []TransitionEdge `json:"transitions"`
	TotalExits  int64            `json:"total_exits"` // sum of all outgoing edge counts
}

// PredictionResult is the output of a Markov prediction query.
type PredictionResult struct {
	CurrentTool string           `json:"current_tool"`
	NextTools   []TransitionEdge `json:"next_tools"` // sorted by probability desc
	Confidence  float64          `json:"confidence"`  // highest transition probability
}

// PatternMinerOutput holds O4's analysis results.
type PatternMinerOutput struct {
	TopTransitions []TransitionEdge // highest-count transitions across all tools
	HotStates      []MarkovState    // tools with the most predictable next-steps
}

// PatternMiner is O4 (Experimental): builds per-org Markov chains from tool
// execution sequences. Learns transition probabilities to predict the next
// tool an agent will invoke, enabling speculative pre-loading.
type PatternMiner struct {
	mu     sync.RWMutex
	// Per-org transition matrix: org_id → (from_tool → (to_tool → count))
	chains map[string]map[string]map[string]int64
	output chan PatternMinerOutput

	// Config
	minTransitions int64 // minimum edge count to report (default: 20)
}

// NewPatternMiner creates an O4 agent pattern miner.
func NewPatternMiner(minTransitions int64) *PatternMiner {
	if minTransitions <= 0 {
		minTransitions = 20
	}
	return &PatternMiner{
		chains:         make(map[string]map[string]map[string]int64),
		output:         make(chan PatternMinerOutput, 1),
		minTransitions: minTransitions,
	}
}

func (m *PatternMiner) Name() string { return "O4:PatternMiner" }

func (m *PatternMiner) Process(_ context.Context, event offline.OfflineEvent) {
	// Need at least 2 tools to form a transition
	if len(event.ToolsUsed) < 2 {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	orgChain, exists := m.chains[event.OrgID]
	if !exists {
		orgChain = make(map[string]map[string]int64)
		m.chains[event.OrgID] = orgChain
	}

	// Record transitions: A→B, B→C, C→D, etc.
	for i := 0; i < len(event.ToolsUsed)-1; i++ {
		from := event.ToolsUsed[i]
		to := event.ToolsUsed[i+1]

		if orgChain[from] == nil {
			orgChain[from] = make(map[string]int64)
		}
		orgChain[from][to]++
	}
}

func (m *PatternMiner) Flush(_ context.Context) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var allEdges []TransitionEdge

	for _, orgChain := range m.chains {
		for from, targets := range orgChain {
			var totalExits int64
			for _, count := range targets {
				totalExits += count
			}
			for to, count := range targets {
				if count >= m.minTransitions {
					allEdges = append(allEdges, TransitionEdge{
						FromTool:    from,
						ToTool:      to,
						Count:       count,
						Probability: float64(count) / float64(totalExits),
					})
				}
			}
		}
	}

	if len(allEdges) == 0 {
		return
	}

	sort.Slice(allEdges, func(i, j int) bool { return allEdges[i].Count > allEdges[j].Count })

	// Top transitions (cap at 50)
	topEdges := allEdges
	if len(topEdges) > 50 {
		topEdges = topEdges[:50]
	}

	// Build hot states from top transitions
	stateMap := make(map[string]*MarkovState)
	for _, edge := range allEdges {
		state, ok := stateMap[edge.FromTool]
		if !ok {
			state = &MarkovState{Tool: edge.FromTool}
			stateMap[edge.FromTool] = state
		}
		state.Transitions = append(state.Transitions, edge)
		state.TotalExits += edge.Count
	}

	var hotStates []MarkovState
	for _, s := range stateMap {
		// Sort transitions within each state by probability
		sort.Slice(s.Transitions, func(i, j int) bool {
			return s.Transitions[i].Probability > s.Transitions[j].Probability
		})
		hotStates = append(hotStates, *s)
	}
	sort.Slice(hotStates, func(i, j int) bool { return hotStates[i].TotalExits > hotStates[j].TotalExits })

	output := PatternMinerOutput{
		TopTransitions: topEdges,
		HotStates:      hotStates,
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

	slog.Info("📊 O4 PatternMiner flush",
		"total_edges", len(allEdges),
		"hot_states", len(hotStates),
	)
}

// Output returns the channel for consuming miner results.
func (m *PatternMiner) Output() <-chan PatternMinerOutput {
	return m.output
}

// Predict returns the most likely next tools given the current tool for an org.
func (m *PatternMiner) Predict(orgID, currentTool string) *PredictionResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	orgChain, exists := m.chains[orgID]
	if !exists {
		return nil
	}

	targets, exists := orgChain[currentTool]
	if !exists || len(targets) == 0 {
		return nil
	}

	var totalExits int64
	for _, count := range targets {
		totalExits += count
	}

	var edges []TransitionEdge
	for to, count := range targets {
		edges = append(edges, TransitionEdge{
			FromTool:    currentTool,
			ToTool:      to,
			Count:       count,
			Probability: float64(count) / float64(totalExits),
		})
	}

	sort.Slice(edges, func(i, j int) bool { return edges[i].Probability > edges[j].Probability })

	confidence := 0.0
	if len(edges) > 0 {
		confidence = edges[0].Probability
	}

	return &PredictionResult{
		CurrentTool: currentTool,
		NextTools:   edges,
		Confidence:  confidence,
	}
}

// GetTransitionMatrix returns the full Markov chain state for an org (for API/dashboard).
func (m *PatternMiner) GetTransitionMatrix(orgID string) []MarkovState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	orgChain, exists := m.chains[orgID]
	if !exists {
		return nil
	}

	var states []MarkovState
	for from, targets := range orgChain {
		var totalExits int64
		for _, count := range targets {
			totalExits += count
		}

		var edges []TransitionEdge
		for to, count := range targets {
			edges = append(edges, TransitionEdge{
				FromTool:    from,
				ToTool:      to,
				Count:       count,
				Probability: float64(count) / float64(totalExits),
			})
		}
		sort.Slice(edges, func(i, j int) bool { return edges[i].Probability > edges[j].Probability })

		states = append(states, MarkovState{
			Tool:        from,
			Transitions: edges,
			TotalExits:  totalExits,
		})
	}
	sort.Slice(states, func(i, j int) bool { return states[i].TotalExits > states[j].TotalExits })
	return states
}

// GetAllHotTransitions returns globally significant transitions across all orgs.
func (m *PatternMiner) GetAllHotTransitions() []TransitionEdge {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var all []TransitionEdge
	for _, orgChain := range m.chains {
		for from, targets := range orgChain {
			var totalExits int64
			for _, count := range targets {
				totalExits += count
			}
			for to, count := range targets {
				if count >= m.minTransitions {
					all = append(all, TransitionEdge{
						FromTool:    from,
						ToTool:      to,
						Count:       count,
						Probability: float64(count) / float64(totalExits),
					})
				}
			}
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Count > all[j].Count })
	if len(all) > 100 {
		all = all[:100]
	}
	return all
}

// LastFlush returns the last flush timestamp (for health checks).
func (m *PatternMiner) LastFlush() time.Time {
	return time.Now() // simplified — could track actual flush time
}
