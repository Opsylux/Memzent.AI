package offline

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"memzent-gateway/internal/cache"
)

const (
	// DefaultStreamName is the Valkey Stream key for offline events.
	DefaultStreamName = "memzent:offline:events"
	// DefaultConsumerGroup is the consumer group for offline miners.
	DefaultConsumerGroup = "offline-miners"
	// DefaultMaxStreamLen caps the stream to prevent unbounded growth.
	DefaultMaxStreamLen = 10000
)

// StreamPlane is an alternative offline plane that uses Valkey Streams for
// crash-durable, multi-instance event distribution. It replaces in-memory
// channels with XADD/XREADGROUP for the event bus while keeping the same
// Miner interface for processing.
type StreamPlane struct {
	cache    *cache.MemzentCache
	miners   []Miner
	config   StreamConfig
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	running  atomic.Bool

	// Metrics
	EventsEmitted   atomic.Uint64
	EventsProcessed atomic.Uint64
	EventsDropped   atomic.Uint64
	EventsFailed    atomic.Uint64
}

// StreamConfig configures the Valkey Streams-backed offline plane.
type StreamConfig struct {
	StreamName    string        // Valkey Stream key (default: memzent:offline:events)
	ConsumerGroup string       // Consumer group name (default: offline-miners)
	ConsumerName  string       // This instance's consumer name (should be unique per gateway)
	MaxStreamLen  int64         // MAXLEN ~ cap on stream entries (default: 10000)
	WorkerCount   int           // Number of stream reader goroutines (default: 2)
	BatchSize     int64         // Entries per XREADGROUP call (default: 50)
	BlockTimeout  time.Duration // Block duration per read (default: 5s)
	FlushInterval time.Duration // Miner flush interval (default: 30s)
}

// DefaultStreamConfig returns production defaults for Valkey Streams mode.
func DefaultStreamConfig(consumerName string) StreamConfig {
	if consumerName == "" {
		consumerName = fmt.Sprintf("gateway-%d", time.Now().UnixNano()%10000)
	}
	return StreamConfig{
		StreamName:    DefaultStreamName,
		ConsumerGroup: DefaultConsumerGroup,
		ConsumerName:  consumerName,
		MaxStreamLen:  DefaultMaxStreamLen,
		WorkerCount:   2,
		BatchSize:     50,
		BlockTimeout:  5 * time.Second,
		FlushInterval: 30 * time.Second,
	}
}

// NewStreamPlane creates a Valkey Streams-backed offline plane.
func NewStreamPlane(c *cache.MemzentCache, cfg StreamConfig, miners ...Miner) *StreamPlane {
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 2
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 50
	}
	if cfg.BlockTimeout <= 0 {
		cfg.BlockTimeout = 5 * time.Second
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = 30 * time.Second
	}
	return &StreamPlane{
		cache:  c,
		miners: miners,
		config: cfg,
	}
}

// Start creates the consumer group and launches workers.
func (sp *StreamPlane) Start(ctx context.Context) {
	if sp.running.Swap(true) {
		return
	}

	// Create consumer group (idempotent — ignores BUSYGROUP)
	err := sp.cache.XGroupCreate(ctx, sp.config.StreamName, sp.config.ConsumerGroup, "0")
	if err != nil {
		slog.Error("Failed to create stream consumer group", "error", err)
	}

	workerCtx, cancel := context.WithCancel(ctx)
	sp.cancel = cancel

	// Launch stream consumer workers
	for i := 0; i < sp.config.WorkerCount; i++ {
		sp.wg.Add(1)
		go sp.worker(workerCtx, i)
	}

	// Launch periodic flusher
	sp.wg.Add(1)
	go sp.flusher(workerCtx)

	slog.Info("🌊 Stream Offline Plane started",
		"stream", sp.config.StreamName,
		"group", sp.config.ConsumerGroup,
		"consumer", sp.config.ConsumerName,
		"workers", sp.config.WorkerCount,
	)
}

// Emit publishes an event to the Valkey Stream. Non-blocking: if the write fails,
// the event is dropped and a counter incremented.
func (sp *StreamPlane) Emit(event OfflineEvent) {
	if !sp.running.Load() {
		return
	}
	sp.EventsEmitted.Add(1)

	// Fire-and-forget async write to avoid blocking the live path
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, err := sp.cache.XAddJSON(ctx, sp.config.StreamName, sp.config.MaxStreamLen, event)
		if err != nil {
			sp.EventsDropped.Add(1)
			slog.Warn("Stream emit failed", "error", err)
		}
	}()
}

// Stop gracefully drains workers and flushes miners.
func (sp *StreamPlane) Stop() {
	if !sp.running.Swap(false) {
		return
	}
	sp.cancel()
	sp.wg.Wait()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, m := range sp.miners {
		m.Flush(ctx)
	}
	slog.Info("🌊 Stream Offline Plane stopped",
		"emitted", sp.EventsEmitted.Load(),
		"processed", sp.EventsProcessed.Load(),
		"dropped", sp.EventsDropped.Load(),
	)
}

// Stats returns current plane metrics.
func (sp *StreamPlane) Stats() map[string]uint64 {
	return map[string]uint64{
		"emitted":   sp.EventsEmitted.Load(),
		"processed": sp.EventsProcessed.Load(),
		"dropped":   sp.EventsDropped.Load(),
		"failed":    sp.EventsFailed.Load(),
	}
}

// worker reads from the Valkey Stream via XREADGROUP and dispatches to miners.
func (sp *StreamPlane) worker(ctx context.Context, id int) {
	defer sp.wg.Done()
	consumer := fmt.Sprintf("%s-%d", sp.config.ConsumerName, id)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		entries, err := sp.cache.XReadGroup(
			ctx, sp.config.ConsumerGroup, consumer,
			sp.config.StreamName, sp.config.BatchSize, sp.config.BlockTimeout,
		)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			sp.EventsFailed.Add(1)
			slog.Warn("Stream read error", "worker", id, "error", err)
			time.Sleep(time.Second) // backoff on error
			continue
		}

		if len(entries) == 0 {
			continue
		}

		var ackIDs []string
		for _, entry := range entries {
			data, ok := entry.Fields["data"]
			if !ok {
				ackIDs = append(ackIDs, entry.ID)
				continue
			}

			var event OfflineEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				slog.Warn("Stream event unmarshal failed", "id", entry.ID, "error", err)
				ackIDs = append(ackIDs, entry.ID)
				continue
			}

			// Dispatch to all miners
			for _, m := range sp.miners {
				m.Process(ctx, event)
			}
			sp.EventsProcessed.Add(1)
			ackIDs = append(ackIDs, entry.ID)
		}

		// Acknowledge processed entries
		if len(ackIDs) > 0 {
			if err := sp.cache.XAck(ctx, sp.config.StreamName, sp.config.ConsumerGroup, ackIDs...); err != nil {
				slog.Warn("Stream ACK failed", "worker", id, "error", err)
			}
		}
	}
}

// flusher periodically triggers Flush on all miners.
func (sp *StreamPlane) flusher(ctx context.Context) {
	defer sp.wg.Done()
	ticker := time.NewTicker(sp.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, m := range sp.miners {
				m.Flush(ctx)
			}
		case <-ctx.Done():
			return
		}
	}
}
