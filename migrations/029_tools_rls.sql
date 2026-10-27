-- Migration: Enable RLS on `tools`, and add a missing UPDATE policy on
-- `organizations`.
--
-- 1. `tools` (created in 001_create_tools_registry.sql, org_id column added
--    in 003/010/011) never had row-level security enabled. It is queried
--    directly from Supabase in the dashboard
--    (services/dashboard/src/app/actions.ts getOrgTools/createMemzentTool)
--    using a caller-supplied org_id, so without RLS any authenticated user
--    could read or write another organization's tool rows directly. This
--    mirrors the org_tools/organizations policies from
--    007_core_rls_policies.sql / 009_fix_rbac_recursion.sql, reusing the
--    `check_is_admin()` helper for consistency.
--
-- 2. `organizations` never had an UPDATE policy, so
--    actions.ts:updateOrgProfile() has been failing under RLS for every
--    caller (permission denied). Add an admin-only UPDATE policy.

ALTER TABLE tools ENABLE ROW LEVEL SECURITY;

-- Org-owned tools are visible to members of that org; system-wide tools
-- (org_id IS NULL) remain visible to everyone, matching the Go gateway's
-- `WHERE org_id = $1 OR org_id IS NULL` visibility rule.
CREATE POLICY "Users can view tools in their org or global tools"
ON tools FOR SELECT
TO authenticated
USING (
    org_id IS NULL
    OR EXISTS (
        SELECT 1 FROM members
        WHERE members.org_id = tools.org_id
        AND members.user_id = auth.uid()
    )
);

-- Only org admins may create/update/delete tools owned by their own org.
-- System-wide tools (org_id IS NULL) are excluded from self-service
-- mutation and remain managed out-of-band.
CREATE POLICY "Admins can manage tools in their org"
ON tools FOR ALL
TO authenticated
USING (org_id IS NOT NULL AND public.check_is_admin(org_id))
WITH CHECK (org_id IS NOT NULL AND public.check_is_admin(org_id));

-- Admins may update their own organization's profile fields.
CREATE POLICY "Admins can update their organization"
ON organizations FOR UPDATE
TO authenticated
USING (public.check_is_admin(id))
WITH CHECK (public.check_is_admin(id));
