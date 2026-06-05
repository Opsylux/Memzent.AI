import { Book, ExternalLink, Terminal, Key, Database, Cpu } from 'lucide-react'
import Link from 'next/link'

const docs = [
  {
    title: 'Quickstart',
    desc: 'Get your first agent routing through Memzent in under 5 minutes.',
    icon: Terminal,
    href: 'https://docs.memzent.ai/quickstart',
    color: 'text-memzent-glow',
  },
  {
    title: 'API Reference',
    desc: 'Full REST API documentation for the Memzent Gateway.',
    icon: Key,
    href: 'https://docs.memzent.ai/api',
    color: 'text-memzent-purple',
  },
  {
    title: 'MCP Tool Registry',
    desc: 'How to register, discover, and invoke MCP tools through Memzent.',
    icon: Database,
    href: 'https://docs.memzent.ai/tools',
    color: 'text-memzent-accent',
  },
  {
    title: 'Provider Setup',
    desc: 'Configure OpenAI, Anthropic, Gemini, or Ollama as your LLM backend.',
    icon: Cpu,
    href: 'https://docs.memzent.ai/providers',
    color: 'text-memzent-glow',
  },
]

export default function DocsPage() {
  return (
    <div className="space-y-10 pb-20">
      <div className="flex items-center gap-4 mb-4">
        <div className="w-2 h-8 rounded-full bg-gradient-to-b from-memzent-glow to-memzent-purple" />
        <div>
          <h1 className="text-3xl font-black tracking-tighter uppercase">Documentation</h1>
          <p className="text-[10px] font-black text-white/20 uppercase tracking-[0.3em] italic">Guides & API Reference</p>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {docs.map((doc) => (
          <a
            key={doc.title}
            href={doc.href}
            target="_blank"
            rel="noopener noreferrer"
            className="group p-8 rounded-3xl bg-white/[0.02] border border-white/5 hover:border-memzent-glow/20 transition-all flex flex-col gap-4"
          >
            <div className="flex items-center justify-between">
              <div className={`w-12 h-12 rounded-2xl bg-white/5 border border-white/10 flex items-center justify-center ${doc.color}`}>
                <doc.icon size={24} />
              </div>
              <ExternalLink size={14} className="text-white/20 group-hover:text-memzent-glow transition-colors" />
            </div>
            <h3 className="text-lg font-black tracking-tight">{doc.title}</h3>
            <p className="text-sm font-medium text-white/40 leading-relaxed">{doc.desc}</p>
          </a>
        ))}
      </div>

      <div className="glass rounded-3xl p-8 border-white/5">
        <div className="flex items-center gap-3 mb-4">
          <Book size={18} className="text-memzent-glow" />
          <h3 className="text-sm font-black uppercase tracking-widest">Need Help?</h3>
        </div>
        <p className="text-sm text-white/50 font-medium">
          Join our Discord community or open an issue on{' '}
          <a href="https://github.com/Opsylux/Memzent.AI" target="_blank" rel="noopener" className="text-memzent-glow hover:underline">
            GitHub
          </a>{' '}
          for support.
        </p>
      </div>
    </div>
  )
}
