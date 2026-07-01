// services/gateway/internal/invalidation/fingerprint.go
package invalidation

import (
	"crypto/sha256"
	"encoding/hex"
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
// effective preference/context fingerprint. When these differ between callers
// (or drift for one caller over time) a cached answer may no longer be
// appropriate, so the fingerprint is folded into the cache key to isolate them.
type PreferenceInputs struct {
	Role         string // RBAC role from auth context (e.g. "admin", "member")
	SystemPrompt string // system message / persona steering the response
}

// Fingerprint produces a deterministic, order-independent token string from the
// preference inputs. Provider and model are intentionally excluded because the
// cache key already partitions on them.
func Fingerprint(in PreferenceInputs) string {
	set := map[string]struct{}{}
	if r := strings.TrimSpace(strings.ToLower(in.Role)); r != "" {
		set["role:"+r] = struct{}{}
	}

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

// PreferenceTag returns a short, stable hash of the preference fingerprint,
// suitable for embedding as a cache-key segment. Returns "" when there are no
// preference signals, so default requests keep un-partitioned keys and maximum
// cache sharing.
func PreferenceTag(role, systemPrompt string) string {
	fp := Fingerprint(PreferenceInputs{Role: role, SystemPrompt: systemPrompt})
	if fp == "" {
		return ""
	}
	h := sha256.Sum256([]byte(fp))
	return hex.EncodeToString(h[:6]) // 12 hex chars — ample to avoid collisions
}
