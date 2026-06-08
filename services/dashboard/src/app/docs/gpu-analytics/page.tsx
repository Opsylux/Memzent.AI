import { BarChart3, Cpu, TrendingUp, Zap, Shield, Activity } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";
import { CodeBlock } from "@/components/docs/code-block";
import { DOCS_CONFIG } from "@/config/docs-config";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "GPU Analytics & Avoidance",
  description: "Track GPU avoidance rate, entity extraction quality, and cache layer distribution with Memzent's GPU analytics dashboard and Prometheus metrics.",
};

export default function GpuAnalyticsPage() {
  const prometheusMetrics = `# Entity extraction counters
memzent_entity_extraction_total{type="account"}    245
memzent_entity_extraction_total{type="customer"}   189
memzent_entity_extraction_total{type="amount"}     312
memzent_entity_extraction_total{type="date"}        67

# Cache layer distribution
memzent_cache_layer_total{layer="L1"}              1024
memzent_cache_layer_total{layer="L1.5"}             456
memzent_cache_layer_total{layer="L1b"}              892
memzent_cache_layer_total{layer="L2"}               234
memzent_cache_layer_total{layer="MISS"}             567

# GPU avoidance
memzent_gpu_avoided_total                          2606
memzent_gpu_required_total                          567

# GPU Avoidance Rate = avoided / (avoided + required) = 82.1%`;

  const gpuEndpoint = `curl -X GET https://${DOCS_CONFIG.domain}/v1/stats \\
  -H "X-API-Key: memzent_YOUR_KEY"

# Response includes:
{
  "cache_stats": {
    "total_requests": 3173,
    "cache_hits": 2606,
    "cache_misses": 567,
    "hit_rate": 0.821,
    "layer_distribution": {
      "L1": 1024, "L1.5": 456, "L1b": 892, "L2": 234
    }
  },
  "entity_stats": {
    "total_extractions": 813,
    "by_type": { "account": 245, "customer": 189 }
  }
}`;

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-3 mb-4">
        <div className="p-2 rounded-xl bg-green-500/10 border border-green-500/20">
          <BarChart3 size={20} className="text-green-400" />
        </div>
        <h1 className="text-3xl font-black tracking-tight">GPU Analytics &amp; Avoidance</h1>
      </div>
      <p className="text-white/50 text-sm leading-relaxed mb-10">
        The GPU Avoidance Rate is Memzent&apos;s primary business metric — it measures how many
        requests are resolved <strong>without</strong> hitting the LLM. Higher avoidance = lower cost,
        lower latency, and better user experience.
      </p>

      {/* GPU Avoidance Rate */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <TrendingUp size={16} className="text-green-400" />
          What is GPU Avoidance Rate?
        </h2>
        <div className="p-5 rounded-xl border border-green-500/20 bg-green-500/5">
          <div className="text-center mb-4">
            <div className="text-3xl font-black text-green-400">GPU Avoidance Rate</div>
            <div className="text-sm text-white/50 mt-1">
              <code>cache_hits / total_requests × 100%</code>
            </div>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 text-center">
            <div className="p-3 rounded-lg bg-white/5">
              <div className="text-xs font-black text-white/40">Good</div>
              <div className="text-lg font-black text-yellow-400">60-75%</div>
            </div>
            <div className="p-3 rounded-lg bg-white/5">
              <div className="text-xs font-black text-white/40">Great</div>
              <div className="text-lg font-black text-green-400">75-90%</div>
            </div>
            <div className="p-3 rounded-lg bg-white/5">
              <div className="text-xs font-black text-white/40">Exceptional</div>
              <div className="text-lg font-black text-memzent-glow">90%+</div>
            </div>
          </div>
        </div>
      </section>

      {/* What Contributes */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Zap size={16} className="text-memzent-glow" />
          What Drives GPU Avoidance?
        </h2>
        <div className="space-y-3">
          <div className="p-4 rounded-xl border border-white/10 bg-white/5">
            <div className="flex items-center gap-2 mb-1">
              <Shield size={14} className="text-memzent-glow" />
              <h4 className="text-sm font-black text-white">Entity-Aware Caching (L1b)</h4>
            </div>
            <p className="text-xs text-white/50">
              The biggest contributor. Repeat requests with identical entities bypass all vector math
              and LLM calls. Typical improvement: 20-30% increase in avoidance rate.
            </p>
          </div>
          <div className="p-4 rounded-xl border border-white/10 bg-white/5">
            <div className="flex items-center gap-2 mb-1">
              <Activity size={14} className="text-blue-400" />
              <h4 className="text-sm font-black text-white">Semantic Cache (L2)</h4>
            </div>
            <p className="text-xs text-white/50">
              Catches paraphrased questions with entity post-filter. Contributes 10-15% to
              avoidance rate for diverse user bases.
            </p>
          </div>
          <div className="p-4 rounded-xl border border-white/10 bg-white/5">
            <div className="flex items-center gap-2 mb-1">
              <Cpu size={14} className="text-purple-400" />
              <h4 className="text-sm font-black text-white">Workflow Shortcuts</h4>
            </div>
            <p className="text-xs text-white/50">
              Pre-approved multi-step workflows skip individual routing. Contributes 5-10% for
              orgs with active workflow registries.
            </p>
          </div>
        </div>
      </section>

      {/* Prometheus Metrics */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <BarChart3 size={16} className="text-memzent-glow" />
          Prometheus Metrics
        </h2>
        <p className="text-white/50 text-sm mb-4">
          All metrics are exposed on <code>/metrics</code> in Prometheus format:
        </p>
        <CodeBlock code={prometheusMetrics} language="text" />
      </section>

      {/* API Endpoint */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Cpu size={16} className="text-memzent-glow" />
          Stats API
        </h2>
        <CodeBlock code={gpuEndpoint} language="bash" />
      </section>

      {/* Dashboard */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <BarChart3 size={16} className="text-memzent-glow" />
          Dashboard
        </h2>
        <p className="text-white/50 text-sm">
          The GPU Analytics dashboard (available in the Command Center under Analytics → GPU)
          shows real-time avoidance rate, cache layer distribution charts, entity extraction
          breakdowns, and historical trends. Admin and editor roles can view the full dashboard;
          viewers see a summary card on the overview page.
        </p>
      </section>

      <DocsPager currentPath="/docs/gpu-analytics" />
    </div>
  );
}
