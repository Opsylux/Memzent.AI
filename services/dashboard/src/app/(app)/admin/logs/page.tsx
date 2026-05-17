import { AuditLogFeed } from "@/components/audit-log-feed";
import { getCurrentOrg } from "@/lib/user-context";
import { Terminal, Shield } from "lucide-react";
import { redirect } from "next/navigation";

export default async function AdminLogsPage() {
   const org = await getCurrentOrg();

   // RBAC Hardening
   if (org?.role !== 'platform_staff' && org?.role !== 'admin') {
      if (org?.role !== 'platform_staff') {
         redirect("/");
      }
   }

   return (
      <div className="h-[calc(100vh-120px)] flex flex-col space-y-8">
         <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
               <div className="w-2 h-8 rounded-full bg-gradient-to-b from-memzent-glow to-memzent-purple" />
               <div>
                  <h1 className="text-3xl font-black tracking-tighter uppercase text-white">System Audit Trace</h1>
                  <p className="text-[10px] font-black text-white/20 uppercase tracking-[0.3em] italic">Global Intelligence Feed</p>
               </div>
            </div>

            <div className="flex gap-4">
               <div className="px-4 py-2 rounded-xl bg-white/5 border border-white/10 flex items-center gap-3">
                  <div className="w-2 h-2 rounded-full bg-memzent-glow animate-pulse" />
                  <span className="text-[10px] font-black text-white/40 uppercase tracking-widest leading-none">Global Stream Online</span>
               </div>
            </div>
         </div>

         <div className="flex-1 min-h-0">
            {/* Passing no orgId to the feed to request global logs from the gateway */}
            <AuditLogFeed />
         </div>

         {/* Admin Quick Insight */}
         <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mt-4">
            <div className="p-4 rounded-2xl bg-memzent-purple/5 border border-memzent-purple/10 flex items-center justify-between group hover:border-memzent-purple/30 transition-all">
               <div className="flex items-center gap-3">
                  <Shield size={16} className="text-memzent-purple" />
                  <span className="text-[10px] font-black text-white/60 uppercase tracking-widest">Retention Mode</span>
               </div>
               <span className="text-[10px] font-black text-memzent-purple italic">30 DAYS (Active)</span>
            </div>
            <div className="p-4 rounded-2xl bg-white/[0.02] border border-white/5 flex items-center justify-between">
               <div className="flex items-center gap-3">
                  <Terminal size={16} className="text-memzent-glow/40" />
                  <span className="text-[10px] font-black text-white/60 uppercase tracking-widest">Persistence</span>
               </div>
               <span className="text-[10px] font-black text-memzent-glow italic uppercase">Postgres Hardened</span>
            </div>
            <div className="p-4 rounded-2xl bg-white/[0.02] border border-white/5 flex items-center justify-between">
               <div className="flex items-center gap-3">
                  <Activity size={16} className="text-white/20" />
                  <span className="text-[10px] font-black text-white/60 uppercase tracking-widest">Feed Status</span>
               </div>
               <span className="text-[10px] font-black text-memzent-accent italic uppercase tracking-tighter animate-pulse">Live polling...</span>
            </div>
         </div>
      </div>
   );
}

import { Activity } from "lucide-react";
