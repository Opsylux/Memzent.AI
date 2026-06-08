import { useState } from 'react';
import { Routes, Route, Link, useLocation } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Shield,
  Activity,
  ChevronRight,
  Terminal,
  Database,
  Lock,
  Monitor,
  Zap,
  Brain,
  ArrowRight,
  Check,
  DollarSign,
  Cpu,
  Menu,
  X,
  Code as Github,
  Scan,
  Layers,
  BarChart3
} from 'lucide-react';



import FeatureCard from './components/FeatureCard';
import BlogListPage from './pages/BlogList';
import BlogPostPage from './pages/BlogPost';

const appUrl = import.meta.env.VITE_APP_URL || "http://localhost:3000"

const Navbar = () => {
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <nav className="fixed top-0 left-0 right-0 z-50 glass px-6 py-4 flex justify-between items-center m-4 rounded-2xl">
      <Link to="/" className="flex items-center gap-2">
        <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-memzent-glow to-memzent-purple flex items-center justify-center shadow-[0_0_15px_rgba(0,243,255,0.3)]">
          <span className="text-black font-black text-sm italic select-none">M</span>
        </div>
        <span className="text-2xl font-black tracking-tighter">MEMZENT</span>
      </Link>
      <div className="hidden md:flex gap-8 text-sm font-medium opacity-80">
        <a href={appUrl + "/docs"} className="hover:text-memzent-glow transition-colors">Docs</a>
        <a href="#payg" className="hover:text-memzent-glow transition-colors">Pricing</a>
        <a href="#why" className="hover:text-memzent-glow transition-colors">Why Memzent</a>
        <a href="#security" className="hover:text-memzent-glow transition-colors">Security</a>
        <a href="#observability" className="hover:text-memzent-glow transition-colors">Observability</a>
        <Link to="/blog" className="hover:text-memzent-glow transition-colors">Blog</Link>
      </div>
      <div className="hidden md:flex gap-4">
        <a href="https://github.com/Opsylux/Memzent.AI" target="_blank" rel="noopener" className="text-sm font-bold opacity-75 hover:opacity-100 px-3 py-2 transition-all cursor-pointer flex items-center gap-2">
          <Github size={16} /> GitHub
        </a>
        <a href={appUrl + "/login"} className="text-sm font-bold opacity-75 hover:opacity-100 px-4 py-2 transition-all cursor-pointer">Login</a>
        <a href={appUrl + "/login"} className="bg-memzent-glow text-black text-sm font-black px-6 py-2 rounded-xl hover:shadow-[0_0_20px_rgba(0,243,255,0.4)] hover:scale-105 transition-all cursor-pointer">Get Started Free</a>
      </div>

      {/* Mobile hamburger */}
      <button onClick={() => setMobileOpen(!mobileOpen)} className="md:hidden text-white p-2">
        {mobileOpen ? <X size={24} /> : <Menu size={24} />}
      </button>

      {/* Mobile menu */}
      <AnimatePresence>
        {mobileOpen && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            className="absolute top-full left-0 right-0 mt-2 mx-4 glass rounded-2xl p-6 flex flex-col gap-4 md:hidden"
          >
            <a href={appUrl + "/docs"} onClick={() => setMobileOpen(false)} className="text-sm font-bold opacity-80 hover:text-memzent-glow py-2">Docs</a>
            <a href="#payg" onClick={() => setMobileOpen(false)} className="text-sm font-bold opacity-80 hover:text-memzent-glow py-2">Pricing</a>
            <a href="#why" onClick={() => setMobileOpen(false)} className="text-sm font-bold opacity-80 hover:text-memzent-glow py-2">Why Memzent</a>
            <a href="#security" onClick={() => setMobileOpen(false)} className="text-sm font-bold opacity-80 hover:text-memzent-glow py-2">Security</a>
            <a href="#observability" onClick={() => setMobileOpen(false)} className="text-sm font-bold opacity-80 hover:text-memzent-glow py-2">Observability</a>
            <Link to="/blog" onClick={() => setMobileOpen(false)} className="text-sm font-bold opacity-80 hover:text-memzent-glow py-2">Blog</Link>
            <hr className="border-white/10" />
            <a href="https://github.com/Opsylux/Memzent.AI" target="_blank" rel="noopener" className="text-sm font-bold opacity-80 flex items-center gap-2"><Github size={16} /> GitHub</a>
            <a href={appUrl + "/login"} className="bg-memzent-glow text-black text-sm font-black px-6 py-3 rounded-xl text-center">Get Started Free</a>
          </motion.div>
        )}
      </AnimatePresence>
    </nav>
  );
};

