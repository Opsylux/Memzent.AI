"use client"

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  Database,
  ShieldCheck,
  Settings,
  Activity,
  Zap,
  Globe
} from "lucide-react";
import { motion } from "framer-motion";

const NAV_ITEMS = [
  { icon: LayoutDashboard, label: "Intelligence Hub", href: "/" },
  { icon: Database, label: "Tool Registry", href: "/tools" },
  { icon: ShieldCheck, label: "Security Guardrails", href: "/security" },
  { icon: Activity, label: "System Health", href: "/health" },
  { icon: Settings, label: "Node Settings", href: "/settings" },
];

export function MemzentSidebar() {
  const pathname = usePathname();

  return (
    <aside className="fixed left-0 top-0 bottom-0 w-72 glass border-r border-white/5 z-50 flex flex-col p-6 m-4 rounded-3xl neural-bg">
      <div className="flex items-center gap-3 mb-12 px-2">
        <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-memzent-glow to-memzent-purple flex items-center justify-center p-2 shadow-[0_0_20px_rgba(0,243,255,0.3)]">
          <Zap className="text-black fill-black" size={20} />
        </div>
        <div>
          <h1 className="font-black text-xl tracking-tighter leading-none">MEMZENT_OS</h1>
          <p className="text-[10px] font-bold text-memzent-glow uppercase tracking-widest mt-1 opacity-60 italic">Enterprise v1.0</p>
        </div>
      </div>

      <nav className="flex-1 space-y-2">
        <div className="text-[10px] font-black text-white/40 uppercase tracking-widest mb-4 px-2">Core Systems</div>
        {NAV_ITEMS.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link key={item.href} href={item.href}>
              <div className={`relative flex items-center gap-3 px-4 py-3 rounded-2xl transition-all group ${isActive ? "bg-white/5 text-memzent-glow" : "text-white/60 hover:text-white hover:bg-white/5"
                }`}>
                {isActive && (
                  <motion.div
                    layoutId="active-nav"
                    className="absolute inset-0 bg-memzent-glow/5 border border-memzent-glow/20 rounded-2xl"
                    transition={{ type: "spring", bounce: 0.2, duration: 0.6 }}
                  />
                )}
                <item.icon size={18} className={isActive ? "text-memzent-glow" : "group-hover:text-white transition-colors"} />
                <span className="text-sm font-bold tracking-tight relative z-10">{item.label}</span>
                {isActive && <div className="ml-auto w-1 h-1 rounded-full bg-memzent-glow shadow-[0_0_10px_#00f3ff]" />}
              </div>
            </Link>
          );
        })}
      </nav>

      <div className="mt-auto pt-6 border-t border-white/5 space-y-4">
        <div className="glass p-4 rounded-2xl border-white/5 flex items-center gap-3">
          <div className="w-2 h-2 rounded-full bg-memzent-accent shadow-[0_0_8px_#00ff8e] animate-pulse" />
          <div className="text-[10px] font-mono font-bold text-white/60 uppercase">Gateway Node 01 - Online</div>
        </div>
        <div className="flex items-center gap-3 px-2">
          <div className="w-8 h-8 rounded-lg bg-white/5 border border-white/10 flex items-center justify-center text-white/40 hover:text-memzent-glow cursor-pointer transition-colors">
            <Globe size={14} />
          </div>
          <div className="text-[10px] font-bold text-white/50 uppercase tracking-tighter">Region: US-EAST-1</div>
        </div>
      </div>
    </aside>
  );
}
