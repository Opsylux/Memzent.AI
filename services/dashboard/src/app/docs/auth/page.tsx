import { Key, Shield, Fingerprint, AlertCircle } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";
import { CodeBlock } from "@/components/docs/code-block";
import { DOCS_CONFIG } from "@/config/docs-config";

export default function AuthPage() {
  const apiKeyExample = `curl -X POST https://${DOCS_CONFIG.domain}/v1/chat \\
  -H "X-API-Key: memzent_f7c9...8e2a" \\
  -d '{"messages": [{"role": "user", "content": "Hello World"}]}'`;

  const jwtExample = `curl -X POST https://${DOCS_CONFIG.domain}/v1/chat \\
  -H "Authorization: Bearer eyJhbGci..." \\
  -d '{"messages": [{"role": "user", "content": "Hello World"}]}'`;

  return (
    <div className="space-y-12">
      <header className="space-y-4">
        <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-memzent-glow/5 border border-memzent-glow/20 w-fit">
          <span className="text-[10px] font-black text-memzent-glow uppercase tracking-tighter italic">Security</span>
        </div>
        <h1 className="text-4xl font-black tracking-tighter uppercase sm:text-5xl">Authentication</h1>
        <p className="text-lg text-white/60 leading-relaxed font-medium">
          Memzent supports two authentication methods. Choose the one that fits your use case — both provide the same level of access and security.
        </p>
      </header>

      {/* Comparison */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
        <div className="p-5 rounded-2xl bg-white/[0.02] border border-white/5 space-y-3">
          <div className="flex items-center gap-2 text-memzent-glow">
            <Key size={16} />
            <span className="text-xs font-black uppercase">API Keys</span>
          </div>
          <p className="text-[11px] text-white/40 font-bold leading-relaxed">Best for server-to-server calls, background jobs, and internal service integrations where a human is not present.</p>
        </div>
        <div className="p-5 rounded-2xl bg-white/[0.02] border border-white/5 space-y-3">
          <div className="flex items-center gap-2 text-memzent-purple">
            <Fingerprint size={16} />
            <span className="text-xs font-black uppercase">JWT Tokens</span>
          </div>
          <p className="text-[11px] text-white/40 font-bold leading-relaxed">Best for web and mobile apps where a logged-in user identity needs to be attached to each request.</p>
        </div>
      </div>

      {/* API Keys */}
      <section className="space-y-5">
        <h2 className="text-2xl font-black tracking-tighter uppercase">API Keys</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          API Keys are long-lived secrets tied to your organization. Generate them in the API Keys section of your Dashboard and pass the key in the <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">X-API-Key</code> header.
        </p>
        <CodeBlock code={apiKeyExample} language="bash" filename="cURL — API Key" />
      </section>

      {/* JWT */}
      <section className="space-y-5 pt-2">
        <h2 className="text-2xl font-black tracking-tighter uppercase">JWT Tokens</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          If your application already uses an identity provider (such as Supabase, Auth0, or Firebase), Memzent can verify the token it issues. Pass the token in the standard <code className="text-memzent-purple bg-memzent-purple/5 px-1 rounded font-mono">Authorization: Bearer</code> header.
        </p>
        <CodeBlock code={jwtExample} language="bash" filename="cURL — JWT Token" />
      </section>

      {/* Zero Trust */}
      <section className="space-y-5 pt-2">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Zero Trust Verification</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Regardless of which method you use, Memzent performs a real-time check against your organization&apos;s live membership records on every request. Tokens and keys are treated as identity hints — permissions are always verified against the source of truth.
        </p>
        <div className="p-5 rounded-2xl bg-white/[0.02] border border-white/5 space-y-4">
          <div className="flex items-center gap-2 text-memzent-accent">
            <Shield size={15} />
            <span className="text-xs font-black uppercase">What this means for you</span>
          </div>
          <ul className="space-y-2 text-[11px] text-white/40 font-bold">
            <li className="flex gap-2"><span className="text-memzent-glow shrink-0">→</span> Revoking a user from your org immediately cuts off their access — no token expiry needed.</li>
            <li className="flex gap-2"><span className="text-memzent-glow shrink-0">→</span> Rotating an API key invalidates the old one instantly across all services.</li>
            <li className="flex gap-2"><span className="text-memzent-glow shrink-0">→</span> Changing a user&apos;s role takes effect on the next request — no cache to flush.</li>
          </ul>
        </div>

        <div className="p-4 rounded-xl bg-white/[0.01] border border-white/5 flex items-start gap-3">
          <AlertCircle size={13} className="text-white/20 mt-0.5 shrink-0" />
          <p className="text-[11px] text-white/30 font-bold leading-relaxed">
            Memzent does not rely on claims embedded inside JWT tokens for authorization decisions. It only uses them to identify the user.
          </p>
        </div>
      </section>

      {/* Scopes */}
      <section className="space-y-5 pt-2">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Permission Scopes</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          API keys are issued with specific scopes that determine what operations they can perform.
          Requests to endpoints requiring a scope the key lacks will receive a <code className="text-red-400 font-mono">403 Forbidden</code>.
        </p>
        <div className="overflow-x-auto">
          <table className="w-full text-xs border border-white/5 rounded-lg overflow-hidden">
            <thead>
              <tr className="bg-white/[0.03] border-b border-white/5">
                <th className="text-left px-4 py-2 font-black text-white/60">Scope</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Grants Access To</th>
              </tr>
            </thead>
            <tbody className="text-white/40">
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-memzent-glow/70">chat:execute</td><td className="px-4 py-2">POST /v1/chat — send prompts and receive responses</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-memzent-glow/70">tools:read</td><td className="px-4 py-2">GET /v1/tools — list registered tools</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-memzent-glow/70">tools:write</td><td className="px-4 py-2">POST/PUT/DELETE /v1/tools — manage tool registry</td></tr>
              <tr><td className="px-4 py-2 font-mono text-memzent-glow/70">audit:read</td><td className="px-4 py-2">GET /v1/audit — view audit logs and analytics</td></tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* Auth Errors */}
      <section className="space-y-5 pt-2">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Error Responses</h2>
        <div className="space-y-3 text-[11px] text-white/40 font-bold">
          <div className="p-3 rounded-lg border border-white/5 bg-white/[0.02]">
            <code className="text-red-400">401</code> — <span className="text-white/60">Missing identity (JWT or API Key)</span> — No auth header provided
          </div>
          <div className="p-3 rounded-lg border border-white/5 bg-white/[0.02]">
            <code className="text-red-400">401</code> — <span className="text-white/60">Invalid Authorization Header Format</span> — Malformed Bearer token
          </div>
          <div className="p-3 rounded-lg border border-white/5 bg-white/[0.02]">
            <code className="text-red-400">403</code> — <span className="text-white/60">Forbidden: API key lacks required scope</span> — Key doesn&apos;t have the needed scope
          </div>
          <div className="p-3 rounded-lg border border-white/5 bg-white/[0.02]">
            <code className="text-red-400">403</code> — <span className="text-white/60">Organizational context missing</span> — Valid auth but no org membership found
          </div>
        </div>
      </section>

      <DocsPager />
    </div>
  );
}
