"use client"

import { Search, Bell, Cpu, Cloud, Database, LogOut } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { signOutAction } from "@/app/actions";
import { supabase } from "@/lib/supabase";
import { useRouter } from "next/navigation";

interface MemzentTopNavProps {
  orgName: string
  email: string
  initials: string
  tier: string
  role: string
}

export function MemzentTopNav({ orgName, email, initials, tier, role }: MemzentTopNavProps) {
  const router = useRouter()

  const handleSignOut = async () => {
    try {
      await signOutAction()
    } catch {
      // Server action may fail after redeploy (stale action ID)
      // Fallback: sign out via client-side Supabase
      await supabase.auth.signOut()
      router.push('/login')
      router.refresh()
    }
  }

  return (
    <header className="sticky top-0 z-40 p-3 sm:p-4 lg:p-6 flex items-center justify-between glass border-white/5 rounded-2xl lg:rounded-3xl neural-bg mx-3 mt-3 sm:mx-4 sm:mt-4 lg:m-4 min-h-[60px] lg:min-h-[80px]">
      {/* Search — hidden on small mobile, offset for hamburger */}
      <div className="flex-1 max-w-2xl relative group ml-10 lg:ml-0 hidden sm:block">
        <Search className="absolute left-4 lg:left-6 top-1/2 -translate-y-1/2 text-white/20 group-hover:text-memzent-glow transition-colors" size={16} />
        <input
          type="text"
          placeholder={`Search ${orgName}...`}
          className="w-full h-10 lg:h-14 bg-white/[0.03] border border-white/5 rounded-xl lg:rounded-2xl pl-10 lg:pl-14 pr-4 text-sm font-bold tracking-tight focus:outline-none focus:border-memzent-glow/30 focus:bg-white/[0.05] transition-all"
        />
      </div>

      {/* Spacer for mobile (hamburger button lives in sidebar component) */}
      <div className="sm:hidden ml-10 flex-1" />

      <div className="flex items-center gap-2 sm:gap-4 lg:gap-6">
        {/* Platform status indicators — desktop + platform_staff only */}
        {role === 'platform_staff' && (
          <div className="hidden xl:flex items-center gap-8 border-r border-white/10 pr-8 mr-2 text-[10px] font-black tracking-widest uppercase">
            <div className="flex flex-col gap-1 items-end">
              <div className="flex items-center gap-2 text-memzent-glow">
                <Cpu size={12} />
                <span>ROUTER_RUST</span>
              </div>
              <div className="text-white/50">99.2% Uptime</div>
            </div>
            <div className="flex flex-col gap-1 items-end">
              <div className="flex items-center gap-2 text-memzent-purple">
                <Cloud size={12} />
                <span>GATEWAY_GO</span>
              </div>
              <div className="text-white/50">0.42ms Latency</div>
            </div>
            <div className="flex flex-col gap-1 items-end">
              <div className="flex items-center gap-2 text-memzent-accent">
                <Database size={12} />
                <span>VECTOR_QDRANT</span>
              </div>
              <div className="text-white/50">1.21M Points</div>
            </div>
          </div>
        )}

        <div className="flex items-center gap-2 sm:gap-4">
          {/* Notifications bell */}
          <div className="relative p-2 sm:p-3 lg:p-4 rounded-xl hover:bg-white/5 cursor-pointer transition-all border border-transparent hover:border-white/10 group">
            <Bell size={18} className="text-white/60 group-hover:text-white" />
            <div className="absolute top-2 right-2 sm:top-3 sm:right-3 lg:top-4 lg:right-4 w-2 h-2 bg-memzent-glow rounded-full shadow-[0_0_8px_#00f3ff] border-2 border-[#050505]" />
          </div>

          {/* Sign out button — always visible */}
          <button
            onClick={handleSignOut}
            className="p-2 sm:p-3 lg:p-4 rounded-xl hover:bg-red-500/10 cursor-pointer transition-all border border-transparent hover:border-red-500/20 group"
            title="Sign Out"
          >
            <LogOut size={18} className="text-white/40 group-hover:text-red-400" />
          </button>

          <div className="hidden sm:block h-8 lg:h-12 w-px bg-white/10 mx-1" />

          {/* User avatar + info — compact on mobile */}
          <div className="flex items-center gap-2 sm:gap-3">
            <div className="hidden sm:block text-right">
              <div className="flex items-center gap-2 justify-end">
                {(role === 'platform_staff' || role === 'admin') && (
                  <Badge variant="outline" className="text-[8px] font-black px-1.5 py-0 rounded-md border-memzent-purple/30 text-memzent-purple uppercase tracking-tighter">
                    {role === 'platform_staff' ? 'STAFF' : 'ADMIN'}
                  </Badge>
                )}
                <div className="text-xs lg:text-sm font-black tracking-tight truncate max-w-[100px] lg:max-w-[140px] uppercase italic">{orgName}</div>
              </div>
              <div className="text-[9px] lg:text-[10px] font-bold text-white/50 uppercase truncate max-w-[100px] lg:max-w-[140px] tracking-widest">{email}</div>
            </div>
            <div className={`w-9 h-9 sm:w-10 sm:h-10 lg:w-12 lg:h-12 rounded-xl bg-gradient-to-br from-memzent-matrix to-memzent-dark border flex items-center justify-center p-0.5 shadow-xl transition-all ${role === 'platform_staff' ? 'border-memzent-purple/50 shadow-[0_0_15px_rgba(157,0,255,0.2)]' : 'border-white/10'
              }`}>
              <div className="w-full h-full rounded-lg bg-memzent-dark flex items-center justify-center">
                <span className={`text-[10px] sm:text-xs font-black opacity-80 ${role === 'platform_staff' ? 'text-memzent-purple' : 'text-memzent-glow'
                  }`}>{initials}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}
