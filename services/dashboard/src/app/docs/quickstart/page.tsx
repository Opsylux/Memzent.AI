import { CodeBlock } from "@/components/docs/code-block";
import { ArrowRight, Key, Zap, CheckCircle2 } from "lucide-react";

export default function QuickStart() {
  const curlExample = `curl -X POST https://aura.gateway.yourdomain.com/v1/chat \\
  -H "X-API-Key: aura_f7c9...8e2a" \\
  -H "Content-Type: application/json" \\
  -d '{
    "message": "Find all high-priority audit logs from the last 24 hours"
  }'`;

  const nodeExample = `import { AuraClient } from "@opsylux/aura-mcp";

const aura = new AuraClient({
  apiKey: process.env.AURA_API_KEY,
  endpoint: "https://aura.gateway.yourdomain.com"
});

const response = await aura.chat("Summarize tool usage for org_01");
console.log(response.text);`;

  return (
    <div className="space-y-12">
      <header className="space-y-4">
        <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-aura-glow/5 border border-aura-glow/20 w-fit">
          <span className="text-[10px] font-black text-aura-glow uppercase tracking-tighter italic">Step_01_Basics</span>
        </div>
        <h1 className="text-4xl font-black tracking-tighter uppercase sm:text-5xl">Quick Start Guide</h1>
        <p className="text-lg text-white/60 leading-relaxed font-medium">
          Get Aura up and running in less than 5 minutes. We'll show you how to authenticate and make your first semantic request.
        </p>
      </header>

      {/* Step 1 */}
      <section className="space-y-6">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 rounded-full bg-white/5 border border-white/10 flex items-center justify-center text-xs font-black text-white/40">1</div>
          <h2 className="text-2xl font-black tracking-tighter uppercase">Obtain an API Key</h2>
        </div>
        <div className="space-y-4 text-sm text-white/60 leading-relaxed font-medium pl-12">
          <p>
            Navigate to the <a href="/keys" className="text-aura-glow underline">API Keys</a> section in your Aura Dashboard. Click <strong>+ Generate Secret Key</strong> and copy the result immediately.
          </p>
          <div className="p-4 rounded-xl bg-aura-glow/5 border border-aura-glow/10 flex items-start gap-3">
             <Key size={16} className="text-aura-glow mt-0.5" />
             <p className="text-xs text-aura-glow font-bold">
               Aura API Keys are only shown once during generation. Keep them secure and never expose them in client-side code.
             </p>
          </div>
        </div>
      </section>

      {/* Step 2 */}
      <section className="space-y-6">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 rounded-full bg-white/5 border border-white/10 flex items-center justify-center text-xs font-black text-white/40">2</div>
          <h2 className="text-2xl font-black tracking-tighter uppercase">Make Your First Request</h2>
        </div>
        <div className="space-y-4 text-sm text-white/60 leading-relaxed font-medium pl-12">
          <p>
            Send a POST request to the Gateway's chat endpoint. Include your key in the <code className="text-aura-glow bg-white/5 px-1 rounded">X-API-Key</code> header.
          </p>
          
          <div className="pt-4">
             <CodeBlock 
                code={curlExample} 
                language="bash" 
                filename="terminal" 
             />
          </div>
        </div>
      </section>

      {/* Step 3 */}
      <section className="space-y-6">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 rounded-full bg-white/5 border border-white/10 flex items-center justify-center text-xs font-black text-white/40">3</div>
          <h2 className="text-2xl font-black tracking-tighter uppercase">Analyze the Execution Trace</h2>
        </div>
        <div className="space-y-4 text-sm text-white/60 leading-relaxed font-medium pl-12">
          <p>
            Check your Dashboard to see the real-time **Neural Execution Trace**. You’ll see how Aura evaluated the intent, checked the semantic cache, and routed the request to the appropriate tool.
          </p>
          <div className="flex flex-wrap gap-4 pt-4">
             <div className="flex items-center gap-2 px-3 py-2 rounded-xl bg-white/[0.02] border border-white/5">
                <CheckCircle2 size={14} className="text-aura-accent" />
                <span className="text-[10px] font-black uppercase text-white/40">Auth Verified</span>
             </div>
             <div className="flex items-center gap-2 px-3 py-2 rounded-xl bg-white/[0.02] border border-white/5">
                <CheckCircle2 size={14} className="text-aura-accent" />
                <span className="text-[10px] font-black uppercase text-white/40">Intent Mapped</span>
             </div>
             <div className="flex items-center gap-2 px-3 py-2 rounded-xl bg-white/[0.02] border border-white/5">
                <CheckCircle2 size={14} className="text-aura-accent" />
                <span className="text-[10px] font-black uppercase text-white/40">Response Cached</span>
             </div>
          </div>
        </div>
      </section>

      <section className="pt-12 border-t border-white/5">
        <div className="stat-card p-8 bg-gradient-to-br from-aura-purple/10 to-transparent border-aura-purple/20 relative overflow-hidden flex flex-col items-center text-center gap-6">
           <Zap size={32} className="text-aura-purple mx-auto animate-pulse" />
           <h3 className="text-2xl font-black uppercase tracking-tighter">Ready for Production?</h3>
           <p className="text-sm text-white/40 max-w-md font-bold leading-relaxed">
             Move beyond hello-world and explore our deep-dives into Multi-Tenant RBAC and Vector Matching algorithms.
           </p>
           <div className="flex items-center gap-4">
              <Link href="/docs/rbac" className="flex items-center gap-2 px-6 py-3 rounded-xl bg-aura-glow text-black text-xs font-black uppercase tracking-widest hover:scale-105 transition-all">
                Explore RBAC <ArrowRight size={14} />
              </Link>
           </div>
        </div>
      </section>
    </div>
  );
}
