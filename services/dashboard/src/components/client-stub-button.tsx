'use client'

import React from 'react'
import { ExternalLink, MoreHorizontal } from 'lucide-react'

export function ClientStubButton({
  message,
  type
}: {
  message: string,
  type: 'external' | 'more'
}) {
  return (
    <button
      className="p-2 rounded-lg hover:bg-white/10 hover:text-memzent-glow transition-all"
      onClick={() => alert(message)}
    >
      {type === 'external' ? <ExternalLink size={16} /> : <MoreHorizontal size={16} />}
    </button>
  )
}
