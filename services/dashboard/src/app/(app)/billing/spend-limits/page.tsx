'use client'

import { useState, useEffect, useCallback } from 'react'
import { DollarSign, Gauge, TrendingUp, Shield, AlertTriangle, Save, RotateCcw, Loader2 } from 'lucide-react'
import { getBudgetStatus, getSpendTimeseries, getSpendLimits, setSpendLimits } from '@/app/actions'

interface BudgetData {
  balance: number
  burn_rate_24h: number
  projected_days: number
  spend_24h: number
  spend_7d: number
  spend_30d: number
  provider_breakdown?: Record<string, number>
}

interface SpendLimitsData {
  daily_limit: number | null
  monthly_limit: number | null
  daily_token_limit: number | null
  monthly_token_limit: number | null
  daily_spend: number
  monthly_spend: number
  daily_tokens: number
  monthly_tokens: number
}

interface TimeseriesPoint {
  date: string
  spend: number
  tokens: number
}

export default function SpendLimitsPage() {
  const [budget, setBudget] = useState<BudgetData | null>(null)
  const [limits, setLimits] = useState<SpendLimitsData | null>(null)
  const [timeseries, setTimeseries] = useState<TimeseriesPoint[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  // Form state
  const [dailyLimit, setDailyLimit] = useState<string>('')
  const [monthlyLimit, setMonthlyLimit] = useState<string>('')
  const [dailyTokenLimit, setDailyTokenLimit] = useState<string>('')
  const [monthlyTokenLimit, setMonthlyTokenLimit] = useState<string>('')

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const [budgetRes, limitsRes, tsRes] = await Promise.all([
        getBudgetStatus(),
        getSpendLimits(),
        getSpendTimeseries(30),
      ])
      setBudget(budgetRes)
      setLimits(limitsRes)
      setTimeseries(Array.isArray(tsRes) ? tsRes : [])

      if (limitsRes) {
        setDailyLimit(limitsRes.daily_limit != null ? String(limitsRes.daily_limit) : '')
        setMonthlyLimit(limitsRes.monthly_limit != null ? String(limitsRes.monthly_limit) : '')
        setDailyTokenLimit(limitsRes.daily_token_limit != null ? String(limitsRes.daily_token_limit) : '')
        setMonthlyTokenLimit(limitsRes.monthly_token_limit != null ? String(limitsRes.monthly_token_limit) : '')
      }
    } catch {
      // Silently handle
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { loadData() }, [loadData])

  const handleSave = async () => {
    setSaving(true)
    setMessage(null)
    try {
      await setSpendLimits({
        daily_limit: dailyLimit ? parseFloat(dailyLimit) : null,
        monthly_limit: monthlyLimit ? parseFloat(monthlyLimit) : null,
        daily_token_limit: dailyTokenLimit ? parseInt(dailyTokenLimit) : null,
        monthly_token_limit: monthlyTokenLimit ? parseInt(monthlyTokenLimit) : null,
      })
      setMessage({ type: 'success', text: 'Spend limits updated successfully' })
      await loadData()
    } catch (err: any) {
      setMessage({ type: 'error', text: err.message || 'Failed to update limits' })
    } finally {
      setSaving(false)
    }
  }

  const clearLimits = async () => {
    setSaving(true)
    setMessage(null)
    try {
      await setSpendLimits({
        daily_limit: null,
        monthly_limit: null,
        daily_token_limit: null,
        monthly_token_limit: null,
      })
      setDailyLimit('')
      setMonthlyLimit('')
      setDailyTokenLimit('')
      setMonthlyTokenLimit('')
      setMessage({ type: 'success', text: 'All spend limits removed' })
      await loadData()
    } catch (err: any) {
      setMessage({ type: 'error', text: err.message || 'Failed to clear limits' })
    } finally {
      setSaving(false)
    }
  }

  // Simple bar chart renderer for timeseries
  const maxSpend = Math.max(...timeseries.map(t => t.spend), 0.01)

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="animate-spin text-memzent-glow" size={32} />
      </div>
    )
  }

  return (
    <div className="space-y-10 pb-20">
      <header>
        <h1 className="text-4xl font-black tracking-tighter text-white mb-2 uppercase italic">
          Spend Limits & Budget
        </h1>
        <p className="text-white/40 font-black uppercase tracking-[0.3em] text-[10px] italic">
          Configure spending caps and monitor burn rate
        </p>
      </header>

      {/* Budget Overview Cards */}
      {budget && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="stat-card p-5 border-white/5">
            <div className="flex items-center gap-2 mb-3">
              <DollarSign size={14} className="text-memzent-glow" />
              <span className="text-[10px] font-black uppercase tracking-widest text-white/30">Balance</span>
            </div>
            <div className="text-2xl font-black text-white">${budget.balance?.toFixed(2) ?? '0.00'}</div>
            <div className="text-[10px] text-white/30 mt-1">
              {budget.projected_days > 0 ? `~${Math.round(budget.projected_days)} days remaining` : 'No usage data'}
            </div>
          </div>

          <div className="stat-card p-5 border-white/5">
            <div className="flex items-center gap-2 mb-3">
              <TrendingUp size={14} className="text-yellow-400" />
              <span className="text-[10px] font-black uppercase tracking-widest text-white/30">Burn Rate (24h)</span>
            </div>
            <div className="text-2xl font-black text-white">${budget.burn_rate_24h?.toFixed(4) ?? '0.00'}</div>
            <div className="text-[10px] text-white/30 mt-1">per day</div>
          </div>

          <div className="stat-card p-5 border-white/5">
            <div className="flex items-center gap-2 mb-3">
              <Gauge size={14} className="text-blue-400" />
              <span className="text-[10px] font-black uppercase tracking-widest text-white/30">Spend (7d)</span>
            </div>
            <div className="text-2xl font-black text-white">${budget.spend_7d?.toFixed(4) ?? '0.00'}</div>
          </div>

          <div className="stat-card p-5 border-white/5">
            <div className="flex items-center gap-2 mb-3">
              <Gauge size={14} className="text-purple-400" />
              <span className="text-[10px] font-black uppercase tracking-widest text-white/30">Spend (30d)</span>
            </div>
            <div className="text-2xl font-black text-white">${budget.spend_30d?.toFixed(4) ?? '0.00'}</div>
          </div>
        </div>
      )}

      {/* Provider Breakdown */}
      {budget?.provider_breakdown && Object.keys(budget.provider_breakdown).length > 0 && (
        <div className="stat-card p-6 border-white/5">
          <h3 className="text-xs font-black uppercase tracking-widest text-white/40 mb-4">Provider Breakdown (30d)</h3>
          <div className="space-y-3">
            {Object.entries(budget.provider_breakdown).map(([provider, spend]) => {
              const pct = budget.spend_30d > 0 ? ((spend as number) / budget.spend_30d) * 100 : 0
              return (
                <div key={provider} className="flex items-center gap-4">
                  <span className="text-xs font-bold text-white/60 w-24 truncate">{provider}</span>
                  <div className="flex-1 h-2 bg-white/5 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-memzent-glow/60 rounded-full transition-all"
                      style={{ width: `${Math.max(pct, 1)}%` }}
                    />
                  </div>
                  <span className="text-xs font-bold text-white/40 w-20 text-right">
                    ${(spend as number).toFixed(4)}
                  </span>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* Spend Timeseries Chart */}
      {timeseries.length > 0 && (
        <div className="stat-card p-6 border-white/5">
          <h3 className="text-xs font-black uppercase tracking-widest text-white/40 mb-4">Daily Spend (Last 30 Days)</h3>
          <div className="flex items-end gap-[2px] h-32">
            {timeseries.map((point, i) => (
              <div
                key={i}
                className="flex-1 bg-memzent-glow/30 hover:bg-memzent-glow/60 rounded-t transition-colors group relative"
                style={{ height: `${Math.max((point.spend / maxSpend) * 100, 2)}%` }}
              >
                <div className="absolute bottom-full mb-2 left-1/2 -translate-x-1/2 hidden group-hover:block z-10">
                  <div className="bg-black/90 border border-white/10 rounded-lg px-2 py-1 text-[9px] text-white/80 whitespace-nowrap font-bold">
                    {new Date(point.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                    <br />${point.spend.toFixed(4)}
                  </div>
                </div>
              </div>
            ))}
          </div>
          <div className="flex justify-between mt-2 text-[9px] text-white/20 font-bold">
            <span>{timeseries.length > 0 ? new Date(timeseries[0].date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) : ''}</span>
            <span>{timeseries.length > 0 ? new Date(timeseries[timeseries.length - 1].date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) : ''}</span>
          </div>
        </div>
      )}

      {/* Current Spend vs Limits */}
      {limits && (
        <div className="stat-card p-6 border-white/5">
          <h3 className="text-xs font-black uppercase tracking-widest text-white/40 mb-4 flex items-center gap-2">
            <Shield size={12} className="text-memzent-glow" />
            Current Spend vs Limits
          </h3>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {[
              { label: 'Daily Dollar', current: limits.daily_spend, limit: limits.daily_limit, unit: '$' },
              { label: 'Monthly Dollar', current: limits.monthly_spend, limit: limits.monthly_limit, unit: '$' },
              { label: 'Daily Tokens', current: limits.daily_tokens, limit: limits.daily_token_limit, unit: '' },
              { label: 'Monthly Tokens', current: limits.monthly_tokens, limit: limits.monthly_token_limit, unit: '' },
            ].map(item => {
              const pct = item.limit ? (item.current / item.limit) * 100 : 0
              const isOver = pct > 90
              return (
                <div key={item.label} className="p-4 rounded-xl bg-white/[0.02] border border-white/5">
                  <div className="flex justify-between mb-2">
                    <span className="text-[10px] font-black uppercase tracking-widest text-white/40">{item.label}</span>
                    {isOver && <AlertTriangle size={12} className="text-yellow-400" />}
                  </div>
                  <div className="text-sm font-bold text-white/70">
                    {item.unit}{item.unit === '$' ? item.current?.toFixed(4) : Math.round(item.current || 0).toLocaleString()}
                    <span className="text-white/30"> / </span>
                    {item.limit != null
                      ? `${item.unit}${item.unit === '$' ? item.limit.toFixed(2) : item.limit.toLocaleString()}`
                      : <span className="text-white/20 italic">No limit</span>
                    }
                  </div>
                  {item.limit != null && (
                    <div className="h-1.5 bg-white/5 rounded-full overflow-hidden mt-2">
                      <div
                        className={`h-full rounded-full transition-all ${isOver ? 'bg-red-500' : pct > 70 ? 'bg-yellow-500' : 'bg-memzent-glow'}`}
                        style={{ width: `${Math.min(pct, 100)}%` }}
                      />
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* Configure Limits Form */}
      <div className="stat-card p-6 border-white/5">
        <h3 className="text-xs font-black uppercase tracking-widest text-white/40 mb-6 flex items-center gap-2">
          <Shield size={12} className="text-memzent-glow" />
          Configure Spend Limits
        </h3>

        {message && (
          <div className={`mb-4 p-3 rounded-xl text-xs font-bold ${message.type === 'success' ? 'bg-green-500/10 text-green-400 border border-green-500/20' : 'bg-red-500/10 text-red-400 border border-red-500/20'}`}>
            {message.text}
          </div>
        )}

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
          <div>
            <label className="text-[10px] font-black uppercase tracking-widest text-white/30 block mb-2">
              Daily Dollar Limit
            </label>
            <div className="relative">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-white/30 text-sm">$</span>
              <input
                type="number"
                step="0.01"
                value={dailyLimit}
                onChange={(e) => setDailyLimit(e.target.value)}
                placeholder="No limit"
                className="w-full pl-7 pr-4 py-2.5 rounded-xl bg-white/5 border border-white/10 text-sm text-white placeholder:text-white/20 focus:border-memzent-glow/30 focus:outline-none transition-colors"
              />
            </div>
          </div>

          <div>
            <label className="text-[10px] font-black uppercase tracking-widest text-white/30 block mb-2">
              Monthly Dollar Limit
            </label>
            <div className="relative">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-white/30 text-sm">$</span>
              <input
                type="number"
                step="0.01"
                value={monthlyLimit}
                onChange={(e) => setMonthlyLimit(e.target.value)}
                placeholder="No limit"
                className="w-full pl-7 pr-4 py-2.5 rounded-xl bg-white/5 border border-white/10 text-sm text-white placeholder:text-white/20 focus:border-memzent-glow/30 focus:outline-none transition-colors"
              />
            </div>
          </div>

          <div>
            <label className="text-[10px] font-black uppercase tracking-widest text-white/30 block mb-2">
              Daily Token Limit
            </label>
            <input
              type="number"
              step="1000"
              value={dailyTokenLimit}
              onChange={(e) => setDailyTokenLimit(e.target.value)}
              placeholder="No limit"
              className="w-full px-4 py-2.5 rounded-xl bg-white/5 border border-white/10 text-sm text-white placeholder:text-white/20 focus:border-memzent-glow/30 focus:outline-none transition-colors"
            />
          </div>

          <div>
            <label className="text-[10px] font-black uppercase tracking-widest text-white/30 block mb-2">
              Monthly Token Limit
            </label>
            <input
              type="number"
              step="10000"
              value={monthlyTokenLimit}
              onChange={(e) => setMonthlyTokenLimit(e.target.value)}
              placeholder="No limit"
              className="w-full px-4 py-2.5 rounded-xl bg-white/5 border border-white/10 text-sm text-white placeholder:text-white/20 focus:border-memzent-glow/30 focus:outline-none transition-colors"
            />
          </div>
        </div>

        <p className="text-[10px] text-white/20 mt-3 italic">
          Leave empty to remove a limit. Daily resets at midnight UTC, monthly on the 1st.
        </p>

        <div className="flex gap-3 mt-6">
          <button
            onClick={handleSave}
            disabled={saving}
            className="flex items-center gap-2 px-6 py-2.5 rounded-xl bg-memzent-glow text-black text-xs font-black uppercase tracking-widest hover:shadow-[0_0_20px_rgba(0,243,255,0.3)] transition-all disabled:opacity-50"
          >
            {saving ? <Loader2 size={14} className="animate-spin" /> : <Save size={14} />}
            Save Limits
          </button>
          <button
            onClick={clearLimits}
            disabled={saving}
            className="flex items-center gap-2 px-6 py-2.5 rounded-xl bg-white/5 border border-white/10 text-xs font-black uppercase tracking-widest text-white/60 hover:text-white hover:bg-white/10 transition-all disabled:opacity-50"
          >
            <RotateCcw size={14} />
            Clear All
          </button>
        </div>
      </div>
    </div>
  )
}
