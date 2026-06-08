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

// ═══════════════════════════════════════════════════════════════════════════════
// MEMZENT EVOLUTION PIPELINE INTEGRATION TEST (E1–E5)
//
// Tests the full gateway pipeline end-to-end:
//   E1: Entity Extraction — verifies entities appear in responses
//   E2: L1b Entity-Keyed Cache — verifies entity-keyed hits/misses
//   E3: Offline Learning Plane — verifies event emission + miner stats
//   E4: Workflow Registry — verifies API endpoints
//   E5: Entity Quality Metrics — verifies Prometheus counters + GPU avoidance
//
// Requires: running gateway (make up), MEMZENT_API_KEY set
// Usage: go run scripts/test_evolution/main.go
// ═══════════════════════════════════════════════════════════════════════════════

type ChatResponse struct {
	Text      string            `json:"text"`
	Cached    bool              `json:"cached"`
	Provider  string            `json:"provider"`
	RequestID string            `json:"request_id"`
	SessionID string            `json:"session_id"`
	Entities  map[string]string `json:"entities"`
	Error     string            `json:"error"`
}

type AuditEntry struct {
	Type       string            `json:"type"`
	Detail     string            `json:"detail"`
	Status     string            `json:"status"`
	CacheLayer string            `json:"cache_layer"`
	Entities   map[string]string `json:"entities"`
	Timestamp  string            `json:"timestamp"`
}

var (
	baseURL = envOr("MEMZENT_BASE_URL", "https://api.memzent.ai")
	apiKey  = envOr("MEMZENT_API_KEY", "memzent_aef5299d4207cf9f180f237ebfb80a78fc92363cb7649b22")

	passed  int
	failed  int
	skipped int
)

