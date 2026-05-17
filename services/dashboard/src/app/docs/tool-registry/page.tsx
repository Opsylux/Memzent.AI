import { Cpu, RefreshCw, GitBranch, Shield, Zap, AlertCircle } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";
import { CodeBlock } from "@/components/docs/code-block";
import { DOCS_CONFIG } from "@/config/docs-config";

export default function ToolRegistryPage() {
  const registerToolExample = `curl -X POST https://${DOCS_CONFIG.domain}/v1/tools \\
  -H "X-API-Key: your_admin_key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "id": "customer_lookup",
    "name": "Customer Lookup",
    "description": "Search customer accounts by email, phone, or ID in CRM",
    "connector_type": "rest",
    "endpoint": "https://api.internal.com/crm/customers",
    "timeout_seconds": 10,
    "requires_auth": true
  }'`;

  const manualSyncExample = `# Trigger an immediate sync
curl -X POST https://${DOCS_CONFIG.domain}/v1/tools/sync \\
  -H "X-API-Key: your_admin_key"`;

  const syncStatusExample = `# Check when the last sync ran
curl https://${DOCS_CONFIG.domain}/v1/tools/status \\
  -H "X-API-Key: your_admin_key"

# Response:
# {
#   "status": "healthy",
#   "last_refresh": "2026-04-19T18:30:00Z"
# }`;

  const listToolsExample = `curl https://${DOCS_CONFIG.domain}/v1/tools \\
  -H "X-API-Key: your_key"`;

  return (
    <div className="space-y-12">
      <header className="space-y-4">
        <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-memzent-glow/5 border border-memzent-glow/20 w-fit">
          <span className="text-[10px] font-black text-memzent-glow uppercase tracking-tighter italic">Tool_Registry</span>
        </div>
        <h1 className="text-4xl font-black tracking-tighter uppercase sm:text-5xl">Dynamic Tool Registry</h1>
        <p className="text-lg text-white/60 leading-relaxed font-medium">
          Connect any data source or API to Memzent without restarting anything. New tools are automatically discovered and made available for AI routing within seconds.
        </p>
      </header>

      {/* How It Works */}
      <section className="space-y-6">
        <h2 className="text-2xl font-black tracking-tighter uppercase">How It Works</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {[
            {
              icon: <Cpu size={18} />,
              color: "text-memzent-glow",
              bg: "bg-memzent-glow/10",
              title: "Register via API",
              desc: "Send a tool definition to the registry. It is stored and immediately available for your organization."
            },
            {
              icon: <RefreshCw size={18} />,
              color: "text-memzent-purple",
              bg: "bg-memzent-purple/10",
              title: "Automatic Discovery",
              desc: "Every 30 seconds, Memzent checks for newly registered tools and adds them to the AI routing pool automatically."
            },
            {
              icon: <GitBranch size={18} />,
              color: "text-memzent-accent",
              bg: "bg-memzent-accent/10",
              title: "Instant Routing",
              desc: "Once discovered, Memzent routes relevant prompts to your tool based on the meaning of what the user asked."
            }
          ].map((item) => (
            <div key={item.title} className="p-5 rounded-2xl bg-white/[0.02] border border-white/5 space-y-3">
              <div className={`w-9 h-9 rounded-lg ${item.bg} flex items-center justify-center ${item.color}`}>
                {item.icon}
              </div>
              <h3 className="text-xs font-black uppercase tracking-tight">{item.title}</h3>
              <p className="text-[11px] text-white/40 leading-relaxed font-bold">{item.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Connector Types */}
      <section className="space-y-5 pt-2">
        <h2 className="text-2xl font-black tracking-tighter uppercase">What You Can Connect</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Memzent supports multiple connection types. Choose the one that matches your data source — Memzent handles the integration automatically.
        </p>
        <div className="space-y-2">
          {[
            { label: "REST API", desc: "Any HTTP/JSON API. Provide the URL and Memzent handles authentication and request formatting." },
            { label: "Database", desc: "Query your database directly using a secure, read-only connection string stored in the tool config." },
            { label: "GraphQL", desc: "Run queries and mutations against GraphQL endpoints." },
            { label: "Webhook", desc: "Fire-and-wait for async tools — Memzent sends the request and waits for a response callback." },
            { label: "Internal Service", desc: "High-speed binary protocol for calling internal microservices." },
            { label: "Native Tool", desc: "Built-in Memzent tools like semantic search, available to all organizations automatically." },
          ].map((c) => (
            <div key={c.label} className="flex items-start gap-4 p-3 rounded-xl hover:bg-white/[0.02] transition-colors">
              <span className="text-[10px] font-black text-memzent-glow font-mono bg-memzent-glow/5 px-2 py-1 rounded border border-memzent-glow/10 min-w-[96px] text-center shrink-0">{c.label}</span>
              <p className="text-xs text-white/40 font-bold leading-relaxed">{c.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Registering */}
      <section className="space-y-5 pt-2">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Registering a Tool</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Admins can register a new tool with a single API call. The tool becomes available for routing within 30 seconds.
        </p>
        <CodeBlock code={registerToolExample} language="bash" filename="POST /v1/tools" />

        <div className="p-4 rounded-xl bg-memzent-glow/5 border border-memzent-glow/10 flex items-start gap-3">
          <Zap size={16} className="text-memzent-glow mt-0.5 shrink-0" />
          <p className="text-xs text-memzent-glow font-bold leading-relaxed">
            Write tool descriptions the way a user would ask for the tool — not the way a developer would name it. This directly affects how accurately Memzent can match prompts to your tool.
          </p>
        </div>
      </section>

      {/* Listing */}
      <section className="space-y-5 pt-2">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Listing Available Tools</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Any authenticated user can list the tools available to their organization. The response includes all registered tools and any built-in Memzent capabilities.
        </p>
        <CodeBlock code={listToolsExample} language="bash" filename="GET /v1/tools" />
      </section>

      {/* Sync */}
      <section className="space-y-5 pt-2">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Sync & Health</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Tools are automatically synced every 30 seconds. Admins can also trigger an immediate sync, or check when the last sync ran.
        </p>
        <div className="space-y-5">
          <CodeBlock code={manualSyncExample} language="bash" filename="POST /v1/tools/sync" />
          <CodeBlock code={syncStatusExample} language="bash" filename="GET /v1/tools/status" />
        </div>
      </section>

      {/* Permissions */}
      <section className="space-y-5 pt-2">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Tool Permissions</h2>
        <div className="p-5 rounded-2xl bg-white/[0.02] border border-white/5 space-y-4">
          <div className="flex items-center gap-3">
            <Shield size={16} className="text-memzent-purple" />
            <h3 className="text-sm font-black uppercase tracking-tight">Public vs. Restricted Tools</h3>
          </div>
          <div className="space-y-3 text-[11px] text-white/40 font-bold">
            <p>Set <code className="text-memzent-purple bg-memzent-purple/10 px-1 rounded font-mono">requires_auth: true</code> to restrict the tool to users who have been explicitly granted access by an Admin.</p>
            <p>Set <code className="text-memzent-purple bg-memzent-purple/10 px-1 rounded font-mono">requires_auth: false</code> for tools that any authenticated member of your organization should be able to use — for example, a general knowledge search.</p>
          </div>
        </div>
        <div className="p-4 rounded-xl bg-white/[0.01] border border-white/5 flex items-start gap-3">
          <AlertCircle size={13} className="text-white/20 mt-0.5 shrink-0" />
          <p className="text-[11px] text-white/30 font-bold leading-relaxed">
            Only Admins can register, update, or remove tools. Members can only list and use the tools they already have access to.
          </p>
        </div>
      </section>

      <DocsPager />
    </div>
  );
}
