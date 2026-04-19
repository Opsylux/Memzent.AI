'use client'

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Book, Zap, Shield, Key, Code, MessageCircle } from "lucide-react";

const docSections = [
  {
    title: "Core Concepts",
    items: [
      { name: "Introduction", href: "/docs", icon: Book },
      { name: "Aura Architecture", href: "/docs/architecture", icon: Zap },
      { name: "Semantic Proxying", href: "/docs/semantic-proxy", icon: MessageCircle },
    ]
  },
  {
    title: "Getting Started",
    items: [
      { name: "Quickstart", href: "/docs/quickstart", icon: Zap },
      { name: "Authentication", href: "/docs/auth", icon: Key },
      { name: "Your First Request", href: "/docs/first-request", icon: Code },
    ]
  },
  {
    title: "Security & RBAC",
    items: [
      { name: "RBAC Overview", href: "/docs/rbac", icon: Shield },
      { name: "Managing Permissions", href: "/docs/permissions", icon: Shield },
    ]
  }
];

export function DocsSidebar() {
  const pathname = usePathname();

  return (
    <div className="space-y-8">
      {docSections.map((section) => (
        <div key={section.title} className="space-y-3">
          <h4 className="text-[10px] font-black uppercase tracking-[0.2em] text-white/20 px-4 italic">
            {section.title}
          </h4>
          <nav className="flex flex-col gap-1">
            {section.items.map((item) => {
              const isActive = pathname === item.href;
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={`flex items-center gap-3 px-4 py-2 rounded-lg text-xs font-bold transition-all ${
                    isActive
                      ? "bg-aura-glow/10 text-aura-glow border border-aura-glow/20 shadow-[0_0_15px_rgba(0,243,255,0.03)]"
                      : "text-white/40 hover:text-white hover:bg-white/5"
                  }`}
                >
                  <item.icon size={14} className={isActive ? "text-aura-glow" : "text-white/20"} />
                  {item.name}
                </Link>
              );
            })}
          </nav>
        </div>
      ))}
    </div>
  );
}
