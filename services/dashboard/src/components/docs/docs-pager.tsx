'use client'

import Link from "next/link";
import { usePathname } from "next/navigation";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { docSections } from "./docs-sidebar";

export function DocsPager() {
  const pathname = usePathname();

  // Flatten items for easy navigation
  const allItems = docSections.flatMap(section => section.items);
  const currentIndex = allItems.findIndex(item => item.href === pathname);

  if (currentIndex === -1) return null;

  const prev = allItems[currentIndex - 1];
  const next = allItems[currentIndex + 1];

  return (
    <div className="flex items-center justify-between gap-4 mt-16 pt-8 border-t border-white/5">
      {prev ? (
        <Link
          href={prev.href}
          className="group flex flex-col gap-2 p-4 rounded-xl border border-white/5 bg-white/[0.02] hover:bg-white/[0.04] transition-all flex-1 max-w-[240px]"
        >
          <div className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-white/20 group-hover:text-memzent-glow transition-colors">
            <ChevronLeft size={12} />
            Previous
          </div>
          <div className="text-sm font-black text-white group-hover:translate-x-1 transition-transform">
            {prev.name}
          </div>
        </Link>
      ) : (
        <div className="flex-1" />
      )}

      {next ? (
        <Link
          href={next.href}
          className="group flex flex-col items-end gap-2 p-4 rounded-xl border border-white/5 bg-white/[0.02] hover:bg-white/[0.04] transition-all flex-1 max-w-[240px] text-right"
        >
          <div className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-white/20 group-hover:text-memzent-glow transition-colors">
            Next
            <ChevronRight size={12} />
          </div>
          <div className="text-sm font-black text-white group-hover:-translate-x-1 transition-transform">
            {next.name}
          </div>
        </Link>
      ) : (
        <div className="flex-1" />
      )}
    </div>
  );
}
