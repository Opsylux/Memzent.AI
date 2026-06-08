'use client'

import React, { useState } from 'react'
import { Plus, Pencil, Trash2, X, Loader2, Eye } from 'lucide-react'
import { createWebhook, updateWebhook, deleteWebhook, getWebhookDeliveries } from '@/app/actions'
import { useRouter } from 'next/navigation'

const EVENT_TYPES = ['cache_hit', 'tool_execution', 'rate_limit', 'key_rotated', 'tool_registered', 'session_created']

interface WebhookActionsProps {
  webhooks: any[]
  orgId?: string
  mode: 'header' | 'row'
  webhookId?: string
}

export function WebhookActions({ webhooks, orgId, mode, webhookId }: WebhookActionsProps) {
  const [showCreate, setShowCreate] = useState(false)
  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)
  const [showLogs, setShowLogs] = useState(false)

  if (mode === 'header') {
    return (
      <>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 px-4 py-2 text-xs font-bold rounded-lg bg-memzent-glow/20 text-memzent-glow hover:bg-memzent-glow/30 transition-all"
        >
          <Plus size={14} /> Add Webhook
        </button>
        {showCreate && <CreateWebhookDialog orgId={orgId} onClose={() => setShowCreate(false)} />}
      </>
    )
  }

  const wh = webhooks[0]
  return (
    <>
      <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
        <button onClick={() => setShowLogs(true)} className="p-2 rounded-lg hover:bg-white/10 transition-all" title="View deliveries">
          <Eye size={14} className="text-white/60" />
        </button>
        <button onClick={() => setShowEdit(true)} className="p-2 rounded-lg hover:bg-white/10 transition-all" title="Edit">
          <Pencil size={14} className="text-white/60" />
        </button>
        <button onClick={() => setShowDelete(true)} className="p-2 rounded-lg hover:bg-white/10 hover:text-red-400 transition-all" title="Delete">
          <Trash2 size={14} className="text-white/60" />
        </button>
      </div>
      {showEdit && <EditWebhookDialog webhook={wh} orgId={orgId} onClose={() => setShowEdit(false)} />}
      {showDelete && <DeleteWebhookDialog webhook={wh} orgId={orgId} onClose={() => setShowDelete(false)} />}
      {showLogs && <DeliveryLogsDialog webhook={wh} orgId={orgId} onClose={() => setShowLogs(false)} />}
    </>
  )
}

