'use client'

import { useState, useEffect } from 'react'
import { Badge } from '@/components/ui/badge'
import { Activity, CheckCircle2, XCircle, Clock, TrendingDown, Zap, GitBranch } from 'lucide-react'
import { getWorkflows, approveWorkflow, rejectWorkflow, getOfflineStats } from '@/app/actions'
import { supabase } from '@/lib/supabase'

const statusConfig: Record<string, { color: string; icon: React.ReactNode }> = {
  discovered: { color: 'bg-blue-500/10 text-blue-400 border-blue-500/20', icon: <Activity size={14} /> },
  simulated: { color: 'bg-cyan-500/10 text-cyan-400 border-cyan-500/20', icon: <Zap size={14} /> },
  pending_review: { color: 'bg-amber-500/10 text-amber-400 border-amber-500/20', icon: <Clock size={14} /> },
  approved: { color: 'bg-green-500/10 text-green-400 border-green-500/20', icon: <CheckCircle2 size={14} /> },
  active: { color: 'bg-memzent-purple/10 text-memzent-purple border-memzent-purple/20', icon: <Zap size={14} /> },
  stale: { color: 'bg-orange-500/10 text-orange-400 border-orange-500/20', icon: <TrendingDown size={14} /> },
  demoted: { color: 'bg-red-500/10 text-red-400 border-red-500/20', icon: <XCircle size={14} /> },
}

