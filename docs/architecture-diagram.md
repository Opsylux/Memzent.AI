# Memzent Architecture Diagrams

> Render these with [mermaid.live](https://mermaid.live), VS Code Mermaid extension, or any Mermaid-compatible tool. Export as PNG/SVG for LinkedIn, blog posts, etc.

---

## 1. High-Level System Architecture

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': { 'primaryColor': '#00f3ff', 'primaryTextColor': '#fff', 'primaryBorderColor': '#00f3ff', 'lineColor': '#666', 'secondaryColor': '#7c3aed', 'tertiaryColor': '#1a1a2e'}}}%%

flowchart TB
    subgraph Clients["🧑‍💻 Clients"]
        direction LR
        A1["AI Agents"]
        A2["Chat Apps"]
        A3["MCP Clients"]
    end

    subgraph Gateway["⚡ Go Gateway :8080"]
        direction TB
        G1["Auth & RBAC"]
        G2["Rate Limiting"]
        G3["Billing Engine"]
        G4["4-Layer Cache"]
        G5["Session & Memory"]
        G6["MCP Tool Execution"]
        G7["LLM Orchestrator"]
    end

    subgraph Router["🧠 Rust Router :50051 gRPC"]
        direction TB
        R1["FastEmbed\nall-MiniLM-L6-v2"]
        R2["Semantic Similarity"]
        R3["Tool Matching"]
    end

    subgraph Storage["💾 Data Layer"]
        direction LR
        S1[("Valkey\n(Cache)")]
        S2[("Qdrant\n(Vectors)")]
        S3[("Postgres\n(RBAC/Billing)")]
    end

    subgraph Providers["🤖 LLM Providers"]
        direction LR
        P1["OpenAI"]
        P2["Anthropic"]
        P3["Gemini"]
        P4["Ollama"]
    end

    Clients -->|"REST /v1/*"| Gateway
    Gateway -->|"gRPC"| Router
    Router --> S2
    Gateway --> S1
    Gateway --> S3
    Gateway -->|"HTTP"| Providers
    Gateway -->|"MCP Protocol"| Tools["🔧 MCP Tools"]
```

---

## 2. Evolution Pipeline (E1–E6) — Request Flow

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': { 'primaryColor': '#00f3ff', 'primaryTextColor': '#fff', 'primaryBorderColor': '#00f3ff', 'lineColor': '#666'}}}%%

flowchart LR
    Request["📨 Incoming\nRequest"] --> RL["🚦 Rate Limit\n& Auth"]
    RL --> BL["💰 Billing\nPre-check"]
    BL --> E1

    subgraph Pipeline["Evolution Pipeline"]
        direction LR
        E1["🔬 E1\nEntity\nExtraction\n<1ms"]
        E1 --> L1["⚡ L1\nLiteral\nHash"]
        L1 --> L15["🔄 L1.5\nCanonical\nHash"]
        L15 --> E2["⚡ E2/L1b\nEntity-Keyed\nHot Path"]
        E2 --> L2["🧠 L2\nSemantic\nSimilarity"]
    end

    L2 -->|"MISS"| E4["🔄 E4\nWorkflow\nShortcut"]
    E4 --> Tools["🔧 Tool\nExecution"]
    Tools --> LLM["🤖 LLM\nSynthesis"]
    LLM --> Cache["💾 Cache\nSet"]
    Cache --> Response["✅ Response"]

    L1 -->|"HIT"| Response
    L15 -->|"HIT"| Response
    E2 -->|"HIT"| Response
    L2 -->|"HIT"| Response
    E4 -->|"SHORTCUT"| Response

    subgraph Async["⏳ Async (Non-blocking)"]
        E3["📊 E3\nOffline\nLearning"]
        E5["📈 E5\nGPU Avoidance\nMetrics"]
        E6["🔮 E6\nPattern\nMining"]
    end

    Cache -.->|"Event Bus"| Async
```

---

## 3. Four-Layer Cache Architecture

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': { 'primaryColor': '#00f3ff', 'primaryTextColor': '#fff'}}}%%

flowchart TD
    P["Prompt"] --> L1

    L1{"L1: Literal Hash\n(SHA-256)"}
    L1 -->|"HIT ⚡ <1ms"| R["Cached Response"]
    L1 -->|"MISS"| L15

    L15{"L1.5: Canonical Hash\n(Normalized + SHA-256)"}
    L15 -->|"HIT ⚡ <1ms"| R
    L15 -->|"MISS"| L1b

    L1b{"L1b: Entity-Keyed\n(Deterministic Valkey Key)"}
    L1b -->|"HIT ⚡ <2ms"| R
    L1b -->|"MISS"| L2

    L2{"L2: Semantic Similarity\n(Qdrant Vector Search\n+ Entity Post-Filter)"}
    L2 -->|"HIT 🧠 ~15ms"| R
    L2 -->|"MISS"| LLM["LLM Inference\n~2-5s"]
    LLM --> Store["Store in All Layers"]
    Store --> R

    style L1 fill:#1a1a2e,stroke:#00f3ff,color:#fff
    style L15 fill:#1a1a2e,stroke:#7c3aed,color:#fff
    style L1b fill:#1a1a2e,stroke:#22c55e,color:#fff
    style L2 fill:#1a1a2e,stroke:#f59e0b,color:#fff
    style R fill:#0d3320,stroke:#22c55e,color:#fff
    style LLM fill:#3b1323,stroke:#ef4444,color:#fff
```

---

## 4. Entity Extraction Guard (Why Similarity Alone Fails)

```mermaid
%%{init: {'theme': 'dark'}}%%

flowchart LR
    subgraph Problem["❌ Without Entity Guard"]
        P1["Transfer $100\nfrom 123 → 456"]
        P2["Transfer $100\nfrom 456 → 123"]
        P1 -.->|"Similarity: 0.98\nFALSE CACHE HIT!"| P2
    end

    subgraph Solution["✅ With Entity Guard"]
        S1["Transfer $100\nfrom 123 → 456"]
        S2["Transfer $100\nfrom 456 → 123"]
        E["Entity Key:\norg:x:m:ollama:e:\naccount_from=123\naccount_to=456"]
        F["Entity Key:\norg:x:m:ollama:e:\naccount_from=456\naccount_to=123"]
        S1 --> E
        S2 --> F
        E -.-|"Different Keys\n→ CACHE MISS ✅"| F
    end
```

---

## 5. Deployment Architecture (Docker Compose)

```mermaid
%%{init: {'theme': 'dark'}}%%

graph TB
    subgraph Docker["Docker Network: memzent-net"]
        GW["gateway:8080\n(Go 1.25)"]
        RT["router:50051\n(Rust/Tonic)"]
        VK["valkey:6379\n(Cache)"]
        QD["qdrant:6333/6334\n(Vectors)"]
        PG["Supabase Postgres\n(External)"]
        DB["dashboard:3000\n(Next.js 16)"]
        WS["website:5173\n(Vite + React)"]
    end

    Users["Users"] --> DB
    Users --> WS
    DB -->|"REST /v1/*"| GW
    GW -->|"gRPC :50051"| RT
    GW --> VK
    RT --> QD
    GW --> PG

    style GW fill:#1a4a3a,stroke:#22c55e
    style RT fill:#1a2a4a,stroke:#3b82f6
    style VK fill:#4a1a1a,stroke:#ef4444
    style QD fill:#4a3a1a,stroke:#f59e0b
    style DB fill:#1a1a4a,stroke:#7c3aed
    style WS fill:#1a1a4a,stroke:#7c3aed
```

---

## 6. GPU Avoidance Funnel

```mermaid
%%{init: {'theme': 'dark'}}%%

flowchart TD
    Total["100% Requests"] --> L1Hit["L1/L1.5 Cache Hit\n~30%"]
    Total --> L1bHit["L1b Entity Cache Hit\n~20%"]
    Total --> L2Hit["L2 Semantic Hit\n~25%"]
    Total --> WF["Workflow Shortcut\n~5%"]
    Total --> LLM["LLM Required\n~20%"]

    L1Hit --> Avoided["✅ 80% GPU Avoided"]
    L1bHit --> Avoided
    L2Hit --> Avoided
    WF --> Avoided

    LLM --> GPU["❌ GPU Used\n~20%"]

    style Avoided fill:#0d3320,stroke:#22c55e,color:#fff
    style GPU fill:#3b1323,stroke:#ef4444,color:#fff
    style Total fill:#1a1a2e,stroke:#00f3ff,color:#fff
```

---

## Usage

### Render to PNG (for LinkedIn)
1. Go to [mermaid.live](https://mermaid.live)
2. Paste any diagram above
3. Click "Actions" → Download PNG
4. **LinkedIn recommended**: 1200×628px

### Render locally with CLI
```bash
# Install mermaid-cli
npm install -g @mermaid-js/mermaid-cli

# Render a specific diagram (extract the mermaid block first)
mmdc -i diagram.mmd -o architecture.png -t dark -w 1200 -H 628 -b '#0a0a1a'
```

### Best diagram for each context

| Context | Recommended Diagram |
|---------|-------------------|
| **LinkedIn intro post** | #1 High-Level Architecture |
| **Technical blog** | #2 Evolution Pipeline |
| **Explaining the cache** | #3 Four-Layer Cache |
| **Why entity extraction?** | #4 Entity Guard |
| **Self-hosting docs** | #5 Deployment |
| **Business value pitch** | #6 GPU Avoidance Funnel |
