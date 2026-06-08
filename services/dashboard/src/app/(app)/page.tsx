import { getMemzentTools, getMemzentStats, getOrgKeyCount, getOrgAuditStats } from "../actions";
import { getCurrentOrg } from "@/lib/user-context";
import { MetricCard } from "@/components/metric-card";
import { QuickActions } from "@/components/quick-actions";
import { CacheRing } from "@/components/cache-ring";
import {
  Zap,
  Activity,
  ShieldCheck,
  Layers,
  ArrowUpRight,
  ArrowRight,
  Database,
  Cpu,
  Terminal,
  TrendingUp
} from "lucide-react";
import { RoutingVisualizer } from "@/components/routing-visualizer";
import { AuditLogFeed } from "@/components/audit-log-feed";
import Link from 'next/link';

export default async function Page() {
  const org = await getCurrentOrg();
  const orgId = org?.orgId;

  const [initialTools, stats, keyCount, auditStats] = await Promise.all([
    getMemzentTools(orgId),
    getMemzentStats(orgId),
    orgId ? getOrgKeyCount(orgId) : Promise.resolve(0),
    orgId ? getOrgAuditStats(orgId) : Promise.resolve({ count24h: 0 })
  ]);

  const total = stats.total_requests || 0;
  const hits = stats.cache_hits || 0;
  const cacheHitPercent = total > 0 ? (hits / total) * 100 : 0;
  const semanticSavings = cacheHitPercent.toFixed(1);
  const uptimeHours = Math.floor((stats.uptime_seconds || 0) / 3600);
  const providerCount = stats.provider_count || 0;
  const defaultProvider = stats.default_provider || "Ollama";
  const activeProviders = Array.isArray(stats.active_providers) ? stats.active_providers : [];

  return (
    <div className="space-y-10 pb-20">
      <header className="flex flex-col md:flex-row md:items-end md:justify-between gap-4">
        <div className="flex items-center gap-4">
          <div className="w-1.5 h-10 rounded-full bg-gradient-to-b from-memzent-glow to-memzent-purple" />
          <div>
            <h1 className="text-2xl md:text-3xl font-black tracking-tight text-readable-primary">
              {org?.orgName || "Dashboard"}
            </h1>
            <p className="text-sm text-readable-muted mt-1">
              Semantic proxy · cache · routing for your agents
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2 text-xs font-medium text-readable-muted">
          <span className="w-2 h-2 rounded-full bg-memzent-accent animate-pulse" />
          Gateway {uptimeHours > 0 ? `${uptimeHours}h uptime` : "online"} · {defaultProvider} default
        </div>
      </header>

      <QuickActions />

      {/* KPI Section */}
      <section className="grid grid-cols-1 md:grid-cols-3 xl:grid-cols-5 gap-6">
        <MetricCard
          label="Cache hit rate"
          value={`${semanticSavings}%`}
          trend={hits > 0 ? "Saving tokens" : "No hits yet"}
          trendDirection={hits > 0 ? "up" : "down"}
          icon={<Zap size={24} />}
          color="cyan"
          detail={`${hits.toLocaleString()} of ${total.toLocaleString()} requests`}
          ringSlot={<CacheRing percent={cacheHitPercent} />}
        />
        <MetricCard
          label="Total Engine Throughput"
          value={total.toLocaleString()}
          trend="Requests"
          icon={<Layers size={24} />}
          color="purple"
          detail="Organization-scoped Flow"
        />
        <MetricCard
          label="Memzent Tools Registry"
          value={`${initialTools?.length || 0}`}
          trend="Online"
          icon={<ShieldCheck size={24} />}
          color="cyan"
          detail="Active Context Bindings"
        />
        <MetricCard
          label="Token Balance"
          value={`$${(stats.token_balance || 0).toFixed(4)}`}
          trend="PAYG"
          trendDirection={(stats.token_balance || 0) > 5 ? "up" : "down"}
          icon={<Database size={24} />}
          color="cyan"
          detail="Available Funds"
        />
        <MetricCard
          label="API Security"
          value={`${keyCount}`}
          trend={keyCount > 0 ? "Protected" : "Setup Required"}
          trendDirection={keyCount > 0 ? "up" : "down"}
          icon={<ShieldCheck size={24} />}
          color="purple"
          detail={`${auditStats.count24h} Events (24h)`}
        />
      </section>

      {/* Main Intelligence Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Semantic Flow & Visualizer */}
        <section className="lg:col-span-2 space-y-8">
          <div className="flex items-center justify-between px-2">
            <div>
              <h2 className="text-xl font-black tracking-tight text-readable-primary">Live routing test</h2>
              <p className="text-sm text-readable-muted mt-1">Send a prompt and watch tool matching in real time</p>
            </div>
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-memzent-glow/5 border border-memzent-glow/20">
                <div className="w-1.5 h-1.5 rounded-full bg-memzent-glow animate-pulse" />
                <span className="text-[10px] font-black text-memzent-glow uppercase tracking-tighter">Live Monitor</span>
              </div>
            </div>
          </div>

          <div className="glass rounded-3xl overflow-hidden border-white/5 shadow-2xl relative">
            <div className="absolute top-0 right-0 p-4 opacity-10 pointer-events-none">
              <TrendingUp size={120} className="text-memzent-glow" />
            </div>
            <RoutingVisualizer steps={[]} orgId={orgId} />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="stat-card glow-purple p-8 neural-bg border-white/5 flex flex-col justify-between min-h-[300px]">
              <div className="space-y-6">
                <div className="flex items-center gap-4">
                  <div className="w-12 h-12 rounded-2xl bg-memzent-purple/10 border border-memzent-purple/20 flex items-center justify-center text-memzent-purple shadow-[0_0_15px_rgba(157,0,255,0.2)]">
                    <Cpu size={24} />
                  </div>
                  <div>
                    <h3 className="text-lg font-black tracking-tight leading-none uppercase">Intelligence Status</h3>
                    <p className="text-[10px] font-bold text-white/20 uppercase tracking-widest mt-1 italic">Model Discovery & Routing</p>
                  </div>
                </div>
                <p className="text-sm text-readable-secondary leading-relaxed">
                  <span className="text-readable-primary font-bold">{initialTools?.length || 0} tools</span> registered
                  across <span className="text-readable-primary font-bold">{providerCount} LLM providers</span>.
                  Router uses vector similarity (threshold 0.65) to match prompts to tools before synthesis.
                </p>
              </div>

              <div className="grid grid-cols-2 gap-4 mt-8 pt-6 border-t border-white/5">
                <div className="p-4 rounded-2xl bg-white/[0.02] border border-white/5">
                  <div className="text-[11px] font-medium text-readable-muted mb-1">Vector engine</div>
                  <div className="text-xs font-bold text-memzent-glow truncate">Rust + Qdrant</div>
                </div>
                <div className="p-4 rounded-2xl bg-white/[0.02] border border-white/5">
                  <div className="text-[11px] font-medium text-readable-muted mb-1">Default provider</div>
                  <div className="text-xs font-bold text-memzent-purple truncate">{defaultProvider}</div>
                </div>
              </div>
            </div>

            <div className="stat-card glow-cyan p-8 neural-bg border-white/5 flex flex-col justify-between min-h-[300px]">
              <div className="space-y-6">
                <div className="flex items-center gap-4">
                  <div className="w-12 h-12 rounded-2xl bg-memzent-glow/10 border border-memzent-glow/20 flex items-center justify-center text-memzent-glow shadow-[0_0_15px_rgba(0,243,255,0.2)]">
                    <Zap size={24} />
                  </div>
                  <div>
                    <h3 className="text-lg font-black tracking-tight leading-none uppercase">Active Clusters</h3>
                    <p className="text-[10px] font-bold text-white/20 uppercase tracking-widest mt-1 italic">Multi-Provider Inventory</p>
                  </div>
                </div>
                <div className="flex flex-wrap gap-2">
                  {activeProviders.length > 0 ? activeProviders.map((provider: string) => (
                    <span key={provider} className="text-[10px] px-3 py-1.5 rounded-xl bg-white/5 border border-white/10 uppercase tracking-widest text-slate-200 font-black hover:border-memzent-glow/30 transition-all cursor-crosshair">
                      {provider}
                    </span>
                  )) : (
                    <span className="text-[10px] text-white/20 italic tracking-widest uppercase">No Active Providers Connected</span>
                  )}
                </div>
              </div>

              <div className="mt-8 pt-6 border-t border-white/5 flex items-center justify-between">
                <div className="space-y-1">
                  <div className="text-[10px] font-black text-white/20 uppercase">Routing Status</div>
                  <div className="text-sm font-black text-memzent-accent uppercase tracking-tighter italic">Engine Ready for Synthesis</div>
                </div>
                <div className="w-2.5 h-2.5 rounded-full bg-memzent-accent shadow-[0_0_8px_#00ff8e]" />
              </div>
            </div>
          </div>
        </section>

        {/* Sidebar Analytics */}
        <section className="space-y-8">
          {/* Active Tools Mini Feed */}
          <div className="stat-card border-white/10 p-8 neural-bg relative overflow-hidden group">
            <div className="flex items-center justify-between mb-8">
              <h3 className="text-xs font-black tracking-widest text-white/40 uppercase leading-none italic font-roboto-mono flex items-center gap-2">
                <Database size={14} className="text-memzent-glow" />
                Recent Registry Updates
              </h3>
              <ArrowUpRight size={14} className="text-white/20 group-hover:text-memzent-glow transition-colors" />
            </div>
            <div className="space-y-6">
              {(initialTools || []).slice(0, 5).map((tool: any) => (
                <div key={tool.id} className="flex items-center justify-between group/item cursor-pointer">
                  <div className="flex items-center gap-4">
                    <div className="w-1.5 h-1.5 rounded-full bg-memzent-accent shadow-[0_0_8px_#00ff8e]" />
                    <div>
                      <div className="text-sm font-black text-white group-hover/item:text-memzent-glow transition-colors">{tool.name}</div>
                      <div className="text-[10px] font-mono text-white/20 uppercase tracking-tighter truncate w-32 font-bold">{tool.connector_type || 'MCP'}</div>
                    </div>
                  </div>
                  <div className="text-[9px] font-black text-white/30 px-2 py-1 rounded-lg border border-white/5 group-hover/item:border-memzent-glow/20 transition-all uppercase tracking-widest">
                    ACTIVE
                  </div>
                </div>
              ))}
              <Link href="/tools" className="flex items-center justify-center gap-2 w-full py-4 text-[10px] font-black uppercase tracking-[0.25em] text-white/20 hover:text-memzent-glow transition-all italic border-t border-white/5 group/link">
                EXPLORE FULL REGISTRY
                <ArrowRight size={12} className="group-hover/link:translate-x-1 transition-transform" />
              </Link>
            </div>
            {/* Grain Overlay */}
            <div className="absolute inset-0 pointer-events-none opacity-[0.03] grayscale bg-[url('https://grainy-gradients.vercel.app/noise.svg')]" />
          </div>

          {/* Activity Logs (Real Feed) */}
          <AuditLogFeed orgId={orgId} />
        </section>
      </div>
    </div>
  );
}