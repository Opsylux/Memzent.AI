import React from 'react'

interface MarkdownProps {
  content: string
}

export function Markdown({ content }: MarkdownProps) {
  if (!content) return null

  // Split content by code blocks first
  const parts = content.split(/(```[\s\S]*?```)/g)

  return (
    <div className="space-y-4 text-xs font-semibold text-white/70 leading-relaxed break-words max-w-full">
      {parts.map((part, index) => {
        if (part.startsWith('```')) {
          // Code block! Extract language and code
          const lines = part.split('\n')
          const firstLine = lines[0] || '```'
          const lang = firstLine.replace('```', '').trim() || 'plaintext'
          const code = lines.slice(1, -1).join('\n')

          return (
            <div key={index} className="relative group my-4 max-w-full">
              <div className="absolute top-2 right-3 text-[8px] font-black uppercase tracking-widest text-white/20 select-none">
                {lang}
              </div>
              <pre className="bg-black/60 border border-white/10 rounded-xl px-5 py-4 font-mono text-[11px] leading-relaxed text-memzent-glow max-w-full whitespace-pre-wrap break-all">
                <code>{code}</code>
              </pre>
            </div>
          )
        }

        // Standard text blocks - split by double-newline
        const blocks = part.split(/\n\n+/g)

        return (
          <React.Fragment key={index}>
            {blocks.map((block, blockIdx) => {
              const trimmedBlock = block.trim()
              if (!trimmedBlock) return null

              // Check if block is a header
              if (trimmedBlock.startsWith('#')) {
                const match = trimmedBlock.match(/^(#{1,6})\s+(.*)$/)
                if (match) {
                  const level = match[1].length
                  const text = match[2]
                  const parsedText = parseInline(text)
                  
                  if (level === 1) {
                    return <h1 key={blockIdx} className="text-lg font-black uppercase text-white tracking-tight mt-5 mb-2 break-words">{parsedText}</h1>
                  } else if (level === 2) {
                    return <h2 key={blockIdx} className="text-sm font-black uppercase text-white tracking-tight mt-4 mb-2 break-words">{parsedText}</h2>
                  } else {
                    return <h3 key={blockIdx} className="text-[10px] font-black uppercase tracking-[0.2em] text-white/90 mt-3 mb-2 break-words">{parsedText}</h3>
                  }
                }
              }

              // Check if block is a list (contains lines starting with - or * or numbers)
              const lines = trimmedBlock.split('\n')
              const isList = lines.some(line => /^(?:[-*]|\d+\.)\s+/.test(line.trim()))

              if (isList) {
                return (
                  <ul key={blockIdx} className="list-none space-y-2 pl-2 my-3 max-w-full break-words">
                    {lines.map((line, lineIdx) => {
                      const trimmedLine = line.trim()
                      const listMatch = trimmedLine.match(/^(?:([-*])|(\d+\.))\s+(.*)$/)
                      if (listMatch) {
                        const isNumbered = !!listMatch[2]
                        const bullet = listMatch[1] || listMatch[2]
                        const itemText = listMatch[3]
                        const parsedText = parseInline(itemText)
                        
                        return (
                          <li key={lineIdx} className="flex items-start gap-2.5 text-xs font-semibold leading-relaxed">
                            {isNumbered ? (
                              <span className="text-[10px] font-black font-mono text-memzent-glow mt-0.5">{bullet}</span>
                            ) : (
                              <div className="w-1.5 h-1.5 rounded-full bg-memzent-glow mt-1.5 shadow-[0_0_6px_rgba(0,243,255,0.6)] flex-shrink-0" />
                            )}
                            <span className="text-white/80">{parsedText}</span>
                          </li>
                        )
                      }
                      // Continuation line
                      return (
                        <li key={lineIdx} className="pl-4 text-xs font-semibold text-white/70 leading-relaxed">
                          {parseInline(trimmedLine)}
                        </li>
                      )
                    })}
                  </ul>
                )
              }

              // Default Paragraph
              return (
                <p key={blockIdx} className="text-xs font-semibold text-white/70 leading-relaxed mb-3 break-words max-w-full">
                  {parseInline(trimmedBlock)}
                </p>
              )
            })}
          </React.Fragment>
        )
      })}
    </div>
  )
}

function parseInline(text: string): React.ReactNode[] {
  // Regex to match bold (**), italic (*), and inline code (`)
  const parts = text.split(/(\*\*.*?\*\*|\*.*?\*|`.*?`)/g)
  
  return parts.map((part, idx) => {
    if (part.startsWith('**') && part.endsWith('**')) {
      return <strong key={idx} className="font-black text-white">{part.slice(2, -2)}</strong>
    }
    if (part.startsWith('*') && part.endsWith('*')) {
      return <em key={idx} className="italic text-white/90">{part.slice(1, -1)}</em>
    }
    if (part.startsWith('`') && part.endsWith('`')) {
      return <code key={idx} className="bg-white/5 border border-white/10 px-1.5 py-0.5 rounded font-mono text-[10px] text-memzent-glow break-all">{part.slice(1, -1)}</code>
    }
    return part
  })
}
