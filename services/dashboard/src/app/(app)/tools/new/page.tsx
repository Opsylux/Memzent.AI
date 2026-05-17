'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Globe, Database, Shield, ArrowRight, Save, Info, Plus, Trash2 } from 'lucide-react'
import { createMemzentTool } from '../../../actions'
import { supabase } from '@/lib/supabase'

export default function NewToolPage() {
  const router = useRouter()
  const [step, setStep] = useState(1)
  const [loading, setLoading] = useState(false)
  const [orgId, setOrgId] = useState<string | null>(null)

  const [formData, setFormData] = useState({
    id: '',
    name: '',
    description: '',
    connector_type: 'rest',
    endpoint: '',
    config: {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      query: ''
    }
  })

  useEffect(() => {
    async function load() {
      const { data: { user } } = await supabase.auth.getUser()
      if (user) {
        setOrgId(user.id) // Mocking org_id as user_id
      }
    }
    load()
  }, [])

  const handleSubmit = async () => {
    if (!orgId) return
    setLoading(true)
    try {
      await createMemzentTool(orgId, formData)
      router.push('/dashboard') // Or to a specific tools page if we add one
    } catch (e) {
      alert('Failed to provision tool: ' + (e as Error).message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-4xl mx-auto">
      <header className="mb-12">
        <div className="flex items-center gap-4 mb-4">
          <Button variant="ghost" size="sm" onClick={() => router.back()} className="text-white/40 hover:text-white">← Back</Button>
          <Badge className="bg-memzent-glow/10 text-memzent-glow border-memzent-glow/20 px-3 uppercase font-black text-[10px]">Registry v2.0</Badge>
        </div>
        <h1 className="text-4xl font-black tracking-tighter text-white mb-2 italic">
          PROVISION NEW ARMAMENT
        </h1>
        <p className="text-slate-400 font-bold uppercase tracking-widest text-xs">
          Connect external infrastructure to the Memzent AI Mesh.
        </p>
      </header>

      <div className="space-y-8">
        {/* Step 1: Identity */}
        <Card className={`border-white/5 bg-white/5 transition-opacity ${step !== 1 ? 'opacity-40 grayscale pointer-events-none' : ''}`}>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg">
              <span className="w-6 h-6 rounded-lg bg-white/10 flex items-center justify-center text-[10px] font-black">01</span>
              Tool Identity
            </CardTitle>
          </CardHeader>
          <CardContent className="grid grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-[10px] font-black uppercase tracking-widest text-white/40">Unique ID (lowercase, no spaces)</label>
              <input
                className="w-full bg-black/40 border border-white/10 rounded-xl px-4 py-3 text-sm focus:border-memzent-glow outline-none transition-all"
                placeholder="e.g. search_internal_docs"
                value={formData.id}
                onChange={(e) => setFormData({ ...formData, id: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-black uppercase tracking-widest text-white/40">Display Name</label>
              <input
                className="w-full bg-black/40 border border-white/10 rounded-xl px-4 py-3 text-sm focus:border-memzent-glow outline-none transition-all"
                placeholder="Search Internal Docs"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              />
            </div>
            <div className="space-y-2 col-span-2">
              <label className="text-[10px] font-black uppercase tracking-widest text-white/40">Functional Purpose (Hint for LLM Routing)</label>
              <textarea
                className="w-full bg-black/40 border border-white/10 rounded-xl px-4 py-3 text-sm focus:border-memzent-glow outline-none transition-all h-20 resize-none"
                placeholder="Executes semantic search across internal PDF knowledge base..."
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              />
            </div>
          </CardContent>
          {step === 1 && (
            <CardFooter>
              <Button onClick={() => setStep(2)} className="bg-white text-black font-black uppercase tracking-widest text-xs px-8 py-6 rounded-2xl hover:bg-memzent-glow transition-all flex items-center gap-2">
                Continue Configuration <ArrowRight size={14} />
              </Button>
            </CardFooter>
          )}
        </Card>

        {/* Step 2: Protocol Selection */}
        <Card className={`border-white/5 bg-white/5 transition-opacity ${step !== 2 ? 'opacity-40 grayscale pointer-events-none' : ''}`}>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg">
              <span className="w-6 h-6 rounded-lg bg-white/10 flex items-center justify-center text-[10px] font-black">02</span>
              Protocol Selection
            </CardTitle>
          </CardHeader>
          <CardContent className="grid grid-cols-2 gap-4">
            <div
              onClick={() => setFormData({ ...formData, connector_type: 'rest' })}
              className={`p-6 rounded-3xl border cursor-pointer transition-all ${formData.connector_type === 'rest' ? 'border-memzent-glow bg-memzent-glow/5' : 'border-white/5 hover:border-white/20'}`}
            >
              <Globe size={32} className={`mb-4 ${formData.connector_type === 'rest' ? 'text-memzent-glow' : 'text-white/20'}`} />
              <div className="font-black text-sm uppercase mb-1">REST API</div>
              <p className="text-[10px] text-white/40 font-bold leading-relaxed">Connect to any JSON-based external web service via HTTP.</p>
            </div>
            <div
              onClick={() => setFormData({ ...formData, connector_type: 'sql' })}
              className={`p-6 rounded-3xl border cursor-pointer transition-all ${formData.connector_type === 'sql' ? 'border-memzent-accent bg-memzent-accent/5' : 'border-white/5 hover:border-white/20'}`}
            >
              <Database size={32} className={`mb-4 ${formData.connector_type === 'sql' ? 'text-memzent-accent' : 'text-white/20'}`} />
              <div className="font-black text-sm uppercase mb-1">Direct SQL</div>
              <p className="text-[10px] text-white/40 font-bold leading-relaxed">Execute secure prepared statements against a Postgres database.</p>
            </div>
          </CardContent>
          {step === 2 && (
            <CardFooter className="flex justify-between">
              <Button variant="ghost" onClick={() => setStep(1)} className="text-white/40 font-black">Back</Button>
              <Button onClick={() => setStep(3)} className="bg-white text-black font-black uppercase tracking-widest text-xs px-8 py-6 rounded-2xl hover:bg-memzent-glow transition-all flex items-center gap-2">
                Define Connection <ArrowRight size={14} />
              </Button>
            </CardFooter>
          )}
        </Card>

        {/* Step 3: Connection Details */}
        <Card className={`border-white/5 bg-white/5 transition-opacity ${step !== 3 ? 'opacity-40 grayscale pointer-events-none' : ''}`}>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg">
              <span className="w-6 h-6 rounded-lg bg-white/10 flex items-center justify-center text-[10px] font-black">03</span>
              Connection Details
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {formData.connector_type === 'rest' ? (
              <div className="space-y-6">
                <div className="space-y-2">
                  <label className="text-[10px] font-black uppercase tracking-widest text-white/40">Endpoint URL</label>
                  <input
                    className="w-full bg-black/40 border border-white/10 rounded-xl px-4 py-3 text-sm font-mono focus:border-memzent-glow outline-none transition-all"
                    placeholder="https://api.example.com/v1/search"
                    value={formData.endpoint}
                    onChange={(e) => setFormData({ ...formData, endpoint: e.target.value })}
                  />
                </div>
                <div className="p-4 border border-blue-500/20 bg-blue-500/5 rounded-2xl flex gap-4">
                  <Info size={18} className="text-blue-400 shrink-0 mt-1" />
                  <p className="text-[10px] text-blue-400/80 font-bold leading-relaxed uppercase">
                    Memzent will send a POST request to this endpoint with the LLM's parsed arguments in the JSON body.
                  </p>
                </div>
              </div>
            ) : (
              <div className="space-y-6">
                <div className="space-y-2">
                  <label className="text-[10px] font-black uppercase tracking-widest text-white/40">Postgres Connection String</label>
                  <input
                    type="password"
                    className="w-full bg-black/40 border border-white/10 rounded-xl px-4 py-3 text-sm font-mono focus:border-memzent-accent outline-none transition-all"
                    placeholder="postgres://user:pass@host:port/db"
                    value={formData.endpoint}
                    onChange={(e) => setFormData({ ...formData, endpoint: e.target.value })}
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-[10px] font-black uppercase tracking-widest text-white/40">Prepared Query</label>
                  <textarea
                    className="w-full bg-black/40 border border-white/10 rounded-xl px-4 py-3 text-sm font-mono focus:border-memzent-accent outline-none transition-all h-24 resize-none"
                    placeholder="SELECT * FROM products WHERE category = $1"
                    value={formData.config.query}
                    onChange={(e) => setFormData({ ...formData, config: { ...formData.config, query: e.target.value } })}
                  />
                </div>
              </div>
            )}
          </CardContent>
          {step === 3 && (
            <CardFooter className="flex justify-between">
              <Button variant="ghost" onClick={() => setStep(2)} className="text-white/40 font-black">Back</Button>
              <Button
                onClick={handleSubmit}
                disabled={loading || !formData.id || !formData.endpoint}
                className="bg-memzent-glow text-black font-black uppercase tracking-widest text-xs px-10 py-6 rounded-2xl hover:shadow-[0_0_20px_rgba(0,243,255,0.4)] transition-all flex items-center gap-2"
              >
                {loading ? 'Initializing...' : <><Save size={16} /> Deploy Tool</>}
              </Button>
            </CardFooter>
          )}
        </Card>
      </div>

      <section className="mt-12 flex items-center gap-4 p-6 bg-white/[0.02] border border-white/5 rounded-3xl text-sm font-bold opacity-40 italic">
        <Shield size={18} />
        Provisioned tools are isolated to your organization and require valid MEMZENT-SC-01 security clearace for execution.
      </section>
    </div>
  )
}
