'use client'

import { useEffect, useState } from 'react'
import {
  Cpu, Zap, BarChart3, Shield, Activity, RefreshCw, Loader2,
  TrendingUp, TrendingDown, Gauge, Layers, Brain, CheckCircle2
} from 'lucide-react'
import { getEntityMetrics, getOfflineStats, getFeatureFlags } from '@/app/actions'
import { supabase } from '@/lib/supabase'

interface EntityMetrics {
  regex_success: number
  regex_failure: number
  regex_success_rate: number
  entity_mismatch: number
  llm_entity_usage: number
  gpu_avoidance_rate: number
  gpu_avoided: number
  gpu_invoked: number
}

interface OfflineStats {
  enabled: boolean
  mode?: string
  plane?: { emitted: number; processed: number; dropped: number }
  hot_patterns?: Array<{ pattern: string; frequency: number }>
  workflow_sequences?: Array<{ pattern: string; frequency: number; success_rate: number }>
}

interface FeatureFlags {
  l1b_cache: boolean
  offline_plane: boolean
  offline_streams: boolean
  workflow_engine: boolean
  entity_metrics: boolean
}

export default function GPUAnalyticsPage() {
  const [metrics, setMetrics] = useState<EntityMetrics | null>(null)
  const [offline, setOffline] = useState<OfflineStats | null>(null)
  const [flags, setFlags] = useState<FeatureFlags | null>(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)

  const fetchAll = async (isRefresh = false) => {
    if (isRefresh) setRefreshing(true)
    else setLoading(true)

    try {
      const { data: { user } } = await supabase.auth.getUser()
      const { data: membership } = await supabase
        .from('members')
        .select('org_id')
        .eq('user_id', user?.id ?? '')
        .limit(1)
        .maybeSingle()

      const orgId = membership?.org_id || user?.id
      const [m, o, f] = await Promise.all([
        getEntityMetrics(orgId),
        getOfflineStats(orgId),
        getFeatureFlags(orgId),
      ])
      setMetrics(m)
      setOffline(o)
      setFlags(f)
    } catch (err) {
      console.error('GPU analytics fetch failed', err)
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => { fetchAll() }, [])

  const gpuRate = metrics?.gpu_avoidance_rate ?? 0
  const totalRequests = (metrics?.gpu_avoided ?? 0) + (metrics?.gpu_invoked ?? 0)

  return (
    <div className="space-y-8 pb-20">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className="w-2 h-8 rounded-full bg-gradient-to-b from-emerald-400 to-cyan-500" />
          <div>
            <h1 className="text-3xl font-black tracking-tighter uppercase">GPU Analytics</h1>
            <p className="text-[10px] font-black text-white/20 uppercase tracking-[0.3em] italic">
              GPU Avoidance · Cache Layers · Entity Quality · Evolution Status
            </p>
          </div>
        </div>
        <button
          onClick={() => fetchAll(true)}
          disabled={refreshing}
          className="flex items-center gap-2 bg-white/5 border border-white/10 text-white/60 font-black px-4 py-2.5 rounded-xl text-[10px] uppercase tracking-[0.2em] hover:border-emerald-400/30 hover:text-emerald-400 transition-all disabled:opacity-30"
        >
          {refreshing ? <Loader2 size={12} className="animate-spin" /> : <RefreshCw size={12} />}
          Refresh
        </button>
      </div>

      {loading ? (
        <div className="flex flex-col items-center justify-center py-32">
          <Loader2 size={32} className="text-emerald-400/30 animate-spin mb-4" />
          <p className="text-[10px] font-black uppercase tracking-[0.3em] text-emerald-400/30 animate-pulse">
            Loading GPU Metrics...
          </p>
        </div>
      ) : (
        <>
          {/* GPU Avoidance Hero Card */}
          <div className="stat-card neural-bg border-white/5 p-8 relative overflow-hidden">
            <div className="absolute inset-0 bg-gradient-to-br from-emerald-500/5 to-transparent" />
            <div className="relative flex items-center justify-between">
              <div>
                <p className="text-[9px] font-black uppercase tracking-[0.3em] text-white/30 mb-2">
                  GPU Avoidance Rate
                </p>
                <p className="text-6xl font-black text-emerald-400 tracking-tight">
                  {(gpuRate * 100).toFixed(1)}%
                </p>
                <p className="text-[10px] text-white/40 mt-2">
                  {metrics?.gpu_avoided?.toLocaleString() ?? 0} requests served without LLM
                  {totalRequests > 0 && ` of ${totalRequests.toLocaleString()} total`}
                </p>
              </div>
              <div className="w-24 h-24 rounded-full border-4 border-emerald-400/20 flex items-center justify-center">
                <Cpu size={36} className="text-emerald-400/60" />
              </div>
            </div>
          </div>

          {/* KPI Cards */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* GPU Avoided */}
            <div className="stat-card neural-bg border-white/5 p-6 relative overflow-hidden group">
              <div className="absolute inset-0 bg-gradient-to-br from-emerald-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
              <div className="relative">
                <div className="flex items-center justify-between mb-4">
                  <div className="w-10 h-10 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center">
                    <TrendingDown size={18} className="text-emerald-400" />
                  </div>
                </div>
                <p className="text-[9px] font-black uppercase tracking-[0.2em] text-white/30 mb-1">GPU Avoided</p>
                <p className="text-3xl font-black text-white tracking-tight">{(metrics?.gpu_avoided ?? 0).toLocaleString()}</p>
              </div>
            </div>

            {/* GPU Invoked */}
            <div className="stat-card neural-bg border-white/5 p-6 relative overflow-hidden group">
              <div className="absolute inset-0 bg-gradient-to-br from-orange-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
              <div className="relative">
                <div className="flex items-center justify-between mb-4">
                  <div className="w-10 h-10 rounded-xl bg-orange-500/10 border border-orange-500/20 flex items-center justify-center">
                    <TrendingUp size={18} className="text-orange-400" />
                  </div>
                </div>
                <p className="text-[9px] font-black uppercase tracking-[0.2em] text-white/30 mb-1">GPU Invoked</p>
                <p className="text-3xl font-black text-white tracking-tight">{(metrics?.gpu_invoked ?? 0).toLocaleString()}</p>
              </div>
            </div>

            {/* Entity Regex Success Rate */}
            <div className="stat-card neural-bg border-white/5 p-6 relative overflow-hidden group">
              <div className="absolute inset-0 bg-gradient-to-br from-cyan-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
              <div className="relative">
                <div className="flex items-center justify-between mb-4">
                  <div className="w-10 h-10 rounded-xl bg-cyan-500/10 border border-cyan-500/20 flex items-center justify-center">
                    <Shield size={18} className="text-cyan-400" />
                  </div>
                </div>
                <p className="text-[9px] font-black uppercase tracking-[0.2em] text-white/30 mb-1">Entity Extraction</p>
                <p className="text-3xl font-black text-white tracking-tight">
                  {((metrics?.regex_success_rate ?? 0) * 100).toFixed(0)}%
                </p>
                <p className="text-[9px] text-white/30 mt-1">
                  {metrics?.regex_success ?? 0} success / {metrics?.regex_failure ?? 0} fail
                </p>
              </div>
            </div>

            {/* Offline Plane Events */}
            <div className="stat-card neural-bg border-white/5 p-6 relative overflow-hidden group">
              <div className="absolute inset-0 bg-gradient-to-br from-purple-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
              <div className="relative">
                <div className="flex items-center justify-between mb-4">
                  <div className="w-10 h-10 rounded-xl bg-purple-500/10 border border-purple-500/20 flex items-center justify-center">
                    <Brain size={18} className="text-purple-400" />
                  </div>
                </div>
                <p className="text-[9px] font-black uppercase tracking-[0.2em] text-white/30 mb-1">Offline Events</p>
                <p className="text-3xl font-black text-white tracking-tight">
                  {(offline?.plane?.processed ?? 0).toLocaleString()}
                </p>
                <p className="text-[9px] text-white/30 mt-1">
                  {offline?.plane?.dropped ?? 0} dropped · {offline?.mode ?? 'disabled'}
                </p>
              </div>
            </div>
          </div>

          {/* Cache Layer Distribution */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Cache Layer Bar */}
            <div className="stat-card neural-bg border-white/5 p-6">
              <div className="flex items-center gap-3 mb-6">
                <Layers size={16} className="text-memzent-glow" />
                <h3 className="text-[11px] font-black uppercase tracking-[0.2em] text-white/60">Cache Layer Distribution</h3>
              </div>
              <div className="space-y-4">
                {[
                  { label: 'L1 Exact', value: metrics?.gpu_avoided ?? 0, color: 'bg-emerald-400', desc: 'Literal hash match' },
                  { label: 'L1b Entity', value: 0, color: 'bg-cyan-400', desc: 'Entity-keyed hot path' },
                  { label: 'L2 Semantic', value: 0, color: 'bg-blue-400', desc: 'Vector similarity' },
                  { label: 'L5 LLM', value: metrics?.gpu_invoked ?? 0, color: 'bg-orange-400', desc: 'Full GPU invocation' },
                ].map(layer => {
                  const max = Math.max(totalRequests, 1)
                  const pct = totalRequests > 0 ? (layer.value / max * 100) : 0
                  return (
                    <div key={layer.label}>
                      <div className="flex justify-between items-center mb-1">
                        <span className="text-[10px] font-bold text-white/60">{layer.label}</span>
                        <span className="text-[9px] text-white/40">{layer.value.toLocaleString()} ({pct.toFixed(1)}%)</span>
                      </div>
                      <div className="h-2 rounded-full bg-white/5 overflow-hidden">
                        <div className={`h-full rounded-full ${layer.color} transition-all duration-1000`} style={{ width: `${pct}%` }} />
                      </div>
                      <p className="text-[8px] text-white/20 mt-0.5">{layer.desc}</p>
                    </div>
                  )
                })}
              </div>
            </div>

            {/* Feature Flags Status */}
            <div className="stat-card neural-bg border-white/5 p-6">
              <div className="flex items-center gap-3 mb-6">
                <Activity size={16} className="text-memzent-glow" />
                <h3 className="text-[11px] font-black uppercase tracking-[0.2em] text-white/60">Evolution Layer Status</h3>
              </div>
              <div className="space-y-3">
                {[
                  { label: 'L1b Entity Cache', key: 'l1b_cache', desc: 'MEMZENT_L1B_ENABLED' },
                  { label: 'Offline Learning', key: 'offline_plane', desc: 'MEMZENT_OFFLINE_ENABLED' },
                  { label: 'Valkey Streams', key: 'offline_streams', desc: 'MEMZENT_OFFLINE_STREAMS' },
                  { label: 'Workflow Engine', key: 'workflow_engine', desc: 'MEMZENT_WORKFLOW_ENABLED' },
                  { label: 'Entity Metrics', key: 'entity_metrics', desc: 'MEMZENT_ENTITY_METRICS_ENABLED' },
                ].map(flag => {
                  const enabled = flags?.[flag.key as keyof FeatureFlags] ?? false
                  return (
                    <div key={flag.key} className="flex items-center justify-between p-3 rounded-lg bg-white/[0.02] border border-white/5">
                      <div>
                        <p className="text-[10px] font-bold text-white/70">{flag.label}</p>
                        <p className="text-[8px] text-white/20 font-mono">{flag.desc}</p>
                      </div>
                      <div className={`px-2 py-0.5 rounded-full text-[8px] font-black uppercase tracking-wider ${
                        enabled ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30' : 'bg-red-500/10 text-red-400/60 border border-red-500/20'
                      }`}>
                        {enabled ? 'ON' : 'OFF'}
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          </div>

          {/* Workflow Sequences (from O3 miner) */}
          {offline?.workflow_sequences && offline.workflow_sequences.length > 0 && (
            <div className="stat-card neural-bg border-white/5 p-6">
              <div className="flex items-center gap-3 mb-6">
                <BarChart3 size={16} className="text-memzent-glow" />
                <h3 className="text-[11px] font-black uppercase tracking-[0.2em] text-white/60">Detected Workflow Sequences</h3>
              </div>
              <div className="space-y-2">
                {offline.workflow_sequences.slice(0, 10).map((seq, i) => (
                  <div key={i} className="flex items-center justify-between p-3 rounded-lg bg-white/[0.02] border border-white/5">
                    <div className="flex items-center gap-3">
                      <span className="text-[9px] font-mono text-white/30">#{i + 1}</span>
                      <span className="text-[10px] font-mono text-cyan-300/80">{seq.pattern}</span>
                    </div>
                    <div className="flex items-center gap-4">
                      <span className="text-[9px] text-white/40">{seq.frequency}x</span>
                      <span className={`text-[9px] font-bold ${seq.success_rate >= 0.95 ? 'text-emerald-400' : seq.success_rate >= 0.8 ? 'text-yellow-400' : 'text-red-400'}`}>
                        {(seq.success_rate * 100).toFixed(0)}%
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Entity Quality Details */}
          <div className="stat-card neural-bg border-white/5 p-6">
            <div className="flex items-center gap-3 mb-6">
              <Gauge size={16} className="text-memzent-glow" />
              <h3 className="text-[11px] font-black uppercase tracking-[0.2em] text-white/60">Entity Extraction Quality</h3>
            </div>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              {[
                { label: 'Regex Success', value: metrics?.regex_success ?? 0, icon: CheckCircle2, color: 'text-emerald-400' },
                { label: 'Regex Failure', value: metrics?.regex_failure ?? 0, icon: TrendingDown, color: 'text-orange-400' },
                { label: 'Entity Mismatch', value: metrics?.entity_mismatch ?? 0, icon: Shield, color: 'text-red-400' },
                { label: 'LLM Extraction', value: metrics?.llm_entity_usage ?? 0, icon: Cpu, color: 'text-purple-400' },
              ].map(item => (
                <div key={item.label} className="text-center p-4 rounded-lg bg-white/[0.02] border border-white/5">
                  <item.icon size={20} className={`${item.color} mx-auto mb-2 opacity-60`} />
                  <p className="text-2xl font-black text-white">{item.value.toLocaleString()}</p>
                  <p className="text-[8px] font-bold uppercase tracking-wider text-white/30 mt-1">{item.label}</p>
                </div>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  )
}
