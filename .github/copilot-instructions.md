# Project Aura: Architectural Abstract

## Overview
Aura is an enterprise-grade AI infrastructure project designed for semantic routing and token optimization. It uses a high-performance Go Gateway for traffic management and a Rust-based Semantic Router for vector-based decision-making.

## Core Architecture
- **Gateway (Go):** Entry point, handles HTTP/Auth, and manages the Semantic Cache via Valkey.
- **Router (Rust):** High-speed gRPC server that interfaces with Qdrant for vector similarity search.
- **Intelligence (Qdrant):** Stores tool metadata and semantic embeddings.
- **Persistence (Postgres):** Stores governance, RBAC, and user audit logs.

## Engineering Standards
- **Communication:** Services communicate strictly via gRPC (defined in `/proto`).
- **Data Safety:** Rust handles all vector math; Go handles all external-facing business logic.
- **Caching:** Every request must check the Valkey cache before hitting the Rust Router.