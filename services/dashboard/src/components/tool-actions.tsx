'use client'

import React, { useState } from 'react'
import { Pencil, Trash2, X, Loader2 } from 'lucide-react'
import { updateTool, deleteTool } from '@/app/actions'
import { useRouter } from 'next/navigation'

interface ToolData {
  id: string
  name: string
  description: string
  connector_type: string
  endpoint: string
  timeout_seconds: number
  enabled: boolean
  requires_auth: boolean
  input_schema?: Record<string, any>
  output_schema?: Record<string, any>
}

export function ToolActions({ tool, orgId }: { tool: ToolData; orgId?: string }) {
  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)

  return (
    <>
      <div className="flex items-center justify-end gap-2 opacity-20 group-hover:opacity-100 transition-opacity">
        <button
          onClick={() => setShowEdit(true)}
          className="p-2 rounded-lg hover:bg-white/10 hover:text-memzent-glow transition-all"
          title="Edit tool"
        >
          <Pencil size={16} />
        </button>
        <button
          onClick={() => setShowDelete(true)}
          className="p-2 rounded-lg hover:bg-white/10 hover:text-red-400 transition-all"
          title="Delete tool"
        >
          <Trash2 size={16} />
        </button>
      </div>

      {showEdit && <EditToolDialog tool={tool} orgId={orgId} onClose={() => setShowEdit(false)} />}
      {showDelete && <DeleteToolDialog tool={tool} orgId={orgId} onClose={() => setShowDelete(false)} />}
    </>
  )
}

