-- Migration 022: Add configurable similarity threshold per organization
-- Allows orgs to tune semantic cache precision (default 0.88)
ALTER TABLE organizations
  ADD COLUMN IF NOT EXISTS similarity_threshold FLOAT DEFAULT 0.88;

COMMENT ON COLUMN organizations.similarity_threshold IS
  'Cosine similarity threshold for semantic cache hits (0.0-1.0). Higher = stricter matching.';
