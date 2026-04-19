import { Shield, Lock, Users, Fingerprint } from "lucide-react";

export default function RBACGuide() {
  return (
    <div className="space-y-12">
      <header className="space-y-4">
        <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-aura-purple/5 border border-aura-purple/20 w-fit">
          <span className="text-[10px] font-black text-aura-purple uppercase tracking-tighter italic">Security_Spec_03</span>
        </div>
        <h1 className="text-4xl font-black tracking-tighter uppercase sm:text-5xl">RBAC & Security</h1>
        <p className="text-lg text-white/60 leading-relaxed font-medium">
          Aura enforces a **Zero-Trust** security model where every request is verified against a real-time database state, not just a static JWT claim.
        </p>
      </header>

      <section className="space-y-6">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Verified Role Enforcement</h2>
        <div className="space-y-4 text-sm text-white/60 leading-relaxed font-medium">
          <p>
            While many proxies trust the `role` claim inside a JWT, Aura treats JWTs as identity hints only. The Gateway performs a sub-1ms check against the persistent `members` table to verify:
          </p>
          <ul className="list-disc pl-6 space-y-3 marker:text-aura-glow">
            <li><strong className="text-white">Organization Membership</strong>: Is the user actually a member of `org_id`?</li>
            <li><strong className="text-white">Verified Role</strong>: Does the user have `admin` or `member` privileges in this specific organization?</li>
          </ul>
        </div>
      </section>

      <section className="space-y-8 pt-6">
        <h2 className="text-2xl font-black tracking-tighter uppercase">Standard Roles</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="p-4 rounded-xl bg-white/[0.02] border border-white/5 space-y-2">
            <div className="flex items-center gap-2 text-aura-glow">
              <Lock size={14} />
              <span className="text-xs font-black uppercase">Admin / Owner</span>
            </div>
            <p className="text-[11px] text-white/40 font-bold">Full control over organization tools, billing, and API keys. Can manage all member keys.</p>
          </div>
          <div className="p-4 rounded-xl bg-white/[0.02] border border-white/5 space-y-2">
            <div className="flex items-center gap-2 text-white/40">
              <Users size={14} />
              <span className="text-xs font-black uppercase">Member</span>
            </div>
            <p className="text-[11px] text-white/40 font-bold">Can create and use personal API keys. Limited to tool execution and personal audit viewing.</p>
          </div>
        </div>
      </section>

      <section className="space-y-6 pt-6">
         <h2 className="text-2xl font-black tracking-tighter uppercase">Row-Level Security (RLS)</h2>
         <p className="text-sm text-white/60 leading-relaxed font-medium">
           Aura uses PostgreSQL RLS at the storage layer to ensure true multi-tenant isolation. Even if a user bypasses the Gateway logic, the database itself will reject any cross-tenant data access.
         </p>
      </section>
    </div>
  );
}
