import Link from "next/link";
import { Shield, ChevronRight, Menu } from "lucide-react";
import { DocsSidebar } from "@/components/docs/docs-sidebar";

export default function DocsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen bg-memzent-dark flex flex-col">
      {/* Public Docs Header */}
      <header className="h-16 border-b border-white/5 sticky top-0 bg-memzent-dark/80 backdrop-blur-md z-50 flex items-center justify-between px-6">
        <div className="flex items-center gap-8">
          <Link href="https://memzent.ai" className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-memzent-glow to-memzent-purple flex items-center justify-center shadow-[0_0_15px_rgba(0,243,255,0.3)]">
              <Shield size={18} className="text-black" strokeWidth={3} />
            </div>
            <span className="text-xl font-black tracking-tighter text-white">MEMZENT</span>
          </Link>
          <nav className="hidden md:flex items-center gap-6">
            <Link href="https://memzent.ai" className="text-xs font-bold text-white/40 uppercase tracking-widest hover:text-white transition-colors">Home</Link>
            <Link href="/docs" className="text-xs font-bold text-white uppercase tracking-widest hover:text-memzent-glow transition-colors">Documentation</Link>
            <Link href="https://github.com/Opsylux/MemzentMCP" target="_blank" className="text-xs font-bold text-white/40 uppercase tracking-widest hover:text-white transition-colors">GitHub</Link>
          </nav>
        </div>
        <div className="flex items-center gap-4">
          <Link href="/login" className="px-4 py-2 rounded-xl bg-white/5 border border-white/10 text-xs font-black uppercase tracking-widest hover:bg-white/10 transition-all">
            Dashboard
          </Link>
        </div>
      </header>

      <div className="flex-1 flex max-w-7xl mx-auto w-full relative">
        {/* Desktop Sidebar */}
        <aside className="hidden lg:block w-64 h-[calc(100vh-4rem)] sticky top-16 overflow-y-auto border-r border-white/5 py-10 pr-6">
          <DocsSidebar />
        </aside>

        {/* Content */}
        <main className="flex-1 px-6 lg:px-12 py-10 min-w-0">
          <div className="max-w-3xl prose prose-invert prose-memzent">
            {children}
          </div>

          {/* Footer inside content area */}
          <footer className="mt-20 pt-10 border-t border-white/5 flex flex-col md:flex-row justify-between gap-6 pb-20">
            <div>
              <div className="text-[10px] font-black uppercase tracking-[0.2em] text-white/20 mb-2 italic">Memzent_Neural_Mesh</div>
              <p className="text-xs text-white/40 font-medium max-w-xs leading-relaxed">
                The intelligent semantic proxy for agentic infrastructure. Optimize latency and ROI with every request.
              </p>
            </div>
            <div className="flex gap-12">
              <div className="flex flex-col gap-3">
                <span className="text-[10px] font-black uppercase tracking-widest text-white/20">Resources</span>
                <Link href="#" className="text-xs font-bold text-white/40 hover:text-memzent-glow transition-colors">API Keys</Link>
                <Link href="#" className="text-xs font-bold text-white/40 hover:text-memzent-glow transition-colors">RBAC Specs</Link>
              </div>
              <div className="flex flex-col gap-3">
                <span className="text-[10px] font-black uppercase tracking-widest text-white/20">Enterprise</span>
                <Link href="#" className="text-xs font-bold text-white/40 hover:text-memzent-glow transition-colors">SLA</Link>
                <Link href="#" className="text-xs font-bold text-white/40 hover:text-memzent-glow transition-colors">Security</Link>
              </div>
            </div>
          </footer>
        </main>
      </div>

      <div className="fixed bottom-0 right-0 w-[500px] h-[500px] bg-memzent-glow/5 blur-[150px] -z-10 rounded-full pointer-events-none" />
    </div>
  );
}
