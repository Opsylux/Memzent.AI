-- Migration 024: Spend Limits & Budget Forecasting
-- Adds configurable daily/monthly spend limits per organization

ALTER TABLE organizations ADD COLUMN IF NOT EXISTS daily_spend_limit NUMERIC(15, 6) DEFAULT NULL;
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS monthly_spend_limit NUMERIC(15, 6) DEFAULT NULL;
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS spend_alert_threshold NUMERIC(5, 2) DEFAULT 0.80;  -- Alert at 80% of limit

-- Add provider column to billing_ledger for structured analytics
ALTER TABLE billing_ledger ADD COLUMN IF NOT EXISTS provider TEXT;

-- Index for efficient spend aggregation queries
CREATE INDEX IF NOT EXISTS idx_billing_ledger_org_date ON billing_ledger(org_id, created_at DESC) WHERE amount < 0;
