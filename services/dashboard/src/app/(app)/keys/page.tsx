'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { 
  Key, Plus, Trash2, Copy, CheckCircle2, ShieldAlert,
  Shield, Eye, Cpu, Settings, Activity, Server
} from 'lucide-react'
import { getApiKeys, createApiKey, revokeApiKey } from '../../actions'
import { supabase } from '@/lib/supabase'

const ROLE_OPTIONS = [
  { id: 'viewer', label: 'Viewer Token', desc: 'ReadOnly access to logs and registries', icon: <Eye size={14} className="text-white/60" /> },
  { id: 'agent', label: 'Agent Token', desc: 'Default execution token for autonomous agents', icon: <Cpu size={14} className="text-memzent-glow" /> },
  { id: 'admin', label: 'Admin Token', desc: 'Bypass authorization checks with full access', icon: <Shield size={14} className="text-memzent-purple" /> }
] as const

const SCOPE_OPTIONS = [
  { id: 'chat:execute', label: 'Execute Chat & Caching', desc: 'Allows prompt evaluation and cache storage' },
  { id: 'tools:read', label: 'Read Registry Tools', desc: 'Allows discovery of available active tools' },
  { id: 'tools:write', label: 'Register & Sync Tools', desc: 'Allows tool ingestion and vector syncing' },
  { id: 'audit:read', label: 'Read Diagnostics & Audits', desc: 'Access metric pipelines and log feeds' }
]

