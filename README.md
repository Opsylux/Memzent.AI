# Memzent.ai — The Open-Source Semantic Proxy for LLMs

<p align="center">
  <a href="https://memzent.ai"><img src="https://img.shields.io/badge/Website-memzent.ai-blue?style=flat-square" /></a>
  <a href="https://github.com/Opsylux/Memzent.AI/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-Apache_2.0-green?style=flat-square" /></a>
  <a href="https://github.com/Opsylux/Memzent.AI/stargazers"><img src="https://img.shields.io/github/stars/Opsylux/Memzent.AI?style=flat-square" /></a>
  <a href="https://github.com/Opsylux/Memzent.AI/issues"><img src="https://img.shields.io/github/issues/Opsylux/Memzent.AI?style=flat-square" /></a>
  <a href="https://discord.gg/memzent"><img src="https://img.shields.io/discord/placeholder?label=Discord&style=flat-square" /></a>
</p>

<p align="center">
  <strong>Intelligent memory, caching, and security layer for autonomous AI agents.</strong><br/>
  Intercepts traffic between clients, MCP tools, and LLM providers. Minimizes latency. Maximizes token ROI.
</p>

<p align="center">
  <a href="#-quick-start-self-hosted">Self-Host</a> ·
  <a href="https://app.memzent.ai">Memzent Cloud</a> ·
  <a href="https://memzent.ai/blog">Blog</a> ·
  <a href="#-contributing">Contributing</a> ·
  <a href="https://discord.gg/memzent">Discord</a>
</p>

---

## Memzent Cloud vs Self-Hosted

|  | Self-Hosted (This Repo) | Memzent Cloud |
|--|-------------------------|---------------|
| **Infrastructure** | You manage | We manage |
| **Pricing** | Free forever | Pay-as-you-go |
| **Code** | Apache 2.0 | Same codebase |
| **Support** | Community (GitHub + Discord) | Priority support + SLA |
| **Updates** | Pull from `main` | Auto-deployed |
| **Scale** | Your hardware | Auto-scaling infra |
| **Features** | Everything | Everything + managed backups, monitoring, team SSO |

> **All code in this repository is open source under Apache 2.0.** The managed cloud is the same code — we just run it for you.

---

## 🚀 Quick Start (Self-Hosted)

```bash
# Clone the repo
git clone https://github.com/Opsylux/Memzent.AI.git
cd Memzent.AI

# Start the full stack (Gateway + Router + Qdrant + Valkey)
docker-compose up -d

# Verify it's running
curl http://localhost:8080/health

# Send your first request
curl -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"messages": [{"role": "user", "content": "Hello, Memzent!"}]}'
```

> Requires Docker and Docker Compose. See [SELF_HOSTING.md](SELF_HOSTING.md) for detailed configuration.

---

## 🏗️ Core Architecture

Memzent utilizes a distributed, multi-language architecture to balance high-speed semantic routing with robust business logic.

| Service | Language | Port | Role |
| :--- | :--- | :--- | :--- |
| **Gateway** | Go 1.25 | `8080` | Entry point, RBAC, JWT Auth, 4-Layer Cache, Entity Extraction, Provider Routing |
| **Router** | Rust (Tonic) | `50051` | gRPC service for vector-based tool selection, embeddings & prompt compression |
| **Dashboard** | Next.js 16 (React 19) | `3000` | Administrative control tower, docs & blog |
| **MCP Server** | Go | `50052` | Tool execution & context protocol adapter |
| **Website** | Vite / React 19 | `5173` | Marketing landing page |

---

## 🛡️ Enterprise Pillars

### 1. Four-Layer Semantic Caching (Valkey + Qdrant)
Memzent uses a four-stage cache hierarchy before ever touching an LLM:

| Layer | Method | Latency | What it catches |
|-------|--------|---------|-----------------|
| **L1** | SHA-256 literal hash | <1ms | Exact duplicate prompts |
| **L1.5** | Canonical hash | <1ms | Formatting/whitespace differences |
| **L1b** | Entity-keyed hot path | 1-2ms | Same entities, different phrasing |
| **L2** | Vector similarity (Qdrant) | 15-50ms | Semantic paraphrasing (≥0.95 threshold) |

Cache hits at any layer short-circuit to instant responses. On a miss, all layers are back-filled for future hits.

### 2. Entity-Aware Cache Guard (Evolution Pipeline E1)
Regex-based extraction (<1ms) identifies 6 typed entities — accounts, customers, invoices, amounts, dates, identifiers — with **directional awareness**. Prevents false cache hits when entity values or positions differ.

