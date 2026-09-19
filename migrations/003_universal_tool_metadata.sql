-- Migration: Universal Tool Metadata & Multi-tenancy
--
-- NOTE: this migration references organizations(id) via FK, but the
-- `organizations` table itself is defined in 004_org_rbac.sql (which numeric
-- migration ordering would normally run *after* this file). To keep this
-- migration safe to apply standalone/in-order, we idempotently ensure a
-- minimal `organizations` table exists first; 004_org_rbac.sql's
-- `CREATE TABLE IF NOT EXISTS organizations` will then simply no-op (or add
-- any columns not yet present) when it runs.
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    stripe_customer_id TEXT UNIQUE,
    subscription_tier TEXT DEFAULT 'free', -- 'free', 'pro', 'business'
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- 1. Add Organization Isolation and Dynamic Configuration
ALTER TABLE tools ADD COLUMN IF NOT EXISTS org_id UUID REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE tools ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}'::JSONB;

-- 2. Update Indexes for SaaS Scale
-- Drop old indexes and create org-scoped ones
DROP INDEX IF EXISTS idx_tools_enabled;
DROP INDEX IF EXISTS idx_tools_connector_type;

CREATE INDEX IF NOT EXISTS idx_tools_org_enabled ON tools(org_id, enabled);
CREATE INDEX IF NOT EXISTS idx_tools_org_type ON tools(org_id, connector_type) WHERE enabled = true;

-- 3. Comment on Columns
COMMENT ON COLUMN tools.endpoint IS 'Base URL or primary locator for the tool (URL for REST, ConnString for SQL)';
COMMENT ON COLUMN tools.config IS 'Tool-specific dynamic configuration (Headers for REST, Queries for SQL)';
