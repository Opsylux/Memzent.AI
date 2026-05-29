"use server"

import { createClient } from '@/lib/supabase-server'
import { cache } from 'react'

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
export const getCurrentOrg = cache(async (): Promise<OrgContext | null> => {
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

    // No membership found -- call the secure database RPC to generate a real one!
    const { data: newOrgId, error: provisionErr } = await supabase.rpc('provision_personal_org')
    if (provisionErr) {
      console.error('Failed to provision personal org:', provisionErr)
      return null
    }

    const name = user.email?.split('@')[0] || 'Personal'
    return {
      userId: user.id,
      email: user.email || '',
      orgId: newOrgId || user.id, // Fallback just in case
      orgName: name + "'s Workspace",
      tier: 'free',
      role: 'admin',
      initials: name.substring(0, 2).toUpperCase(),
    }
  } catch (e) {
    console.error('Failed to resolve org context:', e)
    return null
  }
})
