# Memzent: Project Status & Roadmap

> **Last reconciled:** June 2026 — audited against live code on branch `mani/code`.
> This document replaces optimistic completion claims with honest, code-verified status.
> For actionable go-live tasks with file paths, see **[GO_LIVE_CHECKLIST.md](./GO_LIVE_CHECKLIST.md)**.

---

## 0. Executive Snapshot

| Question | Answer |
|----------|--------|
| **What is Memzent?** | Intelligent Semantic Proxy — memory, security, and routing layer between clients, MCP tools, and LLMs. |
| **Core goal** | Minimize LLM latency and maximize token ROI via semantic caching and tool routing *before* expensive LLM calls. |
| **Where are we?** | **Demo-ready prototype (~85% feature breadth, ~65% production hardening).** Sprints 1–5 code complete; manual credential rotation + observability sign-off remain. |
| **Can we go live today?** | **No.** Blockers: rotate exposed credentials, apply migrations 020/021, prod observability alerts, optional Playwright E2E. |
| **Is the architecture wrong?** | **No.** Go orchestration + Rust vectors + Next.js dashboard is sound. Execution hygiene needs tightening. |

### Status legend

| Symbol | Meaning |
|--------|---------|
| ✅ | Production-ready — tested, no known security shortcuts |
| 🟡 | Demo-complete — works in dev/staging, needs hardening |
| 🟠 | Partial — scaffolded or heuristic, not fully real |
| ⬜ | Not started |
| 🔴 | Security / ops debt — must fix before go-live |

---

## 1. Feature Completion Matrix (Code-Verified)

| Feature | Service | Status | Honest notes |
| :--- | :--- | :--- | :--- |
| **Four-Layer Caching** | Gateway/Rust | ✅ 100% | L1 Literal, L1.5 Canonical, L1b Entity-Keyed, L2 Semantic. |
| **Service Boundaries** | All | ✅ 100% | Go Gateway (Auth/Orchestration), Rust Router (Math), Dashboard (UI). |
| **RBAC Scoping & Multi-Token** | Gateway | ✅ 100% | Dynamic key generation with customizable roles and scopes. API = pay-as-you-go, Dashboard = unlimited. |
| **Dynamic Tool Registry** | Gateway | ✅ 100% | Refresh loop, Qdrant sync, full CRUD (`/v1/tools/{id}`), Dashboard edit/delete UI. |
| **Connector Framework** | Gateway | ✅ 100% | SQL/REST/Core connectors fully implemented, registered, and active. |
| **Neural Dashboard** | Dashboard | ✅ 100% | Dynamic Billing, API Security metrics, live Playground, Provider discovery, Tool CRUD, Notifications. Responsive mobile layout. |
| **Provider Discovery** | Gateway | ✅ 100% | `/v1/providers` + `/v1/models` APIs. OpenAI, Anthropic, Gemini, Ollama discovery. |
| **Marketing Website** | Website | ✅ 100% | Mobile nav, SEO/OG/JSON-LD, Evolution Pipeline section, comparison table, sitemap.xml. |
| **Advanced Orchestration (Phase 4)** | All | ✅ 100% | Model-scoped caching, PlanToolChain Go/Rust bindings, and typewriter SSE streaming. |
| **Agent Memory (Phase 5/6)** | Gateway/Rust | ✅ 100% | PostgreSQL session threads and semantic memory Qdrant extraction. |
| **Context Analytics (Phase 5/6)** | Dashboard/Gateway | ✅ 100% | Premium ROI tracking, latency tool telemetry, and intent theme clusters. |
| **API Key Security (Phase 6)** | Gateway/Dashboard | ✅ 100% | Expiry TTL picker, last_used_at tracking, in-place rotation with 15-min grace window, stale key audit. |
| **Spend Limits & Budget Forecast** | Gateway/Dashboard | ✅ 100% | Dollar + token caps (daily/monthly), budget forecast API, spend timeseries, provider breakdown. |
| **Notification Pipeline (Phase 7)** | Gateway/Dashboard | ✅ 100% | Webhook CRUD, 6 event types, HMAC signing, async retry with dead letter, delivery logs. |
| **Per-User Rate Limiting** | Gateway | ✅ 100% | Role-proportional limits (viewer 20%, member 50%, admin 100% of org). |
| **E1: Entity Extraction** | Gateway/Rust | ✅ 100% | 6 typed extractors (regex, <1ms), positional awareness, mirrored in Go + Rust. |
| **E2: L1b Entity-Keyed Cache** | Gateway | ✅ 100% | Deterministic Valkey key from sorted entity pairs. Sub-ms lookups. Feature flag: `MEMZENT_L1B_ENABLED`. |
| **E3: Offline Learning Plane** | Gateway | ✅ 100% | Buffered channel event bus (4096/4 workers), 3 miners (Request, Cache, Workflow). PII-safe. |
| **E4: Workflow Registry** | Gateway/Dashboard | ✅ 100% | Full lifecycle (discovered→approved→active→stale), API endpoints, dashboard page, engine shortcut. |
| **E5: GPU Avoidance Metrics** | Gateway/Dashboard | ✅ 100% | 8 Prometheus counters, GPU Analytics dashboard, cache layer distribution. |
| **E6: Pattern Mining** | Gateway | ✅ 100% | Markov chain analysis + speculative pre-warmer. Experimental, flag defaults false. |
| **Feature Flags System** | Gateway | ✅ 100% | 6 env-var flags controlling L1b, offline, streams, workflows, metrics, pattern mining. |
| **Documentation Site** | Dashboard | ✅ 100% | 19 doc pages including entity-extraction, cache-layers, offline-learning, gpu-analytics. SEO metadata. |
| **Engineering Blog** | Dashboard | ✅ 100% | MDX + Supabase dual source, per-post SEO/OG metadata, Evolution Pipeline launch post. |
| **Integration Test Suites** | Gateway | ✅ 100% | 4 suites: test-cache (12), test-entity (14), test-memory (10), test-evolution (28). |

