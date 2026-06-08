import { DollarSign, Shield, BarChart3, AlertTriangle, Clock, Gauge } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";
import { CodeBlock } from "@/components/docs/code-block";
import { DOCS_CONFIG } from "@/config/docs-config";

export default function SpendLimitsPage() {
  const budgetExample = `curl -X GET https://${DOCS_CONFIG.domain}/v1/billing/budget \\
  -H "X-API-Key: memzent_f7c9...8e2a"`;

  const budgetResponse = `{
  "org_id": "5127e445-bb64-4057-ac66-0f86fb68284c",
  "current_balance": 842.50,
  "tier": "pro",
  "spend_summaries": [
    { "period": "24h", "total_spend": 12.34, "request_count": 187, "cache_hits": 42, "cache_savings": 3.20 },
    { "period": "7d",  "total_spend": 74.80, "request_count": 1340, "cache_hits": 312, "cache_savings": 18.90 },
    { "period": "30d", "total_spend": 298.60, "request_count": 5120, "cache_hits": 1280, "cache_savings": 78.40 }
  ],
  "provider_breakdown": [
    { "provider": "openai",    "total_spend": 180.20, "request_count": 3100 },
    { "provider": "anthropic", "total_spend": 92.40,  "request_count": 1420 },
    { "provider": "ollama",    "total_spend": 26.00,  "request_count": 600 }
  ],
  "daily_avg_spend": 10.69,
  "projected_days_remaining": 78.8,
  "burn_rate_per_hour": 0.445
}`;

  const setLimitsExample = `curl -X PUT https://${DOCS_CONFIG.domain}/v1/billing/spend-limits \\
  -H "X-API-Key: memzent_f7c9...8e2a" \\
  -H "Content-Type: application/json" \\
  -d '{
    "daily_limit": 50.00,
    "monthly_limit": 1000.00,
    "daily_token_limit": 500000,
    "monthly_token_limit": 10000000
  }'`;

  const getLimitsExample = `curl -X GET https://${DOCS_CONFIG.domain}/v1/billing/spend-limits \\
  -H "X-API-Key: memzent_f7c9...8e2a"`;

  const limitsResponse = `{
  "daily_spend": 12.34,
  "monthly_spend": 298.60,
  "daily_limit": 50.00,
  "monthly_limit": 1000.00,
  "daily_exceeded": false,
  "monthly_exceeded": false,
  "daily_tokens_used": 124500,
  "monthly_tokens_used": 4820000,
  "daily_token_limit": 500000,
  "monthly_token_limit": 10000000,
  "daily_tokens_exceeded": false,
  "monthly_tokens_exceeded": false
}`;

  const timeseriesExample = `curl -X GET "https://${DOCS_CONFIG.domain}/v1/billing/spend-timeseries?days=14" \\
  -H "X-API-Key: memzent_f7c9...8e2a"`;

  const timeseriesResponse = `[
  { "day": "2026-05-24", "spend": 9.82, "requests": 164 },
  { "day": "2026-05-25", "spend": 11.40, "requests": 192 },
  { "day": "2026-05-26", "spend": 8.73, "requests": 141 },
  ...
]`;

  return (
    <div className="space-y-12">
      <header className="space-y-4">
        <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-memzent-glow/5 border border-memzent-glow/20 w-fit">
          <span className="text-[10px] font-black text-memzent-glow uppercase tracking-tighter italic">Billing</span>
        </div>
        <h1 className="text-4xl font-black tracking-tighter uppercase sm:text-5xl">Spend Limits & Budget Forecast</h1>
        <p className="text-lg text-white/60 leading-relaxed font-medium">
          Protect your organization from runaway token bills with configurable spend caps (dollars and tokens),
          real-time enforcement, and a budget forecast API designed for planning tools and FinOps integrations.
        </p>
      </header>

      {/* Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
        <div className="p-5 rounded-2xl bg-white/[0.02] border border-white/5 space-y-3">
          <div className="flex items-center gap-2 text-memzent-glow">
            <Shield size={16} />
            <span className="text-xs font-black uppercase">Hard Caps</span>
          </div>
          <p className="text-[11px] text-white/40 font-bold leading-relaxed">
            Set daily and monthly limits in both dollars and tokens. The engine blocks requests the moment a cap is exceeded.
          </p>
        </div>
        <div className="p-5 rounded-2xl bg-white/[0.02] border border-white/5 space-y-3">
          <div className="flex items-center gap-2 text-memzent-purple">
            <BarChart3 size={16} />
            <span className="text-xs font-black uppercase">Budget Forecast</span>
          </div>
          <p className="text-[11px] text-white/40 font-bold leading-relaxed">
            GET a single JSON payload with burn rate, projected days remaining, and provider-level breakdown for your planning tools.
          </p>
        </div>
        <div className="p-5 rounded-2xl bg-white/[0.02] border border-white/5 space-y-3">
          <div className="flex items-center gap-2 text-amber-400">
            <AlertTriangle size={16} />
            <span className="text-xs font-black uppercase">Webhook Alerts</span>
          </div>
          <p className="text-[11px] text-white/40 font-bold leading-relaxed">
            Receive a <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">spend_limit</code> webhook event when any cap is hit — route it to Slack, PagerDuty, or your FinOps pipeline.
          </p>
        </div>
      </div>

      {/* How it Works */}
      <section className="space-y-5">
        <h2 className="text-2xl font-black tracking-tighter uppercase">How Enforcement Works</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Every request through the Memzent engine runs a billing pre-check <strong className="text-white/80">before</strong> hitting any LLM provider.
          The check evaluates four caps in order — if any is exceeded, the request is rejected with a clear error message:
        </p>

        <div className="space-y-3 pl-4 border-l-2 border-memzent-glow/20">
          <div className="space-y-1">
            <div className="text-xs font-black text-white/80 uppercase flex items-center gap-2">
              <DollarSign size={14} className="text-memzent-glow" /> 1. Token Balance
            </div>
            <p className="text-[11px] text-white/40 font-bold">Account balance must be greater than zero.</p>
          </div>
          <div className="space-y-1">
            <div className="text-xs font-black text-white/80 uppercase flex items-center gap-2">
              <Clock size={14} className="text-memzent-glow" /> 2. Daily Dollar Cap
            </div>
            <p className="text-[11px] text-white/40 font-bold">Sum of today&apos;s spend must be under <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">daily_spend_limit</code>.</p>
          </div>
          <div className="space-y-1">
            <div className="text-xs font-black text-white/80 uppercase flex items-center gap-2">
              <Clock size={14} className="text-memzent-glow" /> 3. Monthly Dollar Cap
            </div>
            <p className="text-[11px] text-white/40 font-bold">Sum of this month&apos;s spend must be under <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">monthly_spend_limit</code>.</p>
          </div>
          <div className="space-y-1">
            <div className="text-xs font-black text-white/80 uppercase flex items-center gap-2">
              <Gauge size={14} className="text-memzent-glow" /> 4. Token Caps
            </div>
            <p className="text-[11px] text-white/40 font-bold">Daily and monthly token consumption must be under <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">daily_token_limit</code> / <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">monthly_token_limit</code>.</p>
          </div>
        </div>

        <div className="p-4 rounded-xl bg-amber-500/5 border border-amber-500/20 text-xs text-amber-300/80 font-bold">
          <strong className="text-amber-300">Note:</strong> All limits are opt-in. If no limits are configured, the engine only checks token balance.
          Limits reset automatically — daily at midnight UTC, monthly on the 1st.
        </div>
      </section>

      {/* Budget Status API */}
      <section className="space-y-5">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Budget Status API</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Retrieve a comprehensive budget report with burn rate, projections, and provider-level spend breakdown.
          Ideal for pulling into planning tools, FinOps dashboards, or CI/CD budget gates.
        </p>

        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <span className="px-2 py-0.5 rounded bg-green-500/10 text-green-400 text-[10px] font-black">GET</span>
            <code className="text-xs font-mono text-white/80">/v1/billing/budget</code>
          </div>
          <p className="text-[11px] text-white/40 font-bold">Requires <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">audit:read</code> scope.</p>
        </div>

        <CodeBlock code={budgetExample} language="bash" title="Request" />
        <CodeBlock code={budgetResponse} language="json" title="Response" />

        <div className="overflow-x-auto">
          <table className="w-full text-xs">
            <thead>
              <tr className="border-b border-white/10">
                <th className="text-left py-2 pr-4 font-black text-white/60 uppercase text-[10px]">Field</th>
                <th className="text-left py-2 pr-4 font-black text-white/60 uppercase text-[10px]">Type</th>
                <th className="text-left py-2 font-black text-white/60 uppercase text-[10px]">Description</th>
              </tr>
            </thead>
            <tbody className="text-white/40 font-bold">
              <tr className="border-b border-white/5"><td className="py-2 pr-4 font-mono text-memzent-glow">current_balance</td><td className="py-2 pr-4">float</td><td className="py-2">Available token credits in dollars</td></tr>
              <tr className="border-b border-white/5"><td className="py-2 pr-4 font-mono text-memzent-glow">burn_rate_per_hour</td><td className="py-2 pr-4">float</td><td className="py-2">Current dollar spend rate per hour (7-day average)</td></tr>
              <tr className="border-b border-white/5"><td className="py-2 pr-4 font-mono text-memzent-glow">projected_days_remaining</td><td className="py-2 pr-4">float</td><td className="py-2">Estimated days until balance hits zero</td></tr>
              <tr className="border-b border-white/5"><td className="py-2 pr-4 font-mono text-memzent-glow">daily_avg_spend</td><td className="py-2 pr-4">float</td><td className="py-2">Average daily spend over last 7 days</td></tr>
              <tr className="border-b border-white/5"><td className="py-2 pr-4 font-mono text-memzent-glow">spend_summaries[]</td><td className="py-2 pr-4">array</td><td className="py-2">Aggregated spend for 24h, 7d, 30d windows</td></tr>
              <tr className="border-b border-white/5"><td className="py-2 pr-4 font-mono text-memzent-glow">provider_breakdown[]</td><td className="py-2 pr-4">array</td><td className="py-2">Per-provider spend + request counts (30d)</td></tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* Spend Limits API */}
      <section className="space-y-5">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Configuring Spend Limits</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Set and retrieve daily/monthly caps for both dollars and tokens. Use <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">null</code> to
          remove a specific limit while keeping others active.
        </p>

        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <span className="px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 text-[10px] font-black">PUT</span>
            <code className="text-xs font-mono text-white/80">/v1/billing/spend-limits</code>
          </div>
          <p className="text-[11px] text-white/40 font-bold">Requires <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">audit:read</code> scope.</p>
        </div>

        <CodeBlock code={setLimitsExample} language="bash" title="Set Limits" />

        <div className="overflow-x-auto">
          <table className="w-full text-xs">
            <thead>
              <tr className="border-b border-white/10">
                <th className="text-left py-2 pr-4 font-black text-white/60 uppercase text-[10px]">Field</th>
                <th className="text-left py-2 pr-4 font-black text-white/60 uppercase text-[10px]">Type</th>
                <th className="text-left py-2 font-black text-white/60 uppercase text-[10px]">Description</th>
              </tr>
            </thead>
            <tbody className="text-white/40 font-bold">
              <tr className="border-b border-white/5"><td className="py-2 pr-4 font-mono text-memzent-glow">daily_limit</td><td className="py-2 pr-4">float | null</td><td className="py-2">Max dollars per day (midnight UTC reset)</td></tr>
              <tr className="border-b border-white/5"><td className="py-2 pr-4 font-mono text-memzent-glow">monthly_limit</td><td className="py-2 pr-4">float | null</td><td className="py-2">Max dollars per month (1st-of-month reset)</td></tr>
              <tr className="border-b border-white/5"><td className="py-2 pr-4 font-mono text-memzent-glow">daily_token_limit</td><td className="py-2 pr-4">int | null</td><td className="py-2">Max tokens (input + output) per day</td></tr>
              <tr className="border-b border-white/5"><td className="py-2 pr-4 font-mono text-memzent-glow">monthly_token_limit</td><td className="py-2 pr-4">int | null</td><td className="py-2">Max tokens (input + output) per month</td></tr>
            </tbody>
          </table>
        </div>

        <div className="space-y-2 mt-6">
          <div className="flex items-center gap-2">
            <span className="px-2 py-0.5 rounded bg-green-500/10 text-green-400 text-[10px] font-black">GET</span>
            <code className="text-xs font-mono text-white/80">/v1/billing/spend-limits</code>
          </div>
        </div>

        <CodeBlock code={getLimitsExample} language="bash" title="Get Current Status" />
        <CodeBlock code={limitsResponse} language="json" title="Response" />
      </section>

      {/* Timeseries API */}
      <section className="space-y-5">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Spend Timeseries</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Retrieve daily spend data for charting or trend analysis. Use the <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">days</code> query
          parameter to control the window (default: 30).
        </p>

        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <span className="px-2 py-0.5 rounded bg-green-500/10 text-green-400 text-[10px] font-black">GET</span>
            <code className="text-xs font-mono text-white/80">/v1/billing/spend-timeseries?days=14</code>
          </div>
        </div>

        <CodeBlock code={timeseriesExample} language="bash" title="Request" />
        <CodeBlock code={timeseriesResponse} language="json" title="Response" />
      </section>

      {/* Error Responses */}
      <section className="space-y-5">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Error Responses</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          When a spend limit is exceeded, the gateway returns <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">HTTP 402</code> with a descriptive message:
        </p>

        <div className="space-y-3">
          <div className="p-4 rounded-xl bg-red-500/5 border border-red-500/20 font-mono text-xs text-red-300/80 space-y-2">
            <div><strong className="text-red-300">Dollar cap:</strong> &quot;daily spend limit reached ($50.12 of $50.00). Resets at midnight UTC&quot;</div>
            <div><strong className="text-red-300">Token cap:</strong> &quot;daily token limit reached (502,340 of 500,000). Resets at midnight UTC&quot;</div>
            <div><strong className="text-red-300">Balance:</strong> &quot;payment required: token balance depleted&quot;</div>
          </div>
        </div>
      </section>

      {/* Integration Examples */}
      <section className="space-y-5">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Integration Examples</h2>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
          <div className="p-5 rounded-2xl bg-white/[0.02] border border-white/5 space-y-3">
            <span className="text-xs font-black uppercase text-white/80">CI/CD Budget Gate</span>
            <p className="text-[11px] text-white/40 font-bold leading-relaxed">
              Call <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">GET /v1/billing/budget</code> in
              your pipeline. If <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">projected_days_remaining &lt; 7</code>,
              fail the build or send an alert to the team channel.
            </p>
          </div>
          <div className="p-5 rounded-2xl bg-white/[0.02] border border-white/5 space-y-3">
            <span className="text-xs font-black uppercase text-white/80">FinOps Dashboard</span>
            <p className="text-[11px] text-white/40 font-bold leading-relaxed">
              Poll <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">GET /v1/billing/spend-timeseries?days=90</code> and
              pipe into Grafana, Datadog, or your own charting — each data point includes spend and request count per day.
            </p>
          </div>
          <div className="p-5 rounded-2xl bg-white/[0.02] border border-white/5 space-y-3">
            <span className="text-xs font-black uppercase text-white/80">Slack Alerts</span>
            <p className="text-[11px] text-white/40 font-bold leading-relaxed">
              Subscribe a webhook to the <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">spend_limit</code> event type.
              The payload includes which limit was hit and current usage — route it to Slack via an incoming webhook.
            </p>
          </div>
          <div className="p-5 rounded-2xl bg-white/[0.02] border border-white/5 space-y-3">
            <span className="text-xs font-black uppercase text-white/80">Terraform / IaC</span>
            <p className="text-[11px] text-white/40 font-bold leading-relaxed">
              Use <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">PUT /v1/billing/spend-limits</code> in
              provisioning scripts to enforce org-level guardrails at infrastructure deploy time.
            </p>
          </div>
        </div>
      </section>

      <DocsPager />
    </div>
  );
}
