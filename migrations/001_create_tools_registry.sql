-- Migration: Create tools table for dynamic tool registry
-- This enables Phase 2: Dynamic Tool Registration without gateway restarts

CREATE TABLE IF NOT EXISTS tools (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    connector_type VARCHAR(50) NOT NULL DEFAULT 'mcp',
    endpoint VARCHAR(1024) NOT NULL,
    input_schema JSONB,
    output_schema JSONB,
    timeout_seconds INTEGER DEFAULT 15,
    enabled BOOLEAN DEFAULT true,
    requires_auth BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for fast lookups
CREATE INDEX IF NOT EXISTS idx_tools_enabled ON tools(enabled);
CREATE INDEX IF NOT EXISTS idx_tools_connector_type ON tools(connector_type) WHERE enabled = true;
CREATE INDEX IF NOT EXISTS idx_tools_created_at ON tools(created_at DESC);

-- Seed initial tools from MCP server (Phase 1 compatibility)
INSERT INTO tools (id, name, description, connector_type, endpoint, timeout_seconds, enabled, requires_auth)
VALUES 
    ('db_query', 'Database Query', 'Execute SQL queries and retrieve data', 'mcp', 'db_query', 15, true, true),
    ('get_user', 'Get User Information', 'Fetch user profile and metadata by user ID', 'mcp', 'get_user', 10, true, true),
    ('read_database', 'Read Database Metrics', 'Retrieve cluster metrics and database statistics', 'mcp', 'read_database', 20, true, true),
    ('aura_search', 'Neural Semantic Search', 'Perform semantic vector similarity search across knowledge base', 'mcp', 'aura_search', 15, true, false)
ON CONFLICT (id) DO NOTHING;
