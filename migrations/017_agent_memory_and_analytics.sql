-- Migration: 017_agent_memory_and_analytics.sql
-- Description: Provision tables for chat sessions, persistent message history, and tool execution telemetry with row-level security.

-- 1. Create chat_sessions table
CREATE TABLE IF NOT EXISTS chat_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id TEXT, -- Can represent user UUID or API key identifier
    title TEXT NOT NULL DEFAULT 'New Conversation',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- 2. Create chat_messages table
CREATE TABLE IF NOT EXISTS chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- 3. Create tool_executions table for detailed latency and ROI analytics
CREATE TABLE IF NOT EXISTS tool_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    tool_id TEXT NOT NULL,
    request_id TEXT,
    duration_ms INTEGER NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('success', 'failure')),
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- --- Enable Row-Level Security (RLS) ---
ALTER TABLE chat_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE chat_messages ENABLE ROW LEVEL SECURITY;
ALTER TABLE tool_executions ENABLE ROW LEVEL SECURITY;

-- --- RLS Policies for chat_sessions ---
CREATE POLICY "Users can view their org chat sessions" ON chat_sessions 
    FOR SELECT 
    USING (
        EXISTS (
            SELECT 1 FROM members
            WHERE members.org_id = chat_sessions.org_id
            AND members.user_id = auth.uid()
        )
    );

CREATE POLICY "Users can create their org chat sessions" ON chat_sessions 
    FOR INSERT 
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM members
            WHERE members.org_id = chat_sessions.org_id
            AND members.user_id = auth.uid()
        )
    );

CREATE POLICY "Users can update their org chat sessions" ON chat_sessions 
    FOR UPDATE 
    USING (
        EXISTS (
            SELECT 1 FROM members
            WHERE members.org_id = chat_sessions.org_id
            AND members.user_id = auth.uid()
        )
    );

CREATE POLICY "Users can delete their org chat sessions" ON chat_sessions 
    FOR DELETE 
    USING (
        EXISTS (
            SELECT 1 FROM members
            WHERE members.org_id = chat_sessions.org_id
            AND members.user_id = auth.uid()
        )
    );

-- --- RLS Policies for chat_messages ---
CREATE POLICY "Users can view messages in their sessions" ON chat_messages 
    FOR SELECT 
    USING (
        EXISTS (
            SELECT 1 FROM chat_sessions s
            JOIN members m ON s.org_id = m.org_id
            WHERE s.id = chat_messages.session_id
            AND m.user_id = auth.uid()
        )
    );

CREATE POLICY "Users can insert messages in their sessions" ON chat_messages 
    FOR INSERT 
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM chat_sessions s
            JOIN members m ON s.org_id = m.org_id
            WHERE s.id = chat_messages.session_id
            AND m.user_id = auth.uid()
        )
    );

-- --- RLS Policies for tool_executions ---
CREATE POLICY "Users can view their org tool executions" ON tool_executions 
    FOR SELECT 
    USING (
        EXISTS (
            SELECT 1 FROM members
            WHERE members.org_id = tool_executions.org_id
            AND members.user_id = auth.uid()
        )
    );

-- --- Performance Indexes ---
CREATE INDEX IF NOT EXISTS idx_chat_sessions_org_updated ON chat_sessions(org_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_chat_messages_session ON chat_messages(session_id, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_tool_executions_org ON tool_executions(org_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tool_executions_tool ON tool_executions(tool_id);

-- Documentation
COMMENT ON TABLE chat_sessions IS 'Stores metadata for specific interactive chat threads.';
COMMENT ON TABLE chat_messages IS 'Short-term persistent context of conversation histories within a chat session.';
COMMENT ON TABLE tool_executions IS 'Latency, ROI, and status analytics of all native and external tool calls.';
