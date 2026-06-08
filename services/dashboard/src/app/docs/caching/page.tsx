import { Layers, Zap, Database, Clock, ToggleLeft } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";
import { CodeBlock } from "@/components/docs/code-block";
import { DOCS_CONFIG } from "@/config/docs-config";

export default function CachingPage() {
  const skipCacheHeader = `curl -X POST https://${DOCS_CONFIG.domain}/v1/chat \\
  -H "X-API-Key: memzent_YOUR_KEY" \\
  -H "X-Skip-Cache: true" \\
  -H "Content-Type: application/json" \\
  -d '{
    "messages": [{"role": "user", "content": "Latest news on AI regulations"}]
  }'`;

  const skipCacheBody = `curl -X POST https://${DOCS_CONFIG.domain}/v1/chat \\
  -H "X-API-Key: memzent_YOUR_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "messages": [{"role": "user", "content": "Latest news on AI regulations"}],
    "skip_cache": true
  }'`;

  const cacheHitResponse = `HTTP/1.1 200 OK
X-Cache: HIT
X-Request-ID: a1b2c3d4...
Content-Type: application/json

{
  "text": "AI regulations have been evolving...",
  "cached": true,
  "request_id": "a1b2c3d4..."
}`;

  const cacheMissResponse = `HTTP/1.1 200 OK
X-Cache: MISS
X-Request-ID: e5f6g7h8...
Content-Type: application/json

{
  "text": "AI regulations have been evolving...",
  "cached": false,
  "provider": "OpenAI (gpt-4o-mini)",
  "request_id": "e5f6g7h8..."
}`;

  const thresholdGet = `curl -X GET https://${DOCS_CONFIG.domain}/v1/settings/threshold \\
  -H "X-API-Key: memzent_YOUR_KEY"

# Response: {"similarity_threshold": 0.95}`;

  const thresholdSet = `curl -X PUT https://${DOCS_CONFIG.domain}/v1/settings/threshold \\
  -H "X-API-Key: memzent_YOUR_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{"similarity_threshold": 0.92}'`;

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-3 mb-4">
        <div className="p-2 rounded-xl bg-memzent-glow/10 border border-memzent-glow/20">
          <Layers size={20} className="text-memzent-glow" />
        </div>
        <h1 className="text-3xl font-black tracking-tight">Caching Deep Dive</h1>
      </div>
      <p className="text-white/50 text-sm leading-relaxed mb-10">
        Memzent&apos;s triple-layer semantic cache reduces latency by up to 95% and costs by up to 90%.
        Understand how each layer works, when to bypass cache, and how to tune the similarity threshold.
      </p>

      {/* Three Layers */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Layers size={16} className="text-memzent-glow" />
          Three-Layer Cache Architecture
        </h2>

        <div className="space-y-4">
          <div className="p-4 rounded-xl border border-memzent-glow/20 bg-memzent-glow/5">
            <div className="flex items-center gap-2 mb-2">
              <span className="text-xs font-black text-memzent-glow bg-memzent-glow/10 px-2 py-0.5 rounded">Layer 1</span>
              <h4 className="text-sm font-black text-white">Literal Match</h4>
            </div>
            <p className="text-xs text-white/50">
              Exact SHA-256 hash of the prompt text. Fastest lookup — sub-millisecond via Valkey.
              Matches only if the prompt is character-for-character identical.
            </p>
          </div>

          <div className="p-4 rounded-xl border border-blue-500/20 bg-blue-500/5">
            <div className="flex items-center gap-2 mb-2">
              <span className="text-xs font-black text-blue-400 bg-blue-500/10 px-2 py-0.5 rounded">Layer 1.5</span>
              <h4 className="text-sm font-black text-white">Canonical Match</h4>
            </div>
            <p className="text-xs text-white/50">
              Normalized version of the prompt — extra whitespace removed, lowercased, punctuation normalized.
              Catches formatting differences: <code>&quot;What is AI?&quot;</code> matches <code>&quot;what is ai ?&quot;</code>.
            </p>
          </div>

          <div className="p-4 rounded-xl border border-purple-500/20 bg-purple-500/5">
            <div className="flex items-center gap-2 mb-2">
              <span className="text-xs font-black text-purple-400 bg-purple-500/10 px-2 py-0.5 rounded">Layer 2</span>
              <h4 className="text-sm font-black text-white">Semantic Match</h4>
            </div>
            <p className="text-xs text-white/50">
              Vector similarity via the Rust Router + Qdrant. Embeds the prompt using <code>all-MiniLM-L6-v2</code> (384-dim)
              and searches for semantically similar previously-answered prompts. Threshold: <strong>0.95</strong>.
              Includes a numeric guard that prevents false positives when parameter values differ.
            </p>
          </div>
        </div>
      </section>

      {/* Cache Scoping */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Database size={16} className="text-memzent-glow" />
          Cache Isolation
        </h2>
        <p className="text-white/50 text-sm mb-4">
          Cache entries are scoped by multiple dimensions to prevent cross-contamination:
        </p>
        <div className="overflow-x-auto">
          <table className="w-full text-xs border border-white/5 rounded-lg overflow-hidden">
            <thead>
              <tr className="bg-white/[0.03] border-b border-white/5">
                <th className="text-left px-4 py-2 font-black text-white/60">Dimension</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Effect</th>
              </tr>
            </thead>
            <tbody className="text-white/40">
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono">org_id</td><td className="px-4 py-2">Org A&apos;s cache never leaks to Org B</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono">model</td><td className="px-4 py-2">GPT-4 responses separated from Claude responses</td></tr>
              <tr><td className="px-4 py-2 font-mono">cache_type</td><td className="px-4 py-2">Literal (p), canonical (c), semantic (s) stored separately</td></tr>
            </tbody>
          </table>
        </div>
        <p className="text-white/40 text-xs mt-3">
          Cache key format: <code className="text-memzent-glow/70">org:{`{org_id}`}:{`{type}`}:{`{model}`}:{`{hash}`}</code>
        </p>
      </section>

      {/* Durable Fallback */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Clock size={16} className="text-memzent-glow" />
          Durable Fallback (Zero-Loss)
        </h2>
        <p className="text-white/50 text-sm leading-relaxed">
          All cache entries are persisted to PostgreSQL (<code className="text-memzent-glow/70">persistent_cache</code> table)
          using write-through. If Valkey crashes or restarts, the gateway automatically reads from Postgres
          and backfills Valkey asynchronously — maintaining 100% cache availability with zero added latency
          on the hot path.
        </p>
      </section>

      {/* Bypassing Cache */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <ToggleLeft size={16} className="text-memzent-glow" />
          Bypassing Cache
        </h2>
        <p className="text-white/50 text-sm mb-4">
          For time-sensitive queries or when you need a fresh LLM response, bypass cache using either method:
        </p>

        <h4 className="text-xs font-black uppercase tracking-widest text-white/40 mb-2">Option 1: Header</h4>
        <CodeBlock code={skipCacheHeader} language="bash" title="X-Skip-Cache Header" />

        <h4 className="text-xs font-black uppercase tracking-widest text-white/40 mb-2 mt-6">Option 2: Body Field</h4>
        <CodeBlock code={skipCacheBody} language="bash" title="skip_cache Field" />
      </section>

      {/* Response Headers */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4">Cache Response Headers</h2>
        <p className="text-white/50 text-sm mb-4">
          The <code className="text-memzent-glow/70">X-Cache</code> response header and
          <code className="text-memzent-glow/70 ml-1">cached</code> body field tell you whether the response was served from cache:
        </p>

        <h4 className="text-xs font-black uppercase tracking-widest text-white/40 mb-2">Cache HIT</h4>
        <CodeBlock code={cacheHitResponse} language="http" title="Cache Hit Response" />

        <h4 className="text-xs font-black uppercase tracking-widest text-white/40 mb-2 mt-6">Cache MISS</h4>
        <CodeBlock code={cacheMissResponse} language="http" title="Cache Miss Response" />
      </section>

      {/* Tuning Threshold */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Zap size={16} className="text-memzent-glow" />
          Tuning the Similarity Threshold
        </h2>
        <p className="text-white/50 text-sm mb-4">
          The semantic cache threshold controls how similar two prompts must be to trigger a cache hit.
          Higher values = stricter matching (fewer false positives), lower values = more cache hits (risk of wrong answers).
        </p>

        <div className="overflow-x-auto mb-6">
          <table className="w-full text-xs border border-white/5 rounded-lg overflow-hidden">
            <thead>
              <tr className="bg-white/[0.03] border-b border-white/5">
                <th className="text-left px-4 py-2 font-black text-white/60">Threshold</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Behavior</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Best For</th>
              </tr>
            </thead>
            <tbody className="text-white/40">
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono">0.98+</td><td className="px-4 py-2">Ultra-strict (near-identical only)</td><td className="px-4 py-2">Financial, medical, legal</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-memzent-glow/70">0.95</td><td className="px-4 py-2">Default — balanced correctness/savings</td><td className="px-4 py-2">General use</td></tr>
              <tr><td className="px-4 py-2 font-mono">0.90</td><td className="px-4 py-2">Aggressive — more hits, slight risk</td><td className="px-4 py-2">FAQ bots, repetitive queries</td></tr>
            </tbody>
          </table>
        </div>

        <CodeBlock code={thresholdGet} language="bash" title="Check Current Threshold" />
        <CodeBlock code={thresholdSet} language="bash" title="Update Threshold" />
      </section>

      {/* Billing impact */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4">Cost Impact</h2>
        <p className="text-white/50 text-sm leading-relaxed">
          Cache hits receive a <strong>90% discount</strong> on billing compared to full LLM calls.
          This means a $1.00 prompt costs only $0.10 when served from cache. Monitor your cache savings
          in the <code className="text-memzent-glow/70">/v1/billing/budget</code> response under <code className="text-memzent-glow/70">cache_savings</code>.
        </p>
      </section>

      <DocsPager />
    </div>
  );
}