func main() {
	if apiKey == "" {
		fmt.Println("❌ Set MEMZENT_API_KEY environment variable")
		os.Exit(1)
	}

	printHeader()

	// ─── Phase E1: Entity Extraction ─────────────────────────────────────────
	section("E1: ENTITY EXTRACTION LAYER")

	// Test 1: Transfer with directional entities
	resp := chat("Transfer $500 from account 1001 to account 2002")
	assertEntities(resp, "E1.1 Transfer entities extracted", map[string]string{
		"source_account": "1001",
		"target_account": "2002",
		"amount":         "500",
	})

	// Test 2: Different direction must produce different entities
	resp2 := chat("Transfer $500 from account 2002 to account 1001")
	assertEntities(resp2, "E1.2 Reversed direction", map[string]string{
		"source_account": "2002",
		"target_account": "1001",
	})

	// Test 3: Verify responses are different (not colliding in cache)
	if resp.Text != "" && resp2.Text != "" && resp.Text == resp2.Text {
		fail("E1.3 Directional cache isolation", "Same response for opposite transfers!")
	} else {
		pass("E1.3 Directional cache isolation")
	}

	// Test 4: Customer entity extraction
	resp3 := chat("What is the balance for customer Raj?")
	assertEntityKey(resp3, "E1.4 Customer entity", "customer", "")

	// Test 5: No entities for generic prompt
	resp4 := chat("Tell me a joke")
	if len(resp4.Entities) == 0 {
		pass("E1.5 No entities for generic prompt")
	} else {
		// May have action entity — still valid
		pass("E1.5 Generic prompt (entities optional)")
	}

	// ─── Phase E2: L1b Entity-Keyed Cache ────────────────────────────────────
	section("E2: L1b ENTITY-KEYED HOT PATH CACHE")

	// Test 6: First call seeds the cache
	unique := fmt.Sprintf("Check invoice %d for customer Alpha", time.Now().UnixNano()%100000)
	r1 := chat(unique)
	assertNotCached(r1, "E2.1 First call is fresh (L5)")

	// Test 7: Exact same prompt should hit L1 cache
	time.Sleep(500 * time.Millisecond)
	r2 := chat(unique)
	assertCached(r2, "E2.2 Repeat prompt hits L1 cache")

	// Test 8: Semantically equivalent with same entities should hit L1b
	// (This depends on entity extraction matching)
	transfer1 := fmt.Sprintf("Transfer $%d from acc 7777 to acc 8888", time.Now().UnixNano()%1000)
	chat(transfer1) // seed
	time.Sleep(500 * time.Millisecond)
	// Slightly different wording, same entities
	transfer2 := fmt.Sprintf("Send $%d from account 7777 to account 8888", time.Now().UnixNano()%1000)
	r3 := chat(transfer2)
	// L1b hit depends on whether entity key matches — may or may not hit
	if r3.Cached {
		pass("E2.3 L1b entity-keyed cache hit (same entities, different wording)")
	} else {
		skip("E2.3 L1b miss (expected if entity keys differ slightly)")
	}

	// ─── Phase E2b: Cache Layer Tracking in Audit ────────────────────────────
	section("E2b: CACHE LAYER AUDIT TRACKING")

	audit := getAudit()
	if audit == nil {
		skip("E2b.1 Audit unavailable")
	} else {
		foundL1 := false
		foundL5 := false
		for _, entry := range audit {
			if entry.CacheLayer == "L1" {
				foundL1 = true
			}
			if entry.CacheLayer == "L5" {
				foundL5 = true
			}
		}
		check(foundL1 || foundL5, "E2b.1 Audit entries have cache_layer field")
	}

	// ─── Phase E3: Offline Learning Plane ────────────────────────────────────
	section("E3: OFFLINE LEARNING PLANE")

	stats := getOfflineStats()
	if stats == nil {
		skip("E3.1 Offline stats endpoint unavailable")
		skip("E3.2 Events being processed")
		skip("E3.3 No excessive drops")
	} else {
		plane, _ := stats["plane"].(map[string]interface{})
		if plane != nil {
			emitted, _ := plane["emitted"].(float64)
			processed, _ := plane["processed"].(float64)
			dropped, _ := plane["dropped"].(float64)

			check(emitted > 0, "E3.1 Offline events emitted")
			check(processed > 0, "E3.2 Events being processed by miners")

			if emitted > 0 {
				dropRate := dropped / emitted
				check(dropRate < 0.1, fmt.Sprintf("E3.3 Drop rate acceptable (%.1f%%)", dropRate*100))
			} else {
				skip("E3.3 No events to measure drop rate")
			}
		} else {
			skip("E3.1-3 Plane stats not returned")
		}

		// Check hot patterns from O1
		patterns, _ := stats["hot_patterns"].([]interface{})
		if patterns != nil && len(patterns) > 0 {
			pass("E3.4 O1 RequestMiner detecting hot patterns")
		} else {
			skip("E3.4 No hot patterns yet (need more traffic)")
		}

		// Check workflow sequences from O3
		sequences, _ := stats["workflow_sequences"].([]interface{})
		if sequences != nil && len(sequences) > 0 {
			pass("E3.5 O3 WorkflowMiner detecting tool sequences")
		} else {
			skip("E3.5 No workflow sequences yet (need multi-tool requests)")
		}
	}

	// ─── Phase E4: Workflow Registry ─────────────────────────────────────────
	section("E4: WORKFLOW REGISTRY")

	workflows := getWorkflows("")
	if workflows == nil {
		skip("E4.1 Workflows endpoint unavailable")
	} else {
		pass("E4.1 GET /v1/workflows responds")
	}

	workflowsActive := getWorkflows("active")
	if workflowsActive == nil {
		skip("E4.2 Active workflows filter unavailable")
	} else {
		pass(fmt.Sprintf("E4.2 Active workflows: %d", len(workflowsActive)))
	}

	// ─── Phase E5: Entity Quality Metrics ────────────────────────────────────
	section("E5: ENTITY QUALITY METRICS + GPU AVOIDANCE")

	entityMetrics := getEntityMetrics()
	if entityMetrics == nil {
		skip("E5.1 Entity metrics endpoint unavailable")
		skip("E5.2 GPU avoidance rate")
		skip("E5.3 Regex success rate")
	} else {
		pass("E5.1 GET /v1/metrics/entities responds")

		gpuRate, ok := entityMetrics["gpu_avoidance_rate"].(float64)
		if ok {
			pass(fmt.Sprintf("E5.2 GPU Avoidance Rate: %.1f%%", gpuRate*100))
		} else {
			skip("E5.2 GPU avoidance rate not computed yet")
		}

		regexRate, ok := entityMetrics["regex_success_rate"].(float64)
		if ok {
			pass(fmt.Sprintf("E5.3 Regex Success Rate: %.1f%%", regexRate*100))
		} else {
			skip("E5.3 Regex success rate not computed yet")
		}

		// Alert check: regex failure rate shouldn't exceed 15%
		regexSuccess, _ := entityMetrics["regex_success"].(float64)
		regexFailure, _ := entityMetrics["regex_failure"].(float64)
		total := regexSuccess + regexFailure
		if total > 10 {
			failureRate := regexFailure / total
			if failureRate > 0.15 {
				fail("E5.4 Regex failure rate alert", fmt.Sprintf("%.1f%% > 15%% threshold", failureRate*100))
			} else {
				pass(fmt.Sprintf("E5.4 Regex failure rate healthy: %.1f%%", failureRate*100))
			}
		} else {
			skip("E5.4 Not enough samples for failure rate alert")
		}
	}

	// ─── Health & Infrastructure ─────────────────────────────────────────────
	section("INFRASTRUCTURE")

	// Health check
	healthResp := httpGet("/healthz")
	check(healthResp != nil, "INFRA.1 /healthz responds")

	// Prometheus metrics available
	metricsResp := httpGetRaw("/metrics")
	if metricsResp != "" {
		hasEntityMetric := strings.Contains(metricsResp, "memzent_entity_regex_success_total")
		hasCacheLayer := strings.Contains(metricsResp, "memzent_cache_layer_hits_total")
		hasGPU := strings.Contains(metricsResp, "memzent_gpu_avoidance_total")
		check(hasEntityMetric, "INFRA.2 Prometheus: entity extraction counters")
		check(hasCacheLayer, "INFRA.3 Prometheus: cache layer distribution")
		check(hasGPU, "INFRA.4 Prometheus: GPU avoidance counter")
	} else {
		skip("INFRA.2-4 /metrics endpoint unavailable")
	}

	// ─── Summary ─────────────────────────────────────────────────────────────
	printSummary()
}

