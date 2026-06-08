# Memzent Go-Live Checklist

> Actionable, file-level tasks to move from **demo-ready** to **production-ready**.
> Reconciled with code audit — June 2026, branch `mani/code`.
> Status tracker: **[PROJECT_STATUS.md](./PROJECT_STATUS.md)**

**Estimated effort:** 3–4 focused weeks (1 engineer) for P0 + P1.

---

## How to use this checklist

- `[ ]` = not done
- `[x]` = done
- **Owner** = suggested role (Backend / DevOps / Frontend / Security)
- **Verify** = command or manual test to confirm completion

---

## Sprint 1 — Security & Secrets (P0)

> **Goal:** No credentials in git, no auth bypasses in production paths.
> **Block go-live until all items checked.**

### 1.1 Rotate exposed credentials

> **Manual action required** — secrets were removed from tracked files (June 2026 sprint 1), but may still exist in git history.

- [ ] **Rotate Supabase pooler password** (was in old `docker-compose.yml` comments)
- [ ] **Rotate OpenAI API key** (was in old compose comments)
- [ ] **Rotate Stripe test keys** (was in old compose / dashboard env)
- [ ] **Rotate agent API key** (was hardcoded in `test_flow.go` — now uses `MEMZENT_API_KEY` env)
- [ ] **Invalidate git history exposure** — consider `git filter-repo` or accept rotation as sufficient for test keys
- **Owner:** Security / DevOps
- **Verify:** `rg "sk-|memzent_[a-f0-9]{20}|postgresql://postgres\." --glob "*.{yml,go,ts,env*}"` returns no live secrets

### 1.2 Remove secrets from tracked files

| File | Action |
|------|--------|
| `docker-compose.yml` | Replace inline secrets with `${VAR}` references only; use `.env.example` for template |
| `services/gateway/scripts/test_flow.go` | Read `MEMZENT_API_KEY`, `MEMZENT_ORG_ID`, `JWT_SECRET`, `GATEWAY_URL` from `os.Getenv` |
| Create `.env.example` | Document all required env vars without real values |

- [x] Create `.env.example` at repo root
- [x] Add `.env` to `.gitignore` (already present)
- [x] Scrub `docker-compose.yml` — gateway uncommented, all services use `env_file: .env`
- [x] Refactor `test_flow.go` to use env vars
- **Owner:** Backend
- **Verify:** `go run scripts/test_flow.go` works with env vars set; no literal `memzent_` keys in repo

### 1.3 Remove production RBAC bypasses

| File | Line(s) | Change |
|------|---------|--------|
| `services/gateway/internal/auth/rbac.go` | 47–50 | Gate `admin-01` bypass behind `ENVIRONMENT != production` |
| `services/gateway/internal/auth/rbac.go` | 52–56 | Gate permissive `chat:execute` behind dev environment |
| `services/gateway/internal/auth/rbac_sqlmock_test.go` | — | `TestRBACClient_CheckPermission_ProductionStrict` |

- [x] Env-gate dev bypasses (`RBACClient.environment`)
- [x] Provision `chat:execute` on org create — `migrations/021_provision_chat_execute.sql`
- **Owner:** Backend
- **Verify:** With `ENVIRONMENT=production`, org without `chat:execute` in `org_tools` gets 403

### 1.4 Harden auth surface

| File | Action |
|------|--------|
| `services/gateway/main.go` | Restrict CORS: replace `*` with configurable `CORS_ALLOWED_ORIGINS` |
| `services/gateway/main.go` | Ensure `/generate-token` only registers when `ENABLE_DEV_TOKEN=true` (already done — verify unset in prod) |
| `services/gateway/main.go` | Static Supabase JWK seed only in non-production (or explicit `SUPABASE_STATIC_JWK`) |
| `services/gateway/internal/config/config.go` | `CORSAllowedOrigins`, production JWT/CORS validation |
| `README.md` | Warn: change default `JWT_SECRET` before any deployment |

- [x] CORS restriction implemented (`corsMiddleware`)
- [x] Static JWK seed gated for production
- [x] Gateway refuses default `JWT_SECRET` when `ENVIRONMENT=production`
- **Owner:** Backend
- **Verify:** Preflight from unauthorized origin blocked; JWT verification works via JWKS only in prod

### 1.5 Cache + RBAC ordering decision

| File | Action |
|------|--------|
| `services/gateway/internal/engine/engine.go` | **Option A:** Document org-scoped cache hit model |
| `AGENTS.md` | Execution flow documents cache-before-RBAC on HIT paths |

- [x] Decision: **Option A** — org+model-scoped cache keys; RBAC on miss only
- [x] Documented in `engine.go` comment block
- **Owner:** Backend + Security
- **Verify:** Threat model comment in `engine.Process` cache section

---

## Sprint 2 — Database & Deploy (P0)

### 2.1 Apply pending migrations

> Helper docs: `scripts/apply-pending-migrations.md` · verify: `scripts/verify-supabase-migrations.sql`

