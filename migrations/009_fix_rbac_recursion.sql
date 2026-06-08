-- Memzent Migration: Fix RBAC Recursion & Harden RLS
-- Resolves "infinite recursion detected in policy for relation 'members'"

-- 1. Create a helper function to bypass RLS recursion
-- This function runs as SECURITY DEFINER (using postgres privileges) 
-- to safely check roles without triggering the RLS policy on the same table.
CREATE OR REPLACE FUNCTION public.check_is_admin(org_id_param UUID)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM public.members
        WHERE org_id = org_id_param
        AND user_id = auth.uid()
        AND role IN ('admin', 'owner', 'platform_staff')
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 2. Update Members Policies
DROP POLICY IF EXISTS "Users can view their own membership" ON members;
DROP POLICY IF EXISTS "Admins can view all members in their org" ON members;

CREATE POLICY "Users can view their own membership"
ON members FOR SELECT
TO authenticated
USING (user_id = auth.uid());

CREATE POLICY "Admins can view entire org roster"
ON members FOR SELECT
TO authenticated
USING (public.check_is_admin(org_id));

-- 3. Update Organizations Policies
DROP POLICY IF EXISTS "Users can view organizations they are members of" ON organizations;
CREATE POLICY "Users can view their organizations"
ON organizations FOR SELECT
TO authenticated
USING (
    EXISTS (
        -- We check membership directly; since the user can always see their own row in 'members',
        -- this subquery will NOT recurse infinitely.
        SELECT 1 FROM members 
        WHERE members.org_id = organizations.id 
        AND members.user_id = auth.uid()
    )
);

-- 4. Update Org Tools Policies
DROP POLICY IF EXISTS "Admins can manage tools in their org" ON org_tools;
CREATE POLICY "Admins can manage tools"
ON org_tools FOR ALL
TO authenticated
USING (public.check_is_admin(org_id))
WITH CHECK (public.check_is_admin(org_id));

-- 5. Update API Keys Policies (Harden for Option A)
DROP POLICY IF EXISTS "Users can view their own personal keys" ON api_keys;
CREATE POLICY "Identified users vs Admin access for keys"
ON api_keys FOR SELECT
TO authenticated
USING (
    user_id = auth.uid() OR public.check_is_admin(org_id)
);
