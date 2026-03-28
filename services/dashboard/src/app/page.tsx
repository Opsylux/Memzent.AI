import { getAuraTools, getAuraStats } from "./actions";
import { MetricCard } from "@/components/metric-card";
import { 
  Zap, 
  Activity, 
  ShieldCheck, 
  Layers, 
  ArrowUpRight,
  Database,
  Cpu,
  Terminal
} from "lucide-react";
import { RoutingVisualizer } from "@/components/routing-visualizer";
import Link from 'next/link';

export default async function Page() {
  const [initialTools, stats] = await Promise.all([
    getAuraTools(),
    getAuraStats()
  ]);

  const total = stats.total_requests || 0;
  const hits = stats.cache_hits || 0;
  const semanticSavings = total > 0 ? ((hits / total) * 100).toFixed(1) : "0.0";
  const uptimeHours = Math.floor((stats.uptime_seconds || 0) / 3600);

  return (
    <div className="space-y-12">
      {/* KPI Section */}
      <section className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <MetricCard 
          label="Semantic Savings" 
          value={`${semanticSavings}%`} 
          trend="Cache Hits"
          trendDirection="up"
          icon={<Zap size={24} />}
          color="cyan"
          detail={`${hits} requests saved from LLM`}
        />
        <MetricCard 
          label="Total Engine Throughput" 
          value={total.toLocaleString()} 
          trend="Requests" 
          icon={<Layers size={24} />}
          color="purple"
          detail="Organization-wide Flow"
        />
        <MetricCard 
          label="Aura Tools Registry" 
          value={`${initialTools?.length || 0}`} 
          trend="Online" 
          icon={<ShieldCheck size={24} />}
          color="cyan"
          detail="Active Context Bindings"
        />
        <MetricCard 
          label="Gateway Uptime" 
          value={`${uptimeHours}h`} 
          trend="Stable" 
          icon={<Activity size={24} />}
          color="purple"
          detail="High-Speed Engine Status"
        />
      </section>

      {/* Main Intelligence Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Semantic Flow & Visualizer */}
        <section className="lg:col-span-2 space-y-6">
          <div className="flex items-center justify-between px-2">
            <div>
              <h2 className="text-xl font-black tracking-tighter">Neural Execution Trace</h2>
              <p className="text-xs font-bold text-white/30 uppercase tracking-widest mt-1">Real-time Intent-to-Tool Mapping</p>
            </div>
            <button className="glass px-4 py-2 rounded-xl text-[10px] font-black tracking-widest uppercase hover:bg-white/5 transition-all text-white/40 hover:text-aura-glow">
              View All Traces
            </button>
          </div>
          <RoutingVisualizer steps={[]} />
          
          <div className="stat-card glow-purple p-8 neural-bg border-white/5">
             <div className="flex items-center gap-4 mb-8">
                <div className="w-12 h-12 rounded-2xl bg-aura-purple/10 border border-aura-purple/20 flex items-center justify-center text-aura-purple">
                   <Cpu size={24} />
                </div>
                <div>
                   <h3 className="text-lg font-black tracking-tight leading-none">Intelligence Status</h3>
                   <p className="text-xs font-bold text-white/20 uppercase tracking-widest mt-1">Model Discovery & Routing Engine</p>
                </div>
             </div>
             <p className="text-sm font-medium text-white/50 leading-relaxed mb-8">
               Aura Router is currently managing <span className="text-white font-bold">12 Active Tools</span> across 4 providers. 
               Semantic clustering is <span className="text-aura-glow font-bold">OPTIMIZED</span> with 98% intent matching accuracy.
             </p>
             <div className="grid grid-cols-2 gap-4">
                <div className="p-4 rounded-2xl bg-white/5 border border-white/5">
                   <div className="text-[10px] font-black text-white/20 uppercase mb-1">Embedding Engine</div>
                   <div className="text-sm font-bold text-aura-glow">ModernBERT-v3 (Qdrant)</div>
                </div>
                <div className="p-4 rounded-2xl bg-white/5 border border-white/5">
                   <div className="text-[10px] font-black text-white/20 uppercase mb-1">Inference Provider</div>
                   <div className="text-sm font-bold text-aura-purple">Aura_Llama_3.1_70B</div>
                </div>
             </div>
          </div>
        </section>

        {/* Sidebar Analytics */}
        <section className="space-y-8">
          {/* Active Tools Mini Feed */}
          <div className="stat-card border-white/10 p-8 neural-bg">
            <div className="flex items-center justify-between mb-8">
              <h3 className="text-xs font-black tracking-widest text-white/40 uppercase leading-none italic font-roboto-mono flex items-center gap-2">
                <Database size={14} className="text-aura-glow" /> 
                Recent Tools Discovered
              </h3>
              <ArrowUpRight size={14} className="text-white/20" />
            </div>
            <div className="space-y-6">
              {(initialTools || []).slice(0, 5).map((tool: any) => (
                <div key={tool.id} className="flex items-center justify-between group cursor-pointer">
                  <div className="flex items-center gap-4">
                    <div className="w-2 h-2 rounded-full bg-aura-accent shadow-[0_0_8px_#00ff8e]" />
                    <div>
                      <div className="text-sm font-bold text-white group-hover:text-aura-glow transition-colors">{tool.name}</div>
                      <div className="text-[10px] font-mono text-white/20 uppercase tracking-tighter truncate w-32">{tool.id}</div>
                    </div>
                  </div>
                  <div className="text-[10px] font-black text-white/40 px-2 py-0.5 rounded-lg border border-white/10 group-hover:border-aura-glow/20 transition-all uppercase">
                    {tool.provider || 'MCP'}
                  </div>
                </div>
              ))}
              <Link href="/tools" className="block text-center pt-4 text-[10px] font-black uppercase tracking-[0.2em] text-white/20 hover:text-aura-glow transition-colors italic">
                EXPLORE FULL REGISTRY.
              </Link>
            </div>
          </div>

          {/* Activity Logs (Mock Feed) */}
          <div className="stat-card glow-cyan border-white/10 p-8 neural-bg relative overflow-hidden">
             <div className="flex items-center justify-between mb-8">
               <h3 className="text-xs font-black tracking-widest text-white/40 uppercase leading-none italic font-roboto-mono flex items-center gap-2">
                 <Terminal size={14} className="text-aura-glow" /> 
                 Live Audit Logs
               </h3>
               <div className="flex items-center gap-1.5">
                  <div className="w-1.5 h-1.5 rounded-full bg-aura-glow animate-pulse" />
                  <span className="text-[10px] font-bold text-aura-glow/60 uppercase">Live Feed</span>
               </div>
             </div>
             
             <div className="font-mono text-[10px] space-y-4 opacity-50">
               <div className="flex gap-3">
                  <span className="text-aura-glow">22:42:01</span>
                  <span>[AUTH] SuperUser_Node_01 Access Granted (JWT)</span>
               </div>
               <div className="flex gap-3">
                  <span className="text-aura-glow">22:41:58</span>
                  <span className="text-aura-purple">[CACHE] Semantic HIT: "analyze project aura"</span>
               </div>
               <div className="flex gap-3">
                  <span className="text-aura-glow">22:41:55</span>
                  <span className="text-aura-accent">[MCP] Executing postgres_tool (42ms)</span>
               </div>
               <div className="flex gap-3">
                  <span className="text-aura-glow">22:41:52</span>
                  <span>[ROUT] Best Match: local_db (Score: 0.99)</span>
               </div>
             </div>

             {/* Grain/Texture Overlay */}
             <div className="absolute inset-0 pointer-events-none opacity-[0.03] grayscale bg-[url('https://grainy-gradients.vercel.app/noise.svg')]" />
          </div>
        </section>
      </div>
    </div>
  );
}