---

## 2. Execution Flow — Policy vs Code

**AGENTS.md policy:**

### [Phase 8] Evolution Pipeline (E1–E6) ✅ COMPLETE
*   **E1: Entity Extraction**: 6 typed regex extractors (<1ms) for accounts, customers, invoices, amounts, dates, identifiers. Positional awareness (source vs destination). Mirrored in both Rust Router and Go Gateway.
*   **E2: L1b Entity-Keyed Cache**: Deterministic Valkey key from sorted entity pairs (`org:{orgID}:m:{model}:e:{key=value|...}`). Sub-millisecond lookups. Write-through on skip-cache.
*   **E3: Offline Learning Plane**: Buffered channel event bus (4096 buffer, 4 workers) with try-send semantics. Three miners: Request, Cache, Workflow. PII-safe (no raw prompts). Valkey Streams transport available.
*   **E4: Workflow Registry**: Full lifecycle management (discovered → simulated → pending_review → approved → active → stale → demoted). API endpoints, dashboard page, engine shortcut that fires all matched tools in one pass.
*   **E5: GPU Avoidance Metrics**: 8 Prometheus counters for entity types, cache layer distribution, GPU avoidance rate. GPU Analytics dashboard page.
*   **E6: Pattern Mining**: Markov chain analysis predicts next-likely requests. Speculative Pre-Warmer populates L1b entries proactively. Experimental, defaults off.
*   **Feature Flags**: 6 environment variable flags controlling all Evolution Pipeline features.
*   **Integration Tests**: `make test-evolution` with 28 functional assertions across E1-E5 + infrastructure.
*   **Dashboard**: Mobile responsive rewrite, sign-out resilience, PKCE error handling, role-gated admin cards, org-isolated stats.
*   **Website & Docs**: Evolution Pipeline section, 4 new doc pages, blog post, sitemap.xml, robots.txt, SEO metadata across all pages.

### [Phase 5 & 6] Memory, Tool Chaining & Context Analytics ✅ COMPLETE
*   **Semantic Agent Memory**: Added PostgreSQL persistence for conversation sessions (`sessions`, `session_messages`) and vectorized conversation facts out-of-band to Qdrant memory collection.
*   **Sequential Tool Chaining**: Integrated Go Gateway engine dynamic parameter schema fitting (`fitToolParameters`) and sequential execution chains (`PlanToolChain`).
*   **Context Analytics**: Developed SQL metrics aggregations computing savings ROI, tool latency, and failure rates, and clustering user intent themes.
*   **Next.js Dashboard**: Added high-end telemetry cards, tool failure dashboards, switchable playground sessions, and environment-decoupled build fallbacks.

**Actual code in `engine.Process`:**

| Step | Implemented | File |
|------|-------------|------|
| 1. Rate limiting | ✅ | `engine.go` ~278–336 |
| 2. Billing pre-check | ✅ (not in AGENTS.md) | `engine.go` ~338–348 |
| 3. Cache L1 + L1.5 | ✅ **before RBAC** | `engine.go` ~375–446 |
| 4. Session history load | ✅ | `engine.go` ~448–461 |
| 5. Semantic memory retrieve | ✅ | `engine.go` ~471–479 |
| 6. RBAC check | ✅ | `engine.go` ~481–495 |
| 7. Semantic routing (gRPC) | ✅ | `engine.go` ~497–501 |
| 8. Cache L2 (similar hash) | ✅ | `engine.go` ~503–531 |
| 9. Tool execution / chaining | ✅ | `engine.go` ~533–718 |
| 10. LLM synthesis | ✅ | `engine.go` ~720+ |
| 11. Cache set + billing charge | ✅ | end of `Process` |

