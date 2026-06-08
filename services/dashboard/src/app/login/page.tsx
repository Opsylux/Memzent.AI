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
    // Show error from auth callback redirect
    const params = new URLSearchParams(window.location.search)
    const error = params.get('error')
    if (error === 'session_expired') {
      setMessage('Your login session expired. Please sign in again.')
    } else if (error === 'auth_exchange_failed') {
      setMessage('Authentication failed. Please try again.')
    }

    // Manually parse tokens in case Supabase drops us at /login directly
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
        emailRedirectTo: `${process.env.NEXT_PUBLIC_APP_URL}/auth/callback`,
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
        redirectTo: `${process.env.NEXT_PUBLIC_APP_URL}/auth/callback`,
      },
    })
  }

  return (
    <div className="flex items-center justify-center min-h-screen neural-bg">
      <Card className="w-[400px] border-white/10 bg-white/[0.03] backdrop-blur-xl text-white">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl font-bold tracking-tight text-white">Welcome to Memzent</CardTitle>
          <CardDescription className="text-white/60">
            Sign in to the memory of an agent
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <Button
              variant="outline"
              className="border-white/20 bg-white/5 hover:bg-white/10 text-white transition-colors"
              onClick={() => handleOAuthLogin('github')}
            >
              <Github className="mr-2 h-4 w-4" />
              GitHub
            </Button>
            <Button
              variant="outline"
              className="border-white/20 bg-white/5 hover:bg-white/10 text-white transition-colors"
              onClick={() => handleOAuthLogin('google')}
            >
              <Mail className="mr-2 h-4 w-4" />
              Google
            </Button>
          </div>

          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t border-white/10" />
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

          <form onSubmit={handleEmailLogin} className="space-y-3">
            <Input
              type="email"
              placeholder="name@company.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="border-white/20 bg-white/5 text-white placeholder:text-white/30 focus:ring-memzent-glow/50"
              required
            />
            <Button
              type="submit"
              className="w-full bg-memzent-glow text-black font-bold hover:shadow-[0_0_20px_rgba(0,243,255,0.3)] transition-all"
              disabled={loading}
            >
              {loading ? 'Sending...' : 'Sign in with Email'}
            </Button>
          </form>
          {message && <p className="text-center text-sm text-blue-400">{message}</p>}
        </CardContent>
        <CardFooter className="flex flex-col text-center text-xs text-white/50">
          <p>By signing in, you agree to our Terms of Service</p>
          <p>and Privacy Policy.</p>
        </CardFooter>
      </Card>
    </div>
  )
}
