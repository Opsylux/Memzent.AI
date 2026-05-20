import { getMemzentTools } from "../../actions";
import { getCurrentOrg } from "@/lib/user-context";
import {
  Database,
  Search,
  Filter,
  RefreshCcw,
  MoreHorizontal,
  ExternalLink,
  ShieldAlert
} from "lucide-react";
import { RegisterToolBtn } from "@/components/register-tool-btn";
import { SyncRegistryBtn } from "@/components/sync-registry-btn";
import { ClientStubButton } from "@/components/client-stub-button";

export default async function ToolsPage() {
  const org = await getCurrentOrg();
  const tools = await getMemzentTools(org?.orgId);

  const providerCount = new Set((tools || []).map((tool: any) => tool.provider || 'Memzent_Internal')).size;

  return (
    <div className="space-y-12">
      <header className="flex flex-col md:flex-row justify-between items-start md:items-end gap-6 mb-12">
        <div>
          <h1 className="text-4xl font-black tracking-tighter">TOOL_REGISTRY</h1>
          <p className="text-sm font-bold text-white/50 uppercase tracking-widest mt-1">
            {org?.orgName ? `${org.orgName} — ` : ''}Managed Model Context Protocol Explorer
          </p>
        </div>
        <div className="flex items-center gap-4 w-full md:w-auto">
          <SyncRegistryBtn orgId={org?.orgId} />
          <RegisterToolBtn orgId={org?.orgId} />
        </div>
      </header>

      {/* Stats Bar */}
      <section className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="stat-card border-memzent-glow/10 shadow-[0_0_20px_rgba(0,243,255,0.02)]">
          <div className="text-[10px] font-black text-white/50 uppercase mb-1">Total Discovered</div>
          <div className="text-3xl font-black">{(tools || []).length}</div>
        </div>
        <div className="stat-card border-memzent-purple/10">
          <div className="text-[10px] font-black text-white/50 uppercase mb-1">Active Providers</div>
          <div className="text-3xl font-black">{providerCount}</div>
        </div>
        <div className="stat-card border-memzent-accent/10">
          <div className="text-[10px] font-black text-white/50 uppercase mb-1">Health Status</div>
          <div className="text-3xl font-black text-memzent-accent">EXCELLENT</div>
        </div>
      </section>

      {/* Tool Explorer */}
      <section className="stat-card neural-bg border-white/5 p-0 overflow-hidden">
        <div className="p-8 border-b border-white/5 flex flex-col md:flex-row justify-between gap-6">
          <div className="relative group flex-1 max-w-md">
            <Search size={16} className="absolute left-4 top-1/2 -translate-y-1/2 text-white/20 group-hover:text-memzent-glow transition-colors" />
            <input
              type="text"
              placeholder="Filter Registry..."
              className="w-full h-12 bg-white/[0.03] border border-white/5 rounded-xl pl-12 pr-4 text-xs font-bold tracking-tight focus:outline-none focus:border-memzent-glow/20 transition-all"
            />
          </div>
          <div className="flex items-center gap-2">
            <button className="glass h-12 px-4 rounded-xl flex items-center gap-2 text-[10px] font-black uppercase text-white/60 hover:text-white transition-all">
              <Filter size={14} /> Providers
            </button>
            <button className="glass h-12 px-4 rounded-xl flex items-center gap-2 text-[10px] font-black uppercase text-white/60 hover:text-white transition-all">
              <ShieldAlert size={14} /> Unsafe Tools
            </button>
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-white/[0.02] border-b border-white/5 text-[10px] uppercase font-black tracking-widest text-white/60">
                <th className="px-8 py-6">IDENTIFIER / NAME</th>
                <th className="px-8 py-6">PROVIDER TYPE</th>
                <th className="px-8 py-6">STATUS</th>
                <th className="px-8 py-6">REL_SCORE</th>
                <th className="px-8 py-6 text-right">ACTIONS</th>
              </tr>
            </thead>
            <tbody>
              {(tools || []).map((tool: any) => (
                <tr key={tool.id} className="border-b border-white/5 hover:bg-white/[0.03] transition-all group">
                  <td className="px-8 py-6">
                    <div className="flex items-center gap-4">
                      <div className="w-10 h-10 rounded-lg bg-white/5 border border-white/10 flex items-center justify-center text-white/40 group-hover:text-memzent-glow transition-colors">
                        <Database size={18} />
                      </div>
                      <div>
                        <div className="text-sm font-black tracking-tight group-hover:text-white transition-colors">{tool.name}</div>
                        <div className="text-[10px] font-mono text-white/45 uppercase truncate w-64">{tool.id}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-8 py-6">
                    <span className="text-[10px] font-black px-2 py-1 rounded-md bg-white/5 border border-white/5 uppercase">
                      {tool.provider || 'Memzent_Internal'}
                    </span>
                  </td>
                  <td className="px-8 py-6 text-[10px] font-bold">
                    <div className="flex items-center gap-2">
                      <div className="w-1.5 h-1.5 rounded-full bg-memzent-accent shadow-[0_0_8px_#00ff8e] animate-pulse" />
                      <span className="text-memzent-accent/90 uppercase">ONLINE</span>
                    </div>
                  </td>
                  <td className="px-8 py-6 font-mono text-xs font-black text-white/65 italic">
                    0.992
                  </td>
                  <td className="px-8 py-6 text-right">
                    <div className="flex items-center justify-end gap-2 opacity-20 group-hover:opacity-100 transition-opacity">
                      <ClientStubButton type="external" message="Tool Configuration editing scheduled for Phase 3." />
                      <ClientStubButton type="more" message="Additional tool actions pending." />
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </div>
  );
}
