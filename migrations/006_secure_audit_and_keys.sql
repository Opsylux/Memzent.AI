-- Migration: Secure Audit Logs and Hardened API Keys
-- Part of Aura RC1 (Public Readiness)

-- 1. Create Audit Logs table for persistent observability
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id TEXT, -- Can be a UUID or a string (for API Keys)
    action TEXT NOT NULL, -- e.g. "chat_complete", "tool_execution", "key_created"
    metadata JSONB DEFAULT '{}',
    request_id TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Enable RLS for audit_logs
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;

-- Policy: Users can view their own organization's audit logs
CREATE POLICY "Users can view their org audit logs"
    ON audit_logs FOR SELECT
    USING (
        EXISTS (
            SELECT 1 FROM members
            WHERE members.org_id = audit_logs.org_id
            AND members.user_id = auth.uid()
        )
    );

-- Index for performance
CREATE INDEX IF NOT EXISTS idx_audit_logs_org_created ON audit_logs(org_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_request_id ON audit_logs(request_id);

-- 2. Clean up api_keys table for hardening
-- We keep key_prefix for fast lookup, but we'll enforce key_hash being a bcrypt hash in logic.
-- Note: Existing keys will be functionally broken as they are plain text, 
-- but we already warned the user about this transition.

-- Update comment for clarity
COMMENT ON COLUMN api_keys.key_hash IS 'Bcrypt hash of the raw API key';

-- Optional: Ensure key_prefix index exists (from previous migration but just in case)
CREATE INDEX IF NOT EXISTS idx_api_keys_prefix_search ON api_keys(key_prefix);
