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

When implementing new routing features, the AI must ensure the Engine follows this exact sequence:
1. `Rate Limiting` (Token bucket check - Not implemented or need to review it, make sure its alinged with tiers)
2. `Cache Check` (Hash lookup in Valkey)
3. `RBAC Check` (Postgres query for user permissions)
4. `Semantic Routing` (gRPC call to Rust/Qdrant)
5. `Tool Execution` (Fire off matched MCP tools)
6. `Synthesis` (Pass context + prompt to Ollama/OpenAI/Anthropic)
7. `Cache Set` (Store synthesized output in Valkey)

## 5. Agent Skills & Instructions

For detailed implementation patterns, pending roadmap items, and feature checklists, refer to:
- **[.cursorrules](file:///c:/Users/nnaga/OneDrive/Documents/GitHub/MemzentMCP/.cursorrules)**: IDE-specific rules for Antigravity, Copilot, and Cursor.
- **[PROJECT_STATUS.md](file:///c:/Users/nnaga/OneDrive/Documents/GitHub/MemzentMCP/PROJECT_STATUS.md)**: Live roadmap and "What's Pending" tracker.
- **[INSTRUCTIONS.md](file:///c:/Users/nnaga/OneDrive/Documents/GitHub/MemzentMCP/INSTRUCTIONS.md)**: Step-by-step checklists and "Pop Questions" for feature implementation.