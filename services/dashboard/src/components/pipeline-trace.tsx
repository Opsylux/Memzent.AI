"use client";

import { CheckCircle2, Circle, Loader2, MinusCircle, XCircle } from "lucide-react";

export type PipelineStatus = "idle" | "running" | "cache_hit" | "llm_hit" | "error";

const STEPS = [
  { id: "rate", label: "Rate limit", sub: "Tier bucket" },
  { id: "l1", label: "Cache L1", sub: "Exact match" },
  { id: "l15", label: "Cache L1.5", sub: "Normalized" },
  { id: "rbac", label: "RBAC", sub: "org_tools" },
  { id: "route", label: "Router", sub: "Vector tools" },
  { id: "l2", label: "Cache L2", sub: "Semantic" },
  { id: "llm", label: "LLM", sub: "Synthesis" },
] as const;

type StepId = (typeof STEPS)[number]["id"];
type StepState = "pending" | "active" | "done" | "skipped" | "error";

function resolveStates(status: PipelineStatus, activeIndex: number): Record<StepId, StepState> {
  const base = Object.fromEntries(STEPS.map((s) => [s.id, "pending" as StepState])) as Record<StepId, StepState>;

  if (status === "idle") return base;

  if (status === "error") {
    STEPS.forEach((s, i) => {
      if (i < activeIndex) base[s.id] = "done";
      else if (i === activeIndex) base[s.id] = "error";
      else base[s.id] = "skipped";
    });
    return base;
  }

  if (status === "cache_hit") {
    base.rate = "done";
    base.l1 = "done";
    ["l15", "rbac", "route", "l2", "llm"].forEach((id) => {
      base[id as StepId] = "skipped";
    });
    return base;
  }

  if (status === "running") {
    STEPS.forEach((s, i) => {
      if (i < activeIndex) base[s.id] = "done";
      else if (i === activeIndex) base[s.id] = "active";
    });
    return base;
  }

  // llm_hit — full pipeline
  STEPS.forEach((s) => {
    base[s.id] = "done";
  });
  return base;
}

function StepIcon({ state }: { state: StepState }) {
  if (state === "active") return <Loader2 size={14} className="animate-spin text-memzent-glow" />;
  if (state === "done") return <CheckCircle2 size={14} className="text-memzent-accent" />;
  if (state === "skipped") return <MinusCircle size={14} className="text-white/25" />;
  if (state === "error") return <XCircle size={14} className="text-red-400" />;
  return <Circle size={14} className="text-white/20" />;
}

interface PipelineTraceProps {
  status: PipelineStatus;
  activeStep?: number;
  elapsedMs?: number;
  className?: string;
}

export function PipelineTrace({ status, activeStep = 2, elapsedMs, className = "" }: PipelineTraceProps) {
  const states = resolveStates(status, activeStep);

  return (
    <div className={`space-y-1 ${className}`}>
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-xs font-bold uppercase tracking-wider text-readable-label">
          Execution pipeline
        </h3>
        {elapsedMs != null && status !== "idle" && status !== "running" && (
          <span className="text-xs font-mono font-bold text-memzent-glow">{elapsedMs}ms</span>
        )}
      </div>

      <div className="relative">
        <div className="absolute left-[11px] top-3 bottom-3 w-px bg-white/10" />
        <ul className="space-y-3">
          {STEPS.map((step) => {
            const state = states[step.id];
            const isHighlight = state === "active" || (status === "cache_hit" && step.id === "l1" && state === "done");

            return (
              <li
                key={step.id}
                className={`relative flex items-center gap-3 pl-0 pr-2 py-1.5 rounded-lg transition-colors ${
                  isHighlight ? "bg-memzent-glow/5" : ""
                }`}
              >
                <div
                  className={`relative z-10 flex h-6 w-6 shrink-0 items-center justify-center rounded-full border ${
                    state === "done"
                      ? "border-memzent-accent/30 bg-memzent-accent/10"
                      : state === "active"
                        ? "border-memzent-glow/40 bg-memzent-glow/10"
                        : "border-white/10 bg-black/40"
                  }`}
                >
                  <StepIcon state={state} />
                </div>
                <div className="min-w-0 flex-1">
                  <div
                    className={`text-xs font-bold leading-none ${
                      state === "pending" || state === "skipped" ? "text-readable-muted" : "text-readable-primary"
                    }`}
                  >
                    {step.label}
                  </div>
                  <div className="text-[11px] text-readable-muted mt-0.5">{step.sub}</div>
                </div>
              </li>
            );
          })}
        </ul>
      </div>

      {status === "cache_hit" && (
        <p className="mt-4 text-[11px] font-medium text-memzent-glow/90 rounded-lg border border-memzent-glow/20 bg-memzent-glow/5 px-3 py-2">
          Short-circuited at L1 — no LLM call. ~80% billing discount applied.
        </p>
      )}
    </div>
  );
}
