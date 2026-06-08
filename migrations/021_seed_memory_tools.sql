-- Migration: 021_seed_memory_tools.sql
-- Description: Seed the store_memory and recall_memory native (core) tools in the registry table.

INSERT INTO tools (id, name, description, connector_type, endpoint, input_schema, enabled, requires_auth)
VALUES 
    (
        'store_memory', 
        'Store Long-Term Memory', 
        'Store a permanent fact or user preference in long-term memory', 
        'core', 
        'store_memory', 
        '{"type": "object", "properties": {"fact": {"type": "string"}}, "required": ["fact"]}'::jsonb, 
        true, 
        true
    ),
    (
        'recall_memory', 
        'Recall Long-Term Memory', 
        'Retrieve permanent facts or user preferences matching a query', 
        'core', 
        'recall_memory', 
        '{"type": "object", "properties": {"query": {"type": "string"}}, "required": ["query"]}'::jsonb, 
        true, 
        true
    )
ON CONFLICT (id) DO NOTHING;
