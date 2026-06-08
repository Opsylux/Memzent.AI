import { Code, Send, Shield, Zap, Database, Globe } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";
import { CodeBlock } from "@/components/docs/code-block";
import { DOCS_CONFIG } from "@/config/docs-config";

export default function APIReferencePage() {
  const chatRequest = `curl -X POST https://${DOCS_CONFIG.domain}/v1/chat \\
  -H "X-API-Key: memzent_YOUR_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "messages": [{"role": "user", "content": "Explain quantum computing"}],
    "provider": "openai",
    "model": "gpt-4o-mini",
    "skip_cache": false,
    "session_id": "optional-session-id"
  }'`;

  const chatResponse = `{
  "text": "Quantum computing leverages quantum mechanical phenomena...",
  "cached": false,
  "provider": "OpenAI (gpt-4o-mini)",
  "request_id": "a1b2c3d4e5f6...",
  "session_id": "sess_abc123"
}`;

  const providersResponse = `[
  { "name": "ollama", "default_model": "llama3.2", "supported_models": ["llama3.2", "llama3", "mistral", "phi3"] },
  { "name": "openai", "default_model": "gpt-4o-mini", "supported_models": ["gpt-4o-mini", "gpt-4", "gpt-4-turbo"] },
  { "name": "anthropic", "default_model": "claude-sonnet-4-20250514", "supported_models": ["claude-sonnet-4-20250514", "claude-opus-4-20250514"] },
  { "name": "gemini", "default_model": "gemini-2.5-flash", "supported_models": ["gemini-2.5-flash", "gemini-2.5-pro", "gemini-2.0-flash"] }
]`;

  const toolsRegister = `curl -X POST https://${DOCS_CONFIG.domain}/v1/tools/register \\
  -H "X-API-Key: memzent_YOUR_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "name": "weather_lookup",
    "description": "Get current weather for a city",
    "endpoint": "https://api.weather.com/v1/current",
    "connector_type": "rest_api",
    "input_schema": {"city": "string"},
    "output_schema": {"temp": "number", "condition": "string"},
    "timeout_seconds": 10,
    "requires_auth": false
  }'`;

  const sessionCreate = `curl -X POST https://${DOCS_CONFIG.domain}/v1/sessions \\
  -H "X-API-Key: memzent_YOUR_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{"title": "Research Session"}'`;

  const webhookCreate = `curl -X POST https://${DOCS_CONFIG.domain}/v1/webhooks \\
  -H "X-API-Key: memzent_YOUR_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "url": "https://your-app.com/webhooks/memzent",
    "events": ["cache_hit", "tool_execution", "rate_limit"],
    "description": "Production webhook"
  }'`;

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-3 mb-4">
        <div className="p-2 rounded-xl bg-memzent-glow/10 border border-memzent-glow/20">
          <Code size={20} className="text-memzent-glow" />
        </div>
        <h1 className="text-3xl font-black tracking-tight">API Reference</h1>
      </div>
      <p className="text-white/50 text-sm leading-relaxed mb-10">
        Complete endpoint reference for the Memzent Gateway API. All endpoints require authentication
        via <code className="text-memzent-glow/80">X-API-Key</code> or <code className="text-memzent-glow/80">Authorization: Bearer</code> header.
      </p>

      {/* Base URL */}
      <div className="p-4 rounded-xl border border-white/5 bg-white/[0.02] mb-8">
        <div className="flex items-center gap-2 mb-2">
          <Globe size={14} className="text-memzent-glow" />
          <span className="text-xs font-black uppercase tracking-widest text-white/40">Base URL</span>
        </div>
        <code className="text-sm text-memzent-glow font-mono">https://{DOCS_CONFIG.domain}</code>
      </div>

      {/* Chat Endpoint */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Send size={16} className="text-memzent-glow" />
          POST /v1/chat
        </h2>
        <p className="text-white/50 text-sm mb-4">
          Send a prompt to the Memzent gateway. The engine checks cache, routes semantically,
          executes matched tools, and synthesizes an LLM response.
        </p>

        <div className="overflow-x-auto mb-4">
          <table className="w-full text-xs border border-white/5 rounded-lg overflow-hidden">
            <thead>
              <tr className="bg-white/[0.03] border-b border-white/5">
                <th className="text-left px-4 py-2 font-black text-white/60">Field</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Type</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Required</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Description</th>
              </tr>
            </thead>
            <tbody className="text-white/40">
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-memzent-glow/70">messages</td><td className="px-4 py-2">Message[]</td><td className="px-4 py-2">✅</td><td className="px-4 py-2">Array of {`{role, content}`} objects</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-memzent-glow/70">provider</td><td className="px-4 py-2">string</td><td className="px-4 py-2">—</td><td className="px-4 py-2">ollama, openai, anthropic, gemini</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-memzent-glow/70">model</td><td className="px-4 py-2">string</td><td className="px-4 py-2">—</td><td className="px-4 py-2">Model override (e.g. gpt-4, claude-sonnet-4-20250514)</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-memzent-glow/70">skip_cache</td><td className="px-4 py-2">boolean</td><td className="px-4 py-2">—</td><td className="px-4 py-2">Bypass semantic cache (default: false)</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-memzent-glow/70">session_id</td><td className="px-4 py-2">string</td><td className="px-4 py-2">—</td><td className="px-4 py-2">Attach to conversation session for memory</td></tr>
              <tr><td className="px-4 py-2 font-mono text-memzent-glow/70">stream</td><td className="px-4 py-2">boolean</td><td className="px-4 py-2">—</td><td className="px-4 py-2">Enable streaming response</td></tr>
            </tbody>
          </table>
        </div>

        <h4 className="text-xs font-black uppercase tracking-widest text-white/40 mb-2">Headers</h4>
        <div className="overflow-x-auto mb-4">
          <table className="w-full text-xs border border-white/5 rounded-lg overflow-hidden">
            <thead>
              <tr className="bg-white/[0.03] border-b border-white/5">
                <th className="text-left px-4 py-2 font-black text-white/60">Header</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Description</th>
              </tr>
            </thead>
            <tbody className="text-white/40">
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono">X-API-Key</td><td className="px-4 py-2">Your Memzent API key (required if no Bearer token)</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono">X-Memzent-Provider</td><td className="px-4 py-2">Override provider (alternative to body field)</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono">X-Memzent-Model</td><td className="px-4 py-2">Override model (alternative to body field)</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono">X-Skip-Cache</td><td className="px-4 py-2">Set to &quot;true&quot; to bypass cache</td></tr>
              <tr><td className="px-4 py-2 font-mono">X-Request-ID</td><td className="px-4 py-2">Optional request tracking ID (auto-generated if absent)</td></tr>
            </tbody>
          </table>
        </div>

        <h4 className="text-xs font-black uppercase tracking-widest text-white/40 mb-2">Response Headers</h4>
        <div className="overflow-x-auto mb-4">
          <table className="w-full text-xs border border-white/5 rounded-lg overflow-hidden">
            <tbody className="text-white/40">
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono">X-Cache</td><td className="px-4 py-2">HIT or MISS — indicates whether response was served from cache</td></tr>
              <tr><td className="px-4 py-2 font-mono">X-Request-ID</td><td className="px-4 py-2">Request tracking identifier</td></tr>
            </tbody>
          </table>
        </div>

        <CodeBlock code={chatRequest} language="bash" title="Request" />
        <CodeBlock code={chatResponse} language="json" title="Response" />
      </section>

      {/* Providers */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Zap size={16} className="text-memzent-glow" />
          GET /v1/providers
        </h2>
        <p className="text-white/50 text-sm mb-4">
          List all configured LLM providers and their supported models.
        </p>
        <CodeBlock code={providersResponse} language="json" title="Response" />
      </section>

      {/* Models */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Zap size={16} className="text-memzent-glow" />
          GET /v1/models
        </h2>
        <p className="text-white/50 text-sm mb-4">
          List all available models grouped by provider. Includes dynamically discovered models from each provider.
        </p>
      </section>

      {/* Tools */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Database size={16} className="text-memzent-glow" />
          Tools API
        </h2>

        <div className="overflow-x-auto mb-6">
          <table className="w-full text-xs border border-white/5 rounded-lg overflow-hidden">
            <thead>
              <tr className="bg-white/[0.03] border-b border-white/5">
                <th className="text-left px-4 py-2 font-black text-white/60">Method</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Endpoint</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Description</th>
              </tr>
            </thead>
            <tbody className="text-white/40">
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-green-400">GET</td><td className="px-4 py-2 font-mono">/v1/tools</td><td className="px-4 py-2">List all registered tools for your org</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-yellow-400">POST</td><td className="px-4 py-2 font-mono">/v1/tools/register</td><td className="px-4 py-2">Register a new tool</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-yellow-400">POST</td><td className="px-4 py-2 font-mono">/v1/tools/sync</td><td className="px-4 py-2">Sync tools from MCP server</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-green-400">GET</td><td className="px-4 py-2 font-mono">/v1/tools/status</td><td className="px-4 py-2">Registry health status</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-green-400">GET</td><td className="px-4 py-2 font-mono">/v1/tools/{`{id}`}</td><td className="px-4 py-2">Get tool details</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-blue-400">PUT</td><td className="px-4 py-2 font-mono">/v1/tools/{`{id}`}</td><td className="px-4 py-2">Update tool configuration</td></tr>
              <tr><td className="px-4 py-2 font-mono text-red-400">DELETE</td><td className="px-4 py-2 font-mono">/v1/tools/{`{id}`}</td><td className="px-4 py-2">Disable a tool</td></tr>
            </tbody>
          </table>
        </div>

        <CodeBlock code={toolsRegister} language="bash" title="Register Tool" />
      </section>

      {/* Sessions */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Database size={16} className="text-memzent-glow" />
          Sessions API
        </h2>

        <div className="overflow-x-auto mb-6">
          <table className="w-full text-xs border border-white/5 rounded-lg overflow-hidden">
            <thead>
              <tr className="bg-white/[0.03] border-b border-white/5">
                <th className="text-left px-4 py-2 font-black text-white/60">Method</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Endpoint</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Description</th>
              </tr>
            </thead>
            <tbody className="text-white/40">
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-yellow-400">POST</td><td className="px-4 py-2 font-mono">/v1/sessions</td><td className="px-4 py-2">Create a new conversation session</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-green-400">GET</td><td className="px-4 py-2 font-mono">/v1/sessions/{`{id}`}/messages</td><td className="px-4 py-2">Get session message history</td></tr>
              <tr><td className="px-4 py-2 font-mono text-red-400">DELETE</td><td className="px-4 py-2 font-mono">/v1/sessions/{`{id}`}</td><td className="px-4 py-2">Delete a session</td></tr>
            </tbody>
          </table>
        </div>

        <CodeBlock code={sessionCreate} language="bash" title="Create Session" />
      </section>

      {/* Webhooks */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Zap size={16} className="text-memzent-glow" />
          Webhooks API
        </h2>

        <div className="overflow-x-auto mb-6">
          <table className="w-full text-xs border border-white/5 rounded-lg overflow-hidden">
            <thead>
              <tr className="bg-white/[0.03] border-b border-white/5">
                <th className="text-left px-4 py-2 font-black text-white/60">Method</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Endpoint</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Description</th>
              </tr>
            </thead>
            <tbody className="text-white/40">
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-green-400">GET</td><td className="px-4 py-2 font-mono">/v1/webhooks</td><td className="px-4 py-2">List webhooks</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-yellow-400">POST</td><td className="px-4 py-2 font-mono">/v1/webhooks</td><td className="px-4 py-2">Create webhook subscription</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-blue-400">PUT</td><td className="px-4 py-2 font-mono">/v1/webhooks/{`{id}`}</td><td className="px-4 py-2">Update webhook</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-red-400">DELETE</td><td className="px-4 py-2 font-mono">/v1/webhooks/{`{id}`}</td><td className="px-4 py-2">Delete webhook</td></tr>
              <tr><td className="px-4 py-2 font-mono text-green-400">GET</td><td className="px-4 py-2 font-mono">/v1/webhooks/event-types</td><td className="px-4 py-2">List available event types</td></tr>
            </tbody>
          </table>
        </div>

        <CodeBlock code={webhookCreate} language="bash" title="Create Webhook" />
      </section>

      {/* Billing */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Shield size={16} className="text-memzent-glow" />
          Billing API
        </h2>

        <div className="overflow-x-auto">
          <table className="w-full text-xs border border-white/5 rounded-lg overflow-hidden">
            <thead>
              <tr className="bg-white/[0.03] border-b border-white/5">
                <th className="text-left px-4 py-2 font-black text-white/60">Method</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Endpoint</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Description</th>
              </tr>
            </thead>
            <tbody className="text-white/40">
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-green-400">GET</td><td className="px-4 py-2 font-mono">/v1/billing/budget</td><td className="px-4 py-2">Full budget status, burn rate, projections</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-green-400">GET</td><td className="px-4 py-2 font-mono">/v1/billing/spend-limits</td><td className="px-4 py-2">Current spend vs limits</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-blue-400">PUT</td><td className="px-4 py-2 font-mono">/v1/billing/spend-limits</td><td className="px-4 py-2">Set daily/monthly caps</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-green-400">GET</td><td className="px-4 py-2 font-mono">/v1/billing/spend-timeseries</td><td className="px-4 py-2">Daily spend data (query: ?days=N)</td></tr>
              <tr><td className="px-4 py-2 font-mono text-yellow-400">POST</td><td className="px-4 py-2 font-mono">/v1/billing/checkout</td><td className="px-4 py-2">Create Stripe checkout for top-up</td></tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* Settings */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Zap size={16} className="text-memzent-glow" />
          Other Endpoints
        </h2>

        <div className="overflow-x-auto">
          <table className="w-full text-xs border border-white/5 rounded-lg overflow-hidden">
            <thead>
              <tr className="bg-white/[0.03] border-b border-white/5">
                <th className="text-left px-4 py-2 font-black text-white/60">Method</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Endpoint</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Description</th>
              </tr>
            </thead>
            <tbody className="text-white/40">
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-green-400">GET</td><td className="px-4 py-2 font-mono">/v1/stats</td><td className="px-4 py-2">Cache stats, hit/miss rates, org metrics</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-green-400">GET</td><td className="px-4 py-2 font-mono">/v1/audit</td><td className="px-4 py-2">Audit log entries</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-green-400">GET</td><td className="px-4 py-2 font-mono">/v1/analytics/context</td><td className="px-4 py-2">Analytics context data</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-green-400">GET</td><td className="px-4 py-2 font-mono">/v1/settings/threshold</td><td className="px-4 py-2">Get similarity threshold</td></tr>
              <tr><td className="px-4 py-2 font-mono text-blue-400">PUT</td><td className="px-4 py-2 font-mono">/v1/settings/threshold</td><td className="px-4 py-2">Update similarity threshold</td></tr>
            </tbody>
          </table>
        </div>
      </section>

      <DocsPager />
    </div>
  );
}
