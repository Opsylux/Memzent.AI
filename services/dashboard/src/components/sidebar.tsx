'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useState, useEffect, useRef } from 'react'
import {
  LayoutDashboard,
  Database,
  Key,
  CreditCard,
  Settings,
  Shield,
  LogOut,
  Building2,
  ChevronRight,
  Book,
  Activity,
  Cpu,
  FlaskConical,
  Bell,
  GitBranch,
  Gauge,
  Menu,
  X,
  DollarSign
} from 'lucide-react'
import { signOutAction } from '@/app/actions'
import { supabase } from '@/lib/supabase'
import { useRouter } from 'next/navigation'

const navItems = [
  { name: 'Overview', href: '/', icon: LayoutDashboard },
  { name: 'Analytics', href: '/analytics', icon: Activity },
  { name: 'GPU Analytics', href: '/analytics/gpu', icon: Gauge },
  { name: 'Playground', href: '/playground', icon: FlaskConical },
  { name: 'Memzent Tools', href: '/tools', icon: Database },
  { name: 'Workflows', href: '/workflows', icon: GitBranch },
  { name: 'Providers', href: '/providers', icon: Cpu },
  { name: 'API Keys', href: '/keys', icon: Key },
  { name: 'Notifications', href: '/notifications', icon: Bell },
  { name: 'Audit Logs', href: '/audit', icon: Activity },
  { name: 'Billing', href: '/billing', icon: CreditCard },
  { name: 'Spend Limits', href: '/billing/spend-limits', icon: DollarSign },
  { name: 'Documentation', href: '/docs', icon: Book },
  { name: 'Settings', href: '/settings', icon: Settings },
]

// Items only visible to org admins (not viewers/members)
const adminItems = [
  { name: 'Providers', href: '/providers', icon: Cpu },
  { name: 'API Keys', href: '/keys', icon: Key },
  { name: 'Settings', href: '/settings', icon: Settings },
]

const staffItems = [
  { name: 'Global Nodes', href: '/admin/nodes', icon: Shield },
  { name: 'System Logs', href: '/admin/logs', icon: Database },
]

interface SidebarProps {
  orgName: string
  tier: string
  initials: string
  role: string
}

