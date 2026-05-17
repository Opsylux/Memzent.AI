'use client'

import { useState } from 'react'
import {
  Send, Bot, Zap, Database, Cpu, CheckCircle2, ArrowRight,
  Loader2, Search, Clock, DollarSign, ShieldCheck, Terminal,
  BarChart3, Layers
} from 'lucide-react'
import { executeMemzentPrompt } from '@/app/actions'
import { supabase } from '@/lib/supabase'

const EXAMPLE_PROMPTS = [
  "What database metrics are available?",
  "Search for the latest tool configurations",
  "Summarize the current system health status",
  "List all active MCP endpoints",
]

type TraceStatus = 'idle' | 'running' | 'cache_hit' | 'llm_hit' | 'error'

export default function PlaygroundPage() {
  const [prompt, setPrompt] = useState('')
  const [status, setStatus] = useState<TraceStatus>('idle')
  const [result, setResult] = useState<any>(null)
  const [history, setHistory] = useState<{ prompt: string; cached: boolean; ms: number }[]>([])

  const handleRun = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!prompt.trim() || status === 'running') return

    const start = Date.now()
    setStatus('running')
    setResult(null)

    try {
      const { data: { user } } = await supabase.auth.getUser()
      const { data: membership } = await supabase
        .from('members')
        .select('org_id')
        .eq('user_id', user?.id ?? '')
        .limit(1)
        .maybeSingle()

      const orgId = membership?.org_id || user?.id
      const res = await executeMemzentPrompt(prompt.trim(), orgId)
      const elapsed = Date.now() - start

      setResult({ ...res, elapsed })
      setStatus(res.cached ? 'cache_hit' : 'llm_hit')
      setHistory(prev => [{ prompt: prompt.trim(), cached: res.cached, ms: elapsed }, ...prev.slice(0, 9)])
    } catch (err: any) {
      setResult({ error: err.message })
      setStatus('error')
    }
  }

  const isRunning = status === 'running'

  return (
    <div className="space-y-8 pb-20">
      {/* Header */}
      <div className="flex items-center gap-4 mb-4">
        <div className="w-2 h-8 rounded-full bg-gradient-to-b from-memzent-glow to-memzent-purple" />
        <div>
          <h1 className="text-3xl font-black tracking-tighter uppercase">Neural Playground</h1>
          <p className="text-[10px] font-black text-white/20 uppercase tracking-[0.3em] italic">Live Prompt Execution & Cost Trace</p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Main Execution Panel */}
        <div className="lg:col-span-2 space-y-6">
          {/* Prompt Input */}
          <div className="stat-card neural-bg border-white/5 p-6">
            <div className="flex items-center gap-3 mb-4">
              <Terminal size={16} className="text-memzent-glow" />
              <h2 className="text-xs font-black uppercase tracking-widest text-white/60">Prompt Terminal</h2>
            </div>
            <form onSubmit={handleRun} className="space-y-4">
              <textarea
                value={prompt}
                onChange={e => setPrompt(e.target.value)}
                onKeyDown={e => { if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) handleRun(e as any) }}
                placeholder="Enter a prompt to route through Memzent Gateway..."
                className="w-full bg-black/40 border border-white/10 rounded-2xl px-5 py-4 text-sm font-bold text-white focus:border-memzent-glow outline-none transition-all resize-none placeholder:text-white/10 min-h-[100px]"
                disabled={isRunning}
              />
              <div className="flex items-center justify-between">
                <div className="flex flex-wrap gap-2">
                  {EXAMPLE_PROMPTS.map(p => (
                    <button
                      key={p}
                      type="button"
                      onClick={() => setPrompt(p)}
                      className="text-[9px] font-black uppercase tracking-widest px-3 py-1.5 rounded-lg bg-white/5 border border-white/5 text-white/30 hover:text-memzent-glow hover:border-memzent-glow/20 transition-all"
                    >
                      {p.split(' ').slice(0, 3).join(' ')}...
                    </button>
                  ))}
                </div>
                <button
                  type="submit"
                  disabled={isRunning || !prompt.trim()}
                  className="flex items-center gap-2 bg-memzent-glow text-black font-black px-6 py-3 rounded-xl text-[10px] uppercase tracking-[0.2em] hover:shadow-[0_0_20px_rgba(0,243,255,0.4)] transition-all disabled:opacity-30 disabled:grayscale"
                >
                  {isRunning ? <Loader2 size={14} className="animate-spin" /> : <Send size={14} />}
                  {isRunning ? 'Routing...' : 'Execute'}
                </button>
              </div>
            </form>
          </div>

          {/* Execution Trace Pipeline */}
          <div className="stat-card neural-bg border-white/5 p-6">
            <div className="flex items-center gap-3 mb-6">
              <Layers size={16} className="text-memzent-purple" />
              <h2 className="text-xs font-black uppercase tracking-widest text-white/60">Execution Pipeline</h2>
              {status !== 'idle' && (
                <span className={`ml-auto text-[9px] font-black uppercase tracking-widest px-2 py-1 rounded-md border ${
                  status === 'cache_hit' ? 'text-memzent-glow border-memzent-glow/20 bg-memzent-glow/5' :
                  status === 'llm_hit' ? 'text-memzent-purple border-memzent-purple/20 bg-memzent-purple/5' :
                  status === 'error' ? 'text-red-400 border-red-500/20 bg-red-500/5' :
                  'text-white/30 border-white/10'
                }`}>
                  {status === 'cache_hit' ? '⚡ Cache Hit' : status === 'llm_hit' ? '🤖 LLM Generated' : status === 'error' ? '❌ Error' : 'Processing...'}
                </span>
              )}
            </div>

            <div className="flex flex-col md:flex-row items-center gap-2 md:gap-0 mb-6">
              {[
                { icon: <ShieldCheck size={18} />, label: 'Auth & Rate Limit', done: status !== 'idle' },
                { icon: <Search size={18} />, label: 'Cache Lookup', done: status !== 'idle' && !isRunning },
                { icon: <Cpu size={18} />, label: 'Semantic Route', done: (status === 'llm_hit' || status === 'cache_hit') },
                { icon: <Bot size={18} />, label: status === 'cache_hit' ? 'Cache Return' : 'LLM Synthesis', done: status === 'llm_hit' || status === 'cache_hit' },
                { icon: <CheckCircle2 size={18} />, label: 'Response', done: status === 'llm_hit' || status === 'cache_hit' },
              ].map((step, i) => (
                <div key={i} className="flex items-center w-full md:w-auto">
                  <div className={`flex flex-col items-center gap-2 flex-1 md:flex-none md:w-24 p-3 rounded-xl border transition-all ${
                    step.done && status === 'cache_hit' && i === 2 ? 'bg-memzent-glow/10 border-memzent-glow/20 text-memzent-glow' :
                    step.done ? 'bg-white/5 border-white/10 text-white' :
                    isRunning ? 'bg-white/[0.02] border-white/5 text-white/20 animate-pulse' :
                    'bg-white/[0.02] border-white/5 text-white/10'
                  }`}>
                    {isRunning && !step.done ? <Loader2 size={18} className="animate-spin" /> : step.icon}
                    <span className="text-[8px] font-black uppercase tracking-wider text-center leading-tight">{step.label}</span>
                  </div>
                  {i < 4 && <ArrowRight size={12} className="text-white/10 mx-1 hidden md:block flex-shrink-0" />}
                </div>
              ))}
            </div>

            {/* Result */}
            {result && !result.error && (
              <div className={`p-5 rounded-2xl border transition-all ${
                result.cached
                  ? 'bg-memzent-glow/5 border-memzent-glow/20'
                  : 'bg-memzent-purple/5 border-memzent-purple/20'
              }`}>
                <div className="text-[10px] font-black uppercase tracking-widest text-white/30 mb-3 flex items-center gap-2">
                  {result.cached ? <><Zap size={10} className="text-memzent-glow" /> Semantic Cache Response</> : <><Bot size={10} className="text-memzent-purple" /> LLM Generated Response</>}
                </div>
                <p className="text-sm text-white/80 font-medium leading-relaxed whitespace-pre-wrap">{result.text}</p>
              </div>
            )}

            {result?.error && (
              <div className="p-5 rounded-2xl bg-red-500/5 border border-red-500/20">
                <div className="text-[10px] font-black uppercase tracking-widest text-red-400/60 mb-2">Gateway Error</div>
                <p className="text-sm text-red-400 font-mono">{result.error}</p>
              </div>
            )}

            {status === 'idle' && (
              <div className="py-12 text-center">
                <Database size={32} className="text-memzent-glow/10 mx-auto mb-4 animate-pulse" />
                <p className="text-[10px] font-black uppercase tracking-[0.3em] text-memzent-glow/30 animate-pulse">Awaiting Prompt Evaluation...</p>
              </div>
            )}
          </div>
        </div>

        {/* Right Sidebar */}
        <div className="space-y-6">
          {/* Cost Trace Card */}
          <div className="stat-card neural-bg border-white/5 p-6">
            <div className="flex items-center gap-3 mb-6">
              <DollarSign size={16} className="text-memzent-accent" />
              <h2 className="text-xs font-black uppercase tracking-widest text-white/60">Cost Trace</h2>
            </div>
            <div className="space-y-4">
              {[
                { label: 'Latency', value: result?.elapsed ? `${result.elapsed}ms` : '—', color: 'text-white' },
                { label: 'Cache Status', value: result ? (result.cached ? 'HIT ⚡' : 'MISS') : '—', color: result?.cached ? 'text-memzent-glow' : 'text-white/40' },
                { label: 'Tokens Used', value: result?.usage ? `${result.usage.total_tokens || 0}` : '—', color: 'text-white' },
                { label: 'Est. Cost', value: result?.usage && !result.cached ? `$${((result.usage.total_tokens || 0) * 0.000002).toFixed(6)}` : result?.cached ? '$0.00 (80% off)' : '—', color: result?.cached ? 'text-memzent-accent' : 'text-white/40' },
              ].map(item => (
                <div key={item.label} className="flex items-center justify-between py-2 border-b border-white/5">
                  <span className="text-[10px] font-black uppercase tracking-widest text-white/30">{item.label}</span>
                  <span className={`text-xs font-black font-mono ${item.color}`}>{item.value}</span>
                </div>
              ))}
            </div>
            {result?.cached && (
              <div className="mt-4 p-3 rounded-xl bg-memzent-accent/5 border border-memzent-accent/20">
                <p className="text-[9px] font-black uppercase tracking-widest text-memzent-accent">80% discount applied — semantic cache hit!</p>
              </div>
            )}
          </div>

          {/* Tools Matched */}
          <div className="stat-card neural-bg border-white/5 p-6">
            <div className="flex items-center gap-3 mb-6">
              <Database size={16} className="text-memzent-purple" />
              <h2 className="text-xs font-black uppercase tracking-widest text-white/60">Matched Tools</h2>
            </div>
            {result?.tools && result.tools.length > 0 ? (
              <div className="space-y-3">
                {result.tools.slice(0, 5).map((tool: any, i: number) => (
                  <div key={i} className="flex items-center justify-between p-3 rounded-xl bg-white/[0.02] border border-white/5">
                    <div className="flex items-center gap-3">
                      <div className="w-1.5 h-1.5 rounded-full bg-memzent-purple shadow-[0_0_6px_rgba(157,0,255,0.5)]" />
                      <span className="text-[10px] font-black text-white/60 uppercase truncate w-28">{tool.name || tool.id}</span>
                    </div>
                    <span className="text-[9px] font-black font-mono text-white/20">{tool.relevance_score ? tool.relevance_score.toFixed(3) : 'N/A'}</span>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-[10px] font-black uppercase tracking-widest text-white/10 text-center py-6">
                {status === 'idle' ? 'No execution yet' : result?.cached ? 'Served from cache' : 'No tools matched'}
              </p>
            )}
          </div>

          {/* Execution History */}
          <div className="stat-card neural-bg border-white/5 p-6">
            <div className="flex items-center gap-3 mb-6">
              <BarChart3 size={16} className="text-white/30" />
              <h2 className="text-xs font-black uppercase tracking-widest text-white/60">Run History</h2>
            </div>
            {history.length === 0 ? (
              <p className="text-[10px] font-black uppercase tracking-widest text-white/10 text-center py-4">No runs yet</p>
            ) : (
              <div className="space-y-2">
                {history.map((h, i) => (
                  <button
                    key={i}
                    onClick={() => setPrompt(h.prompt)}
                    className="w-full flex items-center justify-between p-3 rounded-xl bg-white/[0.02] border border-white/5 hover:border-white/10 transition-all text-left group"
                  >
                    <span className="text-[10px] font-bold text-white/40 group-hover:text-white truncate w-32 transition-colors">{h.prompt}</span>
                    <div className="flex items-center gap-2 flex-shrink-0">
                      <span className={`text-[9px] font-black ${h.cached ? 'text-memzent-glow' : 'text-memzent-purple'}`}>
                        {h.cached ? '⚡' : '🤖'}
                      </span>
                      <span className="text-[9px] font-mono text-white/20"><Clock size={8} className="inline mr-1" />{h.ms}ms</span>
                    </div>
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