function CreateWebhookDialog({ orgId, onClose }: { orgId?: string; onClose: () => void }) {
  const router = useRouter()
  const [url, setUrl] = useState('')
  const [description, setDescription] = useState('')
  const [selectedEvents, setSelectedEvents] = useState<string[]>(['cache_hit', 'rate_limit'])
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [secret, setSecret] = useState('')

  const toggleEvent = (evt: string) => {
    setSelectedEvents(prev => prev.includes(evt) ? prev.filter(e => e !== evt) : [...prev, evt])
  }

  const handleCreate = async () => {
    if (!url.trim()) { setError('URL is required'); return }
    if (selectedEvents.length === 0) { setError('Select at least one event'); return }
    setSaving(true); setError('')
    try {
      const result = await createWebhook({ url: url.trim(), events: selectedEvents, description: description.trim() || undefined }, orgId)
      setSecret(result.secret)
      router.refresh()
    } catch (e: any) {
      setError(e.message)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={onClose}>
      <div className="bg-[#0a0a0f] border border-white/10 rounded-2xl p-8 w-full max-w-lg shadow-2xl" onClick={e => e.stopPropagation()}>
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-xl font-black tracking-tight">Create Webhook</h2>
          <button onClick={onClose} className="p-2 rounded-lg hover:bg-white/10"><X size={18} /></button>
        </div>

        {secret ? (
          <div className="space-y-4">
            <div className="p-4 bg-memzent-accent/5 border border-memzent-accent/20 rounded-lg">
              <p className="text-xs font-bold text-memzent-accent mb-2">✅ Webhook created! Save your signing secret:</p>
              <code className="text-xs font-mono bg-black/50 px-3 py-2 rounded block break-all">{secret}</code>
              <p className="text-[10px] text-white/40 mt-2">This secret won't be shown again. Use it to verify HMAC-SHA256 signatures.</p>
            </div>
            <button onClick={onClose} className="w-full px-4 py-2 text-xs font-bold rounded-lg bg-white/5 hover:bg-white/10 transition-all">
              Done
            </button>
          </div>
        ) : (
          <div className="space-y-4">
            <div>
              <label className="text-[10px] font-bold uppercase tracking-widest text-white/50 mb-1 block">Endpoint URL</label>
              <input
                className="w-full bg-white/[0.03] border border-white/10 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-memzent-glow/30 transition-all"
                placeholder="https://your-app.com/webhooks/memzent"
                value={url} onChange={e => setUrl(e.target.value)}
              />
            </div>
            <div>
              <label className="text-[10px] font-bold uppercase tracking-widest text-white/50 mb-1 block">Description (optional)</label>
              <input
                className="w-full bg-white/[0.03] border border-white/10 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-memzent-glow/30 transition-all"
                placeholder="Production webhook for monitoring"
                value={description} onChange={e => setDescription(e.target.value)}
              />
            </div>
            <div>
              <label className="text-[10px] font-bold uppercase tracking-widest text-white/50 mb-2 block">Events to Subscribe</label>
              <div className="flex flex-wrap gap-2">
                {EVENT_TYPES.map(evt => (
                  <button
                    key={evt}
                    onClick={() => toggleEvent(evt)}
                    className={`text-[10px] font-bold px-3 py-1.5 rounded-lg border transition-all ${
                      selectedEvents.includes(evt)
                        ? 'bg-memzent-glow/10 border-memzent-glow/30 text-memzent-glow'
                        : 'bg-white/[0.02] border-white/10 text-white/40 hover:text-white/60'
                    }`}
                  >
                    {evt}
                  </button>
                ))}
              </div>
            </div>

            {error && <p className="text-red-400 text-xs">{error}</p>}

            <div className="flex justify-end gap-3 pt-2">
              <button onClick={onClose} className="px-4 py-2 text-xs font-bold rounded-lg bg-white/5 hover:bg-white/10 transition-all">Cancel</button>
              <button onClick={handleCreate} disabled={saving}
                className="px-4 py-2 text-xs font-bold rounded-lg bg-memzent-glow/20 text-memzent-glow hover:bg-memzent-glow/30 transition-all disabled:opacity-50 flex items-center gap-2">
                {saving && <Loader2 size={14} className="animate-spin" />} Create Webhook
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

function EditWebhookDialog({ webhook, orgId, onClose }: { webhook: any; orgId?: string; onClose: () => void }) {
  const router = useRouter()
  const [url, setUrl] = useState(webhook.url)
  const [description, setDescription] = useState(webhook.description || '')
  const [selectedEvents, setSelectedEvents] = useState<string[]>(webhook.events || [])
  const [enabled, setEnabled] = useState(webhook.enabled)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const toggleEvent = (evt: string) => {
    setSelectedEvents(prev => prev.includes(evt) ? prev.filter(e => e !== evt) : [...prev, evt])
  }

  const handleSave = async () => {
    setSaving(true); setError('')
    try {
      await updateWebhook(webhook.id, { url, events: selectedEvents, enabled, description }, orgId)
      router.refresh()
      onClose()
    } catch (e: any) {
      setError(e.message)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={onClose}>
      <div className="bg-[#0a0a0f] border border-white/10 rounded-2xl p-8 w-full max-w-lg shadow-2xl" onClick={e => e.stopPropagation()}>
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-xl font-black tracking-tight">Edit Webhook</h2>
          <button onClick={onClose} className="p-2 rounded-lg hover:bg-white/10"><X size={18} /></button>
        </div>
        <div className="space-y-4">
          <div>
            <label className="text-[10px] font-bold uppercase tracking-widest text-white/50 mb-1 block">URL</label>
            <input className="w-full bg-white/[0.03] border border-white/10 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-memzent-glow/30" value={url} onChange={e => setUrl(e.target.value)} />
          </div>
          <div>
            <label className="text-[10px] font-bold uppercase tracking-widest text-white/50 mb-1 block">Description</label>
            <input className="w-full bg-white/[0.03] border border-white/10 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-memzent-glow/30" value={description} onChange={e => setDescription(e.target.value)} />
          </div>
          <div>
            <label className="text-[10px] font-bold uppercase tracking-widest text-white/50 mb-2 block">Events</label>
            <div className="flex flex-wrap gap-2">
              {EVENT_TYPES.map(evt => (
                <button key={evt} onClick={() => toggleEvent(evt)}
                  className={`text-[10px] font-bold px-3 py-1.5 rounded-lg border transition-all ${selectedEvents.includes(evt) ? 'bg-memzent-glow/10 border-memzent-glow/30 text-memzent-glow' : 'bg-white/[0.02] border-white/10 text-white/40'}`}>
                  {evt}
                </button>
              ))}
            </div>
          </div>
          <label className="flex items-center gap-2 cursor-pointer">
            <div className={`w-8 h-4 rounded-full transition-all ${enabled ? 'bg-memzent-glow/40' : 'bg-white/10'} relative`} onClick={() => setEnabled(!enabled)}>
              <div className={`w-3 h-3 rounded-full bg-white absolute top-0.5 transition-all ${enabled ? 'left-4' : 'left-0.5'}`} />
            </div>
            <span className="text-xs font-bold text-white/60">Enabled</span>
          </label>
          {error && <p className="text-red-400 text-xs">{error}</p>}
          <div className="flex justify-end gap-3 pt-2">
            <button onClick={onClose} className="px-4 py-2 text-xs font-bold rounded-lg bg-white/5 hover:bg-white/10 transition-all">Cancel</button>
            <button onClick={handleSave} disabled={saving}
              className="px-4 py-2 text-xs font-bold rounded-lg bg-memzent-glow/20 text-memzent-glow hover:bg-memzent-glow/30 disabled:opacity-50 flex items-center gap-2">
              {saving && <Loader2 size={14} className="animate-spin" />} Save
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

function DeleteWebhookDialog({ webhook, orgId, onClose }: { webhook: any; orgId?: string; onClose: () => void }) {
  const router = useRouter()
  const [deleting, setDeleting] = useState(false)

  const handleDelete = async () => {
    setDeleting(true)
    try {
      await deleteWebhook(webhook.id, orgId)
      router.refresh()
      onClose()
    } catch (e) {
      console.error(e)
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={onClose}>
      <div className="bg-[#0a0a0f] border border-white/10 rounded-2xl p-8 w-full max-w-md shadow-2xl" onClick={e => e.stopPropagation()}>
        <h2 className="text-xl font-black tracking-tight mb-2">Delete Webhook</h2>
        <p className="text-sm text-white/60 mb-6">
          Remove webhook for <span className="text-white font-mono text-xs">{webhook.url}</span>? All delivery history will be deleted.
        </p>
        <div className="flex justify-end gap-3">
          <button onClick={onClose} className="px-4 py-2 text-xs font-bold rounded-lg bg-white/5 hover:bg-white/10 transition-all">Cancel</button>
          <button onClick={handleDelete} disabled={deleting}
            className="px-4 py-2 text-xs font-bold rounded-lg bg-red-500/20 text-red-400 hover:bg-red-500/30 disabled:opacity-50 flex items-center gap-2">
            {deleting && <Loader2 size={14} className="animate-spin" />} Delete
          </button>
        </div>
      </div>
    </div>
  )
}

function DeliveryLogsDialog({ webhook, orgId, onClose }: { webhook: any; orgId?: string; onClose: () => void }) {
  const [logs, setLogs] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  React.useEffect(() => {
    getWebhookDeliveries(webhook.id, orgId).then(data => {
      setLogs(data || [])
      setLoading(false)
    })
  }, [webhook.id, orgId])

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={onClose}>
      <div className="bg-[#0a0a0f] border border-white/10 rounded-2xl p-8 w-full max-w-2xl max-h-[80vh] overflow-hidden flex flex-col shadow-2xl" onClick={e => e.stopPropagation()}>
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-xl font-black tracking-tight">Delivery Logs</h2>
          <button onClick={onClose} className="p-2 rounded-lg hover:bg-white/10"><X size={18} /></button>
        </div>
        <div className="flex-1 overflow-y-auto">
          {loading ? (
            <div className="flex items-center justify-center py-12"><Loader2 size={24} className="animate-spin text-white/30" /></div>
          ) : logs.length === 0 ? (
            <p className="text-sm text-white/40 text-center py-12">No deliveries yet</p>
          ) : (
            <div className="space-y-2">
              {logs.map((log: any) => (
                <div key={log.id} className="p-3 bg-white/[0.02] border border-white/5 rounded-lg flex items-center gap-4">
                  <StatusIcon status={log.status} />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-[10px] font-bold text-white/60">{log.event_type}</span>
                      <span className="text-[9px] text-white/30">{new Date(log.created_at).toLocaleString()}</span>
                    </div>
                    {log.error && <p className="text-[10px] text-red-400/80 truncate">{log.error}</p>}
                  </div>
                  <span className={`text-[9px] font-black px-2 py-0.5 rounded ${statusColor(log.status)}`}>
                    {log.status.toUpperCase()}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

function StatusIcon({ status }: { status: string }) {
  if (status === 'delivered') return <div className="w-2 h-2 rounded-full bg-memzent-accent" />
  if (status === 'failed') return <div className="w-2 h-2 rounded-full bg-yellow-500" />
  if (status === 'dead_letter') return <div className="w-2 h-2 rounded-full bg-red-500" />
  return <div className="w-2 h-2 rounded-full bg-white/20" />
}

function statusColor(status: string) {
  if (status === 'delivered') return 'bg-memzent-accent/10 text-memzent-accent'
  if (status === 'failed') return 'bg-yellow-500/10 text-yellow-500'
  if (status === 'dead_letter') return 'bg-red-500/10 text-red-500'
  return 'bg-white/5 text-white/40'
}
