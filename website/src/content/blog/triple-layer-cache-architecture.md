---
title: "Triple-Layer Semantic Cache: How We Achieve Sub-Millisecond Responses"
description: A deep dive into Memzent's three-layer caching architecture — literal, canonical, and semantic — that delivers instant responses while maintaining correctness.
author: Memzent Engineering
category: engineering
tags: caching, valkey, qdrant, performance, architecture
published_at: 2026-06-03
---

## Why Traditional Caching Fails for LLMs

Traditional caching (Redis, Memcached) uses exact-key matching. For LLM prompts, this means only identical prompts hit cache. But users rarely ask the same question the same way twice:

- "What's the capital of France?"
- "capital of france"
- "Tell me the capital city of France"
- "France's capital?"

All of these are the **same question** — but a hash-based cache misses 3 out of 4.

## Our Three-Layer Architecture

We solved this with a cascade of increasingly intelligent matching:

![Cache Layer Diagram](/blog/cache-layers.png)

### Layer 1: Literal Hash (Valkey)

```
Key: org:{org_id}:p:{model}:{sha256(prompt)}
Latency: <1ms
Hit condition: Exact character-for-character match
```

The fastest possible lookup. SHA-256 hash of the raw prompt text stored in Valkey (Redis-compatible). This catches automated/programmatic calls that send identical prompts.

### Layer 1.5: Canonical Hash (Valkey)

```
Key: org:{org_id}:c:{model}:{sha256(normalize(prompt))}
Latency: <1ms
Hit condition: Same text after normalization
```

Before hashing, we normalize the prompt:
- Lowercase everything
- Collapse multiple spaces to single
- Trim leading/trailing whitespace
- Standardize punctuation

This catches human typing variations without any ML overhead.

### Layer 2: Semantic Match (Qdrant via Rust gRPC)

```
Key: Vector similarity in prompts_collection
Latency: 5-15ms
Hit condition: Cosine similarity > 0.95 + numeric guard pass
```

The prompt is embedded into a 384-dimensional vector using `all-MiniLM-L6-v2` (running locally in Rust via FastEmbed — no API calls). This vector is compared against previously stored prompt vectors in Qdrant.

**The Numeric Guard**: After a vector match, we extract numbers from both prompts in positional order. If they differ, the cache hit is rejected. This prevents false positives like `a=10` matching `a=11`.

## Cache Isolation

Every cache entry is scoped by:

| Dimension | Why |
|-----------|-----|
| org_id | Multi-tenant isolation — Org A never sees Org B's cache |
| model | GPT-4 response ≠ Claude response for same prompt |
| cache_type | Literal/canonical/semantic stored in separate key spaces |

## Durable Fallback: Zero-Loss Guarantee

All cache entries are simultaneously persisted to PostgreSQL's `persistent_cache` table (B-Tree indexed). If Valkey crashes:

1. Gateway detects Valkey miss
2. Falls back to Postgres read
3. Returns cached response (no LLM call)
4. Asynchronously backfills Valkey in background

**Result**: Cache hit rate stays at 100% even through infrastructure failures.

## Real-World Performance

From our production deployment over 30 days:

| Metric | Value |
|--------|-------|
| Layer 1 hit rate | 22% |
| Layer 1.5 hit rate | 8% |
| Layer 2 hit rate | 18% |
| Combined hit rate | 48% |
| Avg cache response time | 3ms |
| Avg LLM response time | 2,800ms |
| Cost savings per cached hit | 90% |

## What's Next

- **Cache warming** — Pre-populate cache from common queries
- **TTL policies** — Per-org configurable expiration
- **Cache analytics** — Dashboard showing hit/miss patterns and savings
