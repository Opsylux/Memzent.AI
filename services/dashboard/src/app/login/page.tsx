'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { supabase } from '@/lib/supabase'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Github, Mail, Zap, Shield, Layers } from 'lucide-react'

const FEATURES = [
  { icon: Zap, title: 'Triple-layer cache', desc: 'Literal, canonical, and semantic hits before any LLM call' },
  { icon: Shield, title: 'Org-scoped RBAC', desc: 'API keys, JWT auth, and per-tool permissions' },
  { icon: Layers, title: 'Smart tool routing', desc: 'Vector-matched MCP, REST, and SQL connectors' },
]

export default function LoginPage() {
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')
  const router = useRouter()

  useEffect(() => {
    const urlParams = new URLSearchParams(
      window.location.hash.replace('#', '?') + '&' + window.location.search.replace('?', '&')
    )
    const accessToken = urlParams.get('access_token')
    const refreshToken = urlParams.get('refresh_token')

    if (accessToken && refreshToken) {
      supabase.auth.setSession({
        access_token: accessToken,
        refresh_token: refreshToken,
      }).then(() => {
        router.push('/')
        router.refresh()
      })
    }

    const { data: { subscription } } = supabase.auth.onAuthStateChange((event, session) => {
      if (session && (event === 'SIGNED_IN' || event === 'TOKEN_REFRESHED')) {
        router.push('/')
        router.refresh()
      }
    })
    return () => subscription.unsubscribe()
  }, [router])

  const handleEmailLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    const { error } = await supabase.auth.signInWithOtp({
      email,
      options: {
        emailRedirectTo: `${window.location.origin}/auth/callback`,
      },
    })
    if (error) {
      setMessage(error.message)
    } else {
      setMessage('Check your email for the login link.')
    }
    setLoading(false)
  }

  const handleOAuthLogin = async (provider: 'google' | 'github') => {
    await supabase.auth.signInWithOAuth({
      provider,
      options: {
        redirectTo: `${window.location.origin}/auth/callback`,
      },
    })
  }

  return (
    <div className="min-h-screen neural-bg flex">
      {/* Brand panel */}
      <div className="hidden lg:flex lg:w-[52%] relative flex-col justify-between p-12 border-r border-white/5 overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-memzent-glow/10 via-transparent to-memzent-purple/10 pointer-events-none" />
        <div className="absolute top-1/4 -left-20 w-80 h-80 bg-memzent-glow/20 blur-[100px] rounded-full" />
        <div className="absolute bottom-1/4 right-0 w-64 h-64 bg-memzent-purple/15 blur-[80px] rounded-full" />

        <div className="relative z-10">
          <div className="text-2xl font-black tracking-tight text-readable-primary">
            Memzent<span className="text-memzent-glow">.ai</span>
          </div>
          <p className="text-sm text-readable-muted mt-1">Command center for agent memory</p>
        </div>

        <div className="relative z-10 space-y-8 max-w-md">
          <h1 className="text-4xl font-black tracking-tight text-readable-primary leading-tight">
            The memory &amp; security layer for agentic AI
          </h1>
          <ul className="space-y-5">
            {FEATURES.map((f) => (
              <li key={f.title} className="flex gap-4">
                <div className="shrink-0 w-10 h-10 rounded-xl bg-white/5 border border-white/10 flex items-center justify-center text-memzent-glow">
                  <f.icon size={18} />
                </div>
                <div>
                  <div className="text-sm font-bold text-readable-primary">{f.title}</div>
                  <p className="text-sm text-readable-muted mt-0.5 leading-snug">{f.desc}</p>
                </div>
              </li>
            ))}
          </ul>
        </div>

        <p className="relative z-10 text-xs text-readable-muted">
          Built for teams shipping production agents — not just chat demos.
        </p>
      </div>

      {/* Sign-in panel */}
      <div className="flex-1 flex items-center justify-center p-6 sm:p-10">
        <div className="w-full max-w-[400px] space-y-8">
          <div className="lg:hidden text-center">
            <div className="text-xl font-black text-readable-primary">
              Memzent<span className="text-memzent-glow">.ai</span>
            </div>
          </div>

          <div className="glass rounded-2xl p-8 border-white/10">
            <div className="text-center mb-8">
              <h2 className="text-xl font-bold text-readable-primary">Sign in</h2>
              <p className="text-sm text-readable-muted mt-2">
                Access your org dashboard, keys, and playground
              </p>
            </div>

            <div className="grid grid-cols-2 gap-3 mb-6">
              <Button
                variant="outline"
                className="border-white/15 bg-white/5 hover:bg-white/10 text-readable-primary h-11"
                onClick={() => handleOAuthLogin('github')}
              >
                <Github className="mr-2 h-4 w-4" />
                GitHub
              </Button>
              <Button
                variant="outline"
                className="border-white/15 bg-white/5 hover:bg-white/10 text-readable-primary h-11"
                onClick={() => handleOAuthLogin('google')}
              >
                <Mail className="mr-2 h-4 w-4" />
                Google
              </Button>
            </div>

            <div className="relative mb-6">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t border-white/10" />
              </div>
              <div className="relative flex justify-center text-xs">
                <span className="bg-[#0a0a0a] px-3 text-readable-muted">or email</span>
              </div>
            </div>

            <form onSubmit={handleEmailLogin} className="space-y-4">
              <Input
                type="email"
                placeholder="you@company.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="h-11 border-white/15 bg-black/30 text-readable-primary placeholder:text-readable-muted focus-visible:ring-memzent-glow/40"
                required
              />
              <Button
                type="submit"
                className="w-full h-11 bg-memzent-glow text-black font-bold hover:shadow-[0_0_24px_rgba(0,243,255,0.35)]"
                disabled={loading}
              >
                {loading ? 'Sending link…' : 'Continue with email'}
              </Button>
            </form>

            {message && (
              <p className={`text-center text-sm mt-4 ${message.includes('Check') ? 'text-memzent-accent' : 'text-red-400'}`}>
                {message}
              </p>
            )}
          </div>

          <p className="text-center text-xs text-readable-muted leading-relaxed">
            By signing in you agree to our Terms of Service and Privacy Policy.
          </p>
        </div>
      </div>
    </div>
  )
}
