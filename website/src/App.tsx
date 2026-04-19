import { motion } from 'framer-motion';
import {
  Shield,
  Activity,
  ChevronRight,
  Terminal,
  Database,
  Lock,
  Monitor
} from 'lucide-react';

import FeatureCard from './components/FeatureCard';

const appUrl = import.meta.env.VITE_APP_URL || "http://localhost:3000"

const Pricing = () => (
  <section id="pricing" className="py-40 px-6 max-w-7xl mx-auto">
    <div className="text-center mb-20">
      <h2 className="text-5xl md:text-7xl font-black tracking-tighter mb-6 underline decoration-aura-glow underline-offset-8">PRICING CLUSTERS</h2>
      <p className="text-xl opacity-60 font-bold uppercase tracking-widest text-xs">Capacity-based infrastructure billing for every scale.</p>
    </div>
    <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
      {[
        { name: 'Individual', price: '$0', desc: 'Personal experimentation.', features: ['10 RPM Limit', 'Standard Latency', 'MCP Support'], cta: 'Join Now' },
        { name: 'Pro', price: '$29', desc: 'Professional agentic scale.', features: ['100 RPM Limit', 'Sub-ms Latency', 'Priority Routing'], cta: 'Go Pro', highlight: true },
        { name: 'Business', price: '$99', desc: 'Enterprise data backbone.', features: ['1000+ RPM Limit', 'Semantic RBAC', 'Deep Analytics'], cta: 'Contact Sales' }
      ].map((plan) => (
        <div key={plan.name} className={`p-10 rounded-[40px] glass flex flex-col items-start gap-8 border-white/5 transition-all hover:border-aura-glow/20 ${plan.highlight ? 'bg-aura-glow/5 border-aura-glow/20 scale-105 z-10' : ''}`}>
          <div className="text-3xl font-black italic tracking-tighter">{plan.name}</div>
          <div className="flex items-baseline gap-2">
            <span className="text-6xl font-black tracking-tighter">{plan.price}</span>
            <span className="text-xs font-black uppercase opacity-20">/mo</span>
          </div>
          <p className="text-sm font-bold opacity-40 uppercase tracking-wide">{plan.desc}</p>
          <ul className="space-y-4 flex-1">
            {plan.features.map(f => (
              <li key={f} className="flex gap-3 text-sm font-bold opacity-70">
                <Check size={16} className="text-aura-glow" /> {f}
              </li>
            ))}
          </ul>
          <a
            href={appUrl + "/login"}
            className={`w-full py-5 rounded-2xl text-center text-sm font-black uppercase tracking-[0.2em] transition-all ${plan.highlight ? 'bg-aura-glow text-black shadow-[0_0_20px_rgba(0,243,255,0.3)]' : 'bg-white text-black hover:bg-aura-glow'
              }`}
          >
            {plan.cta}
          </a>
        </div>
      ))}
    </div>
  </section>
);

const Check = ({ className, size }: any) => (
  <svg className={className} width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>
);

const Navbar = () => (
  <nav className="fixed top-0 left-0 right-0 z-50 glass px-6 py-4 flex justify-between items-center m-4 rounded-2xl">
    <div className="flex items-center gap-2">
      <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-aura-glow to-aura-purple flex items-center justify-center">
        <span className="text-black font-black text-xl italic select-none">A</span>
      </div>
      <span className="text-2xl font-black tracking-tighter">AURA</span>
    </div>
    <div className="hidden md:flex gap-8 text-sm font-medium opacity-70">
      <a href="#security" className="hover:text-aura-glow transition-colors">Security</a>
      <a href="#observability" className="hover:text-aura-glow transition-colors">Observability</a>
      <a href="#pricing" className="hover:text-aura-glow transition-colors">Pricing</a>
    </div>
    <div className="flex gap-4">
      <a href={appUrl + "/login"} className="text-sm font-bold opacity-60 hover:opacity-100 px-4 py-2 transition-all cursor-pointer">Login</a>
      <a href={appUrl + "/login"} className="bg-white text-black text-sm font-black px-6 py-2 rounded-xl hover:bg-aura-glow hover:scale-105 transition-all cursor-pointer shadow-[0_0_20px_rgba(255,255,255,0.1)]">Get Started</a>
    </div>
  </nav>
);