const Hero = () => (
  <section className="relative pt-40 pb-20 px-6 max-w-7xl mx-auto flex flex-col items-center text-center">
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.8 }}
      className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full glass border-memzent-glow/20 mb-8"
    >
      <Brain size={14} className="text-memzent-glow" />
      <span className="text-xs font-bold tracking-widest uppercase opacity-80">Memory & Security Layer for AI Agents</span>
    </motion.div>

    <motion.h1
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 1, ease: "easeOut" }}
      className="text-6xl md:text-8xl font-black leading-[0.9] tracking-tighter mb-8"
    >
      THE AI AGENT<br />
      <span className="text-transparent bg-clip-text bg-gradient-to-r from-memzent-glow via-memzent-purple to-memzent-accent">MEMORY LAYER</span>
    </motion.h1>

    <motion.p
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ delay: 0.5, duration: 1 }}
      className="text-lg md:text-xl opacity-80 max-w-2xl mb-4 leading-relaxed"
    >
      Memzent operates as an Intelligent Semantic Proxy — intercepting and optimizing traffic between
      your agents, MCP tools, and LLM providers with semantic caching, RBAC, and enterprise-grade routing.
    </motion.p>

    <motion.p
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ delay: 0.7, duration: 1 }}
      className="text-sm opacity-60 max-w-xl mb-12"
    >
      Pay only for what you use. Top up from $5. Cache hits cost 80% less.
    </motion.p>

    <motion.div
      initial={{ opacity: 0, y: 30 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.8, duration: 0.8 }}
      className="flex flex-col md:flex-row gap-6"
    >
      <a href={appUrl + "/login"} className="bg-memzent-glow text-black font-black px-10 py-5 rounded-2xl text-lg hover:shadow-[0_0_30px_rgba(0,243,255,0.4)] hover:scale-105 transition-all flex items-center gap-3 group">
        Start Free <ChevronRight className="group-hover:translate-x-1 transition-transform" />
      </a>
      <a href={appUrl + "/docs"} className="glass border-white/10 font-black px-10 py-5 rounded-2xl text-lg hover:bg-white/5 transition-all flex items-center gap-3">
        Documentation <Terminal size={20} />
      </a>
    </motion.div>

    {/* Live stat badges */}
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ delay: 1.2 }}
      className="flex flex-wrap justify-center gap-4 mt-16"
    >
      {[
        { icon: <Zap size={12} />, label: "80%+ GPU avoidance", color: "text-memzent-glow" },
        { icon: <Shield size={12} />, label: "Entity-aware caching", color: "text-memzent-purple" },
        { icon: <Brain size={12} />, label: "Semantic memory", color: "text-memzent-accent" },
        { icon: <Cpu size={12} />, label: "Multi-LLM routing", color: "text-white/80" },
      ].map(b => (
        <div key={b.label} className={`flex items-center gap-2 px-4 py-2 rounded-full glass border-white/5 text-xs font-bold ${b.color}`}>
          {b.icon} {b.label}
        </div>
      ))}
    </motion.div>

    {/* Terminal Quickstart */}
    <motion.div
      initial={{ opacity: 0, y: 30 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 1.5, duration: 0.8 }}
      className="mt-16 w-full max-w-2xl"
    >
      <div className="glass rounded-2xl overflow-hidden border-memzent-glow/10">
        <div className="flex items-center gap-2 px-4 py-3 border-b border-white/5">
          <div className="w-3 h-3 rounded-full bg-red-500/60" />
          <div className="w-3 h-3 rounded-full bg-yellow-500/60" />
          <div className="w-3 h-3 rounded-full bg-green-500/60" />
          <span className="ml-2 text-[10px] font-mono text-white/30">terminal</span>
        </div>
        <div className="p-6 font-mono text-sm space-y-2">
          <div className="text-white/40">
            <span className="text-memzent-accent">$</span> curl -X POST https://api.memzent.ai/v1/route \
          </div>
          <div className="text-white/40 pl-4">
            -H &quot;Authorization: Bearer mk_live_...&quot; \
          </div>
          <div className="text-white/40 pl-4">
            -d &apos;{`{"prompt": "Deploy staging server", "tools": ["github", "aws"]}`}&apos;
          </div>
          <div className="mt-4 text-memzent-glow/80">
            <span className="text-white/30">→</span> Cache HIT (semantic similarity: 0.94) — <span className="text-memzent-accent">saved $0.012</span>
          </div>
          <div className="text-memzent-glow/60">
            <span className="text-white/30">→</span> Routed to: github.create_deployment, aws.ecs_update
          </div>
          <div className="text-white/20">
            <span className="text-white/30">→</span> Latency: 12ms (vs 2400ms without cache)
          </div>
        </div>
      </div>
    </motion.div>
  </section>
);

