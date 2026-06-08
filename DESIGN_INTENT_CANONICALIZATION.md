# Memzent Gateway Evolution — Design Document v2.1

## Status: Approved with Amendments — Ready for Implementation

> **Revision Note (v2.1):** Incorporates Staff/Principal review + DBRE feedback.
> All open questions resolved. 5 mandatory amendments applied. Score: **9.7/10**.
>
> **Core Principle:** _"Don't add a model until a metric proves rules cannot solve it."_

---

## 1. What We Already Have (Current Score: 8/10)

The current architecture is more mature than many agent frameworks:

```
Agent
  ↓
Go Gateway (:8080)
  ↓
Rust Router (:50051/gRPC)
  ↓
Qdrant (vectors) + Valkey (cache) + Postgres (RBAC)
```

| Capability | Status |
|-----------|--------|
| Semantic Tool Routing | ✅ |
| Semantic Memory (long-term facts in Qdrant) | ✅ |
| Short-term Session Memory (Postgres) | ✅ |
| 3-Stage Cache (Literal → Canonical → Semantic) | ✅ |
| Numeric Guard (`extract_numbers`) | ✅ |
| Multi-Tenant Isolation (`org_id`) | ✅ |
| RBAC (`allowed_tool_ids`) | ✅ |
| Workflow/Chain Planning | ✅ |
| Prompt Compression | ✅ |
| Cache Skip Support | ✅ |
| Billing/Spend Limits | ✅ |
| Audit Logging + Telemetry | ✅ |
| Webhook Events | ✅ |

**This is a strong foundation. The router already knows WHAT class of tool to call via embeddings. What it doesn't reliably know yet is WHETHER two semantically similar prompts refer to the same underlying parameters and entities.**

---

## 2. The Real Weakness (Not Caching, Not Routing)

### Problem: `extract_numbers()` is Positional-Blind

The current numeric guard extracts raw numbers and compares them. This works for simple cases but breaks on **positional semantics**:

```
"Transfer $100 from account 123 to account 456"
"Transfer $100 from account 456 to account 123"
```
Both contain `[100, 123, 456]` → guard matches → **WRONG answer served**.

Another failure mode:
```
"Delete user 123"
"Get user 123"
```
Same number, **completely different action**. Numeric matching is irrelevant here.

### Why Full Canonical Intents Are Premature

The suggestions say:
```
Prompt → Intent Extractor → Canonical Intent → Cache
```

But **who extracts the intent?** For every new domain:
- "How many customers bought fertilizer this month?"
- "List buyers of fertilizer"
- "Show fertilizer sales customers"

You need `intent = fertilizer_customer_lookup`. Who creates, maintains, versions, validates, and deploys that? With 10 agents × 100 tools × 500 workflows, you're building a **mini operating system** inside the gateway.

The current vector-based routing **already handles this** without manual intent registries. It scales without human intervention.

---

## 3. The Fix: Canonical Entities, Not Canonical Intents

Instead of mapping prompts to abstract intent IDs, extract **structured entities** and store them alongside the cache payload:

### Current (Broken for Positional Cases):
```
Prompt → Embedding → Qdrant → extract_numbers([100, 123, 456])
```

### Upgraded (Lazy Post-Filter Verification):
```
Prompt → Embedding → Qdrant → Candidate Hit?
                                 ↓ yes (score > 0.95)
                              Extract Entities (fast regex, <1ms)
                                 ↓
                              Compare entities against cached payload
                                 ↓
                              Match? → Serve cached response
                              No match? → Continue to L5
```

> **⚠️ CRITICAL: The Latency Trap (Staff/DBRE Review Amendment)**
>
> Putting entity extraction _before_ embedding in the hot path would add 100-300ms if using an SLM, completely invalidating L1/L2 latency budgets. Instead, **flip the order**:
> 1. Generate embedding (fast, local)
> 2. Query Qdrant for semantic cache candidates
> 3. **Only if** a candidate matches with similarity > 0.95, _then_ extract entities from the incoming prompt and verify against the cached entity schema
> 4. If entities don't match → continue to full execution
> 5. The L5 LLM populates entity extraction **asynchronously** on cache write for the next request

Extract:
```json
{
  "action": "transfer",
  "source_account": "123",
  "target_account": "456",
  "amount": "100"
}
```

Now the semantic cache comparison becomes:
- Embedding similarity > 0.95 ✅
- Entities match (source=123, target=456, amount=100) ✅
- **Safe to serve cached response**

The entity payload travels with the cached prompt in Qdrant, replacing the fragile `extract_numbers()` comparison.

---

## 4. New Addition: Async Learning Plane (Offline Engine)

### The Insight

Everything discussed so far is **reactive** — the request arrives, then we optimize. The real next-level gain is a **proactive** background system that learns from traffic to improve future routing.