const Hero = () => (
  <section className="relative pt-40 pb-20 px-6 max-w-7xl mx-auto flex flex-col items-center text-center">
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.8 }}
      className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full glass border-aura-glow/20 mb-8"
    >
      <Shield size={14} className="text-aura-glow" />
      <span className="text-xs font-bold tracking-widest uppercase opacity-70">Enterprise AI Resilience v1.0</span>
    </motion.div>

    <motion.h1
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 1, ease: "easeOut" }}
      className="text-6xl md:text-8xl font-black leading-[0.9] tracking-tighter mb-8"
    >
      THE SECURE<br />
      <span className="text-transparent bg-clip-text bg-gradient-to-r from-aura-glow via-aura-purple to-aura-accent">AI BACKBONE</span>
    </motion.h1>

    <motion.p
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ delay: 0.5, duration: 1 }}
      className="text-lg md:text-xl opacity-60 max-w-2xl mb-12 leading-relaxed"
    >
      Aura is an AI gateway designed for massive scale. Secure every prompt,
      observe every token, and optimize for 90% cost reduction with Semantic Caching.
    </motion.p>

    <motion.div
      initial={{ opacity: 0, y: 30 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.8, duration: 0.8 }}
      className="flex flex-col md:flex-row gap-6"
    >
      <a href={appUrl + "/login"} className="bg-aura-glow text-black font-black px-10 py-5 rounded-2xl text-lg hover:shadow-[0_0_30px_rgba(0,243,255,0.4)] hover:scale-105 transition-all flex items-center gap-3 group">
        Deploy Infrastructure <ChevronRight className="group-hover:translate-x-1 transition-transform" />
      </a>
      <a href={appUrl + "/docs"} className="glass border-white/10 font-black px-10 py-5 rounded-2xl text-lg hover:bg-white/5 transition-all flex items-center gap-3">
        Documentation <Terminal size={20} />
      </a>
    </motion.div>
  </section>
);

