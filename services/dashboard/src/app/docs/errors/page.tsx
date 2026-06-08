import { AlertTriangle, ShieldX, Clock, Ban, RefreshCw } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";
import { CodeBlock } from "@/components/docs/code-block";

export default function ErrorsPage() {
  const error401 = `{
  "error": "unauthorized: Missing identity (JWT or API Key)"
}`;

  const error403 = `{
  "error": "Forbidden: API key lacks required scope 'chat:execute'"
}`;

  const error429 = `{
  "error": "rate limit exceeded: org tier allows 100 req/min, try again in 12s"
}`;

  const error402 = `{
  "error": "insufficient balance: current balance $0.42, estimated cost $0.85"
}`;

  const error500 = `{
  "error": "internal server error"
}`;

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-3 mb-4">
        <div className="p-2 rounded-xl bg-red-500/10 border border-red-500/20">
          <AlertTriangle size={20} className="text-red-400" />
        </div>
        <h1 className="text-3xl font-black tracking-tight">Errors & Rate Limits</h1>
      </div>
      <p className="text-white/50 text-sm leading-relaxed mb-10">
        Understand error responses, HTTP status codes, and rate limiting behavior
        to build resilient integrations with the Memzent API.
      </p>

      {/* Error Codes Table */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4">HTTP Status Codes</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-xs border border-white/5 rounded-lg overflow-hidden">
            <thead>
              <tr className="bg-white/[0.03] border-b border-white/5">
                <th className="text-left px-4 py-2 font-black text-white/60">Code</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Meaning</th>
                <th className="text-left px-4 py-2 font-black text-white/60">When</th>
              </tr>
            </thead>
            <tbody className="text-white/40">
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-green-400">200</td><td className="px-4 py-2">Success</td><td className="px-4 py-2">Request completed successfully</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-yellow-400">400</td><td className="px-4 py-2">Bad Request</td><td className="px-4 py-2">Invalid JSON, missing required fields</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-red-400">401</td><td className="px-4 py-2">Unauthorized</td><td className="px-4 py-2">Missing or invalid API key / JWT</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-red-400">402</td><td className="px-4 py-2">Payment Required</td><td className="px-4 py-2">Insufficient balance or spend limit exceeded</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-red-400">403</td><td className="px-4 py-2">Forbidden</td><td className="px-4 py-2">Valid auth but lacking required scope/role</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-orange-400">429</td><td className="px-4 py-2">Too Many Requests</td><td className="px-4 py-2">Rate limit exceeded for org/user tier</td></tr>
              <tr><td className="px-4 py-2 font-mono text-red-400">500</td><td className="px-4 py-2">Internal Error</td><td className="px-4 py-2">LLM provider failure, internal issue</td></tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* Error format */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <ShieldX size={16} className="text-red-400" />
          Error Response Format
        </h2>
        <p className="text-white/50 text-sm mb-4">
          All errors return a JSON object with an <code className="text-memzent-glow/70">error</code> field
          containing a human-readable message.
        </p>

        <h4 className="text-xs font-black uppercase tracking-widest text-white/40 mb-2">401 — Missing Authentication</h4>
        <CodeBlock code={error401} language="json" title="401 Unauthorized" />

        <h4 className="text-xs font-black uppercase tracking-widest text-white/40 mb-2 mt-6">403 — Insufficient Scope</h4>
        <CodeBlock code={error403} language="json" title="403 Forbidden" />

        <h4 className="text-xs font-black uppercase tracking-widest text-white/40 mb-2 mt-6">402 — Balance Exhausted</h4>
        <CodeBlock code={error402} language="json" title="402 Payment Required" />

        <h4 className="text-xs font-black uppercase tracking-widest text-white/40 mb-2 mt-6">429 — Rate Limited</h4>
        <CodeBlock code={error429} language="json" title="429 Too Many Requests" />

        <h4 className="text-xs font-black uppercase tracking-widest text-white/40 mb-2 mt-6">500 — Server Error</h4>
        <CodeBlock code={error500} language="json" title="500 Internal Error" />
      </section>

      {/* Rate Limits */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Clock size={16} className="text-memzent-glow" />
          Rate Limits
        </h2>
        <p className="text-white/50 text-sm mb-4">
          Rate limits are enforced per-organization using a distributed sliding window (60 seconds) via Valkey.
          Per-user limits are proportional to role within the org.
        </p>

        <div className="overflow-x-auto mb-6">
          <table className="w-full text-xs border border-white/5 rounded-lg overflow-hidden">
            <thead>
              <tr className="bg-white/[0.03] border-b border-white/5">
                <th className="text-left px-4 py-2 font-black text-white/60">Tier</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Org Limit</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Window</th>
              </tr>
            </thead>
            <tbody className="text-white/40">
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono">free</td><td className="px-4 py-2">10 requests</td><td className="px-4 py-2">60 seconds</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono">pro</td><td className="px-4 py-2">100 requests</td><td className="px-4 py-2">60 seconds</td></tr>
              <tr><td className="px-4 py-2 font-mono">business</td><td className="px-4 py-2">1,000 requests</td><td className="px-4 py-2">60 seconds</td></tr>
            </tbody>
          </table>
        </div>

        <h3 className="text-sm font-black mb-3">Per-User Proportional Limits</h3>
        <p className="text-white/50 text-sm mb-4">
          Within an organization, individual users get a proportional share based on role:
        </p>
        <div className="overflow-x-auto mb-6">
          <table className="w-full text-xs border border-white/5 rounded-lg overflow-hidden">
            <thead>
              <tr className="bg-white/[0.03] border-b border-white/5">
                <th className="text-left px-4 py-2 font-black text-white/60">Role</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Multiplier</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Example (Pro Tier)</th>
              </tr>
            </thead>
            <tbody className="text-white/40">
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono">admin</td><td className="px-4 py-2">Full org limit</td><td className="px-4 py-2">100 req/min</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono">member</td><td className="px-4 py-2">Proportional</td><td className="px-4 py-2">~50 req/min</td></tr>
              <tr><td className="px-4 py-2 font-mono">viewer</td><td className="px-4 py-2">Read-only</td><td className="px-4 py-2">Blocked from execution</td></tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* Handling */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <RefreshCw size={16} className="text-memzent-glow" />
          Handling Errors
        </h2>

        <div className="space-y-4">
          <div className="p-4 rounded-xl border border-white/5 bg-white/[0.02]">
            <h4 className="text-xs font-black text-white/60 mb-2 flex items-center gap-2">
              <Ban size={12} className="text-orange-400" />
              Rate Limited (429)
            </h4>
            <p className="text-xs text-white/40">
              Implement exponential backoff. The error message includes a suggested retry delay.
              Cache-hit responses do NOT count toward rate limits.
            </p>
          </div>

          <div className="p-4 rounded-xl border border-white/5 bg-white/[0.02]">
            <h4 className="text-xs font-black text-white/60 mb-2 flex items-center gap-2">
              <Ban size={12} className="text-red-400" />
              Balance Exhausted (402)
            </h4>
            <p className="text-xs text-white/40">
              Top up via <code className="text-memzent-glow/70">POST /v1/billing/checkout</code> or
              adjust spend limits. Monitor budget with <code className="text-memzent-glow/70">GET /v1/billing/budget</code>.
            </p>
          </div>

          <div className="p-4 rounded-xl border border-white/5 bg-white/[0.02]">
            <h4 className="text-xs font-black text-white/60 mb-2 flex items-center gap-2">
              <Ban size={12} className="text-red-400" />
              Provider Failure (500)
            </h4>
            <p className="text-xs text-white/40">
              If the configured LLM provider is down, retry with a different provider using
              <code className="text-memzent-glow/70 mx-1">X-Memzent-Provider</code> header.
            </p>
          </div>
        </div>
      </section>

      <DocsPager />
    </div>
  );
}
