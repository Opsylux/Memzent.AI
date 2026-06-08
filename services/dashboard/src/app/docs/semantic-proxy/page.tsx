import { DocsPager } from "@/components/docs/docs-pager";
import { CodeBlock } from "@/components/docs/code-block";
import { DOCS_CONFIG } from "@/config/docs-config";
import { Zap, AlertCircle } from "lucide-react";

export default function SemanticProxyPage() {
  const skipCacheExample = `curl -X POST https://${DOCS_CONFIG.domain}/v1/chat \\
  -H "X-API-Key: your_key" \\
  -H "X-Skip-Cache: true" \\
  -d '{"messages": [{"role": "user", "content": "Get real-time server status"}]}'`;

  return (
    <div className="space-y-12">
      <header className="space-y-4">
        <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-memzent-glow/5 border border-memzent-glow/20 w-fit">
          <span className="text-[10px] font-black text-memzent-glow uppercase tracking-tighter italic">Core_Concept</span>
        </div>
        <h1 className="text-4xl font-black tracking-tighter uppercase sm:text-5xl">Semantic Caching</h1>
        <p className="text-lg text-white/60 leading-relaxed font-medium">
          Memzent never charges you twice for the same answer. Its semantic memory catches repeat questions — even when phrased differently — and returns instant responses at no extra cost.
        </p>
      </header>

      {/* Three layers */}
      <section className="space-y-5">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Three Layers of Memory</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Before every request reaches an AI model, Memzent checks three tiers of memory in sequence. The faster tiers cost less — and later tiers are only used when necessary.
        </p>

        <div className="grid grid-cols-1 gap-4">
          {[
            {
              label: "L1",
              color: "text-memzent-glow",
              bg: "bg-memzent-glow/10",
              title: "Literal Match (Exact Hash)",
              desc: "SHA-256 hash of the raw prompt text looked up in Valkey. Sub-millisecond. Only matches character-for-character identical prompts."
            },
            {
              label: "L1.5",
              color: "text-memzent-purple",
              bg: "bg-memzent-purple/10",
              title: "Canonical Match (Normalized Hash)",
              desc: "Prompt is normalized (lowercased, whitespace collapsed, punctuation standardized) then hashed. Catches formatting variants: \"What is AI?\" matches \"what is ai ?\"."
            },
            {
              label: "L2",
              color: "text-memzent-accent",
              bg: "bg-memzent-accent/10",
              title: "Semantic Match (Vector Similarity)",
              desc: "Prompt is embedded into a 384-dim vector via all-MiniLM-L6-v2 (Rust Router) and compared against stored prompts in Qdrant. Threshold: 0.95 similarity. Includes a numeric guard that rejects matches when parameter values differ."
            }
          ].map((item) => (
            <div key={item.label} className="p-6 rounded-2xl bg-white/[0.02] border border-white/5 space-y-3">
              <div className="flex items-center gap-3">
                <div className={`w-8 h-8 rounded-lg ${item.bg} flex items-center justify-center ${item.color} text-xs font-black`}>
                  {item.label}
                </div>
                <h3 className="text-sm font-black uppercase tracking-tight">{item.title}</h3>
              </div>
              <p className="text-xs text-white/40 leading-relaxed font-bold">{item.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Skip cache */}
      <section className="space-y-6 pt-4">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Forcing a Fresh Response</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Sometimes you need a fresh answer — for example, when querying live system status or real-time data. Add the <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">X-Skip-Cache: true</code> header to bypass all memory layers and force a new generation.
        </p>
        <CodeBlock code={skipCacheExample} language="bash" filename="terminal" />
        <div className="p-4 rounded-xl bg-white/[0.02] border border-white/5 flex items-start gap-3">
          <AlertCircle size={14} className="text-white/20 mt-0.5 shrink-0" />
          <p className="text-[11px] text-white/30 font-bold leading-relaxed">
            Skipping the cache means a full AI generation will run. Use this only when you know the data changes frequently enough that a cached response would be incorrect.
          </p>
        </div>
      </section>

      {/* Reading the response */}
      <section className="space-y-6 pt-4">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Reading the Response Header</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Every response from Memzent includes an <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">X-Cache</code> header so you can see exactly which memory layer was used.
        </p>
        <div className="space-y-3">
          {[
            { value: "X-Cache: HIT", desc: "The response was retrieved from memory. No AI model was called.", color: "text-memzent-accent" },
            { value: "X-Cache: MISS", desc: "No match was found. Memzent called the AI model and the result is now saved for future requests.", color: "text-white/50" },
          ].map((row) => (
            <div key={row.value} className="flex gap-4 p-4 rounded-xl bg-white/[0.02] border border-white/5 items-start">
              <code className={`text-xs font-mono font-bold shrink-0 ${row.color}`}>{row.value}</code>
              <p className="text-[11px] text-white/40 font-bold leading-relaxed">{row.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Performance tip */}
      <div className="p-6 rounded-2xl bg-gradient-to-br from-memzent-glow/10 to-transparent border border-memzent-glow/20">
        <div className="flex items-center gap-3 mb-3">
          <Zap size={18} className="text-memzent-glow" />
          <span className="text-xs font-black uppercase tracking-tight">Tip: Write Better Descriptions</span>
        </div>
        <p className="text-xs text-white/50 font-bold leading-relaxed">
          The better your tool descriptions match the language your users actually use, the higher your cache hit rate. Think about what a user would ask — not what a developer would name it.
        </p>
      </div>

      <DocsPager />
    </div>
  );
}
