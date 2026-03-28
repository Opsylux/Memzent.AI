# Project Aura: Enterprise AI Infrastructure

**Aura** is a high-performance, semantic AI gateway designed to eliminate the "Token Tax" and security risks associated with unmanaged LLM deployments. By acting as a secure, intelligent middleware between agents and the **Model Context Protocol (MCP)**, Aura ensures every interaction is cost-optimized, audited, and resilient.

---

## 🏗️ Core Architecture

Aura utilizes a distributed, multi-language architecture to balance high-speed semantic routing with robust business logic.

| Service | Language | Port | Role |
| :--- | :--- | :--- | :--- |
| **Gateway** | Go 1.25 | `8080` | Entry point, RBAC, JWT Auth, Semantic Caching |
| **Router** | Rust | `50051` | gRPC service for vector-based tool selection |
| **Dashboard**| Next.js | `3000` | Administrative control tower & observability |
| **MCP Server**| Go | `50052` | Tool execution & context protocol adapter |
| **Website** | Vite 8 | `5173` | Marketing landing page & user portal |

---

## 🛡️ Enterprise Pillars

### 1. Semantic Tool Routing (Rust + Qdrant)
Aura analyzes user intent in real-time and injects only the most relevant tools into the LLM context. This **reduces token waste by up to 90%** and eliminates "Lost in the Middle" hallucinations.

### 2. Bulletproof Governance (Go + Postgres)
Centralized **Role-Based Access Control (RBAC)** and hardware-backed JWT authentication ensure that AI agents only access the data they are authorized to see.

### 3. Deep Observability (Prometheus + OpenTelemetry)
Monitor latency, success rates, and token flow across your entire agentic fleet. Every decision made by the semantic router is logged and traceable.

### 4. ROI Engine (Valkey + Semantic Cache)
Avoid redundant LLM calls. Aura caches repeat intents semantically, delivering **sub-15ms response times** for cached queries and zero-cost retrieval.

---

## 🚀 Getting Started

### Prerequisites
- **Docker & Docker Compose**
- **Node 24+** (for local development)
- **Go 1.25+** (for gateway development)

### One-Command Deployment
The easiest way to start the entire Aura stack is via Docker Compose. This initializes all services including the vector store (Qdrant) and cache (Valkey).

```powershell
docker-compose up -d --build
```

### Service Access
- **Aura Website**: [http://localhost:5173](http://localhost:5173)
- **Admin Dashboard**: [http://localhost:3000](http://localhost:3000)
- **Gateway API**: [http://localhost:8080/v1/chat](http://localhost:8080/v1/chat)
- **Qdrant UI**: [http://localhost:6333/dashboard](http://localhost:6333/dashboard)

---

## 📂 Project Structure

```bash
AuraMCP/
├── services/
│   ├── gateway/      # Go 1.25: Primary Proxy & Auth
│   ├── router/       # Rust: Semantic Decision Engine
│   ├── mcp-server/   # Go: Dynamic Tool Provider
│   ├── dashboard/    # Next.js: Control Tower
│   └── website/      # Vite 8: Brand & Marketing
├── proto/            # Shared gRPC Definitions
├── data/             # Persistent Storage (Postgres/Qdrant)
└── docker-compose.yml # Orchestration Layer
```

---

## 🛠️ Development Workflow

1. **Gateway Logic**: Edit `services/gateway/internal/engine/engine.go` to modify the orchestration flow.
2. **Branding**: Update `website/src/App.tsx` for marketing and UI improvements.
3. **Routing**: Enhance the Rust service in `services/router/src/` for better intent matching.

---

**Built by the Aura Engineering Team.** *Securing the future of Agentic Intelligence.*
