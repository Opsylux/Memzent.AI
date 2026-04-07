"use server"

import { supabase } from '@/lib/supabase';

const GATEWAY_URL = process.env.GATEWAY_INTERNAL_URL || process.env.NEXT_PUBLIC_GATEWAY_URL || 'http://localhost:8080';

export async function getAuraTools() {
    try {
        const res = await fetch(`${GATEWAY_URL}/v1/tools`, { cache: 'no-store' });
        if (!res.ok) return [];
        return res.json();
    } catch (e) {
        console.error("Gateway fetch failed", e);
        return [];
    }
}

export async function getAuraStats() {
    try {
        const res = await fetch(`${GATEWAY_URL}/v1/stats`, { cache: 'no-store' });
        if (!res.ok) return { total_requests: 0, cache_hits: 0, uptime_seconds: 0 };
        return res.json();
    } catch (e) {
        console.error("Gateway stats fetch failed", e);
        return { total_requests: 0, cache_hits: 0, uptime_seconds: 0 };
    }
}

const DEV_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzQ3NjgxMzksInJvbGUiOiJhZG1pbiIsInN1YiI6ImFkbWluLTAxIn0.4Ffru9o6slUOgPCibCoNGpeIMJoLPF_WgRbXH8FqBrM";

export async function executeAuraPrompt(prompt: string) {
    try {
        const res = await fetch(`${GATEWAY_URL}/v1/chat`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${DEV_TOKEN}`
            },
            body: JSON.stringify({
                user_id: "admin-01",
                prompt: prompt
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

// SaaS API Key Management
export async function getApiKeys(orgId: string) {
    const { data, error } = await supabase
        .from('api_keys')
        .select('*')
        .eq('org_id', orgId)
        .order('created_at', { ascending: false });

    if (error) throw error;
    return data;
}

export async function createApiKey(orgId: string, name: string) {
    // In production, you'd generate a secure key and hash it.
    const key = `aura_${Math.random().toString(36).substring(2, 15)}${Math.random().toString(36).substring(2, 15)}`;
    const prefix = key.substring(0, 8);
    
    const { error } = await supabase
        .from('api_keys')
        .insert({
            org_id: orgId,
            name: name,
            key_prefix: prefix,
            key_hash: key // In reality, use bcrypt/argon2
        });

    if (error) throw error;
    return { key };
}

export async function revokeApiKey(id: string) {
    const { error } = await supabase
        .from('api_keys')
        .delete()
        .eq('id', id);

    if (error) throw error;
}

// SaaS Tool Provisioning
export async function createAuraTool(orgId: string, tool: any) {
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
    const { data, error } = await supabase
        .from('tools')
        .select('*')
        .or(`org_id.eq.${orgId},org_id.is.null`)
        .order('created_at', { ascending: false });

    if (error) throw error;
    return data;
}