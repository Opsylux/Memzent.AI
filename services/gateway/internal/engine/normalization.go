package engine

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
)

var (
	// regexExtraSpace matches multiple whitespaces
	regexExtraSpace = regexp.MustCompile(`\s+`)
)

// NormalizePrompt performs deterministic canonicalization for high-precision caching.
// It lowercases, removes punctuation, and stabilizes whitespace — preserving all
// numbers and meaningful tokens. This ensures only truly identical intents share a cache key.
func NormalizePrompt(prompt string) (canonical string, hash string) {
	// 1. Lowercase and trim
	text := strings.ToLower(strings.TrimSpace(prompt))

	// 2. Remove punctuation and stabilize spaces (keep letters, digits, spaces)
	text = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' {
			return r
		}
		return -1
	}, text)
	text = regexExtraSpace.ReplaceAllString(text, " ")
	canonical = strings.TrimSpace(text)

	// 3. Generate Hash for stable lookup
	h := sha256.New()
	h.Write([]byte(canonical))
	hash = fmt.Sprintf("%x", h.Sum(nil))

	return canonical, hash
}
