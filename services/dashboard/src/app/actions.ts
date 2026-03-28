"use server"

export async function getAuraTools() {
    try {
        // Internal Docker network URL
        const res = await fetch("http://aura-gateway:8080/v1/tools", { cache: 'no-store' });
        if (!res.ok) return [];
        return res.json();
    } catch (e) {
        console.error("Gateway fetch failed", e);
        return [];
    }
}

export async function getAuraStats() {
    try {
        const res = await fetch("http://aura-gateway:8080/v1/stats", { cache: 'no-store' });
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
        const res = await fetch("http://aura-gateway:8080/v1/chat", {
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