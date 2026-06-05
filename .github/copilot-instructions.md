# Memzent — Copilot Instructions

## What is Memzent?

An Intelligent Semantic Proxy that intercepts traffic between clients, MCP tools, and LLM providers. It minimizes LLM latency and maximizes token ROI by semantically caching and routing prompts before hitting expensive LLMs. Multi-tenant with org-level isolation.

## Architecture (4 services)

```
Client → Go Gateway (:8080) → Rust Router (:50051/gRPC) → Qdrant (vectors)
              ↕                        ↕
         Valkey (cache)           Qdrant (tools + memory)
              ↕
         Postgres (RBAC, sessions, billing, audit, webhooks)
```

| Service | Path | Language | Role |
|---------|------|----------|------|
| Gateway | `services/gateway/` | Go 1.25 | HTTP, Auth, Caching, LLM orchestration, MCP client |
| Router | `services/router/` | Rust (tonic/tokio) | gRPC server: embeddings, similarity, tool matching |
| Dashboard | `services/dashboard/` | Next.js 16 (React 19) | Admin UI, App Router only |
| Website | `website/` | Vite + React | Marketing site |

## Strict Service Boundaries

- **Go Gateway**: All HTTP, JWT auth, RBAC, Valkey caching, MCP tools, LLM API calls. **Never** do vector math or embeddings in Go.
- **Rust Router**: All vector embeddings (`all-MiniLM-L6-v2` via FastEmbed), similarity scoring, Qdrant queries. **Never** add business logic, auth, or HTTP here. It is a pure gRPC service defined by `/proto/router.proto`.
- **Dashboard**: Next.js App Router (`src/app/`) only. **Never** use `pages/` directory. Tailwind v4 tokens in `@theme inline` in `globals.css` — no `tailwind.config.js`.

## Build & Run Commands

```bash
# Full stack (Docker)
make up              # docker-compose up -d --build
make down            # stop all
make logs            # follow gateway + router logs

# Gateway (Go) — run from services/gateway/
go build ./...       # compile check
go test ./...        # all tests
go test ./internal/engine/ -run TestProcess  # single test

# Router (Rust) — run from services/router/
cargo check          # fast compile check
cargo test           # all tests
cargo run            # start gRPC server

# Dashboard — run from services/dashboard/
npm run dev          # dev server
npm run build        # production build (Turbopack)
npx eslint .        # lint

# Website — run from website/
npm run dev
npm run build

# Integration test (requires running stack)
make test-flow       # 20-worker load test against /v1/chat

# Proto regeneration (after editing proto/router.proto)
make gen-proto
```

## Engine Execution Flow

Every request through `engine.Process()` follows this mandatory sequence:

1. **Rate Limiting** — Distributed via Valkey (org-level + per-user role-proportional)
2. **Permission Check** — Viewer role blocked from execution
3. **Billing Pre-check** — Token balance validation
4. **Cache Check** — 3-stage: Literal hash → Canonical hash → Semantic similarity (Qdrant)
5. **Session Memory** — Append user message, load history
6. **Semantic Routing** — gRPC to Rust for tool matching + memory recall
7. **Tool Execution** — Fire matched MCP/connector tools
8. **LLM Synthesis** — Pass enriched context to provider (Ollama/OpenAI/Anthropic/Gemini)
9. **Cache Set** — Store response in Valkey + Postgres durable cache
10. **Webhook Events** — Emit notifications to subscribed webhooks

## Key Conventions

### Go Gateway
- All business logic in `internal/` — never export packages
- Use `valkey-go` for caching (not standard Redis clients)
- All `/v1/*` endpoints must use `auth.UnifiedAuthMiddleware`
- Always scope queries by `org_id` from request context
- Use `context.Value("org_id")`, `context.Value("user_id")`, `context.Value("user_role")` for auth context
- API responses must return `[]` not `null` for empty arrays (JSON contract)

### Rust Router
- Proto changes require updating `/proto/router.proto` first, then `cargo build` regenerates stubs
- FastEmbed model: `all-MiniLM-L6-v2` (384-dim vectors)
- Qdrant collections: `tools_collection` (tool vectors), `memory_collection` (semantic memory)

### Dashboard
- Server actions in `src/app/actions.ts` call the gateway via `gatewayHeaders()` helper
- `@/*` path alias maps to `./src/*`
- Use `stat-card`, `glass`, `neural-bg` CSS utilities for card components
- Supabase handles auth — org membership via `members` table

### Communication Protocols
- Go ↔ Rust: gRPC only (`/proto/router.proto`)
- Gateway ↔ Tools: Model Context Protocol (MCP)
- Gateway ↔ Cache: Valkey via `valkey-go`
- Dashboard ↔ Gateway: REST (`/v1/*` endpoints) with JWT or X-API-Key auth

## Environment Variables (Gateway)

| Variable | Default | Purpose |
|----------|---------|---------|
| `VALKEY_URL` | `http://localhost:6379` | Cache layer |
| `ROUTER_URL` | `router:50051` | Rust gRPC address |
| `POSTGRES_URL` | `postgres://...` | Supabase Postgres |
| `OLLAMA_URL` | `http://host.docker.internal:11434` | Local LLM |
| `OPENAI_API_KEY` | — | OpenAI provider |
| `ANTHROPIC_API_KEY` | — | Anthropic provider |
| `GEMINI_API_KEY` | — | Google Gemini provider |
| `TOOL_RELEVANCE_THRESHOLD` | `0.7` | Min similarity score for tool matching |
| `ENVIRONMENT` | `development` | Controls CORS, admin bypass |
| `MEMZENT_DEV_ADMIN_BYPASS` | `false` | Skip RBAC in dev |

## Feature Implementation Checklist

When adding a new feature that touches the routing pipeline:
1. Update `/proto/router.proto` if gRPC changes needed → `make gen-proto`
2. Implement in Rust Router (`services/router/src/handlers.rs`)
3. Update Go gRPC client wrapper (`services/gateway/internal/router/`)
4. Integrate into `engine.Process()` following the execution flow order
5. Add dashboard UI page + server action
6. Apply any new migrations to Supabase

## Migrations

SQL migrations live in `/migrations/` (numbered `001_` through `023_`). Apply via Supabase SQL Editor or CLI (`supabase db push`).
