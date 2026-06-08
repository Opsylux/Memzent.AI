-- Provision chat:execute for all orgs so production RBAC (no dev bypass) still allows /v1/chat.
-- New sign-ups and manual provisioning both grant this scope via org_tools.

-- Backfill existing organizations
INSERT INTO public.org_tools (org_id, tool_id)
SELECT o.id, 'chat:execute'
FROM public.organizations o
WHERE NOT EXISTS (
    SELECT 1 FROM public.org_tools t
    WHERE t.org_id = o.id AND t.tool_id = 'chat:execute'
);

-- Update automatic sign-up provisioning
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
DECLARE
    new_org_id UUID;
    display_name TEXT;
BEGIN
    display_name := COALESCE(
        new.raw_user_meta_data->>'full_name',
        split_part(new.email, '@', 1),
        'User'
    );

    INSERT INTO public.organizations (name, slug, subscription_tier)
    VALUES (
        display_name || '''s Workspace',
        'workspace-' || substr(new.id::text, 1, 8) || '-' || substr(gen_random_uuid()::text, 1, 4),
        'free'
    )
    RETURNING id INTO new_org_id;

    INSERT INTO public.members (org_id, user_id, role)
    VALUES (new_org_id, new.id, 'admin');

    INSERT INTO public.org_tools (org_id, tool_id)
    VALUES
        (new_org_id, 'chat:execute'),
        (new_org_id, 'memzent_search'),
        (new_org_id, 'read_database')
    ON CONFLICT DO NOTHING;

    RETURN new;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Update manual provisioning RPC
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

    SELECT org_id INTO v_new_org_id
    FROM public.members
    WHERE user_id = v_user_id
    LIMIT 1;

    IF v_new_org_id IS NOT NULL THEN
        RETURN v_new_org_id;
    END IF;

    SELECT email INTO v_email
    FROM auth.users
    WHERE id = v_user_id;

    v_display_name := COALESCE(split_part(v_email, '@', 1), 'Personal');

    INSERT INTO public.organizations (name, slug, subscription_tier)
    VALUES (
        v_display_name || '''s Workspace',
        'workspace-' || substr(v_user_id::text, 1, 8) || '-' || substr(gen_random_uuid()::text, 1, 4),
        'free'
    )
    RETURNING id INTO v_new_org_id;

    INSERT INTO public.members (org_id, user_id, role)
    VALUES (v_new_org_id, v_user_id, 'admin');

    INSERT INTO public.org_tools (org_id, tool_id)
    VALUES
        (v_new_org_id, 'chat:execute'),
        (v_new_org_id, 'memzent_search'),
        (v_new_org_id, 'read_database')
    ON CONFLICT DO NOTHING;

    RETURN v_new_org_id;
END;
$$;