- [ ] Apply `migrations/020_api_key_rotation.sql` to Supabase production
- [ ] Apply `migrations/021_provision_chat_execute.sql` to Supabase production
- [ ] Verify columns exist: `expires_at`, `prev_key_hash`, `rotated_at`, `last_used_at` on `api_keys`
- [ ] Test key rotation end-to-end: Dashboard `keys/page.tsx` → grace window → old key rejected after 15 min
- **Owner:** Backend / DevOps
- **Verify:** Run `scripts/verify-supabase-migrations.sql` in Supabase SQL Editor

### 2.2 Fix docker-compose for one-command deploy

| File | Action |
|------|--------|
| `docker-compose.yml` | Gateway uses `Dockerfile.prod`, `env_file: .env`, healthcheck on `/healthz` |
| `docker-compose.yml` | Dashboard `GATEWAY_INTERNAL_URL` → `http://memzent-gateway:8080` inside compose network |
| `services/gateway/Dockerfile.prod` | Multi-stage build; copies root `migrations/` into image |
| `.env.example` | Documents all required vars |

- [x] Gateway service enabled with production Dockerfile
- [x] Secrets via `.env` only
- [x] README documents dashboard on port **3002**
- [ ] `docker compose up -d --build` — verify on your machine
- [ ] `curl http://localhost:8080/healthz` returns healthy
- [ ] Dashboard playground reaches gateway
- **Owner:** DevOps
- **Verify:** Fresh machine deploy from README steps only

### 2.3 Fix CI pipelines

| File | Action |
|------|--------|
| `.github/workflows/go.yml` | `services/gateway`, Go 1.25, `go test ./...` |
| `.github/workflows/rust.yml` | `cargo test --all-targets` in `services/router` |
| `.github/workflows/dashboard.yml` | `bun install`, lint, build |
| Root `go.mod` | Removed (stub deleted; module lives in `services/gateway`) |

- [x] CI workflow files added/updated
- [ ] Gateway tests pass in CI on PR to `main` (verify after push)
- [ ] Rust unit tests pass in CI
- [ ] Dashboard builds in CI
- **Owner:** DevOps
- **Verify:** Green checks on a test PR

---

## Sprint 3 — Docs & API Contract (P1)

### 3.1 Fix README API reference

| File | Section | Fix |
|------|---------|-----|
| `README.md` | POST `/v1/chat` request body | `messages[]` format + curl example |
| `README.md` | Service Access | Dashboard port **3002** |
| `README.md` | Getting Started | `.env.example` + migration docs |

- [x] README matches `engine.PromptRequest` in `services/gateway/internal/engine/engine.go`
- **Verify:** Copy-paste README curl example works against local gateway

### 3.2 Reconcile ARCHITECTURE.md

| File | Section | Fix |
|------|---------|-----|
| `ARCHITECTURE.md` | §1 Execution Sequence | 12-step flow aligned with `engine.Process` |
| `ARCHITECTURE.md` | Phase 1.a–4 | Honest completion status |
| `ARCHITECTURE.md` | Duplicate Phase 4 | Removed |

- [x] ARCHITECTURE.md matches PROJECT_STATUS.md
- **Owner:** Docs / Backend

### 3.3 Update AGENTS.md

| File | Change |
|------|--------|
| `AGENTS.md` §4 Execution Flow | 10-step flow + cache-before-RBAC note |
| `AGENTS.md` §5 links | Repo-relative paths |

- [x] AGENTS.md links resolve to project docs in this repo

### 3.4 Fix broken dashboard client

| File | Action |
|------|--------|
| `services/dashboard/src/lib/memzent-client.ts` | `chatMemzent()` — POST `/v1/chat` with `messages` + auth headers |

- [x] Client rewritten; `queryMemzent()` deprecated wrapper retained
- **Owner:** Frontend

---

## Sprint 4 — Product Honesty (P1–P2)

> **Goal:** No misleading product behavior in code or docs.
> **Code status:** ✅ Complete (June 2026). Remaining gaps: OpenAI/Anthropic native streaming, dashboard playground SSE.

### 4.1 Core tools — mock vs real

| File | Current | Action |
|------|---------|--------|
| `services/gateway/main.go` | Demo stubs return `[Demo]` prefixed messages | Documented as dev-only |

- [x] Decision documented in `PROJECT_STATUS.md`
- [x] Responses clearly labeled `[Demo]` — not misleading as live data

### 4.2 SSE streaming

| File | Action |
|------|--------|
| `internal/llm/provider.go` | `StreamingProvider` interface |
| `internal/llm/ollama.go` | Native Ollama NDJSON `GenerateStream` |
| `main.go` | SSE emits live tokens; cache/non-Ollama fallback word-chunks |

- [x] Ollama native streaming implemented
- [x] README documents provider differences
- [ ] OpenAI/Anthropic native streaming (future)

### 4.3 Tool chaining trigger

| File | Action |
|------|--------|
| `engine/engine.go` | `chain: true` flag + router confidence ≥ 0.65; keywords as fallback |

- [x] `chain` field on `PromptRequest`
- [x] README documents chaining behavior

