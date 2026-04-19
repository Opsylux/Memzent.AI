import { Code, Search, Zap } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";
import { CodeBlock } from "@/components/docs/code-block";
import { DOCS_CONFIG } from "@/config/docs-config";

export default function FirstRequestPage() {
  const providersExample = `curl https://${DOCS_CONFIG.domain}/v1/providers \\
  -H "X-API-Key: your_key"`;

  const providersResponse = `[
  { "name": "openai",    "default_model": "gpt-4o",           "supported_models": ["gpt-4o", "gpt-4-turbo", "gpt-3.5-turbo"] },
  { "name": "anthropic", "default_model": "claude-3-5-sonnet", "supported_models": ["claude-3-5-sonnet", "claude-3-haiku"] },
  { "name": "ollama",    "default_model": "llama3.2",          "supported_models": ["llama3.2", "mistral", "phi3"] }
]`;

  const modelsExample = `curl -X POST https://${DOCS_CONFIG.domain}/v1/chat \\
  -H "X-API-Key: your_key" \\
  -H "X-Aura-Provider: openai" \\
  -H "X-Aura-Model: gpt-4-turbo" \\
  -d '{"prompt": "Summarize the latest activity"}'`;

  return (
    <div className="space-y-12">
      <header className="space-y-4">
        <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-aura-glow/5 border border-aura-glow/20 w-fit">
          <span className="text-[10px] font-black text-aura-glow uppercase tracking-tighter italic">Model_Selection</span>
        </div>
        <h1 className="text-4xl font-black tracking-tighter uppercase sm:text-5xl">Choosing a Model</h1>
        <p className="text-lg text-white/60 leading-relaxed font-medium">
          Aura connects to multiple AI providers simultaneously. You can let Aura pick the default model, or specify exactly which model you want on a per-request basis.
        </p>
      </header>

      {/* Discover */}
      <section className="space-y-5">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 rounded-full bg-aura-glow/10 border border-aura-glow/20 flex items-center justify-center text-xs font-black text-aura-glow">1</div>
          <h2 className="text-2xl font-black tracking-tighter uppercase">Discover Available Models</h2>
        </div>
        <div className="space-y-4 pl-12">
          <p className="text-sm text-white/60 leading-relaxed font-medium">
            Use the <code className="text-aura-glow bg-aura-glow/5 px-1 rounded font-mono">/v1/providers</code> endpoint to see which AI providers and models are currently active in your organization&apos;s Aura instance.
          </p>
          <CodeBlock code={providersExample} language="bash" filename="GET /v1/providers" />
          <CodeBlock code={providersResponse} language="json" filename="Response" />
        </div>
      </section>

      {/* Select model */}
      <section className="space-y-5">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 rounded-full bg-aura-glow/10 border border-aura-glow/20 flex items-center justify-center text-xs font-black text-aura-glow">2</div>
          <h2 className="text-2xl font-black tracking-tighter uppercase">Select a Model Per Request</h2>
        </div>
        <div className="space-y-5 pl-12">
          <p className="text-sm text-white/60 leading-relaxed font-medium">
            Override the default model using two optional headers. This lets you route different tasks to the most appropriate model — for example, a faster model for simple queries and a more capable one for complex analysis.
          </p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="p-4 rounded-xl bg-white/[0.02] border border-white/5 space-y-2">
              <span className="text-[10px] font-black uppercase text-aura-glow">Request Header</span>
              <div className="text-xs font-mono text-white/80">X-Aura-Provider</div>
              <p className="text-[10px] text-white/40 font-bold leading-relaxed">Selects the provider — for example <code className="text-aura-glow">openai</code>, <code className="text-aura-glow">anthropic</code>, or <code className="text-aura-glow">ollama</code>.</p>
            </div>
            <div className="p-4 rounded-xl bg-white/[0.02] border border-white/5 space-y-2">
              <span className="text-[10px] font-black uppercase text-aura-purple">Request Header</span>
              <div className="text-xs font-mono text-white/80">X-Aura-Model</div>
              <p className="text-[10px] text-white/40 font-bold leading-relaxed">Selects the specific model — for example <code className="text-aura-purple">gpt-4-turbo</code> or <code className="text-aura-purple">claude-3-5-sonnet</code>.</p>
            </div>
          </div>
          <CodeBlock code={modelsExample} language="bash" filename="cURL — Model Override" />
        </div>
      </section>

      {/* Default model */}
      <section className="space-y-5">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 rounded-full bg-aura-glow/10 border border-aura-glow/20 flex items-center justify-center text-xs font-black text-aura-glow">3</div>
          <h2 className="text-2xl font-black tracking-tighter uppercase">The Default Model</h2>
        </div>
        <div className="space-y-4 pl-12">
          <p className="text-sm text-white/60 leading-relaxed font-medium">
            If you don&apos;t specify a provider or model, Aura uses the default configured for your organization. You can see which model that is by checking the <code className="text-aura-glow bg-aura-glow/5 px-1 rounded font-mono">/v1/providers</code> response — the first entry with <code className="text-aura-glow bg-aura-glow/5 px-1 rounded font-mono">default_model</code> is what will be used.
          </p>
        </div>
      </section>

      {/* Trace info */}
      <div className="p-6 rounded-2xl bg-gradient-to-br from-aura-glow/10 to-transparent border border-aura-glow/20">
        <div className="flex items-center gap-3 mb-3">
          <Search size={18} className="text-aura-glow" />
          <h3 className="text-sm font-black uppercase tracking-tight text-white">Track Which Model Was Used</h3>
        </div>
        <p className="text-xs text-white/50 font-bold leading-relaxed">
          Every response includes a <code className="text-aura-glow">&quot;provider&quot;</code> field in the JSON body confirming which model generated the answer. You can also see this in the Dashboard trace view for each request.
        </p>
      </div>

      <DocsPager />
    </div>
  );
}
