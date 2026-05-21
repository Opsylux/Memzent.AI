package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
)

// ToolMetric contains telemetry details for a specific connector tool
type ToolMetric struct {
	ToolID         string  `json:"tool_id"`
	ExecutionCount int     `json:"execution_count"`
	AvgLatencyMS   int     `json:"avg_latency_ms"`
	FailureRate    float64 `json:"failure_rate"`
}

// SavingsROI calculates cache vs LLM cost returns
type SavingsROI struct {
	CacheHits     int     `json:"cache_hits"`
	EstimatedSaved float64 `json:"estimated_saved"`
	LLMCost       float64 `json:"llm_cost"`
	NetROI        float64 `json:"net_roi"`
}

// IntentCluster represents high-frequency prompt intents resolved from audit logs
type IntentCluster struct {
	Intent    string `json:"intent"`
	Frequency int    `json:"frequency"`
}

// ContextAnalyticsResponse wraps all deep context metrics
type ContextAnalyticsResponse struct {
	ToolMetrics      []ToolMetric    `json:"tool_metrics"`
	SavingsROI       SavingsROI      `json:"savings_roi"`
	SemanticClusters []IntentCluster `json:"semantic_clusters"`
}

// TelemetryAggregator queries PostgreSQL to gather system execution analytics
type TelemetryAggregator struct {
	db *sql.DB
}

// NewTelemetryAggregator instantiates a telemetry aggregator
func NewTelemetryAggregator(db *sql.DB) *TelemetryAggregator {
	return &TelemetryAggregator{db: db}
}

// LogToolExecution records tool execution telemetry in the database
func (ta *TelemetryAggregator) LogToolExecution(ctx context.Context, orgID, toolID, requestID string, durationMS int, status, errMsg string) {
	if ta.db == nil {
		return
	}

	query := `
		INSERT INTO tool_executions (org_id, tool_id, request_id, duration_ms, status, error_message)
		VALUES ($1::uuid, $2, $3, $4, $5, $6)
	`
	go func() {
		bgCtx := context.Background()
		_, err := ta.db.ExecContext(bgCtx, query, orgID, toolID, requestID, durationMS, status, sql.NullString{
			String: errMsg,
			Valid:  errMsg != "",
		})
		if err != nil {
			slog.Error("Failed to log tool execution telemetry", "tool_id", toolID, "error", err)
		}
	}()
}

// GetContextAnalytics aggregates latency profiles, cost ROI, and intent counts for an org
func (ta *TelemetryAggregator) GetContextAnalytics(ctx context.Context, orgID string) (*ContextAnalyticsResponse, error) {
	if ta.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var response ContextAnalyticsResponse

	// 1. Query Tool Metrics
	toolQuery := `
		SELECT tool_id, 
		       COUNT(*) as execution_count, 
		       COALESCE(ROUND(AVG(duration_ms)), 0) as avg_latency_ms,
		       COALESCE(ROUND(SUM(CASE WHEN status = 'failure' THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 1), 0.0) as failure_rate
		FROM tool_executions
		WHERE org_id = $1::uuid
		GROUP BY tool_id
		ORDER BY execution_count DESC
		LIMIT 10
	`
	rows, err := ta.db.QueryContext(ctx, toolQuery, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tool metrics telemetry: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tm ToolMetric
		if err := rows.Scan(&tm.ToolID, &tm.ExecutionCount, &tm.AvgLatencyMS, &tm.FailureRate); err != nil {
			return nil, fmt.Errorf("failed to scan tool metric: %w", err)
		}
		response.ToolMetrics = append(response.ToolMetrics, tm)
	}

	// 2. Query Savings ROI
	// Cache Hits count
	var cacheHits int
	cacheHitsQuery := `
		SELECT COUNT(*) 
		FROM audit_logs 
		WHERE org_id = $1::uuid AND action LIKE 'CACHE:%'
	`
	err = ta.db.QueryRowContext(ctx, cacheHitsQuery, orgID).Scan(&cacheHits)
	if err != nil {
		slog.Warn("Failed to query cache hits analytics", "error", err)
		cacheHits = 0
	}

	// Total LLM expenditure
	var llmCost float64
	llmCostQuery := `
		SELECT COALESCE(SUM(-amount), 0.0) 
		FROM billing_ledger 
		WHERE org_id = $1::uuid AND transaction_type = 'llm_usage'
	`
	err = ta.db.QueryRowContext(ctx, llmCostQuery, orgID).Scan(&llmCost)
	if err != nil {
		slog.Warn("Failed to query LLM costs analytics", "error", err)
		llmCost = 0.0
	}

	// Estimated savings: assuming average cache prompt is ~1000 tokens.
	// Saving = $0.002 per hit (Ollama is free locally, but for commercial tier estimation we use standard LLM rates).
	estSavedPerHit := 0.002
	estimatedSaved := float64(cacheHits) * estSavedPerHit

	var netROI float64
	if llmCost > 0 {
		netROI = (estimatedSaved / llmCost) * 100.0
	} else if estimatedSaved > 0 {
		netROI = 100.0 // Infinite ROI if no LLM cost was incurred due to perfect caching
	}

	response.SavingsROI = SavingsROI{
		CacheHits:      cacheHits,
		EstimatedSaved: estimatedSaved,
		LLMCost:        llmCost,
		NetROI:         netROI,
	}

	// 3. Query Semantic Intent Clusters
	intentQuery := `
		SELECT COALESCE(metadata->>'prompt', 'System Intent'), COUNT(*) as frequency
		FROM audit_logs
		WHERE org_id = $1::uuid AND action LIKE 'CACHE:%' AND metadata->>'prompt' IS NOT NULL
		GROUP BY metadata->>'prompt'
		ORDER BY frequency DESC
		LIMIT 5
	`
	rowsIntents, err := ta.db.QueryContext(ctx, intentQuery, orgID)
	if err == nil {
		defer rowsIntents.Close()
		for rowsIntents.Next() {
			var ic IntentCluster
			if err := rowsIntents.Scan(&ic.Intent, &ic.Frequency); err == nil {
				// Abbreviate long intents for graphing friendliness
				if len(ic.Intent) > 55 {
					ic.Intent = ic.Intent[:52] + "..."
				}
				response.SemanticClusters = append(response.SemanticClusters, ic)
			}
		}
	} else {
		slog.Warn("Failed to query semantic intent cluster telemetry", "error", err)
	}

	return &response, nil
}
