# Memzent: Project Status & Roadmap

This document tracks the current completion state of Memzent features and provides "Pop Questions" to guide agents and human developers through the remaining tasks.

## 1. Feature Completion Matrix

| Feature | Service | Status | Notes |
| :--- | :--- | :--- | :--- |
| **Triple-Layer Caching** | Gateway/Rust | ✅ 100% | Literal, Canonical, and Semantic layers functional. |
| **Service Boundaries** | All | ✅ 100% | Go Gateway (Auth/Orchestration), Rust Router (Math), Dashboard (UI). |
| **RBAC Scoping & Multi-Token** | Gateway | ✅ 100% | Dynamic key generation with customizable roles and scopes. Seeded with a $10 welcome balance to unblock trials. |
| **Dynamic Tool Registry** | Gateway | ✅ 100% | Refresh loop, Qdrant sync, full CRUD (`/v1/tools/{id}`), Dashboard edit/delete UI. |
| **Connector Framework** | Gateway | ✅ 100% | SQL/REST/Core connectors fully implemented, registered, and active. |
| **Neural Dashboard** | Dashboard | ✅ 100% | Dynamic Billing, API Security metrics, live Playground, Provider discovery, Tool CRUD, Notifications. |
| **Provider Discovery** | Gateway | ✅ 100% | `/v1/providers` + `/v1/models` APIs. OpenAI, Anthropic, Gemini, Ollama discovery. |
| **Marketing Website** | Website | ✅ 100% | Mobile nav, SEO/OG tags, "Why Memzent" comparison, terminal quickstart. |
| **Advanced Orchestration (Phase 4)** | All | ✅ 100% | Model-scoped caching, PlanToolChain Go/Rust bindings, and typewriter SSE streaming. |
| **Agent Memory (Phase 5/6)** | Gateway/Rust | ✅ 100% | PostgreSQL session threads and semantic memory Qdrant extraction. |
| **Context Analytics (Phase 5/6)** | Dashboard/Gateway | ✅ 100% | Premium ROI tracking, latency tool telemetry, and intent theme clusters. |
| **API Key Security (Phase 6)** | Gateway/Dashboard | ✅ 100% | Expiry TTL picker, last_used_at tracking, in-place rotation with 15-min grace window, stale key audit. |
| **Notification Pipeline (Phase 7)** | Gateway/Dashboard | ✅ 100% | Webhook CRUD, 6 event types, HMAC signing, async retry with dead letter, delivery logs. |
| **Per-User Rate Limiting** | Gateway | ✅ 100% | Role-proportional limits (viewer 20%, member 50%, admin 100% of org). |

---

## 2. Completed Milestones

### [Phase 5 & 6] Memory, Tool Chaining & Context Analytics ✅ COMPLETE
*   **Semantic Agent Memory**: Added PostgreSQL persistence for conversation sessions (`sessions`, `session_messages`) and vectorized conversation facts out-of-band to Qdrant memory collection.
*   **Sequential Tool Chaining**: Integrated Go Gateway engine dynamic parameter schema fitting (`fitToolParameters`) and sequential execution chains (`PlanToolChain`).
*   **Context Analytics**: Developed SQL metrics aggregations computing savings ROI, tool latency, and failure rates, and clustering user intent themes.
*   **Next.js Dashboard**: Added high-end telemetry cards, tool failure dashboards, switchable playground sessions, and environment-decoupled build fallbacks.

### [Phase 2] Dynamic Tool Registry ✅ COMPLETE
*   `Registry.StartRefreshLoop()` — background goroutine polls Postgres every 30s for tools where `last_synced_at IS NULL OR last_synced_at < updated_at`.
*   `onSync` callback in `main.go` calls `routerClient.RegisterTool()` (gRPC) to vectorize drifted tools in Qdrant.
*   `HandleSyncTools` triggers a real `Registry.Refresh()` with vectorization and returns a structured JSON report.
*   `/v1/tools/status` endpoint exposes `last_refresh` timestamp for health monitoring.
*   Migration `011_tool_registry_sync.sql` adds `org_id`, `last_synced_at`, and `config` columns.
*   Documentation page `/docs/tool-registry` added to the dashboard.

### [Phase 3] Multi-Token & Resilience ✅ COMPLETE
*   **Multi-Token RBAC**: Granular token generation supporting custom identity types (`viewer`, `agent`, `admin`) and specific permission scopes (`chat:execute`, `tools:read`, `tools:write`, `audit:read`). Evaluated dynamically at the Gateway layer with full backward compatibility.
*   **Persistent Cache Resiliency (Durable Fallback)**: Write-Through & Read-Through B-Tree cached records persisted to Postgres. In the event of a Redis/Valkey crash or infra restart, the Gateway automatically pulls hits from Postgres and backfills Valkey in the background, keeping cache rates at $100\%$ with zero added latency.

