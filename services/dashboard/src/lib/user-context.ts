"use server"

import { createClient } from '@/lib/supabase-server'

export interface OrgContext {
  userId: string
  email: string
  orgId: string
  orgName: string
  tier: string
  role: string
  initials: string
}

/**
 * Resolves the current user's organization context from Supabase.
 * 
 * Flow:
 * 1. Get authenticated user from session cookies
 * 2. Look up their membership in the `members` table
 * 3. Join with `organizations` to get org details
 * 4. Fall back to a "Personal" org seeded from user.id if no membership found
 */
export async function getCurrentOrg(): Promise<OrgContext | null> {
  try {
    const supabase = await createClient()
    const { data: { user }, error: userErr } = await supabase.auth.getUser()

    if (userErr || !user) {
      return null
    }

    // Try to find the user's organization via the members table
    const { data: membership } = await supabase
      .from('members')
      .select('org_id, role, organizations(id, name, subscription_tier)')
      .eq('user_id', user.id)
      .limit(1)
      .maybeSingle()

    if (membership?.organizations) {
      const org = membership.organizations as any
      const name = org.name || 'Unnamed Org'
      const role = membership.role || 'member'
      return {
        userId: user.id,
        email: user.email || '',
        orgId: org.id,
        orgName: name,
        tier: org.subscription_tier || 'free',
        role: role,
        initials: name.substring(0, 2).toUpperCase(),
      }
    }

    // Fallback: No membership found — use the user's own ID as a "personal" org
    // This allows the dashboard to function even before the org/member tables exist
    const displayName = user.user_metadata?.full_name || user.email?.split('@')[0] || 'User'
    return {
      userId: user.id,
      email: user.email || '',
      orgId: user.id,
      orgName: `${displayName}'s Workspace`,
      tier: 'free',
      role: 'admin', // Owner of personal workspace is admin
      initials: displayName.substring(0, 2).toUpperCase(),
    }
  } catch (e) {
    console.error('Failed to resolve org context:', e)
    return null
  }
}
