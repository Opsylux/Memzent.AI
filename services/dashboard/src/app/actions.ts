"use server"

import { createClient } from '@/lib/supabase-server'
import { redirect } from 'next/navigation'
import { plans } from '@/app/plans'
import bcrypt from 'bcryptjs'
import crypto from 'crypto'

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

export async function getMemzentTools(orgId?: string) {
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

export async function getMemzentStats(orgId?: string) {
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

export async function getMemzentProviders() {
    try {
        const headers = await gatewayHeaders()
        const res = await fetch(`${GATEWAY_URL}/v1/providers`, {
            cache: 'no-store',
            headers,
        });
        if (!res.ok) return [];
        return res.json();
    } catch (e) {
        console.error("Gateway providers fetch failed", e);
        return [];
    }
}

export async function executeMemzentPrompt(messages: {role: string, content: string}[], sessionId?: string, orgId?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        headers["X-Request-ID"] = crypto.randomUUID()
        const res = await fetch(`${GATEWAY_URL}/v1/chat`, {
            method: "POST",
            headers,
            body: JSON.stringify({
                messages: messages,
                session_id: sessionId,
            }),
            cache: 'no-store'
        });

        if (!res.ok) {
            const body = await res.text();
            try {
                const parsed = JSON.parse(body);
                throw new Error(parsed.error || body);
            } catch {
                throw new Error(body || "Failed to execute prompt");
            }
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

export async function getOrgKeyCount(orgId: string) {
    const supabase = await createClient();
    const { count, error } = await supabase
        .from('api_keys')
        .select('*', { count: 'exact', head: true })
        .eq('org_id', orgId);

    if (error) return 0;
    return count || 0;
}

export async function getOrgAuditStats(orgId: string) {
    const supabase = await createClient();
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);

    const { count, error } = await supabase
        .from('audit_logs')
        .select('*', { count: 'exact', head: true })
        .eq('org_id', orgId)
        .gte('created_at', yesterday.toISOString());

    if (error) return { count24h: 0 };
    return { count24h: count || 0 };
}

export async function createApiKey(orgId: string, name: string, scopes: string[] = [], role: string = 'agent', expiresAt: string | null = null) {
    const supabase = await createClient();

    // Generate a high-entropy 32-char raw key
    const rawKey = `memzent_${crypto.randomBytes(24).toString('hex')}`;
    const prefix = rawKey.substring(0, 16);

    // Hash the key for secure storage
    const salt = await bcrypt.genSalt(10);
    const hash = await bcrypt.hash(rawKey, salt);

    // Get current user to set as the owner
    const { data: { user } } = await supabase.auth.getUser();
    if (!user) throw new Error("Unauthorized: No user session found");

    const insertPayload: Record<string, any> = {
        org_id: orgId,
        user_id: user.id,
        name: name,
        key_prefix: prefix,
        key_hash: hash,
        scopes: scopes,
        role: role,
    };

    // Only set expires_at if a TTL was selected
    if (expiresAt) {
        insertPayload.expires_at = expiresAt;
    }

    const { data, error } = await supabase
        .from('api_keys')
        .insert(insertPayload)
        .select();

    if (error) {
        console.error("Supabase insert error in createApiKey:", error);
        throw error;
    }

    // Return the RAW key to the user - this is their ONLY chance to see it!
    return { key: rawKey };
}


export async function revokeApiKey(id: string) {
    const supabase = await createClient();
    const { error } = await supabase
        .from('api_keys')
        .delete()
        .eq('id', id);

    if (error) throw error;
}

/**
 * Rotates an API key in place.
 * - Generates a new high-entropy key.
 * - Moves the current key_hash → prev_key_hash (gateway accepts both for 15 min grace).
 * - Stores the new key_hash and stamps rotated_at.
 * - Returns the new raw key — this is the ONLY time it will be shown.
 */
export async function rotateApiKey(id: string) {
    const supabase = await createClient();

    // Verify the key belongs to the authenticated user's org before rotating
    const { data: { user } } = await supabase.auth.getUser();
    if (!user) throw new Error("Unauthorized");

    const { data: existing, error: fetchErr } = await supabase
        .from('api_keys')
        .select('key_hash, org_id')
        .eq('id', id)
        .maybeSingle();

    if (fetchErr || !existing) throw new Error("Key not found or unauthorized");

    // Generate a new key
    const newRawKey = `memzent_${crypto.randomBytes(24).toString('hex')}`;
    const newPrefix  = newRawKey.substring(0, 16);
    const salt       = await bcrypt.genSalt(10);
    const newHash    = await bcrypt.hash(newRawKey, salt);

    const { error: updateErr } = await supabase
        .from('api_keys')
        .update({
            key_prefix:    newPrefix,
            key_hash:      newHash,
            prev_key_hash: existing.key_hash,  // grace window — gateway accepts old key for 15 min
            rotated_at:    new Date().toISOString(),
        })
        .eq('id', id);

    if (updateErr) {
        console.error("rotateApiKey update error:", updateErr);
        throw updateErr;
    }

    // Return the new raw key — ONLY time it will be visible
    return { key: newRawKey };
}

export async function createMemzentTool(orgId: string, tool: any) {
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

    // Immediately trigger a background sync to vectorize this new tool in Qdrant with zero delay!
    try {
        await syncMemzentTools(orgId);
    } catch (e) {
        console.error("Immediate background tool sync failed", e);
    }

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

export async function signOutAction() {
    const supabase = await createClient()
    await supabase.auth.signOut()
    redirect('/login')
}

export async function updateOrgProfile(orgId: string, updates: { name?: string; contact_email?: string; default_provider?: string | null; default_model?: string | null }) {
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

// ─── Similarity Threshold ──────────────────────────────────────────────────

export async function getSimilarityThreshold(orgId: string): Promise<number> {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/settings/threshold`, {
            cache: 'no-store',
            headers,
        });
        if (!res.ok) return 0.88;
        const data = await res.json();
        return data.similarity_threshold ?? 0.88;
    } catch {
        return 0.88;
    }
}

export async function updateSimilarityThreshold(orgId: string, threshold: number) {
    const headers = await gatewayHeaders(orgId)
    const res = await fetch(`${GATEWAY_URL}/v1/settings/threshold`, {
        method: 'PUT',
        headers,
        body: JSON.stringify({ similarity_threshold: threshold }),
        cache: 'no-store',
    });
    if (!res.ok) {
        const err = await res.text();
        throw new Error(err || 'Failed to update threshold');
    }
    return res.json();
}

export async function registerMemzentTool(orgId: string, tool: any) {
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

export async function syncMemzentTools(orgId?: string) {
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

export async function getMemzentAudit(orgId?: string) {
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

export async function createCheckoutSession(orgId: string, payload: { tier?: string, amount?: number }) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/billing/checkout`, {
            method: "POST",
            headers,
            body: JSON.stringify(payload),
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

// ─── Sessions API ──────────────────────────────────────────────────────────

export async function getSessions(orgId?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/sessions`, {
            method: "GET",
            headers,
            cache: 'no-store'
        });

        if (!res.ok) return [];
        return res.json();
    } catch (e) {
        console.error("Gateway sessions fetch failed", e);
        return [];
    }
}

export async function createSession(orgId: string, title?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/sessions`, {
            method: "POST",
            headers,
            body: JSON.stringify({ title: title || "New Conversation" }),
            cache: 'no-store'
        });

        if (!res.ok) {
            const err = await res.text();
            throw new Error(err || "Failed to create session");
        }

        return res.json();
    } catch (e: any) {
        console.error("Gateway session creation failed", e);
        throw new Error(e.message);
    }
}

export async function getSessionMessages(sessionId: string, orgId?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/sessions/${sessionId}/messages`, {
            method: "GET",
            headers,
            cache: 'no-store'
        });

        if (!res.ok) return [];
        return res.json();
    } catch (e) {
        console.error("Gateway session messages fetch failed", e);
        return [];
    }
}

export async function deleteSession(sessionId: string, orgId?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/sessions/${sessionId}`, {
            method: "DELETE",
            headers,
            cache: 'no-store'
        });

        if (!res.ok) {
            const err = await res.text();
            throw new Error(err || "Failed to delete session");
        }

        return res.json();
    } catch (e: any) {
        console.error("Gateway session deletion failed", e);
        throw new Error(e.message);
    }
}

// ─── Context Analytics API ─────────────────────────────────────────────────

export async function getContextAnalytics(orgId?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/analytics/context`, {
            method: "GET",
            headers,
            cache: 'no-store'
        });

        if (!res.ok) {
            return {
                tool_metrics: [],
                savings_roi: { cache_hits: 0, estimated_saved: 0, llm_cost: 0, net_roi: 0 },
                semantic_clusters: []
            };
        }

        return res.json();
    } catch (e) {
        console.error("Gateway context analytics fetch failed", e);
        return {
            tool_metrics: [],
            savings_roi: { cache_hits: 0, estimated_saved: 0, llm_cost: 0, net_roi: 0 },
            semantic_clusters: []
        };
    }
}

// ─── Tool CRUD API ─────────────────────────────────────────────────────────

export async function getTool(toolId: string, orgId?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/tools/${toolId}`, {
            method: "GET",
            headers,
            cache: 'no-store'
        });
        if (!res.ok) return null;
        return res.json();
    } catch (e) {
        console.error("Gateway tool fetch failed", e);
        return null;
    }
}

export async function updateTool(toolId: string, data: Record<string, any>, orgId?: string) {
    const headers = await gatewayHeaders(orgId)
    const res = await fetch(`${GATEWAY_URL}/v1/tools/${toolId}`, {
        method: "PUT",
        headers,
        body: JSON.stringify(data),
        cache: 'no-store'
    });
    if (!res.ok) {
        const err = await res.text();
        throw new Error(err || "Failed to update tool");
    }
    return res.json();
}

export async function deleteTool(toolId: string, orgId?: string) {
    const headers = await gatewayHeaders(orgId)
    const res = await fetch(`${GATEWAY_URL}/v1/tools/${toolId}`, {
        method: "DELETE",
        headers,
        cache: 'no-store'
    });
    if (!res.ok) {
        const err = await res.text();
        throw new Error(err || "Failed to delete tool");
    }
    return res.json();
}

// ─── Budget & Spend API ────────────────────────────────────────────────────

export async function getBudgetStatus(orgId?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/billing/budget`, {
            method: "GET",
            headers,
            cache: 'no-store'
        });
        if (!res.ok) return null;
        return res.json();
    } catch (e) {
        console.error("Gateway budget status fetch failed", e);
        return null;
    }
}

export async function getSpendTimeseries(days: number = 30, orgId?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/billing/spend-timeseries?days=${days}`, {
            method: "GET",
            headers,
            cache: 'no-store'
        });
        if (!res.ok) return [];
        return res.json();
    } catch (e) {
        console.error("Gateway spend timeseries fetch failed", e);
        return [];
    }
}

