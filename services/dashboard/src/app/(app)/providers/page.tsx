import { getCurrentOrg } from "@/lib/user-context"
import { getMemzentStats } from "@/app/actions"
import { Brain, Cpu, Database, Zap, Layers, CheckCircle, Globe } from "lucide-react"

const PROVIDER_COSTS: Record<string, { input: string; output: string; color: string }> = {
  openai: { input: "$2.50", output: "$10.00", color: "text-green-400" },
  anthropic: { input: "$3.00", output: "$15.00", color: "text-orange-400" },
  ollama: { input: "$0.10*", output: "$0.10*", color: "text-memzent-glow" },
  gemini: { input: "$1.25", output: "$5.00", color: "text-blue-400" },
}

export default async function ProvidersPage() {
  const org = await getCurrentOrg()
  const stats = await getMemzentStats(org?.orgId)

  const activeProviders: string[] = Array.isArray(stats.active_providers)
    ? stats.active_providers.map((p: string) => p.toLowerCase())
    : []

  const defaultProvider = (stats.default_provider || "ollama").toLowerCase()

  const allProviders = [
    {
      id: "ollama",
      name: "Ollama",
      tagline: "Local open-source LLM engine",
      models: ["llama3.2", "llama3", "mistral", "phi3", "qwen2.5-coder", "qwen3.6"],
      icon: <Brain size={28} />,
      type: "Local",
      note: "* Infrastructure cost only — no API key required",
    },
    {
      id: "openai",
      name: "OpenAI",
      tagline: "GPT-4o & GPT-5 series",
      models: ["gpt-4o", "gpt-4o-mini", "gpt-5.1"],
      icon: <Cpu size={28} />,
      type: "Cloud",
    },
    {
      id: "anthropic",
      name: "Anthropic",
      tagline: "Claude 3.x series",
      models: ["claude-3-5-sonnet-20241022", "claude-3-opus-20240229", "claude-3-haiku-20240307"],
      icon: <Layers size={28} />,
      type: "Cloud",
    },
    {
      id: "gemini",
      name: "Google Gemini",
      tagline: "Gemini 1.5 Pro & Flash",
      models: ["gemini-1.5-pro", "gemini-1.5-flash", "gemini-2.0-flash"],
      icon: <Globe size={28} />,
      type: "Cloud",
    },
  ]

  return (
    <div className="space-y-12 pb-20">
      {/* Header */}
      <div className="flex items-center gap-4 mb-4">
        <div className="w-2 h-8 rounded-full bg-gradient-to-b from-memzent-purple to-memzent-glow" />
        <div>
          <h1 className="text-3xl font-black tracking-tighter uppercase">Provider Mesh</h1>
          <p className="text-[10px] font-black text-white/20 uppercase tracking-[0.3em] italic">LLM Compute Node Discovery & Cost Ledger</p>
        </div>
      </div>

      {/* Stats Row */}
      <section className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="stat-card neural-bg border-white/5 p-6">
          <div className="text-[10px] font-black uppercase tracking-widest text-white/20 mb-2">Configured Providers</div>
          <div className="text-3xl font-black">{allProviders.length}</div>
        </div>
        <div className="stat-card border-memzent-glow/10 neural-bg p-6">
          <div className="text-[10px] font-black uppercase tracking-widest text-memzent-glow/60 mb-2">Active in Engine</div>
          <div className="text-3xl font-black text-memzent-glow">{activeProviders.length || 1}</div>
        </div>
        <div className="stat-card border-memzent-purple/10 neural-bg p-6">
          <div className="text-[10px] font-black uppercase tracking-widest text-memzent-purple/60 mb-2">Default Provider</div>
          <div className="text-xl font-black capitalize text-memzent-purple">{defaultProvider}</div>
        </div>
      </section>

      {/* Provider Cards */}
      <section className="grid grid-cols-1 md:grid-cols-2 gap-8">
        {allProviders.map((provider) => {
          const isActive = activeProviders.includes(provider.id) || provider.id === defaultProvider
          const isDefault = provider.id === defaultProvider
          const cost = PROVIDER_COSTS[provider.id]

          return (
            <div
              key={provider.id}
              className={`stat-card neural-bg p-8 border transition-all duration-300 ${
                isDefault
                  ? "border-memzent-glow/30 shadow-[0_0_30px_rgba(0,243,255,0.05)]"
                  : isActive
                    ? "border-memzent-purple/20"
                    : "border-white/5 opacity-60"
              }`}
            >
              {/* Card Header */}
              <div className="flex items-start justify-between mb-6">
                <div className="flex items-center gap-4">
                  <div className={`w-14 h-14 rounded-2xl flex items-center justify-center border ${
                    isDefault
                      ? "bg-memzent-glow/10 border-memzent-glow/20 text-memzent-glow shadow-[0_0_15px_rgba(0,243,255,0.15)]"
                      : isActive
                        ? "bg-memzent-purple/10 border-memzent-purple/20 text-memzent-purple"
                        : "bg-white/5 border-white/10 text-white/20"
                  }`}>
                    {provider.icon}
                  </div>
                  <div>
                    <h3 className="text-xl font-black tracking-tight">{provider.name}</h3>
                    <p className="text-[10px] font-bold text-white/30 uppercase tracking-widest mt-1">{provider.tagline}</p>
                  </div>
                </div>
                <div className="flex flex-col items-end gap-2">
                  <span className={`text-[9px] font-black uppercase tracking-widest px-2 py-1 rounded-md border ${
                    provider.type === "Local"
                      ? "text-memzent-accent border-memzent-accent/20 bg-memzent-accent/5"
                      : "text-white/40 border-white/10 bg-white/5"
                  }`}>
                    {provider.type}
                  </span>
                  {isDefault && (
                    <span className="text-[9px] font-black uppercase tracking-widest px-2 py-1 rounded-md border text-memzent-glow border-memzent-glow/20 bg-memzent-glow/5 flex items-center gap-1">
                      <Zap size={8} /> Default
                    </span>
                  )}
                  {isActive && !isDefault && (
                    <span className="text-[9px] font-black uppercase tracking-widest px-2 py-1 rounded-md border text-memzent-purple border-memzent-purple/20 bg-memzent-purple/5 flex items-center gap-1">
                      <CheckCircle size={8} /> Active
                    </span>
                  )}
                </div>
              </div>

              {/* Cost Row */}
              {cost && (
                <div className="grid grid-cols-2 gap-4 mb-6 p-4 rounded-2xl bg-black/30 border border-white/5">
                  <div>
                    <div className="text-[9px] font-black uppercase tracking-widest text-white/20 mb-1">Input / 1M Tokens</div>
                    <div className={`text-base font-black font-mono ${cost.color}`}>{cost.input}</div>
                  </div>
                  <div>
                    <div className="text-[9px] font-black uppercase tracking-widest text-white/20 mb-1">Output / 1M Tokens</div>
                    <div className={`text-base font-black font-mono ${cost.color}`}>{cost.output}</div>
                  </div>
                </div>
              )}

              {/* Cache savings note */}
              <div className="flex items-center gap-2 mb-6 px-1">
                <Zap size={10} className="text-memzent-accent" />
                <span className="text-[9px] font-black uppercase tracking-widest text-memzent-accent/60">80% discount applied on semantic cache hits</span>
              </div>

              {/* Model List */}
              <div>
                <div className="text-[9px] font-black uppercase tracking-widest text-white/20 mb-3">Available Models</div>
                <div className="flex flex-wrap gap-2">
                  {provider.models.map(m => (
                    <span
                      key={m}
                      className="text-[9px] font-black font-mono px-3 py-1.5 rounded-lg bg-white/5 border border-white/5 text-white/50 hover:text-white hover:border-white/10 transition-all"
                    >
                      {m}
                    </span>
                  ))}
                </div>
              </div>

              {provider.note && (
                <p className="text-[9px] font-black uppercase tracking-widest text-white/10 mt-4 pt-4 border-t border-white/5 italic">{provider.note}</p>
              )}
            </div>
          )
        })}
      </section>

      {/* Cache Economy Banner */}
      <section className="stat-card neural-bg border-memzent-accent/10 p-10 relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-r from-memzent-glow/5 via-transparent to-memzent-purple/5 pointer-events-none" />
        <div className="relative z-10 flex flex-col md:flex-row items-center justify-between gap-8">
          <div className="space-y-3 max-w-xl">
            <h3 className="text-2xl font-black tracking-tighter uppercase">Cache Economy</h3>
            <p className="text-[10px] font-black text-white/30 uppercase tracking-[0.2em] leading-relaxed">
              Memzent's semantic cache intercepts repeat intents across your organization. Cache hits are charged at an 80% discount — your infra cost only. The gateway automatically routes identical semantic intents to the cache before touching the LLM.
            </p>
          </div>
          <div className="grid grid-cols-3 gap-6 text-center">
            {[
              { label: "LLM Cost", value: "$2.50", sub: "per 1M tokens", color: "text-white/60" },
              { label: "Cache Cost", value: "$0.50", sub: "80% savings", color: "text-memzent-accent" },
              { label: "Your Savings", value: "80%", sub: "avg per request", color: "text-memzent-glow" },
            ].map(item => (
              <div key={item.label} className="p-4 rounded-2xl bg-white/5 border border-white/5">
                <div className="text-[9px] font-black uppercase tracking-widest text-white/20 mb-1">{item.label}</div>
                <div className={`text-2xl font-black ${item.color}`}>{item.value}</div>
                <div className="text-[9px] font-bold text-white/20 mt-1">{item.sub}</div>
              </div>
            ))}
          </div>
        </div>
      </section>
    </div>
  )
}
