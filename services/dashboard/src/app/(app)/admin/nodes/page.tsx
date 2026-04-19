import { getAuraStats } from "@/app/actions";
import { getCurrentOrg } from "@/lib/user-context";
import { MetricCard } from "@/components/metric-card";
import { 
  Activity, 
  Cpu, 
  Layers, 
  ShieldAlert,
  Terminal,
  Server
} from "lucide-react";
import { redirect } from "next/navigation";

export default async function AdminNodesPage() {
  const org = await getCurrentOrg();
  
  // RBAC Hardening
  if (org?.role !== 'platform_staff' && org?.role !== 'admin') {
     // For a strictly multi-tenant app, only platform_staff would see this.
     // In a personal setup, admin might also see it.
     // But following sidebar.tsx, we check platform_staff.
     if (org?.role !== 'platform_staff') {
        redirect("/");
     }
  }

  const stats = await getAuraStats(org?.orgId); // Scoped to admin org to prevent Forbidden crash

  const uptimeHours = Math.floor((stats.uptime_seconds || 0) / 3600);
  const uptimeMinutes = Math.floor(((stats.uptime_seconds || 0) % 3600) / 60);

  return (
    <div className="space-y-12 pb-20">
      <div className="flex items-center gap-4 mb-4">
        <div className="w-2 h-8 rounded-full bg-gradient-to-b from-aura-purple to-aura-accent" />
        <div>
          <h1 className="text-3xl font-black tracking-tighter uppercase">Infrastructure Nodes</h1>
          <p className="text-[10px] font-black text-white/20 uppercase tracking-[0.3em] italic">System-wide Health & Performance</p>
        </div>
      </div>

      <section className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <MetricCard 
          label="Gateway Uptime" 
          value={`${uptimeHours}h ${uptimeMinutes}m`} 
          trend="Steady State" 
          icon={<Activity size={24} />}
          color="purple"
          detail="Core Traffic Orchestrator"
        />
        <MetricCard 
          label="Intelligence Status" 
          value="Online" 
          trend="Optimized" 
          icon={<Cpu size={24} />}
          color="cyan"
          detail="Semantic Router & Qdrant"
        />
        <MetricCard 
          label="Global Throughput" 
          value={stats.total_requests?.toLocaleString() || "0"} 
          trend="Total Trace" 
          icon={<Layers size={24} />}
          color="purple"
          detail="Combined Network Flow"
        />
      </section>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div className="stat-card p-8 neural-bg border-white/5 space-y-6">
          <div className="flex items-center gap-4">
             <div className="w-12 h-12 rounded-2xl bg-aura-glow/10 border border-aura-glow/20 flex items-center justify-center text-aura-glow">
                <Server size={24} />
             </div>
             <h3 className="text-lg font-black tracking-tight uppercase">Node Inventory</h3>
          </div>
          <div className="space-y-4">
             <div className="flex items-center justify-between p-4 rounded-xl bg-white/[0.02] border border-white/5">
                <div className="flex items-center gap-3">
                   <Terminal size={14} className="text-aura-glow" />
                   <span className="text-xs font-bold text-white">aura-gateway</span>
                </div>
                <span className="text-[9px] font-black text-aura-accent px-2 py-1 rounded-md bg-aura-accent/10 border border-aura-accent/20">HEALTHY</span>
             </div>
             <div className="flex items-center justify-between p-4 rounded-xl bg-white/[0.02] border border-white/5">
                <div className="flex items-center gap-3">
                   <Terminal size={14} className="text-aura-purple" />
                   <span className="text-xs font-bold text-white">aura-router (rust)</span>
                </div>
                <span className="text-[9px] font-black text-aura-accent px-2 py-1 rounded-md bg-aura-accent/10 border border-aura-accent/20">HEALTHY</span>
             </div>
             <div className="flex items-center justify-between p-4 rounded-xl bg-white/[0.02] border border-white/5">
                <div className="flex items-center gap-3">
                   <Terminal size={14} className="text-aura-glow" />
                   <span className="text-xs font-bold text-white">aura-qdrant</span>
                </div>
                <span className="text-[9px] font-black text-aura-accent px-2 py-1 rounded-md bg-aura-accent/10 border border-aura-accent/20">HEALTHY</span>
             </div>
          </div>
        </div>

        <div className="stat-card p-8 neural-bg border-white/5 space-y-6">
           <div className="flex items-center gap-4">
              <div className="w-12 h-12 rounded-2xl bg-red-500/10 border border-red-500/20 flex items-center justify-center text-red-400">
                 <ShieldAlert size={24} />
              </div>
              <h3 className="text-lg font-black tracking-tight uppercase">Maintenance & Warnings</h3>
           </div>
           <div className="flex flex-col items-center justify-center p-12 text-center space-y-4">
              <div className="text-xs font-bold text-white/20 uppercase tracking-widest">No active system warnings</div>
              <div className="w-1.5 h-1.5 rounded-full bg-aura-accent animate-pulse" />
           </div>
        </div>
      </div>
    </div>
  );
}
