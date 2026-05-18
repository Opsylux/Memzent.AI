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
			name:          "number masking 2+ digits",
			input:         "write123 and test45",
			wantCanonical: "write<id> and test<id>",
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
			name:          "complex real world prompt",
			input:         "Get me the ticket #4582 from the CRM, please.",
			wantCanonical: "get me the ticket <id> from the crm please",
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
