-- Migration: 027_remove_welcome_balance.sql
-- Remove the $10 welcome balance. API usage is now pay-as-you-go from $0.
-- Dashboard/app usage (JWT auth) remains unlimited (no billing check).

-- 1. Reset default to 0 for new organizations
ALTER TABLE organizations ALTER COLUMN token_balance SET DEFAULT 0.000000;

-- 2. Note: Existing orgs keep their current balance. To reset all:
-- UPDATE organizations SET token_balance = 0.000000 WHERE token_balance = 10.0000;
-- (Run manually if desired — not auto-applied to avoid disrupting active users)
