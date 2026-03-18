import { getAuraTools } from "./actions";
import { ToolGrid, RoutingVisualizer } from "@/components/routing-visualizer";

export default async function Page() {
  const initialTools = await getAuraTools();

  return (
    <main className="p-8 bg-slate-950 min-h-screen text-white">
      <div className="max-w-6xl mx-auto space-y-8">
        <header className="flex justify-between items-end border-b border-slate-900 pb-6">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">AURA_COMMAND</h1>
            <p className="text-slate-500 text-sm">Real-time status of your AI infrastructure</p>
          </div>
          <div className="text-right">
            <p className="text-xs font-mono text-blue-500 flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
              GATEWAY_CONNECTED
            </p>
          </div>
        </header>

        {/* Execution Trace */}
        <section className="space-y-4">
          <RoutingVisualizer steps={[]} />
        </section>

        {/* Tools Grid */}
        <section className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold text-slate-400 uppercase tracking-wider">Aura Tools</h2>
            <span className="text-xs text-slate-600 font-mono">{(initialTools || []).length} discovered</span>
          </div>
          <ToolGrid tools={initialTools || []} />
        </section>
      </div>
    </main>
  );
}