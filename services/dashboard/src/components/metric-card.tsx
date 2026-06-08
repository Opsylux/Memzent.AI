"use client"

import { ReactNode } from "react";
import { motion } from "framer-motion";
import { TrendingUp, TrendingDown } from "lucide-react";

interface MetricCardProps {
  label: string;
  value: string;
  trend?: string;
  trendDirection?: 'up' | 'down';
  icon: ReactNode;
  color: string;
  detail: string;
  /** Optional 0–100 progress ring (e.g. cache hit rate) */
  ringPercent?: number;
  ringSlot?: ReactNode;
}

export function MetricCard({
  label,
  value,
  trend,
  trendDirection,
  icon,
  color,
  detail,
  ringSlot,
}: MetricCardProps) {
  return (
    <motion.div
      whileHover={{ y: -5 }}
      className={`stat-card relative overflow-hidden group ${color === 'cyan' ? 'glow-cyan' : 'glow-purple'}`}
    >
      <div className="flex justify-between items-start mb-6">
        <div className={`w-12 h-12 rounded-2xl flex items-center justify-center bg-white/5 border border-white/10 ${color === 'cyan' ? 'text-memzent-glow' : 'text-memzent-purple'
          }`}>
          {icon}
        </div>
        <div className="flex items-center gap-2">
          {ringSlot}
        {trend && (
          <div className={`flex items-center gap-1 text-[10px] font-black px-2 py-1 rounded-lg border ${trendDirection === 'up' ? "bg-memzent-accent/10 text-memzent-accent border-memzent-accent/20" : "bg-red-500/10 text-red-500 border-red-500/20"
            }`}>
            {trendDirection === 'up' ? <TrendingUp size={12} /> : <TrendingDown size={12} />}
            {trend}
          </div>
        )}
        </div>
      </div>

      <div className="space-y-1">
        <h3 className="text-xs font-bold text-readable-label uppercase tracking-wider leading-none">{label}</h3>
        <div className="text-3xl font-black tracking-tight text-readable-primary mt-1">{value}</div>
      </div>

      <div className="mt-6 flex items-center justify-between">
        <span className="text-[11px] font-medium text-readable-muted">{detail}</span>
        <div className={`w-1 h-1 rounded-full ${color === 'cyan' ? 'bg-memzent-glow shadow-[0_0_10px_#00f3ff]' : 'bg-memzent-purple shadow-[0_0_10px_#9d00ff]'}`} />
      </div>

      {/* Background Glow */}
      <div className={`absolute -right-8 -bottom-8 w-32 h-32 rounded-full blur-[60px] opacity-10 group-hover:opacity-20 transition-opacity ${color === 'cyan' ? 'bg-memzent-glow' : 'bg-memzent-purple'
        }`} />
    </motion.div>
  );
}
