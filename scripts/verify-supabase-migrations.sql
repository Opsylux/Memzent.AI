-- Run in Supabase SQL Editor after applying pending migrations.
-- Expected: all rows show applied_at IS NOT NULL.

-- 1. Verify API key rotation columns (migration 020)
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_schema = 'public'
  AND table_name = 'api_keys'
  AND column_name IN ('expires_at', 'prev_key_hash', 'rotated_at', 'last_used_at')
ORDER BY column_name;

-- 2. Verify chat:execute provisioned for all orgs (migration 021)
SELECT COUNT(*) AS orgs_missing_chat_execute
FROM public.organizations o
WHERE NOT EXISTS (
    SELECT 1 FROM public.org_tools t
    WHERE t.org_id = o.id AND t.tool_id = 'chat:execute'
);

-- 3. Gateway auto-migration tracker (if gateway has connected)
SELECT version, applied_at
FROM schema_migrations
WHERE version IN ('020_api_key_rotation.sql', '021_provision_chat_execute.sql')
ORDER BY version;
