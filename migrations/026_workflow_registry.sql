-- Migration 026: Workflow Registry (Evolution Phase E4)
-- Stores workflow candidates discovered by O3 Workflow Miner and approved workflows.
-- Lifecycle: discovered → simulated → pending_review → approved → active → stale → demoted

CREATE TABLE IF NOT EXISTS workflow_candidates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL,
    pattern         TEXT NOT NULL,
    frequency       INT NOT NULL DEFAULT 0,
    tool_ids        TEXT[] NOT NULL,
    entity_schema   JSONB,                -- expected entity keys for this workflow
    status          TEXT NOT NULL DEFAULT 'discovered',
    -- Lifecycle: discovered → simulated → pending_review → approved → active → stale → demoted
    replay_accuracy FLOAT,
    replay_count    INT,
    last_hit_at     TIMESTAMPTZ,
    hit_count_7d    INT DEFAULT 0,
    accuracy_7d     FLOAT DEFAULT 1.0,
    tokens_saved    BIGINT DEFAULT 0,     -- estimated tokens saved by this workflow
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now(),
    reviewed_by     TEXT,
    reviewed_at     TIMESTAMPTZ,
    promoted_at     TIMESTAMPTZ,
    demoted_at      TIMESTAMPTZ,
    demotion_reason TEXT,                 -- 'frequency_drop' | 'accuracy_drop' | 'manual'
    CONSTRAINT valid_status CHECK (status IN (
        'discovered', 'simulated', 'pending_review', 'approved', 'active', 'stale', 'demoted'
    ))
);

-- Index for fast org-scoped lookups
CREATE INDEX IF NOT EXISTS idx_workflow_candidates_org_status
    ON workflow_candidates (org_id, status);

-- Index for demotion background job
CREATE INDEX IF NOT EXISTS idx_workflow_candidates_active_hits
    ON workflow_candidates (status, last_hit_at)
    WHERE status IN ('active', 'stale');

-- Unique constraint: one pattern per org
CREATE UNIQUE INDEX IF NOT EXISTS idx_workflow_candidates_org_pattern
    ON workflow_candidates (org_id, pattern)
    WHERE status NOT IN ('demoted');

-- Workflow execution log for accuracy tracking
CREATE TABLE IF NOT EXISTS workflow_executions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id     UUID NOT NULL REFERENCES workflow_candidates(id) ON DELETE CASCADE,
    org_id          UUID NOT NULL,
    prompt_hash     TEXT NOT NULL,
    entities        JSONB,
    success         BOOLEAN NOT NULL,
    latency_ms      INT,
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_workflow_executions_workflow
    ON workflow_executions (workflow_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_workflow_executions_org
    ON workflow_executions (org_id, created_at DESC);
