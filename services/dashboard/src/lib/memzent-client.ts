// services/dashboard/src/lib/memzent-client.ts
const GATEWAY_URL = process.env.NEXT_PUBLIC_GATEWAY_URL || 'http://localhost:8080';

export async function queryMemzent(prompt: string) {
    const response = await fetch(`${GATEWAY_URL}/v1/chat?prompt=${encodeURIComponent(prompt)}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
        // Next.js 15+ / Bun caching logic
        next: { revalidate: 60 }
    });

    if (!response.ok) throw new Error('Memzent Gateway offline');
    return response.json();
}