### 3. Multi-Provider LLM Routing

```bash
curl -X POST https://api.memzent.ai/v1/chat \
  -H "X-API-Key: memzent_YOUR_KEY" \
  -H "X-Memzent-Provider: openai" \
  -H "X-Memzent-Model: gpt-4o" \
  -d '{"messages": [{"role": "user", "content": "Hello"}]}'
```

**Supported providers**: `ollama` (default), `openai`, `anthropic`, `gemini`

### 4. Bulletproof Governance (Go + Postgres)
Role-Based Access Control (RBAC) with JWT + API key auth. Per-user rate limiting (role-proportional), spend limits (daily/monthly dollar + token caps), and full audit trail.

### 5. Semantic Tool Routing (Rust + Qdrant)
Analyzes user intent in real-time and injects only the most relevant MCP tools into the LLM context — reducing token waste by up to 90%.

### 6. Deep Observability (Prometheus)
Every request is tracked via Prometheus metrics at `/metrics`. Entity extraction counters, cache layer distribution, GPU avoidance rates, and token flow.

---

## 🧬 Evolution Pipeline (E1–E6)

Six layers of intelligence that eliminate redundant GPU inference:

| Stage | Feature | Description |
|-------|---------|-------------|
| **E1** | Entity Extraction | Regex-based typed extraction (<1ms) with positional awareness |
| **E2** | L1b Hot Path Cache | Entity-keyed deterministic Valkey lookup, sub-millisecond |
| **E3** | Offline Learning Plane | Async telemetry mining (PII-safe) with 3 miners |
| **E4** | Workflow Registry | Auto-discovered multi-step sequences execute as shortcuts |
| **E5** | GPU Avoidance Metrics | 8 Prometheus counters tracking avoidance rate |
| **E6** | Pattern Mining | Markov chain prediction + speculative pre-warming (experimental) |

**GPU Avoidance Rate** = `cache_hits / total_requests` — the primary business metric. Production target: **80%+**.

### Feature Flags

| Variable | Default | Purpose |
|----------|---------|---------|
| `MEMZENT_L1B_ENABLED` | `true` | L1b entity-keyed cache |
| `MEMZENT_OFFLINE_ENABLED` | `true` | Offline learning plane |
| `MEMZENT_WORKFLOW_ENABLED` | `true` | Workflow registry |
| `MEMZENT_ENTITY_METRICS_ENABLED` | `true` | GPU avoidance counters |
| `MEMZENT_PATTERN_MINING_ENABLED` | `false` | E6 Markov chain (experimental) |
| `MEMZENT_OFFLINE_STREAMS` | `false` | Valkey Streams transport |

---

## 🚀 Getting Started

### Prerequisites
- **Docker & Docker Compose**
- **Go 1.25+** (for gateway development)
- **Rust** (for router development)
- **Ollama** running locally at `http://localhost:11434` with `llama3.2` pulled

### One-Command Deployment

```bash
make up       # docker-compose up -d --build
make down     # stop all
make logs     # follow gateway + router logs
```

