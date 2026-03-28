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
      <a href="#roi" className="hover:text-aura-glow transition-colors">ROI & Cost</a>
    </div>
    <div className="flex gap-4">
      <button className="text-sm font-bold opacity-60 hover:opacity-100 px-4 py-2 transition-all">Portal</button>
      <button className="bg-white text-black text-sm font-black px-6 py-2 rounded-xl hover:bg-aura-glow hover:scale-105 transition-all">Get Started</button>
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
      <button className="bg-aura-glow text-black font-black px-10 py-5 rounded-2xl text-lg hover:shadow-[0_0_30px_rgba(0,243,255,0.4)] hover:scale-105 transition-all flex items-center gap-3 group">
        Deploy Infrastructure <ChevronRight className="group-hover:translate-x-1 transition-transform" />
      </button>
      <button className="glass border-white/10 font-black px-10 py-5 rounded-2xl text-lg hover:bg-white/5 transition-all flex items-center gap-3">
        Documentation <Terminal size={20} />
      </button>
    </motion.div>
  </section>
);

const FeatureCard = ({ icon: Icon, title, desc, color }: any) => (
  <motion.div 
    initial={{ opacity: 0, y: 20 }}
    whileInView={{ opacity: 1, y: 0 }}
    className="bento-card flex flex-col items-start gap-4"
  >
    <div className={`w-12 h-12 rounded-2xl flex items-center justify-center bg-white/5 border border-white/10 ${color}`}>
      <Icon size={24} />
    </div>
    <h3 className="text-2xl font-black tracking-tight">{title}</h3>
    <p className="opacity-50 font-medium leading-relaxed">{desc}</p>
  </motion.div>
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
      <Footer />
    </div>
  );
}
