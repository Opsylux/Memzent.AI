---
title: "Use Case: Building a Customer Support Bot with Semantic Memory"
description: How a SaaS company used Memzent to build a support bot that remembers customer context across sessions and saves 75% on LLM costs.
author: Memzent Team
category: use-case
tags: support-bot, sessions, memory, cost-savings
published_at: 2026-06-04
---

## The Challenge

A mid-size SaaS company (50K users) was running a customer support chatbot powered by GPT-4. Their pain points:

- **$12,000/month** in OpenAI API costs
- **No memory** — each conversation started fresh, requiring customers to repeat context
- **Inconsistent answers** — the same question got different (sometimes contradictory) responses
- **Slow responses** — 3-5 seconds per message for complex queries

## The Solution: Memzent as a Semantic Proxy

By placing Memzent between their app and OpenAI, they gained:

### 1. Instant Answers for Repeat Questions

Support queries are highly repetitive. "How do I reset my password?" appears hundreds of times per day with slight variations:

- "How to reset password"
- "I forgot my password, help"
- "password reset instructions"
- "Can't login, need password change"

All of these semantically match — Memzent's Layer 2 cache catches them all after the first answer is generated.

### 2. Persistent Customer Context

Using Sessions + Semantic Memory:

```bash
# Create a session per customer conversation
curl -X POST https://api.memzent.ai/v1/sessions \
  -H "X-API-Key: $KEY" \
  -d '{"title": "Customer: user_12345"}'

# All subsequent messages include session_id
curl -X POST https://api.memzent.ai/v1/chat \
  -H "X-API-Key: $KEY" \
  -d '{
    "messages": [{"role": "user", "content": "I use the Pro plan with 5 team members"}],
    "session_id": "sess_abc123"
  }'
```

Memzent automatically extracts facts like "Customer uses Pro plan with 5 team members" and recalls them in future sessions — even months later.

### 3. Tool-Enriched Responses

They registered their internal APIs as Memzent tools:

- **Account lookup** — fetch customer plan, billing, team size
- **Knowledge base search** — search help docs for relevant articles
- **Ticket status** — check open support tickets

When a customer asks "What's the status of my refund?", Memzent automatically routes to the ticket-status tool, fetches live data, and includes it in the LLM prompt.

## Results After 30 Days

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| Monthly LLM cost | $12,000 | $3,100 | -74% |
| Avg response time | 3.8s | 0.4s | -89% |
| Cache hit rate | 0% | 58% | — |
| Customer satisfaction | 3.2/5 | 4.6/5 | +44% |

## Key Takeaways

1. **Most support queries are repeated** — semantic caching is a natural fit
2. **Memory reduces token usage** — no need to re-send full context every time
3. **Tool integration eliminates hallucinations** — live data beats training data
4. **Zero code changes needed** — just point your API calls through Memzent