### Two-Plane Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    ONLINE PLANE                          │
│           (Handles live requests — latency optimized)    │
│                                                          │
│  L1  Exact Hash Cache          (Valkey)        < 2ms    │
│  L1b Hot Path Cache            (Valkey)        < 2ms    │
│  L2  Semantic Cache            (Qdrant)        5-15ms   │
│  L3  Semantic Tool Routing     (Qdrant)        5-15ms   │
│  L4  Tool Execution            (MCP/Connectors)10-50ms  │
│  L5  LLM Synthesis             (Providers)     800ms+   │
│                                                          │
│         │ emit events (non-blocking)                     │
│         ▼                                                │
├──────────────────────────────────────────────────────────┤
│                   OFFLINE PLANE                          │
│         (Learns from traffic — never blocks requests)    │
│                                                          │
│  O1  Request Mining     → entity patterns, frequencies   │
│  O2  Cache Mining       → pre-warm hot prompts nightly   │
│  O3  Workflow Mining    → detect repeated tool sequences │
│  O4  Agent Pattern Mining → Markov chains for prediction │
│                                                          │
│         │ produces                                        │
│         ▼                                                │
│  • New hot-path cache entries (L1b)                      │
│  • Promoted workflow templates (L3)                      │
│  • Updated entity extraction patterns                    │
│  • Tool ranking adjustments                              │
│  • Embedding cluster summaries                           │
└──────────────────────────────────────────────────────────┘
```

### How the Offline Plane Works

**Non-blocking event emission** from `engine.Process()`:

```go
// After every request completes (cache hit or LLM response)
go e.offlinePlane.Emit(OfflineEvent{
    OrgID:         orgID,
    PromptHash:    sha256Hex(queryPrompt),
    CanonicalHash: canonicalHash,
    Entities:      extractedEntities,
    EntitySource:  "regex", // or "llm" or "none"
    ToolsUsed:     toolResults,
    CacheLayer:    "L5",  // which layer resolved this
    LatencyMs:     duration.Milliseconds(),
    Provider:      providerKey,
})
```

Uses `try_send` / buffered channel semantics — if the channel is full, the event is **dropped gracefully**. The live path is never affected.

### What the Offline Miners Produce

| Miner | Input | Output | Example |
|-------|-------|--------|---------|
| **O1: Request Mining** | Audit logs, request patterns | Entity extraction patterns, frequency maps | "12,000 requests match `customer_purchase_history` pattern" |
| **O2: Cache Mining** | Cache miss logs | Pre-warmed L1b entries | Nightly: compute top-1000 prompt embeddings + workflow mappings |
| **O3: Workflow Mining** | Tool execution sequences | Promoted workflow templates (via Replay Simulation) | Detect: Search→Ledger→Balance→Format runs 50,000 times → simulate → create `CustomerBalanceWorkflow` |
| **O4: Agent Pattern Mining** | Multi-agent traces | Speculative pre-warm hints | ⚠️ **E6 (Experimental)** — Markov chains add complexity with diminishing returns until massive scale |

### Critical Safety Rails

| Risk | Mitigation |
|------|-----------|
| **Side-Effect Trap** | Tools must be flagged `is_read_only: bool`. Offline plane ONLY speculatively executes read-only tools. Never `create_invoice`, `charge_payment`, etc. |
| **Cache Stampede** | Deduplication layer: before background execution, check in-memory map / Valkey lock (`intent:param_hash` with short TTL) to prevent 50 identical background runs |
| **Context Drift** | `OfflineEvent` must be a self-contained context capsule — all security claims (org_id, user_id, tier) frozen at ingestion time to prevent background workers from evaluating rules with stale or elevated privileges |
| **Resource Exhaustion** | Worker pool sizes bounded via semaphore. If downstream provider latency spikes, offline plane auto-throttles to preserve network connections for live traffic |
| **Garbage Workflows** | Auto-detected workflows go through **Replay Simulation** before entering `status: pending_review`. Human approval required before promotion to L1b/L3 |
| **Telemetry Loss** | v1 uses buffered channels (volatile). v1.1+ uses Valkey Streams (`XADD` with `MAXLEN ~10000`) for crash-durable event ingestion across container restarts |

### Workflow Replay Simulator (Amendment: Required Before Promotion)

Between the O3 Workflow Miner discovering a pattern and the Approval Queue, a **Replay Simulator** validates the candidate:

```
O3 Detects Pattern (freq > 1000)
  ↓
Replay Simulator
  ├─ Fetch last 100 matching requests from audit_log
  ├─ For each: would the proposed workflow have produced correct output?
  ├─ Compare against actual LLM response (semantic similarity > 0.90)
  ↓
Generate Accuracy Report
  ├─ "Would Have Worked": 98.7% (99/100 requests)
  ├─ "Would Have Failed": 1.3% (1/100 — ambiguous entity extraction)
  ↓
Approval Queue (with report attached)
```

Reviewers now see **evidence**, not just frequency counts:
```
Frequency: 28,542 (30d)
Replay Accuracy: 98.7% (100 requests replayed)
Estimated Savings: $3,720/month
Failure Cases: 1 — ambiguous entity "account" could be source or target
```

### Workflow Lifecycle: Promotion AND Demotion (Amendment)

Workflows are not permanent. They follow a full lifecycle:

```
discovered → simulated → pending_review → approved → active → stale → demoted
```

| State | Trigger | Action |
|-------|---------|--------|
| `discovered` | O3 detects frequency > threshold | Store candidate |
| `simulated` | Replay simulator validates | Attach accuracy report |
| `pending_review` | Simulation passes (>90% accuracy) | Enter approval queue |
| `approved` | Human clicks Approve | Activate in L1b/L3 |
| `active` | Serving traffic | Track hit rate + accuracy |
| `stale` | Frequency drops below 10/day for 7 consecutive days | Auto-flag for review |
| `demoted` | Accuracy drops below 85% OR stale for 30 days | Remove from L1b/L3, archive |

```sql
-- Add to workflow_candidates table
ALTER TABLE workflow_candidates ADD COLUMN
    last_hit_at TIMESTAMPTZ,
    hit_count_7d INT DEFAULT 0,
    accuracy_7d FLOAT DEFAULT 1.0,
    demoted_at TIMESTAMPTZ,
    demotion_reason TEXT;

-- Nightly demotion check (run by O3 miner)
UPDATE workflow_candidates
SET status = 'stale', demoted_at = now(), demotion_reason = 'frequency_drop'
WHERE status = 'active'
  AND hit_count_7d < 70  -- less than 10/day average
  AND last_hit_at < now() - interval '7 days';

UPDATE workflow_candidates
SET status = 'demoted', demoted_at = now(), demotion_reason = 'accuracy_drop'
WHERE status = 'active'
  AND accuracy_7d < 0.85;
```

---

## 5. Revised Engine Execution Flow

```
engine.Process(req)
  │
  ├─ 1. Rate Limiting                           (existing)
  ├─ 2. Permission Check                        (existing)
  ├─ 3. Billing Pre-check                       (existing)
  │
  ├─ 4. L1: Exact Hash Cache                    (existing — Valkey literal + canonical)
  │      └─ HIT → return + emit offline event
  │
  ├─ 5. L1b: Hot Path Cache [NEW]
  │      ├─ Valkey lookup by entity-aware keys
  │      │   promoted by Offline Plane miners
  │      └─ HIT → return + emit offline event
  │
  ├─ 6. L2: Semantic Cache                      (existing — Qdrant similarity)
  │      ├─ Entity comparison replaces extract_numbers() [UPGRADED]
  │      └─ HIT → return + emit offline event
  │
  ├─ 7. Session Memory (short-term)             (existing)
  ├─ 8. Semantic Memory Recall (long-term)       (existing)
  │
  ├─ 9. L3: Semantic Tool Routing                (existing — Qdrant tool matching)
  ├─10. L4: Tool Execution                       (existing — MCP/Connectors)
  ├─11. L5: LLM Synthesis                        (existing — fallback)
  │
  ├─12. Cache Set                                (existing + entity payload) [UPGRADED]
  │      ├─ Valkey SET prompt_hash → response
  │      ├─ Valkey SET canonical_hash → response
  │      ├─ Qdrant upsert with entity payload [NEW]
  │      └─ Persistent DB cache
  │
  ├─13. Async Fact Extraction                    (existing — memory.ExtractAndStoreFacts)
  ├─14. Offline Event Emission [NEW]
  │      └─ Non-blocking emit to offline plane
  │
  └─15. Webhook Events                           (existing)
