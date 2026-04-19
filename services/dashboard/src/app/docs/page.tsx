import { Zap, Shield, Search, Layers } from "lucide-react";

export default function DocsIntroduction() {
  return (
    <div className="space-y-12">
      <header className="space-y-4">
        <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-aura-glow/5 border border-aura-glow/20 w-fit">
          <span className="text-[10px] font-black text-aura-glow uppercase tracking-tighter">Documentation</span>
        </div>
        <h1 className="text-4xl font-black tracking-tighter uppercase sm:text-5xl">Introduction to Aura</h1>
        <p className="text-lg text-white/60 leading-relaxed font-medium">
          Aura is an enterprise-grade AI infrastructure mesh that acts as an **Intelligent Semantic Proxy** between your clients and your Large Language Model (LLM) endpoints.
        </p>
      </header>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="p-6 rounded-2xl bg-white/[0.02] border border-white/5 space-y-4">
          <div className="w-10 h-10 rounded-lg bg-aura-glow/10 flex items-center justify-center text-aura-glow">
            <Search size={20} />
          </div>
          <h3 className="text-sm font-black uppercase tracking-tight">Semantic Caching</h3>
          <p className="text-xs text-white/40 leading-relaxed font-bold">
            Aura caches responses semantically, not just by exact string matches. This minimizes LLM latency and maximizes token ROI by serving cached intent.
          </p>
        </div>
        <div className="p-6 rounded-2xl bg-white/[0.02] border border-white/5 space-y-4">
          <div className="w-10 h-10 rounded-lg bg-aura-purple/10 flex items-center justify-center text-aura-purple">
            <Layers size={20} />
          </div>
          <h3 className="text-sm font-black uppercase tracking-tight">Intelligent Routing</h3>
          <p className="text-xs text-white/40 leading-relaxed font-bold">
            Dynamically route requests to specialized tools (MCP) or model endpoints based on the high-dimensional intent of the user prompt.
          </p>
        </div>
      </div>

      <section className="space-y-6 pt-6">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Why use a Semantic Proxy?</h2>
        <div className="space-y-4 text-sm text-white/60 leading-relaxed font-medium">
          <p>
            Standard load balancers and proxies operate at the network layer. Aura operates at the **intent layer**. By understanding what the user is actually asking for, Aura can:
          </p>
          <ul className="list-disc pl-6 space-y-3 marker:text-aura-glow">
            <li><strong className="text-white">Reduce Latency</strong>: Serve repeat intents in sub-1ms directly from the edge cache.</li>
            <li><strong className="text-white">Lower Costs</strong>: Avoid redundant calls to expensive LLMs like GPT-4 or Claude 3 by caching synthesized outputs.</li>
            <li><strong className="text-white">Hardened Security</strong>: Enforce unified RBAC and Row-Level Security (RLS) across all your tool-calling capabilities.</li>
          </ul>
        </div>
      </section>

      <section className="space-y-6 pt-6">
        <h2 className="text-2xl font-black tracking-tighter uppercase">System Architecture</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Aura is built on a high-performance distributed architecture consisting of three core services:
        </p>
        <div className="space-y-4">
          <div className="flex gap-4 p-4 rounded-xl bg-white/[0.01] border border-white/5">
            <div className="text-xs font-black text-aura-glow border-r border-white/10 pr-4 whitespace-nowrap">GO GATEWAY</div>
            <div className="text-xs text-white/40 font-bold uppercase tracking-wide">Orchestrator, Auth, RBAC & Cache Management</div>
          </div>
          <div className="flex gap-4 p-4 rounded-xl bg-white/[0.01] border border-white/5">
            <div className="text-xs font-black text-aura-purple border-r border-white/10 pr-4 whitespace-nowrap">RUST ROUTER</div>
            <div className="text-xs text-white/40 font-bold uppercase tracking-wide">Vector Math, Semantic Similarity & Tool Matching</div>
          </div>
          <div className="flex gap-4 p-4 rounded-xl bg-white/[0.01] border border-white/5">
            <div className="text-xs font-black text-aura-accent border-r border-white/10 pr-4 whitespace-nowrap">MCP SERVER</div>
            <div className="text-xs text-white/40 font-bold uppercase tracking-wide">Execution of connected tools and data-sources</div>
          </div>
        </div>
      </section>
    </div>
  );
}
