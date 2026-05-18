package billing

import (
	"os"
	"testing"
)

func TestNewCostCalculator_DefaultRates(t *testing.T) {
	c := NewCostCalculator()

	tests := []struct {
		key      string
		wantIn   float64
		wantOut  float64
	}{
		{"openai:gpt-4o-mini", 0.150, 0.600},
		{"openai:gpt-4o", 5.00, 15.00},
		{"anthropic:claude-3-5-sonnet-20240620", 3.00, 15.00},
		{"gemini:gemini-1.5-flash", 0.075, 0.300},
		{"ollama", 0.05, 0.05},
	}

	for _, tt := range tests {
		gotIn, ok := c.Rates[tt.key]
		if !ok {
			t.Errorf("Rates[%q]: key not found", tt.key)
			continue
		}
		if gotIn != tt.wantIn {
			t.Errorf("Rates[%q] = %v, want %v", tt.key, gotIn, tt.wantIn)
		}

		gotOut, ok := c.OutputRates[tt.key]
		if !ok {
			t.Errorf("OutputRates[%q]: key not found", tt.key)
			continue
		}
		if gotOut != tt.wantOut {
			t.Errorf("OutputRates[%q] = %v, want %v", tt.key, gotOut, tt.wantOut)
		}
	}
}

func TestNewCostCalculator_DynamicOllamaRate(t *testing.T) {
	os.Setenv("OLLAMA_BASE_COST_PER_1M_TOKENS", "0.10")
	defer os.Unsetenv("OLLAMA_BASE_COST_PER_1M_TOKENS")

	c := NewCostCalculator()
	if c.Rates["ollama"] != 0.10 {
		t.Errorf("Expected dynamic ollama rate 0.10, got %v", c.Rates["ollama"])
	}
}

func TestCalculateCost_KnownModel(t *testing.T) {
	c := NewCostCalculator()

	// 1M input + 1M output tokens @ gpt-4o-mini: 0.150 + 0.600 = 0.750
	got := c.CalculateCost("openai", "gpt-4o-mini", 1_000_000, 1_000_000)
	want := 0.750
	if got != want {
		t.Errorf("CalculateCost(openai, gpt-4o-mini, 1M, 1M) = %v, want %v", got, want)
	}
}

func TestCalculateCost_CaseInsensitive(t *testing.T) {
	c := NewCostCalculator()
	got1 := c.CalculateCost("OpenAI", "GPT-4O-MINI", 1_000_000, 0)
	got2 := c.CalculateCost("openai", "gpt-4o-mini", 1_000_000, 0)
	if got1 != got2 {
		t.Errorf("Case sensitivity issue: %v != %v", got1, got2)
	}
}

func TestCalculateCost_UnknownModelFallsBackToProvider(t *testing.T) {
	c := NewCostCalculator()
	// Unknown model should fall back to provider-level rate for ollama
	got := c.CalculateCost("ollama", "unknown-model-xyz", 1_000_000, 0)
	if got <= 0 {
		t.Errorf("Expected fallback to provider rate, got %v", got)
	}
}

func TestCalculateCost_ZeroTokens(t *testing.T) {
	c := NewCostCalculator()
	got := c.CalculateCost("openai", "gpt-4o", 0, 0)
	if got != 0.0 {
		t.Errorf("Expected 0 cost for 0 tokens, got %v", got)
	}
}

func TestCalculateCacheDiscount_DefaultRate(t *testing.T) {
	os.Unsetenv("CACHE_DISCOUNT_PERCENTAGE")
	c := NewCostCalculator()

	// Discount reduces cost to 20% of full cost (80% off)
	fullCost := c.CalculateCost("openai", "gpt-4o-mini", 10_000, 100)
	discounted := c.CalculateCacheDiscount("openai", "gpt-4o-mini", 10_000)
	expectedDiscount := fullCost * 0.20

	diff := discounted - expectedDiscount
	if diff < -0.0001 || diff > 0.0001 {
		t.Errorf("CalculateCacheDiscount mismatch: got %v, want ~%v", discounted, expectedDiscount)
	}
}

func TestCalculateCacheDiscount_CustomRate(t *testing.T) {
	os.Setenv("CACHE_DISCOUNT_PERCENTAGE", "50")
	defer os.Unsetenv("CACHE_DISCOUNT_PERCENTAGE")

	c := NewCostCalculator()
	fullCost := c.CalculateCost("openai", "gpt-4o-mini", 10_000, 100)
	discounted := c.CalculateCacheDiscount("openai", "gpt-4o-mini", 10_000)
	expectedDiscount := fullCost * 0.50

	diff := discounted - expectedDiscount
	if diff < -0.0001 || diff > 0.0001 {
		t.Errorf("Custom discount: got %v, want ~%v", discounted, expectedDiscount)
	}
}
