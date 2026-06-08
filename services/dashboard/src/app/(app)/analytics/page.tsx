'use client'

import { useEffect, useState } from 'react'
import {
  BarChart3, TrendingUp, DollarSign, Zap, Database, Activity,
  Brain, Layers, ArrowUpRight, ArrowDownRight, RefreshCw, Loader2,
  Clock, AlertTriangle, CheckCircle2, Target
} from 'lucide-react'
import { getContextAnalytics } from '@/app/actions'
import { supabase } from '@/lib/supabase'

interface ToolMetric {
  tool_id: string
  execution_count: number
  avg_latency_ms: number
  failure_rate: number
}

interface SavingsROI {
  cache_hits: number
  estimated_saved: number
  llm_cost: number
  net_roi: number
}

interface IntentCluster {
  intent: string
  frequency: number
}

interface ContextAnalytics {
  tool_metrics: ToolMetric[]
  savings_roi: SavingsROI
  semantic_clusters: IntentCluster[]
}

export default function AnalyticsPage() {
  const [data, setData] = useState<ContextAnalytics | null>(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)

  const fetchAnalytics = async (isRefresh = false) => {
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
      const result = await getContextAnalytics(orgId)
      setData(result)
    } catch (err) {
      console.error('Analytics fetch failed', err)
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => {
    fetchAnalytics()
  }, [])

  const roi = data?.savings_roi ?? { cache_hits: 0, estimated_saved: 0, llm_cost: 0, net_roi: 0 }
  const tools = data?.tool_metrics ?? []
  const clusters = data?.semantic_clusters ?? []
  const maxExecCount = Math.max(...tools.map(t => t.execution_count), 1)
  const maxClusterFreq = Math.max(...clusters.map(c => c.frequency), 1)

  return (
    <div className="space-y-8 pb-20">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className="w-2 h-8 rounded-full bg-gradient-to-b from-memzent-accent to-memzent-purple" />
          <div>
            <h1 className="text-3xl font-black tracking-tighter uppercase">Context Analytics</h1>
            <p className="text-[10px] font-black text-white/20 uppercase tracking-[0.3em] italic">
              Deep Intelligence · Cost Optimization · Intent Profiling
            </p>
          </div>
        </div>
        <button
          onClick={() => fetchAnalytics(true)}
          disabled={refreshing}
          className="flex items-center gap-2 bg-white/5 border border-white/10 text-white/60 font-black px-4 py-2.5 rounded-xl text-[10px] uppercase tracking-[0.2em] hover:border-memzent-glow/30 hover:text-memzent-glow transition-all disabled:opacity-30"
        >
          {refreshing ? <Loader2 size={12} className="animate-spin" /> : <RefreshCw size={12} />}
          Refresh
        </button>
      </div>

      {loading ? (
        <div className="flex flex-col items-center justify-center py-32">
          <Loader2 size={32} className="text-memzent-glow/30 animate-spin mb-4" />
          <p className="text-[10px] font-black uppercase tracking-[0.3em] text-memzent-glow/30 animate-pulse">
            Aggregating Telemetry Data...
          </p>
        </div>
      ) : (
        <>
          {/* KPI Cards Row */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* Cache Hit Savings */}
            <div className="stat-card neural-bg border-white/5 p-6 relative overflow-hidden group">
              <div className="absolute inset-0 bg-gradient-to-br from-memzent-glow/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
              <div className="relative">
                <div className="flex items-center justify-between mb-4">
                  <div className="w-10 h-10 rounded-xl bg-memzent-glow/10 border border-memzent-glow/20 flex items-center justify-center">
                    <Zap size={18} className="text-memzent-glow" />
                  </div>
                  {roi.net_roi > 0 && (
                    <span className="flex items-center gap-1 text-[9px] font-black text-emerald-400">
                      <ArrowUpRight size={10} /> +{roi.net_roi.toFixed(0)}%
                    </span>
                  )}
                </div>
                <p className="text-[9px] font-black uppercase tracking-[0.2em] text-white/30 mb-1">Cache Hits</p>
                <p className="text-3xl font-black text-white tracking-tight">{roi.cache_hits.toLocaleString()}</p>
              </div>
            </div>

            {/* Estimated Savings */}
            <div className="stat-card neural-bg border-white/5 p-6 relative overflow-hidden group">
              <div className="absolute inset-0 bg-gradient-to-br from-emerald-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
              <div className="relative">
                <div className="flex items-center justify-between mb-4">
                  <div className="w-10 h-10 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center">
                    <DollarSign size={18} className="text-emerald-400" />
                  </div>
                  <span className="text-[9px] font-black text-emerald-400">SAVED</span>
                </div>
                <p className="text-[9px] font-black uppercase tracking-[0.2em] text-white/30 mb-1">Estimated Savings</p>
                <p className="text-3xl font-black text-emerald-400 tracking-tight">${roi.estimated_saved.toFixed(4)}</p>
              </div>
            </div>

            {/* LLM Spend */}
            <div className="stat-card neural-bg border-white/5 p-6 relative overflow-hidden group">
              <div className="absolute inset-0 bg-gradient-to-br from-memzent-purple/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
              <div className="relative">
                <div className="flex items-center justify-between mb-4">
                  <div className="w-10 h-10 rounded-xl bg-memzent-purple/10 border border-memzent-purple/20 flex items-center justify-center">
                    <Brain size={18} className="text-memzent-purple" />
                  </div>
                  {roi.llm_cost > 0 && (
                    <span className="flex items-center gap-1 text-[9px] font-black text-red-400/60">
                      <ArrowDownRight size={10} /> SPENT
                    </span>
                  )}
                </div>
                <p className="text-[9px] font-black uppercase tracking-[0.2em] text-white/30 mb-1">LLM Expenditure</p>
                <p className="text-3xl font-black text-white tracking-tight">${roi.llm_cost.toFixed(4)}</p>
              </div>
            </div>

            {/* Net ROI */}
            <div className="stat-card neural-bg border-white/5 p-6 relative overflow-hidden group">
              <div className="absolute inset-0 bg-gradient-to-br from-memzent-accent/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
              <div className="relative">
                <div className="flex items-center justify-between mb-4">
                  <div className="w-10 h-10 rounded-xl bg-memzent-accent/10 border border-memzent-accent/20 flex items-center justify-center">
                    <TrendingUp size={18} className="text-memzent-accent" />
                  </div>
                </div>
                <p className="text-[9px] font-black uppercase tracking-[0.2em] text-white/30 mb-1">Net Token ROI</p>
                <p className={`text-3xl font-black tracking-tight ${roi.net_roi > 0 ? 'text-memzent-accent' : 'text-white/40'}`}>
                  {roi.net_roi > 0 ? `${roi.net_roi.toFixed(1)}%` : '—'}
                </p>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Tool Execution Telemetry */}
            <div className="lg:col-span-2 stat-card neural-bg border-white/5 p-6">
              <div className="flex items-center gap-3 mb-6">
                <Activity size={16} className="text-memzent-glow" />
                <h2 className="text-xs font-black uppercase tracking-widest text-white/60">Tool Execution Telemetry</h2>
                <span className="ml-auto text-[9px] font-black uppercase tracking-widest text-white/20">
                  {tools.length} ACTIVE TOOLS
                </span>
              </div>

              {tools.length === 0 ? (
                <div className="py-16 text-center">
                  <Database size={28} className="text-white/5 mx-auto mb-3" />
                  <p className="text-[10px] font-black uppercase tracking-[0.3em] text-white/10">
                    No tool executions recorded yet
                  </p>
                </div>
              ) : (
                <div className="space-y-3">
                  {tools.map((tool) => (
                    <div key={tool.tool_id} className="p-4 rounded-xl bg-white/[0.02] border border-white/5 hover:border-white/10 transition-all group">
                      <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-3">
                          <div className={`w-2 h-2 rounded-full ${tool.failure_rate > 10 ? 'bg-red-400 shadow-[0_0_6px_rgba(239,68,68,0.5)]' : 'bg-memzent-glow shadow-[0_0_6px_rgba(0,243,255,0.4)]'}`} />
                          <span className="text-[10px] font-black text-white/70 uppercase tracking-wider">{tool.tool_id}</span>
                        </div>
                        <div className="flex items-center gap-4">
                          <span className="flex items-center gap-1 text-[9px] font-mono text-white/30">
                            <Clock size={8} /> {tool.avg_latency_ms}ms
                          </span>
                          <span className={`flex items-center gap-1 text-[9px] font-black ${tool.failure_rate > 10 ? 'text-red-400' : 'text-emerald-400'}`}>
                            {tool.failure_rate > 10 ? <AlertTriangle size={8} /> : <CheckCircle2 size={8} />}
                            {(100 - tool.failure_rate).toFixed(1)}% OK
                          </span>
                        </div>
                      </div>

                      {/* Execution bar */}
                      <div className="relative h-2 bg-white/5 rounded-full overflow-hidden">
                        <div
                          className={`absolute left-0 top-0 h-full rounded-full transition-all duration-700 ${
                            tool.failure_rate > 10
                              ? 'bg-gradient-to-r from-red-500/60 to-red-400/40'
                              : 'bg-gradient-to-r from-memzent-glow/60 to-memzent-purple/40'
                          }`}
                          style={{ width: `${(tool.execution_count / maxExecCount) * 100}%` }}
                        />
                      </div>

                      <div className="flex items-center justify-between mt-2">
                        <span className="text-[8px] font-black uppercase tracking-widest text-white/15">{tool.execution_count} EXECUTIONS</span>
                        {tool.failure_rate > 0 && (
                          <span className="text-[8px] font-black uppercase tracking-widest text-red-400/40">{tool.failure_rate.toFixed(1)}% FAILURE</span>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Semantic Intent Clusters */}
            <div className="stat-card neural-bg border-white/5 p-6">
              <div className="flex items-center gap-3 mb-6">
                <Target size={16} className="text-memzent-purple" />
                <h2 className="text-xs font-black uppercase tracking-widest text-white/60">Intent Themes</h2>
              </div>

              {clusters.length === 0 ? (
                <div className="py-16 text-center">
                  <Layers size={28} className="text-white/5 mx-auto mb-3" />
                  <p className="text-[10px] font-black uppercase tracking-[0.3em] text-white/10">
                    Awaiting intent clustering data
                  </p>
                </div>
              ) : (
                <div className="space-y-3">
                  {clusters.map((cluster, idx) => (
                    <div key={idx} className="p-4 rounded-xl bg-white/[0.02] border border-white/5 hover:border-memzent-purple/20 transition-all group">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-[9px] font-bold text-white/50 leading-tight line-clamp-2 break-words flex-1 mr-2">
                          {cluster.intent}
                        </span>
                        <span className="text-[9px] font-black font-mono text-memzent-purple flex-shrink-0">
                          ×{cluster.frequency}
                        </span>
                      </div>

                      {/* Frequency bar */}
                      <div className="relative h-1.5 bg-white/5 rounded-full overflow-hidden">
                        <div
                          className="absolute left-0 top-0 h-full rounded-full bg-gradient-to-r from-memzent-purple/60 to-memzent-glow/30 transition-all duration-700"
                          style={{ width: `${(cluster.frequency / maxClusterFreq) * 100}%` }}
                        />
                      </div>
                    </div>
                  ))}
                </div>
              )}

              {/* Cost Efficiency Insight */}
              <div className="mt-6 p-4 rounded-xl bg-gradient-to-br from-memzent-glow/5 to-memzent-purple/5 border border-memzent-glow/10">
                <div className="flex items-center gap-2 mb-2">
                  <BarChart3 size={12} className="text-memzent-glow" />
                  <span className="text-[9px] font-black uppercase tracking-widest text-memzent-glow/60">Insight</span>
                </div>
                <p className="text-[10px] text-white/40 leading-relaxed">
                  {roi.cache_hits > 0
                    ? `Semantic caching has intercepted ${roi.cache_hits} redundant LLM calls, saving an estimated $${roi.estimated_saved.toFixed(4)} in token costs.`
                    : 'No cache hits recorded yet. Cache savings will accumulate as repeated prompts are intercepted by the semantic proxy.'}
                </p>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