export function Sidebar({ orgName, tier, initials, role }: SidebarProps) {
  const pathname = usePathname()
  const router = useRouter()
  const [mobileOpen, setMobileOpen] = useState(false)
  const prevPathname = useRef(pathname)

  // Close mobile menu on route change (avoid setState directly in effect)
  if (prevPathname.current !== pathname) {
    prevPathname.current = pathname
    if (mobileOpen) setMobileOpen(false)
  }

  // Prevent body scroll when mobile menu is open
  useEffect(() => {
    if (mobileOpen) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
    return () => { document.body.style.overflow = '' }
  }, [mobileOpen])

  const handleSignOut = async () => {
    try {
      await signOutAction()
    } catch {
      // Server action may fail after redeploy (stale action ID)
      await supabase.auth.signOut()
      router.push('/login')
      router.refresh()
    }
  }

  const tierColors: Record<string, string> = {
    free: 'text-white/40 bg-white/5 border-white/10',
    pro: 'text-memzent-purple bg-memzent-purple/10 border-memzent-purple/20',
    business: 'text-memzent-glow bg-memzent-glow/10 border-memzent-glow/20',
  }

  const filteredNavItems = navItems.filter((item) => {
    if (role === 'viewer' || role === 'member') {
      return !adminItems.some(ai => ai.href === item.href)
    }
    return true
  })

  const sidebarContent = (
    <>
      {/* Brand */}
      <div className="flex items-center justify-between mb-6 lg:mb-8 px-2">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-memzent-glow to-memzent-purple flex items-center justify-center shadow-[0_0_15px_rgba(0,243,255,0.3)]">
            <Shield size={18} className="text-black" strokeWidth={3} />
          </div>
          <span className="text-xl font-black tracking-tighter text-white">MEMZENT</span>
        </div>
        {/* Close button — mobile only */}
        <button
          onClick={() => setMobileOpen(false)}
          className="lg:hidden p-2 rounded-lg hover:bg-white/5 text-white/60"
        >
          <X size={20} />
        </button>
      </div>

      {/* Organization Switcher */}
      <div className="mb-6 lg:mb-8 px-2">
        <button className="w-full flex items-center gap-3 p-3 rounded-xl bg-white/[0.03] border border-white/5 hover:border-memzent-glow/20 transition-all group">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-memzent-glow/20 to-memzent-purple/20 flex items-center justify-center text-[10px] font-black text-memzent-glow border border-memzent-glow/10">
            {initials}
          </div>
          <div className="flex-1 text-left min-w-0">
            <div className="text-xs font-black text-white truncate">{orgName}</div>
            <div className={`text-[9px] font-black uppercase tracking-widest mt-0.5 inline-flex items-center gap-1 px-1.5 py-0.5 rounded border ${tierColors[tier] || tierColors.free}`}>
              {tier}
            </div>
          </div>
          <ChevronRight size={14} className="text-white/20 group-hover:text-memzent-glow transition-colors" />
        </button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-1 overflow-y-auto">
        <div className="text-[10px] font-black uppercase text-white/20 tracking-[0.2em] mb-3 px-4 italic">Navigation</div>
        {filteredNavItems.map((item) => {
          const isActive = pathname === item.href
          return (
            <Link
              key={item.name}
              href={item.href}
              className={`flex items-center gap-3 px-4 py-2.5 lg:py-3 rounded-xl text-sm font-bold transition-all ${isActive
                ? 'bg-memzent-glow/10 text-memzent-glow border border-memzent-glow/20 shadow-[0_0_20px_rgba(0,243,255,0.05)]'
                : 'text-white/40 hover:text-white hover:bg-white/5'
                }`}
            >
              <item.icon size={18} />
              {item.name}
            </Link>
          )
        })}

        {role === 'platform_staff' && (
          <div className="pt-6 space-y-1">
            <div className="text-[10px] font-black uppercase text-memzent-purple/40 tracking-[0.2em] mb-3 px-4 italic">Admin_Ops</div>
            {staffItems.map((item) => {
              const isActive = pathname === item.href
              return (
                <Link
                  key={item.name}
                  href={item.href}
                  className={`flex items-center gap-3 px-4 py-2.5 lg:py-3 rounded-xl text-sm font-bold transition-all ${isActive
                    ? 'bg-memzent-purple/10 text-memzent-purple border border-memzent-purple/20 shadow-[0_0_20px_rgba(151,71,255,0.05)]'
                    : 'text-white/40 hover:text-memzent-purple hover:bg-white/5'
                    }`}
                >
                  <item.icon size={18} />
                  {item.name}
                </Link>
              )
            })}
          </div>
        )}
      </nav>

      {/* Footer with Sign Out */}
      <div className="pt-4 border-t border-white/5 space-y-2 mt-4">
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
    </>
  )

  return (
    <>
      {/* Mobile hamburger trigger */}
      <button
        onClick={() => setMobileOpen(true)}
        className="lg:hidden fixed top-4 left-4 z-50 p-3 rounded-xl bg-slate-900/90 border border-white/10 backdrop-blur-xl shadow-lg"
        aria-label="Open menu"
      >
        <Menu size={20} className="text-white" />
      </button>

      {/* Mobile overlay */}
      {mobileOpen && (
        <div
          className="lg:hidden fixed inset-0 bg-black/60 backdrop-blur-sm z-40"
          onClick={() => setMobileOpen(false)}
        />
      )}

      {/* Mobile sidebar (slide-in) */}
      <aside
        className={`lg:hidden fixed inset-y-0 left-0 z-50 w-72 bg-slate-950/95 backdrop-blur-xl border-r border-white/5 flex flex-col p-5 transform transition-transform duration-300 ease-in-out ${
          mobileOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        {sidebarContent}
      </aside>

      {/* Desktop sidebar (always visible) */}
      <aside className="hidden lg:flex w-64 border-r border-white/5 bg-slate-950/50 backdrop-blur-xl h-screen sticky top-0 flex-col p-6">
        {sidebarContent}
      </aside>
    </>
  )
}
