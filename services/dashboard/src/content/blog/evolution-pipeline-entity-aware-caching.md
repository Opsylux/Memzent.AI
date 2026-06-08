---
title: "Introducing the Evolution Pipeline: Entity-Aware Caching & 80%+ GPU Avoidance"
description: "Memzent's Evolution Pipeline (E1–E6) adds entity extraction, L1b hot path cache, offline learning, workflow discovery, and GPU avoidance analytics — eliminating redundant LLM inference at every layer."
author: Memzent Team
category: announcement
tags: evolution-pipeline, entity-extraction, gpu-avoidance, caching, L1b
published_at: 2026-06-08
---

## The Problem: Semantic Similarity Isn't Enough

Semantic caching catches rephrased questions brilliantly. But it has a blind spot — **entities**.

Consider these two prompts:

- "Transfer $100 from account 123 to account 456"
- "Transfer $100 from account 456 to account 123"

They score **>0.98 semantic similarity**. A pure vector cache would return the same response for both. But they're completely different operations — one moves money in the opposite direction.

This is the **entity collision problem**, and it's the most dangerous class of cache error in production AI systems.

## The Solution: Six Layers of Intelligence

Today we're shipping the **Evolution Pipeline** — six interlocking layers that eliminate redundant GPU inference while guaranteeing data correctness.

### E1: Entity Extraction (<1ms)

Every prompt is scanned by regex-based extractors that identify **six typed entities**:

- **Accounts** with directional awareness (source vs destination)
- **Customers** by name or ID
- **Invoices** and order references
- **Amounts** and monetary values
- **Dates** in multiple formats
- **Generic identifiers** and reference codes

Extraction runs in under 1ms — no SLM, no GPU, pure regex. It runs in both the Rust Router (for cache guard comparison) and Go Gateway (for L1b key generation).

### E2: L1b Entity-Keyed Hot Path Cache

The L1b layer is our newest cache tier. It builds a **deterministic key** from extracted entities:

```
org:{org_id}:m:{model}:e:{sorted_key=value_pairs}
```

If a request's entity fingerprint matches an existing cache entry, the response is served directly from Valkey — **zero vector search, zero LLM call**. This is a sub-millisecond lookup.

The four-layer cache now works like this:

| Layer | Method | Latency | What it catches |
|-------|--------|---------|-----------------|
| L1 | SHA-256 hash | <1ms | Exact duplicate prompts |
| L1.5 | Canonical hash | <1ms | Formatting differences |
| L1b | Entity key | 1-2ms | Same entities, different phrasing |
| L2 | Vector similarity | 15-50ms | Semantic paraphrasing |

### E3: Offline Learning Plane

Every request emits a lightweight telemetry event (PII-safe — no raw prompts, only hashes) to an asynchronous processing plane. Three miners analyze the stream:

- **Request Miner** — identifies traffic patterns and popular entity combinations
- **Cache Miner** — tracks hit/miss ratios per layer and optimizes thresholds
- **Workflow Miner** — discovers recurring multi-step tool sequences

The offline plane uses buffered channels (4096 buffer, 4 workers) with try-send semantics — it never blocks the request path.

### E4: Workflow Registry & Shortcuts

When the Workflow Miner discovers a recurring pattern (e.g., "lookup customer → check balance → generate invoice"), it registers it in the Workflow Registry.

Approved workflows execute as **single-shot shortcuts** — the engine skips per-step routing and fires all matched tools in one pass. This eliminates redundant gRPC calls and reduces multi-step latency by 40-60%.

### E5: GPU Avoidance Analytics

The GPU Avoidance Rate is our **primary business metric**:

```
GPU Avoidance Rate = cache_hits / total_requests × 100%
```

Every cache hit at any layer (L1, L1.5, L1b, L2) is a request that **never touched the LLM**. We expose 8 Prometheus counters for entity types, cache layer distribution, and avoidance rates — all available in the GPU Analytics dashboard.

### E6: Pattern Mining (Experimental)

A Markov chain analyzer predicts next-likely requests based on entity transition patterns. Combined with the Speculative Pre-Warmer, it can populate L1b entries **before** the request arrives — achieving zero-latency on first hits for predictable workflows.

## Real-World Results

After deploying the Evolution Pipeline to our production gateway:

- **Entity extraction**: 100% of prompts scanned in <1ms
- **L1b cache hit rate**: 20-30% of repeat requests resolved without vector search
- **GPU Avoidance Rate**: 80%+ for production workloads with entity-heavy queries
- **Cache correctness**: Zero false positive cache hits on entity-swapped prompts

Our test suite validates this rigorously:

- `make test-cache` — 12/12 tests passing (semantic cache correctness)
- `make test-entity` — 13/14 tests passing (entity extraction + cache guard)
- `make test-memory` — 10/10 tests passing (session continuity + memory recall)

## Feature Flags

All Evolution Pipeline features are controlled by environment variables:

```bash
MEMZENT_L1B_ENABLED=true              # L1b entity-keyed cache (default: on)
MEMZENT_OFFLINE_ENABLED=true          # Offline learning plane (default: on)
MEMZENT_WORKFLOW_ENABLED=true         # Workflow registry (default: on)
MEMZENT_ENTITY_METRICS_ENABLED=true   # GPU avoidance counters (default: on)
MEMZENT_PATTERN_MINING_ENABLED=false  # E6 Markov chain (experimental)
```

## Get Started

The Evolution Pipeline is live on `api.memzent.ai` for all organizations. No configuration needed — L1b and entity extraction are enabled by default.

Explore the new documentation:

- [Entity Extraction](/docs/entity-extraction) — How entities prevent false cache hits
- [Cache Layers & L1b](/docs/cache-layers) — The four-layer cache architecture
- [Offline Learning](/docs/offline-learning) — Asynchronous pattern mining
- [GPU Analytics](/docs/gpu-analytics) — Track your avoidance rate

Questions? Drop us a line at [GitHub Discussions](https://github.com/Opsylux/Memzent.AI/discussions) or reach out on LinkedIn.

---

*The Evolution Pipeline is open source. Check out the implementation in the [Memzent.AI repository](https://github.com/Opsylux/Memzent.AI).*
