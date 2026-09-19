package engine

import (
	"memzent-gateway/internal/llm"
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
// buildCacheKeyV — version-tagged cache keys (issue #11)
// ---------------------------------------------------------------------------

func TestBuildCacheKeyV_EmptyVersionMatchesLegacy(t *testing.T) {
	e := newTestEngine()
	legacy := e.buildCacheKey("org1", "p", "gpt-4o", "hello")
	versioned := e.buildCacheKeyV("org1", "", "", "p", "gpt-4o", "hello")
	if legacy != versioned {
		t.Errorf("empty version must reproduce legacy format: %q vs %q", legacy, versioned)
	}
}

func TestBuildCacheKeyV_VersionSegmentAndFlushable(t *testing.T) {
	e := newTestEngine()
	got := e.buildCacheKeyV("org1", "3", "", "p", "gpt-4o", "hello")
	want := "org:org1:ver:3:m:gpt-4o:p:hello"
	if got != want {
		t.Errorf("buildCacheKeyV = %q, want %q", got, want)
	}
	// Must remain under the org:<orgID>: prefix so FlushByPattern("org:org1:*") works.
	if len(got) < len("org:org1:") || got[:len("org:org1:")] != "org:org1:" {
		t.Errorf("versioned key %q lost org prefix", got)
	}
}

func TestBuildCacheKeyV_DifferentVersionsIsolate(t *testing.T) {
	e := newTestEngine()
	k1 := e.buildCacheKeyV("org1", "1", "", "p", "gpt-4o", "hello")
	k2 := e.buildCacheKeyV("org1", "2", "", "p", "gpt-4o", "hello")
	if k1 == k2 {
		t.Error("different cache versions must produce different keys (invalidation)")
	}
}

func TestBuildCacheKeyV_PreferenceSegmentIsolates(t *testing.T) {
	e := newTestEngine()
	// A preference tag inserts a pf: segment while preserving the org prefix.
	got := e.buildCacheKeyV("org1", "3", "abc123", "p", "gpt-4o", "hello")
	want := "org:org1:ver:3:pf:abc123:m:gpt-4o:p:hello"
	if got != want {
		t.Errorf("buildCacheKeyV with pref = %q, want %q", got, want)
	}
	// Different preference tags must not collide (per-preference isolation).
	other := e.buildCacheKeyV("org1", "3", "def456", "p", "gpt-4o", "hello")
	if got == other {
		t.Error("different preference tags must produce different keys")
	}
	// Empty pref with a version keeps the version-only format.
	noPref := e.buildCacheKeyV("org1", "3", "", "p", "gpt-4o", "hello")
	if noPref != "org:org1:ver:3:m:gpt-4o:p:hello" {
		t.Errorf("empty pref changed version-only format: %q", noPref)
	}
}

func TestBuildEntityCacheKeyV_Versioned(t *testing.T) {
	e := newTestEngine()
	ents := map[string]string{"action": "transfer"}
	legacy := e.buildEntityCacheKey("org1", "gpt-4", ents)
	if legacy != "org:org1:m:gpt-4:e:action=transfer" {
		t.Errorf("legacy entity key = %q", legacy)
	}
	versioned := e.buildEntityCacheKeyV("org1", "7", "", "gpt-4", ents)
	if versioned != "org:org1:ver:7:m:gpt-4:e:action=transfer" {
		t.Errorf("versioned entity key = %q", versioned)
	}
	withPref := e.buildEntityCacheKeyV("org1", "7", "abc123", "gpt-4", ents)
	if withPref != "org:org1:ver:7:pf:abc123:m:gpt-4:e:action=transfer" {
		t.Errorf("pref entity key = %q", withPref)
	}
}

// ---------------------------------------------------------------------------
// PromptRequest struct — JSON field validation
// ---------------------------------------------------------------------------

func TestPromptRequest_Defaults(t *testing.T) {
	req := PromptRequest{Messages: []llm.Message{{Role: "user", Content: "hello"}}}
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
