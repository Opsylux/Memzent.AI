// services/gateway/internal/invalidation/invalidation.go
//
// Package invalidation implements Memzent's cache invalidation strategy layer
// (issue #11). It complements TTL expiry and the entity-aware cache guard with:
//
//   - Event-driven invalidation: MCP tools/connectors emit change events; entries
//     generated from a tool's output are busted via a Valkey reverse index.
//   - Version-tagged cache keys: a per-org cache version is embedded in cache
//     keys. Bumping it (on policy/config/tool changes) makes prior entries
//     unreachable so they fall away through natural TTL eviction.
//   - Preference isolation: a preference tag (role + system prompt) is embedded
//     in the cache key, so callers with different preferences never share (or
//     overwrite) each other's cached answers.
package invalidation

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"memzent-gateway/internal/metrics"
)

// Change types carried by an InvalidationEvent.
const (
	ChangeToolData   = "tool_data_changed"   // a tool's underlying data changed -> targeted bust
	ChangeToolConfig = "tool_config_changed" // a tool's config/schema changed  -> org version bump
	ChangePolicy     = "policy_changed"      // RBAC / permissions changed       -> org version bump
	ChangeConfig     = "config_changed"      // org settings changed             -> org version bump
)

// InvalidationEvent is the schema for a cache-invalidation signal emitted when
// state that a cached answer depended on has changed.
type InvalidationEvent struct {
	OrgID      string    `json:"org_id"`
	ChangeType string    `json:"change_type"`
	ToolIDs    []string  `json:"tool_ids,omitempty"` // required for tool_data_changed
	Reason     string    `json:"reason,omitempty"`
	Timestamp  time.Time `json:"timestamp,omitempty"`
}

// Result summarizes what an invalidation event busted.
type Result struct {
	ChangeType    string `json:"change_type"`
	KeysDeleted   int64  `json:"keys_deleted"`
	VersionBumped bool   `json:"version_bumped"`
	NewVersion    string `json:"new_version,omitempty"`
}

// Store abstracts the Valkey operations the Invalidator needs, so it can be
// unit-tested without a live cache. *cache.MemzentCache satisfies this.
type Store interface {
	Incr(ctx context.Context, key string) (int64, error)
	GetRaw(ctx context.Context, key string) (string, error)
	SAdd(ctx context.Context, key string, ttl time.Duration, members ...string) error
	SPopAll(ctx context.Context, key string) ([]string, error)
	DelKeys(ctx context.Context, keys ...string) (int64, error)
}

// Invalidator implements the engine-facing invalidation behaviour and the
// operator-facing event handling. It is safe for concurrent use.
type Invalidator struct {
	store Store
	db    *sql.DB       // optional: purge durable persistent_cache rows too
	ttl   time.Duration // TTL for reverse-index sets

	mu       sync.RWMutex
	verCache map[string]verEntry // in-memory org->version cache (hot-path guard)
	verTTL   time.Duration
}

type verEntry struct {
	version string
	expires time.Time
}

// New builds an Invalidator. ttl bounds the lifetime of reverse-index
// bookkeeping (typically the LLM cache TTL).
func New(store Store, db *sql.DB, ttl time.Duration) *Invalidator {
	if ttl <= 0 {
		ttl = time.Hour
	}
	return &Invalidator{
		store:    store,
		db:       db,
		ttl:      ttl,
		verCache: make(map[string]verEntry),
		verTTL:   5 * time.Second,
	}
}

func versionKey(orgID string) string       { return "cachever:" + orgID }
func toolIndexKey(org, tool string) string { return fmt.Sprintf("toolkeys:%s:%s", org, tool) }

// Version returns the current cache version tag for an org. Missing => "0".
// Uses a short in-memory cache so the hot path avoids a Valkey round-trip on
// every request; bounded staleness (verTTL) is acceptable for versioning.
func (i *Invalidator) Version(ctx context.Context, orgID string) string {
	if i == nil || i.store == nil || orgID == "" {
		return ""
	}
	i.mu.RLock()
	e, ok := i.verCache[orgID]
	i.mu.RUnlock()
	if ok && time.Now().Before(e.expires) {
		return e.version
	}

	v, err := i.store.GetRaw(ctx, versionKey(orgID))
	if err != nil {
		slog.Warn("cache version lookup failed; using baseline", "org_id", orgID, "error", err)
		return "0"
	}
	if v == "" {
		v = "0"
	}
	i.setVerCache(orgID, v)
	return v
}

