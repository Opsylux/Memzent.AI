"use client"

import { useState, useEffect, useRef } from "react"
import { Send, Bot, User, Loader2, X, MessageSquare, Zap, ShieldCheck } from "lucide-react"
import { executeAuraPrompt } from "@/app/actions"

interface Message {
  role: 'user' | 'assistant'
  content: string
}

export function NeuralAssistant({ orgId }: { orgId?: string }) {
  const [isOpen, setIsOpen] = useState(false)
  const [messages, setMessages] = useState<Message[]>([
    { role: 'assistant', content: "INTELLIGENCE_MESH_ONLINE. How can I assist with your neural tool integration?" }
  ])
  const [input, setInput] = useState("")
  const [loading, setLoading] = useState(false)
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [messages])

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim() || loading) return

    const userMsg = input.trim()
    setInput("")
    setMessages(prev => [...prev, { role: 'user', content: userMsg }])
    setLoading(true)

    try {
      // For now, we reuse the existing executeAuraPrompt
      // but later we can add a specific 'assistant' flag/intent
      const res = await executeAuraPrompt(userMsg, orgId)
      console.log("Chat response:", res)
      setMessages(prev => [...prev, { role: 'assistant', content: res.text || JSON.stringify(res) }])
    } catch (err: any) {
      setMessages(prev => [...prev, { role: 'assistant', content: `CRITICAL_ERROR: ${err.message}` }])
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      {/* Floating Trigger */}
      <button 
        onClick={() => setIsOpen(true)}
        className="fixed bottom-8 right-8 w-16 h-16 rounded-full bg-gradient-to-br from-aura-purple to-aura-glow flex items-center justify-center text-black shadow-[0_0_30px_rgba(0,243,255,0.3)] hover:scale-110 transition-all z-50 group"
      >
        <MessageSquare size={24} className="group-hover:rotate-12 transition-transform" />
      </button>

      {/* Chat Window */}
      {isOpen && (
        <div className="fixed bottom-32 right-8 w-96 h-[600px] glass bg-aura-dark/95 border-white/10 rounded-3xl z-50 flex flex-col overflow-hidden shadow-2xl animate-in slide-in-from-bottom-5 duration-300">
          <header className="p-6 border-b border-white/5 bg-white/[0.02] flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-aura-purple/20 flex items-center justify-center text-aura-purple">
                <Bot size={18} />
              </div>
              <div>
                <h3 className="text-xs font-black uppercase tracking-widest italic">Neural_Assistant</h3>
                <div className="flex items-center gap-2 mt-0.5">
                  <div className="w-1 h-1 rounded-full bg-aura-accent animate-pulse" />
                  <span className="text-[9px] font-bold text-white/20 uppercase tracking-[0.2em]">Active_Co-Pilot</span>
                </div>
              </div>
            </div>
            <button onClick={() => setIsOpen(false)} className="text-white/20 hover:text-white transition-colors">
              <X size={20} />
            </button>
          </header>

          <div 
            ref={scrollRef}
            className="flex-1 overflow-y-auto p-6 space-y-6 scrollbar-hide"
          >
            {messages.map((msg, idx) => (
              <div key={idx} className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
                <div className={`max-w-[85%] flex gap-3 ${msg.role === 'user' ? 'flex-row-reverse' : ''}`}>
                  <div className={`w-8 h-8 rounded-lg flex-shrink-0 flex items-center justify-center ${
                    msg.role === 'user' ? 'bg-aura-glow/10 text-aura-glow' : 'bg-white/5 text-white/40'
                  }`}>
                    {msg.role === 'user' ? <User size={16} /> : <Bot size={16} />}
                  </div>
                  <div className={`p-4 rounded-2xl text-xs font-bold leading-relaxed ${
                    msg.role === 'user' 
                      ? 'bg-aura-glow/10 border border-aura-glow/20 text-white rounded-tr-none shadow-[0_0_15px_rgba(0,243,255,0.05)]' 
                      : 'bg-white/5 border border-white/5 text-white/60 rounded-tl-none'
                  }`}>
                    {msg.content}
                  </div>
                </div>
              </div>
            ))}
            {loading && (
              <div className="flex justify-start">
                <div className="flex gap-3">
                  <div className="w-8 h-8 rounded-lg bg-white/5 flex items-center justify-center text-white/20">
                    <Loader2 size={16} className="animate-spin" />
                  </div>
                  <div className="p-4 rounded-2xl bg-white/5 border border-white/5 text-[10px] uppercase font-black tracking-widest text-white/20 animate-pulse">
                    Thinking...
                  </div>
                </div>
              </div>
            )}
          </div>

          <form onSubmit={handleSend} className="p-6 border-t border-white/5 bg-white/[0.02]">
            <div className="relative group">
              <input 
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder="Ask about tools, endpoints, or keys..."
                className="w-full h-12 bg-black/40 border border-white/10 rounded-xl pl-4 pr-12 text-xs font-bold text-white focus:border-aura-glow outline-none transition-all"
              />
              <button 
                type="submit"
                disabled={loading || !input.trim()}
                className="absolute right-2 top-1/2 -translate-y-1/2 w-8 h-8 bg-aura-glow rounded-lg flex items-center justify-center text-black hover:scale-105 transition-all disabled:opacity-30 disabled:scale-100 disabled:grayscale"
              >
                <Send size={14} />
              </button>
            </div>
            <div className="mt-4 flex items-center justify-between text-[9px] font-black uppercase tracking-widest text-white/10 px-1">
              <div className="flex items-center gap-1.5 hover:text-aura-purple transition-colors cursor-help">
                <Zap size={10} />
                <span>Low Latency</span>
              </div>
              <div className="flex items-center gap-1.5 hover:text-aura-glow transition-colors cursor-help">
                <ShieldCheck size={10} />
                <span>Encrypted</span>
              </div>
            </div>
          </form>

          <div className="absolute inset-0 pointer-events-none opacity-[0.03] grayscale bg-[url('https://grainy-gradients.vercel.app/noise.svg')]" />
        </div>
      )}
    </>
  )
}
