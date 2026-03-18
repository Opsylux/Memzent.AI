"use server"

export async function getAuraTools() {
    try {
        // Internal Docker network URL
        const res = await fetch("http://gateway:8080/v1/tools", { cache: 'no-store' });
        if (!res.ok) return [];
        return res.json();
    } catch (e) {
        console.error("Gateway fetch failed", e);
        return [];
    }
}