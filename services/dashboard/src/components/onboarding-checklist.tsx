'use client'

import { Zap, Key, Database, Cpu, ArrowRight, Check } from 'lucide-react'
import Link from 'next/link'

interface OnboardingProps {
  hasKeys: boolean
  hasTools: boolean
  hasProvider: boolean
}

const steps = [
  {
    id: 'keys',
    title: 'Create an API Key',
    desc: 'Generate your first API key to authenticate requests to the Memzent Gateway.',
    icon: Key,
    href: '/keys',
    color: 'text-memzent-purple',
    check: 'hasKeys',
  },
  {
    id: 'provider',
    title: 'Connect a Provider',
    desc: 'Add OpenAI, Anthropic, or Ollama credentials so Memzent can route your prompts.',
    icon: Cpu,
    href: '/providers',
    color: 'text-memzent-glow',
    check: 'hasProvider',
  },
  {
    id: 'tools',
    title: 'Register Your First Tool',
    desc: 'Add an MCP tool to enable semantic routing and agent workflows.',
    icon: Database,
    href: '/tools',
    color: 'text-memzent-accent',
    check: 'hasTools',
  },
]

export function OnboardingChecklist({ hasKeys, hasTools, hasProvider }: OnboardingProps) {
  const checks: Record<string, boolean> = { hasKeys, hasTools, hasProvider }
  const completed = steps.filter(s => checks[s.check]).length
  const allDone = completed === steps.length

  if (allDone) return null

  return (
    <div className="glass rounded-3xl p-8 border-memzent-glow/10 relative overflow-hidden">
      <div className="absolute inset-0 bg-gradient-to-br from-memzent-glow/5 via-transparent to-memzent-purple/5 pointer-events-none" />
      <div className="relative z-10 space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Zap size={20} className="text-memzent-glow" />
            <h3 className="text-lg font-black tracking-tight uppercase">Get Started</h3>
          </div>
          <span className="text-xs font-black text-white/30">{completed}/{steps.length} complete</span>
        </div>

        {/* Progress bar */}
        <div className="w-full h-1.5 bg-white/5 rounded-full overflow-hidden">
          <div
            className="h-full bg-gradient-to-r from-memzent-glow to-memzent-accent rounded-full transition-all duration-500"
            style={{ width: `${(completed / steps.length) * 100}%` }}
          />
        </div>

        <div className="space-y-4">
          {steps.map((step) => {
            const done = checks[step.check]
            return (
              <Link
                key={step.id}
                href={step.href}
                className={`flex items-center gap-4 p-4 rounded-2xl border transition-all ${
                  done
                    ? 'bg-white/[0.02] border-white/5 opacity-50'
                    : 'bg-white/[0.03] border-white/10 hover:border-memzent-glow/20 hover:bg-white/[0.05]'
                }`}
              >
                <div className={`w-10 h-10 rounded-xl flex items-center justify-center border ${
                  done
                    ? 'bg-memzent-accent/10 border-memzent-accent/20'
                    : `bg-white/5 border-white/10 ${step.color}`
                }`}>
                  {done ? <Check size={18} className="text-memzent-accent" /> : <step.icon size={18} />}
                </div>
                <div className="flex-1 min-w-0">
                  <div className={`text-sm font-black ${done ? 'line-through text-white/40' : 'text-white'}`}>{step.title}</div>
                  <div className="text-[11px] text-white/30 font-medium truncate">{step.desc}</div>
                </div>
                {!done && <ArrowRight size={14} className="text-white/20" />}
              </Link>
            )
          })}
        </div>
      </div>
    </div>
  )
}
