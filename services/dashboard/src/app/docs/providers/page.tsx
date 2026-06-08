import { Cpu, Zap, Globe, Settings } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";
import { CodeBlock } from "@/components/docs/code-block";
import { DOCS_CONFIG } from "@/config/docs-config";

export default function ProvidersPage() {
  const listProviders = `curl -X GET https://${DOCS_CONFIG.domain}/v1/providers \\
  -H "X-API-Key: memzent_YOUR_KEY"`;

  const listModels = `curl -X GET https://${DOCS_CONFIG.domain}/v1/models \\
  -H "X-API-Key: memzent_YOUR_KEY"`;

  const overrideExample = `curl -X POST https://${DOCS_CONFIG.domain}/v1/chat \\
  -H "X-API-Key: memzent_YOUR_KEY" \\
  -H "X-Memzent-Provider: anthropic" \\
  -H "X-Memzent-Model: claude-sonnet-4-20250514" \\
  -H "Content-Type: application/json" \\
  -d '{
    "messages": [{"role": "user", "content": "Explain recursion"}]
  }'`;

  const bodyOverride = `curl -X POST https://${DOCS_CONFIG.domain}/v1/chat \\
  -H "X-API-Key: memzent_YOUR_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "messages": [{"role": "user", "content": "Write a haiku about coding"}],
    "provider": "gemini",
    "model": "gemini-2.5-pro"
  }'`;

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-3 mb-4">
        <div className="p-2 rounded-xl bg-memzent-glow/10 border border-memzent-glow/20">
          <Cpu size={20} className="text-memzent-glow" />
        </div>
        <h1 className="text-3xl font-black tracking-tight">Providers & Models</h1>
      </div>
      <p className="text-white/50 text-sm leading-relaxed mb-10">
        Memzent supports multiple LLM providers simultaneously. Switch between them per-request
        with zero code changes — the caching, billing, and RBAC layers work identically across all providers.
      </p>

      {/* Provider Table */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Globe size={16} className="text-memzent-glow" />
          Supported Providers
        </h2>

        <div className="overflow-x-auto mb-6">
          <table className="w-full text-xs border border-white/5 rounded-lg overflow-hidden">
            <thead>
              <tr className="bg-white/[0.03] border-b border-white/5">
                <th className="text-left px-4 py-2 font-black text-white/60">Provider</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Default Model</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Models</th>
                <th className="text-left px-4 py-2 font-black text-white/60">Requires</th>
              </tr>
            </thead>
            <tbody className="text-white/40">
              <tr className="border-b border-white/5">
                <td className="px-4 py-2 font-mono text-memzent-glow/70">ollama</td>
                <td className="px-4 py-2">llama3.2</td>
                <td className="px-4 py-2">llama3.2, llama3, mistral, phi3 + any locally installed</td>
                <td className="px-4 py-2">OLLAMA_URL</td>
              </tr>
              <tr className="border-b border-white/5">
                <td className="px-4 py-2 font-mono text-memzent-glow/70">openai</td>
                <td className="px-4 py-2">gpt-4o-mini</td>
                <td className="px-4 py-2">gpt-4o-mini, gpt-4, gpt-4-turbo, gpt-3.5-turbo</td>
                <td className="px-4 py-2">OPENAI_API_KEY</td>
              </tr>
              <tr className="border-b border-white/5">
                <td className="px-4 py-2 font-mono text-memzent-glow/70">anthropic</td>
                <td className="px-4 py-2">claude-sonnet-4-20250514</td>
                <td className="px-4 py-2">claude-sonnet-4-20250514, claude-opus-4-20250514, claude-3-5-sonnet-20241022, claude-3-5-haiku-20241022</td>
                <td className="px-4 py-2">ANTHROPIC_API_KEY</td>
              </tr>
              <tr>
                <td className="px-4 py-2 font-mono text-memzent-glow/70">gemini</td>
                <td className="px-4 py-2">gemini-2.5-flash</td>
                <td className="px-4 py-2">gemini-2.5-flash, gemini-2.5-pro, gemini-2.0-flash, gemini-1.5-pro</td>
                <td className="px-4 py-2">GEMINI_API_KEY</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div className="p-4 rounded-xl border border-yellow-500/20 bg-yellow-500/5 mb-6">
          <p className="text-xs text-yellow-200/70">
            <strong>Note:</strong> Only providers with configured API keys are available at runtime.
            Ollama is always available if reachable. Models are dynamically discovered from each provider on startup.
          </p>
        </div>
      </section>

      {/* Discovering */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Zap size={16} className="text-memzent-glow" />
          Discovering Available Providers
        </h2>
        <CodeBlock code={listProviders} language="bash" title="List Providers" />
        <CodeBlock code={listModels} language="bash" title="List All Models" />
      </section>

      {/* Overriding */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Settings size={16} className="text-memzent-glow" />
          Switching Providers Per-Request
        </h2>
        <p className="text-white/50 text-sm mb-4">
          Override the provider and model using either <strong>headers</strong> or <strong>request body</strong> fields.
          Header values take priority over body fields.
        </p>

        <h4 className="text-xs font-black uppercase tracking-widest text-white/40 mb-2 mt-6">Option 1: Headers</h4>
        <CodeBlock code={overrideExample} language="bash" title="Header Override" />

        <h4 className="text-xs font-black uppercase tracking-widest text-white/40 mb-2 mt-6">Option 2: Body Fields</h4>
        <CodeBlock code={bodyOverride} language="bash" title="Body Override" />
      </section>

      {/* Priority */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4">Resolution Priority</h2>
        <div className="space-y-2 text-sm text-white/50">
          <div className="flex items-center gap-3">
            <span className="text-memzent-glow font-mono text-xs w-4">1.</span>
            <span><code className="text-memzent-glow/70">X-Memzent-Provider</code> / <code className="text-memzent-glow/70">X-Memzent-Model</code> headers</span>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-memzent-glow font-mono text-xs w-4">2.</span>
            <span><code className="text-memzent-glow/70">provider</code> / <code className="text-memzent-glow/70">model</code> fields in request body</span>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-memzent-glow font-mono text-xs w-4">3.</span>
            <span>Org-level default model (from settings)</span>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-memzent-glow font-mono text-xs w-4">4.</span>
            <span>Gateway default provider (configured at deploy time)</span>
          </div>
        </div>
      </section>

      {/* Cache scoping */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4">Model-Scoped Caching</h2>
        <p className="text-white/50 text-sm leading-relaxed">
          Cache entries are scoped by <strong>org + model</strong>. The same prompt sent to
          <code className="text-memzent-glow/70 mx-1">gpt-4o-mini</code> and
          <code className="text-memzent-glow/70 mx-1">claude-sonnet-4-20250514</code> will produce
          separate cache entries — you&apos;ll never get a GPT response when requesting Claude.
        </p>
      </section>

      <DocsPager />
    </div>
  );
}
