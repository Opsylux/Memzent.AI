package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/valkey-io/valkey-go"
)

// StreamEntry represents a single entry read from a Valkey Stream.
type StreamEntry struct {
	ID     string
	Fields map[string]string
}

// XAdd appends a message to a Valkey Stream with automatic ID and optional MAXLEN cap.
func (c *MemzentCache) XAdd(ctx context.Context, stream string, maxLen int64, fields map[string]string) (string, error) {
	// Build raw command since valkey-go's typed builder has version-dependent API
	rawArgs := []string{"XADD", stream}
	if maxLen > 0 {
		rawArgs = append(rawArgs, "MAXLEN", "~", fmt.Sprintf("%d", maxLen))
	}
	rawArgs = append(rawArgs, "*")
	for k, v := range fields {
		rawArgs = append(rawArgs, k, v)
	}

	resp := c.client.Do(ctx, c.client.B().Arbitrary(rawArgs[0]).Keys(stream).Args(rawArgs[2:]...).Build())
	if err := resp.Error(); err != nil {
		return "", err
	}
	return resp.ToString()
}

// XGroupCreate creates a consumer group on a stream. If the group already exists, returns nil.
func (c *MemzentCache) XGroupCreate(ctx context.Context, stream, group, startID string) error {
	resp := c.client.Do(ctx, c.client.B().Arbitrary("XGROUP", "CREATE").Keys(stream).Args(group, startID, "MKSTREAM").Build())
	if err := resp.Error(); err != nil {
		// "BUSYGROUP" means group already exists — that's fine
		if err.Error() == "BUSYGROUP Consumer Group name already exists" ||
			(len(err.Error()) > 9 && err.Error()[:9] == "BUSYGROUP") {
			return nil
		}
		return err
	}
	return nil
}

// XReadGroup reads entries from a stream as part of a consumer group.
// count: max entries to read. block: how long to block waiting (0 = no block).
func (c *MemzentCache) XReadGroup(ctx context.Context, group, consumer, stream string, count int64, block time.Duration) ([]StreamEntry, error) {
	blockMs := int64(block / time.Millisecond)

	args := []string{"XREADGROUP", "GROUP", group, consumer, "COUNT", fmt.Sprintf("%d", count)}
	if blockMs > 0 {
		args = append(args, "BLOCK", fmt.Sprintf("%d", blockMs))
	}
	args = append(args, "STREAMS", stream, ">")

	resp := c.client.Do(ctx, c.client.B().Arbitrary(args[0]).Keys(stream).Args(args[1:]...).Build())
	if err := resp.Error(); err != nil {
		if valkey.IsValkeyNil(err) {
			return nil, nil
		}
		return nil, err
	}

	// Parse XREADGROUP response: array of [stream_name, [[id, [field, value, ...]], ...]]
	arr, err := resp.ToArray()
	if err != nil {
		return nil, err
	}
	if len(arr) == 0 {
		return nil, nil
	}

	// First element is the stream data
	streamData, err := arr[0].ToArray()
	if err != nil {
		return nil, err
	}
	if len(streamData) < 2 {
		return nil, nil
	}

	// Second element is the entries array
	entries, err := streamData[1].ToArray()
	if err != nil {
		return nil, err
	}

	var results []StreamEntry
	for _, entry := range entries {
		entryArr, err := entry.ToArray()
		if err != nil {
			continue
		}
		if len(entryArr) < 2 {
			continue
		}

		id, err := entryArr[0].ToString()
		if err != nil {
			continue
		}

		fieldArr, err := entryArr[1].ToArray()
		if err != nil {
			continue
		}

		fields := make(map[string]string)
		for i := 0; i+1 < len(fieldArr); i += 2 {
			k, _ := fieldArr[i].ToString()
			v, _ := fieldArr[i+1].ToString()
			fields[k] = v
		}

		results = append(results, StreamEntry{ID: id, Fields: fields})
	}

	return results, nil
}

// XAck acknowledges messages in a consumer group.
func (c *MemzentCache) XAck(ctx context.Context, stream, group string, ids ...string) error {
	if len(ids) == 0 {
		return nil
	}
	resp := c.client.Do(ctx, c.client.B().Xack().Key(stream).Group(group).Id(ids...).Build())
	return resp.Error()
}

// XAddJSON is a convenience that serializes an object to JSON and adds it as a single "data" field.
func (c *MemzentCache) XAddJSON(ctx context.Context, stream string, maxLen int64, payload any) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}
	return c.XAdd(ctx, stream, maxLen, map[string]string{"data": string(data)})
}

// Client exposes the underlying Valkey client (for packages that need raw access).
func (c *MemzentCache) Client() valkey.Client {
	return c.client
}
