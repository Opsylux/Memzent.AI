package prewarmer

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.TTL != 5*time.Minute {
		t.Errorf("DefaultConfig TTL = %v, want 5m", cfg.TTL)
	}
	if cfg.MaxBatchSize != 50 {
		t.Errorf("DefaultConfig MaxBatchSize = %d, want 50", cfg.MaxBatchSize)
	}
}

func TestNew_AppliesDefaults(t *testing.T) {
	pw := New(nil, nil, Config{}, nil)
	if pw.config.TTL != 5*time.Minute {
		t.Errorf("TTL should default to 5m, got %v", pw.config.TTL)
	}
	if pw.config.MaxBatchSize != 50 {
		t.Errorf("MaxBatchSize should default to 50, got %d", pw.config.MaxBatchSize)
	}
}

func TestNew_RespectsCustomConfig(t *testing.T) {
	cfg := Config{TTL: 10 * time.Minute, MaxBatchSize: 100}
	pw := New(nil, nil, cfg, nil)
	if pw.config.TTL != 10*time.Minute {
		t.Errorf("TTL should be 10m, got %v", pw.config.TTL)
	}
	if pw.config.MaxBatchSize != 100 {
		t.Errorf("MaxBatchSize should be 100, got %d", pw.config.MaxBatchSize)
	}
}

func TestStartStop_Idempotent(t *testing.T) {
	// Start/stop without cache or miner should not panic
	pw := New(nil, nil, DefaultConfig(), nil)
	// Just verify double-stop doesn't panic
	pw.Stop()
	pw.Stop()
}
