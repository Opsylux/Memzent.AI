import { ArrowRight, Zap, ShieldCheck, Search, Layers, RefreshCw, Sparkles } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";

export default function ArchitecturePage() {
  return (
    <div className="space-y-14">
      <header className="space-y-4">
        <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-memzent-glow/5 border border-memzent-glow/20 w-fit">
          <span className="text-[10px] font-black text-memzent-glow uppercase tracking-tighter italic">How_It_Works</span>
        </div>
        <h1 className="text-4xl font-black tracking-tighter uppercase sm:text-5xl">How Memzent Works</h1>
        <p className="text-lg text-white/60 leading-relaxed font-medium max-w-2xl">
          Memzent sits between your application and your AI models. It understands what your users are asking — not just the words they use — and intelligently routes every request to get the fastest, most accurate answer possible.
        </p>
      </header>

      {/* The Big Picture */}
      <section className="space-y-6">
        <h2 className="text-2xl font-black tracking-tighter uppercase">The Big Picture</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Think of Memzent as an intelligent concierge. Your app sends a question, and Memzent handles everything: checking if the answer already exists, finding the right tools to gather context, and then synthesizing a response — all before returning a result to you.
        </p>

        {/* Visual flow diagram */}
        <div className="flex flex-col sm:flex-row items-center gap-2 py-6 overflow-x-auto">
          {[
            { label: "Client" },
            { label: "Go Gateway (:8080)" },
            { label: "Rust Router (gRPC)" },
            { label: "Qdrant (Vectors)" },
          ].map((node, i, arr) => (
            <div key={node.label} className="flex items-center gap-2 shrink-0">
              <div className={`px-4 py-2 rounded-xl text-xs font-black uppercase tracking-tight border ${node.label.includes("Gateway")
                ? "bg-memzent-glow/10 border-memzent-glow/30 text-memzent-glow shadow-[0_0_20px_rgba(0,243,255,0.08)]"
                : "bg-white/[0.02] border-white/10 text-white/50"
                }`}>
                {node.label}
              </div>
              {i < arr.length - 1 && (
                <ArrowRight size={14} className="text-white/20 shrink-0" />
              )}
            </div>
          ))}
        </div>
        <div className="flex flex-col sm:flex-row items-center gap-2 py-2 overflow-x-auto">
          {[
            { label: "Valkey (Cache)" },
            { label: "Postgres (RBAC/Billing)" },
            { label: "LLM Providers" },
            { label: "MCP Tools" },
          ].map((node) => (
            <div key={node.label} className="px-3 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-tight border bg-white/[0.01] border-white/5 text-white/30">
              {node.label}
            </div>
          ))}
        </div>
      </section>

      {/* What happens with every request */}
      <section className="space-y-6">
        <h2 className="text-2xl font-black tracking-tighter uppercase">What Happens With Every Request</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Every time your application sends a prompt, Memzent follows a deliberate sequence to maximize speed, accuracy, and cost efficiency.
        </p>

        <div className="space-y-3 pt-2">
          {[
            {
              icon: <ShieldCheck size={16} />,
              color: "text-white/50",
              step: "01",
              title: "Rate Limiting",
              desc: "Distributed rate limiting via Valkey checks org tier (free: 10/min, pro: 100/min, business: 1000/min) and per-user proportional limits based on role."
            },
            {
              icon: <ShieldCheck size={16} />,
              color: "text-memzent-accent",
              step: "02",
              title: "Auth & Permission Check",
              desc: "API key or JWT is verified. RBAC checks confirm the user has the required scope (chat:execute, tools:read, etc). Viewer role is blocked from execution."
            },
            {
              icon: <Zap size={16} />,
              color: "text-yellow-400",
              step: "03",
              title: "Billing Pre-check",
              desc: "Token balance is verified and spend limits (daily/monthly dollar + token caps) are checked. Requests are blocked with 402 if limits are exceeded."
            },
            {
              icon: <Search size={16} />,
              color: "text-memzent-glow",
              step: "04",
              title: "Triple-Layer Cache Check",
              desc: "Layer 1: Exact hash match in Valkey. Layer 1.5: Canonical (normalized) hash match. Layer 2: Semantic vector similarity via Rust Router + Qdrant (threshold 0.95). With durable Postgres fallback if Valkey is down."
            },
            {
              icon: <Layers size={16} />,
              color: "text-memzent-purple",
              step: "05",
              title: "Session Memory & Recall",
              desc: "Conversation history is loaded from the session. Long-term semantic memories (user preferences, facts) are recalled from Qdrant if relevance > 0.65."
            },
            {
              icon: <Layers size={16} />,
              color: "text-blue-400",
              step: "06",
              title: "Semantic Routing (Rust gRPC)",
              desc: "Prompt is embedded via all-MiniLM-L6-v2 (384-dim). Qdrant is searched for matching tools and similar prompts. Tools above the relevance threshold are selected."
            },
            {
              icon: <RefreshCw size={16} />,
              color: "text-memzent-accent",
              step: "07",
              title: "Tool Execution",
              desc: "Matched MCP tools and connectors are called to gather live context. Supports sequential chaining when prompts require multi-step workflows."
            },
            {
              icon: <Sparkles size={16} />,
              color: "text-memzent-glow",
              step: "08",
              title: "LLM Synthesis",
              desc: "The enriched prompt (with tool results, memory, session history) is sent to the selected provider — Ollama, OpenAI, Anthropic, or Gemini."
            },
            {
              icon: <Zap size={16} />,
              color: "text-memzent-accent",
              step: "09",
              title: "Cache Set & Billing",
              desc: "Response is cached in Valkey + Postgres. Billing is deducted (90% discount for cache hits). Semantic memory facts are auto-extracted in background."
            },
            {
              icon: <RefreshCw size={16} />,
              color: "text-white/50",
              step: "10",
              title: "Webhook Events",
              desc: "Subscribed webhooks receive notifications (cache_hit, tool_execution, rate_limit, etc.) with signed payloads for audit and monitoring."
            },
          ].map((item) => (
            <div
              key={item.step}
              className="flex gap-5 p-5 rounded-2xl hover:bg-white/[0.02] transition-colors border border-transparent hover:border-white/5 group"
            >
              <div className="flex flex-col items-center gap-2 shrink-0">
                <div className={`w-8 h-8 rounded-lg flex items-center justify-center bg-white/5 ${item.color} transition-colors`}>
                  {item.icon}
                </div>
                <div className="flex-1 w-px bg-white/5" />
              </div>
              <div className="space-y-1.5 pb-4">
                <div className="flex items-center gap-3">
                  <span className="text-[10px] font-black text-white/20 italic">{item.step}</span>
                  <h3 className="text-sm font-black uppercase tracking-tight text-white/80">{item.title}</h3>
                </div>
                <p className="text-xs text-white/40 leading-relaxed font-bold">{item.desc}</p>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* Key Outcomes */}
      <section className="space-y-6">
        <h2 className="text-2xl font-black tracking-tighter uppercase">What This Means for Your Team</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="p-6 rounded-2xl bg-white/[0.02] border border-white/5 space-y-3">
            <div className="text-2xl font-black text-memzent-glow">Sub-1ms</div>
            <div className="text-xs font-black uppercase tracking-tight">For cached responses</div>
            <p className="text-[11px] text-white/40 leading-relaxed font-bold">
              Questions your users repeat get answered from memory — instantly — at essentially zero cost.
            </p>
          </div>
          <div className="p-6 rounded-2xl bg-white/[0.02] border border-white/5 space-y-3">
            <div className="text-2xl font-black text-memzent-purple">Any Model</div>
            <div className="text-xs font-black uppercase tracking-tight">One unified API</div>
            <p className="text-[11px] text-white/40 leading-relaxed font-bold">
              Switch between AI providers without changing your application code. Memzent handles all provider details.
            </p>
          </div>
          <div className="p-6 rounded-2xl bg-white/[0.02] border border-white/5 space-y-3">
            <div className="text-2xl font-black text-memzent-accent">Live Tools</div>
            <div className="text-xs font-black uppercase tracking-tight">Real-time context</div>
            <p className="text-[11px] text-white/40 leading-relaxed font-bold">
              Your AI answers using fresh, live data from your own systems — no hallucinations from stale training.
            </p>
          </div>
        </div>
      </section>

      <DocsPager />
    </div>
  );
}
