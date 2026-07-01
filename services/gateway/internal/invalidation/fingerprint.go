// services/gateway/internal/invalidation/fingerprint.go
package invalidation

import (
	"regexp"
	"sort"
	"strings"
)

// wordRe splits fingerprint source text into lowercase word tokens.
var wordRe = regexp.MustCompile(`[a-z0-9]+`)

// stopWords are ignored when tokenizing a system prompt so that trivial phrasing
// differences don't register as preference drift.
var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "you": true, "are": true, "is": true,
	"be": true, "to": true, "of": true, "and": true, "or": true, "as": true,
	"with": true, "for": true, "your": true, "this": true, "that": true,
}

// PreferenceInputs captures the request-level signals that make up a user's
// effective preference/context fingerprint. When any of these drift
// significantly between when a response was cached and when it is served, the
// cached answer may no longer be appropriate.
type PreferenceInputs struct {
	Role         string // RBAC role from auth context (e.g. "admin", "member")
	Provider     string // resolved LLM provider
	Model        string // resolved model
	SystemPrompt string // system message / persona steering the response
}

// Fingerprint produces a deterministic, order-independent token string from the
// preference inputs. Stored alongside a cache entry and compared on retrieval.
func Fingerprint(in PreferenceInputs) string {
	set := map[string]struct{}{}
	add := func(prefix, v string) {
		v = strings.TrimSpace(strings.ToLower(v))
		if v != "" {
			set[prefix+":"+v] = struct{}{}
		}
	}
	add("role", in.Role)
	add("provider", in.Provider)
	add("model", in.Model)

	for _, w := range wordRe.FindAllString(strings.ToLower(in.SystemPrompt), -1) {
		if len(w) < 2 || stopWords[w] {
			continue
		}
		set["sp:"+w] = struct{}{}
	}

	tokens := make([]string, 0, len(set))
	for t := range set {
		tokens = append(tokens, t)
	}
	sort.Strings(tokens)
	return strings.Join(tokens, " ")
}

// Similarity returns the Jaccard similarity (0..1) between two fingerprints.
// Two empty fingerprints are considered identical (1.0).
func Similarity(a, b string) float64 {
	sa := tokenSet(a)
	sb := tokenSet(b)
	if len(sa) == 0 && len(sb) == 0 {
		return 1.0
	}
	if len(sa) == 0 || len(sb) == 0 {
		return 0.0
	}
	inter := 0
	for t := range sa {
		if _, ok := sb[t]; ok {
			inter++
		}
	}
	union := len(sa) + len(sb) - inter
	if union == 0 {
		return 1.0
	}
	return float64(inter) / float64(union)
}

func tokenSet(s string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, t := range strings.Fields(s) {
		out[t] = struct{}{}
	}
	return out
}
