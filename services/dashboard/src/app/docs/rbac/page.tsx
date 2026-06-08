import { Shield, Lock, Users, CheckCircle2, AlertCircle } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";
import Link from "next/link";

export default function RBACGuide() {
  return (
    <div className="space-y-12">
      <header className="space-y-4">
        <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-memzent-purple/5 border border-memzent-purple/20 w-fit">
          <span className="text-[10px] font-black text-memzent-purple uppercase tracking-tighter italic">Security</span>
        </div>
        <h1 className="text-4xl font-black tracking-tighter uppercase sm:text-5xl">Access Control</h1>
        <p className="text-lg text-white/60 leading-relaxed font-medium">
          Memzent follows a Zero-Trust model — every request is checked against your organization&apos;s live access rules, not just a credential that was issued at login.
        </p>
      </header>

      {/* How it works */}
      <section className="space-y-5">
        <h2 className="text-2xl font-black tracking-tighter uppercase">How Verification Works</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          When a request arrives, Memzent performs two checks before allowing anything to proceed:
        </p>
        <div className="space-y-3">
          {[
            {
              title: "Is this user in the right organization?",
              desc: "Membership is verified in real time. If a user is removed from your organization, they lose access on their very next request — no manual revocation needed."
            },
            {
              title: "Does this user have the right role?",
              desc: "Roles are looked up live. Changing a user from Member to Admin (or removing them entirely) takes effect immediately."
            }
          ].map((item, i) => (
            <div key={i} className="flex gap-4 p-5 rounded-2xl bg-white/[0.02] border border-white/5">
              <div className="w-7 h-7 rounded-lg bg-memzent-purple/10 flex items-center justify-center text-memzent-purple shrink-0 mt-0.5">
                <Shield size={13} />
              </div>
              <div className="space-y-1">
                <div className="text-xs font-black uppercase tracking-tight text-white/80">{item.title}</div>
                <p className="text-[11px] text-white/40 font-bold leading-relaxed">{item.desc}</p>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* Roles */}
      <section className="space-y-5 pt-2">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Roles</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Every organization member has one of three roles. Roles control what they can do — not just what they can see.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
          <div className="p-6 rounded-2xl bg-white/[0.02] border border-white/5 space-y-4">
            <div className="flex items-center gap-2 text-memzent-glow">
              <Lock size={15} />
              <span className="text-xs font-black uppercase">Admin</span>
            </div>
            <ul className="space-y-2">
              {[
                "Full org management",
                "Manage all API keys",
                "Register and remove tools",
                "View all audit activity",
                "Manage billing and spend limits",
                "Invite and remove members",
                "Full rate limit allocation",
              ].map((item) => (
                <li key={item} className="flex items-center gap-2">
                  <CheckCircle2 size={12} className="text-memzent-glow shrink-0" />
                  <span className="text-[11px] text-white/40 font-bold">{item}</span>
                </li>
              ))}
            </ul>
          </div>

          <div className="p-6 rounded-2xl bg-white/[0.02] border border-white/5 space-y-4">
            <div className="flex items-center gap-2 text-blue-400">
              <Users size={15} />
              <span className="text-xs font-black uppercase">Member (Agent)</span>
            </div>
            <ul className="space-y-2">
              {[
                "Execute chat prompts",
                "Create personal API keys",
                "Use permitted tools",
                "View their own activity",
                "Discover available models",
                "Proportional rate limit share",
              ].map((item) => (
                <li key={item} className="flex items-center gap-2">
                  <CheckCircle2 size={12} className="text-blue-400/50 shrink-0" />
                  <span className="text-[11px] text-white/40 font-bold">{item}</span>
                </li>
              ))}
            </ul>
          </div>

          <div className="p-6 rounded-2xl bg-white/[0.02] border border-white/5 space-y-4">
            <div className="flex items-center gap-2 text-white/40">
              <Users size={15} />
              <span className="text-xs font-black uppercase">Viewer</span>
            </div>
            <ul className="space-y-2">
              {[
                "Read-only access",
                "View tools and models",
                "View audit logs (own)",
                "Cannot execute prompts",
                "Cannot register tools",
                "Blocked from chat:execute",
              ].map((item) => (
                <li key={item} className="flex items-center gap-2">
                  <CheckCircle2 size={12} className="text-white/20 shrink-0" />
                  <span className="text-[11px] text-white/40 font-bold">{item}</span>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </section>

      {/* Org isolation */}
      <section className="space-y-5 pt-2">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Organization Isolation</h2>
        <p className="text-sm text-white/60 leading-relaxed font-medium">
          Memzent is built for multi-tenant environments. An Admin in Organization A has zero access to Organization B&apos;s tools, data, or audit logs — even if they are a member of both.
        </p>
        <div className="p-4 rounded-xl bg-white/[0.01] border border-white/5 flex items-start gap-3">
          <AlertCircle size={13} className="text-white/20 mt-0.5 shrink-0" />
          <p className="text-[11px] text-white/30 font-bold leading-relaxed">
            Data isolation is enforced at the database level — not just in application code. Even if an API request incorrectly specifies the wrong organization, the underlying data layer will reject it.
          </p>
        </div>
      </section>

      {/* Links */}
      <div className="flex flex-wrap gap-4 pt-4 border-t border-white/5">
        <Link href="/docs/permissions" className="text-xs text-memzent-glow font-black uppercase tracking-widest hover:underline">
          Managing Permissions →
        </Link>
        <Link href="/docs/auth" className="text-xs text-white/30 font-black uppercase tracking-widest hover:text-white transition-colors">
          Authentication Guide →
        </Link>
      </div>

      <DocsPager />
    </div>
  );
}
