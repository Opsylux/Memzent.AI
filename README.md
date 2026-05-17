# Memzent.ai: Memory of Agent

**Memzent (memzent.ai)** delivers the critical memory and security layer for autonomous workflows. Operating as an Intelligent Semantic Proxy, it intercepts and optimizes traffic between clients, MCP tools, and LLM providers. By combining semantic search and caching with enterprise-grade routing and RBAC, Memzent transforms stateless LLM calls into secure, context-aware agentic systems.

---

## 🏗️ Core Architecture

Memzent utilizes a distributed, multi-language architecture to balance high-speed semantic routing with robust business logic.

| Service | Language | Port | Role |
| :--- | :--- | :--- | :--- |
| **Gateway** | Go 1.25 | `8080` | Entry point, RBAC, JWT Auth, Triple-Layer Semantic Cache, Provider Routing |
| **Router** | Rust (Tonic) | `50051` | gRPC service for vector-based tool selection & prompt compression |
| **Dashboard** | Next.js 15+ | `3000` | Administrative control tower & observability |
| **MCP Server** | Go | `50052` | Tool execution & context protocol adapter |
| **Website** | Vite / React 19 | `5173` | Marketing landing page & user portal |

> See [ARCHITECTURE.md](./ARCHITECTURE.md) for the full sequence diagram and service topology.

---

## 🛡️ Enterprise Pillars

### 1. Triple-Layer Semantic Caching (Valkey + Qdrant)
Memzent uses a three-stage cache hierarchy before ever touching an LLM:
- **L1 – Literal**: SHA-256 exact hash match. `<5ms`.
- **L1.5 – Canonical**: Numeric noise (`write011`, `write202`) is masked to a stable form. Catches logically identical queries. `<5ms`.
- **L2 – Semantic (Vector)**: Cosine similarity via Qdrant at ≥0.88 threshold. `~10-30ms`.

Cache hits at any layer short-circuit to instant responses. On a semantic hit, all lower layers are back-filled for future precision hits.

### 2. Multi-Provider LLM Routing
Target any supported LLM backend per request via HTTP headers — no restart required.

```powershell
# Use OpenAI with a specific model
$headers = @{
    "Authorization" = "Bearer <token>"
    "X-Memzent-Provider" = "openai"
    "X-Memzent-Model" = "gpt-4o"
}

# Bypass cache for a real-time fresh response
$headers["X-Skip-Cache"] = "true"
```

**Supported providers**: `ollama` (default), `openai`, `anthropic`, `gemini`

### 3. Bulletproof Governance (Go + Postgres)
Centralized **Role-Based Access Control (RBAC)** and hardware-backed JWT authentication ensure AI agents only access the data they're authorized to see.

### 4. Semantic Tool Routing (Rust + Qdrant)
Memzent analyzes user intent in real-time and injects only the most relevant MCP tools into the LLM context — **reducing token waste by up to 90%** and eliminating hallucinations.

### 5. Deep Observability (Prometheus)
Every request is tracked via Prometheus metrics at `/metrics`. Monitor latency, cache hit rates, and token flow across your agentic fleet.

---

## 🚀 Getting Started

### Prerequisites
- **Docker & Docker Compose**
- **Go 1.24+** (for gateway development)
- **Rust** (for router development)
- **Ollama** running locally at `http://localhost:11434` with `llama3.2` pulled

### One-Command Deployment

```powershell
docker compose up -d --build
```

### Service Access
- **Gateway API**: [http://localhost:8080/v1/chat](http://localhost:8080/v1/chat)
- **Admin Dashboard**: [http://localhost:3000](http://localhost:3000)
- **Memzent Website**: [http://localhost:5173](http://localhost:5173)
- **Qdrant UI**: [http://localhost:6333/dashboard](http://localhost:6333/dashboard)
- **Metrics**: [http://localhost:8080/metrics](http://localhost:8080/metrics)
- **Dashboard Gateway URL**: set `NEXT_PUBLIC_GATEWAY_URL=http://localhost:8080` when running dashboard outside Docker

---

## 📡 API Reference

### POST `/v1/chat`

**Headers:**

| Header | Description |
| :--- | :--- |
| `Authorization: Bearer <jwt>` | Required. JWT token |
| `X-Memzent-Provider` | Optional. `ollama` / `openai` / `anthropic` / `gemini` |
| `X-Memzent-Model` | Optional. Model override (e.g. `gpt-4o`, `llama3.2:1b`) |
| `X-Skip-Cache` | Optional. `true` to bypass all 3 cache layers |

**Request Body:**

```json
{
  "user_id": "admin-01",
  "prompt": "how to reduce wal_buffer_waits in Aurora Postgres?",
  "provider": "openai",
  "model": "gpt-4o",
  "skip_cache": false
}
```

**Response:**

```json
{
  "text": "To reduce wal_buffer_waits...",
  "cached": false,
  "provider": "OpenAI"
}
```

**Response Headers:**

| Header | Value |
| :--- | :--- |
| `X-Cache` | `HIT` or `MISS` |
| `X-Memzent-Provider` | Active provider name |

---

## 📂 Project Structure

```
MemzentMCP/
├── services/
│   ├── gateway/        # Go 1.25: Primary Proxy, Auth, Cache, Provider Router
│   │   └── internal/
│   │       ├── engine/        # Orchestration engine (Triple-Layer Cache logic)
│   │       ├── llm/           # Provider implementations (Ollama, OpenAI, Anthropic, Gemini)
│   │       ├── router/        # gRPC client to Rust Router
│   │       ├── cache/         # Valkey semantic cache
│   │       ├── auth/          # JWT middleware + RBAC
│   │       └── mcp/           # MCP client
│   ├── router/         # Rust: Semantic Decision Engine (Qdrant + Tonic)
│   ├── mcp-server/     # Go: MCP Tool Provider
│   ├── dashboard/      # Next.js 15: Control Tower
│   └── website/        # Vite 8: Brand & Marketing
├── proto/              # Shared gRPC Definitions (router.proto)
├── data/               # Persistent Storage (Postgres/Qdrant volumes)
├── ARCHITECTURE.md     # Full system architecture & sequence diagrams
└── docker-compose.yml  # Orchestration Layer
```

---

## 🔑 Generating a JWT Token

```powershell
cd services/gateway
go run scripts/make_token.go
```

JWT secret: `memzent-enterprise-secret-2026` (configurable via `JWT_SECRET` env var).

---

**Built by the Memzent Engineering Team.** *Securing the future of Agentic Intelligence.*
