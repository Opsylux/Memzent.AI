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

// Tests that verify entity-aware semantic cache correctly distinguishes
// positionally-different prompts that the old sort-based guard conflated.
type TestCase struct {
	Name           string
	Prompt         string
	ExpectCached   *bool
	MustNotContain string
	SkipCache      bool
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
	fmt.Printf("\033[1;36m  MEMZENT ENTITY EXTRACTION CACHE GUARD TEST SUITE\033[0m\n")
	fmt.Printf("\033[1;36m══════════════════════════════════════════════════════\033[0m\n")
	fmt.Printf("  Target: %s\n", baseURL)
	fmt.Printf("  Purpose: Verify positional entity awareness in cache guard\n\n")

	// Flush cache for clean state
	fmt.Printf("\033[1;33m[SETUP]\033[0m Flushing org cache...\n")
	if err := flushCache(); err != nil {
		fmt.Printf("  ⚠️  Cache flush failed: %v\n", err)
	} else {
		fmt.Printf("  ✓ Cache flushed\n")
	}
	fmt.Println()

	falseVal := false
	trueVal := true

	tests := []TestCase{
		// ──── Group 1: Transfer Direction (THE core bug) ────
		// "Transfer $100 from 123 to 456" vs "Transfer $100 from 456 to 123"
		// Old guard: both produce [100, 123, 456] sorted → wrongly matches
		// New guard: source_account=123,target_account=456 ≠ source_account=456,target_account=123
		{
			Name:      "1a. Transfer from 123 to 456 (seed, skip cache)",
			Prompt:    "Transfer $100 from account 123 to account 456",
			SkipCache: true,
		},
		{
			Name:         "1b. REVERSED: Transfer from 456 to 123 → MUST NOT be cached",
			Prompt:       "Transfer $100 from account 456 to account 123",
			ExpectCached: &falseVal,
		},
		{
			Name:         "1c. Same direction again → SHOULD be cached",
			Prompt:       "Transfer $100 from account 123 to account 456",
			ExpectCached: &trueVal,
		},

		// ──── Group 2: Different Amounts, Same Accounts ────
		{
			Name:      "2a. Transfer $200 from 123 to 456 (seed)",
			Prompt:    "Transfer $200 from account 123 to account 456",
			SkipCache: true,
		},
		{
			Name:         "2b. Transfer $500 same accounts → MUST NOT be cached (amount differs)",
			Prompt:       "Transfer $500 from account 123 to account 456",
			ExpectCached: &falseVal,
		},

		// ──── Group 3: Invoice/Customer Swap ────
		// "Invoice #789 for customer #101" vs "Invoice #101 for customer #789"
		{
			Name:      "3a. Invoice 789 for customer 101 (seed)",
			Prompt:    "Show invoice #789 for customer #101",
			SkipCache: true,
		},
		{
			Name:         "3b. Swapped IDs → MUST NOT be cached",
			Prompt:       "Show invoice #101 for customer #789",
			ExpectCached: &falseVal,
		},

		// ──── Group 4: Customer Name Extraction ────
		{
			Name:      "4a. Balance for customer Raj (seed)",
			Prompt:    "What is the balance for customer Raj?",
			SkipCache: true,
		},
		{
			Name:           "4b. Balance for customer Priya → MUST NOT return Raj's data",
			Prompt:         "What is the balance for customer Priya?",
			ExpectCached:   &falseVal,
			MustNotContain: "Raj",
		},
		{
			Name:         "4c. Same customer Raj again → SHOULD be cached",
			Prompt:       "What is the balance for customer Raj?",
			ExpectCached: &trueVal,
		},

		// ──── Group 5: Semantic equivalence WITH same entities → cache hit ────
		{
			Name:      "5a. How much does customer Raj owe? (seed)",
			Prompt:    "How much does customer Raj owe?",
			SkipCache: true,
		},
		{
			Name:         "5b. Outstanding amount for Raj → semantically same + same entity → SHOULD cache",
			Prompt:       "What is the outstanding amount for customer Raj?",
			ExpectCached: &trueVal,
		},

		// ──── Group 6: Action Differentiation ────
		{
			Name:      "6a. Create invoice 500 for customer 101 (seed)",
			Prompt:    "Create invoice for $500 for customer #101",
			SkipCache: true,
		},
		{
			Name:         "6b. Delete invoice for customer 101 → different action → MUST NOT cache",
			Prompt:       "Delete invoice for customer #101",
			ExpectCached: &falseVal,
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

		ok := true
		var reason string

		if tc.ExpectCached != nil && resp.Cached != *tc.ExpectCached {
			ok = false
			reason = fmt.Sprintf("expected cached=%v, got cached=%v", *tc.ExpectCached, resp.Cached)
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
	}

	fmt.Println()
	if failed > 0 {
		os.Exit(1)
	}
}

func sendChat(prompt string, skipCache bool) (*ChatResponse, time.Duration, error) {
	body := map[string]any{
		"messages": []map[string]string{{"role": "user", "content": prompt}},
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", baseURL+"/v1/chat", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	if skipCache {
		req.Header.Set("X-Skip-Cache", "true")
	}

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	duration := time.Since(start)
	if err != nil {
		return nil, duration, fmt.Errorf("HTTP error: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, duration, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(data, &chatResp); err != nil {
		return nil, duration, fmt.Errorf("JSON parse error: %w", err)
	}
	return &chatResp, duration, nil
}

func flushCache() error {
	req, _ := http.NewRequest("POST", baseURL+"/v1/cache/flush", nil)
	req.Header.Set("X-API-Key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
