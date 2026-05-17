'use client'

import { useState, useEffect } from 'react'
// import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Key, Plus, Trash2, Copy, CheckCircle2, ShieldAlert } from 'lucide-react'
import { getApiKeys, createApiKey, revokeApiKey } from '../../actions'
import { supabase } from '@/lib/supabase'

export default function ApiKeysPage() {
  const [keys, setKeys] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [newKeyName, setNewKeyName] = useState('')
  const [createdKey, setCreatedKey] = useState<string | null>(null)
  const [orgId, setOrgId] = useState<string | null>(null)
  const [orgName, setOrgName] = useState<string>('')

  useEffect(() => {
    async function load() {
      const { data: { user } } = await supabase.auth.getUser()
      if (user) {
        // Try to get org from members table
        const { data: membership } = await supabase
          .from('members')
          .select('org_id, organizations(id, name)')
          .eq('user_id', user.id)
          .limit(1)
          .maybeSingle()

        let resolvedOrgId = user.id  // Fallback: personal org
        let resolvedOrgName = user.email?.split('@')[0] || 'Personal'

        if (membership?.organizations) {
          const org = membership.organizations as any
          resolvedOrgId = org.id
          resolvedOrgName = org.name
        }

        setOrgId(resolvedOrgId)
        setOrgName(resolvedOrgName)

        try {
          const data = await getApiKeys(resolvedOrgId)
          setKeys(data)
        } catch {
          // Table might not exist yet
        }
      }
      setLoading(false)
    }
    load()
  }, [])

  const handleCreate = async () => {
    if (!newKeyName || !orgId) return
    const { key } = await createApiKey(orgId, newKeyName)
    setCreatedKey(key)
    setNewKeyName('')
    const data = await getApiKeys(orgId)
    setKeys(data)
  }

  const handleRevoke = async (id: string) => {
    if (!confirm('Are you sure you want to revoke this key?')) return
    await revokeApiKey(id)
    if (orgId) {
      const data = await getApiKeys(orgId)
      setKeys(data)
    }
  }

  return (
    <div className="space-y-12">
      <header className="mb-12">
        <h1 className="text-4xl font-black tracking-tighter text-white mb-2 uppercase italic">
          SECERN_KEYS
        </h1>
        <p className="text-white/20 font-black uppercase tracking-[0.3em] text-[10px] italic">
          {orgName ? `${orgName} — ` : ''}Neural API Access Control & Rotation
        </p>
      </header>

      {createdKey && (
        <div className="stat-card border-memzent-glow/50 bg-memzent-glow/10 animate-pulse p-8 neural-bg mb-8">
          <h3 className="text-memzent-glow flex items-center gap-2 font-black uppercase tracking-tighter text-lg mb-2">
            <CheckCircle2 size={24} />
            Key Generated Successfully
          </h3>
          <p className="text-white font-bold text-sm mb-6">
            Transfer this token to a secure vault. We will never display the full hash again.
          </p>
          <div className="flex items-center gap-3 p-6 bg-black/40 rounded-2xl font-mono text-sm border border-white/10 group">
            <span className="flex-1 text-memzent-glow select-all break-all font-black">{createdKey}</span>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => navigator.clipboard.writeText(createdKey)}
              className="hover:text-memzent-glow hover:bg-white/5"
            >
              <Copy size={20} />
            </Button>
          </div>
          <Button
            onClick={() => setCreatedKey(null)}
            className="mt-6 bg-memzent-glow text-black font-black uppercase text-[10px] tracking-widest px-8"
          >
            Acknowledged & Stored
          </Button>
        </div>
      )}

      <div className="stat-card neural-bg border-white/5 p-0 overflow-hidden">
        <div className="p-8 border-b border-white/5 flex flex-col md:flex-row items-center justify-between gap-6">
          <div>
            <h3 className="text-lg font-black tracking-tight uppercase italic">Active Nodes</h3>
            <p className="text-[10px] font-bold text-white/20 uppercase tracking-widest mt-1">Authentication Registry</p>
          </div>
          <div className="flex gap-4 w-full md:w-auto">
            <input
              type="text"
              placeholder="Key Label (e.g., k8s-cluster-01)"
              value={newKeyName}
              onChange={(e) => setNewKeyName(e.target.value)}
              className="bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-xs font-bold focus:outline-none focus:border-memzent-glow/50 text-white w-full md:w-64"
            />
            <Button onClick={handleCreate} disabled={!newKeyName} className="bg-memzent-glow text-black font-black uppercase tracking-widest text-[10px] px-6 h-12 shadow-[0_0_15px_rgba(0,243,255,0.2)]">
              <Plus size={16} className="mr-2" />
              Generate
            </Button>
          </div>
        </div>

        <div className="divide-y divide-white/5">
          {loading ? (
            <div className="py-20 text-center text-white/10 font-black italic uppercase tracking-[0.4em] text-sm">Synchronizing Registry...</div>
          ) : keys.length === 0 ? (
            <div className="py-20 text-center text-white/10 font-black italic uppercase tracking-[0.4em] text-sm">No Active Tokens Found</div>
          ) : (
            keys.map((k) => (
              <div key={k.id} className="flex items-center justify-between p-8 hover:bg-white/[0.02] transition-all group">
                <div className="flex items-center gap-6">
                  <div className="w-14 h-14 rounded-2xl bg-white/5 border border-white/10 flex items-center justify-center text-white/20 group-hover:text-memzent-purple transition-colors shadow-inner">
                    <Key size={24} />
                  </div>
                  <div>
                    <div className="text-base font-black tracking-tight text-white uppercase italic">{k.name}</div>
                    <div className="text-[10px] font-mono text-white/20 uppercase font-black flex items-center gap-3 mt-1">
                      <span className="text-white/40">IDENTIFIER:</span> <span className="text-memzent-purple">{k.key_prefix}</span>
                      <span className="w-1 h-1 rounded-full bg-white/10" />
                      <span className="text-white/40">PROVISIONED:</span> {new Date(k.created_at).toLocaleDateString()}
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-6">
                  <Badge variant="outline" className="border-memzent-glow/20 bg-memzent-glow/5 text-[10px] uppercase font-black text-memzent-glow px-4 py-1 tracking-widest">
                    SECURE
                  </Badge>
                  <Button
                    onClick={() => handleRevoke(k.id)}
                    variant="ghost"
                    size="icon"
                    className="text-white/10 hover:text-red-500 hover:bg-red-500/5 transition-all"
                  >
                    <Trash2 size={20} />
                  </Button>
                </div>
              </div>
            )
            ))}
        </div>
      </div>

      <footer className="stat-card border-memzent-purple/20 bg-memzent-purple/5 p-8 relative overflow-hidden group">
        <div className="absolute top-0 right-0 p-8 text-memzent-purple/5 group-hover:text-memzent-purple/10 transition-colors pointer-events-none">
          <ShieldAlert size={120} />
        </div>
        <h3 className="text-xs font-black text-memzent-purple uppercase tracking-[0.3em] mb-4 italic">Security Directive</h3>
        <p className="text-[10px] text-white/40 leading-relaxed font-black uppercase max-w-2xl tracking-widest">
          Neural API keys grant full execution power over the intelligence mesh. Leaked tokens will compromise organization privacy. Implement strict rotation and never leak secrets to public repositories.
        </p>
      </footer>
    </div>
  )
}
