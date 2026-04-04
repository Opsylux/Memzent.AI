package engine

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
)

var (
	// regexDigits matches numeric IDs (3 or more digits) or common patterns like write011
	regexDigits = regexp.MustCompile(`\d{2,}`)
	// regexExtraSpace matches multiple whitespaces
	regexExtraSpace = regexp.MustCompile(`\s+`)
)

// NormalizePrompt performs deterministic canonicalization for high-precision caching.
// It masks numbers, removes noise, and standardizes the intent.
func NormalizePrompt(prompt string) (canonical string, hash string) {
	// 1. Lowercase and trim
	text := strings.ToLower(strings.TrimSpace(prompt))

	// 2. ID Masking: Replace numeric sequences (2+ digits) with a generic <ID> token.
	// This ensures "write011" and "write111" become identical intents.
	text = regexDigits.ReplaceAllString(text, "<ID>")

	// 3. Remove punctuation and stabilize spaces
	text = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' || r == '<' || r == '>' {
			return r
		}
		return -1
	}, text)
	text = regexExtraSpace.ReplaceAllString(text, " ")
	canonical = strings.TrimSpace(text)

	// 4. Generate Hash for stable lookup
	h := sha256.New()
	h.Write([]byte(canonical))
	hash = fmt.Sprintf("%x", h.Sum(nil))

	return canonical, hash
}
