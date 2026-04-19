# Aura: Project Status & Roadmap

This document tracks the current completion state of Aura features and provides "Pop Questions" to guide agents and human developers through the remaining tasks.

## 1. Feature Completion Matrix

| Feature | Service | Status | Notes |
| :--- | :--- | :--- | :--- |
| **Triple-Layer Caching** | Gateway/Rust | ✅ 100% | Literal, Canonical, and Semantic layers functional. |
| **Service Boundaries** | All | ✅ 100% | Go Gateway (Auth/Orchestration), Rust Router (Math), Dashboard (UI). |
| **RBAC Scoping** | Gateway | ✅ 90% | JWT + Org-based filtering implemented; multi-org billing in progress. |
| **Dynamic Tool Registry** | Gateway | ✅ 100% | **Phase 2 Complete:** Refresh loop, Qdrant sync, `/v1/tools/sync`, `/v1/tools/status`. |
| **Connector Framework** | Gateway | 🟡 40% | **Phase 3:** MCP is stable. SQL/REST/Core exist but need deep implementation. |
| **Neural Dashboard** | Dashboard | ✅ 90% | UI complete; Docs hardened with Navigation & Variable domain. |
| **Provider Discovery** | Gateway | ✅ 100% | `/v1/providers` API for dynamic model/provider listing. |

---

## 2. Pending Tasks & Directives

### [Phase 2] Dynamic Tool Registry ✅ COMPLETE
**Completed Tasks:**
*   **Task 2.1**: `Registry.StartRefreshLoop()` — background goroutine polls Postgres every 30s for tools where `last_synced_at IS NULL OR last_synced_at < updated_at`.
*   **Task 2.2**: `onSync` callback in `main.go` calls `routerClient.RegisterTool()` (gRPC) to vectorize each drifted tool in Qdrant.
*   **Task 2.3**: `HandleSyncTools` now triggers a real `Registry.Refresh()` with vectorization and returns a structured JSON report.
*   **Task 2.4**: New `/v1/tools/status` endpoint exposes `last_refresh` timestamp for health monitoring.
*   **Task 2.5**: Migration `011_tool_registry_sync.sql` adds `org_id`, `last_synced_at`, and `config` columns.
*   **Task 2.6**: Documentation page `/docs/tool-registry` added to the dashboard.

---

### [Phase 3] Multi-Connector Framework (Priority: Medium)
**Goal**: Support tools that aren't MCP-based (e.g., direct SQL or REST).

*   **Task 3.1**: Finish `RESTConnector.Execute`. Implement standard HTTP client with JSON mapping.
*   **Task 3.2**: Finish `SQLConnector.Execute`. Ensure row serialization to JSON is robust.
*   **Task 3.3**: Implement `GraphQLConnector`.

> [!IMPORTANT]
> **Pop Question**: Should a SQL tool execute with the `org_id` of the user, or a system-wide read-only credential?
> *   *Answer*: The connector should use a tool-specific connection string stored securely in the Registry's `config` column.

---

### [Phase 4] Advanced Orchestration (Priority: Low)
**Goal**: Multi-step agentic flows.

*   **Task 4.1**: Define "Tool Chain" schema in Protobuf.
*   **Task 4.2**: Implement result streaming (SSE) in Gateway.

---

## 3. Design Decisions & Future Assessments

### Org Isolation
- **Current Decision:** RLS (Row-Level Security) with `org_id` filtering is sufficient for the current phase.
- **Future Assessment (Enterprise):** Evaluate schema-level or physical database isolation as a premium feature for Enterprise subscriptions in a future phase.

### Vector Model & Routing
- **Current Decision:** Users and agents should NOT be restricted to a single model; they should be able to choose their preferred model.
- **Action Item / Design Review:** We need a design review on how to handle caching for multiple models. This may involve adding custom headers (e.g., `X-Aura-Model`) to specify the model, and adjusting the cache lookup logic to ensure cache keys are scoped to the requested model. We'll also need to consider if/how we can share cached responses across different models if appropriate.