func (i *Invalidator) setVerCache(orgID, v string) {
	i.mu.Lock()
	i.verCache[orgID] = verEntry{version: v, expires: time.Now().Add(i.verTTL)}
	i.mu.Unlock()
}

// Bump increments an org's cache version, immediately making all previously
// cached entries for that org unreachable. Returns the new version.
func (i *Invalidator) Bump(ctx context.Context, orgID string) (string, error) {
	if i == nil || i.store == nil || orgID == "" {
		return "", nil
	}
	n, err := i.store.Incr(ctx, versionKey(orgID))
	if err != nil {
		return "", err
	}
	v := fmt.Sprintf("%d", n)
	i.setVerCache(orgID, v)
	return v, nil
}

// TagKeys records that the given cache keys were produced using the output of
// the given tools, so they can later be busted precisely. Fire-and-forget: it
// never blocks or fails the caller.
func (i *Invalidator) TagKeys(ctx context.Context, orgID string, toolIDs []string, keys []string) {
	if i == nil || i.store == nil || len(toolIDs) == 0 || len(keys) == 0 {
		return
	}
	for _, tid := range toolIDs {
		if tid == "" {
			continue
		}
		if err := i.store.SAdd(ctx, toolIndexKey(orgID, tid), i.ttl, keys...); err != nil {
			slog.Debug("failed to tag cache keys for tool", "tool_id", tid, "error", err)
		}
	}
}

// InvalidateTool busts every cache entry recorded for a tool via TagKeys, and
// purges matching durable rows when a DB is configured. Returns keys deleted.
func (i *Invalidator) InvalidateTool(ctx context.Context, orgID, toolID string) (int64, error) {
	if i == nil || i.store == nil || orgID == "" || toolID == "" {
		return 0, nil
	}
	keys, err := i.store.SPopAll(ctx, toolIndexKey(orgID, toolID))
	if err != nil {
		return 0, err
	}
	if len(keys) == 0 {
		return 0, nil
	}
	deleted, err := i.store.DelKeys(ctx, keys...)
	if err != nil {
		return deleted, err
	}
	i.purgeDurable(ctx, keys)
	return deleted, nil
}

func (i *Invalidator) purgeDurable(ctx context.Context, keys []string) {
	if i.db == nil || len(keys) == 0 {
		return
	}
	placeholders := make([]string, len(keys))
	args := make([]any, len(keys))
	for idx, k := range keys {
		placeholders[idx] = fmt.Sprintf("$%d", idx+1)
		args[idx] = k
	}
	q := "DELETE FROM persistent_cache WHERE cache_key IN (" + strings.Join(placeholders, ",") + ")"
	if _, err := i.db.ExecContext(ctx, q, args...); err != nil {
		slog.Warn("durable cache purge failed (non-fatal)", "error", err)
	}
}

// HandleEvent applies an InvalidationEvent and records metrics. Tool-data changes
// bust indexed keys precisely; all other change types bump the org cache version.
func (i *Invalidator) HandleEvent(ctx context.Context, ev InvalidationEvent) (Result, error) {
	res := Result{ChangeType: ev.ChangeType}
	if i == nil || ev.OrgID == "" {
		return res, fmt.Errorf("org_id required")
	}
	metrics.CacheInvalidationEventsTotal.WithLabelValues(ev.ChangeType).Inc()

	switch ev.ChangeType {
	case ChangeToolData:
		for _, tid := range ev.ToolIDs {
			n, err := i.InvalidateTool(ctx, ev.OrgID, tid)
			if err != nil {
				return res, err
			}
			res.KeysDeleted += n
		}
	case ChangeToolConfig, ChangePolicy, ChangeConfig:
		v, err := i.Bump(ctx, ev.OrgID)
		if err != nil {
			return res, err
		}
		res.VersionBumped = true
		res.NewVersion = v
	default:
		return res, fmt.Errorf("unknown change_type %q", ev.ChangeType)
	}
	return res, nil
}