function EditToolDialog({ tool, orgId, onClose }: { tool: ToolData; orgId?: string; onClose: () => void }) {
  const router = useRouter()
  const [form, setForm] = useState({
    name: tool.name,
    description: tool.description,
    connector_type: tool.connector_type,
    endpoint: tool.endpoint,
    timeout_seconds: tool.timeout_seconds,
    enabled: tool.enabled,
    requires_auth: tool.requires_auth,
  })
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const handleSave = async () => {
    setSaving(true)
    setError('')
    try {
      await updateTool(tool.id, form, orgId)
      router.refresh()
      onClose()
    } catch (e: any) {
      setError(e.message || 'Failed to update tool')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={onClose}>
      <div className="bg-[#0a0a0f] border border-white/10 rounded-2xl p-8 w-full max-w-lg shadow-2xl" onClick={e => e.stopPropagation()}>
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-xl font-black tracking-tight">Edit Tool</h2>
          <button onClick={onClose} className="p-2 rounded-lg hover:bg-white/10 transition-all">
            <X size={18} />
          </button>
        </div>

        <div className="space-y-4">
          <Field label="Name" value={form.name} onChange={v => setForm(f => ({ ...f, name: v }))} />
          <Field label="Description" value={form.description} onChange={v => setForm(f => ({ ...f, description: v }))} multiline />
          <div className="grid grid-cols-2 gap-4">
            <SelectField
              label="Connector"
              value={form.connector_type}
              onChange={v => setForm(f => ({ ...f, connector_type: v }))}
              options={['mcp', 'rest', 'sql', 'graphql', 'grpc', 'webhook']}
            />
            <Field
              label="Timeout (s)"
              value={String(form.timeout_seconds)}
              onChange={v => setForm(f => ({ ...f, timeout_seconds: parseInt(v) || 15 }))}
              type="number"
            />
          </div>
          <Field label="Endpoint" value={form.endpoint} onChange={v => setForm(f => ({ ...f, endpoint: v }))} />
          <div className="flex gap-6">
            <Toggle label="Enabled" checked={form.enabled} onChange={v => setForm(f => ({ ...f, enabled: v }))} />
            <Toggle label="Requires Auth" checked={form.requires_auth} onChange={v => setForm(f => ({ ...f, requires_auth: v }))} />
          </div>
        </div>

        {error && <p className="text-red-400 text-xs mt-4">{error}</p>}

        <div className="flex justify-end gap-3 mt-8">
          <button onClick={onClose} className="px-4 py-2 text-xs font-bold rounded-lg bg-white/5 hover:bg-white/10 transition-all">
            Cancel
          </button>
          <button
            onClick={handleSave}
            disabled={saving}
            className="px-4 py-2 text-xs font-bold rounded-lg bg-memzent-glow/20 text-memzent-glow hover:bg-memzent-glow/30 transition-all disabled:opacity-50 flex items-center gap-2"
          >
            {saving && <Loader2 size={14} className="animate-spin" />}
            Save Changes
          </button>
        </div>
      </div>
    </div>
  )
}

function DeleteToolDialog({ tool, orgId, onClose }: { tool: ToolData; orgId?: string; onClose: () => void }) {
  const router = useRouter()
  const [deleting, setDeleting] = useState(false)
  const [error, setError] = useState('')

  const handleDelete = async () => {
    setDeleting(true)
    setError('')
    try {
      await deleteTool(tool.id, orgId)
      router.refresh()
      onClose()
    } catch (e: any) {
      setError(e.message || 'Failed to delete tool')
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={onClose}>
      <div className="bg-[#0a0a0f] border border-white/10 rounded-2xl p-8 w-full max-w-md shadow-2xl" onClick={e => e.stopPropagation()}>
        <h2 className="text-xl font-black tracking-tight mb-2">Delete Tool</h2>
        <p className="text-sm text-white/60 mb-6">
          Are you sure you want to disable <span className="text-white font-bold">{tool.name}</span>?
          This will remove it from semantic routing. The tool can be re-enabled later.
        </p>

        {error && <p className="text-red-400 text-xs mb-4">{error}</p>}

        <div className="flex justify-end gap-3">
          <button onClick={onClose} className="px-4 py-2 text-xs font-bold rounded-lg bg-white/5 hover:bg-white/10 transition-all">
            Cancel
          </button>
          <button
            onClick={handleDelete}
            disabled={deleting}
            className="px-4 py-2 text-xs font-bold rounded-lg bg-red-500/20 text-red-400 hover:bg-red-500/30 transition-all disabled:opacity-50 flex items-center gap-2"
          >
            {deleting && <Loader2 size={14} className="animate-spin" />}
            Delete Tool
          </button>
        </div>
      </div>
    </div>
  )
}

// Reusable form components
function Field({ label, value, onChange, multiline, type = 'text' }: {
  label: string; value: string; onChange: (v: string) => void; multiline?: boolean; type?: string
}) {
  const cls = "w-full bg-white/[0.03] border border-white/10 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-memzent-glow/30 transition-all"
  return (
    <div>
      <label className="text-[10px] font-bold uppercase tracking-widest text-white/50 mb-1 block">{label}</label>
      {multiline ? (
        <textarea className={`${cls} h-20 resize-none`} value={value} onChange={e => onChange(e.target.value)} />
      ) : (
        <input className={cls} type={type} value={value} onChange={e => onChange(e.target.value)} />
      )}
    </div>
  )
}

function SelectField({ label, value, onChange, options }: {
  label: string; value: string; onChange: (v: string) => void; options: string[]
}) {
  return (
    <div>
      <label className="text-[10px] font-bold uppercase tracking-widest text-white/50 mb-1 block">{label}</label>
      <select
        className="w-full bg-white/[0.03] border border-white/10 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-memzent-glow/30 transition-all"
        value={value}
        onChange={e => onChange(e.target.value)}
      >
        {options.map(o => <option key={o} value={o}>{o.toUpperCase()}</option>)}
      </select>
    </div>
  )
}

function Toggle({ label, checked, onChange }: { label: string; checked: boolean; onChange: (v: boolean) => void }) {
  return (
    <label className="flex items-center gap-2 cursor-pointer">
      <div
        className={`w-8 h-4 rounded-full transition-all ${checked ? 'bg-memzent-glow/40' : 'bg-white/10'} relative`}
        onClick={() => onChange(!checked)}
      >
        <div className={`w-3 h-3 rounded-full bg-white absolute top-0.5 transition-all ${checked ? 'left-4' : 'left-0.5'}`} />
      </div>
      <span className="text-xs font-bold text-white/60">{label}</span>
    </label>
  )
}
