# Memzent Gateway — Open Source Components

> This document defines exactly which components of the Memzent Gateway are
> open-sourced under Apache 2.0, what each component does, and how to use them.

---

## What Is Open Source

The **Memzent Gateway Core** is the open-source foundation of the Memzent
Agentic Infrastructure Platform. It provides the HTTP entry point, multi-provider
LLM routing, MCP protocol integration, L1 literal caching, connector framework,
and observability layer — everything you need to run a production-ready AI
gateway today.

```
memzent/                          ← public repo (Apache 2.0)
├── services/
│   ├── gateway/                  ← Go gateway core  ✅ open
│   │   ├── main.go
│   │   ├── internal/
│   │   │   ├── auth/             ← JWT + JWKS middleware  ✅ open
│   │   │   ├── cache/            ← L1 Valkey cache  ✅ open
│   │   │   ├── config/           ← Config loader  ✅ open
│   │   │   ├── connectors/       ← MCP, REST, SQL, Core  ✅ open
│   │   │   ├── engine/           ← Orchestration engine  ✅ open
│   │   │   │   ├── engine.go
│   │   │   │   └── normalization.go  ← L1.5 canonical cache  ✅ open
│   │   │   ├── llm/              ← OpenAI, Anthropic, Gemini, Ollama  ✅ open
│   │   │   ├── mcp/              ← MCP client + compressor  ✅ open
│   │   │   ├── metrics/          ← Prometheus + audit log  ✅ open
│   │   │   ├── router/           ← gRPC client stub  ✅ open
│   │   │   └── tools/            ← Tool registry  ✅ open
│   │   └── migrations/           ← Core DB schema  ✅ open
│   ├── mcp-server/               ← MCP protocol adapter  ✅ open
│   ├── dashboard/                ← Next.js admin UI  ✅ open
│   └── website/                  ← Marketing site  ✅ open
├── proto/
│   └── router.proto              ← gRPC contract  ✅ open
├── docker-compose.yml            ← One-command local stack  ✅ open
└── README.md
```

---

## Component Reference

### 1. Gateway Entry Point (`main.go`)

The HTTP server and middleware chain. Handles incoming requests and wires
all components together.

**What it does:**
- Starts the HTTP server on port `8080`
- Registers all middleware (CORS, metrics, auth)
- Mounts API routes with scope enforcement
- Manages graceful shutdown on SIGTERM

**Key routes exposed:**

| Route | Method | Scope required | Description |
|---|---|---|---|
| `/v1/chat` | POST | `chat:execute` | Main inference endpoint with SSE streaming |
| `/v1/tools` | GET | `tools:read` | List all registered tools |
| `/v1/tools` | POST | `tools:write` | Register a new tool |
| `/v1/tools/sync` | GET | `tools:write` | Trigger manual Qdrant sync |
| `/v1/tools/status` | GET | `tools:read` | Registry sync status |
| `/v1/audit` | GET | `audit:read` | Audit log (last 50 events) |
| `/v1/stats` | GET | `audit:read` | Cache hit rate, uptime, token balance |
| `/v1/providers` | GET | — | List active LLM providers and models |
| `/healthz` | GET | — | Liveness probe |
| `/readyz` | GET | — | Readiness probe (checks Valkey) |
| `/metrics` | GET | — | Prometheus metrics |

**Example request:**
```bash
curl -X POST http://localhost:8080/v1/chat \
  -H "Authorization: Bearer <your-jwt>" \
  -H "Content-Type: application/json" \
  -H "X-Memzent-Provider: openai" \
  -H "X-Memzent-Model: gpt-4o" \
  -d '{"messages": [{"role": "user", "content": "What is MCP?"}]}'
```

---

### 2. Orchestration Engine (`internal/engine/engine.go`)

The central orchestrator. Every `/v1/chat` request passes through here.

**Request pipeline — in order:**