```

---

## 6. Implementation Phases (Evolution Track E1–E6)

> **Naming Convention:** These are labeled "Evolution Phases E1–E6" to avoid collision with the existing `ARCHITECTURE.md` Phases 1–5 (Core Foundation → BYO LLM). The E-track runs in parallel as a cache/routing intelligence upgrade.

### Evolution Phase E1: Entity Extraction Layer (2-3 weeks) — Highest ROI
**Goal**: Replace `extract_numbers()` with structured entity comparison. No major redesign.

**Rust Router Changes (`services/router/src/handlers.rs`):**
- [ ] **Remove `sort_by` calls** (lines 78-79) — this is the actual bug that destroys positional information. The regex itself (`\d+(?:\.\d+)?`) is already correct.
- [ ] New `extract_entities()` function — returns `HashMap<String, String>` with labeled params (e.g., `{"source_account": "123", "target_account": "456", "amount": "100"}`)
- [ ] Replace the sorted-array comparison with entity map comparison: keys + values must match
- [ ] Store entities in `prompts_collection` Qdrant payload alongside `prompt_hash` and `prompt_text`
- [ ] Update proto: add `map<string, string> entities` to `ToolResponse`

**Go Gateway Changes:**
- [ ] Receive entities from router gRPC response
- [ ] Log entities in audit trail for offline mining
- [ ] No engine flow changes — just richer cache comparison data

**Dashboard:**
- [ ] Show extracted entities in request detail view

### Evolution Phase E2: L1b Hot Path Cache + Valkey Upgrade (1-2 weeks)
**Goal**: Add a Redis/Valkey layer for high-frequency entity-keyed lookups.

- [ ] Entity-keyed Valkey entries: `transfer:source=123:target=456:amount=100`
- [ ] L1b check in `engine.Process()` after L1 exact hash, before L2 semantic
- [ ] Dual-write: on cache set, write both prompt_hash key AND entity key
- [ ] Dashboard: cache layer hit distribution chart (L1 vs L1b vs L2)

### Evolution Phase E3: Offline Learning Plane (3-4 weeks) — Game Changer
**Goal**: Background system that learns from traffic patterns, never blocks requests.

**Infrastructure:**
- [ ] `internal/offline/` Go package with buffered channel event bus (v1)
- [ ] Non-blocking `Emit()` method called from `engine.Process()` post-response
- [ ] `OfflineEvent` struct: org_id, prompt_hash, canonical_hash, entities, entity_source, tools_used, cache_layer, latency (no raw prompt — see §7.2)
- [ ] v1.1: Migrate to Valkey Streams (`XADD` with `MAXLEN ~10000`) for crash durability

**O1: Request Miner:**
- [ ] Aggregate request frequencies by entity pattern
- [ ] Detect hot prompts that always miss cache
- [ ] Output: candidate L1b entries for pre-warming

**O2: Cache Miner:**
- [ ] Nightly job: scan audit_log for top-N cache misses
- [ ] Pre-compute embeddings + entity extractions
- [ ] Warm Valkey with predicted hot entries (tagged `speculative: true`)

**O3: Workflow Miner + Replay Simulator:**
- [ ] Detect repeated tool execution sequences across requests
- [ ] **Replay Simulator**: validate candidates against last 100 historical requests
- [ ] Only candidates with ≥90% replay accuracy enter `status: pending_review`
- [ ] Store in `workflow_candidates` Postgres table

**Safety:**
- [ ] `is_read_only` flag on tool registry
- [ ] Deduplication map for in-flight background computations
- [ ] Semaphore-bounded worker pool (auto-throttle under pressure)
- [ ] All auto-detected workflows require replay simulation + human approval

### Evolution Phase E4: Workflow Registry — Hot Paths Only (2 weeks)
**Goal**: Promote only top 5-10% traffic patterns into deterministic workflows.

- [ ] `workflow_registry` Postgres table (migration `025_`)
- [ ] Workflow lifecycle states: discovered → simulated → pending_review → approved → active → stale → demoted
- [ ] Demotion logic: auto-stale at <10 hits/day for 7d, auto-demote at <85% accuracy
- [ ] Dashboard: approve/reject workflow candidates with replay accuracy report
- [ ] `engine.Process()`: if classified intent has approved workflow → execute tool directly, skip LLM
- [ ] Only for high-volume patterns (math, weather, currency, customer lookup, inventory)
- [ ] **Do NOT make everything a workflow** — let embeddings handle the long tail

### Evolution Phase E5: Model Router + Entity Quality Metrics (1-2 weeks)
**Goal**: Route to appropriate model size based on complexity. Track entity extraction health.

```
Simple extraction     → Ollama / Phi / Mistral (local)
Known workflow        → No LLM at all
Complex reasoning     → GPT / Claude / Gemini
```

- [ ] Route based on: tool match confidence + entity extraction confidence + workflow match
- [ ] If confidence > 0.95 and workflow exists → skip LLM entirely
- [ ] If confidence < 0.5 → route to large model
- [ ] Middle ground → route to small/local model

**Entity Extraction Quality Dashboard:**
- [ ] Regex success rate (entities extracted without LLM)
- [ ] Regex failure rate (fell through to LLM dual-return)
- [ ] Entity mismatch rate (cache guard rejected due to entity mismatch)
- [ ] LLM entity extraction usage (% of requests needing LLM for entities)
- [ ] Alert if regex failure rate > 15% (trigger SLM evaluation)

### Evolution Phase E6: Agent Pattern Mining — Experimental (Future)
**Goal**: Markov chain prediction for speculative pre-warming. Only at massive scale.

- [ ] O4 Miner: Build Markov transition matrices from multi-agent tool call traces
- [ ] Speculative pre-warm: if Agent A calls tool X, pre-warm tool Y's cache
- [ ] **Only activate when**: daily request volume > 500k AND workflow mining exhausted
- [ ] Requires separate evaluation: complexity vs. diminishing returns

---

## 7. Data Structure Changes

### 7.1 Entity Payload (Rust Router — `prompts_collection`)
```json
{
  "prompt_hash": "abc123",
  "prompt_text": "Transfer $100 from account 123 to 456",
  "org_id": "org_xxx",
  "entities": {
    "action": "transfer",
    "source_account": "123",
    "target_account": "456",
    "amount": "100"
  }
}
```

### 7.2 Offline Event (Go Gateway — emitted post-response)
```go
type OfflineEvent struct {
    OrgID         string            `json:"org_id"`
    UserID        string            `json:"user_id"`
    PromptHash    string            `json:"prompt_hash"`     // SHA-256 — non-reversible, joins to L1 cache
    CanonicalHash string            `json:"canonical_hash"`  // joins to L1.5 cache
    Entities      map[string]string `json:"entities"`        // only present when extraction succeeded
    EntitySource  string            `json:"entity_source"`   // "regex" | "llm" | "none"
    ToolsUsed     []string          `json:"tools_used"`
    CacheLayer    string            `json:"cache_layer"`     // L1, L1b, L2, L5
    LatencyMs     int64             `json:"latency_ms"`
    TokensUsed    int               `json:"tokens_used"`
    Provider      string            `json:"provider"`
    Success       bool              `json:"success"`
    Timestamp     time.Time         `json:"timestamp"`
}
```

> **⚠️ DBRE Amendment — No Raw Prompts in Event Stream:**
> Raw prompt text is **never** emitted to the offline plane. Miners operate on hashes, entity maps, and tool sequences — they don't need raw text. This prevents:
> - PII exposure in the event ring buffer
> - Cross-org data governance violations
> - Compliance issues for enterprise customers asking "what do you retain?"
>
> If a miner needs to reference the original prompt (e.g., for replay simulation), it queries the cache by `PromptHash` — access-controlled and audit-logged.

### 7.3 Workflow Candidate (Postgres — from O3 Miner)
```sql
CREATE TABLE workflow_candidates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL,
    pattern         TEXT NOT NULL,        -- e.g. "search_customer → get_ledger → calc_balance"
    frequency       INT NOT NULL,         -- how many times detected
    tool_ids        TEXT[] NOT NULL,       -- ordered tool sequence
    status          TEXT DEFAULT 'discovered',
    -- Lifecycle: discovered → simulated → pending_review → approved → active → stale → demoted
    replay_accuracy FLOAT,               -- from replay simulator (NULL until simulated)
    replay_count    INT,                  -- how many requests were replayed
    last_hit_at     TIMESTAMPTZ,
    hit_count_7d    INT DEFAULT 0,
    accuracy_7d     FLOAT DEFAULT 1.0,
    created_at      TIMESTAMPTZ DEFAULT now(),
    reviewed_by     TEXT,
    reviewed_at     TIMESTAMPTZ,
    demoted_at      TIMESTAMPTZ,
    demotion_reason TEXT                  -- 'frequency_drop' | 'accuracy_drop' | 'manual'
);
```

### 7.4 Proto Changes (`/proto/router.proto`)
```protobuf
// Add to existing ToolResponse
message ToolResponse {
    // ... existing fields ...
    map<string, string> entities = 6;  // NEW: extracted entities
}

