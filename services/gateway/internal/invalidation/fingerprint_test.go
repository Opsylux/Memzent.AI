// services/gateway/internal/invalidation/fingerprint_test.go
package invalidation

import "testing"

func TestFingerprint_Deterministic(t *testing.T) {
	in := PreferenceInputs{Role: "admin", SystemPrompt: "You are a concise financial assistant."}
	a := Fingerprint(in)
	b := Fingerprint(in)
	if a != b {
		t.Fatalf("fingerprint not deterministic: %q vs %q", a, b)
	}
	if a == "" {
		t.Fatal("expected non-empty fingerprint")
	}
}

func TestFingerprint_IgnoresStopWordsAndCase(t *testing.T) {
	a := Fingerprint(PreferenceInputs{SystemPrompt: "You are a Concise assistant"})
	b := Fingerprint(PreferenceInputs{SystemPrompt: "concise assistant"})
	if a != b {
		t.Errorf("expected case/stopword-insensitive equality: %q vs %q", a, b)
	}
}

func TestPreferenceTag_EmptyForNoSignals(t *testing.T) {
	if tag := PreferenceTag("", ""); tag != "" {
		t.Errorf("no preference signals should yield empty tag, got %q", tag)
	}
}

func TestPreferenceTag_StableAndDistinct(t *testing.T) {
	admin := PreferenceTag("admin", "concise assistant")
	if admin == "" {
		t.Fatal("expected non-empty tag for role+system prompt")
	}
	// Deterministic for identical inputs.
	if again := PreferenceTag("admin", "concise assistant"); again != admin {
		t.Errorf("tag not stable: %q vs %q", admin, again)
	}
	// Different preferences produce a different tag (key partitioning).
	if viewer := PreferenceTag("viewer", "verbose assistant"); viewer == admin {
		t.Error("distinct preferences must produce distinct tags")
	}
}
