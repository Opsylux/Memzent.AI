// services/gateway/internal/invalidation/fingerprint_test.go
package invalidation

import "testing"

func TestFingerprint_Deterministic(t *testing.T) {
	in := PreferenceInputs{Role: "admin", Provider: "openai", Model: "gpt-4o", SystemPrompt: "You are a concise financial assistant."}
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

func TestSimilarity_IdenticalAndDisjoint(t *testing.T) {
	fp := Fingerprint(PreferenceInputs{Role: "admin", Model: "gpt-4o"})
	if s := Similarity(fp, fp); s != 1.0 {
		t.Errorf("identical fingerprints similarity = %v, want 1.0", s)
	}
	if s := Similarity("role:admin", "role:viewer provider:openai"); s != 0.0 {
		t.Errorf("disjoint fingerprints similarity = %v, want 0.0", s)
	}
	if s := Similarity("", ""); s != 1.0 {
		t.Errorf("empty fingerprints similarity = %v, want 1.0", s)
	}
}

func TestSimilarity_PartialDrift(t *testing.T) {
	// Same role+model, different persona wording -> partial overlap, < 1.0.
	a := Fingerprint(PreferenceInputs{Role: "admin", Model: "gpt-4o", SystemPrompt: "concise helpful assistant"})
	b := Fingerprint(PreferenceInputs{Role: "admin", Model: "gpt-4o", SystemPrompt: "verbose detailed assistant"})
	s := Similarity(a, b)
	if s <= 0.0 || s >= 1.0 {
		t.Errorf("partial drift similarity = %v, want between 0 and 1", s)
	}
}
