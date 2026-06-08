// services/gateway/internal/cache/valkey.go
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/valkey-io/valkey-go"
)

type MemzentCache struct {
	client valkey.Client
}

func NewMemzentCache(ctx context.Context, addr string) (*MemzentCache, error) {
	// Valkey Go is native and extremely fast
	client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{addr}})
	if err != nil {
		return nil, err
	}
	return &MemzentCache{client: client}, nil
}

func (c *MemzentCache) GetSemanticResult(ctx context.Context, key string) (string, error) {
	resp := c.client.Do(ctx, c.client.B().Get().Key(key).Build())
	if err := resp.Error(); err != nil {
		if valkey.IsValkeyNil(err) {
			return "", nil
		}
		return "", err
	}
	return resp.ToString()
}

func (c *MemzentCache) SetResult(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.client.Do(ctx, c.client.B().Set().Key(key).Value(value).Px(ttl).Build()).Error()
}

func (c *MemzentCache) Ping(ctx context.Context) error {
	return c.client.Do(ctx, c.client.B().Ping().Build()).Error()
}

func (c *MemzentCache) Close() {
	c.client.Close()
}

// FlushByPattern deletes all keys matching the given glob pattern.
// Use with care — intended for cache invalidation after schema changes.
func (c *MemzentCache) FlushByPattern(ctx context.Context, pattern string) (int64, error) {
	var deleted int64
	var cursor uint64

	for {
		cmd := c.client.B().Scan().Cursor(cursor).Match(pattern).Count(100).Build()
		resp := c.client.Do(ctx, cmd)
		if err := resp.Error(); err != nil {
			return deleted, err
		}

		entry, err := resp.AsScanEntry()
		if err != nil {
			return deleted, err
		}

		if len(entry.Elements) > 0 {
			delArgs := c.client.B().Del().Key(entry.Elements...).Build()
			delResp := c.client.Do(ctx, delArgs)
			if err := delResp.Error(); err == nil {
				n, _ := delResp.ToInt64()
				deleted += n
			}
		}

		cursor = entry.Cursor
		if cursor == 0 {
			break
		}
	}
	return deleted, nil
}

// RateLimit implements a sliding window rate limiter using Valkey.
// Returns (allowed bool, err error). Uses a 60-second sliding window.
// key: unique identifier (e.g. "rl:<orgID>:<userID>")
// limit: max requests allowed per window
func (c *MemzentCache) RateLimit(ctx context.Context, key string, limit int64) (bool, error) {
	now := time.Now().UnixMilli()
	windowMs := int64(60000) // 60-second window
	windowStart := now - windowMs

	// Lua script: remove expired entries, add current request, count, set TTL
	script := valkey.NewLuaScript(`
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local window_start = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local window_ms = tonumber(ARGV[4])

		redis.call('ZREMRANGEBYSCORE', key, '-inf', window_start)
		local count = redis.call('ZCARD', key)

		if count < limit then
			redis.call('ZADD', key, now, now .. '-' .. math.random(1000000))
			redis.call('PEXPIRE', key, window_ms)
			return 1
		end
		return 0
	`)

	resp := script.Exec(ctx, c.client, []string{key}, []string{
		fmt.Sprintf("%d", now),
		fmt.Sprintf("%d", windowStart),
		fmt.Sprintf("%d", limit),
		fmt.Sprintf("%d", windowMs),
	})

	if err := resp.Error(); err != nil {
		return true, err // fail-open on error
	}

	val, err := resp.ToInt64()
	if err != nil {
		return true, err
	}

	return val == 1, nil
}
