---
title: "Announcing Sessions & Semantic Memory: Your LLM Now Remembers"
description: Persistent conversation sessions and automatic long-term memory extraction — your LLM builds knowledge about users across every interaction.
author: Memzent Team
category: announcement
tags: sessions, memory, qdrant, context, personalization
published_at: 2026-05-25
---

## The Stateless Problem

LLMs are stateless. Every request starts from zero. Your users have to repeat themselves every conversation:

> "I'm using PostgreSQL 16..."
> "As I mentioned before, I'm on the Pro plan..."
> "Remember, I prefer Python examples..."

This wastes tokens, frustrates users, and produces generic responses.

## Sessions: Short-Term Memory

Create a session to maintain conversation context:

```bash
# Create a session
curl -X POST https://api.memzent.ai/v1/sessions \
  -H "X-API-Key: memzent_YOUR_KEY" \
  -d '{"title": "Debugging database connection"}'

# Response: {"id": "sess_a1b2c3...", "session_id": "sess_a1b2c3..."}
```

Now attach it to your chat requests:

```bash
curl -X POST https://api.memzent.ai/v1/chat \
  -H "X-API-Key: memzent_YOUR_KEY" \
  -d '{
    "messages": [{"role": "user", "content": "I am using PostgreSQL 16 with pgvector"}],
    "session_id": "sess_a1b2c3..."
  }'

# Later in the same session:
curl -X POST https://api.memzent.ai/v1/chat \
  -H "X-API-Key: memzent_YOUR_KEY" \
  -d '{
    "messages": [{"role": "user", "content": "How do I add an index?"}],
    "session_id": "sess_a1b2c3..."
  }'
```

The second request automatically includes the previous messages as context. The LLM knows you're asking about PostgreSQL 16 pgvector indexes — not generic indexing.

## Semantic Memory: Long-Term Learning

This is where it gets powerful. After every exchange, Memzent **automatically extracts permanent facts** and stores them as vectors:

![Memory Architecture](/blog/memory-architecture.png)

### What Gets Extracted

- "User uses PostgreSQL 16 with pgvector" ✓
- "User prefers Python examples" ✓
- "User is on Pro plan with 5 team members" ✓
- "User's timezone is EST" ✓

### What Gets Ignored

- Casual dialogue ("thanks!", "got it")
- Temporal state ("I'm debugging right now")
- Generic questions ("what is REST?")

### How Recall Works

On every request, Memzent queries stored memories with a relevance threshold of 0.65:

```
User asks: "Show me how to create an embedding column"

Memory recall finds:
→ "User uses PostgreSQL 16 with pgvector" (score: 0.82)
→ "User prefers Python examples" (score: 0.71)

These are injected into LLM context automatically.
```

The LLM now generates a PostgreSQL 16 + pgvector specific answer with Python code — without the user specifying any of that.

## Cross-Session Persistence

Semantic memories persist **across all future sessions**. A fact learned in January is recalled in June — if it's relevant to the current question.

```
January session: "I deploy on AWS us-east-1 with ECS"
                  ↓ extracted and stored as vector

June session: "Help me set up a new service"
              ↓ memory recall finds AWS/ECS context
              ↓ LLM generates AWS ECS-specific answer
```

## Privacy & Deletion

- Memories are scoped by **org + user** — never cross-contaminated
- Deleting a session does NOT delete semantic memories (they're persistent knowledge)
- Users can request memory deletion via admin API (coming soon)

## Technical Architecture

```
User Message
    ↓
[Session Manager] → Append to history, load previous messages
    ↓
[Memory Manager] → Query Qdrant for relevant facts (threshold 0.65)
    ↓
[LLM Synthesis] → Generate response with full context
    ↓
[Fact Extraction] → Background: extract new facts from exchange
    ↓
[Qdrant Storage] → Vectorize and store permanent facts
```

## API Reference

```bash
# Create session
POST /v1/sessions
Body: {"title": "optional description"}

# Get session history
GET /v1/sessions/{id}/messages

# Delete session
DELETE /v1/sessions/{id}

# Use in chat (just add session_id)
POST /v1/chat
Body: {"messages": [...], "session_id": "sess_..."}
```

## What's Next

- **Memory management UI** in the dashboard
- **Explicit memory injection** — tell Memzent to remember specific facts
- **Memory scoping** — org-wide vs user-private memories
- **Forgetting** — API to delete specific memories on request
