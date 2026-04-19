-- Aura Migration: User-Scoped API Keys (Option A)
-- Adds individual ownership while maintaining admin oversight.

-- 1. Add user_id column
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES auth.users(id);

-- 2. Backfill existing keys (Optional: Assign to the first admin found in their org)
UPDATE api_keys
SET user_id = (
    SELECT user_id FROM members 
    WHERE members.org_id = api_keys.org_id 
    AND members.role = 'admin' 
    LIMIT 1
)
WHERE user_id IS NULL;

-- 3. Update RLS Policies
DROP POLICY IF EXISTS "Users can view their own org keys" ON api_keys;
CREATE POLICY "Users can view their own personal keys"
    ON api_keys FOR SELECT
    TO authenticated
    USING (
        -- User owns the key
        user_id = auth.uid() OR
        -- OR User is an admin of the org that owns the key
        EXISTS (
            SELECT 1 FROM members
            WHERE members.org_id = api_keys.org_id
            AND members.user_id = auth.uid()
            AND members.role IN ('admin', 'owner', 'platform_staff')
        )
    );

DROP POLICY IF EXISTS "Admins can create keys" ON api_keys;
CREATE POLICY "Any member can create keys in their org"
    ON api_keys FOR INSERT
    TO authenticated
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM members
            WHERE members.org_id = api_keys.org_id
            AND members.user_id = auth.uid()
        )
    );

DROP POLICY IF EXISTS "Admins can revoke keys" ON api_keys;
CREATE POLICY "Users can revoke their own keys or admins can revoke any"
    ON api_keys FOR DELETE
    TO authenticated
    USING (
        user_id = auth.uid() OR
        EXISTS (
            SELECT 1 FROM members
            WHERE members.org_id = api_keys.org_id
            AND members.user_id = auth.uid()
            AND members.role IN ('admin', 'owner', 'platform_staff')
        )
    );