```
1. Rate limiting      → per-user token bucket (golang.org/x/time/rate)
2. L1 cache check     → SHA-256 hash lookup in Valkey
3. L1.5 cache check   → Canonical normalised hash lookup in Valkey
4. RBAC check         → JWT scope + Postgres permission lookup
5. Semantic routing   → gRPC call to Rust router (tool selection)
6. Tool execution     → Connector framework (MCP / REST / SQL / Core)
7. LLM synthesis      → Provider call with injected tool context
8. Cache populate     → Write result to L1, L1.5, and Postgres fallback
9. Response           → JSON or SSE stream with X-Cache header
```

**Key types:**
```go
type PromptRequest struct {
    UserID    string        `json:"user_id"`
    Messages  []llm.Message `json:"messages"`
    Provider  string        `json:"provider,omitempty"`
    Model     string        `json:"model,omitempty"`
    SkipCache bool          `json:"skip_cache,omitempty"`
    Stream    bool          `json:"stream,omitempty"`
}

type PromptResponse struct {
    Text      string `json:"text"`
    Cached    bool   `json:"cached"`
    Provider  string `json:"provider,omitempty"`
    Tools     []any  `json:"tools,omitempty"`
    RequestID string `json:"request_id,omitempty"`
}
```

**SSE streaming:**

Set `"stream": true` or `Accept: text/event-stream` header. Each chunk:
```
data: {"text": "Paris ", "cached": false, "provider": "openai", "request_id": "abc123"}

data: [DONE]
```

---

### 3. Canonical Normalisation (`internal/engine/normalization.go`)

The L1.5 cache layer. One of Memzent's most distinctive innovations.

**What it does:**

Transforms any prompt into a canonical form before hashing, so logically
identical queries with different numeric IDs hit the same cache entry.

```
"Show me order 1234"  →  "show me order <id>"  →  hash: abc...
"Show me order 5678"  →  "show me order <id>"  →  hash: abc...
                                                   ↑ same — cache HIT
```

**Normalisation steps:**
1. Lowercase and trim whitespace
2. Replace all 2+ digit sequences with `<id>` token
3. Remove punctuation, stabilise spaces
4. SHA-256 hash the canonical form

**Public API:**
```go
// Returns (canonical string, sha256 hash)
canonical, hash := engine.NormalizePrompt("Show me order 1234")
// canonical: "show me order <id>"
// hash:      "3f4a..."
```

**Why this matters:**

No competitor implements this step. LiteLLM and Portkey use single-layer
exact-match or pure vector caching. The L1.5 canonical layer catches the
common enterprise pattern of repeated queries that differ only in record IDs —
giving cache hits at sub-5ms latency instead of a full LLM round-trip.

---

### 4. L1 Cache (`internal/cache/valkey.go`)

Valkey (Redis-compatible) client for the L1 literal cache layer.

**What it does:**
- Stores and retrieves LLM responses keyed by prompt hash
- Uses the native Valkey Go client for maximum throughput
- TTL-based expiry (configurable via `LLM_CACHE_TTL` env var)

**Cache key format:**
```
org:<org_id>:m:<model>:l1:<sha256_of_raw_prompt>      ← L1 literal
org:<org_id>:m:<model>:l15:<sha256_of_canonical>      ← L1.5 canonical
```

**API:**
```go
result, err := cache.GetSemanticResult(ctx, hashKey)  // "" = miss
err = cache.SetResult(ctx, hashKey, responseText, ttl)
err = cache.Ping(ctx)
```

---

### 5. Authentication Middleware (`internal/auth/`)

JWT validation and RBAC enforcement.

**Two auth modes supported:**
- **Static JWT** — validated with `JWT_SECRET` env var (HS256)
- **JWKS/Supabase** — dynamic key discovery via `JWKS_URL` (ES256/RS256)

**Identity types and scopes:**

| Identity | Allowed scopes |
|---|---|
| `viewer` | `tools:read` |
| `agent` | `chat:execute`, `tools:read` |
| `admin` | `chat:execute`, `tools:read`, `tools:write`, `audit:read` |

