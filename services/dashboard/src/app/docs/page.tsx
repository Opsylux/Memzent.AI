import { Zap, Shield, Search, Layers, ArrowRight } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";
import Link from "next/link";

export default function DocsIntroduction() {
  return (
    <div className="space-y-14">
      <header className="space-y-4">
        <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-memzent-glow/5 border border-memzent-glow/20 w-fit">
          <span className="text-[10px] font-black text-memzent-glow uppercase tracking-tighter">Documentation</span>
        </div>
        <h1 className="text-4xl font-black tracking-tighter uppercase sm:text-5xl">Introduction to Memzent</h1>
        <p className="text-lg text-white/60 leading-relaxed font-medium max-w-2xl">
          Memzent is an AI infrastructure layer that sits between your application and your AI models. It makes every request faster, smarter, and cheaper — without changing your existing code.
        </p>
      </header>

      {/* Feature cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-5">
        {[
          {
            icon: <Search size={20} />,
            color: "text-memzent-glow",
            bg: "bg-memzent-glow/10",
            title: "Semantic Memory",
            desc: "Memzent remembers the meaning of past answers, not just exact words. Similar questions get instant responses — at zero cost."
          },
          {
            icon: <Layers size={20} />,
            color: "text-memzent-purple",
            bg: "bg-memzent-purple/10",
            title: "Smart Routing",
            desc: "Every prompt is matched to the most relevant tool or data source first, so your AI always responds with accurate, current information."
          },
          {
            icon: <Shield size={20} />,
            color: "text-memzent-accent",
            bg: "bg-memzent-accent/10",
            title: "Zero-Trust Security",
            desc: "Every request is verified against your organization's live permissions — before any AI or tool is invoked."
          },
          {
            icon: <Zap size={20} />,
            color: "text-memzent-glow",
            bg: "bg-memzent-glow/10",
            title: "Any AI Model",
            desc: "Switch between OpenAI, Anthropic, Gemini, or your own self-hosted model with a single header. One API for all of them."
          }
        ].map((item) => (
          <div key={item.title} className="p-6 rounded-2xl bg-white/[0.02] border border-white/5 space-y-4 hover:border-white/10 transition-colors">
            <div className={`w-10 h-10 rounded-xl ${item.bg} flex items-center justify-center ${item.color}`}>
              {item.icon}
            </div>
            <h3 className="text-sm font-black uppercase tracking-tight">{item.title}</h3>
            <p className="text-xs text-white/40 leading-relaxed font-bold">{item.desc}</p>
          </div>
        ))}
      </div>

      {/* Why section */}
      <section className="space-y-6">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Why use Memzent?</h2>
        <div className="space-y-4 text-sm text-white/60 leading-relaxed font-medium">
          <p>
            Most teams connect their app directly to an AI API and manage caching, routing, and security themselves. As usage grows, that becomes expensive and fragile.
          </p>
          <ul className="list-disc pl-6 space-y-3 marker:text-memzent-glow">
            <li><strong className="text-white">Cut AI costs by up to 80%</strong> — repeat questions are answered from memory instantly, without calling the model.</li>
            <li><strong className="text-white">Eliminate hallucinations</strong> — Memzent enriches every prompt with live data from your tools before the model ever sees it.</li>
            <li><strong className="text-white">One secure entry point</strong> — manage all AI access, rate limits, and permissions from a single place.</li>
          </ul>
        </div>
      </section>

      {/* How it works overview */}
      <section className="space-y-6">
        <h2 className="text-2xl font-black tracking-tighter uppercase">How the Flow Works</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Every request through the gateway goes through a precise execution pipeline:
        </p>
        <div className="space-y-3">
          {[
            { label: "Rate Limit & Auth", desc: "Distributed rate limiting per org/tier, then JWT or API key verification with RBAC scope checks." },
            { label: "Billing Pre-check", desc: "Verify token balance and spend limits (daily/monthly caps) before processing." },
            { label: "Cache Check", desc: "Four-layer lookup: L1 literal hash → L1.5 canonical → L1b entity-keyed hot path → L2 semantic similarity (vector match via Qdrant with entity post-filter guard)." },
            { label: "Session & Memory", desc: "Load conversation history, recall long-term semantic memories relevant to this prompt." },
            { label: "Semantic Routing", desc: "gRPC call to Rust Router for tool matching, prompt compression, and vector search." },
            { label: "Tool Execution", desc: "Fire matched MCP/connector tools to gather live context for the LLM." },
            { label: "LLM Synthesis", desc: "Send enriched prompt to the selected provider (Ollama, OpenAI, Anthropic, or Gemini)." },
            { label: "Cache & Respond", desc: "Store the response in cache, deduct billing, emit webhook events, and return to client." },
          ].map((item, i) => (
            <div key={item.label} className="flex gap-5 items-start p-4 rounded-xl hover:bg-white/[0.02] transition-colors">
              <span className="text-xl font-black text-white/10 italic shrink-0 mt-0.5">0{i + 1}</span>
              <div className="space-y-1">
                <div className="text-xs font-black uppercase tracking-tight text-white/80">{item.label}</div>
                <div className="text-[11px] text-white/40 font-bold leading-relaxed">{item.desc}</div>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* Quick API */}
      <section className="space-y-6">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Quick API Overview</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-xs border border-white/5 rounded-lg overflow-hidden">
            <thead>
              <tr className="bg-white/[0.03] border-b border-white/5">
                <th className="text-left px-4 py-2 font-black text-white/60">Endpoint</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Purpose</th>
              </tr>
            </thead>
            <tbody className="text-white/40">
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-memzent-glow/70">POST /v1/chat</td><td className="px-4 py-2">Send prompts, get AI responses</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-memzent-glow/70">GET /v1/providers</td><td className="px-4 py-2">List available LLM providers</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-memzent-glow/70">POST /v1/tools/register</td><td className="px-4 py-2">Register tools for semantic routing</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-memzent-glow/70">POST /v1/sessions</td><td className="px-4 py-2">Create conversation sessions</td></tr>
              <tr className="border-b border-white/5"><td className="px-4 py-2 font-mono text-memzent-glow/70">GET /v1/billing/budget</td><td className="px-4 py-2">Check balance and spend analytics</td></tr>
              <tr><td className="px-4 py-2 font-mono text-memzent-glow/70">POST /v1/webhooks</td><td className="px-4 py-2">Subscribe to real-time events</td></tr>
            </tbody>
          </table>
        </div>
        <p className="text-xs text-white/40">
          See the full <Link href="/docs/api-reference" className="text-memzent-glow hover:underline">API Reference</Link> for all endpoints.
        </p>
      </section>

      {/* CTA */}
      <div className="flex items-center gap-4 pt-4 border-t border-white/5">
        <Link href="/docs/quickstart" className="flex items-center gap-2 px-5 py-3 rounded-xl bg-memzent-glow text-black text-xs font-black uppercase tracking-widest hover:scale-105 transition-all">
          Get Started <ArrowRight size={13} />
        </Link>
        <Link href="/docs/architecture" className="text-xs text-white/40 font-black uppercase tracking-widest hover:text-white transition-colors">
          Read the Overview →
        </Link>
      </div>

      <DocsPager />
    </div>
  );
}
