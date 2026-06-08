package prewarmer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"memzent-gateway/internal/cache"
	"memzent-gateway/internal/offline/miners"
)

// Config configures the speculative pre-warming worker.
type Config struct {
	TTL          time.Duration // TTL for speculative cache entries (default: 5m)
	MaxBatchSize int           // Max entries to pre-warm per cycle (default: 50)
}

// DefaultConfig returns production defaults.
func DefaultConfig() Config {
	return Config{
		TTL:          5 * time.Minute,
		MaxBatchSize: 50,
	}
}

// ResponseLookup retrieves a cached response by key from persistent storage.
type ResponseLookup func(ctx context.Context, key string) (string, error)

// PreWarmer consumes O2 CacheMiner output and writes speculative cache entries
// to Valkey for high-frequency miss patterns. Tracks prediction accuracy via
// speculative hit/miss counters.
type PreWarmer struct {
	cache      *cache.MemzentCache
	cacheMiner *miners.CacheMiner
	config     Config
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	running    atomic.Bool

	// Metrics
	EntriesWarmed atomic.Uint64
	WarmFailures  atomic.Uint64

	// responseStore provides "last known good response" for a given cache key.
	responseStore ResponseLookup
}

// New creates a new speculative pre-warming worker.
func New(c *cache.MemzentCache, cm *miners.CacheMiner, cfg Config, lookup ResponseLookup) *PreWarmer {
	if cfg.TTL <= 0 {
		cfg.TTL = 5 * time.Minute
	}
	if cfg.MaxBatchSize <= 0 {
		cfg.MaxBatchSize = 50
	}
	return &PreWarmer{
		cache:         c,
		cacheMiner:    cm,
		config:        cfg,
		responseStore: lookup,
	}
}

// Start launches the pre-warming worker goroutine.
func (pw *PreWarmer) Start(ctx context.Context) {
	if pw.running.Swap(true) {
		return
	}

	workerCtx, cancel := context.WithCancel(ctx)
	pw.cancel = cancel

	pw.wg.Add(1)
	go pw.worker(workerCtx)

	slog.Info("🔥 Speculative Pre-Warmer started",
		"ttl", pw.config.TTL,
		"max_batch", pw.config.MaxBatchSize,
	)
}

// Stop gracefully shuts down the pre-warmer.
func (pw *PreWarmer) Stop() {
	if !pw.running.Swap(false) {
		return
	}
	pw.cancel()
	pw.wg.Wait()
	slog.Info("🔥 Speculative Pre-Warmer stopped",
		"entries_warmed", pw.EntriesWarmed.Load(),
		"warm_failures", pw.WarmFailures.Load(),
	)
}

// Stats returns current pre-warmer metrics.
func (pw *PreWarmer) Stats() map[string]interface{} {
	return map[string]interface{}{
		"entries_warmed":      pw.EntriesWarmed.Load(),
		"warm_failures":       pw.WarmFailures.Load(),
		"prediction_accuracy": pw.cacheMiner.PredictionAccuracy(),
		"speculative_hits":    pw.cacheMiner.SpeculativeHits.Load(),
		"speculative_misses":  pw.cacheMiner.SpeculativeMisses.Load(),
	}
}

func (pw *PreWarmer) worker(ctx context.Context) {
	defer pw.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case output := <-pw.cacheMiner.Output():
			pw.processTargets(ctx, output.PreWarmTargets)
		}
	}
}

func (pw *PreWarmer) processTargets(ctx context.Context, targets []miners.SpeculativeEntry) {
	if len(targets) == 0 {
		return
	}

	batch := targets
	if len(batch) > pw.config.MaxBatchSize {
		batch = batch[:pw.config.MaxBatchSize]
	}

	warmed := 0
	for _, target := range batch {
		if ctx.Err() != nil {
			return
		}

		// Look up cached response by prompt hash key
		cacheKey := fmt.Sprintf("org:%s:p:default:%s", target.OrgID, target.PromptHash)

		var response string
		var err error

		if pw.responseStore != nil {
			response, err = pw.responseStore(ctx, cacheKey)
			if err != nil {
				slog.Debug("Pre-warm: persistent lookup failed", "key", cacheKey, "error", err)
			}
		}

		// Try canonical hash fallback
		if response == "" && target.CanonicalHash != "" {
			canonKey := fmt.Sprintf("org:%s:c:default:%s", target.OrgID, target.CanonicalHash)
			if pw.responseStore != nil {
				response, _ = pw.responseStore(ctx, canonKey)
			}
		}

		if response == "" {
			continue
		}

		// Write speculative entry to Valkey
		err = pw.cache.SetResult(ctx, cacheKey, response, pw.config.TTL)
		if err != nil {
			pw.WarmFailures.Add(1)
			continue
		}

		// Also warm entity key if available
		if target.EntityKey != "" {
			entityCacheKey := fmt.Sprintf("org:%s:m:default:e:%s", target.OrgID, target.EntityKey)
			_ = pw.cache.SetResult(ctx, entityCacheKey, response, pw.config.TTL)
		}

		pw.EntriesWarmed.Add(1)
		warmed++
	}

	if warmed > 0 {
		slog.Info("🔥 Pre-warmer: speculative entries written",
			"warmed", warmed, "targets", len(batch),
		)
	}
}
