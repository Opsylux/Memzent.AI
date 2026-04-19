# Aura: Project Status & Roadmap

This document tracks the current completion state of Aura features and provides "Pop Questions" to guide agents and human developers through the remaining tasks.

## 1. Feature Completion Matrix

| Feature | Service | Status | Notes |
| :--- | :--- | :--- | :--- |
| **Triple-Layer Caching** | Gateway/Rust | ✅ 100% | Literal, Canonical, and Semantic layers functional. |
| **Service Boundaries** | All | ✅ 100% | Go Gateway (Auth/Orchestration), Rust Router (Math), Dashboard (UI). |
| **RBAC Scoping** | Gateway | ✅ 90% | JWT + Org-based filtering implemented; multi-org billing in progress. |
| **Dynamic Tool Registry** | Gateway | 🟡 60% | **Phase 2:** Postgres schema and basic handlers exist. Periodical refresh pending. |
| **Connector Framework** | Gateway | 🟡 40% | **Phase 3:** MCP is stable. SQL/REST/Core exist but need deep implementation. |
| **Neural Dashboard** | Dashboard | 🟡 80% | UI tokens and core views complete. Real-time log streaming pending. |

---

## 2. Pending Tasks & Directives

### [Phase 2] Dynamic Tool Registry (Priority: High)
**Goal**: Allow tools to be added to the database via API and automatically synced to the Gateway/Router.

*   **Task 2.1**: Implement `Registry.Refresh()` loop in Gateway to poll Postgres for tool updates.
*   **Task 2.2**: Integrate `toolRegistry` into the `auraEngine` so that tools are filtered by the Registry before being sent to the Router.
*   **Task 2.3**: Implement Tool Sync to Qdrant. When a tool is added to Postgres, its description must be vectorized in Qdrant.

> [!IMPORTANT]
> **Pop Question**: When adding a new tool to the database, who is responsible for generating its vector embedding?
> *   *Answer*: The Gateway must call the Rust Router's `UpsertTool` endpoint (to be implemented).

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

## 3. Pending Questions for User
- **Org Isolation**: Do we need hard physical isolation (separate DBs) for different orgs, or is row-level security (RLS) sufficient?
- **Vector Model**: Should we expose the choice of embedding model in `X-Aura-Model`, or keep it locked at the Router level for hash consistency?
