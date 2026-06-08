-- Migration: 020_api_key_rotation.sql
-- Adds key rotation, expiry TTL, and last_used_at tracking to api_keys.
-- This enables autonomous agents to rotate credentials without downtime
-- and enforces time-bounded key lifetimes.

-- 1. Expiry: Optional TTL. NULL = no expiry (legacy keys stay valid).
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ DEFAULT NULL;

-- 2. Rotation overlap: holds the PREVIOUS hash during the grace window (~15 min).
--    Both key_hash and prev_key_hash are accepted until prev_key_hash is cleared.
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS prev_key_hash TEXT DEFAULT NULL;

-- 3. Rotation audit timestamp.
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS rotated_at TIMESTAMPTZ DEFAULT NULL;

-- 4. last_used_at: Gateway updates this on every successful auth check.
--    Note: column already exists from migration 005 but was never written to by the gateway.
--    This migration is idempotent; the IF NOT EXISTS guard protects re-runs.
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMPTZ DEFAULT NULL;

-- Self-documentation
COMMENT ON COLUMN api_keys.expires_at     IS 'Optional expiry timestamp. NULL = key never expires. Enforced by gateway VerifyAPIKey.';
COMMENT ON COLUMN api_keys.prev_key_hash  IS 'Holds the previous bcrypt hash during a rotation grace window (15 min). Cleared by gateway after grace expires.';
COMMENT ON COLUMN api_keys.rotated_at     IS 'Timestamp of the last key rotation event.';
COMMENT ON COLUMN api_keys.last_used_at   IS 'Updated by gateway on every successful API key auth. Used for stale key detection.';

-- Performance: index for expiry sweeps and stale-key audits
CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at   ON api_keys(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_last_used_at ON api_keys(last_used_at);

-- Allow the gateway service role to update last_used_at and clear prev_key_hash
-- (RLS UPDATE policy: org members OR service role can update their own key metadata)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE tablename = 'api_keys' AND policyname = 'Service can update key metadata'
    ) THEN
        CREATE POLICY "Service can update key metadata"
            ON api_keys FOR UPDATE
            USING (
                -- Org admins can rotate/update their keys
                EXISTS (
                    SELECT 1 FROM members
                    WHERE members.org_id = api_keys.org_id
                    AND members.user_id = auth.uid()
                    AND members.role IN ('admin', 'owner', 'staff', 'platform_staff')
                )
            );
    END IF;
END $$;
