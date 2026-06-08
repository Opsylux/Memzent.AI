---
title: "Spend Limits & Budget Forecasting: Never Get a Surprise Bill Again"
description: Set daily and monthly caps on both dollar spend and token usage. Get real-time burn rate projections and automated alerts before limits are hit.
author: Memzent Team
category: announcement
tags: billing, spend-limits, budget, cost-control, forecasting
published_at: 2026-05-28
---

## The Problem: Runaway AI Costs

LLM costs are unpredictable. A single viral feature or prompt injection can burn through your entire monthly budget in hours. Teams need guardrails — not just visibility.

## Introducing Spend Limits

Set hard caps on your AI spending at both daily and monthly granularity:

```bash
curl -X PUT https://api.memzent.ai/v1/billing/spend-limits \
  -H "X-API-Key: memzent_YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "daily_limit": 50.00,
    "monthly_limit": 1000.00,
    "daily_token_limit": 500000,
    "monthly_token_limit": 10000000
  }'
```

When a limit is exceeded, requests are blocked with a clear `402 Payment Required` error — no silent failures, no unexpected charges.

## Budget Dashboard

Check your full budget status anytime:

```bash
curl https://api.memzent.ai/v1/billing/budget \
  -H "X-API-Key: memzent_YOUR_KEY"
```

```json
{
  "current_balance": 842.50,
  "tier": "pro",
  "daily_avg_spend": 10.69,
  "projected_days_remaining": 78.8,
  "burn_rate_per_hour": 0.445,
  "spend_summaries": [
    { "period": "24h", "total_spend": 12.34, "cache_hits": 42, "cache_savings": 3.20 },
    { "period": "7d",  "total_spend": 74.80, "cache_hits": 312, "cache_savings": 18.90 },
    { "period": "30d", "total_spend": 298.60, "cache_hits": 1280, "cache_savings": 78.40 }
  ],
  "provider_breakdown": [
    { "provider": "openai", "total_spend": 180.20 },
    { "provider": "anthropic", "total_spend": 92.40 },
    { "provider": "ollama", "total_spend": 26.00 }
  ]
}
```

## How Enforcement Works

Spend limits are checked in the engine pipeline **before** any LLM call:

1. Rate limit check ✓
2. Auth & RBAC check ✓
3. **Billing pre-check** ← spend limits enforced here
4. Cache check
5. ...LLM call...

This means you're never charged for a request that exceeds your limits.

## Spend Timeseries for Charts

Build custom dashboards with daily spend data:

```bash
curl "https://api.memzent.ai/v1/billing/spend-timeseries?days=30" \
  -H "X-API-Key: memzent_YOUR_KEY"
```

Returns daily spend, request counts, and cache savings — perfect for Grafana, Datadog, or your own dashboard.

## Cache Savings Are Visible

Every budget response includes `cache_savings` — the dollar amount you saved by serving from cache instead of calling the LLM. Cache hits receive a 90% billing discount, so a $1.00 prompt costs only $0.10 from cache.

## Opt-In Design

All limits are optional. Set `null` to remove any limit:

```bash
curl -X PUT https://api.memzent.ai/v1/billing/spend-limits \
  -H "X-API-Key: memzent_YOUR_KEY" \
  -d '{"daily_limit": null, "monthly_limit": 1000.00}'
```

## Top-Up via Stripe

When your balance runs low, top up directly from the API:

```bash
curl -X POST https://api.memzent.ai/v1/billing/checkout \
  -H "X-API-Key: memzent_YOUR_KEY" \
  -d '{"tier": "pro", "amount": 100.00}'
```

Returns a Stripe checkout URL — complete payment, balance updates instantly.

## What's Next

- **Webhook alerts** at 80%, 90%, 100% of limit
- **Per-user spend allocation** within an org
- **Auto-pause** with configurable resume policies
- **Cost attribution** by session, tool, or prompt category
