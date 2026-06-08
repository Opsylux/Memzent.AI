import { Bell, Zap, Shield, Code, RefreshCw } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";
import { CodeBlock } from "@/components/docs/code-block";
import { DOCS_CONFIG } from "@/config/docs-config";

export default function WebhooksPage() {
  const createWebhook = `curl -X POST https://${DOCS_CONFIG.domain}/v1/webhooks \\
  -H "X-API-Key: memzent_YOUR_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "url": "https://your-app.com/webhooks/memzent",
    "events": ["cache_hit", "tool_execution", "rate_limit"],
    "description": "Production alerts"
  }'`;

  const webhookResponse = `{
  "id": "wh_8f2a4b6c...",
  "org_id": "5127e445-bb64-4057...",
  "url": "https://your-app.com/webhooks/memzent",
  "secret": "whsec_a1b2c3d4e5f6...",
  "events": ["cache_hit", "tool_execution", "rate_limit"],
  "enabled": true,
  "description": "Production alerts",
  "created_at": "2026-06-01T10:00:00Z"
}`;

  const eventPayload = `{
  "id": "evt_9x8y7z...",
  "type": "cache_hit",
  "org_id": "5127e445-bb64-4057...",
  "timestamp": "2026-06-06T05:22:11Z",
  "data": {
    "query": "what is quantum computing",
    "score": 0.97,
    "latency_ms": 12,
    "model": "gpt-4o-mini"
  }
}`;

  const verifySignature = `import hmac
import hashlib

def verify_webhook(payload: bytes, signature: str, secret: str) -> bool:
    expected = hmac.new(
        secret.encode(),
        payload,
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(f"sha256={expected}", signature)

# In your webhook handler:
# signature = request.headers["X-Memzent-Signature"]
# verify_webhook(request.body, signature, "whsec_a1b2c3...")`;

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-3 mb-4">
        <div className="p-2 rounded-xl bg-memzent-glow/10 border border-memzent-glow/20">
          <Bell size={20} className="text-memzent-glow" />
        </div>
        <h1 className="text-3xl font-black tracking-tight">Webhooks</h1>
      </div>
      <p className="text-white/50 text-sm leading-relaxed mb-10">
        Subscribe to real-time events from the Memzent gateway. Get notified about cache hits,
        tool executions, rate limiting, and more — delivered to your endpoint with cryptographic verification.
      </p>

      {/* Event Types */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Zap size={16} className="text-memzent-glow" />
          Event Types
        </h2>

        <div className="overflow-x-auto mb-6">
          <table className="w-full text-xs border border-white/5 rounded-lg overflow-hidden">
            <thead>
              <tr className="bg-white/[0.03] border-b border-white/5">
                <th className="text-left px-4 py-2 font-black text-white/60">Event</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Fired When</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Data Fields</th>
              </tr>
            </thead>
            <tbody className="text-white/40">
              <tr className="border-b border-white/5">
                <td className="px-4 py-2 font-mono text-memzent-glow/70">cache_hit</td>
                <td className="px-4 py-2">A prompt matches the semantic cache</td>
                <td className="px-4 py-2">query, score, latency_ms, model</td>
              </tr>
              <tr className="border-b border-white/5">
                <td className="px-4 py-2 font-mono text-memzent-glow/70">tool_execution</td>
                <td className="px-4 py-2">A registered tool is invoked</td>
                <td className="px-4 py-2">tool_name, duration_ms, success</td>
              </tr>
              <tr className="border-b border-white/5">
                <td className="px-4 py-2 font-mono text-memzent-glow/70">rate_limit</td>
                <td className="px-4 py-2">A request is rate-limited</td>
                <td className="px-4 py-2">user_id, tier, limit, window</td>
              </tr>
              <tr className="border-b border-white/5">
                <td className="px-4 py-2 font-mono text-memzent-glow/70">key_rotated</td>
                <td className="px-4 py-2">An API key is rotated</td>
                <td className="px-4 py-2">key_prefix, rotated_by</td>
              </tr>
              <tr className="border-b border-white/5">
                <td className="px-4 py-2 font-mono text-memzent-glow/70">tool_registered</td>
                <td className="px-4 py-2">A new tool is registered</td>
                <td className="px-4 py-2">tool_name, connector_type</td>
              </tr>
              <tr>
                <td className="px-4 py-2 font-mono text-memzent-glow/70">session_created</td>
                <td className="px-4 py-2">A new session is started</td>
                <td className="px-4 py-2">session_id, title</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* Creating */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Code size={16} className="text-memzent-glow" />
          Creating a Webhook
        </h2>
        <CodeBlock code={createWebhook} language="bash" title="Create Webhook" />
        <CodeBlock code={webhookResponse} language="json" title="Response (save the secret!)" />

        <div className="p-4 rounded-xl border border-yellow-500/20 bg-yellow-500/5 mt-4">
          <p className="text-xs text-yellow-200/70">
            <strong>Important:</strong> The <code>secret</code> is only returned on creation.
            Store it securely — you&apos;ll need it to verify webhook signatures.
          </p>
        </div>
      </section>

      {/* Payload */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4">Event Payload Structure</h2>
        <p className="text-white/50 text-sm mb-4">
          Every webhook delivery has the same envelope structure with event-specific data:
        </p>
        <CodeBlock code={eventPayload} language="json" title="Webhook Payload" />
      </section>

      {/* Headers */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Shield size={16} className="text-memzent-glow" />
          Delivery Headers
        </h2>
        <div className="overflow-x-auto">
          <table className="w-full text-xs border border-white/5 rounded-lg overflow-hidden">
            <thead>
              <tr className="bg-white/[0.03] border-b border-white/5">
                <th className="text-left px-4 py-2 font-black text-white/60">Header</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Description</th>
              </tr>
            </thead>
            <tbody className="text-white/40">
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono">X-Memzent-Signature</td><td className="px-4 py-2">HMAC-SHA256 signature of the payload body</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono">X-Memzent-Event</td><td className="px-4 py-2">Event type (e.g. cache_hit, tool_execution)</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono">X-Memzent-Delivery</td><td className="px-4 py-2">Unique delivery ID for idempotency</td></tr>
              <tr><td className="px-4 py-2 font-mono">User-Agent</td><td className="px-4 py-2">Memzent-Webhook/1.0</td></tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* Verification */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <RefreshCw size={16} className="text-memzent-glow" />
          Verifying Signatures
        </h2>
        <p className="text-white/50 text-sm mb-4">
          Always verify the <code className="text-memzent-glow/70">X-Memzent-Signature</code> header
          to ensure webhook payloads are authentic and haven&apos;t been tampered with.
        </p>
        <CodeBlock code={verifySignature} language="python" title="Python Verification" />
      </section>

      <DocsPager />
    </div>
  );
}
