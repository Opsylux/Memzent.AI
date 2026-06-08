# Self-Hosting Memzent

This guide covers deploying Memzent on your own infrastructure.

## Prerequisites

- Docker 24+ and Docker Compose v2
- 4GB RAM minimum (8GB recommended)
- Ports: 8080 (Gateway), 50051 (Router), 6333 (Qdrant), 6379 (Valkey)

## Quick Start

```bash
git clone https://github.com/Opsylux/Memzent.AI.git
cd Memzent.AI
docker-compose up -d
```

This starts:
- **Gateway** (Go) — `:8080` — HTTP API, auth, caching, LLM orchestration
- **Router** (Rust) — `:50051` — Semantic embeddings, vector search, tool matching
- **Qdrant** — `:6333` — Vector database for semantic cache + memory
- **Valkey** — `:6379` — Fast cache layer (Redis-compatible)

## Configuration

All configuration is via environment variables. Set them in a `.env` file or pass to Docker:

### Required

| Variable | Description |
|----------|-------------|
| `POSTGRES_URL` | PostgreSQL connection string (for RBAC, sessions, billing) |

### LLM Providers (at least one required)

| Variable | Description |
|----------|-------------|
| `OLLAMA_URL` | Ollama endpoint (default: `http://host.docker.internal:11434`) |
| `OPENAI_API_KEY` | OpenAI API key |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `GEMINI_API_KEY` | Google Gemini API key |

### Optional Tuning

| Variable | Default | Description |
|----------|---------|-------------|
| `VALKEY_URL` | `redis://localhost:6379` | Valkey/Redis endpoint |
| `ROUTER_URL` | `router:50051` | Rust Router gRPC address |
| `TOOL_RELEVANCE_THRESHOLD` | `0.7` | Minimum similarity score for tool matching |
| `ENVIRONMENT` | `development` | Set to `production` for strict CORS |
| `JWT_SECRET` | — | Secret for JWT signing |
| `RATE_LIMIT_FREE` | `10` | Requests/min for free tier |
| `RATE_LIMIT_PRO` | `100` | Requests/min for pro tier |

## Database Setup

Memzent uses PostgreSQL for persistent storage (RBAC, sessions, billing, audit logs).

```bash
# Apply migrations
psql $POSTGRES_URL < migrations/001_initial.sql
psql $POSTGRES_URL < migrations/002_sessions.sql
# ... apply all numbered migrations in order
```

Or use the provided migration script:
```bash
for f in migrations/*.sql; do psql $POSTGRES_URL < "$f"; done
```

## Running Without Docker

### Gateway (Go)

```bash
cd services/gateway
go build -o memzent-gateway .
POSTGRES_URL="..." VALKEY_URL="..." ROUTER_URL="localhost:50051" ./memzent-gateway
```

### Router (Rust)

```bash
cd services/router
cargo build --release
QDRANT_URL="http://localhost:6333" ./target/release/memzent-router
```

### Qdrant

Follow [Qdrant's installation guide](https://qdrant.tech/documentation/quick-start/).

### Valkey

```bash
docker run -d -p 6379:6379 valkey/valkey:latest
```

## Production Deployment

For production, we recommend:

1. **Reverse proxy** (nginx/Caddy) in front of the Gateway for TLS termination
2. **Managed PostgreSQL** (Supabase, AWS RDS, etc.)
3. **Persistent volumes** for Qdrant data
4. **Health checks**: `GET /health` on the Gateway returns `200 OK`
5. **Resource limits**: Gateway ~512MB RAM, Router ~256MB RAM, Qdrant ~2GB RAM

### Docker Compose (Production)

```yaml
# Override for production
services:
  gateway:
    environment:
      - ENVIRONMENT=production
      - POSTGRES_URL=${POSTGRES_URL}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    deploy:
      resources:
        limits:
          memory: 512M
  router:
    deploy:
      resources:
        limits:
          memory: 256M
  qdrant:
    volumes:
      - qdrant_data:/qdrant/storage
    deploy:
      resources:
        limits:
          memory: 2G
```

## Upgrading

```bash
git pull origin main
docker-compose up -d --build
# Apply any new migrations
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Gateway can't reach Router | Check `ROUTER_URL` env var, ensure Router container is healthy |
| Cache misses everything | Verify Valkey is running, check `VALKEY_URL` |
| Semantic search not working | Ensure Qdrant is running and collections are initialized |
| Auth failures | Check `JWT_SECRET` matches between services |

## Support

- **Community**: [GitHub Discussions](https://github.com/Opsylux/Memzent.AI/discussions)
- **Discord**: [discord.gg/memzent](https://discord.gg/memzent)
- **Managed hosting**: [app.memzent.ai](https://app.memzent.ai) (we run it for you)
