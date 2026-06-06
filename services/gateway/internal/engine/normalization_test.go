package engine

import (
	"testing"
)

func TestNormalizePrompt(t *testing.T) {
	cases := []struct {
		name          string
		input         string
		wantCanonical string
	}{
		{
			name:          "basic lowercase and trim",
			input:         "  Hello World  ",
			wantCanonical: "hello world",
		},
		{
			name:          "numbers fully preserved",
			input:         "write123 and test45",
			wantCanonical: "write123 and test45",
		},
		{
			name:          "math parameters preserved",
			input:         "calculate a=10, b=15",
			wantCanonical: "calculate a10 b15",
		},
		{
			name:          "single digit remains",
			input:         "version 1 is good",
			wantCanonical: "version 1 is good",
		},
		{
			name:          "punctuation removal",
			input:         "Hello, world! What's up?",
			wantCanonical: "hello world whats up",
		},
		{
			name:          "multiple spaces normalized",
			input:         "hello    world",
			wantCanonical: "hello world",
		},
		{
			name:          "ticket IDs preserved (no masking)",
			input:         "Get me the ticket #4582 from the CRM, please.",
			wantCanonical: "get me the ticket 4582 from the crm please",
		},
		{
			name:          "ordinals preserved",
			input:         "What is the 15th fibonacci number?",
			wantCanonical: "what is the 15th fibonacci number",
		},
		{
			name:          "different numbers produce different canonical forms",
			input:         "what is (a+b)^2 where a=5, b=15",
			wantCanonical: "what is ab2 where a5 b15",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			canonical, hash := NormalizePrompt(tc.input)
			if canonical != tc.wantCanonical {
				t.Errorf("NormalizePrompt(%q) canonical = %q, want %q", tc.input, canonical, tc.wantCanonical)
			}
			if len(hash) != 64 {
				t.Errorf("NormalizePrompt(%q) hash length = %d, want 64", tc.input, len(hash))
			}
		})
	}
}
