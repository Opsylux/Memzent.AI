"use server"

import { createClient } from '@/lib/supabase-server'

const GATEWAY_URL = process.env.GATEWAY_INTERNAL_URL || process.env.NEXT_PUBLIC_GATEWAY_URL || 'http://localhost:8080';

/**
 * Build standard headers for Gateway calls.
 * Injects org_id from Supabase session so the Gateway can scope responses.
 */
async function gatewayHeaders(orgId?: string): Promise<Record<string, string>> {
    const headers: Record<string, string> = {
        "Content-Type": "application/json",
    }

    // If orgId is explicitly provided, use it
    if (orgId) {
        headers["X-Org-ID"] = orgId
    }

    // Try to get a JWT from the current session for Gateway auth
    try {
        const supabase = await createClient()
        const { data: { session } } = await supabase.auth.getSession()
        if (session?.access_token) {
            headers["Authorization"] = `Bearer ${session.access_token}`
        }
    } catch {
        // Fallback: no session available (e.g. during build)
    }

    return headers
}

// ─── Gateway API ───────────────────────────────────────────────────────────

export async function getAuraTools(orgId?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/tools`, { 
            cache: 'no-store',
            headers,
        });
        if (!res.ok) return [];
        return res.json();
    } catch (e) {
        console.error("Gateway fetch failed", e);
        return [];
    }
}

export async function getAuraStats(orgId?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/stats`, { 
            cache: 'no-store',
            headers,
        });
        if (!res.ok) return { total_requests: 0, cache_hits: 0, uptime_seconds: 0 };
        return res.json();
    } catch (e) {
        console.error("Gateway stats fetch failed", e);
        return { total_requests: 0, cache_hits: 0, uptime_seconds: 0 };
    }
}

export async function executeAuraPrompt(prompt: string, orgId?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/chat`, {
            method: "POST",
            headers,
            body: JSON.stringify({
                prompt: prompt,
            }),
            cache: 'no-store'
        });

        if (!res.ok) {
            const err = await res.text();
            throw new Error(err || "Failed to execute prompt");
        }

        return res.json();
    } catch (e: any) {
        console.error("Gateway prompt execution failed", e);
        throw new Error(e.message);
    }
}

// ─── Supabase Data (Org-Scoped) ────────────────────────────────────────────

// ─── Supabase Data (Org-Scoped) ────────────────────────────────────────────

export async function getApiKeys(orgId: string) {
    const supabase = await createClient();
    const { data, error } = await supabase
        .from('api_keys')
        .select('*')
        .eq('org_id', orgId)
        .order('created_at', { ascending: false });

    if (error) throw error;
    return data;
}

export async function createApiKey(orgId: string, name: string) {
    const supabase = await createClient();
    const key = `aura_${Math.random().toString(36).substring(2, 15)}${Math.random().toString(36).substring(2, 15)}`;
    const prefix = key.substring(0, 8);
    
    const { error } = await supabase
        .from('api_keys')
        .insert({
            org_id: orgId,
            name: name,
            key_prefix: prefix,
            key_hash: key
        });

    if (error) throw error;
    return { key };
}

export async function revokeApiKey(id: string) {
    const supabase = await createClient();
    const { error } = await supabase
        .from('api_keys')
        .delete()
        .eq('id', id);

    if (error) throw error;
}

export async function createAuraTool(orgId: string, tool: any) {
    const supabase = await createClient();
    const { error } = await supabase
        .from('tools')
        .insert({
            id: tool.id,
            org_id: orgId,
            name: tool.name,
            description: tool.description,
            connector_type: tool.connector_type,
            endpoint: tool.endpoint,
            config: tool.config || {},
            input_schema: tool.input_schema || {},
            enabled: true
        });

    if (error) throw error;
    return { success: true };
}

export async function getOrgTools(orgId: string) {
    const supabase = await createClient();
    const { data, error } = await supabase
        .from('tools')
        .select('*')
        .or(`org_id.eq.${orgId},org_id.is.null`)
        .order('created_at', { ascending: false });

    if (error) throw error;
    return data;
}

export async function updateOrgProfile(orgId: string, updates: { name?: string; contact_email?: string }) {
    const supabase = await createClient();
    const { error } = await supabase
        .from('organizations')
        .update(updates)
        .eq('id', orgId)

    if (error) throw error;
    return { success: true };
}

export async function getOrgProfile(orgId: string) {
    const supabase = await createClient();
    const { data, error } = await supabase
        .from('organizations')
        .select('*')
        .eq('id', orgId)
        .maybeSingle()

    if (error) throw error;
    return data;
}

export async function registerAuraTool(orgId: string, tool: any) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/tools/register`, {
            method: "POST",
            headers,
            body: JSON.stringify({
                ...tool,
                org_id: orgId
            }),
            cache: 'no-store'
        });

        if (!res.ok) {
            const err = await res.text();
            throw new Error(err || "Failed to register tool");
        }

        return res.json();
    } catch (e: any) {
        console.error("Gateway tool registration failed", e);
        throw new Error(e.message);
    }
}

export async function syncAuraTools(orgId?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/tools/sync`, {
            method: "POST",
            headers,
            cache: 'no-store'
        });

        if (!res.ok) {
            const err = await res.text();
            throw new Error(err || "Failed to sync tools");
        }

        return res.json();
    } catch (e: any) {
        console.error("Gateway tools sync failed", e);
        throw new Error(e.message);
    }
}

export async function getAuraAudit(orgId?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/audit`, {
            method: "GET",
            headers,
            cache: 'no-store'
        });

        if (!res.ok) {
            const err = await res.text();
            throw new Error(err || "Failed to fetch audit logs");
        }

        return res.json();
    } catch (e: any) {
        console.error("Gateway audit fetch failed", e);
        return [];
    }
}

export async function createCheckoutSession(orgId: string, tier: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/billing/checkout`, {
            method: "POST",
            headers,
            body: JSON.stringify({ tier }),
            cache: 'no-store'
        });

        if (!res.ok) {
            const err = await res.text();
            throw new Error(err || "Failed to create checkout session");
        }

        return res.json();
    } catch (e: any) {
        console.error("Gateway checkout failed", e);
        throw new Error(e.message);
    }
}