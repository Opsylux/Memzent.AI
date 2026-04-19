-- Migration 010: Infrastructure Hardening & Table Reconciliation
-- Standardizes on 'tools' table and ensures 'audit_logs' compatibility with system events.

-- 1. Ensure System Organization Sentinel exists
INSERT INTO organizations (id, name, slug, subscription_tier)
VALUES ('00000000-0000-0000-0000-000000000000', 'Aura Platform (System)', 'aura-system', 'business')
ON CONFLICT (id) DO NOTHING;

-- 2. Reconcile Tools Discrepancy
-- Gateway uses 'tools', but '004' created 'tool_registry'.
-- We move data from 'tool_registry' to 'tools' if it exists, then drop 'tool_registry'.

DO $$ 
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'tool_registry') THEN
        INSERT INTO tools (id, name, description, connector_type, endpoint, enabled, requires_auth, updated_at)
        SELECT id, name, description, 'mcp', 'mcp_server:' || id, is_active, true, created_at
        FROM tool_registry
        ON CONFLICT (id) DO NOTHING;
        
        DROP TABLE tool_registry CASCADE;
    END IF;
END $$;

-- 3. Hardening tools table (ensure it matches registry.go struct)
ALTER TABLE tools ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}'::JSONB;
ALTER TABLE tools ADD COLUMN IF NOT EXISTS org_id UUID REFERENCES organizations(id) ON DELETE CASCADE;

-- 4. Set default org_id for existing tools to System ORG if null
UPDATE tools SET org_id = '00000000-0000-0000-0000-000000000000' WHERE org_id IS NULL;

-- 5. Hardening audit_logs
-- Ensure sentinel can be used and index is efficient
CREATE INDEX IF NOT EXISTS idx_audit_logs_system_events ON audit_logs(created_at DESC) WHERE org_id = '00000000-0000-0000-0000-000000000000';
