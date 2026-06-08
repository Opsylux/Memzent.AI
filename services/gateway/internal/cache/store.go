package cache

import (
	"context"
	"time"
)

// Store is the semantic cache contract used by the engine pipeline.
type Store interface {
	GetSemanticResult(ctx context.Context, key string) (string, error)
	SetResult(ctx context.Context, key, value string, ttl time.Duration) error
	RateLimit(ctx context.Context, key string, limit int64) (bool, error)
}
