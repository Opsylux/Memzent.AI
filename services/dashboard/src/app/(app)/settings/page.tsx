'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Badge } from "@/components/ui/badge";
import { Settings, User, Shield, Bell, Zap, Save, AlertTriangle, Activity } from 'lucide-react'
import { updateOrgProfile, getOrgProfile, getMemzentProviders, getSimilarityThreshold, updateSimilarityThreshold } from '../../actions'
import { supabase } from '@/lib/supabase'

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState<'profile' | 'security' | 'api' | 'alerts'>('profile')
  const [loading, setLoading] = useState(false)
  const [orgId, setOrgId] = useState<string | null>(null)
  const [orgName, setOrgName] = useState('Memzent Global HQ')
  const [contactEmail, setContactEmail] = useState('ops@memzent.io')
  const [saved, setSaved] = useState(false)
  const [members, setMembers] = useState<any[]>([])
  const [defaultProvider, setDefaultProvider] = useState('')
  const [defaultModel, setDefaultModel] = useState('')
  const [providers, setProviders] = useState<any[]>([])
  const [similarityThreshold, setSimilarityThreshold] = useState(0.88)
  const [thresholdSaving, setThresholdSaving] = useState(false)
  const [thresholdSaved, setThresholdSaved] = useState(false)

  useEffect(() => {
    async function load() {
      const { data: { user } } = await supabase.auth.getUser()
      if (user) {
        // Resolve org membership
        const { data: membership } = await supabase
          .from('members')
          .select('org_id, organizations(id, name)')
          .eq('user_id', user.id)
          .limit(1)
          .maybeSingle()

        if (membership?.organizations) {
          const org = membership.organizations as any
          setOrgId(org.id)
          setOrgName(org.name || '')

          // Load members
          const { data: memberList } = await supabase
            .from('members')
            .select('*')
            .eq('org_id', org.id)
          setMembers(memberList || [])

          // Try loading full profile
          try {
            const profile = await getOrgProfile(org.id)
            if (profile) {
              setOrgName(profile.name || '')
              setContactEmail(profile.contact_email || user.email || '')
              setDefaultProvider(profile.default_provider || '')
              setDefaultModel(profile.default_model || '')
            }
          } catch { }

          // Load providers
          try {
            const provs = await getMemzentProviders()
            setProviders(provs || [])
          } catch (e) {
            console.error('Failed to load providers:', e)
          }

          // Load similarity threshold
          try {
            const threshold = await getSimilarityThreshold(org.id)
            setSimilarityThreshold(threshold)
          } catch { }
        } else {
          setOrgId(user.id)
          setOrgName(user.user_metadata?.full_name || user.email?.split('@')[0] || 'Personal')
          setContactEmail(user.email || '')
        }
      }
    }
    load()
  }, [])

  const handleSave = async () => {
    if (!orgId) return
    setLoading(true)
    try {
      await updateOrgProfile(orgId, {
        name: orgName,
        contact_email: contactEmail,
        default_provider: defaultProvider || null,
        default_model: defaultModel || null
      })
      setSaved(true)
      setTimeout(() => setSaved(false), 3000)
    } catch (e) {
      console.error('Failed to update org profile:', e)
    }
    setLoading(false)
  }

  return (
    <div className="space-y-12 pb-20">
      <header className="mb-12">
        <h1 className="text-4xl font-black tracking-tighter text-white mb-2 uppercase italic">
          Settings
        </h1>
        <p className="text-white/50 font-black uppercase tracking-[0.3em] text-[10px] italic">
          {orgName ? `${orgName} — ` : ''}Organization & Configuration
        </p>
      </header>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-12">
        {/* Navigation tabs */}
        <aside className="lg:col-span-1 space-y-2">
          <Button
            variant="ghost"
            onClick={() => setActiveTab('profile')}
            className={`w-full justify-start gap-4 rounded-2xl py-6 font-black uppercase text-[10px] tracking-widest transition-all ${activeTab === 'profile' ? 'bg-white/5 text-memzent-glow border border-memzent-glow/20 shadow-[0_0_15px_rgba(0,243,255,0.1)]' : 'text-white/55 hover:text-white hover:bg-white/5'
              }`}
          >
            <User size={18} /> Profile & Org
          </Button>
          <Button
            variant="ghost"
            onClick={() => setActiveTab('security')}
            className={`w-full justify-start gap-4 rounded-2xl py-6 font-black uppercase text-[10px] tracking-widest group transition-all ${activeTab === 'security' ? 'bg-white/5 text-memzent-purple border border-memzent-purple/20 shadow-[0_0_15px_rgba(157,0,255,0.1)]' : 'text-white/30 hover:text-white hover:bg-white/5'
              }`}
          >
            <Shield size={18} className={activeTab === 'security' ? "text-memzent-purple" : "group-hover:text-memzent-purple"} /> Security & RBAC
          </Button>
          <Button
            variant="ghost"
            onClick={() => setActiveTab('api')}
            className={`w-full justify-start gap-4 rounded-2xl py-6 font-black uppercase text-[10px] tracking-widest group transition-all ${activeTab === 'api' ? 'bg-white/5 text-memzent-glow border border-memzent-glow/20 shadow-[0_0_15px_rgba(0,243,255,0.1)]' : 'text-white/30 hover:text-white hover:bg-white/5'
              }`}
          >
            <Zap size={18} className={activeTab === 'api' ? "text-memzent-glow" : "group-hover:text-memzent-glow"} /> API Integration
          </Button>
          <Button
            variant="ghost"
            onClick={() => setActiveTab('alerts')}
            className={`w-full justify-start gap-4 rounded-2xl py-6 font-black uppercase text-[10px] tracking-widest group transition-all ${activeTab === 'alerts' ? 'bg-white/5 text-memzent-accent border border-memzent-accent/20 shadow-[0_0_15px_rgba(0,277,142,0.1)]' : 'text-white/30 hover:text-white hover:bg-white/5'
              }`}
          >
            <Bell size={18} className={activeTab === 'alerts' ? "text-memzent-accent" : "group-hover:text-memzent-accent"} /> Alerts & Webhooks
          </Button>
        </aside>

        {/* Main Content */}
        <div className="lg:col-span-3 space-y-8">
          {activeTab === 'profile' && (
            <>
              <div className="stat-card glow-purple p-8 neural-bg border-white/5 overflow-hidden relative">
                <div className="flex items-center gap-6 mb-8 pb-6 border-b border-white/5">
                  <div className="w-14 h-14 rounded-2xl bg-memzent-purple/10 border border-memzent-purple/20 flex items-center justify-center text-memzent-purple shadow-[0_0_15px_rgba(157,0,255,0.2)]">
                    <User size={28} />
                  </div>
                  <div>
                    <h3 className="text-xl font-black tracking-tight uppercase italic leading-none">Organization Profile</h3>
                    <p className="text-[10px] font-bold text-white/50 uppercase tracking-widest mt-1">Managed Neural Entity Identity</p>
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-10">
                  <div className="space-y-3">
                    <label className="text-[10px] font-black uppercase tracking-widest text-white/40 italic pl-1">Entity Name</label>
                    <input
                      className="w-full bg-black/40 border border-white/10 rounded-2xl px-5 py-4 text-sm font-bold text-white focus:border-memzent-glow outline-none transition-all placeholder:text-white/10 shadow-inner"
                      value={orgName}
                      onChange={(e) => setOrgName(e.target.value)}
                    />
                  </div>
                  <div className="space-y-3">
                    <label className="text-[10px] font-black uppercase tracking-widest text-white/40 italic pl-1">Technical Contact</label>
                    <input
                      className="w-full bg-black/40 border border-white/10 rounded-2xl px-5 py-4 text-sm font-bold text-white focus:border-memzent-glow outline-none transition-all placeholder:text-white/10 shadow-inner"
                      value={contactEmail}
                      onChange={(e) => setContactEmail(e.target.value)}
                    />
                  </div>
                </div>

                <div className="flex items-center gap-6 mb-8 pb-6 border-b border-white/5 mt-12">
                  <div className="w-14 h-14 rounded-2xl bg-memzent-glow/10 border border-memzent-glow/20 flex items-center justify-center text-memzent-glow shadow-[0_0_15px_rgba(0,243,255,0.2)]">
                    <Zap size={28} />
                  </div>
                  <div>
                    <h3 className="text-xl font-black tracking-tight uppercase italic leading-none">Neural Model Routing</h3>
                    <p className="text-[10px] font-bold text-white/50 uppercase tracking-widest mt-1">Default Fallback Provider & Model Preferences</p>
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-10">
                  <div className="space-y-3">
                    <label className="text-[10px] font-black uppercase tracking-widest text-white/40 italic pl-1">Default Provider</label>
                    <select
                      className="w-full bg-black/40 border border-white/10 rounded-2xl px-5 py-4 text-sm font-bold text-white focus:border-memzent-glow outline-none transition-all shadow-inner"
                      value={defaultProvider}
                      onChange={(e) => {
                        setDefaultProvider(e.target.value)
                        const prov = providers.find(p => p.name === e.target.value)
                        if (prov) {
                          setDefaultModel(prov.default_model || prov.supported_models?.[0] || '')
                        } else {
                          setDefaultModel('')
                        }
                      }}
                    >
                      <option value="" className="bg-[#141414] text-white/60">System default (configured by admin)</option>
                      {providers.map((p) => (
                        <option key={p.name} value={p.name} className="bg-[#141414] text-white">
                          {p.name.toUpperCase()}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div className="space-y-3">
                    <label className="text-[10px] font-black uppercase tracking-widest text-white/40 italic pl-1">Default Model</label>
                    <select
                      className="w-full bg-black/40 border border-white/10 rounded-2xl px-5 py-4 text-sm font-bold text-white focus:border-memzent-glow outline-none transition-all shadow-inner"
                      value={defaultModel}
                      onChange={(e) => setDefaultModel(e.target.value)}
                      disabled={!defaultProvider}
                    >
                      {!defaultProvider ? (
                        <option value="" className="bg-[#141414] text-white/60">Select a provider first</option>
                      ) : (
                        <>
                          <option value="" className="bg-[#141414] text-white/60">Provider default model</option>
                          {(providers.find(p => p.name === defaultProvider)?.supported_models || []).map((m: string) => (
                            <option key={m} value={m} className="bg-[#141414] text-white">
                              {m}
                            </option>
                          ))}
                        </>
                      )}
                    </select>
                  </div>
                </div>

                <div className="flex items-center justify-end gap-4 pt-4 border-t border-white/5">
                  {saved && (
                    <span className="text-[10px] font-black text-memzent-accent uppercase tracking-widest animate-pulse">Changes Synchronized</span>
                  )}
                  <Button
                    onClick={handleSave}
                    disabled={loading}
                    className="bg-memzent-glow text-black font-black uppercase tracking-[0.3em] text-[10px] px-8 h-14 rounded-2xl hover:scale-102 hover:shadow-[0_0_20px_rgba(0,243,255,0.3)] transition-all"
                  >
                    <Save size={16} className="mr-2" /> {loading ? 'Syncing...' : 'Sync Changes'}
                  </Button>
                </div>

                <div className="absolute inset-0 pointer-events-none opacity-[0.02] grayscale bg-[url('https://grainy-gradients.vercel.app/noise.svg')]" />
              </div>

              <div className="stat-card border-red-500/10 bg-red-500/[0.02] p-8 neural-bg">
                <div className="flex items-center gap-6 mb-8">
                  <div className="w-14 h-14 rounded-2xl bg-red-500/10 border border-red-500/20 flex items-center justify-center text-red-500">
                    <AlertTriangle size={28} />
                  </div>
                  <div>
                    <h3 className="text-xl font-black tracking-tight uppercase italic leading-none text-red-500/80">Terminus Zone</h3>
                    <p className="text-[10px] font-bold text-red-500/30 uppercase tracking-widest mt-1 italic font-black">Permanent Infrastructure Dissolution</p>
                  </div>
                </div>

                <p className="text-[10px] text-white/60 mb-8 uppercase font-bold leading-relaxed max-w-xl tracking-widest">
                  Deactivating the organization sector will purge all semantic clusters, vector points, and provisioned tool bindings. This action will initiate an IRREVERSIBLE data scrubbing protocol.
                </p>

                <Button variant="outline" className="border-red-500/20 text-red-500 font-black uppercase tracking-[0.3em] text-[10px] px-8 h-14 rounded-2xl hover:bg-red-500/10 hover:border-red-500/40 transition-all" onClick={() => alert("Dissolution currently restricted to manual database intervention during Beta.")}>
                  Execute Dissolution
                </Button>
              </div>
            </>
          )}

          {activeTab === 'security' && (
            <div className="stat-card glow-purple p-8 neural-bg border-white/5">
              <div className="flex items-center gap-6 mb-8 pb-6 border-b border-white/5">
                <div className="w-14 h-14 rounded-2xl bg-memzent-purple/10 border border-memzent-purple/20 flex items-center justify-center text-memzent-purple">
                  <Shield size={28} />
                </div>
                <div>
                  <h3 className="text-xl font-black tracking-tight uppercase italic leading-none">Members & Permissions</h3>
                  <p className="text-[10px] font-bold text-white/50 uppercase tracking-widest mt-1">Organization Access Control</p>
                </div>
              </div>

              {/* Rate limit breakdown by role */}
              <div className="mb-8 p-6 rounded-2xl bg-white/[0.02] border border-white/5">
                <h4 className="text-[10px] font-black uppercase tracking-widest text-white/50 mb-4">Rate Limits by Role</h4>
                <div className="grid grid-cols-3 gap-4">
                  <div className="p-3 rounded-xl bg-white/[0.03] border border-white/5 text-center">
                    <div className="text-[9px] font-black text-white/40 uppercase mb-1">Viewer</div>
                    <div className="text-sm font-black text-white/60">Read Only</div>
                    <div className="text-[9px] text-white/30 mt-1">No prompt execution</div>
                  </div>
                  <div className="p-3 rounded-xl bg-white/[0.03] border border-memzent-glow/10 text-center">
                    <div className="text-[9px] font-black text-memzent-glow/60 uppercase mb-1">Member</div>
                    <div className="text-sm font-black text-memzent-glow">50% of org limit</div>
                    <div className="text-[9px] text-white/30 mt-1">Shared org balance</div>
                  </div>
                  <div className="p-3 rounded-xl bg-white/[0.03] border border-memzent-purple/10 text-center">
                    <div className="text-[9px] font-black text-memzent-purple/60 uppercase mb-1">Admin</div>
                    <div className="text-sm font-black text-memzent-purple">Full org limit</div>
                    <div className="text-[9px] text-white/30 mt-1">Full access</div>
                  </div>
                </div>
              </div>

              <div className="divide-y divide-white/5">
                {members.map((m) => (
                  <div key={m.id} className="py-6 flex items-center justify-between group">
                    <div className="flex items-center gap-4">
                      <div className="w-10 h-10 rounded-full bg-white/5 border border-white/10 flex items-center justify-center text-[10px] font-black text-white/40 uppercase">
                        {m.user_id.substring(0, 2)}
                      </div>
                      <div>
                        <p className="text-sm font-black text-white italic">{m.user_id}</p>
                        <p className="text-[10px] font-bold text-white/45 uppercase tracking-widest mt-0.5">{m.role}</p>
                      </div>
                    </div>
                    <Badge variant="outline" className={`uppercase text-[9px] font-black tracking-widest px-3 ${
                      m.role === 'admin' || m.role === 'owner'
                        ? 'border-memzent-purple/20 text-memzent-purple'
                        : m.role === 'viewer'
                          ? 'border-white/10 text-white/40'
                          : 'border-memzent-glow/20 text-memzent-glow'
                    }`}>
                      {m.role}
                    </Badge>
                  </div>
                ))}

                {members.length === 0 && (
                  <div className="py-12 text-center text-white/40 font-black uppercase tracking-widest text-[10px]">
                    No members found
                  </div>
                )}
              </div>

              <Button className="w-full mt-8 py-6 rounded-2xl bg-white/5 border border-white/10 text-white/40 hover:text-white hover:bg-white/10 text-[10px] font-black uppercase tracking-widest transition-all" onClick={() => alert("Member invitations coming in the next release.")}>
                Invite Member
              </Button>
            </div>
          )}

          {activeTab === 'api' && (
            <div className="stat-card glow-cyan p-8 neural-bg border-white/5 overflow-hidden relative">
              <div className="flex items-center gap-6 mb-8 pb-6 border-b border-white/5">
                <div className="w-14 h-14 rounded-2xl bg-memzent-glow/10 border border-memzent-glow/20 flex items-center justify-center text-memzent-glow shadow-[0_0_15px_rgba(0,243,255,0.2)]">
                  <Activity size={28} />
                </div>
                <div>
                  <h3 className="text-xl font-black tracking-tight uppercase italic leading-none">Semantic Precision Control</h3>
                  <p className="text-[10px] font-bold text-white/50 uppercase tracking-widest mt-1">Dynamic Similarity Threshold Configuration</p>
                </div>
              </div>

              <div className="space-y-8">
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <label className="text-[10px] font-black uppercase tracking-widest text-white/40 italic">Cosine Similarity Threshold</label>
                    <span className="text-lg font-black text-memzent-glow tabular-nums">{similarityThreshold.toFixed(2)}</span>
                  </div>
                  <input
                    type="range"
                    min="0.50"
                    max="0.99"
                    step="0.01"
                    value={similarityThreshold}
                    onChange={(e) => setSimilarityThreshold(parseFloat(e.target.value))}
                    className="w-full h-2 bg-white/5 rounded-full appearance-none cursor-pointer accent-memzent-glow"
                  />
                  <div className="flex justify-between text-[9px] font-bold text-white/25 uppercase tracking-widest">
                    <span>Broader Matches (0.50)</span>
                    <span>Exact Only (0.99)</span>
                  </div>
                </div>

                <div className="grid grid-cols-3 gap-4">
                  {[
                    { label: 'Relaxed', value: 0.70, desc: 'More cache hits, less precise' },
                    { label: 'Balanced', value: 0.88, desc: 'Default — recommended' },
                    { label: 'Strict', value: 0.95, desc: 'Fewer hits, maximum precision' },
                  ].map((preset) => (
                    <button
                      key={preset.label}
                      onClick={() => setSimilarityThreshold(preset.value)}
                      className={`p-4 rounded-2xl border transition-all text-center ${
                        Math.abs(similarityThreshold - preset.value) < 0.02
                          ? 'border-memzent-glow/40 bg-memzent-glow/5 shadow-[0_0_15px_rgba(0,243,255,0.1)]'
                          : 'border-white/5 bg-white/[0.02] hover:border-white/10'
                      }`}
                    >
                      <div className="text-[10px] font-black uppercase tracking-widest text-white/70">{preset.label}</div>
                      <div className="text-lg font-black text-memzent-glow mt-1">{preset.value}</div>
                      <div className="text-[8px] text-white/30 mt-1 uppercase tracking-wider">{preset.desc}</div>
                    </button>
                  ))}
                </div>

                <div className="flex items-center justify-end gap-4 pt-4 border-t border-white/5">
                  {thresholdSaved && (
                    <span className="text-[10px] font-black text-memzent-accent uppercase tracking-widest animate-pulse">Threshold Synchronized</span>
                  )}
                  <Button
                    onClick={async () => {
                      if (!orgId) return
                      setThresholdSaving(true)
                      try {
                        await updateSimilarityThreshold(orgId, similarityThreshold)
                        setThresholdSaved(true)
                        setTimeout(() => setThresholdSaved(false), 3000)
                      } catch (e) {
                        console.error('Failed to update threshold:', e)
                      }
                      setThresholdSaving(false)
                    }}
                    disabled={thresholdSaving}
                    className="bg-memzent-glow text-black font-black uppercase tracking-[0.3em] text-[10px] px-8 h-14 rounded-2xl hover:scale-102 hover:shadow-[0_0_20px_rgba(0,243,255,0.3)] transition-all"
                  >
                    <Save size={16} className="mr-2" /> {thresholdSaving ? 'Syncing...' : 'Apply Threshold'}
                  </Button>
                </div>
              </div>

              <div className="absolute inset-0 pointer-events-none opacity-[0.02] grayscale bg-[url('https://grainy-gradients.vercel.app/noise.svg')]" />
            </div>
          )}

          {activeTab === 'alerts' && (
            <div className="stat-card p-20 flex flex-col items-center justify-center text-center neural-bg border-dashed border-white/10">
              <div className="w-20 h-20 rounded-full bg-white/5 border border-white/5 flex items-center justify-center mb-6 text-white/10">
                <Bell size={40} />
              </div>
              <h3 className="text-lg font-black tracking-tighter uppercase italic text-white/45">Webhooks & Alerts</h3>
              <p className="text-[10px] font-bold text-white/35 uppercase tracking-[0.3em] mt-2 mb-8">Phase 7 Notification Pipeline Pending</p>
              <Badge variant="outline" className="border-memzent-glow/20 text-memzent-glow/40 uppercase text-[9px] font-black tracking-tighter italic">Coming Soon</Badge>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
