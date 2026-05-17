-- Migration: 016_welcome_token_balance.sql
-- Allocates a complimentary $10 welcome token balance (10.0000 tokens) to all new and existing organizations to ensure trial accounts are active.

-- 1. Alter default column value for future organization inserts to be $10
ALTER TABLE organizations ALTER COLUMN token_balance SET DEFAULT 10.0000;

-- 2. Backfill existing trial organizations that are depleted (0.0000)
UPDATE organizations 
SET token_balance = 10.0000 
WHERE token_balance = 0.0000 OR token_balance IS NULL;

-- 3. Record complimentary grant transactions in billing_ledger for audit completeness
INSERT INTO billing_ledger (org_id, amount, transaction_type, description)
SELECT id, 10.0000, 'grant', 'Complimentary Welcome Balance'
FROM organizations
WHERE id NOT IN (SELECT DISTINCT org_id FROM billing_ledger WHERE transaction_type = 'grant');
