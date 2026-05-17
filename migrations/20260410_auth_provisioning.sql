-- Automatic User & Organization Provisioning
-- Every new sign-up gets a "Personal" workspace and default tool access.

-- 1. Function to handle new user provisioning
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
DECLARE
    new_org_id UUID;
    display_name TEXT;
BEGIN
    -- Determine display name (prefer full_name from metadata, then email, then 'User')
    display_name := COALESCE(
        new.raw_user_meta_data->>'full_name',
        split_part(new.email, '@', 1),
        'User'
    );

    -- Create a default organization for the user
    INSERT INTO public.organizations (name, slug, subscription_tier)
    VALUES (
        display_name || '''s Workspace',
        'workspace-' || substr(new.id::text, 1, 8) || '-' || substr(gen_random_uuid()::text, 1, 4),
        'free'
    )
    RETURNING id INTO new_org_id;

    -- Add the user as the ADMIN of this organization
    INSERT INTO public.members (org_id, user_id, role)
    VALUES (new_org_id, new.id, 'admin');

    -- Provision default tools for the new organization
    -- These IDs must match the ones in public.tool_registry
    INSERT INTO public.org_tools (org_id, tool_id)
    VALUES 
        (new_org_id, 'memzent_search'),
        (new_org_id, 'read_database')
    ON CONFLICT DO NOTHING;

    RETURN new;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 2. Trigger on auth.users (Supabase Auth table)
-- Note: In Supabase, tokens for 'auth' table are restricted, so we use the 'auth' schema explicitly
-- if running as superuser, or target the trigger correctly. 
-- In most hosted environments, you apply this to auth.users.
DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
CREATE TRIGGER on_auth_user_created
    AFTER INSERT ON auth.users
    FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();
