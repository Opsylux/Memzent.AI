// internal/mcp/compressor.go

package mcp

import (
	"strings"
	"regexp"
)

// CompressToolOutput takes raw data and returns a "LLM-Friendly" version.
func CompressToolOutput(rawOutput string, intent string) string {
	// 1. Rule-based Pruning (e.g., remove boilerplate headers)
	lines := strings.Split(rawOutput, "\n")
	var filtered []string

	// Example: If the user is looking for "errors", only keep lines with keywords
	isErrorSearch := strings.Contains(strings.ToLower(intent), "error") || 
                     strings.Contains(strings.ToLower(intent), "fail")

	for _, line := range lines {
		if isErrorSearch {
			if matched, _ := regexp.MatchString("(?i)error|fail|critical|exception", line); matched {
				filtered = append(filtered, line)
			}
		} else {
			// Default: Just take the first 50 lines to prevent context overflow
			if len(filtered) < 50 {
				filtered = append(filtered, line)
			}
		}
	}

	// 2. Add a footer so the LLM knows the data was truncated
	if len(lines) > len(filtered) {
		filtered = append(filtered, "... [Memzent Truncated 90% of data for token efficiency]")
	}

	return strings.Join(filtered, "\n")
}