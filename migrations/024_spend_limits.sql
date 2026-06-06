-- Migration 024: Spend Limits & Budget Forecasting
-- Adds configurable daily/monthly spend limits per organization (both dollar and token caps)

-- Dollar-based caps
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS daily_spend_limit NUMERIC(15, 6) DEFAULT NULL;
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS monthly_spend_limit NUMERIC(15, 6) DEFAULT NULL;

-- Token-based caps (measured in total tokens consumed: input + output)
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS daily_token_limit BIGINT DEFAULT NULL;
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS monthly_token_limit BIGINT DEFAULT NULL;

-- Alert threshold (emit webhook at this % of any limit)
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS spend_alert_threshold NUMERIC(5, 2) DEFAULT 0.80;

-- Add provider + token usage columns to billing_ledger for structured analytics
ALTER TABLE billing_ledger ADD COLUMN IF NOT EXISTS provider TEXT;
ALTER TABLE billing_ledger ADD COLUMN IF NOT EXISTS tokens_used INT DEFAULT 0;

-- Index for efficient spend aggregation queries
CREATE INDEX IF NOT EXISTS idx_billing_ledger_org_date ON billing_ledger(org_id, created_at DESC) WHERE amount < 0;
