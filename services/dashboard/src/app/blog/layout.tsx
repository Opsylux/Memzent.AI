import Link from "next/link";
import { Shield } from "lucide-react";

export default function BlogLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen bg-[#030507] flex flex-col">
      {/* Blog Header */}
      <header className="h-16 border-b border-white/5 sticky top-0 bg-[#030507]/80 backdrop-blur-md z-50 flex items-center justify-between px-6">
        <div className="flex items-center gap-8">
          <Link href="https://memzent.ai" className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-memzent-glow to-memzent-purple flex items-center justify-center shadow-[0_0_15px_rgba(0,243,255,0.3)]">
              <Shield size={18} className="text-black" strokeWidth={3} />
            </div>
            <span className="text-xl font-black tracking-tighter text-white">MEMZENT</span>
          </Link>
          <nav className="hidden md:flex items-center gap-6">
            <Link href="https://memzent.ai" className="text-xs font-bold text-white/40 uppercase tracking-widest hover:text-white transition-colors">Home</Link>
            <Link href="/docs" className="text-xs font-bold text-white/40 uppercase tracking-widest hover:text-white transition-colors">Documentation</Link>
            <Link href="/blog" className="text-xs font-bold text-white uppercase tracking-widest hover:text-memzent-glow transition-colors">Blog</Link>
            <Link href="https://github.com/Opsylux/Memzent.AI" target="_blank" className="text-xs font-bold text-white/40 uppercase tracking-widest hover:text-white transition-colors">GitHub</Link>
          </nav>
        </div>
        <div className="flex items-center gap-4">
          <Link href="/login" className="px-4 py-2 rounded-xl bg-white/5 border border-white/10 text-xs font-black uppercase tracking-widest hover:bg-white/10 transition-all">
            Dashboard
          </Link>
        </div>
      </header>

      {/* Content */}
      <main className="flex-1">
        {children}
      </main>

      {/* Footer */}
      <footer className="border-t border-white/5 py-8 px-6">
        <div className="max-w-6xl mx-auto flex items-center justify-between">
          <span className="text-[10px] font-bold text-white/20 uppercase tracking-widest">
            © 2026 Memzent. All rights reserved.
          </span>
          <div className="flex items-center gap-4">
            <Link href="/docs" className="text-[10px] font-bold text-white/20 uppercase tracking-widest hover:text-white/40 transition-colors">Docs</Link>
            <Link href="https://github.com/Opsylux/Memzent.AI" className="text-[10px] font-bold text-white/20 uppercase tracking-widest hover:text-white/40 transition-colors">GitHub</Link>
          </div>
        </div>
      </footer>
    </div>
  );
}