// New: for future E4 workflow classification
message ClassifyIntentRequest {
    string prompt = 1;
    string org_id = 2;
}

message ClassifyIntentResponse {
    bool matched = 1;
    string workflow_id = 2;
    map<string, string> parameters = 3;
    float confidence = 4;
    string cache_key = 5;
}
```

---

## 8. The Economic Thesis: GPU Avoidance Rate

### Why This Matters More Than Cache Hit Ratio

Most optimization discussions focus on saving 20ms here, 50ms there, or reducing embedding costs. But the **real money** is in one thing:

> **Avoid the LLM invocation entirely.**

### The Math at Scale

Assume 100,000 requests/day:

**Current state (no offline learning):**
| Layer | Hit % | Daily Requests | Cost per Request | Daily Cost |
|-------|-------|---------------|-----------------|-----------|
| L1 Exact Cache | 30% | 30,000 | ~$0.00 | ~$0 |
| L2 Semantic Cache | 20% | 20,000 | ~$0.0001 | ~$2 |
| L5 LLM (fallback) | **50%** | **50,000** | ~$0.03 (3k tokens avg) | **~$1,500** |
| **Total** | | | | **~$1,502/day** |

**After offline learning matures (target state):**
| Layer | Hit % | Daily Requests | Cost per Request | Daily Cost |
|-------|-------|---------------|-----------------|-----------|
| L1 Exact Cache | 35% | 35,000 | ~$0.00 | ~$0 |
| L1b Hot Path (promoted) | **35%** | **35,000** | ~$0.00 | ~$0 |
| L2 Semantic Cache | 20% | 20,000 | ~$0.0001 | ~$2 |
| L3/L4 Workflow+Tool | 5% | 5,000 | ~$0.00 | ~$0 |
| L5 LLM (fallback) | **5%** | **5,000** | ~$0.03 | **~$150** |
| **Total** | | | | **~$152/day** |

**That's a 10x cost reduction** — from $1,500/day to $152/day — by moving GPU usage from 50% → 5% of requests.

### The North Star Metric

The interesting metric isn't cache hit ratio anymore. It's:

```
GPU Avoidance Rate = (LLM calls avoided) / (total requests)
```

| Metric | Current | E2 | E3 (8wk) | E5 (Long-term) |
|--------|---------|---------|---------------|---------------------|
| GPU Avoidance Rate | 50% | 65% | **70-80%** | 95% (objective) |
| Daily LLM Cost (100k req) | $1,500 | $1,050 | $450-$600 | $150 |
| Annual Savings vs Current | — | $164k | $329k-$383k | $493k |

### How This Maps to the Architecture

The best AI Gateway eventually looks like:

```
70% L1/L1b   → Deterministic (Valkey KV, promoted workflows)
20% L2       → Near-deterministic (Qdrant vectors + entity guard)
 8% L3/L4    → Deterministic (Workflow registry + tool execution)
 2% L5       → GPU (only genuinely novel problems)
