-- Migration: 011 - Tool Registry Sync Enhancement
-- Phase 2: Adds org_id scoping and last_synced_at tracking to the tools table
-- This enables the Registry.Refresh() loop to identify drift between Postgres and Qdrant.

-- Add org_id for multi-tenant tool scoping (NULL = system/global tool)
ALTER TABLE tools ADD COLUMN IF NOT EXISTS org_id VARCHAR(255) DEFAULT NULL;

-- Add sync tracking for incremental refresh
ALTER TABLE tools ADD COLUMN IF NOT EXISTS last_synced_at TIMESTAMP DEFAULT NULL;

-- Add config JSONB for tool-specific connector configuration (connection strings, auth, etc.)
ALTER TABLE tools ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}'::jsonb;

-- Index to quickly find tools that have drifted (not yet synced to vector store)
CREATE INDEX IF NOT EXISTS idx_tools_sync_needed ON tools(last_synced_at) WHERE enabled = true;

-- Index for org-scoped tool lookups
CREATE INDEX IF NOT EXISTS idx_tools_org_id ON tools(org_id) WHERE enabled = true;

-- Update seed tools to mark them as needing initial sync
UPDATE tools SET last_synced_at = NULL WHERE last_synced_at IS NULL;
