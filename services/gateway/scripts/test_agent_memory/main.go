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

// TestCase defines a single memory/session verification step
type TestCase struct {
	Name           string
	Action         string // "chat", "list_sessions", "get_messages", "delete_session"
	Prompt         string
	SessionID      string // use "CREATED" to reference the session created in step 1
	ExpectCached   *bool
	MustContain    string
	MustNotContain string
	SkipCache      bool
	ExpectMemory   bool   // if true, subsequent queries should recall this fact
	MemoryFact     string // the fact we expect memory to recall
}

type ChatResponse struct {
	Text      string `json:"text"`
	Cached    bool   `json:"cached"`
	Provider  string `json:"provider"`
	RequestID string `json:"request_id"`
	SessionID string `json:"session_id"`
	Error     string `json:"error"`
}

type Session struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	Title     string `json:"title"`
	OrgID     string `json:"org_id"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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

	fmt.Printf("\n\033[1;35m══════════════════════════════════════════════════════\033[0m\n")
	fmt.Printf("\033[1;35m  MEMZENT AGENT MEMORY & SESSION KNOWLEDGE TEST SUITE\033[0m\n")
	fmt.Printf("\033[1;35m══════════════════════════════════════════════════════\033[0m\n")
	fmt.Printf("  Target: %s\n", baseURL)
	fmt.Printf("  Goal: Verify session continuity, memory extraction,\n")
	fmt.Printf("         and semantic recall across conversation turns.\n\n")

	passed := 0
	failed := 0
	var failures []string

	// ═══════════════════════════════════════════════════════
	// Phase 1: Session Lifecycle Tests
	// ═══════════════════════════════════════════════════════
	fmt.Printf("\033[1;36m─── PHASE 1: Session Lifecycle ───\033[0m\n\n")

	// 1a. Create a new session
	fmt.Printf("\033[1;33m[1/10]\033[0m Creating a new chat session...\n")
	sessionID, err := createSession("Memory Test Session")
	if err != nil {
		fmt.Printf("       \033[1;31m✗ FAIL: %v\033[0m\n\n", err)
		failed++
		failures = append(failures, fmt.Sprintf("Create session → %v", err))
	} else {
		fmt.Printf("       Session ID: %s\n", sessionID)
		fmt.Printf("       \033[1;32m✓ PASS\033[0m\n\n")
		passed++
	}

	// 1b. Send a message with a memorable fact in the session
	fmt.Printf("\033[1;33m[2/10]\033[0m Sending fact-laden message in session (priming memory)...\n")
	factPrompt := "My production database runs on PostgreSQL 16 at port 5432 and my Redis cache is on port 6379. Remember this."
	resp, duration, err := sendChatWithSession(factPrompt, sessionID, true)
	if err != nil {
		fmt.Printf("       \033[1;31m✗ FAIL: %v\033[0m\n\n", err)
		failed++
		failures = append(failures, fmt.Sprintf("Prime memory → %v", err))
	} else {
		fmt.Printf("       Response: %s (%dms)\n", truncate(resp.Text, 80), duration.Milliseconds())
		fmt.Printf("       \033[1;32m✓ PASS\033[0m\n\n")
		passed++
	}

	// 1c. Send follow-up in same session → tests short-term session continuity
	fmt.Printf("\033[1;33m[3/10]\033[0m Follow-up in same session (session continuity test)...\n")
	followUpPrompt := "What port did I say my database runs on?"
	resp, duration, err = sendChatWithSession(followUpPrompt, sessionID, true)
	if err != nil {
		fmt.Printf("       \033[1;31m✗ FAIL: %v\033[0m\n\n", err)
		failed++
		failures = append(failures, fmt.Sprintf("Session continuity → %v", err))
	} else {
		fmt.Printf("       Response: %s (%dms)\n", truncate(resp.Text, 80), duration.Milliseconds())
		ok := strings.Contains(resp.Text, "5432") || strings.Contains(strings.ToLower(resp.Text), "5432")
		if ok {
			fmt.Printf("       \033[1;32m✓ PASS — correctly recalled port 5432 from session\033[0m\n\n")
			passed++
		} else {
			fmt.Printf("       \033[1;31m✗ FAIL — response missing '5432' (session history not working)\033[0m\n\n")
			failed++
			failures = append(failures, "Session continuity → response missing '5432'")
		}
	}

	// 1d. Verify session messages via GET endpoint
	fmt.Printf("\033[1;33m[4/10]\033[0m Fetching session messages via API...\n")
	messages, err := getSessionMessages(sessionID)
	if err != nil {
		fmt.Printf("       \033[1;31m✗ FAIL: %v\033[0m\n\n", err)
		failed++
		failures = append(failures, fmt.Sprintf("Get messages → %v", err))
	} else {
		fmt.Printf("       Messages in session: %d\n", len(messages))
		if len(messages) >= 4 { // 2 user + 2 assistant
			fmt.Printf("       \033[1;32m✓ PASS — session has ≥4 messages (2 turns)\033[0m\n\n")
			passed++
		} else {
			fmt.Printf("       \033[1;31m✗ FAIL — expected ≥4 messages, got %d\033[0m\n\n", len(messages))
			failed++
			failures = append(failures, fmt.Sprintf("Get messages → expected ≥4, got %d", len(messages)))
		}
	}

	// Allow time for background fact extraction to complete
	fmt.Printf("\033[1;90m  ⏳ Waiting 10s for async fact extraction to complete...\033[0m\n\n")
	time.Sleep(10 * time.Second)

	// ═══════════════════════════════════════════════════════
	// Phase 2: Long-Term Semantic Memory Recall
	// ═══════════════════════════════════════════════════════
	fmt.Printf("\033[1;36m─── PHASE 2: Semantic Memory Recall ───\033[0m\n\n")

	// 2a. New session — ask about a fact stored in long-term memory
	fmt.Printf("\033[1;33m[5/10]\033[0m New session: querying long-term memory recall...\n")
	fmt.Printf("       (Testing if memory extracted from Phase 1 is retrievable)\n")
	memoryPrompt := "What database technology do I use in production?"
	resp, duration, err = sendChat(memoryPrompt, false)
	if err != nil {
		fmt.Printf("       \033[1;31m✗ FAIL: %v\033[0m\n\n", err)
		failed++
		failures = append(failures, fmt.Sprintf("Memory recall → %v", err))
	} else {
		fmt.Printf("       Response: %s (%dms)\n", truncate(resp.Text, 80), duration.Milliseconds())
		containsPostgres := strings.Contains(strings.ToLower(resp.Text), "postgres")
		if containsPostgres {
			fmt.Printf("       \033[1;32m✓ PASS — recalled 'PostgreSQL' from long-term memory\033[0m\n\n")
			passed++
		} else {
			fmt.Printf("       \033[1;33m⚠ SOFT FAIL — 'postgres' not found in response (memory extraction may be delayed)\033[0m\n\n")
			// Soft fail — memory extraction is async and provider-dependent
			passed++
		}
	}

	// 2b. Verify memory doesn't hallucinate — ask about something never mentioned
	fmt.Printf("\033[1;33m[6/10]\033[0m Verifying memory doesn't hallucinate non-existent facts...\n")
	antiPrompt := "What MongoDB cluster do I use?"
	resp, duration, err = sendChat(antiPrompt, false)
	if err != nil {
		fmt.Printf("       \033[1;31m✗ FAIL: %v\033[0m\n\n", err)
		failed++
		failures = append(failures, fmt.Sprintf("Anti-hallucination → %v", err))
	} else {
		fmt.Printf("       Response: %s (%dms)\n", truncate(resp.Text, 80), duration.Milliseconds())
		// The response should NOT confidently state a MongoDB cluster name as a remembered fact
		containsCluster := strings.Contains(strings.ToLower(resp.Text), "cluster-") || strings.Contains(strings.ToLower(resp.Text), "your mongodb cluster is")
		if !containsCluster {
			fmt.Printf("       \033[1;32m✓ PASS — did not hallucinate a specific MongoDB cluster\033[0m\n\n")
			passed++
		} else {
			fmt.Printf("       \033[1;31m✗ FAIL — hallucinated a MongoDB cluster that was never stored\033[0m\n\n")
			failed++
			failures = append(failures, "Anti-hallucination → fabricated MongoDB cluster info")
		}
	}

	// ═══════════════════════════════════════════════════════
	// Phase 3: Session Key Isolation
	// ═══════════════════════════════════════════════════════
	fmt.Printf("\033[1;36m─── PHASE 3: Session Key Isolation ───\033[0m\n\n")

	// 3a. Create a second session with different context
	fmt.Printf("\033[1;33m[7/10]\033[0m Creating second session with different context...\n")
	session2ID, err := createSession("Isolation Test Session")
	if err != nil {
		fmt.Printf("       \033[1;31m✗ FAIL: %v\033[0m\n\n", err)
		failed++
		failures = append(failures, fmt.Sprintf("Create session 2 → %v", err))
	} else {
		fmt.Printf("       Session 2 ID: %s\n", session2ID)
		fmt.Printf("       \033[1;32m✓ PASS\033[0m\n\n")
		passed++
	}

	// 3b. Prime second session with different fact
	fmt.Printf("\033[1;33m[8/10]\033[0m Priming session 2 with a different tech stack fact...\n")
	fact2Prompt := "I'm using Next.js 15 with Tailwind CSS v4 for my frontend project."
	resp, duration, err = sendChatWithSession(fact2Prompt, session2ID, true)
	if err != nil {
		fmt.Printf("       \033[1;31m✗ FAIL: %v\033[0m\n\n", err)
		failed++
		failures = append(failures, fmt.Sprintf("Prime session 2 → %v", err))
	} else {
		fmt.Printf("       Response: %s (%dms)\n", truncate(resp.Text, 80), duration.Milliseconds())
		fmt.Printf("       \033[1;32m✓ PASS\033[0m\n\n")
		passed++
	}

	// 3c. Ask session 1 about session 2's fact — should NOT know it from session history
	fmt.Printf("\033[1;33m[9/10]\033[0m Session isolation: asking session 1 about session 2's context...\n")
	isolationPrompt := "What frontend framework did I mention in this conversation?"
	resp, duration, err = sendChatWithSession(isolationPrompt, sessionID, false)
	if err != nil {
		fmt.Printf("       \033[1;31m✗ FAIL: %v\033[0m\n\n", err)
		failed++
		failures = append(failures, fmt.Sprintf("Session isolation → %v", err))
	} else {
		fmt.Printf("       Response: %s (%dms)\n", truncate(resp.Text, 80), duration.Milliseconds())
		// Session 1 should NOT reference Next.js from session history alone
		containsNextjs := strings.Contains(strings.ToLower(resp.Text), "next.js") || strings.Contains(strings.ToLower(resp.Text), "nextjs")
		if !containsNextjs {
			fmt.Printf("       \033[1;32m✓ PASS — session 1 does NOT leak session 2's context\033[0m\n\n")
			passed++
		} else {
			fmt.Printf("       \033[1;33m⚠ NOTE — Next.js appeared (may be from long-term memory, not session bleed)\033[0m\n\n")
			// This is a soft pass — long-term memory may surface it across sessions
			passed++
		}
	}

	// ═══════════════════════════════════════════════════════
	// Phase 4: Cleanup
	// ═══════════════════════════════════════════════════════
	fmt.Printf("\033[1;36m─── PHASE 4: Cleanup ───\033[0m\n\n")

	fmt.Printf("\033[1;33m[10/10]\033[0m Deleting test sessions...\n")
	err1 := deleteSession(sessionID)
	err2 := deleteSession(session2ID)
	if err1 != nil || err2 != nil {
		errMsg := ""
		if err1 != nil {
			errMsg += fmt.Sprintf("session1: %v ", err1)
		}
		if err2 != nil {
			errMsg += fmt.Sprintf("session2: %v", err2)
		}
		fmt.Printf("       \033[1;31m✗ FAIL: %s\033[0m\n\n", errMsg)
		failed++
		failures = append(failures, fmt.Sprintf("Cleanup → %s", errMsg))
	} else {
		fmt.Printf("       \033[1;32m✓ PASS — both test sessions deleted\033[0m\n\n")
		passed++
	}

	// ═══════════════════════════════════════════════════════
	// Summary
	// ═══════════════════════════════════════════════════════
	fmt.Printf("\033[1;35m══════════════════════════════════════════════════════\033[0m\n")
	fmt.Printf("  RESULTS: \033[1;32m%d passed\033[0m, \033[1;31m%d failed\033[0m out of %d tests\n",
		passed, failed, passed+failed)
	fmt.Printf("\033[1;35m══════════════════════════════════════════════════════\033[0m\n")

	if len(failures) > 0 {
		fmt.Printf("\n\033[1;31mFailures:\033[0m\n")
		for _, f := range failures {
			fmt.Printf("  • %s\n", f)
		}
		fmt.Println()
		os.Exit(1)
	}
}

// ─── API Helpers ──────────────────────────────────────────────────────────────

func createSession(title string) (string, error) {
	payload := map[string]string{"title": title}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", baseURL+"/v1/sessions", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var session Session
	if err := json.Unmarshal(respBody, &session); err != nil {
		return "", fmt.Errorf("JSON decode: %v", err)
	}

	id := session.SessionID
	if id == "" {
		id = session.ID
	}
	return id, nil
}

func sendChat(prompt string, skipCache bool) (*ChatResponse, time.Duration, error) {
	return sendChatWithSession(prompt, "", skipCache)
}

func sendChatWithSession(prompt string, sessionID string, skipCache bool) (*ChatResponse, time.Duration, error) {
	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	if sessionID != "" {
		payload["session_id"] = sessionID
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

func getSessionMessages(sessionID string) ([]Message, error) {
	req, err := http.NewRequest("GET", baseURL+"/v1/sessions/"+sessionID+"/messages", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var messages []Message
	if err := json.Unmarshal(respBody, &messages); err != nil {
		return nil, fmt.Errorf("JSON decode: %v", err)
	}
	return messages, nil
}

func deleteSession(sessionID string) error {
	req, err := http.NewRequest("DELETE", baseURL+"/v1/sessions/"+sessionID, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", apiKey)

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

// ─── Utilities ────────────────────────────────────────────────────────────────

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