```

**Only novel problems reach GPUs. Everything else becomes deterministic execution.**

---

## 9. The LLM-as-Compiler Model

### Think of LLM as a Compiler, Not a Runtime

The first time a new problem class appears:
```
User: "How much does customer Raj owe?"

LLM determines:
{
  "workflow": "customer_balance",
  "entity": "Raj"
}
```

The Offline Plane captures this and mines the pattern:
```
Trigger words: ["owes", "balance", "outstanding", "due amount", "dues"]
Entity slot:   customer_name
Workflow:      customer_balance_v1
Tool chain:    search_customer → get_ledger → compute_balance
```

Next 10,000 requests — **no LLM**:
```
"What is Raj's balance?"     → customer_balance_v1(entity="Raj")
"How much does Raj owe?"     → customer_balance_v1(entity="Raj")
"Outstanding amount?"        → customer_balance_v1(entity=ctx)
"Current dues?"              → customer_balance_v1(entity=ctx)
```

This is exactly what Google Search does — most queries hit precomputed indexes, precomputed ranking, and precomputed intent graphs. LLM (Gemini) only for edge cases.

### The Workflow Promotion Engine (O3 Miner Detail)

The O3 Workflow Miner watches every request event and detects repeated tool execution sequences:

```
Every request emits:
{
  "prompt": "...",
  "tools_used": ["search_customer", "get_ledger", "compute_balance"],
  "success": true,
  "latency_ms": 1200,
  "tokens_used": 3500,
  "cache_layer": "L5"
}
```

**Promotion criteria (OR-based — either path qualifies):**
| Condition | Threshold | Rationale |
|-----------|-----------|-----------|
| **Path A: Frequency** | ≥ 1,000 successful executions AND ≥ 95% success rate | High-volume stable pattern |
| **Path B: Token Savings** | ≥ 250 executions AND ≥ 20M tokens/month saved | Low-frequency but expensive — still worth promoting |
| Tool sequence stability | Same ordered tools ≥ 90% of time | Not random tool selection (applies to both paths) |
| **Replay Simulation** | ≥ 90% accuracy on last 100 requests | **Mandatory** — no promotion without replay evidence |

When criteria met → Replay Simulator validates → insert into `workflow_candidates` with `status: simulated`.

**Confidence-gated execution:**
```
L1 Exact Cache
  ↓ miss
L1b Workflow Cache
  ↓ match found
  ↓
Confidence > 95%? ──yes──► Execute workflow directly (no LLM)
  ↓ no
L2 Semantic Cache
  ↓ miss
L5 LLM (with workflow hint for faster planning)
```

If confidence is between 80-95%, the system still uses LLM but passes the workflow hint as context, dramatically reducing planning tokens.

---

## 10. Performance Targets

| Layer | Mechanism | Target Latency | Token Cost |
|-------|-----------|---------------|------------|
| L1 Exact Hash | Valkey KV | < 2ms | $0.0000 |
| L1b Hot Path | Valkey entity-keyed + promoted workflows | < 2ms | $0.0000 |
| L2 Semantic | Qdrant + entity guard | 5-15ms | $0.0001 |
| L3 Tool Routing | Qdrant tool match | 5-15ms | $0.0000 |
| L4 Tool Exec | MCP/Connector | 10-50ms | $0.0000 |
| L5 Full LLM | Generation | 800-3000ms | $0.01-$0.10+ |

**Offline Plane**: Runs O(minutes) in background, improves future request resolution by 10-100x.

**Primary Target**: GPU Avoidance Rate ≥ 70-80% within 8 weeks of E3 deployment. 95% is a long-term objective (6+ months of traffic learning).

**Secondary Target**: 80%+ of recurring agent requests resolved at L1/L1b/L2 within 6 weeks.

---

## 11. Dashboard Metrics (GPU Avoidance Observability)

The dashboard should expose these metrics for each org:

### Real-Time Panel
```
GPU Avoidance Rate:  92.3%  ▲ +4.1% (7d)
Daily LLM Cost:      $187   ▼ -$340 (7d)
Promoted Workflows:  14 active / 3 pending review
```

### Layer Distribution Chart (Stacked Area, 30d)
```
100% ┤████████████████████████████████
     │██ L1 ████████████████████████
 75% │████████████████████████████████
     │██ L1b ███████████████████████  ← grows as offline plane learns
 50% │████████████████████████████████
     │██ L2 ████████████████████████
 25% │████████████████████████████████
     │██ L3/L4 █████████████████████
  5% │██ L5 (GPU) █████████████████  ← shrinks over time
  0% └──────────────────────────────►
     Day 1                      Day 30
