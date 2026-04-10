"use client"

import { useState } from "react"
import { Plus, X, Globe, Database, Terminal, Shield, Loader2 } from "lucide-react"
import { registerAuraTool } from "@/app/actions"

export function RegisterToolBtn({ orgId }: { orgId?: string }) {
  const [isOpen, setIsOpen] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  
  const [formData, setFormData] = useState({
    id: "",
    name: "",
    description: "",
    connector_type: "mcp",
    endpoint: "",
    timeout_seconds: 15,
  })

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    setError(null)
    
    try {
      if (!orgId) throw new Error("Organization context missing")
      await registerAuraTool(orgId, formData)
      setIsOpen(false)
      // Refresh the page or update state
      window.location.reload()
    } catch (err: any) {
      setError(err.message || "Failed to register tool")
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <>
      <button 
        onClick={() => setIsOpen(true)}
        className="bg-aura-glow text-black px-6 py-3 rounded-2xl text-xs font-black tracking-widest uppercase hover:scale-105 transition-all shadow-[0_0_20px_rgba(0,243,255,0.2)] flex items-center gap-2"
      >
        <Plus size={14} />
        Register New Tool
      </button>

      {isOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={() => setIsOpen(false)} />
          
          <div className="stat-card w-full max-w-2xl bg-[#0a0a0c] border-white/10 p-0 overflow-hidden relative z-10 animate-in fade-in zoom-in duration-200">
            <header className="p-8 border-b border-white/5 flex items-center justify-between">
              <div>
                <h2 className="text-2xl font-black tracking-tighter uppercase italic">Register_Tool_v1</h2>
                <p className="text-[10px] font-bold text-white/20 uppercase tracking-widest mt-1">Provision New Semantic Node</p>
              </div>
              <button onClick={() => setIsOpen(false)} className="p-2 hover:bg-white/5 rounded-xl transition-colors text-white/40 hover:text-white">
                <X size={20} />
              </button>
            </header>

            <form onSubmit={handleSubmit} className="p-8 space-y-6">
              {error && (
                <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-500 text-xs font-bold uppercase tracking-widest flex items-center gap-2">
                  <Shield size={14} /> {error}
                </div>
              )}

              <div className="grid grid-cols-2 gap-6">
                <div className="space-y-2">
                  <label className="text-[10px] font-black uppercase text-white/30 tracking-widest ml-1">Unique Identifier</label>
                  <input 
                    required
                    placeholder="e.g. stripe_billing"
                    className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm font-bold focus:border-aura-glow outline-none transition-all"
                    value={formData.id}
                    onChange={(e) => setFormData({...formData, id: e.target.value})}
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-[10px] font-black uppercase text-white/30 tracking-widest ml-1">Display Name</label>
                  <input 
                    required
                    placeholder="e.g. Stripe Manager"
                    className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm font-bold focus:border-aura-glow outline-none transition-all"
                    value={formData.name}
                    onChange={(e) => setFormData({...formData, name: e.target.value})}
                  />
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-[10px] font-black uppercase text-white/30 tracking-widest ml-1">Description</label>
                <textarea 
                  required
                  placeholder="What does this tool enable the LLM to do?"
                  rows={2}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm font-bold focus:border-aura-glow outline-none transition-all resize-none"
                  value={formData.description}
                  onChange={(e) => setFormData({...formData, description: e.target.value})}
                />
              </div>

              <div className="grid grid-cols-2 gap-6">
                <div className="space-y-2">
                  <label className="text-[10px] font-black uppercase text-white/30 tracking-widest ml-1">Connector Type</label>
                  <select 
                    className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm font-bold focus:border-aura-glow outline-none transition-all appearance-none"
                    value={formData.connector_type}
                    onChange={(e) => setFormData({...formData, connector_type: e.target.value})}
                  >
                    <option value="mcp" className="bg-[#0a0a0c]">MCP (Model Context Protocol)</option>
                    <option value="rest" className="bg-[#0a0a0c]">REST API</option>
                    <option value="sql" className="bg-[#0a0a0c]">Direct SQL</option>
                  </select>
                </div>
                <div className="space-y-2">
                  <label className="text-[10px] font-black uppercase text-white/30 tracking-widest ml-1">Timeout (Seconds)</label>
                  <input 
                    type="number"
                    min="1"
                    className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm font-bold focus:border-aura-glow outline-none transition-all"
                    value={formData.timeout_seconds}
                    onChange={(e) => setFormData({...formData, timeout_seconds: parseInt(e.target.value)})}
                  />
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-[10px] font-black uppercase text-white/30 tracking-widest ml-1">Endpoint / Connection String</label>
                <div className="relative">
                  <div className="absolute left-4 top-1/2 -translate-y-1/2 text-white/20">
                    {formData.connector_type === 'mcp' && <Terminal size={16} />}
                    {formData.connector_type === 'rest' && <Globe size={16} />}
                    {formData.connector_type === 'sql' && <Database size={16} />}
                  </div>
                  <input 
                    required
                    placeholder={formData.connector_type === 'mcp' ? "Tool ID within MCP" : "https://api.service.com/v1" }
                    className="w-full bg-white/5 border border-white/10 rounded-xl pl-12 pr-4 py-3 text-sm font-bold focus:border-aura-glow outline-none transition-all"
                    value={formData.endpoint}
                    onChange={(e) => setFormData({...formData, endpoint: e.target.value})}
                  />
                </div>
              </div>

              <div className="flex items-center justify-end gap-4 pt-4">
                 <button 
                  type="button"
                  onClick={() => setIsOpen(false)}
                  className="px-6 py-3 rounded-xl text-[10px] font-black uppercase tracking-widest hover:bg-white/5 transition-all text-white/30 hover:text-white"
                 >
                   Cancel
                 </button>
                 <button 
                  type="submit"
                  disabled={isSubmitting}
                  className="bg-aura-glow text-black px-8 py-3 rounded-2xl text-[10px] font-black uppercase tracking-widest hover:scale-105 transition-all shadow-[0_0_20px_rgba(0,243,255,0.2)] flex items-center gap-2 disabled:opacity-50 disabled:scale-100"
                 >
                   {isSubmitting ? (
                     <>
                        <Loader2 size={14} className="animate-spin" />
                        Synchronizing...
                     </>
                   ) : (
                     "Authorize Node"
                   )}
                 </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </>
  )
}
