'use client'

import { useState } from "react";
import { Check, Copy } from "lucide-react";

interface CodeBlockProps {
  code: string;
  language?: string;
  filename?: string;
  title?: string;
}

export function CodeBlock({ code, language = "typescript", filename, title }: CodeBlockProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="relative group rounded-xl overflow-hidden border border-white/5 bg-black/40 my-6 shadow-2xl">
      {(filename || title) && (
        <div className="flex items-center justify-between px-4 py-2 border-b border-white/5 bg-white/[0.02]">
          <span className="text-[10px] font-black uppercase tracking-widest text-white/20">{filename || title}</span>
          <span className="text-[10px] font-black uppercase tracking-widest text-white/20">{language}</span>
        </div>
      )}
      <div className="relative">
        <pre className="p-4 overflow-x-auto text-[13px] font-mono leading-relaxed text-slate-300 scrollbar-hide">
          <code className={`language-${language}`}>{code}</code>
        </pre>
        <button
          onClick={handleCopy}
          className="absolute right-4 top-4 p-2 rounded-lg bg-white/5 border border-white/10 text-white/40 hover:text-memzent-glow hover:border-memzent-glow/20 transition-all opacity-0 group-hover:opacity-100"
          title="Copy to clipboard"
        >
          {copied ? <Check size={14} className="text-memzent-glow" /> : <Copy size={14} />}
        </button>
      </div>
    </div>
  );
}