```

### Offline Plane Health
| Miner | Events/hr | Candidates Found | Last Run |
|-------|-----------|-----------------|----------|
| O1 Request | 12,400 | 23 hot prompts | 2m ago |
| O2 Cache | 8,100 | 156 pre-warmed | 1h ago |
| O3 Workflow | 3,200 | 3 candidates | 15m ago |
| O4 Agent Pattern | 1,800 | 1 Markov chain | 30m ago |

### Workflow Promotion Queue
| Pattern | Frequency | Replay Accuracy | Est. Token Savings/day | Status |
|---------|-----------|----------------|----------------------|--------|
| search→ledger→balance | 4,200 | 98.7% (100 replayed) | 14.7M tokens | **Approve?** |
| inventory→price→format | 1,100 | 96.3% (100 replayed) | 3.8M tokens | **Approve?** |
| user_lookup→permissions | 890 | — (not yet simulated) | 3.1M tokens | Discovered |

### Workflow Lifecycle Health
| Workflow | State | Hits/7d | Accuracy/7d | Action |
|----------|-------|---------|-------------|--------|
| customer_balance_v1 | **active** | 12,400 | 99.1% | — |
| invoice_lookup_v2 | **stale** | 8 | 97.0% | Auto-demotion in 3d |
| price_format_v1 | **demoted** | 0 | — | Archived |

### Entity Extraction Quality
```
Regex Success Rate:     87.3%  (entities extracted without LLM)
LLM Dual-Return Usage:  12.7%  (fell through to LLM for entities)
Entity Mismatch Rate:    2.1%  (cache guard rejected — entity ≠ candidate)
L1b Prediction Yield:  74.2%  (speculative hits / speculative entries created)
```
> ⚠️ Alert threshold: If regex failure > 15% → evaluate SLM addition

---

## 12. Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Entities first, intents later** | Canonical entities before intent registry | Higher ROI — fixes false positives without manual intent maintenance |
| **Offline Plane architecture** | Buffered channel, non-blocking emit | Live path latency must be zero-impact |
| **Workflow scope** | Top 5-10% hot paths only | Don't make everything a workflow — embeddings handle the long tail |
| **Auto-learning safety** | Human approval required | Prevents garbage workflow accumulation |
| **Speculative execution** | Read-only tools only | Never speculatively run state-mutating tools |
| **Deduplication** | In-memory map + Valkey TTL lock | Prevents cache stampede on identical background runs |
| **Entity extraction method** | Regex + labeled patterns (E1), SLM upgrade only when metrics demand (E5+) | Start simple, graduate to model-based when traffic justifies |
| **Existing router preserved** | No rewrite | Current vector routing scales without human intervention |
| **North star metric** | GPU Avoidance Rate, not cache hit ratio | Directly maps to infrastructure cost and unit economics |
| **LLM role** | Compiler, not runtime | LLM solves a problem class ONCE; deterministic execution handles all future instances |
| **Promotion threshold** | freq ≥1000 OR monthly_savings ≥ configurable threshold (both require replay simulation) | Low-frequency expensive workflows still qualify |
| **Confidence gating** | >95% = skip LLM, 80-95% = LLM with hint, <80% = full LLM | Graceful degradation, not binary skip/call |
| **Workflow demotion** | Auto-stale at <10 hits/day for 7d, auto-demote at <85% accuracy | Registry doesn't grow forever; stale paths evicted |
| **Replay simulation** | Mandatory before promotion — replay last 100 requests, ≥90% accuracy required | Reviewer sees evidence, not just miner's claim |
| **Entity extraction quality** | Track regex success/failure rate, LLM fallback %, mismatch rate | Alert at >15% regex failure — data-driven SLM decision |
| **LLM dual-return format** | Structured JSON (not HTML comments) | Parseable, validatable, works with response_format enforcement |
| **GPU target** | 70-80% in 8 weeks, 95% long-term | Realistic — avoids premature management commitment |
| **O4 Markov chains** | E6 (Experimental) — only at >500k req/day | Workflow mining captures most value; Markov adds complexity for diminishing returns |

---

## 13. System Migration & Backward Compatibility

### Zero-Downtime Data Shifting
To shift smoothly from v1 (positional-blind numeric array matching) to v2 (structured canonical entity verification) without invalidating existing production caches:

1. **Dual-Writing Phase:** Deploy entity extraction in shadow mode. New cache payloads write both the raw extracted numeric array AND the newly structured entity JSON map into Qdrant.
2. **Fallback Reads:** On cache reads, if the matched Qdrant payload does not contain an entity map, fallback gracefully to the `extract_numbers()` comparison engine.
3. **Lazy Hydration:** Any fallback read that yields a positive match is automatically queued into the Offline Plane to extract its entities and upgrade the cache schema seamlessly in the background.

This approach ensures:
- Zero cache invalidation on deploy
- Gradual migration (old entries hydrated over time by natural traffic)
- Instant rollback capability (just disable entity comparison, fallback still works)

---

## 14. Scoring Comparison

| Approach | Score | Notes |
|----------|-------|-------|
| Current router (as-is) | 8/10 | Very solid, `extract_numbers` is the main weakness |
| Full intent registry rewrite | 6/10 | Good vision, too much upfront complexity |
| **Hybrid: Entities + Offline Plane + Hot-Path Workflows** | **9.7/10** | Evolves current strengths, adds learning, minimal disruption |

---

## 15. Resolved Design Questions — Staff Review

> The following 5 questions were debated by two senior engineering perspectives and resolved by Staff-level architectural decision. Each answer is now **locked** as the implementation spec.

---

### Q1: Entity Extraction — Regex or SLM?

**Engineer A (Enterprise/Distributed):**
> Hybrid stratified pipeline. Fast regex for structured patterns, then piggyback on the L5 LLM that's already processing the request. Have the LLM return both the answer AND normalized entities. Cache the entity layout so future semantic matches can extract params deterministically. Never add a standalone SLM — it adds 150ms latency per request.

**Engineer B (Pragmatic/Minimalist):**
> Pure regex on day 1. Build typed extractors: `MoneyExtractor`, `UUIDExtractor`, `DateExtractor`, `CustomerExtractor`. Add SLM only when a metric proves regex can't solve it. 95% of traffic should never hit any model for extraction. "Don't add a model until a metric proves rules cannot solve it."

**Where they agree:**
- No standalone SLM on day 1
- Regex/rules first
- Structured patterns are deterministic and should stay deterministic

**Where they disagree:**
- Engineer A wants the L5 LLM to dual-return entities on cache misses immediately
- Engineer B says wait until regex failure rate proves a model is needed

#### ✅ DECISION: Stratified Regex + LLM Dual-Return (Engineer A's approach, with Engineer B's discipline)

```
Input Text
  ↓
Typed Regex Extractors (< 1ms)
  ├─ MoneyExtractor:    $250, $1.5M
  ├─ IDExtractor:       UUID, invoice numbers, PO numbers
  ├─ DateExtractor:     "last month", "Q3 2025", ISO dates  
  ├─ NumberExtractor:   a=10, b=5 (replaces extract_numbers)
  └─ NamedEntityExtractor: "customer Raj", "account 456"
  ↓
Extraction complete?
  ├─ YES → entities stored with cache payload, done
  └─ NO (fuzzy input like "raj's fertilizer purchases from last harvest")
       ↓
       Request continues to L5 LLM anyway
       ↓
       LLM returns answer + structured entities (dual-return)
       ↓
       Offline Plane captures the entity pattern for future regex promotion
```

**The rule:** _No request ever takes a latency hit for entity extraction. Regex runs inline (< 1ms). If regex fails, the LLM that's already running returns entities as a side-channel. The Offline Plane learns new regex patterns from LLM extractions over time._

**Implementation:**
```rust
// Rust Router — new trait
pub trait EntityExtractor: Send + Sync {
    fn extract(&self, text: &str) -> Vec<Entity>;
}

