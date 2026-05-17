-- Migration: 014_api_key_scopes.sql
-- Adds granular RBAC scopes and key type/role for multi-token generations for users and agents.

-- Add scopes and role columns to api_keys table
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS scopes TEXT[] DEFAULT '{}';
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS role TEXT DEFAULT 'agent';

-- Update comments for database self-documentation
COMMENT ON COLUMN api_keys.scopes IS 'Array of granular permission scopes (e.g. tools:read, tools:write, chat:execute)';
COMMENT ON COLUMN api_keys.role IS 'The identity role of the token (e.g. user, agent, admin)';
