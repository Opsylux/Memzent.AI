-- Memzent API Keys Table
-- Provides hashed storage for external agent access tokens

CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    key_prefix TEXT NOT NULL, -- First 8 characters for display
    key_hash TEXT NOT NULL UNIQUE, -- The full key (stored as plain text for now, should be hashed in production)
    created_at TIMESTAMPTZ DEFAULT now(),
    last_used_at TIMESTAMPTZ
);

-- Enable RLS
ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;

-- Policies
CREATE POLICY "Users can view their own org keys"
    ON api_keys FOR SELECT
    USING (
        EXISTS (
            SELECT 1 FROM members
            WHERE members.org_id = api_keys.org_id
            AND members.user_id = auth.uid()
        )
    );

CREATE POLICY "Admins can create keys"
    ON api_keys FOR INSERT
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM members
            WHERE members.org_id = api_keys.org_id
            AND members.user_id = auth.uid()
            AND members.role IN ('admin', 'owner', 'staff', 'platform_staff')
        )
    );

CREATE POLICY "Admins can revoke keys"
    ON api_keys FOR DELETE
    USING (
        EXISTS (
            SELECT 1 FROM members
            WHERE members.org_id = api_keys.org_id
            AND members.user_id = auth.uid()
            AND members.role IN ('admin', 'owner', 'staff', 'platform_staff')
        )
    );

-- Index for Gateway lookup performance
CREATE INDEX IF NOT EXISTS idx_api_keys_org ON api_keys(org_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_prefix ON api_keys(key_prefix);
