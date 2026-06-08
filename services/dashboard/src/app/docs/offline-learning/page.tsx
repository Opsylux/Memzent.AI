import { Activity, Cpu, BarChart3, Gauge, TrendingDown } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";
import { CodeBlock } from "@/components/docs/code-block";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Offline Learning Plane",
  description: "Memzent's asynchronous offline learning system: request miners, cache optimization, workflow discovery, and pattern analysis without blocking request flow.",
};

export default function OfflineLearningPage() {
  const eventStructure = `// OfflineEvent — emitted after every request
{
  "org_id":       "abc123",
  "prompt_hash":  "sha256:...",     // No raw prompt (PII-safe)
  "model":        "gpt-4o-mini",
  "cache_layer":  "L1b",           // Which layer resolved it
  "latency_ms":   45,
  "entity_count": 3,
  "entity_types": ["account", "amount"],
  "tool_ids":     ["tool-uuid-1"],
  "timestamp":    "2025-01-15T10:30:00Z"
}`;

  const featureFlags = `# Offline Learning Feature Flags
MEMZENT_OFFLINE_ENABLED=true       # Enable offline plane (default: true)
MEMZENT_OFFLINE_STREAMS=false      # Use Valkey Streams instead of channels
MEMZENT_PATTERN_MINING_ENABLED=false  # E6 Markov chain (experimental)`;

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-3 mb-4">
        <div className="p-2 rounded-xl bg-purple-500/10 border border-purple-500/20">
          <Activity size={20} className="text-purple-400" />
        </div>
        <h1 className="text-3xl font-black tracking-tight">Offline Learning Plane</h1>
      </div>
      <p className="text-white/50 text-sm leading-relaxed mb-10">
        The Offline Learning Plane processes request telemetry asynchronously — mining patterns,
        optimizing cache strategies, and discovering workflow opportunities without adding
        any latency to the request path.
      </p>

      {/* Architecture */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Cpu size={16} className="text-memzent-glow" />
          Architecture
        </h2>
        <div className="space-y-4">
          <div className="p-4 rounded-xl border border-white/10 bg-white/5">
            <h4 className="text-sm font-black text-white mb-2">Event Bus</h4>
            <p className="text-xs text-white/50">
              Uses a try-send buffered channel (4096 buffer, 4 worker goroutines).
              Events are never blocked — if the buffer is full, the event is dropped (counted as a
              <code> memzent_offline_drops_total</code> Prometheus metric). For distributed deployments,
              Valkey Streams transport is available via the <code>MEMZENT_OFFLINE_STREAMS</code> flag.
            </p>
          </div>
          <div className="p-4 rounded-xl border border-white/10 bg-white/5">
            <h4 className="text-sm font-black text-white mb-2">PII Safety</h4>
            <p className="text-xs text-white/50">
              Offline events contain <strong>no raw prompts</strong> — only prompt hashes (SHA-256),
              entity types, cache layers, and latency metrics. This ensures compliance with data
              privacy requirements while still enabling pattern analysis.
            </p>
          </div>
        </div>
      </section>

      {/* Event Structure */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <BarChart3 size={16} className="text-memzent-glow" />
          Event Structure
        </h2>
        <CodeBlock code={eventStructure} language="json" />
      </section>

      {/* Three Miners */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Gauge size={16} className="text-memzent-glow" />
          Mining Components
        </h2>
        <div className="space-y-4">
          <div className="p-4 rounded-xl border border-memzent-glow/20 bg-memzent-glow/5">
            <h4 className="text-sm font-black text-white mb-2">Request Miner</h4>
            <p className="text-xs text-white/50">
              Analyzes request patterns — most-used models, peak traffic times, entity type distributions.
              Feeds data to the cache pre-warmer for speculative cache population.
            </p>
          </div>
          <div className="p-4 rounded-xl border border-blue-500/20 bg-blue-500/5">
            <h4 className="text-sm font-black text-white mb-2">Cache Miner</h4>
            <p className="text-xs text-white/50">
              Tracks cache hit/miss ratios per layer, identifies under-performing cache entries,
              and suggests threshold adjustments. Powers the GPU Avoidance Rate metric.
            </p>
          </div>
          <div className="p-4 rounded-xl border border-purple-500/20 bg-purple-500/5">
            <h4 className="text-sm font-black text-white mb-2">Workflow Miner</h4>
            <p className="text-xs text-white/50">
              Detects recurring multi-step request sequences (e.g., &ldquo;lookup customer → check balance →
              generate invoice&rdquo;). Discovered workflows are registered in the Workflow Registry
              for single-shot execution optimization.
            </p>
          </div>
        </div>
      </section>

      {/* Speculative Pre-Warmer */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <TrendingDown size={16} className="text-memzent-glow" />
          Speculative Pre-Warmer
        </h2>
        <p className="text-white/50 text-sm mb-4">
          When the Request Miner identifies high-frequency entity combinations, the Pre-Warmer
          speculatively populates L1b cache entries. This means the first request for a common
          entity pattern may already be cached — achieving zero-latency on first hit.
        </p>
      </section>

      {/* Configuration */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Cpu size={16} className="text-memzent-glow" />
          Configuration
        </h2>
        <CodeBlock code={featureFlags} language="bash" />
      </section>

      <DocsPager />
    </div>
  );
}