export async function getSpendLimits(orgId?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/billing/spend-limits`, {
            method: "GET",
            headers,
            cache: 'no-store'
        });
        if (!res.ok) return null;
        return res.json();
    } catch (e) {
        console.error("Gateway spend limits fetch failed", e);
        return null;
    }
}

export async function setSpendLimits(data: { daily_limit?: number | null; monthly_limit?: number | null; daily_token_limit?: number | null; monthly_token_limit?: number | null }, orgId?: string) {
    const headers = await gatewayHeaders(orgId)
    const res = await fetch(`${GATEWAY_URL}/v1/billing/spend-limits`, {
        method: "PUT",
        headers,
        body: JSON.stringify(data),
        cache: 'no-store'
    });
    if (!res.ok) {
        const err = await res.text();
        throw new Error(err || "Failed to set spend limits");
    }
    return res.json();
}

// ─── Webhooks API (Phase 7) ────────────────────────────────────────────────

export async function getWebhooks(orgId?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/webhooks`, {
            method: "GET",
            headers,
            cache: 'no-store'
        });
        if (!res.ok) return [];
        return res.json();
    } catch (e) {
        console.error("Gateway webhooks fetch failed", e);
        return [];
    }
}

export async function createWebhook(data: { url: string; events: string[]; description?: string }, orgId?: string) {
    const headers = await gatewayHeaders(orgId)
    const res = await fetch(`${GATEWAY_URL}/v1/webhooks`, {
        method: "POST",
        headers,
        body: JSON.stringify(data),
        cache: 'no-store'
    });
    if (!res.ok) {
        const err = await res.text();
        throw new Error(err || "Failed to create webhook");
    }
    return res.json();
}

