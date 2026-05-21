-- Migration: Add default provider and default model preferences to organizations
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS default_provider TEXT;
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS default_model TEXT;