const Pillars = () => (
  <section className="py-20 px-6 max-w-7xl mx-auto space-y-20">
    {/* Security Section */}
    <div id="security" className="grid grid-cols-1 md:grid-cols-2 gap-12 items-center">
      <div>
        <h2 className="text-4xl md:text-5xl font-black tracking-tighter mb-6">BULLETPROOF<br /><span className="text-aura-purple">GOVERNANCE</span></h2>
        <p className="text-lg opacity-60 mb-8 leading-relaxed">
          Aura enforces enterprise-grade security at the semantic layer.
          Audit every prompt, restrict model access via RBAC, and protect your
          data with gRPC mTLS and hardware-backed JWT authentication.
        </p>
        <ul className="space-y-4">
          <li className="flex items-center gap-3 font-bold opacity-80"><Lock className="text-aura-purple" size={18} /> Zero-Trust AI Access</li>
          <li className="flex items-center gap-3 font-bold opacity-80"><Shield className="text-aura-purple" size={18} /> Semantic Data Guardrails</li>
          <li className="flex items-center gap-3 font-bold opacity-80"><Database className="text-aura-purple" size={18} /> Full RBAC & Auth Integration</li>
        </ul>
      </div>
      <div className="grid grid-cols-1 gap-6">
        <FeatureCard
          icon={Shield}
          title="RBAC Gateway"
          desc="Limit tool execution and model access based on deep user identity scopes."
          color="text-aura-purple"
        />
        <FeatureCard
          icon={Lock}
          title="Secure gRPC"
          desc="Distributed microservices communicate via encrypted mTLS for absolute data safety."
          color="text-aura-purple"
        />
      </div>
    </div>

    {/* Observability Section */}
    <div id="observability" className="grid grid-cols-1 md:grid-cols-2 gap-12 items-center md:flex-row-reverse">
      <div className="order-1 md:order-2">
        <h2 className="text-4xl md:text-5xl font-black tracking-tighter mb-6">INTELLIGENCE<br /><span className="text-aura-glow">MONITORING</span></h2>
        <p className="text-lg opacity-60 mb-8 leading-relaxed">
          Stop guessing AI performance. Aura provides deep telemetry into every LLM request.
          Monitor latency, success rates, and token flow in real-time with integrated
          Prometheus and OpenTelemetry support.
        </p>
        <div className="p-6 glass rounded-2xl border-aura-glow/20 font-mono text-sm space-y-2">
          <div className="flex justify-between text-aura-glow opacity-60"><span>LATENCY_P95</span> <span>42ms</span></div>
          <div className="flex justify-between text-aura-glow opacity-80"><span>TOKEN_FLOW</span> <span>1.2M/min</span></div>
          <div className="w-full h-1 bg-white/5 rounded-full overflow-hidden">
            <motion.div initial={{ x: '-100%' }} animate={{ x: '0%' }} transition={{ repeat: Infinity, duration: 2 }} className="w-1/3 h-full bg-aura-glow shadow-[0_0_10px_#00f3ff]" />
          </div>
        </div>
      </div>
      <div className="grid grid-cols-1 gap-6 order-2 md:order-1">
        <FeatureCard
          icon={Activity}
          title="Real-time Metrics"
          desc="Expose structured Prometheus metrics for every router decision and cache hit."
          color="text-aura-glow"
        />
        <FeatureCard
          icon={Monitor}
          title="Trace Everything"
          desc="End-to-end tracing for complex multi-tool agentic workflows."
          color="text-aura-glow"
        />
      </div>
    </div>

    {/* ROI / Cost Section */}
    <div id="roi" className="text-center space-y-8 max-w-4xl mx-auto">
      <h2 className="text-5xl md:text-7xl font-black tracking-tighter text-transparent bg-clip-text bg-gradient-to-r from-aura-glow via-aura-accent to-aura-purple">
        90% COST REDUCTION
      </h2>
      <p className="text-xl opacity-60 leading-relaxed">
        Our Semantic Caching engine (powered by Valkey & Qdrant) detects repeat intents
        across your entire organization. Stop paying for the same tokens twice.
      </p>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="glass p-8 rounded-3xl border-aura-accent/20">
          <div className="text-4xl font-black text-aura-accent mb-2">98%</div>
          <div className="text-xs font-bold uppercase opacity-40">Cache Hit Rate</div>
        </div>
        <div className="glass p-8 rounded-3xl border-aura-accent/20">
          <div className="text-4xl font-black text-aura-accent mb-2">-15ms</div>
          <div className="text-xs font-bold uppercase opacity-40">Latency Impact</div>
        </div>
        <div className="glass p-8 rounded-3xl border-aura-accent/20">
          <div className="text-4xl font-black text-aura-accent mb-2">$0.00</div>
          <div className="text-xs font-bold uppercase opacity-40">Cached Cost</div>
        </div>
      </div>
    </div>
  </section>
);

const Footer = () => (
  <footer className="pt-40 pb-10 px-6 border-t border-white/5">
    <div className="max-w-7xl mx-auto grid grid-cols-1 md:grid-cols-4 gap-12 mb-20 text-sm opacity-40 font-medium tracking-tight">
      <div className="col-span-2">
        <div className="flex items-center gap-2 mb-4 opacity-100 italic font-black text-xl">AURA.</div>
        <p className="max-w-sm">Project Aura: Modern AI infrastructure for security, scale, and efficiency.</p>
      </div>
      <div className="space-y-3">
        <div className="font-black text-white mb-4">PLATFORM</div>
        <a href="#" className="block hover:text-aura-glow">Gateway API</a>
        <a href="#" className="block hover:text-aura-glow">Semantic Router</a>
        <a href="#" className="block hover:text-aura-glow">Monitoring</a>
      </div>
      <div className="space-y-3">
        <div className="font-black text-white mb-4">LEGAL</div>
        <a href="#" className="block hover:text-aura-glow">Compliance</a>
        <a href="#" className="block hover:text-aura-glow">Security Audit</a>
        <a href="#" className="block hover:text-aura-glow">ROI Calculator</a>
      </div>
    </div>
  </footer>
);

export default function App() {
  return (
    <div className="bg-aura-dark text-white min-h-screen selection:bg-aura-glow selection:text-black font-outfit">
      <Navbar />
      <Hero />
      <Pillars />
      <Pricing />
      <Footer />
    </div>
  );
}