export async function updateWebhook(webhookId: string, data: Record<string, any>, orgId?: string) {
    const headers = await gatewayHeaders(orgId)
    const res = await fetch(`${GATEWAY_URL}/v1/webhooks/${webhookId}`, {
        method: "PUT",
        headers,
        body: JSON.stringify(data),
        cache: 'no-store'
    });
    if (!res.ok) {
        const err = await res.text();
        throw new Error(err || "Failed to update webhook");
    }
    return res.json();
}

export async function deleteWebhook(webhookId: string, orgId?: string) {
    const headers = await gatewayHeaders(orgId)
    const res = await fetch(`${GATEWAY_URL}/v1/webhooks/${webhookId}`, {
        method: "DELETE",
        headers,
        cache: 'no-store'
    });
    if (!res.ok) {
        const err = await res.text();
        throw new Error(err || "Failed to delete webhook");
    }
    return res.json();
}

export async function getWebhookDeliveries(webhookId: string, orgId?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/webhooks/${webhookId}/deliveries`, {
            method: "GET",
            headers,
            cache: 'no-store'
        });
        if (!res.ok) return [];
        return res.json();
    } catch (e) {
        console.error("Gateway webhook deliveries fetch failed", e);
        return [];
    }
}

// ── Workflow Registry (E4) ──────────────────────────────────────

export async function getWorkflows(orgId?: string, status?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const url = status
            ? `${GATEWAY_URL}/v1/workflows?status=${status}`
            : `${GATEWAY_URL}/v1/workflows`
        const res = await fetch(url, { cache: 'no-store', headers });
        if (!res.ok) return [];
        return res.json();
    } catch (e) {
        console.error("Gateway workflows fetch failed", e);
        return [];
    }
}

export async function approveWorkflow(orgId: string, workflowId: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/workflows/approve`, {
            method: 'PUT',
            headers,
            body: JSON.stringify({ id: workflowId }),
        });
        return res.json();
    } catch (e) {
        console.error("Workflow approve failed", e);
        return { error: 'failed' };
    }
}

export async function rejectWorkflow(orgId: string, workflowId: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/workflows/reject`, {
            method: 'PUT',
            headers,
            body: JSON.stringify({ id: workflowId }),
        });
        return res.json();
    } catch (e) {
        console.error("Workflow reject failed", e);
        return { error: 'failed' };
    }
}

export async function getOfflineStats(orgId?: string) {
    try {
        const headers = await gatewayHeaders(orgId)
        const res = await fetch(`${GATEWAY_URL}/v1/offline/stats`, {
            cache: 'no-store',
            headers,
        });
        if (!res.ok) return null;
        return res.json();
    } catch (e) {
        console.error("Offline stats fetch failed", e);
        return null;
    }
}