'use client'

import { useState } from 'react'
import { supabase } from '@/lib/supabase'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card'
import { Github, Mail } from 'lucide-react'

export default function LoginPage() {
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')

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
    <div className="flex items-center justify-center min-h-screen bg-slate-950">
      <Card className="w-[400px] border-slate-800 bg-slate-900 text-slate-100">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl font-bold tracking-tight text-white">Welcome to Aura</CardTitle>
          <CardDescription className="text-slate-400">
            Sign in to your enterprise AI mesh
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <Button 
              variant="outline" 
              className="border-slate-700 bg-slate-800 hover:bg-slate-700 transition-colors"
              onClick={() => handleOAuthLogin('github')}
            >
              <Github className="mr-2 h-4 w-4" />
              GitHub
            </Button>
            <Button 
              variant="outline"
              className="border-slate-700 bg-slate-800 hover:bg-slate-700 transition-colors"
              onClick={() => handleOAuthLogin('google')}
            >
              <Mail className="mr-2 h-4 w-4" />
              Google
            </Button>
          </div>
          
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t border-slate-800" />
            </div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-slate-900 px-2 text-slate-500">Or continue with</span>
            </div>
          </div>

          <form onSubmit={handleEmailLogin} className="space-y-3">
            <Input
              type="email"
              placeholder="name@company.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="border-slate-700 bg-slate-950 text-white placeholder:text-slate-600 focus:ring-blue-500"
              required
            />
            <Button 
              type="submit" 
              className="w-full bg-blue-600 hover:bg-blue-500 text-white font-medium"
              disabled={loading}
            >
              {loading ? 'Sending...' : 'Sign in with Email'}
            </Button>
          </form>
          {message && <p className="text-center text-sm text-blue-400">{message}</p>}
        </CardContent>
        <CardFooter className="flex flex-col text-center text-xs text-slate-500">
          <p>By signing in, you agree to our Terms of Service</p>
          <p>and Privacy Policy.</p>
        </CardFooter>
      </Card>
    </div>
  )
}
