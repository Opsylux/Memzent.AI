'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Settings, User, Shield, Bell, Zap, Save, AlertTriangle } from 'lucide-react'
import { createClient } from '@supabase/supabase-js'
import { updateOrgProfile, getOrgProfile } from '../../actions'

const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL || ''
const supabaseAnonKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY || ''
const supabase = createClient(supabaseUrl, supabaseAnonKey)

export default function SettingsPage() {
  const [loading, setLoading] = useState(false)
  const [orgId, setOrgId] = useState<string | null>(null)
  const [orgName, setOrgName] = useState('Aura Global HQ')
  const [contactEmail, setContactEmail] = useState('ops@aura.io')
  const [saved, setSaved] = useState(false)

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
          // Try loading full profile
          try {
            const profile = await getOrgProfile(org.id)
            if (profile) {
              setOrgName(profile.name || '')
              setContactEmail(profile.contact_email || user.email || '')
            }
          } catch {}
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
      await updateOrgProfile(orgId, { name: orgName, contact_email: contactEmail })
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
          CORE_SETTINGS
        </h1>
        <p className="text-white/20 font-black uppercase tracking-[0.3em] text-[10px] italic">
          {orgName ? `${orgName} — ` : ''}Governance & Infrastructure Configuration
        </p>
      </header>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-12">
        {/* Navigation tabs */}
        <aside className="lg:col-span-1 space-y-2">
          <Button variant="ghost" className="w-full justify-start gap-4 bg-white/5 text-aura-glow border border-aura-glow/20 rounded-2xl py-6 font-black uppercase text-[10px] tracking-widest shadow-[0_0_15px_rgba(0,243,255,0.1)] transition-all">
            <User size={18} /> Profile & Org
          </Button>
          <Button variant="ghost" className="w-full justify-start gap-4 text-white/30 hover:text-white hover:bg-white/5 rounded-2xl py-6 font-black uppercase text-[10px] tracking-widest group transition-all">
            <Shield size={18} className="group-hover:text-aura-purple" /> Security & RBAC
          </Button>
          <Button variant="ghost" className="w-full justify-start gap-4 text-white/30 hover:text-white hover:bg-white/5 rounded-2xl py-6 font-black uppercase text-[10px] tracking-widest group transition-all">
            <Zap size={18} className="group-hover:text-aura-glow" /> API Integration
          </Button>
          <Button variant="ghost" className="w-full justify-start gap-4 text-white/30 hover:text-white hover:bg-white/5 rounded-2xl py-6 font-black uppercase text-[10px] tracking-widest group transition-all">
            <Bell size={18} className="group-hover:text-aura-accent" /> Alerts & Webhooks
          </Button>
        </aside>

        {/* Main Content */}
        <div className="lg:col-span-3 space-y-8">
          <div className="stat-card glow-purple p-8 neural-bg border-white/5 overflow-hidden relative">
             <div className="flex items-center gap-6 mb-8 pb-6 border-b border-white/5">
                <div className="w-14 h-14 rounded-2xl bg-aura-purple/10 border border-aura-purple/20 flex items-center justify-center text-aura-purple shadow-[0_0_15px_rgba(157,0,255,0.2)]">
                   <User size={28} />
                </div>
                <div>
                   <h3 className="text-xl font-black tracking-tight uppercase italic leading-none">Organization Profile</h3>
                   <p className="text-[10px] font-bold text-white/20 uppercase tracking-widest mt-1">Managed Neural Entity Identity</p>
                </div>
             </div>

             <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-10">
                <div className="space-y-3">
                   <label className="text-[10px] font-black uppercase tracking-widest text-white/40 italic pl-1">Entity Name</label>
                   <input 
                     className="w-full bg-black/40 border border-white/10 rounded-2xl px-5 py-4 text-sm font-bold text-white focus:border-aura-glow outline-none transition-all placeholder:text-white/10 shadow-inner" 
                     value={orgName}
                     onChange={(e) => setOrgName(e.target.value)}
                   />
                </div>
                <div className="space-y-3">
                   <label className="text-[10px] font-black uppercase tracking-widest text-white/40 italic pl-1">Technical Contact</label>
                   <input 
                     className="w-full bg-black/40 border border-white/10 rounded-2xl px-5 py-4 text-sm font-bold text-white focus:border-aura-glow outline-none transition-all placeholder:text-white/10 shadow-inner" 
                     value={contactEmail}
                     onChange={(e) => setContactEmail(e.target.value)}
                   />
                </div>
             </div>

             <div className="flex items-center justify-end gap-4 pt-4 border-t border-white/5">
                {saved && (
                  <span className="text-[10px] font-black text-aura-accent uppercase tracking-widest animate-pulse">Changes Synchronized</span>
                )}
                <Button 
                  onClick={handleSave}
                  disabled={loading}
                  className="bg-aura-glow text-black font-black uppercase tracking-[0.3em] text-[10px] px-8 h-14 rounded-2xl hover:scale-102 hover:shadow-[0_0_20px_rgba(0,243,255,0.3)] transition-all"
                >
                   <Save size={16} className="mr-2" /> {loading ? 'Syncing...' : 'Sync Changes'}
                </Button>
             </div>
             
             {/* Grain Texture */}
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
             
             <p className="text-[10px] text-white/30 mb-8 uppercase font-bold leading-relaxed max-w-xl tracking-widest">
                Deactivating the organization sector will purge all semantic clusters, vector points, and provisioned tool bindings. This action will initiate an IRREVERSIBLE data scrubbing protocol.
             </p>
             
             <Button variant="outline" className="border-red-500/20 text-red-500 font-black uppercase tracking-[0.3em] text-[10px] px-8 h-14 rounded-2xl hover:bg-red-500/10 hover:border-red-500/40 transition-all">
                Execute Dissolution
             </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
