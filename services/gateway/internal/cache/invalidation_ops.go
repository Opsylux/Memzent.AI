// services/gateway/internal/cache/invalidation_ops.go
package cache

import (
	"context"
	"time"

	"github.com/valkey-io/valkey-go"
)

// Incr atomically increments the integer stored at key and returns the new value.
// If the key does not exist it is first set to 0, so the first call returns 1.
// Used by the cache versioner to bump per-org cache versions.
func (c *MemzentCache) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Do(ctx, c.client.B().Incr().Key(key).Build()).ToInt64()
}

// GetRaw returns the string value at key, or "" (nil error) when the key is missing.
func (c *MemzentCache) GetRaw(ctx context.Context, key string) (string, error) {
	resp := c.client.Do(ctx, c.client.B().Get().Key(key).Build())
	if err := resp.Error(); err != nil {
		if valkey.IsValkeyNil(err) {
			return "", nil
		}
		return "", err
	}
	return resp.ToString()
}

// SetRaw sets key to value with the given TTL. A non-positive ttl stores the key
// without expiry.
func (c *MemzentCache) SetRaw(ctx context.Context, key, value string, ttl time.Duration) error {
	if ttl <= 0 {
		return c.client.Do(ctx, c.client.B().Set().Key(key).Value(value).Build()).Error()
	}
	return c.client.Do(ctx, c.client.B().Set().Key(key).Value(value).Px(ttl).Build()).Error()
}

// SAdd adds one or more members to the set at key and refreshes its TTL.
// The set is used as a reverse index (tool -> cache keys) for targeted invalidation.
func (c *MemzentCache) SAdd(ctx context.Context, key string, ttl time.Duration, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	if err := c.client.Do(ctx, c.client.B().Sadd().Key(key).Member(members...).Build()).Error(); err != nil {
		return err
	}
	if ttl > 0 {
		_ = c.client.Do(ctx, c.client.B().Pexpire().Key(key).Milliseconds(ttl.Milliseconds()).Build()).Error()
	}
	return nil
}

// SPopAll returns all members of the set at key and deletes the set atomically-ish
// (SMEMBERS then DEL). Returns an empty slice when the set is missing.
func (c *MemzentCache) SPopAll(ctx context.Context, key string) ([]string, error) {
	resp := c.client.Do(ctx, c.client.B().Smembers().Key(key).Build())
	if err := resp.Error(); err != nil {
		if valkey.IsValkeyNil(err) {
			return nil, nil
		}
		return nil, err
	}
	members, err := resp.AsStrSlice()
	if err != nil {
		return nil, err
	}
	if len(members) > 0 {
		_ = c.client.Do(ctx, c.client.B().Del().Key(key).Build()).Error()
	}
	return members, nil
}

// DelKeys deletes the given keys and returns the number removed.
func (c *MemzentCache) DelKeys(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	return c.client.Do(ctx, c.client.B().Del().Key(keys...).Build()).ToInt64()
}
