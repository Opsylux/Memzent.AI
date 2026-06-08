-- Memzent SaaS Foundation Migration
-- 1. Create Organizations Table
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    stripe_customer_id TEXT UNIQUE,
    subscription_tier TEXT DEFAULT 'free', -- 'free', 'pro', 'business'
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- 2. Create Organization Memberships
CREATE TABLE IF NOT EXISTS org_memberships (
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL, -- References auth.users(id) in Supabase
    role TEXT DEFAULT 'member', -- 'admin', 'member'
    PRIMARY KEY (org_id, user_id)
);

-- 3. Extend Tools for Multi-tenancy
ALTER TABLE user_tools ADD COLUMN IF NOT EXISTS org_id UUID REFERENCES organizations(id) ON DELETE CASCADE;

-- 4. Auth Hook for JWT Custom Claims (Stateless Org Mapping)
-- This function will be called by Supabase Auth to add org_id to the token
CREATE OR REPLACE FUNCTION public.get_user_org_claims(user_id UUID)
RETURNS JSONB AS $$
DECLARE
    org_data JSONB;
BEGIN
    SELECT jsonb_build_object(
        'org_id', org_id,
        'role', role,
        'tier', (SELECT subscription_tier FROM organizations WHERE id = org_id)
    ) INTO org_data
    FROM org_memberships
    WHERE user_id = $1
    LIMIT 1;

    RETURN COALESCE(org_data, '{}'::JSONB);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 5. Trigger Example (Supabase Auth Hook)
-- This is a conceptual trigger that would be part of Supabase's auth hook logic
-- It ensures every JWT issued contains the user's primary organization ID.
