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
| Preference-drift detection | Per cache hit | Per entry | Rejects hits whose preference context drifted |
| Manual flush (`/v1/cache/flush`) | Operator action | Per org | Purges Valkey / durable / Qdrant |

## 1. Version-tagged cache keys

Every cache key embeds the org's current **cache version**:

```
org:{orgID}:ver:{version}:m:{model}:{type}:{value}
```

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

## 3. Preference-drift detection

Two users (or the same user over time) may share a prompt but differ in the
context that shapes a good answer — role, provider/model, and the system prompt /
persona. Memzent records a **preference fingerprint** with each cached entry and
compares it on retrieval.

- **Fingerprint:** an order-independent token set derived from role, provider,
  model, and normalized system-prompt words (stop-words removed).
- **Comparison:** on a cache hit, the current fingerprint is compared with the
  stored one using **Jaccard similarity**. If similarity is below
  `PREFERENCE_DRIFT_THRESHOLD` (default `0.85`), the hit is treated as a **miss**
  and a fresh response is generated.
- Entries written without a fingerprint (or before this feature) are never
  considered stale, so behaviour degrades gracefully.

## Metrics

Exposed on `/metrics` (Prometheus):

| Metric | Type | Meaning |
|--------|------|---------|
| `memzent_cache_invalidation_events_total{change_type}` | counter | Invalidation events processed, by change type |
| `memzent_stale_hit_avoided_total` | counter | Cache hits rejected due to preference drift (served fresh instead) |

## Configuration

| Env var | Default | Purpose |
|---------|---------|---------|
| `LLM_CACHE_TTL` | `1h` | Lifetime of cache entries, reverse-index sets, and fingerprints |
| `PREFERENCE_DRIFT_THRESHOLD` | `0.85` | Min fingerprint similarity to serve a cached entry |

## Implementation map

- `internal/invalidation/` — event schema, `Invalidator`, fingerprint + Jaccard,
  and the `POST /v1/cache/invalidate` handler.
- `internal/cache/invalidation_ops.go` — low-level Valkey ops (version counter,
  reverse-index sets, fingerprint keys).
- `internal/engine/engine.go` — version threading into cache keys, staleness
  guard on every cache hit, and reverse-index / fingerprint writes on populate.
