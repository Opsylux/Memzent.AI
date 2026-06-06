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
			name:          "long number masking 4+ digits",
			input:         "write12345 and test45",
			wantCanonical: "write<id> and test45",
		},
		{
			name:          "short numbers preserved",
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
			name:          "hash ID masked",
			input:         "Get me the ticket #4582 from the CRM, please.",
			wantCanonical: "get me the ticket <id> from the crm please",
		},
		{
			name:          "short hash ID masked",
			input:         "fix issue #45 now",
			wantCanonical: "fix issue <id> now",
		},
		{
			name:          "ordinals preserved",
			input:         "What is the 15th fibonacci number?",
			wantCanonical: "what is the 15th fibonacci number",
		},
		{
			name:          "math parameters preserved and differentiated",
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
