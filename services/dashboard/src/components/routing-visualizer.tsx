"use client"

import { useState } from "react";
import { CheckCircle2, Cpu, Database, ArrowRight, Terminal, Search, Zap, Loader2 } from "lucide-react";
import { executeMemzentPrompt } from "../app/actions";

export function RoutingVisualizer({ steps, orgId }: { steps?: any[], orgId?: string }) {
    const [prompt, setPrompt] = useState("");
    const [isExecuting, setIsExecuting] = useState(false);
    const [traceResult, setTraceResult] = useState<any>(null);

    const handleExecute = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!prompt.trim()) return;

        setIsExecuting(true);
        setTraceResult(null);

        try {
            const res = await executeMemzentPrompt(prompt, orgId);
            setTraceResult(res);
        } catch (err: any) {
            console.error(err);
            setTraceResult({ error: err.message });
        } finally {
            setIsExecuting(false);
        }
    };

    return (
        <div className="stat-card neural-bg border-white/5 p-6 relative overflow-hidden">
            <header className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8 border-b border-white/5 pb-6">
                <div className="flex items-center gap-2">
                    <Terminal size={14} className="text-memzent-glow" />
                    <h3 className="text-xs font-black text-white/60 uppercase tracking-widest text-center">Live Arena</h3>
                </div>

                <form onSubmit={handleExecute} className="w-full md:w-3/5 flex relative">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-white/20 w-4 h-4" />
                    <input
                        type="text"
                        value={prompt}
                        onChange={(e) => setPrompt(e.target.value)}
                        placeholder="Test intent routing (e.g., 'Fetch database metrics')"
                        className="w-full bg-black/40 border border-white/10 text-xs font-bold text-white px-10 py-3 rounded-xl focus:outline-none focus:border-memzent-glow transition-all placeholder:text-white/10"
                        disabled={isExecuting}
                    />
                    <button
                        type="submit"
                        disabled={isExecuting || !prompt.trim()}
                        className="absolute right-2 top-1/2 -translate-y-1/2 bg-white/5 hover:bg-white/10 text-white p-1.5 rounded-lg border border-white/5 transition-colors disabled:opacity-50"
                    >
                        {isExecuting ? <Loader2 className="w-4 h-4 animate-spin text-memzent-glow" /> : <ArrowRight className="w-4 h-4 text-white/40" />}
                    </button>
                </form>
            </header>

            {!traceResult && !isExecuting && (
                <div className="flex flex-col items-center justify-center py-8 opacity-80">
                    <Database size={24} className="mb-4 text-memzent-glow/20 animate-pulse" />
                    <p className="text-[10px] font-black uppercase tracking-[0.25em] text-memzent-glow/30 animate-pulse">Awaiting Prompt Evaluation...</p>
                </div>
            )}

            {isExecuting && (
                <div className="flex flex-col md:flex-row items-center justify-between gap-6 opacity-60 animate-pulse">
                    <TraceStep icon={<Database size={16} />} label="Secure Ingress" status="active" detail="Evaluating Intent" />
                    <ArrowRight className="text-slate-700 hidden md:block" size={16} />
                    <TraceStep icon={<Search size={16} />} label="Semantic Cache" status="pending" detail="Checking Neural Hash" />
                    <ArrowRight className="text-slate-700 hidden md:block" size={16} />
                    <TraceStep icon={<Cpu size={16} />} label="Intelligence Hub" status="pending" detail="Synthesis Engine" />
                </div>
            )}

            {traceResult && !traceResult.error && (
                <div className="space-y-6">
                    <div className="flex flex-col md:flex-row items-center justify-between gap-6">
                        <TraceStep icon={<Database size={16} />} label="Secure Ingress" status="complete" detail="Verified Logic 200" />
                        <ArrowRight className="text-slate-700 hidden md:block" size={16} />

                        {traceResult.cached ? (
                            <TraceStep icon={<Zap size={16} />} label="Semantic Match" status="active" detail="Context Memory HIT" color="glow-cyan" />
                        ) : (
                            <TraceStep icon={<Cpu size={16} />} label="Memzent Routing" status="complete" detail="Neural Context Mapping" />
                        )}

                        <ArrowRight className="text-slate-700 hidden md:block" size={16} />

                        {traceResult.cached ? (
                            <TraceStep icon={<CheckCircle2 size={16} />} label="Synthesized" status="complete" detail="Sub-1ms Delivery Return" />
                        ) : (
                            <TraceStep
                                icon={<CheckCircle2 size={16} />}
                                label={traceResult.tools && traceResult.tools.length > 0 ? "Tool Activated" : "Conversational Fallback"}
                                status="active"
                                color="glow-purple"
                                detail={traceResult.tools && traceResult.tools.length > 0 ? traceResult.tools[0].name : "Generic Inference"}
                            />
                        )}
                    </div>

                    <div className="mt-8 p-4 bg-slate-950 border border-slate-800 rounded-xl">
                        <div className="text-[10px] font-black uppercase text-slate-500 mb-2">Final Logic Synthesis:</div>
                        <p className="text-sm text-slate-300 font-serif leading-relaxed whitespace-pre-wrap">{traceResult.text}</p>
                    </div>
                </div>
            )}

            {traceResult && traceResult.error && (
                <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-center">
                    <p className="text-red-400 font-mono text-sm">{traceResult.error}</p>
                </div>
            )}

            {traceResult?.cached && <div className="absolute top-0 right-0 w-32 h-32 bg-memzent-glow/5 blur-[80px] rounded-full pointer-events-none" />}
            {traceResult && !traceResult.cached && <div className="absolute bottom-0 left-0 w-32 h-32 bg-memzent-purple/5 blur-[80px] rounded-full pointer-events-none" />}
        </div>
    );
}

