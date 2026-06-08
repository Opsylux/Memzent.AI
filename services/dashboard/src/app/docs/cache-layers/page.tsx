import { Layers, Zap, Database, Key, BarChart3 } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";
import { CodeBlock } from "@/components/docs/code-block";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Cache Layers & L1b",
  description: "Memzent's 4-layer cache architecture: L1 literal, L1.5 canonical, L1b entity-keyed hot path, and L2 semantic vector similarity.",
};

export default function CacheLayersPage() {
  const l1bKeyFormat = `// L1b Key Format (deterministic, entity-keyed):
org:{org_id}:m:{model}:e:{sorted_key=value_pairs}

// Example:
org:abc123:m:gpt-4o-mini:e:account_dest=456|account_source=123|amount=100

// Properties:
// - Sorted entity pairs → order-independent
// - Includes model → no cross-model contamination
// - Org-scoped → multi-tenant isolation`;

  const cacheStages = `Request Flow Through Cache Layers:
─────────────────────────────────────
│ L1:  Literal Hash (SHA-256)      │  ← Exact match, sub-ms
│ L1.5: Canonical Hash             │  ← Normalized text match
│ L1b: Entity-Keyed Hot Path      │  ← Entity fingerprint in Valkey
│ L2:  Semantic Vector (Qdrant)    │  ← Embedding similarity ≥0.95
─────────────────────────────────────
│ MISS → LLM inference             │
│ Response stored in ALL layers    │
─────────────────────────────────────`;

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-3 mb-4">
        <div className="p-2 rounded-xl bg-memzent-purple/10 border border-memzent-purple/20">
          <Layers size={20} className="text-memzent-purple" />
        </div>
        <h1 className="text-3xl font-black tracking-tight">Cache Layers &amp; L1b Hot Path</h1>
      </div>
      <p className="text-white/50 text-sm leading-relaxed mb-10">
        Memzent uses a 4-layer caching architecture. The L1b layer is the newest addition —
        an entity-keyed hot path cache that resolves repeat entity-identical requests without any
        vector search or LLM call.
      </p>

      {/* Overview Diagram */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <BarChart3 size={16} className="text-memzent-glow" />
          Cache Layer Overview
        </h2>
        <CodeBlock code={cacheStages} language="text" />
      </section>

      {/* Layer Details */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Database size={16} className="text-memzent-glow" />
          Layer Details
        </h2>
        <div className="space-y-4">
          <div className="p-4 rounded-xl border border-memzent-glow/20 bg-memzent-glow/5">
            <div className="flex items-center gap-2 mb-2">
              <span className="text-xs font-black text-memzent-glow bg-memzent-glow/10 px-2 py-0.5 rounded">L1</span>
              <h4 className="text-sm font-black text-white">Literal Hash</h4>
            </div>
            <p className="text-xs text-white/50">
              SHA-256 of the raw prompt. Fastest possible lookup — sub-millisecond via Valkey.
              Only matches character-for-character identical prompts.
            </p>
          </div>

          <div className="p-4 rounded-xl border border-blue-500/20 bg-blue-500/5">
            <div className="flex items-center gap-2 mb-2">
              <span className="text-xs font-black text-blue-400 bg-blue-500/10 px-2 py-0.5 rounded">L1.5</span>
              <h4 className="text-sm font-black text-white">Canonical Hash</h4>
            </div>
            <p className="text-xs text-white/50">
              Normalized version — lowercased, whitespace-collapsed, punctuation-stripped.
              Catches formatting differences without any vector math.
            </p>
          </div>

          <div className="p-4 rounded-xl border border-green-500/20 bg-green-500/5">
            <div className="flex items-center gap-2 mb-2">
              <span className="text-xs font-black text-green-400 bg-green-500/10 px-2 py-0.5 rounded">L1b</span>
              <h4 className="text-sm font-black text-white">Entity-Keyed Hot Path</h4>
              <span className="text-[9px] font-black text-green-400 bg-green-500/10 px-1.5 py-0.5 rounded ml-auto">NEW</span>
            </div>
            <p className="text-xs text-white/50 mb-3">
              Built from extracted entities. If a prompt&apos;s entities match a cached entry exactly,
              the response is served directly from Valkey without any vector search.
              This layer handles the most common repeat pattern: same question with same entity values.
            </p>
            <CodeBlock code={l1bKeyFormat} language="javascript" />
          </div>

          <div className="p-4 rounded-xl border border-purple-500/20 bg-purple-500/5">
            <div className="flex items-center gap-2 mb-2">
              <span className="text-xs font-black text-purple-400 bg-purple-500/10 px-2 py-0.5 rounded">L2</span>
              <h4 className="text-sm font-black text-white">Semantic Vector</h4>
            </div>
            <p className="text-xs text-white/50">
              Vector similarity via the Rust Router + Qdrant. Embeds using <code>all-MiniLM-L6-v2</code> (384-dim)
              and matches semantically similar prompts with entity post-filter guard.
              Threshold: 0.95 similarity score. Slowest layer but catches paraphrased questions.
            </p>
          </div>
        </div>
      </section>

      {/* L1b Feature Flag */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Key size={16} className="text-memzent-glow" />
          Configuration
        </h2>
        <div className="space-y-3">
          <div className="p-4 rounded-xl border border-white/10 bg-white/5">
            <h4 className="text-sm font-black text-white mb-1">Feature Flag</h4>
            <p className="text-xs text-white/50 mb-2">
              L1b is controlled by the <code>MEMZENT_L1B_ENABLED</code> environment variable (default: <code>true</code>).
            </p>
          </div>
          <div className="p-4 rounded-xl border border-white/10 bg-white/5">
            <h4 className="text-sm font-black text-white mb-1">Write-Through on Skip Cache</h4>
            <p className="text-xs text-white/50">
              When <code>X-Skip-Cache: true</code> is sent, cache <strong>reads</strong> are skipped but the LLM response
              is still <strong>written</strong> to all cache layers. This ensures the cache stays warm for future requests.
            </p>
          </div>
        </div>
      </section>

      {/* Performance */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Zap size={16} className="text-memzent-glow" />
          Performance Characteristics
        </h2>
        <div className="overflow-x-auto">
          <table className="w-full text-xs">
            <thead>
              <tr className="border-b border-white/10">
                <th className="text-left py-2 px-3 font-black text-white/70">Layer</th>
                <th className="text-left py-2 px-3 font-black text-white/70">Latency</th>
                <th className="text-left py-2 px-3 font-black text-white/70">Backend</th>
                <th className="text-left py-2 px-3 font-black text-white/70">Accuracy</th>
              </tr>
            </thead>
            <tbody className="text-white/50">
              <tr className="border-b border-white/5"><td className="py-2 px-3">L1</td><td className="py-2 px-3">&lt;1ms</td><td className="py-2 px-3">Valkey</td><td className="py-2 px-3">Exact</td></tr>
              <tr className="border-b border-white/5"><td className="py-2 px-3">L1.5</td><td className="py-2 px-3">&lt;1ms</td><td className="py-2 px-3">Valkey</td><td className="py-2 px-3">Normalized</td></tr>
              <tr className="border-b border-white/5"><td className="py-2 px-3">L1b</td><td className="py-2 px-3">1–2ms</td><td className="py-2 px-3">Valkey</td><td className="py-2 px-3">Entity-exact</td></tr>
              <tr><td className="py-2 px-3">L2</td><td className="py-2 px-3">15–50ms</td><td className="py-2 px-3">Qdrant</td><td className="py-2 px-3">Semantic (≥0.95)</td></tr>
            </tbody>
          </table>
        </div>
      </section>

      <DocsPager currentPath="/docs/cache-layers" />
    </div>
  );
}
