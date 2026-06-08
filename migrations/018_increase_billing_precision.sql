-- Alter token_balance and amount to use 6 decimal places for micro-cent precision
ALTER TABLE organizations ALTER COLUMN token_balance TYPE NUMERIC(15, 6);
ALTER TABLE billing_ledger ALTER COLUMN amount TYPE NUMERIC(15, 6);
