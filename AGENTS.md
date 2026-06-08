# AGENTS.md

> **Notice to AI System & Agents**: This file contains the mandatory operating rules, system constraints, and engineering stack for Project Memzent. You must parse and adhere strictly to these constraints before proposing any code edits and if need to change any of the rules,or add any new rules or new design decisions ask the user for approval.

## 1. Project Abstract
**Memzent (memzent.ai)** delivers the critical memory and security layer for autonomous workflows. Operating as an Intelligent Semantic Proxy, it intercepts and optimizes traffic between clients, MCP tools, and LLM providers. By combining semantic search and caching with enterprise-grade routing and RBAC, Memzent transforms stateless LLM calls into secure, context-aware agentic systems.
- **Core Goal**: Minimize LLM latency and maximize token ROI by semantically caching and efficiently routing user prompts *before* hitting expensive LLMs.

## 2. Engineering Standards & Boundaries

You must respect the specific language boundaries. Do not mix responsibilities.

### The Go Gateway (`/services/gateway`)
- **Role**: The "Front Door" orchestrator.
- **Rules**:
  - Handles *all* external HTTP traffic, JWT authentication, and RBAC via Postgres.
  - Handles *all* Semantic Caching logic via Valkey (Glide).
  - Handles *all* MCP Client connections and Tool execution mapping.
  - Handles *all* External API integrations (OpenAI, Anthropic, Ollama).
  - **Forbidden**: Do not do any vector math, embedding, or semantic similarity matching in Go.

### The Rust Router (`/services/router`)
- **Role**: The "Brain". A high-speed gRPC microservice.
- **Rules**:
  - Handles *all* Vector embeddings, similarity scoring, and tool matching algorithms.
  - Solely interacts with the Qdrant Vector DB.
  - **Forbidden**: Do not add business logic, authentication, or HTTP endpoints to Rust. It must remain a pure gRPC service defined by `/proto/router.proto`.

### The Next.js Dashboard (`/services/dashboard`)
- **Role**: The "Command Center".
- **Rules**:
  - Built with Next.js 15+ (React 19), Tailwind CSS v4, and Shadcn UI.
  - **Forbidden**: Do not use `pages/` directory. Strict App Router (`src/app/`) only.
  - Always map Tailwind v4 variables inside the `@theme inline` block in `globals.css` (Do not use `tailwind.config.js`).

## 3. Communication Patterns

- **Go <-> Rust**: Strictly over gRPC using generated protobufs (`/proto/router.proto`).
- **Gateway <-> Tools**: Strictly over the official Model Context Protocol (MCP).
- **Gateway <-> Cache**: Strictly using Valkey via `valkey-go` client.

## 4. Execution Flow Policy

When implementing new routing features, the AI must ensure the Engine (`internal/engine/engine.go` `Process()`) follows this exact sequence:

1. `Rate Limiting` — per-org token bucket, tier resolved from JWT (free 10/min, pro 100/min, business 1000/min). A positive pay-as-you-go balance promotes a free org to the pro limit.
2. `Billing Pre-Check` — reject orgs with a depleted balance before any compute is consumed (bypassed for internal dashboard / JWT sessions).
3. `L1 Literal Cache` — SHA-256 exact-hash lookup in Valkey, org-isolated and model-scoped. Falls back to the persistent Postgres cache on a Valkey miss/crash, then backfills Valkey.
4. `L1.5 Canonical Cache` — normalized-prompt hash (`NormalizePrompt`) lookup in Valkey, same Postgres fallback.
5. `Short-Term Memory` — load prior session messages (`sessionMgr`) when a `SessionID` is present.
6. `Long-Term Memory` — retrieve related semantic facts from the Qdrant memory collection (`memoryMgr`).
7. `RBAC Check` — Postgres org-scoped `chat:execute` permission check plus allowed-tools lookup.
8. `Semantic Routing` — gRPC call to the Rust router for vector tool selection and prompt compression.
9. `L2 Semantic Cache` — fuzzy vector match on the router's similar-prompt hash, org-isolated and model-scoped, same Postgres fallback.
10. `Tool Execution` — multi-connector dispatch (MCP / REST / SQL / Core), with sequential `PlanToolChain` chaining when the prompt implies ordering.
11. `Synthesis` — provider `Generate` (Ollama / OpenAI / Anthropic / Gemini) with SSE typewriter streaming.
12. `Cache Populate + Cost Deduct` — write L1 / L1.5 / L2 (Valkey + Postgres) and deduct the token cost from the billing ledger.

## 5. Agent Skills & Instructions

For detailed implementation patterns, pending roadmap items, and feature checklists, refer to:
- **[.cursorrules](./.cursorrules)**: IDE-specific rules for Antigravity, Copilot, and Cursor.
- **[PROJECT_STATUS.md](./PROJECT_STATUS.md)**: Code-verified feature matrix and honest completion status.
- **[GO_LIVE_CHECKLIST.md](./GO_LIVE_CHECKLIST.md)**: File-level production hardening tasks.
- **[INSTRUCTIONS.md](./INSTRUCTIONS.md)**: Step-by-step checklists and "Pop Questions" for feature implementation.
- **[ARCHITECTURE.md](./ARCHITECTURE.md)**: System topology, execution sequence, and phase roadmap (reconciled June 2026).
