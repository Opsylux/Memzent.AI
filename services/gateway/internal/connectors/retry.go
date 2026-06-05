package connectors

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// RetryConfig defines the retry behavior for tool execution.
type RetryConfig struct {
	MaxAttempts    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
	JitterFactor  float64 // 0.0 to 1.0 — randomizes delay to avoid thundering herd
}

// DefaultRetryConfig provides sensible defaults for tool execution retries.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:    3,
		InitialDelay:  200 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		JitterFactor:  0.3,
	}
}

// ExecuteWithRetry wraps a connector execution with exponential backoff and jitter.
// Returns the first successful response, or the last error after all retries are exhausted.
func ExecuteWithRetry(
	ctx context.Context,
	connector Connector,
	req *ExecutionRequest,
	cfg RetryConfig,
) (*ExecutionResponse, error) {
	var lastErr error
	var lastResp *ExecutionResponse

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		resp, err := connector.Execute(ctx, req)

		// Success: return immediately
		if err == nil && resp != nil && resp.Status != "error" {
			return resp, nil
		}

		// Record the failure
		lastErr = err
		lastResp = resp

		// Don't retry on context cancellation/timeout
		if ctx.Err() != nil {
			break
		}

		// Don't sleep after the last attempt
		if attempt == cfg.MaxAttempts-1 {
			break
		}

		// Calculate delay with exponential backoff + jitter
		delay := calculateBackoff(attempt, cfg)

		select {
		case <-ctx.Done():
			return lastResp, ctx.Err()
		case <-time.After(delay):
			// Continue to next retry
		}
	}

	if lastErr != nil {
		return lastResp, lastErr
	}
	return lastResp, nil
}

func calculateBackoff(attempt int, cfg RetryConfig) time.Duration {
	delay := float64(cfg.InitialDelay) * math.Pow(cfg.BackoffFactor, float64(attempt))
	if delay > float64(cfg.MaxDelay) {
		delay = float64(cfg.MaxDelay)
	}

	// Apply jitter: delay ± (jitterFactor * delay)
	if cfg.JitterFactor > 0 {
		jitter := delay * cfg.JitterFactor * (2*rand.Float64() - 1)
		delay += jitter
	}

	if delay < 0 {
		delay = 0
	}
	return time.Duration(delay)
}
