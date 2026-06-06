package engine

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
)

var (
	// regexLongDigits matches long numeric sequences (4+ digits) that are likely IDs,
	// not mathematical parameters. Short numbers (1-3 digits) are preserved because
	// they often represent meaningful values (a=10, 15th, etc.) that affect the answer.
	regexLongDigits = regexp.MustCompile(`\d{4,}`)
	// regexHashID matches # followed by digits (e.g., ticket #45, issue #123)
	regexHashID = regexp.MustCompile(`#\d+`)
	// regexExtraSpace matches multiple whitespaces
	regexExtraSpace = regexp.MustCompile(`\s+`)
)

// NormalizePrompt performs deterministic canonicalization for high-precision caching.
// It masks numbers, removes noise, and standardizes the intent.
func NormalizePrompt(prompt string) (canonical string, hash string) {
	// 1. Lowercase and trim
	text := strings.ToLower(strings.TrimSpace(prompt))

	// 2. ID Masking: Replace hash-prefixed numbers (#45, #123) and long numeric
	// sequences (4+ digits) with <id>. Short numbers (1-3 digits) are preserved
	// because they typically represent mathematical values or ordinals that
	// change the semantics of the query (e.g., a=10, 15th fibonacci).
	text = regexHashID.ReplaceAllString(text, "<id>")
	text = regexLongDigits.ReplaceAllString(text, "<id>")

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
