'use client'

import { useState, useEffect } from 'react'
import { Badge } from '@/components/ui/badge'
import { Activity, ShieldCheck, Zap, Server, ShieldAlert, KeyRound, Bot, Globe } from 'lucide-react'
import { getMemzentAudit, getMemzentStats } from '@/app/actions'
import { supabase } from '@/lib/supabase'

function formatTimestampUTC(isoString: string) {
  try {
    const clean = isoString.replace('T', ' ').split('.')[0] || isoString
    return clean.endsWith('Z') ? clean.slice(0, -1) : clean
  } catch {
    return isoString
  }
}

export default function AuditPage() {
  const [logs, setLogs] = useState<any[]>([])
  const [stats, setStats] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [orgName, setOrgName] = useState<string>('')

  useEffect(() => {
    async function load() {
      const { data: { user } } = await supabase.auth.getUser()
      if (user) {
        const { data: membership } = await supabase
          .from('members')
          .select('org_id, organizations(id, name)')
          .eq('user_id', user.id)
          .limit(1)
          .maybeSingle()

        let resolvedOrgId = user.id
        let resolvedOrgName = user.email?.split('@')[0] || 'Personal'

        if (membership?.organizations) {
          const org = membership.organizations as any
          resolvedOrgId = org.id
          resolvedOrgName = org.name
        }

        setOrgName(resolvedOrgName)

        try {
          const [auditData, statsData] = await Promise.all([
            getMemzentAudit(resolvedOrgId),
            getMemzentStats(resolvedOrgId)
          ])
          setLogs(auditData || [])
          setStats(statsData)
        } catch {
          // Ignore fetch errors
        }
      }
      setLoading(false)
    }
    load()
  }, [])

  const hitRatio = stats && stats.total_requests > 0
    ? ((stats.cache_hits / stats.total_requests) * 100).toFixed(1)
    : '0.0'

  const getIcon = (type: string) => {
    switch (type) {
      case 'AUTH': return <ShieldCheck size={20} className="text-blue-400" />
      case 'CACHE': return <Zap size={20} className="text-yellow-400" />
      case 'GENERATION': return <Bot size={20} className="text-memzent-purple" />
      case 'GATEWAY': return <Server size={20} className="text-green-400" />
      case 'KEY_GEN': return <KeyRound size={20} className="text-memzent-glow" />
      default: return <Globe size={20} className="text-white/40" />
    }
  }

  return (
    <div className="space-y-12">
      <header className="mb-12">
        <h1 className="text-4xl font-black tracking-tighter text-white mb-2 uppercase italic">
          SECERN_AUDIT
        </h1>
        <p className="text-white/20 font-black uppercase tracking-[0.3em] text-[10px] italic">
          {orgName ? `${orgName} — ` : ''}Network Observability & Security Ledger
        </p>
      </header>

      {/* KPI Row */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="stat-card neural-bg p-8 border-white/5 hover:border-memzent-glow/20 transition-all">
          <div className="text-[10px] font-black uppercase tracking-widest text-white/40 mb-2 whitespace-nowrap">Total Mesh Requests</div>
          <div className="text-3xl font-black text-white">{stats?.total_requests?.toLocaleString() || 0}</div>
        </div>
        <div className="stat-card p-8 border-memzent-glow/20 bg-memzent-glow/5 shadow-[0_0_20px_rgba(0,243,255,0.05)] relative overflow-hidden transition-all">
          <div className="absolute top-0 right-0 p-8 text-memzent-glow/10 pointer-events-none">
            <Zap size={64} />
          </div>
          <div className="text-[10px] font-black uppercase tracking-widest text-memzent-glow mb-2 whitespace-nowrap">Cache Intelligence Ratio</div>
          <div className="text-3xl font-black text-white flex items-end gap-2">
            {hitRatio}% <span className="text-sm font-bold text-white/40 uppercase tracking-widest mb-1 italic">hit_rate</span>
          </div>
        </div>
        <div className="stat-card neural-bg p-8 border-white/5 hover:border-memzent-purple/20 transition-all">
          <div className="text-[10px] font-black uppercase tracking-widest text-white/40 mb-2 whitespace-nowrap">Node Uptime</div>
          <div className="text-3xl font-black text-white flex items-end gap-2">
            {(stats?.uptime_seconds / 3600).toFixed(1)} <span className="text-sm font-bold text-white/40 uppercase tracking-widest mb-1 italic">hours</span>
          </div>
        </div>
      </div>

      <div className="stat-card neural-bg border-white/5 p-0 overflow-hidden">
        <div className="p-8 border-b border-white/5 flex flex-col md:flex-row items-center justify-between gap-6 bg-black/20">
          <div>
            <h3 className="text-lg font-black tracking-tight text-white uppercase italic">Event Ledger</h3>
            <p className="text-[10px] font-bold text-white/20 uppercase tracking-widest mt-1">Real-time organizational footprint</p>
          </div>
          <Badge variant="outline" className="border-green-500/20 bg-green-500/5 text-green-400 font-black tracking-widest uppercase text-[9px] px-3">
            <Activity size={10} className="mr-2 animate-pulse" /> Live
          </Badge>
        </div>

        <div className="divide-y divide-white/5">
          {loading ? (
            <div className="py-20 text-center text-white/10 font-black italic uppercase tracking-[0.4em] text-sm">Intercepting Logs...</div>
          ) : logs.length === 0 ? (
            <div className="py-20 text-center text-white/10 font-black italic uppercase tracking-[0.4em] text-sm">No Auditable Events Found</div>
          ) : (
            logs.map((log, index) => (
              <div key={index} className="flex items-start md:items-center justify-between p-6 hover:bg-white/[0.02] transition-all group flex-col md:flex-row gap-4">
                <div className="flex items-start md:items-center gap-6">
                  <div className="w-12 h-12 shrink-0 rounded-2xl bg-white/5 border border-white/10 flex items-center justify-center shadow-inner group-hover:scale-110 transition-transform">
                    {getIcon(log.type)}
                  </div>
                  <div>
                    <div className="text-sm font-black tracking-tight text-white uppercase italic truncate max-w-sm md:max-w-2xl">{log.detail}</div>
                    <div className="text-[10px] font-mono text-white/20 uppercase font-black flex items-center gap-3 mt-1.5 flex-wrap">
                      <span className="text-white/40">ACTOR:</span> <span className={`${log.user === 'system' ? 'text-memzent-purple' : 'text-memzent-glow'}`}>{log.user || 'Unknown'}</span>
                      <span className="hidden md:inline-block w-1 h-1 rounded-full bg-white/10" />
                      <span className="text-white/40">TYPE:</span> {log.type}
                    </div>
                    {log.entities && Object.keys(log.entities).length > 0 && (
                      <div className="flex flex-wrap gap-1.5 mt-2">
                        {Object.entries(log.entities).map(([key, value]) => (
                          <span key={key} className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-memzent-purple/10 border border-memzent-purple/20 text-[9px] font-mono font-bold text-memzent-purple/80">
                            <span className="text-white/30">{key}:</span>{String(value)}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
                <div className="flex flex-col items-end gap-2 shrink-0">
                  <div className="text-[10px] font-black tracking-widest text-white/40 uppercase font-mono">
                    {formatTimestampUTC(log.timestamp)}
                  </div>
                  <Badge variant="outline" className={`border-white/10 text-[9px] uppercase font-black tracking-widest px-3 ${log.status === 'success' ? 'bg-green-500/10 text-green-400 border-green-500/20' : 'bg-red-500/10 text-red-500 border-red-500/20'}`}>
                    {log.status}
                  </Badge>
                </div>
              </div>
            ))
          )}
        </div>
      </div>

      <footer className="stat-card border-white/5 bg-black/20 p-8 relative overflow-hidden group">
        <div className="absolute top-0 right-0 p-8 text-white/5 pointer-events-none">
          <ShieldAlert size={120} />
        </div>
        <h3 className="text-xs font-black text-white/40 uppercase tracking-[0.3em] mb-4 italic">Compliance Directive</h3>
        <p className="text-[10px] text-white/20 leading-relaxed font-black uppercase max-w-2xl tracking-widest">
          The Semantic Proxy retains event footprints for 30 days per Enterprise requirements. Audit logs cover routing, authentication, generation, and vector caching mechanisms natively.
        </p>
      </footer>
    </div>
  )
}
