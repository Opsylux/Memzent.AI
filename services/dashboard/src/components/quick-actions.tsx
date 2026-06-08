import Link from "next/link";
import { ArrowRight, FlaskConical, Key, Database, BookOpen, CreditCard } from "lucide-react";

const actions = [
  {
    href: "/playground",
    label: "Test a prompt",
    description: "Route through cache & tools live",
    icon: FlaskConical,
    accent: "text-memzent-glow border-memzent-glow/20 bg-memzent-glow/5",
  },
  {
    href: "/keys",
    label: "API keys",
    description: "Create or rotate agent keys",
    icon: Key,
    accent: "text-memzent-purple border-memzent-purple/20 bg-memzent-purple/5",
  },
  {
    href: "/tools",
    label: "Tool registry",
    description: "Register REST, SQL, MCP tools",
    icon: Database,
    accent: "text-memzent-accent border-memzent-accent/20 bg-memzent-accent/5",
  },
  {
    href: "/docs/quickstart",
    label: "Quickstart",
    description: "First request in 5 minutes",
    icon: BookOpen,
    accent: "text-white/80 border-white/10 bg-white/5",
  },
  {
    href: "/billing",
    label: "Billing",
    description: "Balance & token usage",
    icon: CreditCard,
    accent: "text-white/80 border-white/10 bg-white/5",
  },
];

export function QuickActions() {
  return (
    <section className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
      {actions.map((action) => (
        <Link
          key={action.href}
          href={action.href}
          className="group stat-card border-white/5 p-5 hover:border-white/20 transition-all"
        >
          <div className={`inline-flex p-2.5 rounded-xl border mb-4 ${action.accent}`}>
            <action.icon size={18} />
          </div>
          <div className="text-sm font-bold text-readable-primary group-hover:text-memzent-glow transition-colors">
            {action.label}
          </div>
          <p className="text-[11px] text-readable-muted mt-1 leading-snug">{action.description}</p>
          <ArrowRight
            size={14}
            className="mt-3 text-white/20 group-hover:text-memzent-glow group-hover:translate-x-0.5 transition-all"
          />
        </Link>
      ))}
    </section>
  );
}
