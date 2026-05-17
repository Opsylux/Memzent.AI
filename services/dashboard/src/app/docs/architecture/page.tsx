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
            { label: "Your App" },
            { label: "Memzent Gateway" },
            { label: "Your Tools" },
            { label: "AI Model" },
            { label: "Response" },
          ].map((node, i, arr) => (
            <div key={node.label} className="flex items-center gap-2 shrink-0">
              <div className={`px-4 py-2 rounded-xl text-xs font-black uppercase tracking-tight border ${node.label === "Memzent Gateway"
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
              title: "Identity & Access Check",
              desc: "Memzent verifies who is making the request and confirms they have permission to access the tools and models they need. Unauthorized requests are stopped before any work begins."
            },
            {
              icon: <Search size={16} />,
              color: "text-memzent-glow",
              step: "02",
              title: "Did We Already Answer This?",
              desc: "Memzent checks its semantic memory. If a similar question was already answered — even if worded differently — the cached response is returned instantly. No model call, no cost, zero wait."
            },
            {
              icon: <Layers size={16} />,
              color: "text-memzent-purple",
              step: "03",
              title: "Finding the Right Tools",
              desc: "Memzent understands the intent behind the question and identifies which of your connected tools — databases, APIs, knowledge bases — can provide the most relevant context."
            },
            {
              icon: <RefreshCw size={16} />,
              color: "text-memzent-accent",
              step: "04",
              title: "Gathering Context",
              desc: "The matched tools are called to retrieve live data: customer records, documents, metrics, or anything else your tools expose. This context is assembled alongside the original question."
            },
            {
              icon: <Sparkles size={16} />,
              color: "text-memzent-glow",
              step: "05",
              title: "AI Generates the Answer",
              desc: "The enriched prompt is sent to the AI model of your choice. Because Memzent already gathered the relevant context, the model produces a more accurate, grounded response — using fewer tokens."
            },
            {
              icon: <Zap size={16} />,
              color: "text-memzent-accent",
              step: "06",
              title: "Response Saved for Next Time",
              desc: "The answer is returned to your app and saved in Memzent's semantic memory. The next time someone asks something similar, they get an instant response — from any provider, any region."
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
