import { createClient } from '@/lib/supabase-server'
import { NextResponse } from 'next/server'

export async function GET(request: Request) {
  const requestUrl = new URL(request.url)
  const code = requestUrl.searchParams.get('code')
  const accessToken = requestUrl.searchParams.get('access_token')
  const refreshToken = requestUrl.searchParams.get('refresh_token')
  const origin = requestUrl.origin

  const supabase = await createClient()

  // Standard PKCE Auth Code Flow
  if (code) {
    const { data: { user }, error } = await supabase.auth.exchangeCodeForSession(code)
    if (error || !user) {
      console.error('Auth code exchange failed:', error)
      // PKCE verifier missing = user opened callback in different browser/tab
      // or cookies were cleared. Redirect to login with helpful message.
      const errorParam = error?.code === 'pkce_code_verifier_not_found'
        ? 'session_expired'
        : 'auth_exchange_failed'
      return NextResponse.redirect(`${origin}/login?error=${errorParam}`)
    }

    // Provisioning Logic: Ensure user has a default organization
    const { data: membership } = await supabase
      .from('members')
      .select('org_id')
      .eq('user_id', user.id)
      .limit(1)
      .maybeSingle()

    if (!membership) {
      console.log('Provisioning personal organization for new user:', user.id)
      
      // 1. Create Organization
      const displayName = user.user_metadata?.full_name || user.email?.split('@')[0] || 'Member'
      const { data: org, error: orgErr } = await supabase
        .from('organizations')
        .insert({
          name: `${displayName}'s Workspace`,
          subscription_tier: 'free'
        })
        .select()
        .single()

      if (!orgErr && org) {
        // 2. Add as Admin Member
        await supabase
          .from('members')
          .insert({
            org_id: org.id,
            user_id: user.id,
            role: 'admin'
          })
      }
    }

    return NextResponse.redirect(`${origin}/`)
  }

  // Fallback Implicit Auth Flow interception
  if (accessToken && refreshToken) {
    console.log('Intercepted Implicit Auth payload on server callback')
    const { error } = await supabase.auth.setSession({
      access_token: accessToken,
      refresh_token: refreshToken,
    })
    
    if (error) {
      console.error('Auth manual session injection failed:', error)
      return NextResponse.redirect(`${origin}/login?error=token_injection_failed`)
    }
    return NextResponse.redirect(`${origin}/`)
  }

  return NextResponse.redirect(`${origin}/`)
}