export default function WorkflowsPage() {
  const [workflows, setWorkflows] = useState<any[]>([])
  const [offlineStats, setOfflineStats] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState<string>('')
  const [orgId, setOrgId] = useState<string>('')

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
        if (membership?.organizations) {
          const org = membership.organizations as any
          resolvedOrgId = org.id
        }
        setOrgId(resolvedOrgId)

        const [wf, stats] = await Promise.all([
          getWorkflows(resolvedOrgId, filter || undefined),
          getOfflineStats(resolvedOrgId),
        ])
        setWorkflows(wf || [])
        setOfflineStats(stats)
      }
      setLoading(false)
    }
    load()
  }, [filter])

  async function handleApprove(id: string) {
    await approveWorkflow(orgId, id)
    setWorkflows(prev => prev.map(w => w.id === id ? { ...w, status: 'active' } : w))
  }

  async function handleReject(id: string) {
    await rejectWorkflow(orgId, id)
    setWorkflows(prev => prev.map(w => w.id === id ? { ...w, status: 'demoted' } : w))
  }

  const activeCount = workflows.filter(w => w.status === 'active').length
  const pendingCount = workflows.filter(w => ['discovered', 'pending_review', 'simulated'].includes(w.status)).length

  return (
    <div className="space-y-8 max-w-6xl mx-auto">
      <header>
        <h1 className="text-2xl font-black uppercase tracking-tight text-white italic flex items-center gap-3">
          <GitBranch className="text-memzent-purple" />
          Workflow Registry
        </h1>
        <p className="text-xs text-white/30 font-bold uppercase tracking-widest mt-1">
          Evolution Phase E4 — Deterministic Hot Path Promotion
        </p>
      </header>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="stat-card p-5">
          <div className="text-[10px] font-black text-white/40 uppercase tracking-widest">Active Workflows</div>
          <div className="text-3xl font-black text-memzent-purple mt-1">{activeCount}</div>
        </div>
        <div className="stat-card p-5">
          <div className="text-[10px] font-black text-white/40 uppercase tracking-widest">Pending Review</div>
          <div className="text-3xl font-black text-amber-400 mt-1">{pendingCount}</div>
        </div>
        <div className="stat-card p-5">
          <div className="text-[10px] font-black text-white/40 uppercase tracking-widest">Offline Events</div>
          <div className="text-3xl font-black text-cyan-400 mt-1">
            {offlineStats?.plane?.processed?.toLocaleString() ?? '—'}
          </div>
        </div>
        <div className="stat-card p-5">
          <div className="text-[10px] font-black text-white/40 uppercase tracking-widest">Prediction Accuracy</div>
          <div className="text-3xl font-black text-green-400 mt-1">
            {offlineStats?.prediction_accuracy != null && offlineStats.prediction_accuracy >= 0
              ? `${(offlineStats.prediction_accuracy * 100).toFixed(1)}%`
              : '—'}
          </div>
        </div>
      </div>

      {/* Filter Tabs */}
      <div className="flex gap-2 flex-wrap">
        {['', 'active', 'pending_review', 'discovered', 'stale', 'demoted'].map(s => (
          <button
            key={s}
            onClick={() => setFilter(s)}
            className={`px-3 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-widest transition-all ${
              filter === s
                ? 'bg-memzent-purple/20 text-memzent-purple border border-memzent-purple/30'
                : 'bg-white/5 text-white/40 border border-white/10 hover:bg-white/10'
            }`}
          >
            {s || 'All'}
          </button>
        ))}
      </div>

      {/* Workflow List */}
      <div className="stat-card divide-y divide-white/5">
        {loading ? (
          <div className="py-20 text-center text-white/10 font-black italic uppercase tracking-[0.4em] text-sm">
            Mining Workflows...
          </div>
        ) : workflows.length === 0 ? (
          <div className="py-20 text-center text-white/10 font-black italic uppercase tracking-[0.4em] text-sm">
            No Workflow Candidates Found
          </div>
        ) : (
          workflows.map((wf) => {
            const cfg = statusConfig[wf.status] || statusConfig.discovered
            return (
              <div key={wf.id} className="p-6 hover:bg-white/[0.02] transition-all group">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-3 mb-2">
                      <Badge variant="outline" className={`${cfg.color} text-[9px] uppercase font-black tracking-widest px-2 flex items-center gap-1`}>
                        {cfg.icon} {wf.status}
                      </Badge>
                      <span className="text-[10px] font-mono text-white/20">
                        freq: {wf.frequency?.toLocaleString()}
                      </span>
                      {wf.accuracy_7d != null && (
                        <span className={`text-[10px] font-mono ${wf.accuracy_7d >= 0.9 ? 'text-green-400/60' : wf.accuracy_7d >= 0.85 ? 'text-amber-400/60' : 'text-red-400/60'}`}>
                          acc: {(wf.accuracy_7d * 100).toFixed(1)}%
                        </span>
                      )}
                    </div>

                    <div className="text-sm font-black tracking-tight text-white uppercase italic">
                      {wf.pattern}
                    </div>

                    <div className="flex flex-wrap gap-1.5 mt-2">
                      {(wf.tool_ids || []).map((tool: string, i: number) => (
                        <span key={i} className="inline-flex items-center px-2 py-0.5 rounded-md bg-white/5 border border-white/10 text-[9px] font-mono font-bold text-white/50">
                          {i > 0 && <span className="text-memzent-purple/40 mr-1">→</span>}
                          {tool}
                        </span>
                      ))}
                    </div>

                    {wf.tokens_saved > 0 && (
                      <div className="text-[10px] text-green-400/50 font-mono mt-2">
                        💰 {wf.tokens_saved.toLocaleString()} tokens saved
                      </div>
                    )}
                  </div>

                  <div className="flex items-center gap-2 shrink-0">
                    {['discovered', 'simulated', 'pending_review'].includes(wf.status) && (
                      <>
                        <button
                          onClick={() => handleApprove(wf.id)}
                          className="px-3 py-1.5 rounded-lg bg-green-500/10 border border-green-500/20 text-green-400 text-[10px] font-black uppercase tracking-widest hover:bg-green-500/20 transition-all"
                        >
                          Approve
                        </button>
                        <button
                          onClick={() => handleReject(wf.id)}
                          className="px-3 py-1.5 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-[10px] font-black uppercase tracking-widest hover:bg-red-500/20 transition-all"
                        >
                          Reject
                        </button>
                      </>
                    )}
                    {wf.status === 'active' && (
                      <button
                        onClick={() => handleReject(wf.id)}
                        className="px-3 py-1.5 rounded-lg bg-orange-500/10 border border-orange-500/20 text-orange-400 text-[10px] font-black uppercase tracking-widest hover:bg-orange-500/20 transition-all"
                      >
                        Demote
                      </button>
                    )}
                  </div>
                </div>
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}
