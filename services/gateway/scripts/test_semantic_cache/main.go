package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Test cases: pairs of prompts that are semantically similar but should NOT
// return the same cached answer due to different numeric parameters.
type TestCase struct {
	Name           string
	Prompt         string
	ExpectCached   *bool  // nil = don't care, true/false = assert
	MustContain    string // substring the response MUST contain
	MustNotContain string // substring the response MUST NOT contain
	SkipCache      bool   // send X-Skip-Cache: true
}

type ChatResponse struct {
	Text      string `json:"text"`
	Cached    bool   `json:"cached"`
	Provider  string `json:"provider"`
	RequestID string `json:"request_id"`
	Error     string `json:"error"`
}

var (
	baseURL = envOr("MEMZENT_BASE_URL", "https://api.memzent.ai")
	apiKey  = envOr("MEMZENT_API_KEY", "memzent_aef5299d4207cf9f180f237ebfb80a78fc92363cb7649b22")
)

func main() {
	if apiKey == "" {
		fmt.Println("❌ Set MEMZENT_API_KEY environment variable")
		os.Exit(1)
	}

	fmt.Printf("\n\033[1;36m══════════════════════════════════════════════════════\033[0m\n")
	fmt.Printf("\033[1;36m  MEMZENT SEMANTIC CACHE CORRECTNESS TEST SUITE\033[0m\n")
	fmt.Printf("\033[1;36m══════════════════════════════════════════════════════\033[0m\n")
	fmt.Printf("  Target: %s\n\n", baseURL)

	// Flush cache before tests to remove stale entries from prior runs.
	// This is required for idempotent test runs: without it, prompts cached
	// in a previous run would return cached=true for cases that expect fresh hits.
	fmt.Printf("\033[1;33m[SETUP]\033[0m Flushing org cache to start clean...\n")
	if err := flushCache(); err != nil {
		fmt.Printf("  ⚠️  Cache flush failed (may not be deployed yet): %v\n", err)
		fmt.Printf("  Tests will proceed — stale entries may cause false failures.\n")
	} else {
		fmt.Printf("  ✓ Cache flushed successfully\n")
	}
	fmt.Println()

	falseVal := false
	trueVal := true

	tests := []TestCase{
		// ──── Group 1: Numeric Parameter Variation ────
		// These tests verify that prompts with different numeric parameters
		// are NOT served from cache. We avoid asserting on specific LLM answers
		// since local models (Ollama) can be unreliable at math.
		{
			Name:      "1a. Base formula (skip cache to get fresh)",
			Prompt:    "what is (a+b)^2 where a=3, b=4",
			SkipCache: true,
		},
		{
			Name:         "1b. Same formula, different numbers → must NOT be cached",
			Prompt:       "what is (a+b)^2 where a=3, b=7",
			ExpectCached: &falseVal,
		},
		{
			Name:         "1c. Same formula, same numbers → SHOULD be cached",
			Prompt:       "what is (a+b)^2 where a=3, b=4",
			ExpectCached: &trueVal,
		},

		// ──── Group 2: Swapped Parameter Order ────
		{
			Name:      "2a. Baseline: a=10, b=5 (skip cache)",
			Prompt:    "calculate (a+b)^2 when a=10 and b=5",
			SkipCache: true,
		},
		{
			Name:         "2b. Swapped: a=5, b=10 → same result but different prompt",
			Prompt:       "calculate (a+b)^2 when a=5 and b=10",
			ExpectCached: nil, // acceptable either way since result is same (225)
		},
		{
			Name:           "2c. Different numbers: a=5, b=15 → MUST NOT return cached a=5,b=10 answer",
			Prompt:         "calculate (a+b)^2 when a=5 and b=15",
			ExpectCached:   &falseVal,
			MustNotContain: "b = 10",
		},

		// ──── Group 3: Similar Natural Language, Different Intent ────
		{
			Name:      "3a. Population of Paris (skip cache)",
			Prompt:    "What is the population of Paris?",
			SkipCache: true,
		},
		{
			Name:           "3b. Population of London → must NOT return Paris data",
			Prompt:         "What is the population of London?",
			MustNotContain: "Paris",
		},

		// ──── Group 4: X-Skip-Cache header validation ────
		{
			Name:         "4a. Normal request (may cache)",
			Prompt:       "What is 2+2?",
			ExpectCached: nil,
			MustContain:  "4",
		},
		{
			Name:         "4b. Same prompt with skip-cache → must NOT be cached",
			Prompt:       "What is 2+2?",
			SkipCache:    true,
			ExpectCached: &falseVal,
			MustContain:  "4",
		},

		// ──── Group 5: Edge case - numbers embedded in words ────
		{
			Name:      "5a. Fibonacci 10th term (skip cache)",
			Prompt:    "What is the 10th fibonacci number?",
			SkipCache: true,
		},
		{
			Name:           "5b. Fibonacci 15th term → must NOT return 10th's answer",
			Prompt:         "What is the 15th fibonacci number?",
			ExpectCached:   &falseVal,
			MustNotContain: "10th",
		},
	}

	passed := 0
	failed := 0
	var failures []string

	for i, tc := range tests {
		fmt.Printf("\033[1;33m[%d/%d]\033[0m %s\n", i+1, len(tests), tc.Name)
		fmt.Printf("       Prompt: %q\n", tc.Prompt)

		resp, duration, err := sendChat(tc.Prompt, tc.SkipCache)
		if err != nil {
			fmt.Printf("       \033[1;31m✗ ERROR: %v\033[0m\n\n", err)
			failed++
			failures = append(failures, fmt.Sprintf("%s → %v", tc.Name, err))
			continue
		}

		fmt.Printf("       Response: %s (cached=%v, %dms)\n",
			truncate(resp.Text, 80), resp.Cached, duration.Milliseconds())

		// Validate
		ok := true
		var reason string

		if tc.ExpectCached != nil && resp.Cached != *tc.ExpectCached {
			ok = false
			reason = fmt.Sprintf("expected cached=%v, got cached=%v", *tc.ExpectCached, resp.Cached)
		}
		if tc.MustContain != "" && !strings.Contains(resp.Text, tc.MustContain) {
			ok = false
			reason = fmt.Sprintf("response missing expected substring %q", tc.MustContain)
		}
		if tc.MustNotContain != "" && strings.Contains(resp.Text, tc.MustNotContain) {
			ok = false
			reason = fmt.Sprintf("response contains forbidden substring %q", tc.MustNotContain)
		}

		if ok {
			fmt.Printf("       \033[1;32m✓ PASS\033[0m\n\n")
			passed++
		} else {
			fmt.Printf("       \033[1;31m✗ FAIL: %s\033[0m\n\n", reason)
			failed++
			failures = append(failures, fmt.Sprintf("%s → %s", tc.Name, reason))
		}

		// Small delay between requests to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	// Summary
	fmt.Printf("\033[1;36m══════════════════════════════════════════════════════\033[0m\n")
	fmt.Printf("  RESULTS: \033[1;32m%d passed\033[0m, \033[1;31m%d failed\033[0m out of %d tests\n",
		passed, failed, len(tests))
	fmt.Printf("\033[1;36m══════════════════════════════════════════════════════\033[0m\n")

	if len(failures) > 0 {
		fmt.Printf("\n\033[1;31mFailures:\033[0m\n")
		for _, f := range failures {
			fmt.Printf("  • %s\n", f)
		}
		fmt.Println()
		os.Exit(1)
	}
}

func sendChat(prompt string, skipCache bool) (*ChatResponse, time.Duration, error) {
	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", baseURL+"/v1/chat", bytes.NewBuffer(body))
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	if skipCache {
		req.Header.Set("X-Skip-Cache", "true")
	}

	client := &http.Client{Timeout: 60 * time.Second}
	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)
	if err != nil {
		return nil, duration, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, duration, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, duration, fmt.Errorf("JSON decode: %v (body: %s)", err, string(respBody))
	}
	return &chatResp, duration, nil
}

func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func flushCache() error {
	req, err := http.NewRequest("POST", baseURL+"/v1/cache/flush", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
