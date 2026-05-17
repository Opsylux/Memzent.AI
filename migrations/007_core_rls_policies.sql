-- Memzent Core Infrastructure Security - RLS Policies
-- This completes the security link for multi-tenant isolation.

-- 1. Organizations
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view organizations they are members of"
ON organizations FOR SELECT
TO authenticated
USING (
    EXISTS (
        SELECT 1 FROM members 
        WHERE members.org_id = organizations.id 
        AND members.user_id = auth.uid()
    )
);

-- 2. Members
ALTER TABLE members ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view their own membership"
ON members FOR SELECT
TO authenticated
USING (user_id = auth.uid());

CREATE POLICY "Admins can view all members in their org"
ON members FOR SELECT
TO authenticated
USING (
    EXISTS (
        SELECT 1 FROM members m2
        WHERE m2.org_id = members.org_id
        AND m2.user_id = auth.uid()
        AND m2.role = 'admin'
    )
);

-- 3. Org Tools
ALTER TABLE org_tools ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view tools in their org"
ON org_tools FOR SELECT
TO authenticated
USING (
    EXISTS (
        SELECT 1 FROM members
        WHERE members.org_id = org_tools.org_id
        AND members.user_id = auth.uid()
    )
);

CREATE POLICY "Admins can manage tools in their org"
ON org_tools FOR ALL -- Grants INSERT, UPDATE, DELETE
TO authenticated
USING (
    EXISTS (
        SELECT 1 FROM members
        WHERE members.org_id = org_tools.org_id
        AND members.user_id = auth.uid()
        AND members.role = 'admin'
    )
)
WITH CHECK (
    EXISTS (
        SELECT 1 FROM members
        WHERE members.org_id = org_tools.org_id
        AND members.user_id = auth.uid()
        AND members.role = 'admin'
    )
);

-- 4. Tool Registry (Global Table)
ALTER TABLE tool_registry ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Anyone can view global tool definitions"
ON tool_registry FOR SELECT
TO authenticated
USING (true);
