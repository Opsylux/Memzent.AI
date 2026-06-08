package workflow

import "context"

// EngineAdapter wraps the Registry to implement the engine.WorkflowRegistry interface.
type EngineAdapter struct {
	registry *Registry
}

// NewEngineAdapter creates an adapter for use by the engine.
func NewEngineAdapter(r *Registry) *EngineAdapter {
	return &EngineAdapter{registry: r}
}

// MatchWorkflow checks for an active workflow matching the given tool pattern.
func (a *EngineAdapter) MatchWorkflow(ctx context.Context, orgID string, toolPattern string) (bool, []string, string, error) {
	c, err := a.registry.MatchWorkflow(ctx, orgID, toolPattern)
	if err != nil {
		return false, nil, "", err
	}
	if c == nil {
		return false, nil, "", nil
	}
	return true, c.ToolIDs, c.ID, nil
}

// RecordExecution logs a workflow execution.
func (a *EngineAdapter) RecordExecution(ctx context.Context, workflowID, orgID, promptHash string, entities map[string]string, success bool, latencyMs int) error {
	return a.registry.RecordExecution(ctx, workflowID, orgID, promptHash, entities, success, latencyMs)
}