**JWT payload example:**
```json
{
  "sub": "user_123",
  "org_id": "org_456",
  "role": "agent",
  "scopes": ["chat:execute", "tools:read"],
  "exp": 1748000000
}
```

**Request headers:**
```
Authorization: Bearer <jwt>          ← required
X-Org-ID: org_456                    ← optional override
X-Skip-Cache: true                   ← bypass all cache layers
X-Memzent-Provider: openai           ← override LLM provider
X-Memzent-Model: gpt-4o              ← override model
```

---

### 6. LLM Providers (`internal/llm/`)

Multi-provider LLM abstraction. Common `Provider` interface for all backends.

**Supported providers:**

| Provider | Env var required | Default model |
|---|---|---|
| Ollama (local, default) | `OLLAMA_URL` | `llama3.2` |
| OpenAI | `OPENAI_API_KEY` | `gpt-4o` |
| Anthropic | `ANTHROPIC_API_KEY` | `claude-sonnet-4-6` |
| Gemini | `GEMINI_API_KEY` | `gemini-2.0-flash` |

**Provider interface:**
```go
type Provider interface {
    // Generate produces an LLM response. model may be empty to use the provider default.
    Generate(ctx context.Context, messages []Message, tools []any, model string) (string, *TokenUsage, error)
    GetProviderName() string
    GetMetadata() ProviderMetadata
}
```

**Adding a new provider:**
1. Create `internal/llm/myprovider.go` implementing `Provider`
2. Register in `main.go`:
```go
if cfg.MyAPIKey != "" {
    providers["myprovider"] = llm.NewMyProvider(cfg.MyAPIKey)
}
```

---

### 7. Connector Framework (`internal/connectors/`)

Protocol-agnostic tool execution. Routes tool calls to the right backend.

| Connector | Type constant | What it does |
|---|---|---|
| MCP | `TypeMCP` | Executes tools via Model Context Protocol |
| REST | `TypeREST` | Executes tools via HTTP REST |
| SQL | `TypeSQL` | Executes tools via parameterised SQL |
| Core | `TypeCore` | Executes native Go built-in tools |

**Connector interface:**
```go
type Connector interface {
    Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error)
    Validate(req *ExecutionRequest) error
    HealthCheck(ctx context.Context) error
    Type() ConnectorType
}
```

---

### 8. MCP Client (`internal/mcp/`)

Native Model Context Protocol client and tool context compressor.

```go
tools, err := mcpClient.ListTools(ctx)
result, err := mcpClient.ExecuteTool(ctx, toolName, arguments)
```

---

### 9. Tool Registry (`internal/tools/`)

Postgres-backed dynamic tool registry. Syncs to Qdrant every 30 seconds.

**Register a tool via API:**
```bash
curl -X POST http://localhost:8080/v1/tools \
  -H "Authorization: Bearer <admin-jwt>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "get_weather",
    "description": "Get current weather for a city. Input: city name.",
    "connector_type": "rest",
    "endpoint": "https://api.weather.com/v1/current",
    "org_id": "org_456"
  }'
```

---

### 10. Metrics and Audit (`internal/metrics/`)

**Prometheus metrics** exposed at `/metrics`:
- `http_requests_total` — by path, method, status code
- `request_duration_seconds` — histogram per endpoint

**Audit log** at `/v1/audit`:
- In-memory ring buffer (fast, ephemeral)
- Postgres persistence with 30-day retention (`persistent_audit.go`)

---

### 11. gRPC Router Client (`internal/router/`)

Client stub connecting the Go gateway to the Rust semantic router.

The open-source version includes the gRPC stubs and proto-generated code.
The **Rust router service itself** is a commercial component. You can
implement `RouterService` yourself using `proto/router.proto`:

```protobuf
service RouterService {
  rpc RouteQuery     (RouteRequest)     returns (RouteResponse);
  rpc RegisterTool   (RegisterToolRequest) returns (RegisterToolResponse);
  rpc PlanToolChain  (ChainRequest)     returns (ChainResponse);
}
```

Without a router, the gateway falls back to injecting all registered tools
into the LLM context.

