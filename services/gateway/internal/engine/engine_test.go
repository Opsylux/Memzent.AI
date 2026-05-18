package engine

import (
	"testing"
)

// ---------------------------------------------------------------------------
// buildCacheKey
// ---------------------------------------------------------------------------

// We test buildCacheKey with a minimal stub of MemzentEngine since the method
// has no external dependencies.

func newTestEngine() *MemzentEngine {
	return &MemzentEngine{}
}

func TestBuildCacheKey_Format(t *testing.T) {
	e := newTestEngine()

	got := e.buildCacheKey("org123", "lit", "gpt-4o", "hello world")
	want := "org:org123:m:gpt-4o:lit:hello world"
	if got != want {
		t.Errorf("buildCacheKey = %q, want %q", got, want)
	}
}

func TestBuildCacheKey_TypeVariants(t *testing.T) {
	e := newTestEngine()
	cases := []struct {
		keyType string
		want    string
	}{
		{"lit", "org:o1:m:m1:lit:v"},
		{"can", "org:o1:m:m1:can:v"},
		{"sem", "org:o1:m:m1:sem:v"},
	}
	for _, tc := range cases {
		got := e.buildCacheKey("o1", tc.keyType, "m1", "v")
		if got != tc.want {
			t.Errorf("buildCacheKey(%q) = %q, want %q", tc.keyType, got, tc.want)
		}
	}
}

func TestBuildCacheKey_IsolateDifferentOrgs(t *testing.T) {
	e := newTestEngine()
	k1 := e.buildCacheKey("orgA", "lit", "gpt-4o", "prompt")
	k2 := e.buildCacheKey("orgB", "lit", "gpt-4o", "prompt")
	if k1 == k2 {
		t.Error("Different orgs must produce different cache keys")
	}
}

func TestBuildCacheKey_IsolateDifferentModels(t *testing.T) {
	e := newTestEngine()
	k1 := e.buildCacheKey("org1", "lit", "gpt-4o", "prompt")
	k2 := e.buildCacheKey("org1", "lit", "llama3.2", "prompt")
	if k1 == k2 {
		t.Error("Different models must produce different cache keys — critical for model-scoped caching")
	}
}

func TestBuildCacheKey_IsolateDifferentPrompts(t *testing.T) {
	e := newTestEngine()
	k1 := e.buildCacheKey("org1", "lit", "gpt-4o", "prompt A")
	k2 := e.buildCacheKey("org1", "lit", "gpt-4o", "prompt B")
	if k1 == k2 {
		t.Error("Different prompts must produce different cache keys")
	}
}

func TestBuildCacheKey_EmptyComponents(t *testing.T) {
	e := newTestEngine()
	// Must not panic on empty strings — format: org:<orgID>:m:<model>:<keyType>:<value>
	got := e.buildCacheKey("", "", "", "")
	want := "org::m:::"
	if got != want {
		t.Errorf("buildCacheKey empty = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// PromptRequest struct — JSON field validation
// ---------------------------------------------------------------------------

func TestPromptRequest_Defaults(t *testing.T) {
	req := PromptRequest{Prompt: "hello"}
	if req.SkipCache {
		t.Error("SkipCache should default to false")
	}
	if req.Stream {
		t.Error("Stream should default to false")
	}
	if req.Provider != "" {
		t.Error("Provider should default to empty string")
	}
}

// ---------------------------------------------------------------------------
// Atomic counters — thread safety smoke tests
// ---------------------------------------------------------------------------

func TestEngineCounters_AtomicIncrement(t *testing.T) {
	e := newTestEngine()
	e.TotalRequests.Add(1)
	e.CacheHits.Add(1)

	if e.TotalRequests.Load() != 1 {
		t.Errorf("TotalRequests = %d, want 1", e.TotalRequests.Load())
	}
	if e.CacheHits.Load() != 1 {
		t.Errorf("CacheHits = %d, want 1", e.CacheHits.Load())
	}
}

func TestEngineCounters_CacheHitRatio(t *testing.T) {
	e := newTestEngine()
	e.TotalRequests.Add(100)
	e.CacheHits.Add(87)

	total := float64(e.TotalRequests.Load())
	hits := float64(e.CacheHits.Load())
	ratio := hits / total * 100

	if ratio < 87.0 || ratio > 87.1 {
		t.Errorf("Cache hit ratio = %.2f%%, want 87.0%%", ratio)
	}
}
