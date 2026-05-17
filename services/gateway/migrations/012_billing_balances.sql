-- Add token_balance to organizations
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS token_balance NUMERIC(15, 4) DEFAULT 0.0000;

-- Create billing_ledger table to track deductions and top-ups
CREATE TABLE IF NOT EXISTS billing_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    amount NUMERIC(15, 4) NOT NULL, -- Negative for LLM usage, Positive for top-ups
    transaction_type TEXT NOT NULL, -- 'llm_usage', 'cache_hit', 'stripe_topup', 'grant'
    description TEXT,
    metadata JSONB, -- For storing model, provider, tokens used, or stripe session ID
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Index for querying billing history
CREATE INDEX IF NOT EXISTS idx_billing_ledger_org_id ON billing_ledger(org_id);
CREATE INDEX IF NOT EXISTS idx_billing_ledger_created_at ON billing_ledger(created_at DESC);