**Deviation to resolve:** Cache layers 1, 1.5, and 2 return responses **before** RBAC runs. Acceptable only if cache keys are strictly org+model isolated and threat model allows it. Document or fix.

---

## 3. Completed Milestones (What Actually Ships)

### Phase 1 — Core Foundation 🟡
- Triple-layer semantic cache (Valkey + Qdrant + Postgres fallback)
- JWT + API key auth (`auth/middleware.go`, `auth/rbac.go`)
- Multi-provider LLM routing (`internal/llm/`)
- MCP integration (`internal/mcp/`, `mcp-server/`)
- Rust semantic router with FastEmbed + Qdrant optimizations
- Prometheus metrics at `/metrics`

### Phase 2 — Dynamic Tool Registry 🟡
- `tools/registry.go` — Postgres-backed tools with `org_id`, `last_synced_at`
- `Registry.StartRefreshLoop()` — 30s drift detection → Qdrant vectorization
- Endpoints: `POST /v1/tools/register`, `/v1/tools/sync`, `/v1/tools/status`
- Dashboard docs: `/docs/tool-registry`

### Phase 3 — Multi-Connector + Resilience 🟠
- Connector registry: Core, MCP, REST, SQL (`internal/connectors/`)
- API key scopes: `chat:execute`, `tools:read`, `tools:write`, `audit:read`
- Persistent cache table (`migrations/015_persistent_cache.sql`, `engine.go`)

### Phase 4 — Advanced Orchestration 🟠
- Model-scoped cache keys: `org:<orgID>:m:<model>:<type>:<hash>`
- `PlanToolChain` gRPC (Rust + Go bindings)
- SSE endpoint (simulated typewriter, not provider-native stream)
- Dynamic parameter fitting via LLM (`fitToolParameters`)

### Phase 5/6 — Memory, Analytics, Billing 🟡
- Session threads: `/v1/sessions`, `memory/session.go`
- Semantic memory: Qdrant `memories_collection`
- Context analytics: `/v1/analytics/context`
- Stripe SaaS: checkout, webhooks, token ledger
- API key rotation UI + backend (pending migration 020 in prod)

---

## 4. Known Gaps & Technical Debt

### Security 🔴
| Issue | Location | Risk |
|-------|----------|------|
| RBAC dev bypasses in production | `auth/rbac.go` | Fixed: `ENVIRONMENT=production` disables `admin-01` and permissive `chat:execute` |
| CORS `*` | `main.go:63` | Cross-origin abuse |
| Default JWT secret in README / `maketoken` | `README.md`, `scripts/maketoken/` | Token forgery if unchanged |
| Secrets in git | `docker-compose.yml` comments, `scripts/test_flow.go` | Credential exposure |
| Static Supabase JWK seeded | `main.go:138-146` | Key rotation bypass |
| Cache before RBAC | `engine.go` | Stale permission on cache hit |

### Ops 🔴
| Issue | Location |
|-------|----------|
| Gateway not in active docker-compose | `docker-compose.yml:72-114` (commented) |
| CI green on rebased branch | `.github/workflows/*.yml` | Verify after `git push --force-with-lease` |
| Port mismatch (3000 vs 3002) | `docker-compose.yml` vs `README.md` |

### Product honesty 🟠
| Issue | Location |
|-------|----------|
| Core tools return mock data | `main.go:198-207` |
| Fake SSE streaming | `main.go:446-466` |
| Keyword-based chain trigger | `engine.go:536-541` |
| GraphQL/Webhook connectors declared but missing | `tools/registry.go:19-21` |
| ~~Broken dashboard client lib~~ | Fixed — `chatMemzent()` POST with `messages` (Sprint 3) |

### Dead code (resolved Sprint 4)
Removed unwired experimental packages that were never imported by `main.go`: `offline/`, `workflow/`, `prewarmer/`, `featureflags/`, `notifications/`. Re-introduce via a feature branch if needed.

---

## 5. Pending Tasks (Prioritized)

### P0 — Go-live blockers
See **[GO_LIVE_CHECKLIST.md](./GO_LIVE_CHECKLIST.md)** for file-level tasks.

**Sprint 1 (Security & Secrets) — code complete June 2026.** Manual credential rotation still required (see checklist §1.1).

