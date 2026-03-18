import { CheckCircle2, Cpu, Database, ArrowRight, Terminal } from "lucide-react";

export function RoutingVisualizer({ steps }: { steps?: any[] }) {
    return (
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6">
            <header className="flex items-center gap-2 mb-6 border-b border-slate-800 pb-4">
                <Terminal size={14} className="text-slate-500" />
                <h3 className="text-xs font-bold text-slate-400 uppercase tracking-widest text-center">Execution Trace</h3>
            </header>

            <div className="flex flex-col md:flex-row items-center justify-between gap-6">
                <TraceStep icon={<Cpu size={16} />} label="Embedding" status="complete" detail="ModernBERT-v3" />
                <ArrowRight className="text-slate-700 hidden md:block" size={16} />
                <TraceStep icon={<Database size={16} />} label="Router" status="complete" detail="Rust-Router (98%)" />
                <ArrowRight className="text-slate-700 hidden md:block" size={16} />
                <TraceStep icon={<CheckCircle2 size={16} />} label="Tool" status="active" detail="Postgres-DBA" />
            </div>
        </div>
    );
}

function TraceStep({ icon, label, status, detail }: any) {
    const isComplete = status === 'complete';
    const isActive = status === 'active';

    return (
        <div className="flex flex-col items-center gap-3">
            <div className={`p-4 rounded-xl border ${isComplete ? 'bg-blue-500/10 text-blue-400 border-blue-500/20' :
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
                                {tool.provider || 'Aura'}
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