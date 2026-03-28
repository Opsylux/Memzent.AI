package llm

import (
	"context"
	"fmt"
)

type MockProvider struct{}

func NewMockProvider() Provider { return &MockProvider{} }

func (m *MockProvider) GetProviderName() string { return "Mock-Aura" }

func (m *MockProvider) Generate(ctx context.Context, prompt string, tools []any) (string, error) {
	response := fmt.Sprintf("Aura Local Engine (Zero-Cost Execution)\n")
	response += fmt.Sprintf("====================================================\n")
	response += fmt.Sprintf("Intent Analyzed: '%s'\n", prompt)
	
	if len(tools) > 0 {
		response += fmt.Sprintf("\n[MCP Execution Trace]\n")
		for i, t := range tools {
			response += fmt.Sprintf("%d. successfully ingested tool context: %v\n", i+1, t)
		}
		response += "\nSynthesis: The requested metrics indicate that the project repository and its backend infrastructure are completely stable and performing optimally."
	} else {
		response +=("\nSynthesis: No specialized MCP tools were required to route this prompt. Defaulting to standard conversational fallback.")
	}

	return response, nil
}
