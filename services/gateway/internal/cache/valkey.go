// services/gateway/internal/cache/valkey.go
package cache

import (
	"context"
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
