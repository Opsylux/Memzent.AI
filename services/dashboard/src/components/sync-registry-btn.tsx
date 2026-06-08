"use client"

import { useState } from "react"
import { RefreshCcw, Loader2, CheckCircle2, ShieldAlert } from "lucide-react"
import { syncMemzentTools } from "@/app/actions"

export function SyncRegistryBtn({ orgId }: { orgId?: string }) {
  const [status, setStatus] = useState<'idle' | 'syncing' | 'success' | 'error'>('idle')
  const [error, setError] = useState<string | null>(null)

  const handleSync = async () => {
    setStatus('syncing')
    setError(null)

    try {
      await syncMemzentTools(orgId)
      setStatus('success')
      setTimeout(() => setStatus('idle'), 3000)
    } catch (err: any) {
      setError(err.message || "Sync failed")
      setStatus('error')
      setTimeout(() => setStatus('idle'), 5000)
    }
  }

  return (
    <div className="relative">
      <button
        onClick={handleSync}
        disabled={status === 'syncing'}
        className={`glass px-6 py-3 rounded-2xl text-xs font-black tracking-widest uppercase flex items-center gap-2 transition-all group overflow-hidden ${status === 'syncing' ? 'text-memzent-glow' :
          status === 'success' ? 'text-memzent-accent border-memzent-accent/30' :
            status === 'error' ? 'text-red-400 border-red-500/30' :
              'text-white/40 hover:text-white hover:bg-white/5'
          }`}
      >
        {status === 'syncing' ? (
          <Loader2 size={14} className="animate-spin" />
        ) : status === 'success' ? (
          <CheckCircle2 size={14} className="animate-in zoom-in duration-300" />
        ) : status === 'error' ? (
          <ShieldAlert size={14} className="animate-in shake duration-300" />
        ) : (
          <RefreshCcw size={14} className="group-hover:rotate-180 transition-transform duration-500" />
        )}

        {status === 'syncing' ? 'SYNCING_NODES' :
          status === 'success' ? 'SYNC_COMPLETE' :
            status === 'error' ? 'SYNC_FAILED' :
              'Sync Registry'}

        {status === 'syncing' && (
          <div className="absolute inset-0 bg-memzent-glow/5 animate-pulse" />
        )}
      </button>

      {error && status === 'error' && (
        <div className="absolute top-full right-0 mt-2 p-3 bg-red-500/10 border border-red-500/20 rounded-xl text-[10px] font-black uppercase tracking-widest text-red-400 whitespace-nowrap z-50">
          {error}
        </div>
      )}
    </div>
  )
}
