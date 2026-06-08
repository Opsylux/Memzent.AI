import { Scan, Shield, Fingerprint, Tag, Zap, AlertTriangle } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";
import { CodeBlock } from "@/components/docs/code-block";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Entity Extraction",
  description: "How Memzent extracts typed entities from prompts to prevent false cache hits and ensure data isolation between accounts, users, and transactions.",
};

export default function EntityExtractionPage() {
  const entityResponse = `{
  "text": "I cannot facilitate transactions...",
  "cached": false,
  "entities": {
    "account_source": "123",
    "account_dest": "456",
    "amount": "100"
  }
}`;

  const reverseExample = `// Prompt A: "Transfer $100 from account 123 to account 456"
// Prompt B: "Transfer $100 from account 456 to account 123"
//
// Without entity extraction → Cache HIT (wrong! different direction)
// With entity extraction   → Cache MISS (correct! entities differ)`;

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-3 mb-4">
        <div className="p-2 rounded-xl bg-memzent-glow/10 border border-memzent-glow/20">
          <Scan size={20} className="text-memzent-glow" />
        </div>
        <h1 className="text-3xl font-black tracking-tight">Entity Extraction</h1>
      </div>
      <p className="text-white/50 text-sm leading-relaxed mb-10">
        Memzent&apos;s entity extraction layer identifies typed entities — accounts, customers, amounts,
        dates, and identifiers — from every prompt. This prevents the most dangerous class of cache
        errors: returning cached data meant for a different entity.
      </p>

      {/* Why Entity Extraction */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <AlertTriangle size={16} className="text-yellow-400" />
          The Problem It Solves
        </h2>
        <p className="text-white/50 text-sm mb-4">
          Semantic similarity alone cannot distinguish between prompts that are structurally identical
          but reference different data. Consider:
        </p>
        <CodeBlock code={reverseExample} language="javascript" />
        <p className="text-white/50 text-sm mt-4">
          Both prompts have &gt;0.98 semantic similarity. Without entity-aware guards,
          the cache would incorrectly return account 123→456 data for a 456→123 request.
        </p>
      </section>

      {/* Six Entity Types */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Tag size={16} className="text-memzent-glow" />
          Supported Entity Types
        </h2>
        <div className="space-y-3">
          {[
            { type: "Account", pattern: "account #123, acct 456", color: "memzent-glow", desc: "Bank/financial account identifiers with directional awareness (source vs destination)" },
            { type: "Customer", pattern: "customer Raj, client #101", color: "blue-400", desc: "Named or numbered customer/client references" },
            { type: "Invoice", pattern: "invoice #789, inv 1234", color: "purple-400", desc: "Invoice and order identifiers" },
            { type: "Amount", pattern: "$100, 50.00, 1000", color: "green-400", desc: "Monetary amounts and numeric quantities" },
            { type: "Date", pattern: "2024-01-15, Jan 15th", color: "yellow-400", desc: "Date references in various formats" },
            { type: "Identifier", pattern: "ID-12345, ref:ABC", color: "red-400", desc: "Generic identifiers, reference numbers, and codes" },
          ].map((entity) => (
            <div key={entity.type} className="p-4 rounded-xl border border-white/10 bg-white/5">
              <div className="flex items-center gap-2 mb-1">
                <span className="text-xs font-black px-2 py-0.5 rounded bg-white/10 text-white/80">
                  {entity.type}
                </span>
                <code className="text-[10px] text-white/30">{entity.pattern}</code>
              </div>
              <p className="text-xs text-white/50">{entity.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* How It Works */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Zap size={16} className="text-memzent-glow" />
          How It Works
        </h2>
        <div className="space-y-4">
          <div className="p-4 rounded-xl border border-white/10 bg-white/5">
            <h4 className="text-sm font-black text-white mb-2">1. Regex-Only Extraction (&lt;1ms)</h4>
            <p className="text-xs text-white/50">
              Entity extraction is entirely regex-based — no SLM or LLM call. Runs in both the
              Rust Router (for cache guard comparison) and Go Gateway (for L1b key generation).
              Typical extraction time: 0.1–0.5ms.
            </p>
          </div>
          <div className="p-4 rounded-xl border border-white/10 bg-white/5">
            <h4 className="text-sm font-black text-white mb-2">2. Positional Awareness</h4>
            <p className="text-xs text-white/50">
              Unlike simple number extraction, entities preserve their role in the prompt.
              &ldquo;from account 123&rdquo; and &ldquo;to account 123&rdquo; produce different entity keys
              (<code>account_source=123</code> vs <code>account_dest=123</code>).
            </p>
          </div>
          <div className="p-4 rounded-xl border border-white/10 bg-white/5">
            <h4 className="text-sm font-black text-white mb-2">3. Post-Filter Guard</h4>
            <p className="text-xs text-white/50">
              Entity comparison only triggers <strong>after</strong> a semantic cache hit (similarity &gt;0.95).
              If the entities don&apos;t match, the hit is rejected and the request goes to the LLM.
              This is a lazy post-filter — it adds zero cost to cache misses.
            </p>
          </div>
        </div>
      </section>

      {/* API Response */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Fingerprint size={16} className="text-memzent-glow" />
          Entities in API Responses
        </h2>
        <p className="text-white/50 text-sm mb-4">
          When entities are extracted, they appear in the response body:
        </p>
        <CodeBlock code={entityResponse} language="json" />
      </section>

      {/* GPU Avoidance */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Shield size={16} className="text-memzent-glow" />
          GPU Avoidance Metric
        </h2>
        <p className="text-white/50 text-sm">
          Entity extraction enables the <strong>GPU Avoidance Rate</strong> — the percentage of requests
          resolved without hitting the LLM. By combining regex entity extraction with the L1b cache layer,
          Memzent avoids expensive GPU inference for entity-identical repeat requests.
          Track this metric in the <strong>GPU Analytics</strong> dashboard.
        </p>
      </section>

      <DocsPager />
    </div>
  );
}
