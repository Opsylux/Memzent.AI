import { getWebhooks } from "../../actions";
import { getCurrentOrg } from "@/lib/user-context";
import { Bell, Plus, Globe, CheckCircle2, XCircle, Clock } from "lucide-react";
import { WebhookActions } from "@/components/webhook-actions";

export default async function NotificationsPage() {
  const org = await getCurrentOrg();
  const webhooks = await getWebhooks(org?.orgId);

  return (
    <div className="space-y-12">
      <header className="flex flex-col md:flex-row justify-between items-start md:items-end gap-6 mb-12">
        <div>
          <h1 className="text-4xl font-black tracking-tighter">Notifications</h1>
          <p className="text-sm font-bold text-white/50 uppercase tracking-widest mt-1">
            Webhook Event Subscriptions &amp; Delivery Logs
          </p>
        </div>
        <WebhookActions webhooks={webhooks || []} orgId={org?.orgId} mode="header" />
      </header>

      {/* Event Types Reference */}
      <section className="stat-card border-white/5 p-6">
        <h2 className="text-xs font-black uppercase tracking-widest text-white/50 mb-4">Available Event Types</h2>
        <div className="flex flex-wrap gap-2">
          {['cache_hit', 'tool_execution', 'rate_limit', 'key_rotated', 'tool_registered', 'session_created'].map(evt => (
            <span key={evt} className="text-[10px] font-bold px-3 py-1.5 rounded-lg bg-memzent-glow/5 border border-memzent-glow/10 text-memzent-glow/80">
              {evt}
            </span>
          ))}
        </div>
      </section>

      {/* Webhooks List */}
      <section className="stat-card border-white/5 p-0 overflow-hidden">
        <div className="p-8 border-b border-white/5">
          <h2 className="text-sm font-black uppercase tracking-widest">Registered Webhooks</h2>
        </div>

        {(!webhooks || webhooks.length === 0) ? (
          <div className="p-12 text-center">
            <Bell size={32} className="mx-auto text-white/20 mb-4" />
            <p className="text-sm text-white/40 mb-2">No webhooks configured</p>
            <p className="text-xs text-white/30">Create a webhook to receive real-time event notifications via HTTP POST.</p>
          </div>
        ) : (
          <div className="divide-y divide-white/5">
            {webhooks.map((wh: any) => (
              <div key={wh.id} className="p-6 hover:bg-white/[0.02] transition-all group">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex items-start gap-4 flex-1 min-w-0">
                    <div className="w-10 h-10 rounded-lg bg-white/5 border border-white/10 flex items-center justify-center shrink-0">
                      <Globe size={18} className={wh.enabled ? 'text-memzent-glow' : 'text-white/30'} />
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="text-sm font-bold truncate">{wh.url}</span>
                        <span className={`text-[9px] font-black px-1.5 py-0.5 rounded ${wh.enabled ? 'bg-memzent-accent/10 text-memzent-accent' : 'bg-white/5 text-white/40'}`}>
                          {wh.enabled ? 'ACTIVE' : 'DISABLED'}
                        </span>
                      </div>
                      {wh.description && (
                        <p className="text-xs text-white/40 mb-2">{wh.description}</p>
                      )}
                      <div className="flex flex-wrap gap-1.5">
                        {(wh.events || []).map((evt: string) => (
                          <span key={evt} className="text-[9px] font-bold px-2 py-0.5 rounded bg-white/5 text-white/50">
                            {evt}
                          </span>
                        ))}
                      </div>
                    </div>
                  </div>
                  <WebhookActions webhooks={[wh]} orgId={org?.orgId} mode="row" webhookId={wh.id} />
                </div>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
