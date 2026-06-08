package billing

import (
	"os"
	"strconv"
	"strings"
)

// CostCalculator handles computing costs based on LLM models and token usage
type CostCalculator struct {
	Rates       map[string]float64 // provider/model -> cost per 1m tokens (input)
	OutputRates map[string]float64
}

func NewCostCalculator() *CostCalculator {
	c := &CostCalculator{
		Rates:       make(map[string]float64),
		OutputRates: make(map[string]float64),
	}

	// OpenAI Rates
	c.Rates["openai:gpt-4o-mini"] = 0.150   // $0.150 per 1M input tokens
	c.OutputRates["openai:gpt-4o-mini"] = 0.600 // $0.600 per 1M output tokens
	
	c.Rates["openai:gpt-4o"] = 5.00
	c.OutputRates["openai:gpt-4o"] = 15.00

	// Anthropic Rates
	c.Rates["anthropic:claude-3-5-sonnet-20240620"] = 3.00
	c.OutputRates["anthropic:claude-3-5-sonnet-20240620"] = 15.00

	// Gemini Rates
	c.Rates["gemini:gemini-1.5-flash"] = 0.075
	c.OutputRates["gemini:gemini-1.5-flash"] = 0.300

	// Ollama Infra Pricing
	// We read a dynamic infra cost from environment, default to $0.05/1M tokens
	ollamaCostStr := os.Getenv("OLLAMA_BASE_COST_PER_1M_TOKENS")
	ollamaCost := 0.05
	if parsed, err := strconv.ParseFloat(ollamaCostStr, 64); err == nil {
		ollamaCost = parsed
	}
	c.Rates["ollama"] = ollamaCost
	c.OutputRates["ollama"] = ollamaCost

	return c
}

// CalculateCost computes total cost based on token usage.
// Returns fractional dollars (e.g. 0.0001)
func (c *CostCalculator) CalculateCost(provider, model string, promptTokens, completionTokens int) float64 {
	key := strings.ToLower(provider) + ":" + strings.ToLower(model)
	inRate, ok := c.Rates[key]
	if !ok {
		inRate = c.Rates[strings.ToLower(provider)]
	}

	outRate, ok := c.OutputRates[key]
	if !ok {
		outRate = c.OutputRates[strings.ToLower(provider)]
	}

	inCost := (float64(promptTokens) / 1000000.0) * inRate
	outCost := (float64(completionTokens) / 1000000.0) * outRate

	return inCost + outCost
}

func (c *CostCalculator) CalculateCacheDiscount(provider, model string, estimatedPromptTokens int) float64 {
	// Determine dynamic cache discount rate from Env
	discountRateStr := os.Getenv("CACHE_DISCOUNT_PERCENTAGE")
	discountRate := 80.0 // Default 80% discount for cache hits
	if parsed, err := strconv.ParseFloat(discountRateStr, 64); err == nil {
		discountRate = parsed
	}

	// Cost if it were to hit the provider
	baseCost := c.CalculateCost(provider, model, estimatedPromptTokens, 100) // Assume 100 completion tokens average

	discountMultiplier := (100.0 - discountRate) / 100.0
	return baseCost * discountMultiplier
}
