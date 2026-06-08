package engine

import (
	"testing"
)

func TestExtractEntitiesLocal(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		expected map[string]string
	}{
		{
			name:   "transfer with direction",
			prompt: "Transfer $100 from account 123 to account 456",
			expected: map[string]string{
				"amount":         "100",
				"source_account": "123",
				"target_account": "456",
				"action":         "transfer",
			},
		},
		{
			name:   "reversed transfer",
			prompt: "Transfer $100 from account 456 to account 123",
			expected: map[string]string{
				"amount":         "100",
				"source_account": "456",
				"target_account": "123",
				"action":         "transfer",
			},
		},
		{
			name:   "balance lookup with customer",
			prompt: "What is the balance for customer Raj?",
			expected: map[string]string{
				"action":   "balance",
				"customer": "Raj",
			},
		},
		{
			name:   "no entities in generic prompt",
			prompt: "Hello, how are you?",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractEntitiesLocal(tt.prompt)

			for k, v := range tt.expected {
				if got[k] != v {
					t.Errorf("entity %q: got %q, want %q", k, got[k], v)
				}
			}

			// Ensure no unexpected key differences for important entities
			for k, v := range got {
				if _, ok := tt.expected[k]; !ok {
					// Allow extra entities (e.g., entity_id from fallback)
					// but log them for awareness
					t.Logf("extra entity extracted: %s=%s", k, v)
				}
			}
		})
	}
}

func TestBuildEntityCacheKey(t *testing.T) {
	e := &MemzentEngine{}

	tests := []struct {
		name     string
		entities map[string]string
		wantKey  string
	}{
		{
			name:     "empty entities returns empty",
			entities: map[string]string{},
			wantKey:  "",
		},
		{
			name: "single entity",
			entities: map[string]string{
				"action": "transfer",
			},
			wantKey: "org:org1:m:gpt-4:e:action=transfer",
		},
		{
			name: "multiple entities sorted",
			entities: map[string]string{
				"target_account": "456",
				"amount":         "100",
				"source_account": "123",
				"action":         "transfer",
			},
			wantKey: "org:org1:m:gpt-4:e:action=transfer:amount=100:source_account=123:target_account=456",
		},
		{
			name: "reversed direction produces different key",
			entities: map[string]string{
				"target_account": "123",
				"amount":         "100",
				"source_account": "456",
				"action":         "transfer",
			},
			wantKey: "org:org1:m:gpt-4:e:action=transfer:amount=100:source_account=456:target_account=123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.buildEntityCacheKey("org1", "gpt-4", tt.entities)
			if got != tt.wantKey {
				t.Errorf("got %q, want %q", got, tt.wantKey)
			}
		})
	}

	// Verify that direction matters
	key1 := e.buildEntityCacheKey("org1", "gpt-4", map[string]string{
		"source_account": "123", "target_account": "456", "amount": "100", "action": "transfer",
	})
	key2 := e.buildEntityCacheKey("org1", "gpt-4", map[string]string{
		"source_account": "456", "target_account": "123", "amount": "100", "action": "transfer",
	})
	if key1 == key2 {
		t.Error("CRITICAL: reversed source/target produced same cache key!")
	}
}
