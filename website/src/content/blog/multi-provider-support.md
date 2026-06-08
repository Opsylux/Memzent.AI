---
title: "Announcing Multi-Provider Support: One API, Four LLM Providers"
description: Switch between Ollama, OpenAI, Anthropic, and Google Gemini with a single header — no code changes needed. All caching, billing, and RBAC work identically across providers.
author: Memzent Team
category: announcement
tags: providers, openai, anthropic, gemini, ollama, multi-model
published_at: 2026-06-05
---

## The Problem with Provider Lock-in

Every LLM provider has different APIs, authentication, rate limits, and pricing. Teams end up:

- Writing separate integration code for each provider
- Managing multiple API keys and billing accounts
- Losing their caching layer when switching providers
- Rebuilding monitoring when a new model drops

## One API to Rule Them All

With Memzent's multi-provider support, you send one request and control the provider with a header:

```bash
# Use OpenAI
curl -X POST https://api.memzent.ai/v1/chat \
  -H "X-API-Key: memzent_YOUR_KEY" \
  -H "X-Memzent-Provider: openai" \
  -H "X-Memzent-Model: gpt-4o-mini" \
  -d '{"messages": [{"role": "user", "content": "Explain REST APIs"}]}'

# Switch to Anthropic — same code, just change headers
curl -X POST https://api.memzent.ai/v1/chat \
  -H "X-API-Key: memzent_YOUR_KEY" \
  -H "X-Memzent-Provider: anthropic" \
  -H "X-Memzent-Model: claude-sonnet-4-20250514" \
  -d '{"messages": [{"role": "user", "content": "Explain REST APIs"}]}'
```

## Supported Providers

| Provider | Default Model | Available Models |
|----------|--------------|-----------------|
| Ollama | llama3.2 | llama3.2, llama3, mistral, phi3, + locally installed |
| OpenAI | gpt-4o-mini | gpt-4o-mini, gpt-4, gpt-4-turbo |
| Anthropic | claude-sonnet-4-20250514 | claude-sonnet-4, claude-opus-4, claude-3-5-sonnet, claude-3-5-haiku |
| Gemini | gemini-2.5-flash | gemini-2.5-flash, gemini-2.5-pro, gemini-2.0-flash, gemini-1.5-pro |

## Model-Scoped Caching

Cache entries are isolated by model. When you ask "Explain REST APIs" with GPT-4 and then ask the same question with Claude — you get two separate answers cached separately. You'll never accidentally receive a GPT response when requesting Claude.

```
Cache key: org:{org}:s:{model}:{hash}
                         ↑
              Model is part of the key
```

## Dynamic Model Discovery

Don't know what models are available? Query the discovery endpoints:

```bash
# List all configured providers
curl https://api.memzent.ai/v1/providers \
  -H "X-API-Key: memzent_YOUR_KEY"

# List all available models (dynamically discovered)
curl https://api.memzent.ai/v1/models \
  -H "X-API-Key: memzent_YOUR_KEY"
```

For Ollama, models are discovered from your local instance. For cloud providers, models are queried from their APIs at startup.

## Provider Resolution Priority

When multiple override sources conflict, the priority is:

1. **Headers** — `X-Memzent-Provider` / `X-Memzent-Model`
2. **Request body** — `"provider": "..."` / `"model": "..."`
3. **Org settings** — Default model configured in dashboard
4. **Gateway default** — Configured at deploy time

## Unified Billing

Regardless of provider, billing flows through the same ledger:

- Token costs are calculated per-provider (different rates for GPT-4 vs Claude vs local Ollama)
- Cache hits get 90% discount across all providers
- Spend limits apply uniformly
- One balance, one invoice

## What This Enables

- **A/B testing models** — Same prompt to different providers, compare quality
- **Cost optimization** — Route simple queries to cheaper models, complex ones to premium
- **Resilience** — If OpenAI is down, switch to Anthropic with one header
- **Local development** — Use Ollama locally, OpenAI in production, same code

## Get Started

Update your integration to include the provider header, or set your org default in the [dashboard settings](/settings). All existing requests continue working with your current default provider.