### Service Access
- **Gateway API**: [http://localhost:8080/v1/chat](http://localhost:8080/v1/chat)
- **Admin Dashboard**: [http://localhost:3000](http://localhost:3000)
- **Documentation**: [http://localhost:3000/docs](http://localhost:3000/docs)
- **Website**: [http://localhost:5173](http://localhost:5173)
- **Qdrant UI**: [http://localhost:6333/dashboard](http://localhost:6333/dashboard)
- **Metrics**: [http://localhost:8080/metrics](http://localhost:8080/metrics)

---

## 📡 API Reference

### Core Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/v1/chat` | POST | Send prompts, get AI responses |
| `/v1/providers` | GET | List available LLM providers |
| `/v1/models` | GET | List available models |
| `/v1/sessions` | POST | Create conversation sessions |
| `/v1/sessions/{id}/messages` | GET | Retrieve session messages |
| `/v1/tools/register` | POST | Register tools for semantic routing |
| `/v1/tools/{id}` | GET/PUT/DELETE | Tool CRUD |
| `/v1/stats` | GET | Cache stats & analytics |
| `/v1/cache/flush` | POST | Flush cache (scope: valkey\|db\|all) |
| `/v1/settings/threshold` | GET/PUT | Similarity threshold config |

### Billing & Spend Limits

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/v1/billing/budget` | GET | Balance, burn rate, projections |
| `/v1/billing/spend-limits` | GET/PUT | Daily/monthly dollar + token caps |
| `/v1/billing/spend-timeseries` | GET | Daily spend data for charts |
| `/v1/billing/checkout` | POST | Stripe checkout for top-ups |

### Webhooks

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/v1/webhooks` | POST/GET | Create/list webhook subscriptions |
| `/v1/webhooks/{id}` | PUT/DELETE | Update/delete webhooks |
| `/v1/webhooks/{id}/deliveries` | GET | Delivery logs |
| `/v1/webhooks/event-types` | GET | Available event types |

### Request Headers

| Header | Description |
| :--- | :--- |
| `X-API-Key: memzent_...` | API key authentication |
| `Authorization: Bearer <jwt>` | JWT authentication |
| `X-Memzent-Provider` | Optional. `ollama` / `openai` / `anthropic` / `gemini` |
| `X-Memzent-Model` | Optional. Model override (e.g. `gpt-4o`, `llama3.2:1b`) |
| `X-Skip-Cache` | Optional. `true` to skip cache reads (still writes) |
| `X-Session-ID` | Optional. Session UUID for conversation continuity |

---

## 🧪 Testing

```bash
# Gateway unit tests
cd services/gateway && go test ./...

# Integration test suites (require running stack)
make test-cache       # 12 semantic cache correctness tests
make test-entity      # 14 entity extraction + cache guard tests
make test-memory      # 10 agent memory + session isolation tests
make test-evolution   # 28 Evolution Pipeline E1-E5 assertions
make test-flow        # 20-worker load test against /v1/chat
```

---

## 📂 Project Structure

```
Memzent.AI/
├── services/
│   ├── gateway/        # Go 1.25: Primary Proxy, Auth, Cache, Provider Router
│   │   └── internal/
│   │       ├── engine/        # Orchestration engine (4-Layer Cache + Entity Extraction)
│   │       ├── llm/           # Provider implementations (Ollama, OpenAI, Anthropic, Gemini)
│   │       ├── router/        # gRPC client to Rust Router
│   │       ├── cache/         # Valkey semantic cache + L1b hot path
│   │       ├── auth/          # JWT middleware + RBAC
│   │       ├── billing/       # Stripe, ledger, spend limits
│   │       ├── workflow/      # Workflow registry + simulator
│   │       ├── offline/       # Offline learning plane + miners
│   │       ├── prewarmer/     # Speculative L1b cache pre-warming
│   │       ├── featureflags/  # Environment-based feature flags
│   │       ├── notifications/ # Webhook dispatcher + retry
│   │       ├── memory/        # Session threads + semantic memory
│   │       ├── metrics/       # Prometheus + persistent audit
│   │       ├── mcp/           # MCP client
│   │       └── tools/         # Dynamic tool registry
│   ├── router/         # Rust: Semantic Decision Engine (Qdrant + Tonic)
│   ├── mcp-server/     # Go: MCP Tool Provider
│   ├── dashboard/      # Next.js 16 (React 19): Control Tower, Docs, Blog
│   └── website/        # Vite + React 19: Marketing Site
├── proto/              # Shared gRPC Definitions (router.proto)
├── migrations/         # SQL migrations (001-026+)
├── data/               # Persistent Storage (Postgres/Qdrant volumes)
└── docker-compose.yml  # Orchestration Layer
```

---

## 🔑 Authentication

```bash
# Generate a JWT token
cd services/gateway && go run scripts/make_token.go
```

JWT secret configurable via `JWT_SECRET` env var. Production uses API keys (`X-API-Key: memzent_...`) with role-based scopes.

---

## 🤝 Contributing

We love contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

```bash
# Fork → Clone → Branch → Code → Test → PR
git checkout -b feat/my-feature
go test ./... && cargo test
# Open a PR targeting main
```

Good first issues are labelled [`good-first-issue`](https://github.com/Opsylux/Memzent.AI/labels/good-first-issue).

---

## 📜 License

[Apache 2.0](LICENSE) — use it, modify it, deploy it, sell services on top of it. Just give attribution.

---

## ⭐ Star History

If Memzent helps your team, consider giving us a star. It helps others discover the project.

---

<p align="center">
  <strong>Built in the open by <a href="https://github.com/Opsylux">Opsylux</a></strong><br/>
  <em>Securing the future of Agentic Intelligence.</em>
</p>
