-- Migration: 015_persistent_cache.sql
-- Creates the persistent backup cache layer for Valkey (Redis) to prevent data loss on crashes or restarts.

CREATE TABLE IF NOT EXISTS persistent_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    cache_key TEXT NOT NULL UNIQUE,
    response TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL
);

-- Create an index to guarantee sub-millisecond lookups
CREATE INDEX IF NOT EXISTS idx_persistent_cache_key_expiry ON persistent_cache (cache_key, expires_at);

-- Table documentation
COMMENT ON TABLE persistent_cache IS 'Durable write-through cache records to survive Redis/Valkey service disruptions.';
