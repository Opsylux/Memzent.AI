---
title: How Semantic Caching Prevents False Positives in Parametric Queries
description: Deep dive into how Memzent's numeric guard prevents incorrect cache hits when query parameters differ by small values.
author: Memzent Engineering
category: engineering
tags: caching, vectors, embeddings, qdrant, rust
published_at: 2026-06-06
---

## The Problem

Sentence embedding models like `all-MiniLM-L6-v2` are designed to capture semantic meaning — but they struggle with numerical precision. Consider these two prompts:

- "What is (a+b)^2 where a=6 and b=10?"
- "What is (a+b)^2 where a=6 and b=11?"

These are **semantically identical** in structure (same formula, same phrasing) but produce **completely different answers** (256 vs 289). A naive vector similarity check would score these above 0.95 — triggering a false cache hit.

## Our Solution: The Numeric Guard

We implemented a two-layer protection against parametric false positives:

### 1. Raised Similarity Threshold (0.88 → 0.95)

The default semantic cache threshold was too permissive. By raising it to 0.95, we eliminate most ambiguous matches while still catching genuine rephrasings.

### 2. Positional Numeric Comparison

After a vector match exceeds the threshold, we extract all numbers from both prompts **in order of appearance** and compare them:

```rust
fn extract_numbers(text: &str) -> Vec<String> {
    let re = Regex::new(r"\d+\.?\d*").unwrap();
    re.find_iter(text).map(|m| m.as_str().to_string()).collect()
}
```

Key design decisions:

- **Positional, not sorted** — `a=2, b=5` and `a=5, b=2` are treated as different (non-commutative safety)
- **All numbers extracted** — catches parameters anywhere in the prompt
- **Stored alongside vectors** — prompt text is saved in Qdrant payload for comparison

## Results

| Scenario | Before | After |
|----------|--------|-------|
| Same prompt, same numbers | ✅ HIT | ✅ HIT |
| Rephrased, same numbers | ✅ HIT | ✅ HIT |
| Same structure, different numbers | ❌ FALSE HIT | ✅ MISS |
| Completely different prompt | ✅ MISS | ✅ MISS |

## Trade-offs

We intentionally sacrifice one class of valid cache hits: **commutative parameter swaps** (e.g., `a=2, b=5` vs `a=5, b=2` for addition). The cost of one extra LLM call is always less than the cost of returning a wrong answer.

For most real-world workloads, the cache hit rate remains excellent because:
- Users typically ask the same question with the same parameters
- Rephrasings of the same question (different words, same numbers) still hit cache
- The semantic layer catches questions that literal/canonical layers miss