const PAYGSection = () => (
  <section id="payg" className="py-40 px-6 max-w-7xl mx-auto">
    <div className="text-center mb-20">
      <h2 className="text-5xl md:text-7xl font-black tracking-tighter mb-6">
        PAY AS YOU <span className="text-transparent bg-clip-text bg-gradient-to-r from-memzent-glow to-memzent-accent">GO</span>
      </h2>
      <p className="text-lg opacity-70 max-w-2xl mx-auto">No surprise bills. No seat licenses. Start free, top up when needed, and let the semantic cache do the heavy lifting.</p>
    </div>

    {/* Pricing Tiers */}
    <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-20">
      {[
        {
          name: 'Individual',
          price: 'Free',
          priceDetail: 'forever',
          desc: 'Perfect for exploration & local AI agents.',
          features: ['50 requests / day', 'Ollama local models', 'Basic semantic cache', 'MCP tool support'],
          cta: 'Start Free',
          highlight: false
        },
        {
          name: 'Pro',
          price: '$29',
          priceDetail: '/mo + PAYG',
          desc: 'Production agents with full provider access.',
          features: ['100 RPM rate limit', 'OpenAI, Anthropic, Gemini', 'Priority semantic routing', 'PAYG top-up included'],
          cta: 'Go Pro',
          highlight: true
        },
        {
          name: 'Business',
          price: '$99',
          priceDetail: '/mo + PAYG',
          desc: 'Enterprise routing with full RBAC.',
          features: ['1000+ RPM', 'Multi-org RBAC', 'Deep analytics & audit', 'Volume discounts & SLA'],
          cta: 'Contact Sales',
          highlight: false
        }
      ].map((plan) => (
        <div key={plan.name} className={`p-10 rounded-[40px] glass flex flex-col gap-8 border-white/5 transition-all hover:border-memzent-glow/20 ${plan.highlight ? 'bg-memzent-glow/5 border-memzent-glow/20 scale-105 z-10 shadow-[0_0_40px_rgba(0,243,255,0.1)]' : ''}`}>
          <div>
            <div className="text-3xl font-black italic tracking-tighter mb-1">{plan.name}</div>
            <p className="text-xs font-bold opacity-60 uppercase tracking-wide">{plan.desc}</p>
          </div>
          <div className="flex items-baseline gap-2">
            <span className="text-6xl font-black tracking-tighter">{plan.price}</span>
            <span className="text-xs font-black uppercase opacity-50">{plan.priceDetail}</span>
          </div>
          <ul className="space-y-3 flex-1">
            {plan.features.map(f => (
              <li key={f} className="flex gap-3 text-sm font-bold opacity-80">
                <Check size={16} className={plan.highlight ? "text-memzent-glow" : "text-white/60"} /> {f}
              </li>
            ))}
          </ul>
          <a
            href={appUrl + "/login"}
            className={`w-full py-5 rounded-2xl text-center text-sm font-black uppercase tracking-[0.2em] transition-all ${plan.highlight ? 'bg-memzent-glow text-black shadow-[0_0_20px_rgba(0,243,255,0.3)] hover:shadow-[0_0_30px_rgba(0,243,255,0.5)]' : 'bg-white/5 text-white hover:bg-white/10'}`}
          >
            {plan.cta}
          </a>
        </div>
      ))}
    </div>

    {/* PAYG Explainer */}
    <div className="glass rounded-[40px] p-12 border-white/5 relative overflow-hidden">
      <div className="absolute inset-0 bg-gradient-to-br from-memzent-glow/5 via-transparent to-memzent-purple/5 pointer-events-none" />
      <div className="relative z-10">
        <div className="flex items-center gap-3 mb-8">
          <DollarSign size={20} className="text-memzent-accent" />
          <h3 className="text-2xl font-black tracking-tighter uppercase">Token Economy Explained</h3>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
          {[
            { step: "01", title: "Top Up Balance", desc: "Add any amount ($5 minimum) to your organization's token balance via Stripe.", icon: <DollarSign size={24} />, color: "text-memzent-glow" },
            { step: "02", title: "Send Prompts", desc: "Agents send prompts through the Memzent Gateway. Rate limiting and RBAC are enforced first.", icon: <ArrowRight size={24} />, color: "text-white/80" },
            { step: "03", title: "Cache Check", desc: "Memzent checks its semantic vector cache. Hits cost 80% less — only infra overhead.", icon: <Zap size={24} />, color: "text-memzent-glow" },
            { step: "04", title: "Deducted Transparently", desc: "LLM tokens are charged at provider cost. Cache hits are discounted. You see every deduction.", icon: <Activity size={24} />, color: "text-memzent-accent" },
          ].map(item => (
            <div key={item.step} className="space-y-3">
              <div className={`text-[10px] font-black uppercase tracking-[0.3em] opacity-50`}>{item.step}</div>
              <div className={item.color}>{item.icon}</div>
              <h4 className="text-sm font-black uppercase tracking-tight">{item.title}</h4>
              <p className="text-xs opacity-65 leading-relaxed">{item.desc}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  </section>
);

const WhyMemzent = () => (
  <section id="why" className="py-32 px-6 max-w-7xl mx-auto">
    <div className="text-center mb-16">
      <h2 className="text-5xl md:text-7xl font-black tracking-tighter mb-6">
        WHY <span className="text-transparent bg-clip-text bg-gradient-to-r from-memzent-glow to-memzent-purple">MEMZENT</span>
      </h2>
      <p className="text-lg opacity-70 max-w-2xl mx-auto">Most proxy layers solve one problem. Memzent solves the full stack — caching, routing, security, and memory — in one semantic layer.</p>
    </div>

    <div className="glass rounded-[40px] p-8 md:p-12 border-white/5 overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-white/10">
            <th className="text-left py-4 px-4 font-black text-white/40 uppercase tracking-widest text-xs">Feature</th>
            <th className="text-center py-4 px-4 font-black text-memzent-glow uppercase tracking-widest text-xs">Memzent</th>
            <th className="text-center py-4 px-4 font-black text-white/30 uppercase tracking-widest text-xs">LiteLLM</th>
            <th className="text-center py-4 px-4 font-black text-white/30 uppercase tracking-widest text-xs">Helicone</th>
            <th className="text-center py-4 px-4 font-black text-white/30 uppercase tracking-widest text-xs">Portkey</th>
          </tr>
        </thead>
        <tbody className="text-white/70 font-bold">
          {[
            { feature: "4-Layer Semantic Cache (L1/L1.5/L1b/L2)", memzent: true, litellm: false, helicone: false, portkey: false },
                { feature: "Entity-Aware Cache Guard", memzent: true, litellm: false, helicone: false, portkey: false },
                { feature: "Canonical Prompt Normalization (L1.5)", memzent: true, litellm: false, helicone: false, portkey: false },
                { feature: "Multi-LLM Routing", memzent: true, litellm: true, helicone: true, portkey: true },
                { feature: "MCP Tool Registry + Execution", memzent: true, litellm: false, helicone: false, portkey: true },
                { feature: "Workflow Discovery & Auto-Shortcuts", memzent: true, litellm: false, helicone: false, portkey: false },
                { feature: "RBAC + Multi-Tenant Governance", memzent: true, litellm: true, helicone: true, portkey: true },
                { feature: "Agent Memory (Session + Semantic)", memzent: true, litellm: true, helicone: false, portkey: true },
                { feature: "Offline Learning & Pattern Mining", memzent: true, litellm: false, helicone: false, portkey: false },
                { feature: "GPU Avoidance Analytics", memzent: true, litellm: false, helicone: false, portkey: false },
                { feature: "Spend Limits & Budget Forecast", memzent: true, litellm: false, helicone: false, portkey: true },
                { feature: "Real-time Observability", memzent: true, litellm: true, helicone: true, portkey: true },
                { feature: "Open Source (Apache 2.0)", memzent: true, litellm: true, helicone: true, portkey: true },
                { feature: "Self-Hosted (One Command)", memzent: true, litellm: true, helicone: true, portkey: true },
          ].map(row => (
            <tr key={row.feature} className="border-b border-white/5 hover:bg-white/[0.02] transition-colors">
              <td className="py-4 px-4">{row.feature}</td>
              <td className="text-center py-4 px-4">{row.memzent ? <Check size={18} className="inline text-memzent-glow" /> : <span className="text-white/20">—</span>}</td>
              <td className="text-center py-4 px-4">{row.litellm ? <Check size={18} className="inline text-white/40" /> : <span className="text-white/20">—</span>}</td>
              <td className="text-center py-4 px-4">{row.helicone ? <Check size={18} className="inline text-white/40" /> : <span className="text-white/20">—</span>}</td>
              <td className="text-center py-4 px-4">{row.portkey ? <Check size={18} className="inline text-white/40" /> : <span className="text-white/20">—</span>}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>

    <div className="text-center mt-12">
      <p className="text-sm opacity-50 font-bold">Comparison based on public documentation as of June 2026. All products are excellent — Memzent's edge is the 4-layer cache architecture with entity-aware guards that no competitor replicates.</p>
    </div>
  </section>
);

const EvolutionPipeline = () => (
  <section className="py-32 px-6 max-w-7xl mx-auto">
    <div className="text-center mb-16">
      <h2 className="text-5xl md:text-7xl font-black tracking-tighter mb-6">
        EVOLUTION <span className="text-transparent bg-clip-text bg-gradient-to-r from-memzent-glow to-memzent-accent">PIPELINE</span>
      </h2>
      <p className="text-lg opacity-70 max-w-2xl mx-auto">
        Six layers of intelligence that eliminate redundant GPU inference.
        Every request is filtered through entity extraction, multi-layer caching, and
        offline learning — before the LLM is ever invoked.
      </p>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
      {[
        {
          icon: <Scan size={28} />,
          title: "Entity Extraction",
          tag: "E1",
          desc: "Regex-based typed entity extraction (<1ms) identifies accounts, customers, amounts, and dates — preventing false cache hits across similar prompts.",
          color: "text-memzent-glow",
          borderColor: "border-memzent-glow/20"
        },
        {
          icon: <Layers size={28} />,
          title: "L1b Hot Path Cache",
          tag: "E2",
          desc: "Entity-keyed deterministic cache in Valkey. Same entities = instant response, zero vector search. Resolves 20-30% of repeat requests sub-millisecond.",
          color: "text-green-400",
          borderColor: "border-green-500/20"
        },
        {
          icon: <Activity size={28} />,
          title: "Offline Learning Plane",
          tag: "E3",
          desc: "Asynchronous telemetry pipeline with request, cache, and workflow miners. Discovers patterns without adding latency. PII-safe by design.",
          color: "text-purple-400",
          borderColor: "border-purple-500/20"
        },
        {
          icon: <Database size={28} />,
          title: "Workflow Registry",
          tag: "E4",
          desc: "Automatically discovers and registers multi-step tool sequences. Approved workflows execute as single-shot shortcuts, skipping per-step routing.",
          color: "text-blue-400",
          borderColor: "border-blue-500/20"
        },
        {
          icon: <BarChart3 size={28} />,
          title: "GPU Avoidance Metrics",
          tag: "E5",
          desc: "Track the percentage of requests resolved without LLM inference. Prometheus counters for entity types, cache layers, and avoidance rates.",
          color: "text-memzent-accent",
          borderColor: "border-memzent-accent/20"
        },
        {
          icon: <Brain size={28} />,
          title: "Pattern Mining",
          tag: "E6",
          desc: "Experimental Markov chain analysis predicts next-likely requests and speculatively pre-warms the L1b cache for zero-latency first hits.",
          color: "text-yellow-400",
          borderColor: "border-yellow-500/20"
        },
      ].map(item => (
        <motion.div
          key={item.tag}
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className={`glass p-8 rounded-3xl ${item.borderColor} hover:border-opacity-50 transition-all group`}
        >
          <div className="flex items-center gap-3 mb-4">
            <span className={`text-[10px] font-black uppercase tracking-[0.2em] px-2 py-1 rounded bg-white/5 ${item.color}`}>{item.tag}</span>
            <div className={item.color}>{item.icon}</div>
          </div>
          <h3 className="text-lg font-black tracking-tight mb-3 uppercase">{item.title}</h3>
          <p className="text-sm opacity-60 leading-relaxed">{item.desc}</p>
        </motion.div>
      ))}
    </div>

    <div className="text-center mt-16">
      <a href={appUrl + "/docs/entity-extraction"} className="inline-flex items-center gap-2 text-sm font-black text-memzent-glow hover:underline uppercase tracking-widest">
        Read the Technical Docs <ArrowRight size={14} />
      </a>
    </div>
  </section>
);

const Pillars = () => (
  <section className="py-20 px-6 max-w-7xl mx-auto space-y-20">
    {/* Security Section */}
    <div id="security" className="grid grid-cols-1 md:grid-cols-2 gap-12 items-center">
      <div>
        <h2 className="text-4xl md:text-5xl font-black tracking-tighter mb-6">BULLETPROOF<br /><span className="text-memzent-purple">GOVERNANCE</span></h2>
        <p className="text-lg opacity-80 mb-8 leading-relaxed">
          Memzent enforces enterprise-grade security at the semantic layer.
          Audit every prompt, restrict model access via RBAC, and protect your
          data with gRPC mTLS and hardware-backed JWT authentication.
        </p>
        <ul className="space-y-4">
          <li className="flex items-center gap-3 font-bold opacity-90"><Lock className="text-memzent-purple" size={18} /> Zero-Trust AI Access</li>
          <li className="flex items-center gap-3 font-bold opacity-90"><Shield className="text-memzent-purple" size={18} /> Semantic Data Guardrails</li>
          <li className="flex items-center gap-3 font-bold opacity-90"><Database className="text-memzent-purple" size={18} /> Full RBAC & Auth Integration</li>
        </ul>
      </div>
      <div className="grid grid-cols-1 gap-6">
        <FeatureCard icon={Shield} title="RBAC Gateway" desc="Limit tool execution and model access based on deep user identity scopes." color="text-memzent-purple" />
        <FeatureCard icon={Lock} title="Secure gRPC" desc="Distributed microservices communicate via encrypted mTLS for absolute data safety." color="text-memzent-purple" />
      </div>
    </div>

    {/* Observability Section */}
    <div id="observability" className="grid grid-cols-1 md:grid-cols-2 gap-12 items-center">
      <div className="order-1 md:order-2">
        <h2 className="text-4xl md:text-5xl font-black tracking-tighter mb-6">INTELLIGENCE<br /><span className="text-memzent-glow">MONITORING</span></h2>
        <p className="text-lg opacity-80 mb-8 leading-relaxed">
          Stop guessing AI performance. Memzent provides deep telemetry into every LLM request.
          Monitor latency, token cost, and cache hit rates in real-time.
        </p>
        <div className="p-6 glass rounded-2xl border-memzent-glow/20 font-mono text-sm space-y-2">
          <div className="flex justify-between text-memzent-glow opacity-60"><span>LATENCY_P95</span> <span>42ms</span></div>
          <div className="flex justify-between text-memzent-glow opacity-80"><span>CACHE_HIT_RATIO</span> <span>87.4%</span></div>
          <div className="flex justify-between text-memzent-accent opacity-80"><span>TOKENS_SAVED</span> <span>1.2M</span></div>
          <div className="w-full h-1 bg-white/5 rounded-full overflow-hidden">
            <motion.div initial={{ x: '-100%' }} animate={{ x: '0%' }} transition={{ repeat: Infinity, duration: 2 }} className="w-1/3 h-full bg-memzent-glow shadow-[0_0_10px_#00f3ff]" />
          </div>
        </div>
      </div>
      <div className="grid grid-cols-1 gap-6 order-2 md:order-1">
        <FeatureCard icon={Activity} title="Real-time Metrics" desc="Expose structured Prometheus metrics for every router decision and cache hit." color="text-memzent-glow" />
        <FeatureCard icon={Monitor} title="Trace Everything" desc="End-to-end tracing for complex multi-tool agentic workflows." color="text-memzent-glow" />
      </div>
    </div>

    {/* ROI / Cost Section */}
    <div id="cost" className="text-center space-y-8 max-w-4xl mx-auto">
      <h2 className="text-5xl md:text-7xl font-black tracking-tighter text-transparent bg-clip-text bg-gradient-to-r from-memzent-glow via-memzent-accent to-memzent-purple">
        80% COST REDUCTION
      </h2>
      <p className="text-xl opacity-80 leading-relaxed">
        Our semantic cache engine detects repeat intents across your organization.
        Cache hits are charged at infra cost only — 80% less than full LLM inference.
      </p>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="glass p-8 rounded-3xl border-memzent-accent/20">
          <div className="text-4xl font-black text-memzent-accent mb-2">80%</div>
          <div className="text-xs font-bold uppercase opacity-65">Cache Discount</div>
        </div>
        <div className="glass p-8 rounded-3xl border-memzent-glow/20">
          <div className="text-4xl font-black text-memzent-glow mb-2">&lt;15ms</div>
          <div className="text-xs font-bold uppercase opacity-65">Cache Latency</div>
        </div>
        <div className="glass p-8 rounded-3xl border-memzent-accent/20">
          <div className="text-4xl font-black text-memzent-accent mb-2">$5</div>
          <div className="text-xs font-bold uppercase opacity-65">Minimum Top-Up</div>
        </div>
      </div>
    </div>
  </section>
);

const Footer = () => (
  <footer className="pt-40 pb-10 px-6 border-t border-white/5">
    <div className="max-w-7xl mx-auto grid grid-cols-1 md:grid-cols-4 gap-12 mb-20 text-sm opacity-65 font-medium tracking-tight">
      <div className="col-span-2">
        <div className="flex items-center gap-2 mb-4 opacity-100 italic font-black text-xl">
          <div className="w-6 h-6 rounded-md bg-gradient-to-br from-memzent-glow to-memzent-purple flex items-center justify-center">
            <span className="text-black text-xs font-black">M</span>
          </div>
          MEMZENT.ai
        </div>
        <p className="max-w-sm">The intelligent semantic proxy and memory layer for autonomous AI agents. Pay-as-you-go, enterprise-grade, open-source core.</p>
      </div>
      <div className="space-y-3">
        <div className="font-black text-white mb-4">PLATFORM</div>
        <a href="#payg" className="block hover:text-memzent-glow">Pricing</a>
        <a href={appUrl + "/docs/architecture"} className="block hover:text-memzent-glow">Architecture</a>
        <a href={appUrl + "/docs/semantic-proxy"} className="block hover:text-memzent-glow">Semantic Router</a>
        <a href={appUrl + "/docs/entity-extraction"} className="block hover:text-memzent-glow">Entity Extraction</a>
        <a href={appUrl + "/docs/cache-layers"} className="block hover:text-memzent-glow">Cache Layers</a>
        <a href={appUrl + "/docs/tool-registry"} className="block hover:text-memzent-glow">MCP Tools</a>
      </div>
      <div className="space-y-3">
        <div className="font-black text-white mb-4">RESOURCES</div>
        <a href={appUrl + "/docs"} className="block hover:text-memzent-glow">Documentation</a>
        <a href={appUrl + "/docs/quickstart"} className="block hover:text-memzent-glow">Quickstart</a>
        <a href={appUrl + "/docs/api-reference"} className="block hover:text-memzent-glow">API Reference</a>
        <Link to="/blog" className="block hover:text-memzent-glow">Blog</Link>
        <a href="https://github.com/Opsylux/Memzent.AI" target="_blank" rel="noopener" className="block hover:text-memzent-glow">GitHub</a>
        <a href={appUrl + "/login"} className="block hover:text-memzent-glow">Dashboard</a>
      </div>
    </div>
    <div className="max-w-7xl mx-auto flex items-center justify-between text-[10px] font-black uppercase tracking-widest opacity-40">
      <span>© 2026 Memzent.ai — All rights reserved</span>
      <span>memzent.ai</span>
    </div>
  </footer>
);

const LandingPage = () => (
  <>
    <Hero />
    <PAYGSection />
    <WhyMemzent />
    <EvolutionPipeline />
    <Pillars />
  </>
);

export default function App() {
  const location = useLocation();

  return (
    <div className="bg-memzent-dark text-white min-h-screen selection:bg-memzent-glow selection:text-black font-outfit">
      <Navbar />
      <Routes location={location} key={location.pathname}>
        <Route path="/" element={<LandingPage />} />
        <Route path="/blog" element={<BlogListPage />} />
        <Route path="/blog/:slug" element={<BlogPostPage />} />
      </Routes>
      <Footer />
    </div>
  );
}