export default function ApiKeysPage() {
  const [keys, setKeys] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [newKeyName, setNewKeyName] = useState('')
  const [selectedRole, setSelectedRole] = useState<'viewer' | 'agent' | 'admin'>('agent')
  const [selectedScopes, setSelectedScopes] = useState<string[]>(['chat:execute', 'tools:read'])
  const [createdKey, setCreatedKey] = useState<string | null>(null)
  const [orgId, setOrgId] = useState<string | null>(null)
  const [orgName, setOrgName] = useState<string>('')

  useEffect(() => {
    async function load() {
      const { data: { user } } = await supabase.auth.getUser()
      if (user) {
        const { data: membership } = await supabase
          .from('members')
          .select('org_id, organizations(id, name)')
          .eq('user_id', user.id)
          .limit(1)
          .maybeSingle()

        let resolvedOrgId = user.id
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
          // Fallback
        }
      }
      setLoading(false)
    }
    load()
  }, [])

  const handleCreate = async () => {
    if (!newKeyName || !orgId) return
    try {
      const res = await createApiKey(orgId, newKeyName, selectedScopes, selectedRole)
      if (res && res.key) {
        setCreatedKey(res.key)
        setNewKeyName('')
        setSelectedRole('agent')
        setSelectedScopes(['chat:execute', 'tools:read'])
        const data = await getApiKeys(orgId)
        setKeys(data)
      } else {
        alert("Failed to generate key: Invalid server response")
      }
    } catch (err: any) {
      console.error("Key generation failed:", err)
      alert(`Key generation failed: ${err.message || err}`)
    }
  }

  const handleRevoke = async (id: string) => {
    if (!confirm('Are you sure you want to revoke this key?')) return
    await revokeApiKey(id)
    if (orgId) {
      const data = await getApiKeys(orgId)
      setKeys(data)
    }
  }

  const toggleScope = (scope: string) => {
    setSelectedScopes(prev => 
      prev.includes(scope)
        ? prev.filter(s => s !== scope)
        : [...prev, scope]
    )
  }

  return (
    <div className="space-y-12">
      <header className="mb-12">
        <h1 className="text-4xl font-black tracking-tighter text-white mb-2 uppercase italic">
          SECERN_KEYS
        </h1>
        <p className="text-white/50 font-black uppercase tracking-[0.3em] text-[10px] italic">
          {orgName ? `${orgName} — ` : ''}Neural API Access Control & Granular RBAC
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

      {/* Main Creation Form Card */}
      <div className="stat-card neural-bg border-white/5 p-8 space-y-8">
        <div>
          <h3 className="text-lg font-black tracking-tight uppercase italic">Generate Intelligent Token</h3>
          <p className="text-[10px] font-bold text-white/50 uppercase tracking-widest mt-1">Configure identity boundaries and access layers</p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Col 1: Label & Role */}
          <div className="space-y-6 lg:border-r lg:border-white/5 lg:pr-8">
            <div className="space-y-2">
              <label className="text-[10px] font-black uppercase tracking-widest text-white/60">Key Name</label>
              <input
                type="text"
                placeholder="Label (e.g. production-gateway-01)"
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
                className="bg-black/40 border border-white/10 rounded-xl px-4 py-3 text-xs font-bold focus:outline-none focus:border-memzent-glow/50 text-white w-full"
              />
            </div>

            <div className="space-y-3">
              <label className="text-[10px] font-black uppercase tracking-widest text-white/40 block">Key Identity Type</label>
              <div className="space-y-2">
                {ROLE_OPTIONS.map((opt) => (
                  <button
                    key={opt.id}
                    type="button"
                    onClick={() => setSelectedRole(opt.id)}
                    className={`w-full text-left p-3 rounded-xl border flex items-start gap-3 transition-all ${
                      selectedRole === opt.id
                        ? 'bg-white/5 border-white/20 shadow-inner shadow-black'
                        : 'bg-transparent border-white/5 hover:border-white/10'
                    }`}
                  >
                    <div className="mt-0.5">{opt.icon}</div>
                    <div>
                      <div className="text-[10px] font-black uppercase tracking-wider text-white">{opt.label}</div>
                      <div className="text-[8px] text-white/60 font-bold mt-0.5 uppercase tracking-wide leading-tight">{opt.desc}</div>
                    </div>
                  </button>
                ))}
              </div>
            </div>
          </div>

          {/* Col 2: Permission Scopes */}
          <div className="space-y-4 lg:col-span-2">
            <label className="text-[10px] font-black uppercase tracking-widest text-white/40 block">Granular Permission Scopes (RBAC)</label>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {SCOPE_OPTIONS.map((scope) => {
                const isSelected = selectedScopes.includes(scope.id)
                return (
                  <button
                    key={scope.id}
                    type="button"
                    onClick={() => toggleScope(scope.id)}
                    className={`text-left p-4 rounded-2xl border transition-all flex items-start gap-4 ${
                      isSelected
                        ? 'bg-memzent-glow/5 border-memzent-glow/20'
                        : 'bg-black/20 border-white/5 hover:border-white/10'
                    }`}
                  >
                    <div className={`w-4 h-4 rounded border flex items-center justify-center transition-all mt-0.5 ${
                      isSelected 
                        ? 'border-memzent-glow bg-memzent-glow text-black' 
                        : 'border-white/20 bg-black/40'
                    }`}>
                      {isSelected && <div className="w-1.5 h-1.5 rounded-full bg-black" />}
                    </div>
                    <div>
                      <div className={`text-[10px] font-black uppercase tracking-wider ${isSelected ? 'text-memzent-glow' : 'text-white'}`}>
                        {scope.label}
                      </div>
                      <div className="text-[8px] font-bold text-white/55 uppercase tracking-widest mt-1 leading-normal">
                        {scope.desc}
                      </div>
                    </div>
                  </button>
                )
              })}
            </div>

            <div className="pt-6 border-t border-white/5 flex items-center justify-between">
              <div className="text-[9px] font-black uppercase tracking-wider text-white/50">
                {selectedScopes.length} scopes active for <span className="text-memzent-glow">{selectedRole}</span>
              </div>
              <Button 
                onClick={handleCreate} 
                disabled={!newKeyName || selectedScopes.length === 0} 
                className="bg-memzent-glow text-black font-black uppercase tracking-[0.2em] text-[10px] px-8 h-12 shadow-[0_0_20px_rgba(0,243,255,0.2)] hover:shadow-[0_0_30px_rgba(0,243,255,0.4)] transition-all"
              >
                <Plus size={14} className="mr-2" />
                Generate Key
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* Active Keys Registry */}
      <div className="stat-card neural-bg border-white/5 p-0 overflow-hidden">
        <div className="p-8 border-b border-white/5">
          <h3 className="text-lg font-black tracking-tight uppercase italic">Provisioned Credentials</h3>
          <p className="text-[10px] font-bold text-white/50 uppercase tracking-widest mt-1">Multi-token cluster authorization table</p>
        </div>

        <div className="divide-y divide-white/5">
          {loading ? (
            <div className="py-20 text-center text-white/10 font-black italic uppercase tracking-[0.4em] text-sm animate-pulse">Synchronizing Registry...</div>
          ) : keys.length === 0 ? (
            <div className="py-20 text-center text-white/10 font-black italic uppercase tracking-[0.4em] text-sm">No Active Tokens Found</div>
          ) : (
            keys.map((k) => {
              // Parse stored scopes or fallback to basic defaults
              const scopesList = Array.isArray(k.scopes) ? k.scopes : ['chat:execute', 'tools:read']
              const roleLabel = k.role || 'agent'
              
              return (
                <div key={k.id} className="p-8 hover:bg-white/[0.02] transition-all group space-y-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-6">
                      <div className="w-14 h-14 rounded-2xl bg-white/5 border border-white/10 flex items-center justify-center text-white/20 group-hover:text-memzent-purple transition-colors shadow-inner">
                        <Key size={24} />
                      </div>
                      <div>
                        <div className="text-base font-black tracking-tight text-white uppercase italic flex items-center gap-3">
                          {k.name}
                          <Badge className={`text-[8px] uppercase tracking-widest font-black ${
                            roleLabel === 'admin' 
                              ? 'bg-memzent-purple/10 border-memzent-purple/20 text-memzent-purple' 
                              : roleLabel === 'viewer'
                              ? 'bg-white/5 border-white/10 text-white/40'
                              : 'bg-memzent-glow/10 border-memzent-glow/20 text-memzent-glow'
                          }`}>
                            {roleLabel}
                          </Badge>
                        </div>
                        <div className="text-[10px] font-mono text-white/45 uppercase font-black flex items-center gap-3 mt-1.5">
                          <span className="text-white/40">IDENTIFIER:</span> <span className="text-memzent-purple">{k.key_prefix}</span>
                          <span className="w-1 h-1 rounded-full bg-white/10" />
                          <span className="text-white/40">PROVISIONED:</span> {new Date(k.created_at).toLocaleDateString()}
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-4">
                      <Button
                        onClick={() => handleRevoke(k.id)}
                        variant="ghost"
                        size="icon"
                        className="text-white/35 hover:text-red-500 hover:bg-red-500/5 transition-all"
                      >
                        <Trash2 size={20} />
                      </Button>
                    </div>
                  </div>

                  {/* Scopes Badges Grid */}
                  <div className="pl-20 flex flex-wrap gap-2">
                    {scopesList.map((scope: string) => (
                      <span 
                        key={scope} 
                        className="text-[8px] font-black uppercase tracking-wider px-2 py-1 rounded bg-black/40 border border-white/5 text-white/60 flex items-center gap-1.5"
                      >
                        <span className="w-1 h-1 rounded-full bg-white/20" />
                        {scope}
                      </span>
                    ))}
                  </div>
                </div>
              )
            })
          )}
        </div>
      </div>

      <footer className="stat-card border-memzent-purple/20 bg-memzent-purple/5 p-8 relative overflow-hidden group">
        <div className="absolute top-0 right-0 p-8 text-memzent-purple/5 group-hover:text-memzent-purple/10 transition-colors pointer-events-none">
          <ShieldAlert size={120} />
        </div>
        <h3 className="text-xs font-black text-memzent-purple uppercase tracking-[0.3em] mb-4 italic">Security Directive</h3>
        <p className="text-[10px] text-white/65 leading-relaxed font-black uppercase max-w-2xl tracking-widest">
          Granular permission boundaries restrict specific keys from accessing destructive commands. Always assign the minimum required scopes (least privilege principle) when integrating external agents or workflows.
        </p>
      </footer>
    </div>
  )
}