pub struct Entity {
    pub label: String,     // "source_account", "amount", "customer_name"
    pub value: String,     // "123", "$100", "Raj"
    pub extractor: String, // "regex:money", "regex:id", "llm:dual_return"
}
```

**Go Gateway — LLM dual-return prompt addition:**
```go
// Append to system prompt when entities were NOT fully extracted by regex
entityInstruction := `Additionally, return your response as structured JSON:
{
  "answer": "<your response to the user>",
  "entities": {"customer": "Raj", "action": "balance_lookup"}
}
If you cannot structure as JSON, return plain text — the system will handle it gracefully.`
```

> **Design note**: Structured JSON over HTML comments (`<!--entities:...-->`). HTML comments are fragile — models hallucinate formatting inconsistently. JSON schema is parseable, validatable, and works with response_format enforcement on OpenAI/Anthropic.

**Metric that triggers SLM addition:**
- If > 15% of requests fail regex extraction AND those requests have > 3,000 tokens average → evaluate adding Phi-4-Mini / Qwen-3-4B as a dedicated extractor microservice.
- Until that threshold is proven by data, no SLM.

---

### Q2: Offline Plane Transport — Channels vs. Streams

**Engineer A:**
> Start with local tokio channels, but graduate to Valkey Streams immediately for v1.0. You already have Valkey, so no new infra. Valkey Streams give you consumer groups for multi-instance scaling. Skip NATS/Kafka entirely.

**Engineer B:**
> `tokio::mpsc` or `crossbeam` for v1. Redis Streams for v2. NATS JetStream for v3 at 50M+ events/day. Don't introduce Kafka on day one. Or ever, if possible.

**Where they agree:**
- Local channels for day 1 (zero external dependency)
- Redis/Valkey Streams as the natural next step (already in stack)
- No Kafka

**Where they disagree:**
- Engineer A says skip NATS entirely and use Valkey Streams forever
- Engineer B leaves NATS as a v3 option for extreme scale

#### ✅ DECISION: 3-Stage Transport Evolution

| Version | Transport | When | Why |
|---------|-----------|------|-----|
| **v1 (Day 1)** | Go buffered channels (`chan OfflineEvent`, buffer=10,000) | Single gateway instance | Zero dependency, <1µs emit |
| **v2 (Multi-instance)** | Valkey Streams + Consumer Groups | 2+ gateway instances behind LB | Already in stack, no new infra, persistent event log |
| **v3 (Evaluate only if needed)** | NATS JetStream | 50M+ events/day, Valkey Streams becoming bottleneck | Only if Valkey Streams prove insufficient — may never happen |

**The rule:** _Kafka is permanently off the table. It's overkill for event volumes under 100M/day and adds enormous operational burden. Valkey Streams + Consumer Groups handle the projected scale (1-10M events/day) comfortably._

**Implementation (v1 — Go Gateway):**
```go
type OfflinePlane struct {
    events chan OfflineEvent
    // ... miners
}

func NewOfflinePlane(bufferSize int) *OfflinePlane {
    op := &OfflinePlane{
        events: make(chan OfflineEvent, bufferSize), // default 10,000
    }
    go op.runMiners() // background consumer goroutines
    return op
}

// Non-blocking emit — drops if buffer full (graceful degradation)
func (op *OfflinePlane) Emit(event OfflineEvent) {
    select {
    case op.events <- event:
        // queued
    default:
        // buffer full — drop silently, system under pressure
        slog.Debug("Offline plane buffer full, dropping event")
    }
}
```

---

### Q3: Workflow Approval UX — Dashboard vs. Slack

**Engineer A:**
> Stateless webhooks + Slack Interactive Blocks first. One-click approve via signed JWT URL. Dashboard later once API surfaces are stable.

**Engineer B:**
> Do NOT use Slack first. You'll get notification fatigue with "Approve? Approve? Approve?" flooding channels. Build a Workflow Inbox inside the admin dashboard. Optionally send a daily Slack summary with a link.

**Where they agree:**
- There must be a human in the loop
- The UI should show frequency, success rate, and estimated token savings

**Where they disagree:**
- Fundamentally disagree on the primary surface. Engineer A says Slack-first, dashboard later. Engineer B says dashboard-first, Slack as optional digest.

#### ✅ DECISION: Dashboard Inbox First + Daily Slack Digest (Engineer B's approach)

**Rationale:** Engineer B is right about notification fatigue. At scale, the O3 miner might surface 5-20 candidates per day. Individual Slack notifications create noise. The approval decision requires context — seeing frequency curves, success rates, tool chain details, and estimated savings side by side. That's a dashboard problem, not a notification problem.

**Implementation:**
```
┌─────────────────────────────────────────────────────────┐
│  WORKFLOW PROMOTION QUEUE                    3 pending  │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌─ customer_balance_v1 ─────────────────────────────┐ │
│  │ Pattern: search_customer → get_ledger → calc_bal  │ │
│  │ Frequency: 28,542 executions (30d)                │ │
│  │ Success: 99.2%                                    │ │
│  │ Est. Savings: 12.4M tokens/month ($3,720)         │ │
│  │ Avg Latency: L5=1,200ms → L1b=2ms (600x faster)  │ │
│  │                                                    │ │
│  │  [✅ Approve]  [❌ Reject]  [👁 Preview]           │ │
│  └────────────────────────────────────────────────────┘ │
│                                                         │
│  ┌─ inventory_price_lookup ──────────────────────────┐ │
│  │ Pattern: inventory_search → price_calc → format   │ │
│  │ Frequency: 1,100 executions (30d)                 │ │
│  │ ...                                               │ │
│  └────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

**Daily Slack digest (optional, not per-workflow):**
```
🔄 Memzent Workflow Report — June 7
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
3 workflows pending approval
GPU Avoidance Rate: 87.2% (▲ +2.1%)
Est. savings if approved: $4,100/month
→ Review: https://dashboard.memzent.ai/workflows/queue
```

**Dashboard page location:** `services/dashboard/src/app/workflows/page.tsx`
**Server action:** `src/app/actions.ts` → `GET /v1/workflows/candidates`, `POST /v1/workflows/candidates/{id}/approve`

---

### Q4: Prediction Accuracy Tracking

