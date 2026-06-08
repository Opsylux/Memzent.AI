package offline

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// Miner is the interface all offline miners must implement.
type Miner interface {
	// Name returns the miner identifier (e.g., "O1:RequestMiner")
	Name() string
	// Process handles a single OfflineEvent. Must be safe for concurrent calls.
	Process(ctx context.Context, event OfflineEvent)
	// Flush forces any buffered state to be persisted (called on shutdown).
	Flush(ctx context.Context)
}

// PlaneConfig holds configuration for the offline learning plane.
type PlaneConfig struct {
	BufferSize     int           // Channel buffer capacity (default: 4096)
	WorkerCount    int           // Number of consumer goroutines (default: 4)
	FlushInterval  time.Duration // How often miners flush internal state (default: 30s)
	MaxBacklog     int           // If channel len exceeds this, start dropping (= BufferSize)
}

// DefaultConfig returns production defaults.
func DefaultConfig() PlaneConfig {
	return PlaneConfig{
		BufferSize:    4096,
		WorkerCount:   4,
		FlushInterval: 30 * time.Second,
		MaxBacklog:    4096,
	}
}

// Plane is the offline learning plane. It receives events non-blocking from
// the live request path and dispatches them to registered miners in background
// goroutines. The live path is NEVER blocked or slowed by offline processing.
type Plane struct {
	events  chan OfflineEvent
	miners  []Miner
	config  PlaneConfig
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running atomic.Bool

	// Metrics (exported for observability)
	EventsEmitted  atomic.Uint64
	EventsDropped  atomic.Uint64
	EventsProcessed atomic.Uint64
}

// NewPlane creates a new offline learning plane with the given miners.
func NewPlane(cfg PlaneConfig, miners ...Miner) *Plane {
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 4096
	}
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 4
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = 30 * time.Second
	}
	return &Plane{
		events: make(chan OfflineEvent, cfg.BufferSize),
		miners: miners,
		config: cfg,
	}
}

// Start launches the worker goroutines. Safe to call only once.
func (p *Plane) Start(ctx context.Context) {
	if p.running.Swap(true) {
		return // already running
	}

	workerCtx, cancel := context.WithCancel(ctx)
	p.cancel = cancel

	// Launch consumer workers
	for i := 0; i < p.config.WorkerCount; i++ {
		p.wg.Add(1)
		go p.worker(workerCtx, i)
	}

	// Launch periodic flush ticker
	p.wg.Add(1)
	go p.flusher(workerCtx)

	slog.Info("🧠 Offline Learning Plane started",
		"workers", p.config.WorkerCount,
		"buffer_size", p.config.BufferSize,
		"miners", len(p.miners),
	)
}

// Emit sends an event to the offline plane. Non-blocking: if the buffer is full,
// the event is dropped and a counter is incremented. The live path is never affected.
func (p *Plane) Emit(event OfflineEvent) {
	if !p.running.Load() {
		return
	}
	p.EventsEmitted.Add(1)

	select {
	case p.events <- event:
		// Successfully queued
	default:
		// Channel full — drop gracefully
		p.EventsDropped.Add(1)
	}
}

// Stop gracefully shuts down the offline plane. Drains remaining events and
// calls Flush on all miners.
func (p *Plane) Stop() {
	if !p.running.Swap(false) {
		return
	}
	p.cancel()
	close(p.events)
	p.wg.Wait()

	// Final flush of all miners
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, m := range p.miners {
		m.Flush(ctx)
	}
	slog.Info("🧠 Offline Learning Plane stopped",
		"emitted", p.EventsEmitted.Load(),
		"processed", p.EventsProcessed.Load(),
		"dropped", p.EventsDropped.Load(),
	)
}

// Stats returns current metrics for the offline plane.
func (p *Plane) Stats() map[string]uint64 {
	return map[string]uint64{
		"emitted":   p.EventsEmitted.Load(),
		"processed": p.EventsProcessed.Load(),
		"dropped":   p.EventsDropped.Load(),
		"pending":   uint64(len(p.events)),
	}
}

// worker is the main consumer loop. Each worker pulls events and fans out to all miners.
func (p *Plane) worker(ctx context.Context, id int) {
	defer p.wg.Done()
	for {
		select {
		case event, ok := <-p.events:
			if !ok {
				return // channel closed
			}
			p.EventsProcessed.Add(1)
			for _, m := range p.miners {
				m.Process(ctx, event)
			}
		case <-ctx.Done():
			// Drain remaining events before exit
			for {
				select {
				case event, ok := <-p.events:
					if !ok {
						return
					}
					p.EventsProcessed.Add(1)
					for _, m := range p.miners {
						m.Process(ctx, event)
					}
				default:
					return
				}
			}
		}
	}
}

// flusher periodically triggers Flush on all miners to persist intermediate state.
func (p *Plane) flusher(ctx context.Context) {
	defer p.wg.Done()
	ticker := time.NewTicker(p.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, m := range p.miners {
				m.Flush(ctx)
			}
		case <-ctx.Done():
			return
		}
	}
}