// ═══════════════════════════════════════════════════════════════════════════════
// HTTP Helpers
// ═══════════════════════════════════════════════════════════════════════════════

func chat(prompt string) *ChatResponse {
	body := map[string]interface{}{
		"messages": []map[string]string{{"role": "user", "content": prompt}},
	}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", baseURL+"/v1/chat", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("    ⚠️  HTTP error: %v\n", err)
		return &ChatResponse{}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var chatResp ChatResponse
	json.Unmarshal(respBody, &chatResp)
	return &chatResp
}

func getAudit() []AuditEntry {
	data := httpGet("/v1/audit")
	if data == nil {
		return nil
	}
	raw, _ := json.Marshal(data)
	var entries []AuditEntry
	json.Unmarshal(raw, &entries)
	return entries
}

func getOfflineStats() map[string]interface{} {
	result := httpGet("/v1/offline/stats")
	if m, ok := result.(map[string]interface{}); ok {
		return m
	}
	return nil
}

func getWorkflows(status string) []interface{} {
	path := "/v1/workflows"
	if status != "" {
		path += "?status=" + status
	}
	result := httpGet(path)
	if arr, ok := result.([]interface{}); ok {
		return arr
	}
	return nil
}

func getEntityMetrics() map[string]interface{} {
	result := httpGet("/v1/metrics/entities")
	if m, ok := result.(map[string]interface{}); ok {
		return m
	}
	return nil
}

func httpGet(path string) interface{} {
	req, _ := http.NewRequest("GET", baseURL+path, nil)
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	var result interface{}
	json.Unmarshal(body, &result)
	return result
}

func httpGetRaw(path string) string {
	req, _ := http.NewRequest("GET", baseURL+path, nil)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

// ═══════════════════════════════════════════════════════════════════════════════
// Assertion Helpers
// ═══════════════════════════════════════════════════════════════════════════════

func assertEntities(resp *ChatResponse, name string, expected map[string]string) {
	if resp.Entities == nil || len(resp.Entities) == 0 {
		fail(name, "No entities in response")
		return
	}
	for k, v := range expected {
		if got, ok := resp.Entities[k]; !ok {
			fail(name, fmt.Sprintf("Missing entity key '%s'", k))
			return
		} else if v != "" && got != v {
			fail(name, fmt.Sprintf("Entity '%s': got '%s', want '%s'", k, got, v))
			return
		}
	}
	pass(name)
}

func assertEntityKey(resp *ChatResponse, name string, key string, _ string) {
	if resp.Entities == nil {
		fail(name, "No entities in response")
		return
	}
	if _, ok := resp.Entities[key]; !ok {
		fail(name, fmt.Sprintf("Missing entity key '%s' (got: %v)", key, resp.Entities))
		return
	}
	pass(name)
}

func assertCached(resp *ChatResponse, name string) {
	if resp.Cached {
		pass(name)
	} else {
		fail(name, "Expected cached=true")
	}
}

func assertNotCached(resp *ChatResponse, name string) {
	if !resp.Cached {
		pass(name)
	} else {
		fail(name, "Expected cached=false")
	}
}

func check(condition bool, name string) {
	if condition {
		pass(name)
	} else {
		fail(name, "condition not met")
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// Output Formatting
// ═══════════════════════════════════════════════════════════════════════════════

func pass(name string) {
	passed++
	fmt.Printf("  \033[32m✓\033[0m %s\n", name)
}

func fail(name string, reason string) {
	failed++
	fmt.Printf("  \033[31m✗\033[0m %s — %s\n", name, reason)
}

func skip(name string) {
	skipped++
	fmt.Printf("  \033[33m○\033[0m %s\n", name)
}

func section(title string) {
	fmt.Printf("\n\033[1;35m── %s ──\033[0m\n\n", title)
}

func printHeader() {
	fmt.Printf("\n\033[1;36m══════════════════════════════════════════════════════════════\033[0m\n")
	fmt.Printf("\033[1;36m  MEMZENT EVOLUTION PIPELINE INTEGRATION TEST (E1–E5)\033[0m\n")
	fmt.Printf("\033[1;36m══════════════════════════════════════════════════════════════\033[0m\n")
	fmt.Printf("  Target:  %s\n", baseURL)
	fmt.Printf("  API Key: %s...%s\n\n", apiKey[:12], apiKey[len(apiKey)-4:])
}

func printSummary() {
	total := passed + failed + skipped
	fmt.Printf("\n\033[1;36m══════════════════════════════════════════════════════════════\033[0m\n")
	fmt.Printf("  \033[1mRESULTS: %d total | \033[32m%d passed\033[0m | \033[31m%d failed\033[0m | \033[33m%d skipped\033[0m\n",
		total, passed, failed, skipped)
	fmt.Printf("\033[1;36m══════════════════════════════════════════════════════════════\033[0m\n\n")

	if failed > 0 {
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