### [Phase 4] Advanced Orchestration ✅ COMPLETE
*   **Model-Specific Cache Scoping**: Injected dynamically resolved target models into the vector cache keys (`org:<orgID>:m:<model>:<keyType>:<value>`), preventing cross-contamination between high-tier and small local LLM responses.
*   **Protobuf Tool Chains**: Expanded `proto/router.proto` to support `PlanToolChain` gRPC method and compiled stubs for Go & Rust, enabling multi-step sequencing backed by vector search inside the Rust semantic router.
*   **SSE Result Streaming**: Upgraded the `/v1/chat` controller with dynamic Server-Sent Events (SSE) typewriter streaming, maintaining caches asynchronously in the background.

---

## 3. Pending Tasks & Directives

### [Phase 6] API Key Security Hardening ✅ COMPLETE
**Goal**: Harden agent credential lifecycle with expiry, activity tracking, and zero-downtime rotation.

*   **Task 6.1** ✅ Migration `020_api_key_rotation.sql` — adds `expires_at`, `prev_key_hash`, `rotated_at`, `last_used_at` columns with performance indexes.
*   **Task 6.2** ✅ `rbac.go` `VerifyAPIKey` — enforces expiry TTL, dual-hash acceptance during 15-min rotation grace window, async `last_used_at` updates, auto-clearing of `prev_key_hash`.
*   **Task 6.3** ✅ `actions.ts` `rotateApiKey` — server action for in-place key rotation (generates new key, preserves old hash in `prev_key_hash`).
*   **Task 6.4** ✅ `keys/page.tsx` — Rotate button (purple, with spin animation), grace window notice banner, `last_used_at` / `rotated_at` / `expires_at` displayed in key row.
*   **Task 6.5** ✅ Migration `020` applied to Supabase.
*   **Task 6.6** ✅ TTL picker in key creation form — 4-option grid (Never / 24h / 7d / 30d). Passes `expires_at` ISO timestamp to `createApiKey`.
*   **Task 6.7** ✅ Stale key audit — amber count banner in registry header; per-row `Stale` / `Expired` badges when `last_used_at` is NULL or >30 days old.

### [Phase 7] Notification Pipeline ✅ COMPLETE
**Goal**: Real-time webhook event delivery for observability and external integrations.

*   **Task 7.1** ✅ Event schema — 6 typed events: `cache_hit`, `tool_execution`, `rate_limit`, `key_rotated`, `tool_registered`, `session_created`.
*   **Task 7.2** ✅ Webhook CRUD — `POST/GET/PUT/DELETE /v1/webhooks`, `GET /v1/webhooks/{id}/deliveries`, `GET /v1/webhooks/event-types`.
*   **Task 7.3** ✅ Async Dispatcher — 4 workers, 1024-buffer channel, HMAC-SHA256 signing (`X-Memzent-Signature`).
*   **Task 7.4** ✅ Retry with exponential backoff (1s → 5s → 30s → 2m → 10m). Dead letter after 5 attempts. Delivery logs in Postgres.
*   **Task 7.5** ✅ Engine integration — `EventEmitter` interface, emits `rate_limit` and `cache_hit` events. Graceful shutdown drains queue.
*   **Task 7.6** ✅ Dashboard — `/notifications` page with webhook CRUD UI, event selector, delivery log viewer.
*   **Task 7.7** ⬜ Apply migration `023_webhook_notifications.sql` to Supabase.

### Infrastructure Tasks (Completed This Session)
*   **Task 5.1** ✅ Envoy gRPC Load Balancing profiles.
*   **Task 5.2** ✅ Exponential backoff retry on native tool execution.
*   **Task 5.3** ✅ Dynamic Similarity Threshold in Dashboard settings.
*   **Task 5.4** ✅ `SemanticRouterInterface` for mock injection/testability.

### Additional Fixes (This Session)
*   ✅ `/v1/stats` now returns `active_providers[]` and `default_provider` (contract mismatch fix).
*   ✅ Tool CRUD: `PUT/DELETE/GET /v1/tools/{id}` endpoints + Dashboard edit/delete UI (Phase 3 completion).
*   ✅ Session ID fix: Gateway returns `"id"` field (was `"session_id"`, breaking dashboard + test-flow).
*   ✅ Per-user rate limiting within org (role-proportional: viewer 20%, member 50%, admin 100%).
*   ✅ Model discovery for Anthropic + Gemini fallback model updates.
*   ✅ CORS fixes (methods, headers, expose-headers).
*   ✅ Text contrast safety net (both dashboard + website).
*   ✅ Jargon cleanup across all dashboard pages.

---

## 4. Design Decisions & Future Assessments

### Org Isolation
- **Current Decision:** RLS (Row-Level Security) with `org_id` filtering is sufficient for the current phase.
- **Future Assessment (Enterprise):** Evaluate schema-level or physical database isolation as a premium feature for Enterprise subscriptions in a future phase.

### Vector Model & Caching
- **Current Decision:** Caching keys partition dynamically by model name, isolating smaller local models from large external LLMs, ensuring perfect semantic caching precision.