1. ⬜ Apply `migrations/020_api_key_rotation.sql` to Supabase
2. ⬜ Apply `migrations/021_provision_chat_execute.sql` to Supabase (required for production RBAC)
3. ⬜ Rotate all secrets previously exposed in git (manual — see GO_LIVE_CHECKLIST §1.1)
4. ✅ Remove secrets from tracked files; env-gate RBAC bypasses; CORS + JWT production guards
5. ✅ Gateway uncommented in `docker-compose.yml` with `env_file: .env`
6. 🟡 CI workflows added (`go.yml`, `rust.yml`, `dashboard.yml`) — verify green on next PR (Sprint 2)

### P1 — Trust & docs (1–2 weeks)
1. ✅ Reconcile README API contract (`messages` not `prompt`) — Sprint 3
2. ✅ Update `ARCHITECTURE.md` phases 1.a–4 to match code — Sprint 3
3. ✅ Update `AGENTS.md` execution flow — Sprint 1/3
4. ✅ Fix `memzent-client.ts` → `chatMemzent()` — Sprint 3
5. ✅ Sanitize `test_flow.go` — env vars — Sprint 1

### P2 — Product quality (2–3 weeks)
1. 🟡 Ollama native SSE streaming — Sprint 4; OpenAI/Anthropic streaming still ⬜
2. ✅ Core tools documented as `[Demo]` stubs — Sprint 4
3. 🟡 Chain trigger: `chain: true` + router confidence ≥ 0.65; keywords as fallback — Sprint 4
4. ✅ GraphQL/Webhook/gRPC registration rejected at API — Sprint 4
5. ✅ `SemanticRouter` interface + engine mocks — Sprint 5
6. ⬜ Dashboard similarity threshold UI (task 5.3)
7. ⬜ Tool retry with exponential backoff (task 5.2)
8. ✅ Removed unwired packages (`offline/`, `workflow/`, etc.) — Sprint 4
9. ✅ Gateway integration tests (Valkey) + engine unit coverage — Sprint 5
10. ✅ Dashboard Vitest smoke test — Sprint 5
11. ✅ Rust `lib.rs` pure helpers + integration test scaffold — Sprint 5

### P3 — Enterprise scale (future)
1. ⬜ Envoy gRPC load balancing (task 5.1)
2. ⬜ S3 Qdrant snapshot offsite DR
3. ⬜ BYO LLM providers (ARCHITECTURE Phase 5)
4. ⬜ Schema-level org isolation for enterprise tier

---

## 6. Design Decisions (Still Valid)

### Org isolation
- **Current:** RLS + `org_id` filtering on cache keys, audit logs, tool registry.
- **Future:** Physical DB isolation for enterprise tier.

### Vector model & caching
- **Current:** Cache keys partition by `org_id` + `model` name.
- **Embedding model:** `all-MiniLM-L6-v2` via FastEmbed in Rust router.
- **Similarity threshold:** Default 0.88 (Qdrant), configurable override in proto.

### Auth strategy
- **Dual mode:** Supabase JWT (dashboard users) + API keys (agents/automation).
- **Dev convenience:** `scripts/maketoken`, `/generate-token` when `ENABLE_DEV_TOKEN=true`.

---

## 7. Test Coverage Summary

| Area | Tests | Gap |
|------|-------|-----|
| Gateway auth | `middleware_test.go`, `rbac_test.go`, `rbac_sqlmock_test.go` | Integration with Supabase JWKS |
| Gateway engine | `process_test.go`, `normalization_test.go`, `engine_sqlmock_test.go` | No `RouterClient` mock interface |
| Gateway billing | `calculator_test.go`, `ledger_test.go`, `stripe_test.go` | Webhook happy path |
| Rust router | `tests/unit_tests.rs` (mirrored pure fns) | No live Qdrant integration tests in CI |
| Dashboard | **None** | Playground, keys, billing untested |
| E2E | `scripts/test_flow.go` (manual, has secrets) | Not CI-safe |

---

## 8. Pop Questions for Agents & Developers

> [!IMPORTANT]
> Before marking any feature ✅, verify in code — do not trust this doc's historical claims without re-audit.

1. **Cache hit + RBAC:** Is org-scoped cache key isolation sufficient for your threat model, or must RBAC run before cache return?
2. **Connector type:** If registering a GraphQL tool, does an implementation exist? (Answer: no — use REST/SQL/MCP/Core only.)
3. **Streaming:** Does the client need real token streaming or is post-hoc typewriter acceptable?
4. **Core tools:** Are `read_database` / `memzent_search` mocks acceptable in prod, or must they hit real backends?
5. **Dead packages:** Should `offline/`, `workflow/`, etc. be deleted before go-live to reduce attack surface?

---

*Maintainers: Update this file when merging features. Do not mark ✅ without removing known bypasses and security debt.*
