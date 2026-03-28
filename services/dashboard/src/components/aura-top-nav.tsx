"use client"

import { Search, Bell, Cpu, Cloud, Database } from "lucide-react";
import { motion } from "framer-motion";

export function AuraTopNav() {
  return (
    <header className="fixed top-0 right-0 left-72 h-32 z-40 p-6 m-4 ml-0 flex items-center justify-between glass border-white/5 rounded-3xl neural-bg">
      <div className="flex-1 max-w-2xl relative group">
        <Search className="absolute left-6 top-1/2 -translate-y-1/2 text-white/20 group-hover:text-aura-glow transition-colors" size={18} />
        <input 
          type="text" 
          placeholder="Global Neural Search: Tools, Traces, Logs..."
          className="w-full h-16 bg-white/[0.03] border border-white/5 rounded-2xl pl-16 pr-6 text-sm font-bold tracking-tight focus:outline-none focus:border-aura-glow/30 focus:bg-white/[0.05] transition-all"
        />
        <div className="absolute right-6 top-1/2 -translate-y-1/2 flex items-center gap-2 opacity-20 group-hover:opacity-100 transition-all font-mono text-[10px] font-bold">
           <span className="bg-white/10 px-1.5 py-0.5 rounded">⌘</span>
           <span className="bg-white/10 px-1.5 py-0.5 rounded">K</span>
        </div>
      </div>

      <div className="flex items-center gap-6">
        <div className="flex items-center gap-8 border-r border-white/10 pr-8 mr-2 text-[10px] font-black tracking-widest uppercase">
          <div className="flex flex-col gap-1 items-end">
             <div className="flex items-center gap-2 text-aura-glow">
               <Cpu size={12} />
               <span>ROUTER_RUST</span>
             </div>
             <div className="text-white/20">99.2% Uptime</div>
          </div>
          <div className="flex flex-col gap-1 items-end">
             <div className="flex items-center gap-2 text-aura-purple">
               <Cloud size={12} />
               <span>GATEWAY_GO</span>
             </div>
             <div className="text-white/20">0.42ms Latency</div>
          </div>
          <div className="flex flex-col gap-1 items-end">
             <div className="flex items-center gap-2 text-aura-accent">
               <Database size={12} />
               <span>VECTOR_QDRANT</span>
             </div>
             <div className="text-white/20">1.21M Points</div>
          </div>
        </div>

        <div className="flex items-center gap-4">
           <div className="relative p-4 rounded-xl hover:bg-white/5 cursor-pointer transition-all border border-transparent hover:border-white/10 group">
             <Bell size={20} className="text-white/40 group-hover:text-white" />
             <div className="absolute top-4 right-4 w-2 h-2 bg-aura-glow rounded-full shadow-[0_0_8px_#00f3ff] border-2 border-[#050505]" />
           </div>
           
           <div className="h-12 w-px bg-white/10 mx-2" />

           <div className="flex items-center gap-3 pl-2">
             <div className="text-right">
                <div className="text-sm font-black tracking-tight">Enterprise_Node_01</div>
                <div className="text-[10px] font-bold text-white/30 uppercase">SuperUser Auth</div>
             </div>
             <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-aura-matrix to-aura-dark border border-white/10 flex items-center justify-center p-0.5 shadow-xl">
               <div className="w-full h-full rounded-lg bg-aura-dark flex items-center justify-center">
                 <span className="text-xs font-black text-aura-glow opacity-80">EX</span>
               </div>
             </div>
           </div>
        </div>
      </div>
    </header>
  );
}
