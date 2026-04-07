'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { 
  LayoutDashboard, 
  Database, 
  Key, 
  CreditCard, 
  Settings,
  Shield,
  LogOut,
  Building2,
  ChevronRight
} from 'lucide-react'
import { createClient } from '@supabase/supabase-js'
import { useRouter } from 'next/navigation'

const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL || ''
const supabaseAnonKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY || ''
const supabase = createClient(supabaseUrl, supabaseAnonKey)

const navItems = [
  { name: 'Overview', href: '/', icon: LayoutDashboard },
  { name: 'Aura Tools', href: '/tools', icon: Database },
  { name: 'API Keys', href: '/keys', icon: Key },
  { name: 'Billing', href: '/billing', icon: CreditCard },
  { name: 'Settings', href: '/settings', icon: Settings },
]

interface SidebarProps {
  orgName: string
  tier: string
  initials: string
}

export function Sidebar({ orgName, tier, initials }: SidebarProps) {
  const pathname = usePathname()
  const router = useRouter()

  const handleSignOut = async () => {
    await supabase.auth.signOut()
    router.push('/login')
  }

  const tierColors: Record<string, string> = {
    free: 'text-white/40 bg-white/5 border-white/10',
    pro: 'text-aura-purple bg-aura-purple/10 border-aura-purple/20',
    business: 'text-aura-glow bg-aura-glow/10 border-aura-glow/20',
  }

  return (
    <aside className="w-64 border-r border-white/5 bg-slate-950/50 backdrop-blur-xl h-screen sticky top-0 flex flex-col p-6">
      {/* Brand */}
      <div className="flex items-center gap-3 mb-8 px-2">
        <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-aura-glow to-aura-purple flex items-center justify-center shadow-[0_0_15px_rgba(0,243,255,0.3)]">
          <Shield size={18} className="text-black" strokeWidth={3} />
        </div>
        <span className="text-xl font-black tracking-tighter text-white">AURA</span>
      </div>

      {/* Organization Switcher */}
      <div className="mb-8 px-2">
        <button className="w-full flex items-center gap-3 p-3 rounded-xl bg-white/[0.03] border border-white/5 hover:border-aura-glow/20 transition-all group">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-aura-glow/20 to-aura-purple/20 flex items-center justify-center text-[10px] font-black text-aura-glow border border-aura-glow/10">
            {initials}
          </div>
          <div className="flex-1 text-left min-w-0">
            <div className="text-xs font-black text-white truncate">{orgName}</div>
            <div className={`text-[9px] font-black uppercase tracking-widest mt-0.5 inline-flex items-center gap-1 px-1.5 py-0.5 rounded border ${tierColors[tier] || tierColors.free}`}>
              {tier}
            </div>
          </div>
          <ChevronRight size={14} className="text-white/20 group-hover:text-aura-glow transition-colors" />
        </button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-2">
        {navItems.map((item) => {
          const isActive = pathname === item.href
          return (
            <Link
              key={item.name}
              href={item.href}
              className={`flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-bold transition-all ${
                isActive 
                  ? 'bg-aura-glow/10 text-aura-glow border border-aura-glow/20 shadow-[0_0_20px_rgba(0,243,255,0.05)]' 
                  : 'text-white/40 hover:text-white hover:bg-white/5'
              }`}
            >
              <item.icon size={18} />
              {item.name}
            </Link>
          )
        })}
      </nav>

      {/* Footer */}
      <div className="pt-6 border-t border-white/5 space-y-2">
        <div className="flex items-center gap-3 px-4 py-2">
          <Building2 size={14} className="text-white/20" />
          <span className="text-[10px] font-bold text-white/20 uppercase tracking-widest truncate">
            {orgName}
          </span>
        </div>
        <button 
          onClick={handleSignOut}
          className="flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-bold text-white/40 hover:text-red-400 hover:bg-red-500/5 transition-all w-full"
        >
          <LogOut size={18} />
          Sign Out
        </button>
      </div>
    </aside>
  )
}
