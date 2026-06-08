---
title: "Announcing Webhooks: Real-Time Event Notifications from Your AI Pipeline"
description: Subscribe to cache hits, tool executions, rate limits, and more — delivered to your endpoints with HMAC-signed payloads for full observability.
author: Memzent Team
category: announcement
tags: webhooks, events, monitoring, observability, notifications
published_at: 2026-06-02
---

## Full Observability for Your AI Pipeline

Today we're launching **Webhooks** — real-time HTTP notifications for events happening inside your Memzent gateway. Get notified the instant something important happens, without polling.

## Available Events

| Event | Fires When |
|-------|-----------|
| `cache_hit` | A prompt matches semantic cache |
| `tool_execution` | A registered tool is invoked |
| `rate_limit` | A request is rate-limited |
| `key_rotated` | An API key is rotated |
| `tool_registered` | A new tool is added to the registry |
| `session_created` | A new conversation session starts |

## Quick Setup

Create a webhook subscription in one API call:

```bash
curl -X POST https://api.memzent.ai/v1/webhooks \
  -H "X-API-Key: memzent_YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-app.com/webhooks/memzent",
    "events": ["cache_hit", "tool_execution", "rate_limit"],
    "description": "Production monitoring"
  }'
```

Response includes a **signing secret** (shown only once):

```json
{
  "id": "wh_8f2a4b6c...",
  "secret": "whsec_a1b2c3d4e5f6...",
  "events": ["cache_hit", "tool_execution", "rate_limit"],
  "enabled": true
}
```

## Event Payload Structure

Every webhook delivery follows the same envelope:

```json
{
  "id": "evt_9x8y7z...",
  "type": "cache_hit",
  "org_id": "5127e445-bb64-4057...",
  "timestamp": "2026-06-06T05:22:11Z",
  "data": {
    "query": "what is quantum computing",
    "score": 0.97,
    "latency_ms": 12,
    "model": "gpt-4o-mini"
  }
}
```

## Cryptographic Verification

Every delivery includes an HMAC-SHA256 signature in the `X-Memzent-Signature` header. Always verify before processing:

```python
import hmac
import hashlib

def verify_webhook(payload: bytes, signature: str, secret: str) -> bool:
    expected = hmac.new(
        secret.encode(), payload, hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(f"sha256={expected}", signature)
```

## Delivery Headers

| Header | Purpose |
|--------|---------|
| `X-Memzent-Signature` | HMAC-SHA256 payload signature |
| `X-Memzent-Event` | Event type for routing |
| `X-Memzent-Delivery` | Unique ID for idempotency |
| `User-Agent` | `Memzent-Webhook/1.0` |

## Use Cases

### Cost Monitoring Alerts

Subscribe to `cache_hit` events to track savings in real-time. Pipe to Slack when daily savings exceed a threshold.

### Security Auditing

Subscribe to `key_rotated` and `rate_limit` events. Alert your security team when keys are rotated or unusual rate limiting occurs.

### Tool Performance Tracking

Subscribe to `tool_execution` events to monitor tool latency and success rates. Build custom dashboards or send to Datadog/Grafana.

### Session Analytics

Subscribe to `session_created` to track conversation volume and user engagement patterns.

## Managing Webhooks

```bash
# List all webhooks
curl https://api.memzent.ai/v1/webhooks \
  -H "X-API-Key: memzent_YOUR_KEY"

# Update events or URL
curl -X PUT https://api.memzent.ai/v1/webhooks/wh_8f2a... \
  -H "X-API-Key: memzent_YOUR_KEY" \
  -d '{"events": ["cache_hit", "rate_limit"], "enabled": true}'

# Delete a webhook
curl -X DELETE https://api.memzent.ai/v1/webhooks/wh_8f2a... \
  -H "X-API-Key: memzent_YOUR_KEY"

# See available event types
curl https://api.memzent.ai/v1/webhooks/event-types \
  -H "X-API-Key: memzent_YOUR_KEY"
```

## What's Next

- **Retry with backoff** — Failed deliveries will retry up to 5 times
- **Delivery logs** — View delivery history and debug failures in the dashboard
- **Custom filters** — Subscribe to events matching specific conditions (e.g., only cache hits above 0.98 similarity)
