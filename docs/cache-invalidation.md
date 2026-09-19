# Cache Invalidation Strategy

Semantic caching answers the question *"have we already paid for equivalent
work?"*. Invalidation answers the harder follow-up: *"is a previously equivalent
answer still a **safe** answer?"* When the repo, policy, tool data, or user
preferences change, a cached response can silently become wrong. Memzent layers
several complementary mechanisms so stale answers are not served.

## Layers at a glance

| Mechanism | Trigger | Scope | Effect |
|-----------|---------|-------|--------|
| TTL expiry (`LLM_CACHE_TTL`) | Time | Per entry | Entry expires after the TTL |
| Entity-aware guard (E1) | Per request | Per entry | Rejects semantic matches with different operational entities |
| Version-tagged keys | Config/policy/tool-config change | Per org | Prior entries become unreachable |
| Event-driven invalidation | Tool data change | Per tool | Busts entries produced from that tool's output |
| Preference-partitioned keys | Per request | Per preference | Callers with different preferences never share/overwrite entries |
| Manual flush (`/v1/cache/flush`) | Operator action | Per org | Purges Valkey / durable / Qdrant |

## 1. Version-tagged cache keys

Every cache key embeds the org's current **cache version** and, when the caller
has preference signals, a **preference tag**:

```
org:{orgID}:ver:{version}:pf:{prefTag}:m:{model}:{type}:{value}
```

Both segments are optional: an org with no bumps and a request with no preference
signals produces the legacy `org:{orgID}:m:{model}:{type}:{value}` form, so
existing keys and maximum cache sharing are preserved.

The version is an integer counter stored in Valkey at `cachever:{orgID}` and
resolved once per request (with a short in-process cache so the hot path avoids a
round-trip). **Bumping** the version (`INCR`) makes every key generated under the
previous version unreachable; those entries then fall away through normal TTL
eviction — no scan-and-delete required.

Bumps happen when configuration that affects *how* a response is produced
changes, e.g. the org similarity threshold (`PUT /v1/settings/threshold`), and on
`policy_changed` / `config_changed` / `tool_config_changed` invalidation events.

> Keys keep the `org:{orgID}:*` prefix, so `FlushByPattern` and the existing
> `/v1/cache/flush` endpoint continue to match versioned keys.

## 2. Event-driven invalidation

When a connected tool's underlying data changes (repo pushed, DB schema altered),
entries produced *using that tool's output* should be busted precisely rather
than waiting for TTL.

- **Reverse index:** when a fresh response is cached, every tool that contributed
  output is recorded in a Valkey set `toolkeys:{orgID}:{toolID}` pointing at the
  written cache keys.
- **Bust:** an invalidation event reads that set, deletes the referenced Valkey
  keys, purges matching durable `persistent_cache` rows, and clears the index.

### Emitting an event

```
POST /v1/cache/invalidate          (scope: tools:write)
Content-Type: application/json

{ "change_type": "tool_data_changed", "tool_ids": ["github-repo"], "reason": "push" }
```

`change_type` values:

| Change type | Action |
|-------------|--------|
| `tool_data_changed` | Targeted bust of the given `tool_ids` via the reverse index |
| `tool_config_changed` | Bump org cache version |
| `policy_changed` | Bump org cache version (RBAC / permissions changed) |
| `config_changed` | Bump org cache version (org settings changed) |

The org is always taken from the authenticated request context — never from the
body — preserving tenant isolation.

## 3. Preference-partitioned keys

Two users (or the same user over time) may share a prompt but differ in the
context that shapes a good answer — RBAC role and the system prompt / persona.
Rather than storing one shared answer and comparing preferences on read (which
lets different callers overwrite each other's value), Memzent folds a
**preference tag** directly into the cache key. Callers with different
preferences therefore address *different* cache slots and can never serve or
clobber one another's answers.

- **Fingerprint:** an order-independent token set derived from role and
  normalized system-prompt words (stop-words removed). Provider and model are
  intentionally excluded because the key already partitions on the model.
- **Tag:** a short SHA-256 hash of the fingerprint, inserted as a `pf:{tag}`
  key segment. When there are no preference signals the tag is empty and the
  key stays un-partitioned for maximum sharing.
- **Effect:** any change in a caller's preferences yields a new key — a clean
  cache miss and a fresh response — with no read-time comparison or extra
  round-trip.

## Metrics

Exposed on `/metrics` (Prometheus):

| Metric | Type | Meaning |
|--------|------|---------|
| `memzent_cache_invalidation_events_total{change_type}` | counter | Invalidation events processed, by change type |

## Configuration

| Env var | Default | Purpose |
|---------|---------|---------|
| `LLM_CACHE_TTL` | `1h` | Lifetime of cache entries and reverse-index sets |

## Implementation map

- `internal/invalidation/` — event schema, `Invalidator`, preference fingerprint
  + tag, and the `POST /v1/cache/invalidate` handler.
- `internal/cache/invalidation_ops.go` — low-level Valkey ops (version counter,
  reverse-index sets).
- `internal/engine/engine.go` — version + preference threading into cache keys
  and reverse-index writes on populate.