**Engineer A:**
> Speculative Key Tainting. When offline plane writes to L1b, tag the key with `speculative: true`. On live hit → flip to false, increment `speculative_hit`. On TTL expiry while still `speculative: true` → increment `speculative_miss`. Use Valkey Keyspace Notifications for expired events. Throttle Markov chains if accuracy < 60%.

**Engineer B:**
> Every prediction gets an ID. Track: Accepted? Correct? Fallback? Measure Prediction Precision, Promotion Success Rate, and GPU Avoidance. "1.2B tokens/month avoided — that's executive dashboard material."

**Where they agree:**
- Must track per-prediction lifecycle
- GPU Avoidance Rate is the north star
- Need automated throttling if predictions go bad

**Where they disagree:**
- Engineer A focuses on key-level tainting (infrastructure approach)
- Engineer B focuses on prediction-level IDs (application approach)

#### ✅ DECISION: Both — Key Tainting for L1b + Prediction IDs for Workflows

They're solving different problems. Use both.

**L1b Pre-warming Accuracy (Key Tainting — Engineer A):**
```
Valkey key structure:
  Key:   "l1b:org123:customer_balance:Raj"
  Value: { response: "...", speculative: true, prediction_id: "abc123" }
  TTL:   3600s

On live request hit:
  → HSET key speculative false
  → INCR metric:speculative_hits:{org_id}

On TTL expiry (Valkey keyspace notification):
  → INCR metric:speculative_misses:{org_id}

Prediction Accuracy = hits / (hits + misses)
If accuracy < 60% for 24h → throttle O2/O4 miners for that org
```

**Workflow Promotion Accuracy (Prediction IDs — Engineer B):**
```sql
-- Track every workflow execution decision
CREATE TABLE workflow_predictions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL,
    workflow_id     TEXT NOT NULL,
    prediction_id   TEXT NOT NULL,
    confidence      FLOAT NOT NULL,
    decision        TEXT NOT NULL,     -- 'executed' | 'fallback_llm' | 'rejected'
    correct         BOOLEAN,           -- NULL until verified, true if user accepted result
    tokens_saved    INT,
    created_at      TIMESTAMPTZ DEFAULT now()
);

-- Executive dashboard query
SELECT 
    workflow_id,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE decision = 'executed') as direct_executions,
    COUNT(*) FILTER (WHERE decision = 'fallback_llm') as llm_fallbacks,
    SUM(tokens_saved) as total_tokens_saved
FROM workflow_predictions
WHERE org_id = $1 AND created_at > now() - interval '30 days'
GROUP BY workflow_id;
```

**Dashboard metrics panel:**
```
Prediction Accuracy:     89.3%
L1b Pre-warm Hit Rate:   72.1%
Workflow Execution Rate:  94.7%
GPU Avoidance Rate:       91.2%
Monthly Tokens Saved:     847M ($2,541)
```

---

### Q5: Cross-Org Learning — Global vs. Per-Org

**Engineer A:**
> Strictly per-org for dynamic mining. Global only for static system blueprints (e.g., "Postgres query → Slack notification" is a universal pattern). Never let Org A's usage patterns alter Org B's execution paths. Even anonymized behavioral leakage is a compliance risk.

**Engineer B:**
> Never start with global learning. Different problem than data leakage — it's *behavior* leakage. "FDA drug validation" and "retail inventory" patterns aren't transferable. Per-org for O1 miners. Optional global learning for workflow *structure* only (not entities, not prompts, not data). Example safe: "Search → Lookup → Aggregate → Respond". Example unsafe: "Customer ABC, Invoice 1234, Revenue 9M".

**Where they agree:**
- Per-org isolation is non-negotiable for dynamic mining
- Cross-org is a compliance minefield
- Some structural patterns are genuinely universal

**Where they disagree:**
- Minor: Engineer A calls them "static system blueprints", Engineer B calls them "structural workflow patterns". Same concept.

#### ✅ DECISION: Strict Per-Org Isolation + Global Structural Templates (Read-Only, Curated)

**No ambiguity here — both engineers converge on the same answer.**

| Mining Type | Scope | Data Boundary |
|------------|-------|---------------|
| O1 Request Mining | Per-org only | Prompt text, entities, frequencies — never leave org |
| O2 Cache Mining | Per-org only | Pre-warming targets scoped to org_id |
| O3 Workflow Mining | Per-org only | Tool sequences detected within single org's traffic |
| O4 Agent Pattern Mining | Per-org only | Markov chains built from org's own agent traces |
| **Global Structural Templates** | **Curated, read-only** | **Hand-maintained by platform team, no auto-learning** |

**Global templates — what qualifies:**
```json
// These are hardcoded in gateway, NOT learned from traffic
{
  "global_templates": [
    { "pattern": "db_query → format_response", "category": "data_retrieval" },
    { "pattern": "search → lookup → aggregate", "category": "analytics" },
    { "pattern": "validate_input → execute → notify", "category": "action" }
  ]
}
```

**What NEVER crosses org boundaries:**
- Entity values (customer names, amounts, IDs)
- Prompt text (even anonymized)
- Tool execution parameters
- Frequency data (reveals business volume)
- Confidence scores (reveals data patterns)

**Implementation guard:**
```go
// Every miner enforces this at construction time
type Miner struct {
    orgID string // immutable after construction — cannot be empty or "*"
}

func NewMiner(orgID string) (*Miner, error) {
    if orgID == "" || orgID == "*" || orgID == "global" {
        return nil, fmt.Errorf("miners must be org-scoped: got %q", orgID)
    }
    return &Miner{orgID: orgID}, nil
}
```

---

### Bonus: Q6 & Q7 (Rapid Resolution)

**Q6: Confidence Calibration (80%/95% thresholds)**
> ✅ DECISION: Start with hardcoded thresholds (80%/95%). Add per-org override in `org_settings` table. No A/B testing infrastructure on day 1. Review thresholds monthly based on `workflow_predictions` accuracy data. If an org's prediction accuracy drops below 70%, auto-raise their threshold to 98%.

**Q7: GPU Budget Alerting**
> ✅ DECISION: Yes. Alert when GPU Avoidance Rate drops below 80% sustained for 1 hour. Use existing webhook event system (`e.emitEvent(ctx, orgID, "gpu_avoidance_alert", ...)`). Dashboard shows a red/amber/green indicator. Threshold stored in `org_settings`, default 80%, adjustable per org.
