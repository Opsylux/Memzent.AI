'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { supabase } from '@/lib/supabase'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card'
import { Github, Mail } from 'lucide-react'

export default function LoginPage() {
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')
  const router = useRouter()

  useEffect(() => {
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
      if (session) {
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
      setMessage('Check your email for the login link!')
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
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-[#050505] px-2 text-white/50">Or continue with</span>
            </div>
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