### 4.4 Connector completeness

| File | Action |
|------|--------|
| `tools/connector_types.go` | Registration limited to `core`, `mcp`, `rest`, `sql` |
| `tools/handlers.go` | Rejects `graphql`, `webhook`, `grpc` at register time |

- [x] Unsupported types rejected with clear error message

### 4.5 Dead code cleanup

| Path | Status |
|------|--------|
| `internal/offline/`, `workflow/`, `prewarmer/`, `featureflags/`, `notifications/` | Not wired in `main.go` — removed from workspace (never imported) |

- [x] Decision: delete unwired experimental packages
- [x] No imports from production `main.go`
- **Owner:** Backend
- **Verify:** `go build .` in `services/gateway`

---

## Sprint 5 — Testing & Observability (P2)

> **Goal:** Automated test coverage for critical paths; observability sign-off documented.
> **Code status:** ✅ Complete (June 2026). Manual observability items remain for prod deploy.

### 5.1 Integration test suite

| File | Action |
|------|--------|
| `tests/integration/cache_test.go` | Valkey round-trip + engine miss→hit with stub LLM |
| `docker-compose.test.yml` | Valkey + Qdrant + Router for local full-stack tests |
| `scripts/test_flow.go` | Manual load tool — env vars only (README) |

- [x] CI integration job runs Valkey tests (`.github/workflows/go.yml`)
- [ ] Optional: run router gRPC integration locally (`cargo test --test integration_tests -- --ignored`)

### 5.2 Router testability

| File | Action |
|------|--------|
| `internal/router/interface.go` | `SemanticRouter` interface + test constructor |
| `internal/engine/engine.go` | Accepts `SemanticRouter` + `cache.Store` |
| `internal/engine/process_test.go` | Cache hit, RBAC deny, miss→hit, rate limit, billing |

- [x] `go test ./internal/engine/...` covers cache hit, RBAC deny, rate limit paths

### 5.3 Rust integration tests

| File | Action |
|------|--------|
| `tests/integration_tests.rs` | gRPC SelectTools (`#[ignore]` — needs compose stack) |
| `src/lib.rs` | Pure helpers: hash, compress, thresholds, tool UUID |

- [x] Unit tests import `memzent_router` lib (no mirrored logic)
- **Owner:** Backend (Rust)

### 5.4 Dashboard tests

| File | Action |
|------|--------|
| `vitest.config.ts` + `login/page.test.tsx` | Smoke: login page renders |
| `package.json` | `bun run test` in CI |

- [x] Vitest smoke test in CI
- [ ] Playwright E2E for keys page (staging — future)

### 5.5 Production observability

See **[scripts/observability-checklist.md](./scripts/observability-checklist.md)**.

- [ ] Confirm `/metrics` scraped in prod deployment config
- [ ] Confirm audit log retention job running (`main.go` — 30 day retention)
- [ ] Alert on: gateway `readyz` fail, cache hit rate drop, Stripe webhook errors
- [ ] Qdrant snapshot scheduler verified (`SNAPSHOT_INTERVAL_HOURS=24` in compose)
- **Owner:** DevOps

---

## Go-Live Sign-Off Criteria

All must be true before calling Memzent production-ready:

| # | Criterion | Verified by |
|---|-----------|-------------|
| 1 | No secrets in git (`rg` scan clean) | Security |
| 2 | `ENVIRONMENT=production` has no RBAC bypasses | Backend test |
| 3 | Migration 020 applied to Supabase | SQL query |
| 4 | `docker compose up` starts full stack including gateway | DevOps |
| 5 | CI green: gateway tests + rust tests + dashboard build | GitHub Actions |
| 6 | README curl example works copy-paste | Manual |
| 7 | API key rotation tested end-to-end | Manual / E2E |
| 8 | CORS restricted to known origins | Manual |
| 9 | PROJECT_STATUS.md and ARCHITECTURE.md agree | Docs review |
| 10 | `test_flow.go` uses env vars only | Code review |

---

## Quick Reference — Key Files

```
services/gateway/
├── main.go                          # HTTP routes, wiring, CORS, SSE
├── internal/engine/engine.go        # Core pipeline (cache, RBAC, tools, LLM)
├── internal/auth/rbac.go            # RBAC bypasses — fix here
├── internal/auth/middleware.go      # JWT + API key auth
├── internal/tools/handlers.go       # Tool registry API + health probes
├── internal/connectors/             # Core, MCP, REST, SQL
├── scripts/maketoken/main.go        # Dev JWT CLI
└── scripts/test_flow.go             # Load test — sanitize secrets

services/router/src/main.rs          # Embeddings, Qdrant, snapshots
services/dashboard/                  # Next.js admin UI
migrations/020_api_key_rotation.sql  # Apply to Supabase
docker-compose.yml                   # Uncomment gateway, remove secrets
.github/workflows/go.yml             # Fix to test services/gateway
PROJECT_STATUS.md                    # Honest feature matrix
```

---

*Update checkboxes as tasks complete. Link PRs next to items when merging.*
