import { Zap, Shield, Search, Layers, ArrowRight } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";
import Link from "next/link";

export default function DocsIntroduction() {
  return (
    <div className="space-y-14">
      <header className="space-y-4">
        <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-aura-glow/5 border border-aura-glow/20 w-fit">
          <span className="text-[10px] font-black text-aura-glow uppercase tracking-tighter">Documentation</span>
        </div>
        <h1 className="text-4xl font-black tracking-tighter uppercase sm:text-5xl">Introduction to Aura</h1>
        <p className="text-lg text-white/60 leading-relaxed font-medium max-w-2xl">
          Aura is an AI infrastructure layer that sits between your application and your AI models. It makes every request faster, smarter, and cheaper — without changing your existing code.
        </p>
      </header>

      {/* Feature cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-5">
        {[
          {
            icon: <Search size={20} />,
            color: "text-aura-glow",
            bg: "bg-aura-glow/10",
            title: "Semantic Memory",
            desc: "Aura remembers the meaning of past answers, not just exact words. Similar questions get instant responses — at zero cost."
          },
          {
            icon: <Layers size={20} />,
            color: "text-aura-purple",
            bg: "bg-aura-purple/10",
            title: "Smart Routing",
            desc: "Every prompt is matched to the most relevant tool or data source first, so your AI always responds with accurate, current information."
          },
          {
            icon: <Shield size={20} />,
            color: "text-aura-accent",
            bg: "bg-aura-accent/10",
            title: "Zero-Trust Security",
            desc: "Every request is verified against your organization's live permissions — before any AI or tool is invoked."
          },
          {
            icon: <Zap size={20} />,
            color: "text-aura-glow",
            bg: "bg-aura-glow/10",
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
        <h2 className="text-2xl font-black tracking-tighter uppercase">Why use Aura?</h2>
        <div className="space-y-4 text-sm text-white/60 leading-relaxed font-medium">
          <p>
            Most teams connect their app directly to an AI API and manage caching, routing, and security themselves. As usage grows, that becomes expensive and fragile.
          </p>
          <ul className="list-disc pl-6 space-y-3 marker:text-aura-glow">
            <li><strong className="text-white">Cut AI costs by up to 80%</strong> — repeat questions are answered from memory instantly, without calling the model.</li>
            <li><strong className="text-white">Eliminate hallucinations</strong> — Aura enriches every prompt with live data from your tools before the model ever sees it.</li>
            <li><strong className="text-white">One secure entry point</strong> — manage all AI access, rate limits, and permissions from a single place.</li>
          </ul>
        </div>
      </section>

      {/* How it works overview */}
      <section className="space-y-6">
        <h2 className="text-2xl font-black tracking-tighter uppercase">How the Flow Works</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          At a high level, every request goes through three stages before a response is returned.
        </p>
        <div className="space-y-3">
          {[
            { label: "Check Memory", desc: "Has something similar been asked before? If yes, return the cached answer instantly." },
            { label: "Gather Context", desc: "Find and call the tools that can provide fresh, relevant information for this specific question." },
            { label: "Generate & Return", desc: "Send the enriched prompt to the AI model, get the response, and save it for future requests." },
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

      {/* CTA */}
      <div className="flex items-center gap-4 pt-4 border-t border-white/5">
        <Link href="/docs/quickstart" className="flex items-center gap-2 px-5 py-3 rounded-xl bg-aura-glow text-black text-xs font-black uppercase tracking-widest hover:scale-105 transition-all">
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
