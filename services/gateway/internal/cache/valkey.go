// services/gateway/internal/cache/valkey.go
package cache

import (
	"context"
	"time"

	"github.com/valkey-io/valkey-go"
)

type AuraCache struct {
	client valkey.Client
}

func NewAuraCache(ctx context.Context, addr string) (*AuraCache, error) {
	// Valkey Go is native and extremely fast
	client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{addr}})
	if err != nil {
		return nil, err
	}
	return &AuraCache{client: client}, nil
}

func (c *AuraCache) GetSemanticResult(ctx context.Context, key string) (string, error) {
	resp := c.client.Do(ctx, c.client.B().Get().Key(key).Build())
	if err := resp.Error(); err != nil {
		if valkey.IsValkeyNil(err) {
			return "", nil
		}
		return "", err
	}
	return resp.ToString()
}

func (c *AuraCache) SetResult(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.client.Do(ctx, c.client.B().Set().Key(key).Value(value).Px(ttl).Build()).Error()
}

func (c *AuraCache) Close() {
	c.client.Close()
}
