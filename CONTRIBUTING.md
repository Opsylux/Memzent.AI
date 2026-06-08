# Contributing to Memzent

Thank you for your interest in contributing to Memzent! This is an
open-source project licensed under Apache 2.0, and we welcome
contributions from the community.

---

## What You Can Contribute To

**Everything in this repository is open source.** All services, all code:

- `services/gateway/` — Go gateway (HTTP, auth, caching, orchestration)
- `services/router/` — Rust semantic router (embeddings, vector search, tool matching)
- `services/dashboard/` — Next.js admin dashboard
- `services/mcp-server/` — MCP protocol adapter
- `website/` — Marketing site
- `proto/router.proto` — gRPC contract
- `migrations/` — Database schemas
- `docker-compose.yml` — Local stack configuration
- Documentation, tests, and scripts

---

## Architecture Rules — Read Before Writing Code

These rules are enforced on every PR. Violating them means your PR will
be rejected regardless of code quality.

### Rule 1 — No vector math in Go

All embedding generation and cosine similarity computation goes through
the Rust router via gRPC. The Go gateway calls the router — it never
computes vector similarity itself.

```go
// ✅ CORRECT — call the router
result, err := rClient.RouteQuery(ctx, prompt, orgID)

// ❌ WRONG — never do this in Go
similarity := dotProduct(vecA, vecB) / (magnitude(vecA) * magnitude(vecB))
```

### Rule 2 — No HTTP or auth logic in Rust

The Rust service handles gRPC only. It does not validate JWTs, make HTTP
requests to LLM providers, or connect to Postgres.

### Rule 3 — All SQL must be parameterised

Never concatenate user input into SQL strings. This is a security
requirement, not a style preference.

```go
// ✅ CORRECT
row := db.QueryRowContext(ctx,
    "SELECT balance FROM orgs WHERE id = $1", orgID)

// ❌ WRONG — SQL injection vulnerability
row := db.QueryRowContext(ctx,
    "SELECT balance FROM orgs WHERE id = '" + orgID + "'")
```

### Rule 4 — All errors must be logged with context

Never swallow errors silently. Every error must be logged with the
operation name, relevant IDs, and the error message.

```go
// ✅ CORRECT
if err != nil {
    slog.Error("Cache write failed",
        "operation", "SetResult",
        "org_id", orgID,
        "error", err)
    return err
}

// ❌ WRONG
if err != nil {
    return err  // no context, impossible to debug
}
```

### Rule 5 — All external calls must have timeouts

Every call to an external service (LLM, gRPC router, Qdrant, Postgres)
must have an explicit timeout or context deadline.

```go
// ✅ CORRECT
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()
resp, _, err := llmProvider.Generate(ctx, messages, nil, "")

// ❌ WRONG — can hang forever
resp, _, err := llmProvider.Generate(context.Background(), messages, nil, "")
```

### Rule 6 — No new dependencies without discussion

Open a GitHub issue before adding a new Go module or npm package. Keep
the dependency footprint minimal.

---

## Development Setup

### Prerequisites

- Go 1.25+
- Rust (stable toolchain) + `cargo`
- Node.js 22+ and Bun (for the dashboard)
- Docker and Docker Compose
- `golangci-lint` (Go linter)

### Clone and Run

```bash
git clone https://github.com/Opsylux/memzent
cd memzent

# Start the full stack
docker-compose up -d

# Run the gateway locally (hot reload)
cd services/gateway
go run main.go

# Run the dashboard locally
cd services/dashboard
bun install
bun dev
```

### Run Tests

```bash
# Gateway unit tests
cd services/gateway
go test ./...

# With race detector
go test -race ./...

# A specific package
go test ./internal/engine/...
```

### Lint

```bash
cd services/gateway
golangci-lint run ./...
```

---

## Making a Contribution

### 1. Find or create an issue

All contributions start with a GitHub issue. Check existing issues before
opening a new one. For bug fixes, link your PR to the issue.

Good first issues are labelled `good-first-issue`.

### 2. Fork and branch

```bash
git checkout -b fix/cache-stampede-on-cold-start
# or
git checkout -b feat/add-groq-provider
```

Branch naming convention:
- `fix/` — bug fixes
- `feat/` — new features
- `docs/` — documentation only
- `test/` — test additions or fixes
- `refactor/` — code changes with no behaviour change

### 3. Write tests

All new code requires tests. PRs without tests for new behaviour will
not be merged.

```bash
# Test file convention
internal/engine/normalization_test.go  ← tests for normalization.go
internal/cache/valkey_test.go          ← tests for valkey.go
```

### 4. Commit messages

Follow conventional commits:

```
feat(engine): add singleflight pattern for cache stampede prevention
fix(auth): reject expired JWTs with 401 not 500
docs(readme): add Groq provider setup instructions
test(cache): add concurrent write race condition test
```

### 5. Open a pull request

- Target the `main` branch
- Fill in the PR template completely
- Link the related issue
- Ensure all CI checks pass before requesting review

---

## PR Review Checklist

Every PR is reviewed against this checklist:

- [ ] No vector math in Go code
- [ ] No hardcoded secrets or API keys
- [ ] All SQL queries use parameterised placeholders
- [ ] All external calls have context timeouts
- [ ] All errors logged with operation context
- [ ] New functionality has unit tests
- [ ] `go test -race ./...` passes
- [ ] `golangci-lint run ./...` passes with no new errors

---

## Adding a New LLM Provider

The most common contribution. Here is the exact process:

1. Create `services/gateway/internal/llm/myprovider.go`

```go
package llm

import "context"

type MyProvider struct {
    apiKey string
    model  string
}

func NewMyProvider(apiKey, model string) *MyProvider {
    if model == "" {
        model = "my-default-model"
    }
    return &MyProvider{apiKey: apiKey, model: model}
}

func (p *MyProvider) GetProviderName() string { return "myprovider" }

func (p *MyProvider) GetMetadata() ProviderMetadata {
    return ProviderMetadata{Name: "myprovider", DefaultModel: p.model}
}

func (p *MyProvider) Generate(ctx context.Context,
    messages []Message, tools []any, model string) (string, *TokenUsage, error) {
    // implement your provider API call here
    // return: response text, token usage, error
}
```

2. Add config in `services/gateway/internal/config/config.go`:

```go
MyProviderAPIKey string
MyProviderModel  string
```

3. Register in `services/gateway/main.go`:

```go
if cfg.MyProviderAPIKey != "" {
    providers["myprovider"] = llm.NewMyProvider(
        cfg.MyProviderAPIKey, cfg.MyProviderModel)
    slog.Info("Provider registered: MyProvider")
}
```

4. Add env var documentation to the README

5. Write tests in `internal/llm/myprovider_test.go`

6. Open a PR

---

## Adding a New Connector

To add a new tool execution protocol:

1. Create `services/gateway/internal/connectors/myprovider.go`
   implementing the `Connector` interface
2. Add `TypeMyConnector` to the `ConnectorType` constants in `connector.go`
3. Register in `main.go`
4. Document in the README
5. Write tests

---

## Code Style

- `gofmt` and `goimports` — enforced by CI
- `golangci-lint` — enforced by CI
- Structured logging only — use `slog.Info()`, `slog.Error()` etc.
  Never use `fmt.Println` or `log.Printf` in production code
- Context propagation — always pass `ctx` as the first argument
- No global state — pass dependencies explicitly via constructors

---

## Questions

Open a GitHub Discussion for questions. Do not open issues for general
questions — issues are for bugs and feature requests only.

---

*Memzent.AI — Memory of Agent — [memzent.ai](https://memzent.ai)*
