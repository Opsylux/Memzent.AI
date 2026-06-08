package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	concurrency = 20
	totalRuns   = 100
	rampUpDelay = 200 * time.Millisecond
)

type SessionResponse struct {
	ID string `json:"id"`
}

type Metrics struct {
	TotalRequests int64
	SuccessCount  int64
	FailureCount  int64
	TotalDuration int64 // in microseconds
}

type loadConfig struct {
	baseURL   string
	jwtSecret string
	orgID     string
	apiKey    string
}

func loadConfigFromEnv() (loadConfig, error) {
	cfg := loadConfig{
		baseURL:   envOr("GATEWAY_URL", "http://localhost:8080"),
		jwtSecret: os.Getenv("JWT_SECRET"),
		orgID:     os.Getenv("MEMZENT_ORG_ID"),
		apiKey:    os.Getenv("MEMZENT_API_KEY"),
	}
	if cfg.jwtSecret == "" {
		return cfg, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.orgID == "" {
		return cfg, fmt.Errorf("MEMZENT_ORG_ID is required")
	}
	if cfg.apiKey == "" {
		return cfg, fmt.Errorf("MEMZENT_API_KEY is required")
	}
	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	cfg, err := loadConfigFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "test_flow: %v\n", err)
		fmt.Fprintf(os.Stderr, "Required env: JWT_SECRET, MEMZENT_ORG_ID, MEMZENT_API_KEY\n")
		fmt.Fprintf(os.Stderr, "Optional env: GATEWAY_URL (default http://localhost:8080)\n")
		os.Exit(1)
	}

	fmt.Printf("\033[1;36m=== MEMZENT OS: HIGH-THROUGHPUT LOAD GENERATOR ===\033[0m\n")
	fmt.Printf("Config: Concurrency=%d, Target Total Cycles=%d, Gateway=%s\n\n", concurrency, totalRuns, cfg.baseURL)

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        120,
			MaxIdleConnsPerHost: 120,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	jwtTokenString := generateDynamicJwtToken(cfg)
	metrics := &Metrics{}

	taskQueue := make(chan int, totalRuns)
	for i := 0; i < totalRuns; i++ {
		taskQueue <- i
	}
	close(taskQueue)

	var wg sync.WaitGroup
	startTime := time.Now()

	for w := 1; w <= concurrency; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			sessionID, err := createSession(client, cfg, jwtTokenString)
			if err != nil {
				fmt.Printf("\033[1;31m[Worker %d] Failed to initialize session: %v\033[0m\n", workerID, err)
				return
			}

			for range taskQueue {
				runCycle(client, cfg, jwtTokenString, sessionID, metrics)
			}
		}(w)

		time.Sleep(rampUpDelay)
	}

	wg.Wait()
	totalElapsed := time.Since(startTime)

	printReport(metrics, totalElapsed)
}

func runCycle(client *http.Client, cfg loadConfig, jwtToken, sessionID string, m *Metrics) {
	prompt := "Verify performance connectivity metrics."

	reqPayload := map[string]interface{}{
		"session_id": sessionID,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	reqBody, _ := json.Marshal(reqPayload)

	req, _ := http.NewRequestWithContext(context.Background(), "POST", cfg.baseURL+"/v1/chat", bytes.NewBuffer(reqBody))

	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("X-API-Key", cfg.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Org-ID", cfg.orgID)

	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start).Microseconds()

	atomic.AddInt64(&m.TotalRequests, 1)
	atomic.AddInt64(&m.TotalDuration, duration)

	if err != nil {
		atomic.AddInt64(&m.FailureCount, 1)
		fmt.Printf("\033[1;31m[Network Error]: %v\033[0m\n", err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		atomic.AddInt64(&m.SuccessCount, 1)
	} else {
		atomic.AddInt64(&m.FailureCount, 1)
		bodyStr := string(bodyBytes)
		if len(bodyStr) > 200 {
			bodyStr = bodyStr[:200]
		}
		failCount := atomic.LoadInt64(&m.FailureCount)
		if failCount <= 5 {
			fmt.Printf("\033[1;33m[HTTP Reject]: Status %d | Body: %s\033[0m\n", resp.StatusCode, bodyStr)
		}
	}
}

func generateDynamicJwtToken(cfg loadConfig) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  "test-user-01",
		"role": "admin",
		"app_metadata": map[string]interface{}{
			"org_id": cfg.orgID,
			"tier":   "pro",
		},
		"exp": time.Now().Add(time.Hour * 2).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(cfg.jwtSecret))
	return tokenString
}

func createSession(client *http.Client, cfg loadConfig, jwtToken string) (string, error) {
	reqBody, _ := json.Marshal(map[string]string{"title": "Load Test Worker Session"})
	req, _ := http.NewRequest("POST", cfg.baseURL+"/v1/sessions", bytes.NewBuffer(reqBody))

	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("X-API-Key", cfg.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Org-ID", cfg.orgID)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status code %d: %s", resp.StatusCode, string(body))
	}

	var res SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	return res.ID, nil
}

func printReport(m *Metrics, elapsed time.Duration) {
	avgLatency := time.Duration(0)
	if m.TotalRequests > 0 {
		avgLatency = time.Duration(m.TotalDuration/m.TotalRequests) * time.Microsecond
	}
	throughput := float64(m.TotalRequests) / elapsed.Seconds()

	fmt.Println("\n\033[1;36m=== PERFORMANCE RESULTS ===\033[0m")
	fmt.Printf("Elapsed Time:         %v\n", elapsed)
	fmt.Printf("Total Requests Sent:  %d\n", m.TotalRequests)
	fmt.Printf("Successful Responses: \033[1;32m%d\033[0m\n", m.SuccessCount)
	fmt.Printf("Failed Requests:      \033[1;31m%d\033[0m\n", m.FailureCount)
	fmt.Printf("Average Latency:      %v\n", avgLatency)
	fmt.Printf("Throughput (RPS):     %.2f req/sec\n", throughput)
}
