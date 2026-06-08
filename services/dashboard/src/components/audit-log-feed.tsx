"use client"

import { useState, useEffect } from "react"
import { Terminal } from "lucide-react"
import { getMemzentAudit } from "@/app/actions"
import Link from "next/link"

interface AuditEvent {
  timestamp: string
  type: string
  detail: string
  status: string
  user: string
}

function formatTime(isoString: string) {
  try {
    const parts = isoString.split("T")
    if (parts[1]) {
      return parts[1].split(".")[0] || parts[1]
    }
    return isoString
  } catch {
    return isoString
  }
}

export function AuditLogFeed({ orgId }: { orgId?: string }) {
  const [events, setEvents] = useState<AuditEvent[]>([])

  useEffect(() => {
    async function fetchAudit() {
      try {
        const data = await getMemzentAudit(orgId)
        setEvents(data || [])
      } catch (err) {
        console.error("Failed to fetch audit logs:", err)
      }
    }

    fetchAudit()
    const interval = setInterval(fetchAudit, 5000) // Poll every 5s for "live" feel
    return () => clearInterval(interval)
  }, [orgId])

  return (
    <div className="stat-card glow-cyan border-white/10 p-8 neural-bg relative overflow-hidden bg-black/40 h-full flex flex-col">
      <div className="flex items-center justify-between mb-8">
        <h3 className="text-xs font-black tracking-widest text-white/75 uppercase leading-none italic font-roboto-mono flex items-center gap-2">
          <Terminal size={14} className="text-memzent-glow" />
          Live Neural Audit
        </h3>
        <div className="flex items-center gap-1.5">
          <div className="w-1.5 h-1.5 rounded-full bg-memzent-glow animate-pulse shadow-[0_0_8px_#00f3ff]" />
          <span className="text-[10px] font-bold text-memzent-glow/80 uppercase tracking-widest">Live Feed</span>
        </div>
      </div>

      <div className="font-mono text-[9px] space-y-5 flex-1 overflow-y-auto">
        {events.length === 0 ? (
          <div className="text-white/35 italic text-center py-10">Initializing Neural Feed...</div>
        ) : (
          events.map((event, i) => (
            <div key={i} className={`flex gap-4 border-l pl-4 py-1 hover:bg-white/[0.03] transition-colors group ${event.status === 'error' ? 'border-red-500/50' :
              event.status === 'warning' ? 'border-memzent-purple/50' :
                'border-white/5'
              }`}>
              <span className="text-memzent-glow/65 font-bold">
                {formatTime(event.timestamp)}
              </span>
              <span className={`transition-colors ${event.status === 'error' ? 'text-red-400' :
                event.status === 'warning' ? 'text-memzent-purple' :
                  'text-white/65 group-hover:text-white/85'
                }`}>
                [{event.type}] {event.detail}
              </span>
            </div>
          ))
        )}
      </div>

      <div className="mt-8 pt-6 border-t border-white/5">
        <Link href="/audit" className="block w-full">
          <button className="w-full py-3 rounded-xl bg-white/5 border border-white/10 text-[10px] font-black uppercase tracking-[0.3em] text-white/55 hover:text-white hover:bg-white/10 hover:border-white/20 transition-all cursor-pointer">
            OPEN INSPECTOR ENGINE
          </button>
        </Link>
      </div>

      <div className="absolute inset-x-0 bottom-0 h-32 bg-gradient-to-t from-memzent-dark to-transparent pointer-events-none" />
    </div>
  )
}
