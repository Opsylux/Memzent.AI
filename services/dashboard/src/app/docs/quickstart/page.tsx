import { CodeBlock } from "@/components/docs/code-block";
import { ArrowRight, Key, Zap, CheckCircle2, Terminal } from "lucide-react";
import Link from "next/link";
import { DocsPager } from "@/components/docs/docs-pager";
import { DOCS_CONFIG } from "@/config/docs-config";

export default function QuickStart() {
  const curlExample = `curl -X POST https://${DOCS_CONFIG.domain}/v1/chat \\
  -H "X-API-Key: memzent_f7c9...8e2a" \\
  -H "Content-Type: application/json" \\
  -d '{
    "prompt": "Find all high-priority tickets from the last 24 hours"
  }'`;

  const responseExample = `{
  "text": "There are 3 high-priority tickets opened in the last 24 hours...",
  "cached": false,
  "provider": "ollama"
}`;

  const nodeExample = `import { MemzentClient } from "@opsylux/memzent-mcp";

const memzent = new MemzentClient({
  apiKey: process.env.MEMZENT_API_KEY,
  endpoint: "https://${DOCS_CONFIG.domain}"
});

const response = await memzent.chat("Summarize activity for org_01");
console.log(response.text);`;

  return (
    <div className="space-y-12">
      <header className="space-y-4">
        <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-memzent-glow/5 border border-memzent-glow/20 w-fit">
          <span className="text-[10px] font-black text-memzent-glow uppercase tracking-tighter italic">Getting_Started</span>
        </div>
        <h1 className="text-4xl font-black tracking-tighter uppercase sm:text-5xl">Quick Start Guide</h1>
        <p className="text-lg text-white/60 leading-relaxed font-medium">
          Get Memzent up and running in under 5 minutes. All you need is an API key and one HTTP request.
        </p>
      </header>

      {/* Step 1 */}
      <section className="space-y-6">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 rounded-full bg-memzent-glow/10 border border-memzent-glow/20 flex items-center justify-center text-xs font-black text-memzent-glow">1</div>
          <h2 className="text-2xl font-black tracking-tighter uppercase">Get Your API Key</h2>
        </div>
        <div className="space-y-4 pl-12">
          <p className="text-sm text-white/60 leading-relaxed font-medium">
            Go to the <a href="/keys" className="text-memzent-glow underline font-bold">API Keys</a> section of your Dashboard and click <strong className="text-white">+ Generate Secret Key</strong>. Copy it immediately — it is only shown once.
          </p>
          <div className="p-4 rounded-xl bg-memzent-glow/5 border border-memzent-glow/10 flex items-start gap-3">
            <Key size={16} className="text-memzent-glow mt-0.5 shrink-0" />
            <p className="text-xs text-memzent-glow font-bold leading-relaxed">
              Keep your API key secret. Never include it in client-side JavaScript or expose it in a public repository.
            </p>
          </div>
        </div>
      </section>

      {/* Step 2 */}
      <section className="space-y-6">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 rounded-full bg-memzent-glow/10 border border-memzent-glow/20 flex items-center justify-center text-xs font-black text-memzent-glow">2</div>
          <h2 className="text-2xl font-black tracking-tighter uppercase">Send Your First Request</h2>
        </div>
        <div className="space-y-5 pl-12">
          <p className="text-sm text-white/60 leading-relaxed font-medium">
            Send a <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">POST</code> request to the chat endpoint. Include your key in the <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">X-API-Key</code> header.
          </p>
          <CodeBlock code={curlExample} language="bash" filename="cURL" />

          <p className="text-sm text-white/60 leading-relaxed font-medium">You will receive a structured JSON response:</p>
          <CodeBlock code={responseExample} language="json" filename="Response" />
        </div>
      </section>

      {/* Step 3 — Node.js */}
      <section className="space-y-6">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 rounded-full bg-memzent-glow/10 border border-memzent-glow/20 flex items-center justify-center text-xs font-black text-memzent-glow">3</div>
          <h2 className="text-2xl font-black tracking-tighter uppercase">Use the SDK (Optional)</h2>
        </div>
        <div className="space-y-4 pl-12">
          <p className="text-sm text-white/60 leading-relaxed font-medium">
            If you are building a Node.js application, the Memzent SDK handles authentication, retries, and streaming for you.
          </p>
          <CodeBlock code={nodeExample} language="typescript" filename="Node.js SDK" />
        </div>
      </section>

      {/* Step 4 — Check the trace */}
      <section className="space-y-6">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 rounded-full bg-memzent-glow/10 border border-memzent-glow/20 flex items-center justify-center text-xs font-black text-memzent-glow">4</div>
          <h2 className="text-2xl font-black tracking-tighter uppercase">Inspect the Execution Trace</h2>
        </div>
        <div className="space-y-4 pl-12">
          <p className="text-sm text-white/60 leading-relaxed font-medium">
            After your first request, check your Dashboard. You will see a real-time trace showing how Memzent processed the prompt, which tools were called, and whether the response came from memory or a model.
          </p>
          <div className="flex flex-wrap gap-3">
            {["Auth Verified", "Cache Checked", "Tools Matched", "Response Generated"].map((step) => (
              <div key={step} className="flex items-center gap-2 px-3 py-2 rounded-xl bg-white/[0.02] border border-white/5">
                <CheckCircle2 size={13} className="text-memzent-accent" />
                <span className="text-[10px] font-black uppercase text-white/40">{step}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="pt-10 border-t border-white/5">
        <div className="p-8 rounded-2xl bg-gradient-to-br from-memzent-purple/10 to-transparent border border-memzent-purple/20 flex flex-col items-center text-center gap-5">
          <Zap size={28} className="text-memzent-purple animate-pulse" />
          <h3 className="text-xl font-black uppercase tracking-tighter">Ready to go deeper?</h3>
          <p className="text-sm text-white/40 max-w-md font-bold leading-relaxed">
            Learn how to connect your own tools, pick specific AI models per request, and manage team permissions.
          </p>
          <div className="flex flex-wrap items-center gap-3">
            <Link href="/docs/first-request" className="flex items-center gap-2 px-5 py-3 rounded-xl bg-memzent-glow text-black text-xs font-black uppercase tracking-widest hover:scale-105 transition-all">
              Explore Model Selection <ArrowRight size={13} />
            </Link>
            <Link href="/docs/tool-registry" className="text-xs text-white/40 font-black uppercase tracking-widest hover:text-white transition-colors">
              Connect Tools →
            </Link>
          </div>
        </div>
      </section>

      <DocsPager />
    </div>
  );
}
