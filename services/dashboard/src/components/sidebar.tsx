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
  LogOut
} from 'lucide-react'
import { supabase } from '@/lib/supabase'
import { useRouter } from 'next/navigation'

const navItems = [
  { name: 'Overview', href: '/', icon: LayoutDashboard },
  { name: 'Aura Tools', href: '/tools', icon: Database },
  { name: 'API Keys', href: '/keys', icon: Key },
  { name: 'Billing', href: '/billing', icon: CreditCard },
  { name: 'Settings', href: '/settings', icon: Settings },
]

export function Sidebar() {
  const pathname = usePathname()
  const router = useRouter()

  const handleSignOut = async () => {
    await supabase.auth.signOut()
    router.push('/login')
  }

  return (
    <aside className="w-64 border-r border-white/5 bg-slate-950/50 backdrop-blur-xl h-screen sticky top-0 flex flex-col p-6">
      <div className="flex items-center gap-3 mb-12 px-2">
        <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-aura-glow to-aura-purple flex items-center justify-center shadow-[0_0_15px_rgba(0,243,255,0.3)]">
          <Shield size={18} className="text-black" strokeWidth={3} />
        </div>
        <span className="text-xl font-black tracking-tighter text-white">AURA</span>
      </div>

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

      <div className="pt-6 border-t border-white/5">
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
