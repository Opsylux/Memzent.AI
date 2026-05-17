-- Memzent SaaS Database Foundation & Gateway RBAC Setup
-- Execute this file against Cloud Supabase to create the core tables

-- 1. Organizations
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    stripe_customer_id TEXT UNIQUE,
    subscription_tier TEXT DEFAULT 'free', -- 'free', 'pro', 'business'
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- 2. Members (Maps users to organizations)
-- Using table name "members" to match the Next.js getCurrentOrg() logic
CREATE TABLE IF NOT EXISTS members (
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL, -- References auth.users(id) conceptually
    role TEXT DEFAULT 'member', -- 'admin', 'member'
    PRIMARY KEY (org_id, user_id)
);

-- 3. Org Tools (Gateway Live RBAC Permissions)
-- Defines which tools an Organization has been granted access to
CREATE TABLE IF NOT EXISTS org_tools (
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    tool_id VARCHAR(255) REFERENCES tools(id) ON DELETE CASCADE,
    PRIMARY KEY (org_id, tool_id)
);
