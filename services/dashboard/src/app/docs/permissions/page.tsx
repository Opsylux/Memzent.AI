import { Shield, Settings, CheckCircle2, AlertCircle, Lock, Users } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";
import Link from "next/link";

export default function PermissionsPage() {
  return (
    <div className="space-y-12">
      <header className="space-y-4">
        <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-aura-purple/5 border border-aura-purple/20 w-fit">
          <span className="text-[10px] font-black text-aura-purple uppercase tracking-tighter italic">Security</span>
        </div>
        <h1 className="text-4xl font-black tracking-tighter uppercase sm:text-5xl">Managing Permissions</h1>
        <p className="text-lg text-white/60 leading-relaxed font-medium">
          Aura gives you fine-grained control over what each user and API key can access — scoped to your organization and down to individual tools.
        </p>
      </header>

      {/* Scoping */}
      <section className="space-y-5">
        <h2 className="text-2xl font-black tracking-tighter uppercase">How Permissions Are Scoped</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Every permission in Aura belongs to a specific organization. A user who is an Admin in one organization has no special access in another — even if they are a member of both.
        </p>
        <div className="p-5 rounded-2xl bg-white/[0.02] border border-white/5 space-y-4">
          <div className="text-xs font-black uppercase text-white/50 tracking-widest">Permission Hierarchy</div>
          <div className="space-y-2 text-[11px] text-white/40 font-bold">
            <div className="flex gap-3"><span className="text-aura-glow shrink-0">Organization</span><span>→ contains Members, each with a Role</span></div>
            <div className="flex gap-3"><span className="text-aura-purple shrink-0">Role (Admin/Member)</span><span>→ controls what actions are allowed</span></div>
            <div className="flex gap-3"><span className="text-aura-accent shrink-0">Tool Gating</span><span>→ further restricts which tools can be called</span></div>
          </div>
        </div>
      </section>

      {/* Role capabilities */}
      <section className="space-y-5 pt-2">
        <h2 className="text-2xl font-black tracking-tighter uppercase">What Each Role Can Do</h2>

        <div className="space-y-4">
          <div className="p-6 rounded-2xl bg-white/[0.02] border border-white/5 space-y-5">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-aura-glow/10 flex items-center justify-center text-aura-glow">
                <Lock size={14} />
              </div>
              <h3 className="text-sm font-black uppercase tracking-tight">Admin</h3>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {[
                "Generate and revoke all API keys",
                "Register new tools and data sources",
                "View all member activity and history",
                "Manage the billing plan",
                "Invite or remove members",
                "Set per-tool access restrictions",
              ].map((text) => (
                <div key={text} className="flex items-start gap-2">
                  <CheckCircle2 size={13} className="text-aura-glow shrink-0 mt-0.5" />
                  <span className="text-[11px] text-white/40 font-bold">{text}</span>
                </div>
              ))}
            </div>
          </div>

          <div className="p-6 rounded-2xl bg-white/[0.02] border border-white/5 space-y-5">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-white/5 flex items-center justify-center text-white/30">
                <Users size={14} />
              </div>
              <h3 className="text-sm font-black uppercase tracking-tight">Member</h3>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {[
                "Create personal API keys",
                "Use tools they have access to",
                "View their own chat history",
                "Browse available AI models",
              ].map((text) => (
                <div key={text} className="flex items-start gap-2">
                  <CheckCircle2 size={13} className="text-white/20 shrink-0 mt-0.5" />
                  <span className="text-[11px] text-white/40 font-bold">{text}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* Tool gating */}
      <section className="space-y-5 pt-2">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Tool-Level Access Control</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Beyond roles, individual tools can be restricted to a subset of your team. This is useful when you have sensitive tools — for example, a tool with write access to your database — that only a few trusted users should be able to invoke.
        </p>
        <div className="p-5 rounded-2xl bg-aura-purple/5 border border-aura-purple/10 flex items-start gap-4">
          <Settings size={18} className="text-aura-purple mt-0.5 shrink-0" />
          <div className="space-y-2">
            <div className="text-xs font-black uppercase text-white/80">How to configure</div>
            <p className="text-[11px] text-white/40 font-bold leading-relaxed">
              When registering a tool, set <code className="text-aura-purple bg-aura-purple/10 px-1 rounded font-mono">requires_auth: true</code> to gate it behind permissions. Open tools that anyone should be able to use (like search) can be set to <code className="text-aura-purple bg-aura-purple/10 px-1 rounded font-mono">requires_auth: false</code>.
            </p>
          </div>
        </div>

        <div className="p-4 rounded-xl bg-white/[0.01] border border-white/5 flex items-start gap-3">
          <AlertCircle size={13} className="text-white/20 mt-0.5 shrink-0" />
          <p className="text-[11px] text-white/30 font-bold leading-relaxed">
            Only Admins can register, update, or remove tools. Members can only use the tools they already have permission to access.
          </p>
        </div>
      </section>

      <div className="flex flex-wrap gap-4 pt-4 border-t border-white/5">
        <Link href="/docs/rbac" className="text-xs text-aura-glow font-black uppercase tracking-widest hover:underline">
          ← Access Control Overview
        </Link>
        <Link href="/docs/tool-registry" className="text-xs text-white/30 font-black uppercase tracking-widest hover:text-white transition-colors">
          Tool Registry →
        </Link>
      </div>

      <DocsPager />
    </div>
  );
}
