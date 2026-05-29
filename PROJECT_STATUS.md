# Memzent: Project Status & Roadmap

This document tracks the current completion state of Memzent features and provides "Pop Questions" to guide agents and human developers through the remaining tasks.

## 1. Feature Completion Matrix

| Feature | Service | Status | Notes |
| :--- | :--- | :--- | :--- |
| **Triple-Layer Caching** | Gateway/Rust | ✅ 100% | Literal, Canonical, and Semantic layers functional. |
| **Service Boundaries** | All | ✅ 100% | Go Gateway (Auth/Orchestration), Rust Router (Math), Dashboard (UI). |
| **RBAC Scoping & Multi-Token** | Gateway | ✅ 100% | Dynamic key generation with customizable roles and scopes. Seeded with a $10 welcome balance to unblock trials. |
| **Dynamic Tool Registry** | Gateway | ✅ 100% | Refresh loop, Qdrant sync, `/v1/tools/sync`, `/v1/tools/status`. |
| **Connector Framework** | Gateway | ✅ 100% | SQL/REST/Core connectors fully implemented, registered, and active. |
| **Neural Dashboard** | Dashboard | ✅ 100% | Dynamic Billing (Option B), API Security metrics, live Playground with cost trace, and Provider discovery. |
| **Provider Discovery** | Gateway | ✅ 100% | `/v1/providers` API for dynamic model/provider listing. |
| **Marketing Website** | Website | ✅ 100% | Hybrid PAYG billing explainer, sleek dark look, logo bug fixed. |
| **Advanced Orchestration (Phase 4)** | All | ✅ 100% | Model-scoped caching, PlanToolChain Go/Rust bindings, and typewriter SSE streaming. |
| **Agent Memory (Phase 5/6)** | Gateway/Rust | ✅ 100% | PostgreSQL session threads and semantic memory Qdrant extraction. |
| **Context Analytics (Phase 5/6)** | Dashboard/Gateway | ✅ 100% | Premium ROI tracking, latency tool telemetry, and intent theme clusters. |
| **API Key Security Hardening** | Gateway/Dashboard | ✅ Code Complete | Expiry TTL picker, last_used_at tracking, in-place rotation with 15-min grace window, stale key audit. ⬜ Migration 020 needs applying to Supabase. |

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

### [Phase 6] API Key Security Hardening (Priority: High — In Progress)
**Goal**: Harden agent credential lifecycle with expiry, activity tracking, and zero-downtime rotation.

*   **Task 6.1** ✅ Migration `020_api_key_rotation.sql` — adds `expires_at`, `prev_key_hash`, `rotated_at`, `last_used_at` columns with performance indexes.
*   **Task 6.2** ✅ `rbac.go` `VerifyAPIKey` — enforces expiry TTL, dual-hash acceptance during 15-min rotation grace window, async `last_used_at` updates, auto-clearing of `prev_key_hash`.
*   **Task 6.3** ✅ `actions.ts` `rotateApiKey` — server action for in-place key rotation (generates new key, preserves old hash in `prev_key_hash`).
*   **Task 6.4** ✅ `keys/page.tsx` — Rotate button (purple, with spin animation), grace window notice banner, `last_used_at` / `rotated_at` / `expires_at` displayed in key row.
*   **Task 6.5** ⬜ Apply migration `020` to Supabase production:
    ```
    -- Option A: Supabase CLI
    supabase db push
    -- Option B: Paste migrations/020_api_key_rotation.sql into Supabase SQL Editor
    ```
*   **Task 6.6** ✅ TTL picker in key creation form — 4-option grid (Never / 24h / 7d / 30d). Passes `expires_at` ISO timestamp to `createApiKey`.
*   **Task 6.7** ✅ Stale key audit — amber count banner in registry header; per-row `Stale` / `Expired` badges when `last_used_at` is NULL or >30 days old.

*   **Task 5.1**: Define Envoy Load Balancing profiles for gRPC streams.
*   **Task 5.2**: Implement custom retry policies with exponential backoff on native tools.
*   **Task 5.3**: Add Dynamic Similarity Threshold settings to the Next.js Dashboard.
*   **Task 5.4**: Refactor `RouterClient` into a `SemanticRouterInterface` to allow mock injection, unlocking 80-100% code coverage for `engine.Process`.

---

## 4. Design Decisions & Future Assessments

### Org Isolation
- **Current Decision:** RLS (Row-Level Security) with `org_id` filtering is sufficient for the current phase.
- **Future Assessment (Enterprise):** Evaluate schema-level or physical database isolation as a premium feature for Enterprise subscriptions in a future phase.

### Vector Model & Caching
- **Current Decision:** Caching keys partition dynamically by model name, isolating smaller local models from large external LLMs, ensuring perfect semantic caching precision.

