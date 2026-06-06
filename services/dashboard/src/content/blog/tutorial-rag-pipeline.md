---
title: "Tutorial: Building a RAG Pipeline with Memzent Tool Registry"
description: Step-by-step guide to registering tools, enabling semantic routing, and building a retrieval-augmented generation pipeline that automatically enriches LLM prompts.
author: Memzent Engineering
category: tutorial
tags: rag, tools, mcp, connectors, semantic-routing, tutorial
published_at: 2026-05-30
---

## What You'll Build

A complete RAG (Retrieval-Augmented Generation) pipeline where:

1. User asks a question
2. Memzent automatically identifies relevant tools
3. Tools fetch live data (docs, APIs, databases)
4. LLM receives enriched context and generates an accurate answer

No prompt engineering. No manual tool selection. Fully automatic.

## Prerequisites

- A Memzent API key ([get one here](/keys))
- One or more data sources you want the LLM to access

## Step 1: Register Your Tools

Register each data source as a tool with a **descriptive name and description** — Memzent uses these for semantic matching:

```bash
# Knowledge base search tool
curl -X POST https://api.memzent.ai/v1/tools/register \
  -H "X-API-Key: memzent_YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "docs_search",
    "description": "Search the product documentation and help articles for user-facing features, pricing, and how-to guides",
    "endpoint": "https://your-app.com/api/docs/search",
    "connector_type": "rest_api",
    "input_schema": {"query": "string"},
    "output_schema": {"results": "array"},
    "timeout_seconds": 5
  }'
```

```bash
# Customer database lookup
curl -X POST https://api.memzent.ai/v1/tools/register \
  -H "X-API-Key: memzent_YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "customer_lookup",
    "description": "Look up customer account details including plan, billing status, team size, and recent activity",
    "endpoint": "https://your-app.com/api/customers/lookup",
    "connector_type": "rest_api",
    "input_schema": {"customer_id": "string"},
    "output_schema": {"plan": "string", "status": "string"},
    "timeout_seconds": 3,
    "requires_auth": true
  }'
```

## Step 2: Understand Semantic Matching

When you register a tool, Memzent:

1. Embeds the tool's name + description into a 384-dim vector
2. Stores it in Qdrant's `tools_collection`
3. On every request, embeds the user's prompt and finds the most relevant tools
4. Tools scoring above the relevance threshold (default: 0.7) are selected

**Key insight**: Write descriptions from the **user's perspective**, not the developer's. Think "what would a user ask that should trigger this tool?"

## Step 3: Send Requests (Normal Chat API)

No changes to your chat calls — tool routing is automatic:

```bash
curl -X POST https://api.memzent.ai/v1/chat \
  -H "X-API-Key: memzent_YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {"role": "user", "content": "What features are included in the Pro plan?"}
    ]
  }'
```

Behind the scenes:
- Memzent embeds "What features are included in the Pro plan?"
- Finds `docs_search` tool is relevant (high similarity to "product documentation and help articles for user-facing features, pricing")
- Calls `docs_search` with query "Pro plan features"
- Passes the results to the LLM as context
- LLM generates an answer grounded in your actual documentation

## Step 4: Tool Chaining (Advanced)

For complex queries that need multiple data sources, Memzent supports automatic tool chaining:

```bash
curl -X POST https://api.memzent.ai/v1/chat \
  -H "X-API-Key: memzent_YOUR_KEY" \
  -d '{
    "messages": [
      {"role": "user", "content": "First look up customer acme-corp, then find docs about upgrading their plan"}
    ]
  }'
```

Keywords like "first", "then", "after" trigger the chain planner — Memzent executes tools sequentially, passing output from one to the next.

## Step 5: Monitor & Tune

### Check tool relevance scores

Subscribe to `tool_execution` webhook events to see which tools are being matched and their relevance scores.

### Adjust the threshold

If tools are firing too often (low relevance):

```bash
curl -X PUT https://api.memzent.ai/v1/settings/threshold \
  -H "X-API-Key: memzent_YOUR_KEY" \
  -d '{"similarity_threshold": 0.8}'
```

If tools aren't firing when they should, lower the threshold.

### Verify tool status

```bash
curl https://api.memzent.ai/v1/tools/status \
  -H "X-API-Key: memzent_YOUR_KEY"
```

## Tips for Better RAG

1. **Be specific in descriptions** — "Search product documentation for pricing, features, and how-to guides" beats "Search docs"
2. **Use natural language** — Match how users actually phrase questions
3. **Set appropriate timeouts** — Don't let a slow tool block the entire response
4. **Test with real queries** — Send actual user questions and verify the right tools are matched
5. **Monitor cache interaction** — Cached responses include tool results, so subsequent similar questions get instant answers

## Complete Example: Support Bot

```python
import requests

BASE = "https://api.memzent.ai"
HEADERS = {"X-API-Key": "memzent_YOUR_KEY", "Content-Type": "application/json"}

# Register tools
tools = [
    {"name": "faq_search", "description": "Search frequently asked questions", "endpoint": "https://app.com/api/faq"},
    {"name": "ticket_status", "description": "Check support ticket status", "endpoint": "https://app.com/api/tickets"},
    {"name": "billing_info", "description": "Get billing and subscription details", "endpoint": "https://app.com/api/billing"},
]

for tool in tools:
    requests.post(f"{BASE}/v1/tools/register", json={**tool, "connector_type": "rest_api"}, headers=HEADERS)

# Now just chat — tools are matched automatically
response = requests.post(f"{BASE}/v1/chat", json={
    "messages": [{"role": "user", "content": "What's the status of my support ticket?"}]
}, headers=HEADERS)

print(response.json()["text"])
# → "Based on your account, ticket #4521 is currently being reviewed by our team..."
```

The LLM response is grounded in live data from your `ticket_status` tool — no hallucination possible.
