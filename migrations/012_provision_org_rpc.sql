-- This function allows an authenticated user to provision their own default organization
-- securely if the auth trigger failed or if they are legacy accounts.

CREATE OR REPLACE FUNCTION public.provision_personal_org()
RETURNS UUID
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_user_id UUID;
    v_email TEXT;
    v_new_org_id UUID;
    v_display_name TEXT;
BEGIN
    v_user_id := auth.uid();
    IF v_user_id IS NULL THEN
        RAISE EXCEPTION 'Not authenticated';
    END IF;

    -- Check if they already have an org to prevent duplicates
    SELECT org_id INTO v_new_org_id
    FROM public.members
    WHERE user_id = v_user_id
    LIMIT 1;

    IF v_new_org_id IS NOT NULL THEN
        RETURN v_new_org_id;
    END IF;

    -- Get user email from auth.users (Security definer allows this)
    SELECT email INTO v_email
    FROM auth.users
    WHERE id = v_user_id;

    v_display_name := COALESCE(split_part(v_email, '@', 1), 'Personal');

    -- Create organization
    INSERT INTO public.organizations (name, slug, subscription_tier)
    VALUES (
        v_display_name || '''s Workspace',
        'workspace-' || substr(v_user_id::text, 1, 8) || '-' || substr(gen_random_uuid()::text, 1, 4),
        'free'
    )
    RETURNING id INTO v_new_org_id;

    -- Add member
    INSERT INTO public.members (org_id, user_id, role)
    VALUES (v_new_org_id, v_user_id, 'admin');

    -- Provision default tools
    INSERT INTO public.org_tools (org_id, tool_id)
    VALUES 
        (v_new_org_id, 'memzent_search'),
        (v_new_org_id, 'read_database')
    ON CONFLICT DO NOTHING;

    RETURN v_new_org_id;
END;
$$;
