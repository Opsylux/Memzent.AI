const GATEWAY_URL =
  process.env.NEXT_PUBLIC_GATEWAY_URL || 'http://localhost:8080'

export type ChatMessage = {
  role: 'user' | 'assistant' | 'system'
  content: string
}

export type ChatRequest = {
  messages: ChatMessage[]
  session_id?: string
  provider?: string
  model?: string
  skip_cache?: boolean
  stream?: boolean
}

export type ChatResponse = {
  text: string
  cached: boolean
  provider?: string
  request_id?: string
  session_id?: string
}

export type MemzentClientOptions = {
  /** Bearer JWT or raw token (Bearer prefix added automatically) */
  token?: string
  /** Agent API key — alternative to JWT */
  apiKey?: string
  orgId?: string
}

/**
 * POST /v1/chat — matches gateway engine.PromptRequest.
 * Prefer server actions for dashboard pages; use this for client-side calls only.
 */
export async function chatMemzent(
  body: ChatRequest,
  opts: MemzentClientOptions = {}
): Promise<ChatResponse> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  if (opts.apiKey) {
    headers['X-API-Key'] = opts.apiKey
  } else if (opts.token) {
    const t = opts.token.startsWith('Bearer ') ? opts.token : `Bearer ${opts.token}`
    headers['Authorization'] = t
  } else {
    throw new Error('chatMemzent requires token or apiKey')
  }

  if (opts.orgId) {
    headers['X-Org-ID'] = opts.orgId
  }

  const response = await fetch(`${GATEWAY_URL}/v1/chat`, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
    cache: 'no-store',
  })

  if (!response.ok) {
    const errText = await response.text().catch(() => '')
    throw new Error(`Memzent Gateway error ${response.status}: ${errText}`)
  }

  return response.json()
}

/** @deprecated Use chatMemzent — auto-resolves Supabase session when opts omit credentials */
export async function queryMemzent(
  prompt: string,
  opts: MemzentClientOptions = {}
): Promise<ChatResponse> {
  let resolved = opts
  if (!opts.apiKey && !opts.token && typeof window !== 'undefined') {
    try {
      const { supabase } = await import('@/lib/supabase')
      const { data: { session } } = await supabase.auth.getSession()
      if (session?.access_token) {
        resolved = { ...opts, token: session.access_token }
      }
    } catch {
      // fall through to chatMemzent error
    }
  }
  return chatMemzent(
    { messages: [{ role: 'user', content: prompt }] },
    resolved
  )
}
