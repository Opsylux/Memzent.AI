# Changelog

All notable changes to Memzent are documented here.

## [Unreleased]

### Added
- **README**: Complete rewrite with 4-layer cache, Evolution Pipeline, full API reference, test commands
- **PROJECT_STATUS**: Added E1-E6 entries, workflow registry, feature flags, blog/docs/SEO completion
- **Docker**: Healthchecks for dashboard and website, pinned Qdrant to v1.14.0
- **CHANGELOG**: This file

### Fixed
- Removed stale TODO in tools/handlers.go (tool health assumed online)

---

## [0.8.0] — 2026-06-08 (PR #6: Evolution Pipeline)

### Added
- **E1: Entity Extraction** — 6 typed regex extractors (<1ms) with positional awareness (Go + Rust)
- **E2: L1b Entity-Keyed Cache** — Deterministic Valkey key from sorted entity pairs, sub-ms lookups
- **E3: Offline Learning Plane** — Buffered channel event bus (4096/4 workers), 3 miners (Request, Cache, Workflow)
- **E4: Workflow Registry** — Full lifecycle management, API endpoints, dashboard page, engine shortcut
- **E5: GPU Avoidance Metrics** — 8 Prometheus counters, GPU Analytics dashboard
- **E6: Pattern Mining** — Markov chain analysis + speculative pre-warmer (experimental, default off)
- **Feature Flags** — 6 env-var flags controlling all Evolution Pipeline features
- **Integration Tests** — `make test-evolution` with 28 functional assertions
- **Agent Memory Test** — `make test-memory` with 10 session/memory tests
- **Entity Test Suite** — `make test-entity` with 14 cache guard tests
- **Docs**: 4 new pages (entity-extraction, cache-layers, offline-learning, gpu-analytics)
- **Blog**: Evolution Pipeline launch post with SEO metadata
- **Website**: Evolution Pipeline section, updated comparison table, enhanced JSON-LD
- **SEO**: sitemap.xml, robots.txt, generateMetadata on docs/blog, canonical URLs

### Fixed
- SkipCache semantics — now skips reads only, still writes to cache
- N+1 query in workflow shortcut (hoisted ListTools above loop)
- Billing gap — added chargeCacheHit after workflow execution
- Invalid workflow state transition — added status guard
- Cache flush scope — added `?scope=` param (valkey|db|all)
- Stats org isolation — strict WHERE clause, no nil-UUID system events leak
- Dashboard responsive — mobile hamburger menu, slide-in drawer
- Sign-out resilience — client-side fallback when server action fails
- PKCE auth callback — graceful redirect on missing code verifier

---

## [0.7.0] — 2026-06-01

### Added
- Webhook notification pipeline (6 event types, HMAC signing, retry with dead letter)
- Spend limits & budget forecast API (daily/monthly dollar + token caps)
- Per-user role-proportional rate limiting
- API key rotation with 15-min grace window
- Blog system (MDX + Supabase dual source)
- Documentation site (15 pages)

---

## [0.6.0] — 2026-05-15

### Added
- Agent session memory (PostgreSQL + Qdrant semantic extraction)
- Context analytics (ROI tracking, latency telemetry, intent clustering)
- Sequential tool chaining (PlanToolChain gRPC)
- SSE typewriter streaming for /v1/chat
- Model-specific cache scoping

---

## [0.5.0] — 2026-05-01

### Added
- Triple-layer semantic caching (L1 literal, L1.5 canonical, L2 vector)
- Multi-provider LLM routing (Ollama, OpenAI, Anthropic, Gemini)
- Dynamic tool registry with Qdrant sync
- RBAC with JWT + API key auth
- Prometheus metrics at /metrics
- Marketing website with SEO
- Admin dashboard with billing, playground, tools CRUD