---

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.25+ (for local development)
- An API key for at least one LLM provider, or Ollama running locally

### One-Command Stack

```bash
git clone https://github.com/Opsylux/memzent
cd memzent
docker-compose up
```

Services started:
- Gateway → `http://localhost:8080`
- Dashboard → `http://localhost:3000`
- Valkey → `localhost:6379`
- PostgreSQL 16 → `localhost:5432`
- Ollama → `http://localhost:11434`

### Generate a Token

```bash
curl http://localhost:8080/generate-token
# {"token": "eyJ..."}
```

### First Request

```bash
curl -X POST http://localhost:8080/v1/chat \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"messages": [{"role": "user", "content": "Hello, Memzent!"}]}'
```

### Verify Cache Behaviour

```bash
# Run the same query twice
for i in 1 2; do
  curl -si -X POST http://localhost:8080/v1/chat \
    -H "Authorization: Bearer <token>" \
    -H "Content-Type: application/json" \
    -d '{"messages": [{"role": "user", "content": "What is the capital of France?"}]}' \
    | grep X-Cache
done
# X-Cache: MISS   ← first call hits LLM
# X-Cache: HIT    ← second call served in <5ms
```

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `:8080` | Gateway listen port |
| `VALKEY_URL` | `localhost:6379` | Valkey/Redis address |
| `POSTGRES_URL` | — | PostgreSQL connection string |
| `ROUTER_URL` | `localhost:50051` | Rust router gRPC address |
| `MCP_SERVER_URL` | `localhost:50052` | MCP server address |
| `JWT_SECRET` | — | HS256 JWT signing secret |
| `JWKS_URL` | — | JWKS endpoint (Supabase / Auth0) |
| `OLLAMA_URL` | `http://localhost:11434` | Ollama base URL |
| `OLLAMA_MODEL` | `llama3.2` | Default Ollama model |
| `OPENAI_API_KEY` | — | Enables OpenAI provider |
| `OPENAI_MODEL` | `gpt-4o` | Default OpenAI model |
| `ANTHROPIC_API_KEY` | — | Enables Anthropic provider |
| `GEMINI_API_KEY` | — | Enables Gemini provider |
| `LLM_CACHE_TTL` | `1h` | Cache TTL for LLM responses |
| `TOOL_RELEVANCE_THRESHOLD` | `0.88` | Cosine similarity threshold |
| `ENVIRONMENT` | `development` | `production` enables JSON logging |

---

## Architecture

```
Client Request
      │
      ▼
┌──────────────────────────────────────────┐
│        Memzent Gateway  (Go :8080)       │
│                                          │
│  1. Rate Limit  → per-user token bucket  │
│  2. L1 Cache    → Valkey exact hash      │
│  3. L1.5 Cache  → Valkey canonical hash  │
│  4. RBAC        → Postgres + JWT scopes  │
│  5. Tool Route  → gRPC → [Rust Router*]  │
│  6. Execute     → MCP / REST / SQL / Go  │
│  7. LLM Call    → Ollama/OpenAI/Anthropic│
│  8. Cache Write → Valkey + Postgres      │
│  9. Response    → JSON or SSE stream     │
└──────────────────────────────────────────┘

* Rust Router = commercial component
  Gateway operates without it — tool
  selection falls back to all tools.
```

---

## What Is NOT Open Source

| Component | Description |
|---|---|
| `services/router/` | Rust gRPC semantic router — vector math, HNSW, cosine similarity, PlanToolChain |
| L2 Semantic Cache | Qdrant vector similarity search (requires the Rust router) |
| `internal/billing/` | Ledger, CostCalculator, Stripe integration |
| Enterprise RBAC | Fine-grained multi-tenant governance |
| Persistent cache fallback | Postgres write-through for Valkey crash recovery |

Commercial licensing: [memzent.ai/enterprise](https://memzent.ai/enterprise)

---

## Licence

Apache License 2.0 — see [LICENSE](LICENSE)

---

*Memzent.AI — Memory of Agent — [memzent.ai](https://memzent.ai)*
