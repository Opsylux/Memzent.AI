'use client'

import { useState, useEffect, useRef } from 'react'
import {
  Send, Bot, Zap, Database, Cpu, CheckCircle2, ArrowRight,
  Loader2, Search, Clock, DollarSign, ShieldCheck, Terminal,
  BarChart3, Layers, Plus, Trash2, MessageSquare, BrainCircuit, RefreshCw
} from 'lucide-react'
import {
  executeMemzentPrompt,
  getSessions,
  createSession,
  getSessionMessages,
  deleteSession
} from '@/app/actions'
import { supabase } from '@/lib/supabase'
import { Markdown } from '@/components/markdown'
import { PipelineTrace, type PipelineStatus } from '@/components/pipeline-trace'

const EXAMPLE_PROMPTS = [
  "My server IP is 192.168.1.100 and port is 8080",
  "Recall the server details I just mentioned",
  "Summarize system memory",
  "What is the system latency standard?",
]

export default function PlaygroundPage() {
  const [prompt, setPrompt] = useState('')
  const [status, setStatus] = useState<PipelineStatus>('idle')
  const [pipelineStep, setPipelineStep] = useState(0)
  const [result, setResult] = useState<any>(null)
  const [sessions, setSessions] = useState<any[]>([])
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null)
  const [messages, setMessages] = useState<any[]>([])
  const [loadingSessions, setLoadingSessions] = useState(true)
  const [loadingMessages, setLoadingMessages] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const pipelineTimerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => {
    loadSessions()
    return () => {
      if (pipelineTimerRef.current) clearInterval(pipelineTimerRef.current)
    }
  }, [])

  useEffect(() => {
    if (status !== 'running') {
      if (pipelineTimerRef.current) {
        clearInterval(pipelineTimerRef.current)
        pipelineTimerRef.current = null
      }
      return
    }
    setPipelineStep(0)
    pipelineTimerRef.current = setInterval(() => {
      setPipelineStep((s) => (s >= 6 ? 6 : s + 1))
    }, 280)
  }, [status])

  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [messages])

  const loadSessions = async () => {
    setLoadingSessions(true)
    try {
      const { data: { user } } = await supabase.auth.getUser()
      const { data: membership } = await supabase
        .from('members')
        .select('org_id')
        .eq('user_id', user?.id ?? '')
        .limit(1)
        .maybeSingle()

      const orgId = membership?.org_id || user?.id
      const sessList = await getSessions(orgId)
      setSessions(sessList || [])
      
      if (sessList && sessList.length > 0) {
        handleSelectSession(sessList[0].id)
      }
    } catch (e) {
      console.error("Failed to load sessions", e)
    } finally {
      setLoadingSessions(false)
    }
  }

  const handleSelectSession = async (sessId: string) => {
    setActiveSessionId(sessId)
    setLoadingMessages(true)
    setMessages([])
    setResult(null)
    setStatus('idle')
    try {
      const { data: { user } } = await supabase.auth.getUser()
      const { data: membership } = await supabase
        .from('members')
        .select('org_id')
        .eq('user_id', user?.id ?? '')
        .limit(1)
        .maybeSingle()

      const orgId = membership?.org_id || user?.id
      const msgs = await getSessionMessages(sessId, orgId)
      setMessages(msgs || [])
    } catch (e) {
      console.error("Failed to load session messages", e)
    } finally {
      setLoadingMessages(false)
    }
  }

  const handleCreateSession = async () => {
    try {
      const { data: { user } } = await supabase.auth.getUser()
      const { data: membership } = await supabase
        .from('members')
        .select('org_id')
        .eq('user_id', user?.id ?? '')
        .limit(1)
        .maybeSingle()

      const orgId = membership?.org_id || user?.id
      const newSess = await createSession(orgId, `Thread ${new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`)
      if (newSess && newSess.id) {
        setSessions(prev => [newSess, ...prev])
        setActiveSessionId(newSess.id)
        setMessages([])
        setResult(null)
        setStatus('idle')
      }
    } catch (e) {
      console.error("Failed to create session", e)
    }
  }

  const handleDeleteSession = async (sessId: string, e: React.MouseEvent) => {
    e.stopPropagation()
    try {
      const { data: { user } } = await supabase.auth.getUser()
      const { data: membership } = await supabase
        .from('members')
        .select('org_id')
        .eq('user_id', user?.id ?? '')
        .limit(1)
        .maybeSingle()

      const orgId = membership?.org_id || user?.id
      await deleteSession(sessId, orgId)
      setSessions(prev => prev.filter(s => s.id !== sessId))
      if (activeSessionId === sessId) {
        setActiveSessionId(null)
        setMessages([])
        setResult(null)
        setStatus('idle')
      }
    } catch (err) {
      console.error("Failed to delete session", err)
    }
  }

  const handleRun = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!prompt.trim() || status === 'running') return

    let currentSessionId = activeSessionId
    if (!currentSessionId) {
      try {
        const { data: { user } } = await supabase.auth.getUser()
        const { data: membership } = await supabase
          .from('members')
          .select('org_id')
          .eq('user_id', user?.id ?? '')
          .limit(1)
          .maybeSingle()

        const orgId = membership?.org_id || user?.id
        const newSess = await createSession(orgId, `Auto Session`)
        if (newSess && newSess.id) {
          setSessions(prev => [newSess, ...prev])
          currentSessionId = newSess.id
          setActiveSessionId(newSess.id)
        }
      } catch (err) {
        console.error("Failed to auto create session", err)
        return
      }
    }

    const newMsg = { role: 'user', content: prompt.trim() }
    const updatedMessages = [...messages, newMsg]
    setMessages(updatedMessages)
    setPrompt('')

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
      const res = await executeMemzentPrompt(updatedMessages, currentSessionId ?? undefined, orgId)
      const elapsed = Date.now() - start

      setMessages([...updatedMessages, { role: 'assistant', content: res.text }])
      setResult({ ...res, elapsed })
      setStatus(res.cached ? 'cache_hit' : 'llm_hit')
    } catch (err: any) {
      setResult({ error: err.message })
      setStatus('error')
    }
  }

  const isRunning = status === 'running'

  return (
    <div className="space-y-8 pb-20">
      <header className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex items-center gap-4">
          <div className="w-1.5 h-10 rounded-full bg-gradient-to-b from-memzent-glow to-memzent-purple" />
          <div>
            <h1 className="text-2xl font-black tracking-tight text-readable-primary">Playground</h1>
            <p className="text-sm text-readable-muted mt-1">
              Test cache layers, tool routing, and session memory
            </p>
          </div>
        </div>
        {status !== 'idle' && (
          <span className={`inline-flex items-center gap-2 text-xs font-bold px-3 py-1.5 rounded-full border w-fit ${
            status === 'cache_hit' ? 'text-memzent-glow border-memzent-glow/30 bg-memzent-glow/10' :
            status === 'llm_hit' ? 'text-memzent-purple border-memzent-purple/30 bg-memzent-purple/10' :
            status === 'error' ? 'text-red-400 border-red-500/30 bg-red-500/10' :
            'text-white/60 border-white/10 bg-white/5'
          }`}>
            {status === 'running' && <Loader2 size={12} className="animate-spin" />}
            {status === 'cache_hit' ? 'Cache hit' : status === 'llm_hit' ? 'LLM synthesis' : status === 'error' ? 'Failed' : 'Running…'}
          </span>
        )}
      </header>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
        
        {/* Left Column: Sessions list */}
        <div className="stat-card neural-bg border-white/5 p-6 flex flex-col h-[650px]">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-2">
              <MessageSquare size={14} className="text-memzent-glow" />
              <span className="text-xs font-black uppercase tracking-widest text-white/60">Sessions</span>
            </div>
            <button
              onClick={handleCreateSession}
              className="p-2 rounded-lg bg-white/5 border border-white/10 hover:border-memzent-glow/40 hover:text-memzent-glow text-white/70 transition-all"
              title="New Conversation"
            >
              <Plus size={14} />
            </button>
          </div>

          <div className="flex-1 overflow-y-auto space-y-2 pr-1 custom-scrollbar">
            {loadingSessions ? (
              <div className="flex justify-center py-10">
                <Loader2 size={16} className="animate-spin text-white/20" />
              </div>
            ) : sessions.length === 0 ? (
              <p className="text-xs text-readable-muted text-center py-6">No sessions yet — send a message to start</p>
            ) : (
              sessions.map(s => {
                const isActive = activeSessionId === s.id
                return (
                  <div
                    key={s.id}
                    onClick={() => handleSelectSession(s.id)}
                    className={`flex items-center justify-between p-3 rounded-xl border cursor-pointer transition-all group ${
                      isActive
                        ? 'bg-memzent-glow/10 border-memzent-glow/20 text-memzent-glow'
                        : 'bg-white/[0.02] border-white/5 text-white/50 hover:bg-white/[0.04] hover:text-white'
                    }`}
                  >
                    <span className="text-[10px] font-black truncate w-40 uppercase tracking-wide">
                      {s.title || "Thread"}
                    </span>
                    <button
                      onClick={(e) => handleDeleteSession(s.id, e)}
                      className="opacity-0 group-hover:opacity-100 p-1 hover:text-red-400 transition-all rounded"
                    >
                      <Trash2 size={10} />
                    </button>
                  </div>
                )
              })
            )}
          </div>

          <div className="mt-6 pt-4 border-t border-white/5">
            <div className="flex items-center gap-2 mb-2">
              <BrainCircuit size={14} className="text-memzent-purple" />
              <span className="text-[9px] font-black uppercase tracking-widest text-memzent-purple">Agent Memory</span>
            </div>
            <p className="text-[11px] text-readable-muted leading-relaxed">
              Facts are extracted after each turn and stored in Qdrant. Relevant memory is injected on the next request.
            </p>
          </div>
        </div>

        {/* Center Column: Active Chat Area */}
        <div className="lg:col-span-2 flex flex-col h-[650px] space-y-4">
          
          {/* Scrollable Conversation History */}
          <div className="flex-1 stat-card neural-bg border-white/5 p-6 overflow-y-auto flex flex-col space-y-4 custom-scrollbar">
            {loadingMessages ? (
              <div className="flex flex-col items-center justify-center h-full">
                <Loader2 size={24} className="animate-spin text-memzent-glow/30 mb-2" />
                <span className="text-[9px] font-black uppercase tracking-wider text-white/20">Syncing Thread History...</span>
              </div>
            ) : messages.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-full text-center">
                <Bot size={40} className="text-white/5 mb-4 animate-pulse" />
                <h3 className="text-sm font-bold text-readable-secondary mb-1">Start a conversation</h3>
                <p className="text-xs text-readable-muted max-w-xs leading-relaxed">
                  Pick a session or type below — a new thread is created automatically.
                </p>
              </div>
            ) : (
              <div className="space-y-4">
                {messages.map((m, i) => {
                  const isUser = m.role === 'user'
                  return (
                    <div
                      key={i}
                      className={`flex gap-4 ${isUser ? 'justify-end' : 'justify-start'}`}
                    >
                      <div className={`flex gap-3 max-w-[85%] ${isUser ? 'flex-row-reverse' : 'flex-row'}`}>
                        <div className={`w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 border ${
                          isUser
                            ? 'bg-memzent-glow/10 border-memzent-glow/20 text-memzent-glow'
                            : 'bg-memzent-purple/10 border-memzent-purple/20 text-memzent-purple'
                        }`}>
                          {isUser ? <Terminal size={14} /> : <Bot size={14} />}
                        </div>
                        
                        <div className={`p-4 rounded-2xl border ${
                          isUser
                            ? 'bg-white/[0.03] border-white/10 text-white/80'
                            : 'bg-gradient-to-br from-memzent-purple/5 to-transparent border-memzent-purple/10 text-white/90'
                        }`}>
                          <Markdown content={m.content} />
                        </div>
                      </div>
                    </div>
                  )
                })}
                <div ref={messagesEndRef} />
              </div>
            )}
          </div>

          {/* Chat Form */}
          <div className="stat-card neural-bg border-white/5 p-4">
            <form onSubmit={handleRun} className="space-y-3">
              <div className="relative flex items-center">
                <textarea
                  value={prompt}
                  onChange={e => setPrompt(e.target.value)}
                  onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleRun(e as any) } }}
                  placeholder="Ask anything — routes through Memzent gateway…"
                  className="w-full bg-black/40 border border-white/10 rounded-xl pl-4 pr-12 py-3.5 text-sm text-readable-primary focus:border-memzent-glow outline-none transition-all resize-none placeholder:text-readable-muted min-h-[52px] max-h-[120px]"
                  disabled={isRunning}
                />
                <button
                  type="submit"
                  disabled={isRunning || !prompt.trim()}
                  className="absolute right-3 bg-memzent-glow text-black font-black p-2 rounded-lg hover:shadow-[0_0_12px_rgba(0,243,255,0.4)] transition-all disabled:opacity-20 disabled:grayscale"
                >
                  {isRunning ? <Loader2 size={12} className="animate-spin" /> : <Send size={12} />}
                </button>
              </div>

              <div className="flex flex-wrap gap-2 items-center justify-between">
                <div className="flex gap-2">
                  {EXAMPLE_PROMPTS.map(p => (
                    <button
                      key={p}
                      type="button"
                      onClick={() => setPrompt(p)}
                      className="text-[11px] font-medium px-2.5 py-1 rounded-lg bg-white/5 border border-white/5 text-readable-muted hover:text-memzent-glow hover:border-memzent-glow/20 transition-all"
                    >
                      {p.split(' ').slice(0, 3).join(' ')}...
                    </button>
                  ))}
                </div>
              </div>
            </form>
          </div>
        </div>

        {/* Right Column: Pipeline + telemetry */}
        <div className="space-y-5 h-[650px] overflow-y-auto custom-scrollbar">
          <div className="stat-card neural-bg border-white/5 p-6">
            <PipelineTrace
              status={status}
              activeStep={pipelineStep}
              elapsedMs={result?.elapsed}
            />
          </div>

          <div className="stat-card neural-bg border-white/5 p-6">
            <div className="flex items-center gap-3 mb-6">
              <DollarSign size={14} className="text-memzent-accent" />
              <h2 className="text-xs font-bold uppercase tracking-wider text-readable-label">Cost & latency</h2>
            </div>
            <div className="space-y-4">
              {[
                { label: 'Latency', value: result?.elapsed ? `${result.elapsed}ms` : '—', color: 'text-white' },
                { label: 'Cache Status', value: result ? (result.cached ? 'HIT ⚡' : 'MISS') : '—', color: result?.cached ? 'text-memzent-glow' : 'text-white/40' },
                { label: 'Tokens Used', value: result?.usage ? `${result.usage.total_tokens || 0}` : '—', color: 'text-white' },
                { label: 'Est. Cost', value: result?.usage && !result.cached ? `$${((result.usage.total_tokens || 0) * 0.000002).toFixed(6)}` : result?.cached ? '$0.00 (80% off)' : '—', color: result?.cached ? 'text-memzent-accent' : 'text-white/40' },
              ].map(item => (
                <div key={item.label} className="flex items-center justify-between py-2 border-b border-white/5">
                  <span className="text-[11px] text-readable-muted">{item.label}</span>
                  <span className={`text-sm font-bold font-mono ${item.color}`}>{item.value}</span>
                </div>
              ))}
            </div>
            {result?.cached && (
              <div className="mt-4 p-3 rounded-xl bg-memzent-accent/5 border border-memzent-accent/20">
                <p className="text-[11px] font-medium text-memzent-accent">80% billing discount — served from semantic cache</p>
              </div>
            )}
          </div>

          {/* Tools Matched */}
          <div className="stat-card neural-bg border-white/5 p-6">
            <div className="flex items-center gap-3 mb-6">
              <Database size={14} className="text-memzent-purple" />
              <h2 className="text-xs font-black uppercase tracking-widest text-white/60">Matched Tools</h2>
            </div>
            {result?.tools && result.tools.length > 0 ? (
              <div className="space-y-3">
                {result.tools.slice(0, 5).map((tool: any, i: number) => (
                  <div key={i} className="flex items-center justify-between p-3 rounded-xl bg-white/[0.02] border border-white/5">
                    <div className="flex items-center gap-3">
                      <div className="w-1.5 h-1.5 rounded-full bg-memzent-purple shadow-[0_0_6px_rgba(157,0,255,0.5)]" />
                      <span className="text-[9px] font-black text-white/60 uppercase truncate w-24">{tool.name || tool.id}</span>
                    </div>
                    <span className="text-[8px] font-black font-mono text-white/20">{tool.relevance_score ? tool.relevance_score.toFixed(3) : 'N/A'}</span>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-[9px] font-black uppercase tracking-widest text-white/10 text-center py-6">
                {status === 'idle' ? 'No execution yet' : result?.cached ? 'Served from cache' : 'No tools matched'}
              </p>
            )}
          </div>

          {/* Active Memory Profile */}
          <div className="stat-card neural-bg border-white/5 p-6">
            <div className="flex items-center gap-3 mb-4">
              <BrainCircuit size={14} className="text-memzent-glow" />
              <h2 className="text-xs font-black uppercase tracking-widest text-white/60">Memory Context</h2>
            </div>
            <div className="p-3 rounded-xl bg-white/[0.02] border border-white/5 text-[9px] leading-relaxed text-white/50 space-y-2 font-bold">
              <div className="flex items-center gap-2 text-white/80">
                <span className="w-1.5 h-1.5 rounded-full bg-memzent-glow animate-pulse" />
                <span className="uppercase tracking-wider">Semantic Sync Status</span>
              </div>
              <p>
                As conversations progress, facts are automatically extracted and vectorized. In the next turn, relevant contexts are retrieved and appended to the model prompt automatically.
              </p>
              {result && (
                <div className="pt-2 border-t border-white/5 text-[8px] text-memzent-glow/80 font-mono">
                  SESSION_ID: {activeSessionId?.substring(0, 8)}...
                </div>
              )}
            </div>
          </div>
        </div>

      </div>
    </div>
  )
}