function TraceStep({ icon, label, status, detail, color }: any) {
    const isComplete = status === 'complete';
    const isActive = status === 'active';

    return (
        <div className="flex flex-col items-center gap-3">
            <div className={`p-4 rounded-xl border ${color ? color :
                isComplete ? 'bg-blue-500/10 text-blue-400 border-blue-500/20' :
                    isActive ? 'bg-slate-800 text-white border-slate-700 shadow-lg' :
                        'bg-slate-950 text-slate-600 border-slate-900'
                }`}>
                {icon}
            </div>
            <div className="text-center">
                <p className={`text-sm font-medium ${isActive ? 'text-white' : 'text-slate-300'}`}>{label}</p>
                <p className="text-[10px] text-slate-500 font-mono uppercase tracking-tighter">{detail}</p>
            </div>
        </div>
    );
}

export function ToolGrid({ tools }: { tools: any[] }) {
    if (!tools || tools.length === 0) {
        return (
            <div className="p-12 text-center bg-slate-900/50 border border-slate-800 border-dashed rounded-2xl">
                <p className="text-slate-600 text-sm font-mono">No tools discovered yet...</p>
            </div>
        );
    }

    return (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {tools.map((tool: any) => (
                <div key={tool.id} className="p-6 bg-slate-900 border border-slate-800 rounded-2xl flex flex-col justify-between">
                    <div>
                        <div className="flex justify-between items-start mb-4">
                            <span className="text-[10px] bg-blue-500/10 text-blue-400 px-2 py-1 rounded-md uppercase font-bold border border-blue-500/10">
                                {tool.provider || 'Memzent'}
                            </span>
                            <div className="w-1.5 h-1.5 rounded-full bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.5)]" />
                        </div>
                        <h3 className="font-bold text-lg text-slate-100">{tool.name}</h3>
                        <p className="text-slate-500 text-[10px] font-mono mt-1 break-all">{tool.id}</p>
                    </div>
                </div>
            ))}
        </div>
    );
}