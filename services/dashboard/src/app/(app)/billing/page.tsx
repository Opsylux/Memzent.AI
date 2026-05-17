'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Shield, Check, Zap, Sparkles, Building2, ExternalLink, ArrowRight } from 'lucide-react'
import { createCheckoutSession } from '@/app/actions'
import { supabase } from '@/lib/supabase'
import { plans } from '@/app/plans'

export default function BillingPage() {
  const [currentTier, setCurrentTier] = useState('free')
  const [orgName, setOrgName] = useState('')
  const [orgId, setOrgId] = useState<string | null>(null)
  const [loadingPlan, setLoadingPlan] = useState<string | null>(null)

  useEffect(() => {
    async function load() {
      const { data: { user } } = await supabase.auth.getUser()
      if (user) {
        // Resolve org membership
        const { data: membership } = await supabase
          .from('members')
          .select('org_id, organizations(id, name, subscription_tier)')
          .eq('user_id', user.id)
          .limit(1)
          .maybeSingle()

        if (membership?.organizations) {
          const org = membership.organizations as any
          setOrgId(org.id)
          setOrgName(org.name || '')
          setCurrentTier(org.subscription_tier || 'free')
        } else {
          setOrgId(user.id)
          setOrgName(user.email?.split('@')[0] || 'Personal')
        }
      }
    }
    load()
  }, [])

  const handleUpgrade = async (planId: string) => {
    if (planId === currentTier || !orgId || loadingPlan) return
    setLoadingPlan(planId)
    try {
      const { url } = await createCheckoutSession(orgId, planId)
      if (url) {
        window.location.href = url
      }
    } catch (err: any) {
      alert(`Provisioning Error: ${err.message}`)
    } finally {
      setLoadingPlan(null)
    }
  }

  return (
    <div className="space-y-12 pb-20">
      <header className="mb-16">
        <h1 className="text-4xl font-black tracking-tighter text-white mb-2 uppercase italic">
          SUBSCRIPTION_SECTOR
        </h1>
        <p className="text-white/20 font-black uppercase tracking-[0.3em] text-[10px] italic">
          {orgName ? `${orgName} — ` : ''}Scaling Neural Infrastructure Capacity
        </p>
      </header>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        {plans.map((plan) => {
          const isCurrent = plan.id === currentTier
          return (
            <div
              key={plan.id}
              className={`stat-card relative border-white/5 bg-white/5 flex flex-col transition-all duration-500 overflow-hidden group hover:border-memzent-glow/20 ${plan.highlight ? 'border-memzent-purple/30 bg-memzent-purple/5 scale-105 z-10 shadow-[0_0_40px_rgba(151,71,255,0.1)]' : ''
                }`}
            >
              {plan.highlight && (
                <div className="absolute top-0 right-0 p-4">
                  <Badge className="bg-memzent-purple text-white text-[10px] font-black uppercase tracking-widest px-3">Priority</Badge>
                </div>
              )}

              <div className="p-8 pb-0">
                <div className="flex items-center gap-3 mb-4">
                  <div className={`w-10 h-10 rounded-xl flex items-center justify-center border border-white/10 ${plan.id === 'free' ? 'text-white/40 bg-white/5' :
                    plan.id === 'pro' ? 'text-memzent-purple bg-memzent-purple/10 border-memzent-purple/20 shadow-[0_0_15px_rgba(151,71,255,0.2)]' :
                      'text-memzent-glow bg-memzent-glow/10 border-memzent-glow/20 shadow-[0_0_15px_rgba(0,243,255,0.2)]'
                    }`}>
                    {plan.id === 'free' && <Zap size={20} />}
                    {plan.id === 'pro' && <Sparkles size={20} />}
                    {plan.id === 'business' && <Building2 size={20} />}
                  </div>
                  <h3 className="text-2xl font-black italic tracking-tighter uppercase">{plan.name}</h3>
                </div>
                <p className="text-[10px] font-black uppercase text-white/20 tracking-widest mb-8">{plan.description}</p>

                <div className="flex items-baseline gap-2 mb-8">
                  <span className="text-5xl font-black tracking-tighter text-white">{plan.price}</span>
                  <span className="text-[10px] font-black uppercase tracking-[0.2em] text-white/20">/ unit</span>
                </div>
              </div>

              <div className="p-8 pt-0 flex-1 space-y-8">
                <div className="h-px bg-white/5 w-full" />
                <ul className="space-y-4">
                  {plan.features.map((feature) => (
                    <li key={feature} className="flex gap-3 text-[10px] font-black uppercase tracking-widest text-white/40 group-hover:text-white/60 transition-colors">
                      <Check size={14} className={plan.highlight ? "text-memzent-purple" : "text-white/20"} />
                      {feature}
                    </li>
                  ))}
                </ul>
              </div>

              <div className="p-8">
                <Button
                  onClick={() => handleUpgrade(plan.id)}
                  disabled={isCurrent || !!loadingPlan}
                  className={`w-full py-7 rounded-2xl text-[10px] font-black uppercase tracking-[0.3em] transition-all h-14 ${isCurrent
                    ? 'bg-white/5 text-white/20 border border-white/5 cursor-default'
                    : plan.highlight
                      ? 'bg-memzent-purple text-white hover:bg-memzent-purple/80 hover:shadow-[0_0_25px_rgba(151,71,255,0.4)]'
                      : 'bg-white text-black hover:bg-memzent-glow'
                    } ${loadingPlan === plan.id ? 'opacity-50 cursor-wait' : ''}`}
                >
                  {isCurrent ? (
                    <span className="flex items-center gap-2 italic"><Shield size={14} /> ACTIVE_SECTOR</span>
                  ) : loadingPlan === plan.id ? (
                    <span className="flex items-center gap-2 italic">PROVISIONING...</span>
                  ) : (
                    <span className="flex items-center gap-2">{plan.cta} <ArrowRight size={14} /></span>
                  )}
                </Button>
              </div>
            </div>
          )
        })}
      </div>

      <section className="stat-card border-white/5 bg-black/40 p-12 rounded-[40px] relative overflow-hidden group mt-16 neural-bg">
        <div className="flex flex-col md:flex-row items-center justify-between gap-12 relative z-10">
          <div className="space-y-4 flex-1">
            <h3 className="text-2xl font-black tracking-tighter italic uppercase underline decoration-memzent-glow/30 decoration-4 underline-offset-8">Neural Token Top-Up</h3>
            <p className="text-[10px] font-black text-white/20 max-w-lg leading-relaxed uppercase tracking-[0.2em]">
              Add Pay-As-You-Go tokens to your organization's balance. Tokens are automatically deducted for LLM completions and semantic cache hits.
            </p>
          </div>
          <div className="flex flex-col gap-4 w-full md:w-auto">
            <div className="flex items-center gap-4">
              <div className="relative">
                <span className="absolute left-4 top-1/2 -translate-y-1/2 text-white/50 font-black">$</span>
                <input
                  type="number"
                  min="5"
                  defaultValue="20"
                  id="topup-amount"
                  className="w-32 h-16 bg-white/5 border border-white/10 rounded-2xl pl-8 pr-4 text-white font-black text-xl outline-none focus:border-memzent-glow transition-all"
                />
              </div>
              <Button
                onClick={async () => {
                  const input = document.getElementById('topup-amount') as HTMLInputElement;
                  const val = parseFloat(input?.value || "20");
                  if (val < 5) return alert("Minimum top-up is $5.00");
                  setLoadingPlan('topup');
                  try {
                    const { url } = await createCheckoutSession(orgId!, { amount: val });
                    if (url) window.location.href = url;
                  } catch (e: any) {
                    alert("Top-Up failed: " + e.message);
                  } finally {
                    setLoadingPlan(null);
                  }
                }}
                disabled={!orgId || loadingPlan === 'topup'}
                className="bg-memzent-glow text-black font-black px-10 rounded-2xl hover:bg-white transition-all flex items-center gap-3 text-[10px] uppercase tracking-[0.3em] h-16"
              >
                {loadingPlan === 'topup' ? 'PROCESSING...' : 'ADD FUNDS'} <Zap size={16} />
              </Button>
            </div>
          </div>
        </div>
        <div className="absolute inset-x-0 bottom-0 h-64 bg-gradient-to-t from-memzent-glow/5 to-transparent pointer-events-none" />
        <div className="absolute inset-0 pointer-events-none opacity-[0.03] grayscale bg-[url('https://grainy-gradients.vercel.app/noise.svg')]" />
      </section>

      <section className="stat-card border-white/5 bg-black/40 p-12 rounded-[40px] relative overflow-hidden group mt-16 neural-bg">
        <div className="flex flex-col md:flex-row items-center justify-between gap-12 relative z-10">
          <div className="space-y-4">
            <h3 className="text-2xl font-black tracking-tighter italic uppercase underline decoration-memzent-glow/30 decoration-4 underline-offset-8">Custom Capacity_01</h3>
            <p className="text-[10px] font-black text-white/20 max-w-lg leading-relaxed uppercase tracking-[0.2em]">
              Executing over 10M tokens monthly? Require private model hosting or dedicated vector clusters? Contact the infrastructure team for an Enterprise SLA.
            </p>
          </div>
          <Button variant="outline" className="border-white/10 text-white font-black px-10 py-7 rounded-2xl hover:bg-memzent-glow hover:text-black hover:border-memzent-glow transition-all flex items-center gap-3 group text-[10px] uppercase tracking-[0.3em] h-16">
            Contact HQ <ExternalLink size={16} className="group-hover:translate-x-1 group-hover:-translate-y-1 transition-transform" />
          </Button>
        </div>
        <div className="absolute inset-x-0 bottom-0 h-64 bg-gradient-to-t from-memzent-purple/5 to-transparent pointer-events-none" />
        <div className="absolute inset-0 pointer-events-none opacity-[0.03] grayscale bg-[url('https://grainy-gradients.vercel.app/noise.svg')]" />
      </section>
    </div>
  )
}
