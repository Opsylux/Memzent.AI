# Aura: Feature Implementation Instructions

This document provides step-by-step checklists for common engineering tasks in Aura. Follow these to ensure architectural consistency across service boundaries.

---

## 1. Adding a New LLM Provider
Aura supports pluggable LLM providers in the Gateway.

### Checklist:
1.  **Define Provider**: Create a new file in `/services/gateway/internal/llm/` (e.g., `mistral.go`).
2.  **Implement Interface**: Ensure the struct satisfies the `llm.Provider` interface:
    - `Generate(ctx, prompt, modelOverride) (string, error)`
    - `Name() string`
3.  **Register Provider**: In `services/gateway/main.go`, instantiate the provider and add it to the `providers` map.
4.  **Config**: Update `config.go` to include API keys for the new provider.

> [!IMPORTANT]
> **Pop Question**: If the new provider is a local model, should it use the standard HTTP client or a specialized connector?
> *   *Standard*: Use standard HTTP unless the model requires manual gRPC or socket handling.

---

## 2. Implementing a Phase 3 Tool Connector
Connectors allow the Gateway to execute non-MCP tools (SQL, REST, GraphQL).

### Checklist:
1.  **Engine Definition**: Confirm the `ConnectorType` is defined in `internal/tools/registry.go`.
2.  **Connector Logic**: In `internal/connectors/`, implement the `Connector` interface:
    - `Execute(ctx, req) (*ExecutionResponse, error)`
    - `Validate(req) error`
    - `Type() ConnectorType`
3.  **Registry Integration**: Ensure the tool's `connector_type` in the database matches your implementation.
4.  **Dispatcher**: Verify `internal/connectors/registry.go` correctly routes incoming requests to your new connector.

> [!IMPORTANT]
> **Pop Question**: Does every execution of a tool need to be logged in the Audit Log?
> *   *Yes*: Authentication middleware and the Aura Engine handle this; do not reinvent logging inside the connector.

---

## 3. Modifying gRPC Services (Gateway ↔ Router)
All "brain" logic happens in Rust, called via gRPC.

### Checklist:
1.  **Protobuf**: Edit `/proto/router.proto`.
2.  **Contract Check**: Run the build script in both `gateway` and `router` to generate new bindings.
3.  **Rust Handler**: implement the new service trait in `services/router/src/main.rs`.
4.  **Go Client**: Update `services/gateway/internal/router/client.go` to expose the new gRPC call to the rest of the Gateway.

---

## 4. Dashboard Styling (The "Neural" Look)
The Command Center uses a high-end, dark, glassmorphic aesthetic.

### Guidelines:
- **Backgrounds**: Always use the `neural-bg` utility.
- **Components**: Use `stat-card` for data displays.
- **Accents**: Use `glow-cyan` for primary actions and `glow-purple` for secondary/AI-driven states.
- **Tailwind**: Do NOT use `tailwind.config.js`. Use `@theme inline` in `globals.css`.

> [!TIP]
> **Pop Question**: Should I use standard Radix UI colors or Aura theme tokens?
> *   *Aura Tokens*: Always use `--color-aura-*` variables for a cohesive enterprise feel.
