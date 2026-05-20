# Memzent.AI — Strategic Proposal | Confidential

**Product:** MEMZENT.AI (Memory of Agent)  
**Document:** Strategic Proposal — Manager Briefing & Forward Roadmap  
**Prepared by:** Memzent Engineering Team  
**Date:** May 2026  
**Classification:** Confidential — Internal Use Only  
**Version:** 2.1  
**Domain:** [memzent.ai](https://memzent.ai)  

---

## 1. Executive Summary

Memzent.AI is an enterprise-grade Agentic Infrastructure Platform that provides the missing memory, security, and intelligence layer for autonomous AI workflows.

It operates as an **Intelligent Semantic Proxy** between clients, MCP tools, and LLM providers — intercepting every request, caching with semantic precision, enforcing role-based access control, routing only relevant tools into context, and billing per-token with full audit visibility.

The platform has completed five major development phases, including a triple-layer semantic cache, multi-connector tool framework (MCP, REST, SQL), model-scoped caching, multi-step tool chaining, SSE streaming, Stripe-integrated billing, tiered rate limiting, and a persistent cache fallback for infrastructure resilience. It is production-ready pending one final hardening step: Qdrant vector database optimization.

> **Core Value Proposition:** Memzent eliminates the *Token Tax* — the hidden cost of unmanaged LLM deployments — by ensuring that no semantically equivalent query ever hits an LLM twice. Combined with intelligent tool routing that reduces context window bloat by up to 90%, Memzent delivers faster responses at a fraction of the cost, with enterprise governance built in.

This document provides the complete strategic picture: what Memzent is, how it differs from every competitor, what has been built, what remains, and the exact steps required to turn this into an outstanding, market-defining product.

---

## 2. What Is Memzent.AI?

### 2.1 The Problem
Enterprises deploying AI agents at scale face five converging problems that no single product solves today:

* **Token waste:** Semantically identical queries pay full LLM cost every time because no caching layer understands meaning.
* **No governance:** Any user can invoke any tool with any LLM, with no access control, audit trail, or budget accountability.
* **Tool context explosion:** Every registered tool is injected into the LLM context regardless of relevance, bloating token usage and increasing hallucination risk.
* **Provider lock-in:** Switching between OpenAI, Anthropic, Gemini, or local models requires code changes across every consuming service.
* **Zero resilience:** If the cache layer crashes, everything falls back to full-price LLM calls with no recovery mechanism.

### 2.2 How Memzent Solves It
Every AI request passes through a nine-step intelligent pipeline:

| Step | Action | Technology | Outcome |
| :--- | :--- | :--- | :--- |
| **1** | Rate limiting | Token bucket, Go | Prevents abuse per user/tier |
| **2** | Billing check | Ledger + Postgres | Rejects requests from depleted orgs |
| **3** | L1 Literal cache | SHA-256 hash, Valkey | Exact match served in < 5 ms |
| **4** | L1.5 Canonical cache | Normalised hash, Valkey | Logically identical queries matched in < 5 ms |
| **5** | L2 Semantic cache | Cosine similarity, Qdrant | Semantically similar queries matched in 10–30 ms |
| **6** | RBAC enforcement | JWT + PostgreSQL 16 + scoped tokens | Only authorised tools/actions for this identity |
| **7** | Semantic tool routing | Rust gRPC, Qdrant vectors | Only 1–2 relevant tools injected into context |
| **8** | LLM synthesis + SSE streaming | Ollama / OpenAI / Anthropic / Gemini | Typewriter streaming response to client |
| **9** | Cache populate + cost deduct | Valkey + Postgres persistent backup | All 3 cache layers filled; cost deducted from balance |

**Key Insight:** Steps 2 through 5 mean the LLM is never called for a query that has been asked before in any semantically equivalent form, and depleted accounts are blocked before any compute is consumed. Step 9's persistent cache backup means even a Valkey crash cannot destroy cached responses.

### 2.3 The Triple-Layer Semantic Cache
The cache hierarchy is the most differentiated technical component:

* **L1 — Literal:** SHA-256 hash of the raw prompt. Sub-5ms response time. Serves byte-identical repeated queries instantly.
* **L1.5 — Corporate/Canonical:** Numeric identifiers masked to stable placeholders (`write011` and `write202` both become `write<ID>`), then hashed. Catches the enterprise pattern of logically identical queries with different record IDs. No competitor implements this.
* **L2 — Semantic:** Vector embedding via the Rust router, cosine similarity search in Qdrant at `>= 0.88` threshold. **Model-scoped:** GPT-4o and Claude responses are cached separately, preventing cross-model contamination.
* **Persistent Fallback:** All cache entries are write-through to a Postgres `persistent_cache` table. If Valkey crashes, the gateway reads from Postgres and backfills Valkey in the background. Zero cache loss.

---

## 3. What Makes Memzent Outstanding

Seven characteristics differentiate Memzent from every current competitor. Four are technical, three are strategic.

### 3.1 MCP-Native from Day One
The Model Context Protocol (MCP), introduced by Anthropic in November 2024 and adopted by OpenAI in March 2025, is the emerging standard for agent-to-tool communication. Memzent is built around MCP as a first-class citizen — not a bolt-on. The gateway acts as both MCP server and client, and the semantic router natively understands MCP tool descriptions as vector embeddings.

Every competitor (LiteLLM, Portkey, Kong) is retrofitting MCP into architectures designed before MCP existed. This gives us a 12-month window before major players close the gap.

### 3.2 Rust gRPC Semantic Router — The Brain
Vector mathematics runs in a dedicated Rust service via Tonic gRPC — not in the Go gateway, and not in Python. Strict service boundaries ensure no vector math leaks into business logic and no HTTP/auth logic enters the Rust layer. This enables ultra-fast 10–30ms semantic lookups that feel synchronous.

### 3.3 Three-Layer Cache with Persistent Fallback
Competitors that implement semantic caching (Portkey, Helicone) use a single vector similarity layer with no crash recovery. Memzent's three-layer hierarchy with canonical normalisation, model-scoped keys, and a Postgres write-through backup is architecturally unique.

### 3.4 Multi-Step Tool Chaining (PlanToolChain)
The Rust router implements a fully operational `PlanToolChain` gRPC method that computes multi-step tool sequences for complex prompts. Agents can chain tool A's output into tool B's input without manual wiring — this is live and callable today. Competitors handle single tool calls; Memzent handles full workflows.

### 3.5 Built-In Billing and Cost Transparency
Every request is priced in real-time using per-model cost rates (OpenAI, Anthropic, Gemini, or Ollama infrastructure costs). Token balances are checked before LLM invocation, costs are deducted after, and every transaction is audit-logged. Stripe integration handles payments. No self-hosted competitor includes billing natively at the gateway layer — they delegate to external billing systems or hosted SaaS platforms.

### 3.6 Universal Connector Framework
Phase 3 connectors are fully implemented: MCP, REST, and SQL connectors are active and registered. This means Memzent can execute tools over any protocol — not just MCP. GraphQL, gRPC, and webhook connectors are planned.

### 3.7 Agentic Runtime Positioning (Strategic)
Memzent should not compete in the crowded AI Gateway category. The correct positioning is **Agentic Runtime**: the infrastructure layer that governs, caches, routes, bills, and observes every action taken by AI agents in production. This category is being defined in 2025–2026 and no incumbent owns it yet.

---

## 4. Competitive Landscape

Scored 1–5 across 14 capabilities (5 = best in class).

| Capability | Memzent | Portkey | LiteLLM | OpenRouter | Kong AI | Helicone | Leader |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :--- |
| **Semantic caching** | ★★★★★ | ★★★★☆ | ★★☆☆☆ | ★★☆☆☆ | ★★★☆☆ | ★★★☆☆ | **Memzent** |
| **Cache resilience** | ★★★★★ | ★★★☆☆ | ★★☆☆☆ | ★☆☆☆☆ | ★★★☆☆ | ★★☆☆☆ | **Memzent** |
| **MCP native** | ★★★★★ | ★★☆☆• | ★★☆☆☆ | ★☆☆☆☆ | ★☆☆☆☆ | ★☆☆☆☆ | **Memzent** |
| **Tool routing / selection** | ★★★★★ | ★★☆☆☆ | ★☆☆☆☆ | ★☆☆☆☆ | ★★☆☆☆ | ★☆☆☆☆ | **Memzent** |
| **Multi-step tool chaining** | ★★★★☆ | ★☆☆☆☆ | ★☆☆☆☆ | ★☆☆☆☆ | ★☆☆☆☆ | ★☆☆☆☆ | **Memzent** |
| **SSE streaming** | ★★★★☆ | ★★★★☆ | ★★★☆☆ | ★★★★☆ | ★★★☆☆ | ★★☆☆☆ | **Memzent / Portkey** |
| **Built-in billing** | ★★★★☆ | ★★☆☆☆ | ★★☆☆☆ | ★★★★☆ | ★★☆☆☆ | ★★☆☆☆ | **Memzent / OpenRouter** |
| **Observability** | ★★★☆☆ | ★★★★★ | ★★★☆☆ | ★★☆☆☆ | ★★★★☆ | ★★★★★ | **Portkey** |
| **Guardrails / PII** | ★★☆☆☆ | ★★★★★ | ★★★☆☆ | ★☆☆☆☆ | ★★★★☆ | ★★☆☆☆ | **Portkey** |
| **Multi-provider routing** | ★★★★☆ | ★★★★★ | ★★★★★ | ★★★★★ | ★★★★☆ | ★★★☆☆ | **Tie** |
| **Enterprise RBAC** | ★★★★★ | ★★★★☆ | ★★★☆☆ | ★★☆☆☆ | ★★★★★ | ★★☆☆☆ | **Memzent / Kong** |
| **Connector breadth** | ★★★★☆ | ★★★☆☆ | ★★★☆☆ | ★★☆☆☆ | ★★★☆☆ | ★★☆☆☆ | **Memzent** |
| **Performance / latency** | ★★★★★ | ★★★★☆ | ★★★☆☆ | ★★★★☆ | ★★★★★ | ★★★☆☆ | **Memzent / Kong** |
| **Open-source community** | ★☆☆☆☆ | ★★★★☆ | ★★★★★ | ★★★☆☆ | ★★★☆☆ | ★★★☆☆ | **LiteLLM** |

**Summary:** Memzent leads in 9 of 14 capabilities — more than any competitor. The two gaps are observability/guardrails (where Portkey leads) and open-source community (where LiteLLM leads). Both are highly addressable.

---

## 5. What Each Competitor Does Differently

### 5.1 Portkey — The Primary Threat
* **Funding:** $18M total ($15M Series A, Feb 2026 — Elevation Capital + Lightspeed)
* **Scale:** 500B+ tokens/day, 125M+ requests/day, 24,000 organisations
* **Strength:** Observability, guardrails, PII redaction, semantic caching, open-sourced Feb 2026.
* **Weakness:** LLM-request-centric (not agent-session-centric); MCP support is retrofitted late; no built-in billing; no multi-tier cache resilience layer.
* **How Memzent Wins:** MCP-native architecture, multi-step tool chaining, 3-layer cache with persistent fallback, built-in billing, and model-scoped cache keys.

### 5.2 LiteLLM — The Community Darling
* **Funding:** $2.1M seed (YC W24)
* **GitHub Stars:** 47,300+ — largest open-source AI gateway.
* **Strength:** 100+ LLM providers, massive community, simple Python SDK.
* **Weakness:** No semantic caching at all; no MCP support; monolithic Python codebase; no intelligent tool routing.
* **How Memzent Wins:** Different category entirely. Memzent serves enterprise teams needing infrastructure governance; LiteLLM serves developers wanting basic model portability. We do not compete directly for this developer sandbox audience.

### 5.3 OpenRouter — The Marketplace Unicorn
* **Funding:** $168M total. $1.3B valuation (CapitalG / Google, April 2026). $50M+ ARR.
* **Strength:** 400+ models, zero setup, automatic provider failover.
* **Weakness:** No caching layers, no RBAC, no tool routing, 5.5% platform fee, no local data residency.
* **How Memzent Wins:** Different category entirely. OpenRouter is an endpoint marketplace; Memzent is internal infrastructure. OpenRouter's massive valuation proves how large the underlying compute market is.

### 5.4 Kong AI Gateway — The Enterprise Incumbent
* **Funding:** $171M total (parent company, Series D)
* **Strength:** Proven enterprise reliability, massive plugin ecosystem, benchmarked 228% faster than Portkey.
* **Weakness:** Heavy operational overhead, no semantic caching, no native MCP, enterprise features heavily paywalled.
* **How Memzent Wins:** Lighter, MCP-native, semantics-first. Differentiate on intelligence and optimization, not plugin volume.

### 5.5 Helicone — Complementary, Not Competitive
* **Funding:** Angel / YC (limited)
* **Strength:** Best-in-class LLM observability — one-line integration, cost tracking, alerts.
* **Weakness:** Observability-only — does not route, cache, or govern.
* **Opportunity:** Memzent should build Helicone-style agent-trajectory observability natively. This represents an excellent acquisition or partnership opportunity later.

---

## 6. Current Project Status

Memzent has completed five major development phases. One hardening step remains before production deployment.

| Phase | Feature | Status | % | Notes |
| :---: | :--- | :---: | :---: | :--- |
| **1** | Core — cache, RBAC, routing, MCP | Complete | 100% | Triple-layer cache, JWT auth, multi-provider LLM, MCP integration, Rust router |
| **1a** | Rate limiting by tier | Complete | 100% | Implemented in `engine.go`. Tier-scoped limits: free 10/min, pro 100/min, business 1,000/min. Reads tier from JWT claims. |
| **1b** | Qdrant optimisation (quantisation, snapshots) | **Not Started** | 0% | **Go-live blocker.** RAM cost reduction + disaster recovery. Est. 1–2 weeks. |
| **2** | Dynamic Tool Registry + 30s Qdrant sync | Complete | 100% | Zero-downtime tool registration. Postgres + Qdrant refresh loop. |
| **3** | Multi-connector framework (MCP, REST, SQL) | Complete | 100% | All connector types active and registered. |
| **4** | Advanced orchestration | Complete | 100% | Model-scoped cache, `PlanToolChain` gRPC (live), SSE typewriter streaming. |
| **—** | Billing (Ledger, CostCalc, Stripe, $10 trial) | Complete | 100% | Per-model pricing, balance enforcement, Stripe webhooks. |
| **—** | Persistent cache fallback | Complete | 100% | Postgres write-through. Auto-backfill on Valkey crash. |
| **5** | Kubernetes / Envoy HA | Planned | 0% | K8s deploy, Envoy gRPC LB, retry policies. Post go-live. |

> **Go-Live Estimate:** Phase 1b (Qdrant hardening) is the only remaining blocker. This is an isolated engineering task with no architectural dependencies. Estimated timeline: **1–2 weeks of focused development.**

---

## 7. Forward Roadmap — How to Make Memzent Outstanding

The improvements below are sequenced by defensibility and impact. Each one either creates a moat competitors cannot quickly copy, or closes a gap that would otherwise block enterprise adoption.

### 7.1 Immediate (Weeks 1–2): Ship Go-Live Blocker + Open-Source

#### A. Tiered Rate Limiting (Phase 1a) — `✅ COMPLETE`
* Tier-scoped limits are live in `engine.go`: free 10/min, pro 100/min, business 1,000/min.
* Tier is read from JWT claims already populated by the RBAC system. No new infrastructure required.
* *Known hardening item:* the in-memory rate limiter map has no TTL eviction. For multi-tenant scale, replace with an LRU-bounded cache to prevent memory footprint growth.

#### B. Qdrant Production Hardening (Phase 1b)
* Enable scalar quantisation — reduces Qdrant RAM footprint by 75% with negligible recall loss.
* Set `memmap` threshold — cold data offloads to disk automatically.
* Index `org_id` and `user_id` payload fields for fast, isolated filtered search performance.
* Schedule S3 snapshots for automated disaster recovery.
* *Estimated effort:* 1 week. Qdrant configuration updates + migration script.

#### C. Open-Source Under Apache 2.0
* Push the gateway core to GitHub as a public repo under the Memzent.AI brand. The repo is currently named `AuraMCP` — rename to `memzent` before publishing.
* **Strategic Boundary:** Open-source the gateway core, Rust router, and MCP server adapter. Keep the billing ledger package, federated cache pooling, and custom ML router as commercial enterprise modules (under a BSL or AGPL licence). Define this split with build tags or separate repos before the push — the billing package currently lives in the same workspace tree.
* Write a robust README with system architecture diagrams, one-command Docker Compose deployment, and rich API examples.
* Post to Hacker News (*Show HN*), *r/LocalLLaMA*, *r/MachineLearning*, and Anthropic's *MCP Discord*.
* **Target:** 500 GitHub stars in 60 days. Based on comparable Show HN launches in this category, 200–500 first-week stars is highly achievable with an explicit, execution-focused README and a short demo video.

> **Why this is urgent:** LiteLLM has amassed 47,000 stars on a $2.1M seed. Portkey open-sourced in February 2026. The window for building community-driven distribution is closing rapidly. Every month of delay is a month where incumbent mindshare grows unchallenged.

### 7.2 Short-Term (Months 2–4): The Wedge Features

#### D. Agent Trajectory Replay — The Headline Demo
When an AI agent makes 47 sequential tool calls and produces an incorrect answer, engineers cannot easily explain why. Current observability tools log individual flat LLM requests but fail to map complex agent decision graphs.
* Store every agent session as a directed graph: `prompt -> tool selection -> tool result -> next prompt -> final output`.
* Inject `session_id` and `parent_call_id` headers into every request. Store efficiently in Postgres + ClickHouse.
* Build a visual timeline UI in the dashboard: replay, diff across multiple runs, and annotate failures.
* *Why defensible:* Requires an MCP-native architecture to view the full tool execution graph. Portkey and LiteLLM are fundamentally LLM-request-centric — they cannot build this without severe structural re-engineering.

**Business Impact:** This is a highly visual, investor-grade capability. Walking into a VC meeting and demoing a live agent decision graph being replayed and diffed secures a seed round.

#### E. Predictive Tool Pre-Fetching
Agents currently wait for the LLM to complete its reasoning and explicitly emit the next tool call before executing it. Each sequential round-trip adds 200–800ms of friction.
* Extend the Rust router to predict the next 1–2 likely tool calls based on the active session trajectory vector and pre-warm backend tool connections in parallel.
* When the LLM actually requests them, execution has already started.
* *Why defensible:* Requires the combination of semantic routing + unified tool registry + session state that only Memzent has as a single cohesive system.
* **Marketing Headline:** *"Memzent reduces agent runtime latency by 40%."*

#### F. Trust-and-Safety at the Router Layer
Portkey has guardrails, but they evaluate *after* the LLM responds. Memzent can do better.
* Classify user intent inside the high-performance Rust semantic router *before* the downstream LLM is invoked: checking for PII exfiltration, prompt injection, or tool misuse.
* Block instantly at the edge router layer. True prevention, not post-incident auditing.
* *Why defensible:* Requires a Rust-based semantic engine directly in the synchronous request path. No competitor possesses this.
* **Positioning:** *"Other gateways audit what already happened. Memzent prevents what should not."*

### 7.3 Medium-Term (Months 5–8): The Moat

#### G. Cost-Quality ML Router
* Score every request based on: (a) model used, (b) tools involved, and (c) session continuation length as a proxy for answer quality.
* Train a per-organisation model predicting: *"For queries like this, Model X delivers 95% of Model Y quality at 20% of the cost."*
* *Why defensible:* Creates an immense switching-cost moat. The longer an enterprise relies on Memzent, the smarter their internal routing becomes. Leaving means losing months of learned cost optimization.
* **Pitch:** *"Most gateways route. Memzent learns."*

#### H. MCP Tool Marketplace
* Position Memzent as the trusted registry for community MCP tools. Automatically scan, benchmark, and rate security for every tool. Enable one-click installations directly from the dashboard.
* Take a platform revenue cut on paid enterprise tools. Build the *npm for MCP*.
* *Why defensible:* Standard marketplace network effects. Whoever aggregates developer tool adoption first wins permanently.

### 7.4 Long-Term (Months 9–12): The Network Effect

#### I. Federated Semantic Cache
Every enterprise currently caches independently. Two separate companies asking the exact same factual public data question each pay full downstream LLM costs.
* Build an opt-in shared cache pool for non-sensitive query classes (e.g., standard code snippets, public documentation, open facts) classified automatically by the semantic router.
* Every new enterprise customer onboarding automatically increases the cache hit rate for all other customers.
* *Why defensible:* A genuine network-effect moat — the exact type that supports multi-billion dollar valuations in infrastructure software.
* *Privacy:* Bound strictly to public-domain query classes. Never cache PII. Classify ruthlessly at the router boundary.

> **Strategic Note:** This single feature, if executed with privacy-safe classification, is what separates a $50M acquisition from a $5B market leader. It requires customer scale to activate, which is why it sits at months 9–12.

---

## 8. 12-Month Execution Plan

| Period | Goal | Key Deliverables | Success Metric |
| :--- | :--- | :--- | :--- |
| **Weeks 1–4** | Go-live + open-source | Ship 1a (rate limiting) + 1b (Qdrant). Open-source gateway core under Apache 2.0. Show HN launch post. Interview 30 prospective users. | 500 GitHub stars. Production-hardened system. |
| **Months 2–3** | Position + first revenue | Reposition as Agentic Runtime. Rewrite landing page messaging. Onboard 10 paying customers at $200–500/month. | $3K–5K MRR. Clear messaging fit. |
| **Months 3–5** | Wedge features | Complete agent trajectory replay UI. Implement predictive tool pre-fetching. Deploy trust-and-safety router classifier. | Trajectory replay demo ready. 40% latency reduction measured. |
| **Months 5–8** | Moat + funding | Ship Cost-quality ML router MVP. Open MCP marketplace beta. Apply to YC or raise $2–4M institutional seed round. | $20K MRR. Seed round term sheets initiated. |
| **Months 9–12** | Network effect + scale | Deploy federated cache pool beta. Secure 3+ enterprise contracts. Initiate SOC 2 Type I compliance audit. | $50K+ MRR. Series A growth story locked. |

---

## 9. Decision Milestones for Management

The project must be measured against clear, quantitative milestones at three distinct checkpoints:

| Metric | 3-Month Target | 6-Month Target | 12-Month Target |
| :--- | :---: | :---: | :---: |
| **GitHub Stars** | 500+ | 2,000+ | 5,000+ |
| **Monthly Revenue (MRR)** | $3,000+ | $20,000+ | $50,000+ |
| **Paying Customers** | 10+ | 30+ | 100+ |
| **Enterprise Pilots** | — | 2+ | 5+ |
| **Funding Status** | User interviews complete | Seed raise started | Seed closed / YC batch |
| **Team Size** | Solo + 1 DevRel hire | 3–4 (Eng + GTM) | 6–8 (Full Scale Team) |
| **Key Feature Shipped** | Open-source + trajectory replay | ML router + marketplace beta | Federated cache beta |

> **Kill Criteria:** At the 6-month mark, if MRR sits below $10,000 and GitHub stars are under 1,000, the current positioning has failed to capture the market. A definitive decision must be made — either pivot the target buyer entirely or wind down operations. Revenue is the only signal that matters.

---

## 10. Recommendation

### 10.1 Continue and Accelerate — With Immediate Focus on Distribution
Memzent is technically one of the most complete products in this entire category. The architecture is sound, four core development phases are completely shipped, and the MCP-native edge is highly defensible. The critical gap is not the technology — it is distribution. We currently sit at zero GitHub stars, zero customers, and zero institutional funding. The code is outstanding; the world simply does not know it exists yet.

### 10.2 The Five Actions That Must Happen This Month
1. **Open-source the gateway core** under the Apache 2.0 license — *this week*.
2. **Ship Phase 1a** (LRU rate limiting fix) and **Phase 1b** (Qdrant quantization config) — *within two weeks*.
3. **Launch publicly** on Hacker News, r/LocalLLaMA, and the Anthropic MCP Discord — *within 48 hours of open-sourcing*.
4. **Begin 30 user interviews** with production engineering teams building AI agents — *this month*.
5. **Archive the old name (`AuraMCP`).** One project, one unified brand, one repository. `Memzent.AI` is the product.

### 10.3 The One Sentence for Every Conversation
> **Positioning:** Memzent is the agentic runtime that remembers, routes, and governs every action your AI agents take — cutting token costs by up to 90% while making agents faster, safer, and completely auditable.

### 10.4 Probability Assessment (Honest Market Baseline)

| Outcome | On Current Path | With Pivot + Execution |
| :--- | :---: | :---: |
| **Multi-billion dollar company** | < 3% | 5–10% |
| **Meaningful exit ($20–100M)** | 10–15% | 30–45% |
| **Sustainable business ($1–5M ARR)** | 30–40% | 60–75% |
| **Acquisition target for Portkey/Kong/Anthropic** | 20–30% | 40–50% |

*Note on probabilities: These figures reflect typical base rates for bootstrapped infrastructure startups competing in venture-backed, fast-moving categories. Portkey, LiteLLM, and OpenRouter all shared comparable odds at their equivalent repre-launch stage. The variance lies strictly in execution velocity and distribution focus — both of which are entirely within our control.*

---

## Appendix A — Technical Reference

### A.1 Service Architecture

                   +-----------------------+
                   |   Client API Request  |
                   +-----------+-----------+
                               |
                               v
                   +-----------+-----------+
                   |      Go Gateway       | <---> Valkey 8 (L1/L1.5 Cache)
                   |     (Port :8080)      | <---> PostgreSQL 16 (RBAC/Ledger)
                   +-----------+-----------+
                               |
                           | (gRPC)
                               v
                   +-----------+-----------+
                   |  Rust Semantic Router | <---> Qdrant (L2 Cache/Vectors)
                   |     (Port :50051)     |
                   +-----------+-----------+
                               |
                               v
                   +-----------+-----------+
                   |    MCP Server/Tools   |
                   |     (Port :50052)     |
                   +-----------------------+



| Service | Language | Port | Role |
| :--- | :--- | :---: | :--- |
| **Go Gateway** | Go 1.25 | `:8080` | Entry point — authentication, RBAC, billing enforcement, caching, provider routing, SSE streaming. |
| **Rust Router** | Rust (Tonic gRPC) | `:50051` | Vector math, cosine similarity thresholds, tool selection optimization, prompt compression, `PlanToolChain`. |
| **MCP Server** | Go | `:50052` | MCP protocol adapter — tool execution bridge to environment connectors. |
| **Dashboard** | Next.js 15 | `:3000` | Admin tower — billing UI, playground console, API metrics, tool registry documentation. |
| **Website** | Vite / React 19 | `:5173` | Marketing landing page — PAYG billing calculator and explainer. |
| **Valkey** | Valkey 8 | `:6379` | L1 (Literal) + L1.5 (Canonical) cache layers operating fully in-memory. |
| **Qdrant** | Qdrant | `:6333` | L2 semantic cache storage + tool definition embeddings. |
| **PostgreSQL 16** | Postgres | `:5432` | Identity RBAC, billing transaction ledger, tool registry state, audit logs, persistent write-through cache fallback. |
| **Ollama** | LYaMA / Meta | `:11434` | Default local LLM node (disabled by default in production `docker-compose.yaml`). |

### A.2 API Endpoints

| Endpoint | Method | Description |
| :--- | :---: | :--- |
| `POST /v1/chat` | POST | Primary inference endpoint — handles SSE streaming, billing checks, and the full multi-layer cache pipeline. |
| `GET /v1/tools` | GET | List all registered tools with their specific connector types and system metadata. |
| `GET /v1/stats` | GET | Real-time gateway stats — tracking cache hit rates, active provider count, and container uptime. |
| `GET /v1/providers`| GET | List all active downstream inference providers and available models. |
| `POST /v1/tools/register` | POST | Admin Endpoint — register an entirely new tool or database connector in the dynamic registry. |
| `GET /v1/tools/status` | GET | Check last tool registry synchronization timestamp and cluster health. |
| `GET /v1/tools/sync` | GET | Trigger a manual tool registry refresh and synchronize embeddings with Qdrant. |
| `GET /v1/healthz` | GET | Liveness probe for orchestration health checks. |
| `GET /v1/readyz` | GET | Readiness probe checking upstream connections (Valkey, Rust Router, Postgres). |
| `GET /metrics` | GET | Standard Prometheus metrics endpoint for cluster monitoring. |

***
*End of document — Memzent.AI Strategic Proposal v2.1 — May 2026* **Memzent.AI | Memory of Agent | memzent.ai**