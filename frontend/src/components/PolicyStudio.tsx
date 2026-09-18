import React, { useState } from 'react';
import { PolicyItemDTO, MacroDTO } from '../types/api';
import { CheckCircle2, Circle, Lock, Award, TrendingUp, AlertCircle, ChevronDown, ChevronUp, Clock, AlertTriangle, ShieldAlert } from 'lucide-react';

interface PolicyStudioProps {
  policies: PolicyItemDTO[];
  macro: MacroDTO | null;
  onTogglePolicy: (id: string) => Promise<void>;
}

export const PolicyStudio: React.FC<PolicyStudioProps> = ({ policies, macro, onTogglePolicy }) => {
  const [togglingId, setTogglingId] = useState<string | null>(null);
  const [activeCategory, setActiveCategory] = useState<string>('ALL');
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const categories = [
    { id: 'ALL', label: 'All Categories' },
    { id: 'fiscal', label: 'Fiscal & Infra' },
    { id: 'monetary', label: 'Monetary & Banking' },
    { id: 'trade', label: 'Trade & Tariffs' },
    { id: 'industry', label: 'Industry & Tech' },
    { id: 'labor', label: 'Labor & Welfare' },
    { id: 'foreign', label: 'Foreign Affairs' },
  ];

  const filteredPolicies = policies.filter(
    (p) => activeCategory === 'ALL' || p.category.toLowerCase() === activeCategory.toLowerCase()
  );

  const handleToggle = async (policy: PolicyItemDTO) => {
    if (policy.is_locked) {
      setErrorMsg(`Cannot enact '${policy.name}': Requires country status tier '${formatTierName(policy.min_tier)}'.`);
      setTimeout(() => setErrorMsg(null), 4000);
      return;
    }
    setTogglingId(policy.id);
    setErrorMsg(null);
    try {
      await onTogglePolicy(policy.id);
    } catch (err: any) {
      setErrorMsg(err.message || 'Failed to toggle policy');
      setTimeout(() => setErrorMsg(null), 4000);
    } finally {
      setTogglingId(null);
    }
  };

  const formatCost = (cost: number) => {
    if (cost === 0) return '$0 (Neutral)';
    const isRevenue = cost < 0;
    const abs = Math.abs(cost);
    const formatted = abs >= 1e9 ? `$${(abs / 1e9).toFixed(1)}B` : `$${(abs / 1e6).toFixed(0)}M`;
    return isRevenue ? `+${formatted} (Rev)` : `-${formatted} (Cost)`;
  };

  const formatTierName = (tier?: string) => {
    switch (tier) {
      case 'powerhouse':
        return 'Global Powerhouse';
      case 'mid':
        return 'Emerging State';
      default:
        return 'Developing Nation';
    }
  };

  const currentTier = macro?.tier || 'small';
  const tierName = macro?.tier_name || 'Developing Nation';
  const tierProgress = macro?.tier_progress !== undefined ? macro.tier_progress : 25.0;

  const unlockedCount = policies.filter((p) => !p.is_locked).length;
  const activeCount = policies.filter((p) => p.active).length;

  return (
    <div className="space-y-6">
      {/* Editorial Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <div className="text-[11px] font-mono uppercase tracking-[0.08em] text-[#807d72] dark:text-[#a09c92] mb-1">
            National Policies · Executive Directives
          </div>
          <h1 className="text-[26px] font-normal tracking-[-0.025em] text-[#26251e] dark:text-[#edece6] leading-[1.25]">
            Economic Policy Studio
          </h1>
          <p className="text-[14px] text-[#5a5852] dark:text-[#a09c92] mt-1 max-w-2xl leading-[1.5]">
            Enact policies across fiscal, monetary, trade, industry, labor, and foreign affairs. Advanced policies unlock as national economic capacity expands.
          </p>
        </div>
      </div>

      {/* Country Status & Tier Progression Banner */}
      <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-[#efeee8] dark:border-[#262520]">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-[8px] bg-[#fafaf7] dark:bg-[#151412] border border-[#e6e5e0] dark:border-[#2c2b26] flex items-center justify-center text-[#f54e00]">
              <Award size={22} />
            </div>
            <div>
              <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#78756c]">Economic Development Tier</div>
              <div className="text-[18px] font-semibold text-[#26251e] dark:text-[#edece6] flex items-center gap-2">
                <span>{tierName}</span>
                <span
                  className={`text-[11px] font-mono uppercase px-2 py-0.5 rounded-[4px] font-medium ${
                    currentTier === 'powerhouse'
                      ? 'bg-[#6b46c1]/10 text-[#6b46c1]'
                      : currentTier === 'mid'
                      ? 'bg-[#1f8a65]/10 text-[#1f8a65]'
                      : 'bg-[#efeee8] dark:bg-[#262520] text-[#5a5852] dark:text-[#a09c92]'
                  }`}
                >
                  Tier {currentTier === 'powerhouse' ? 'III' : currentTier === 'mid' ? 'II' : 'I'}
                </span>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-4 text-[13px] font-mono">
            <div>
              <span className="text-[#807d72] dark:text-[#78756c]">Active Policies: </span>
              <span className="font-semibold text-[#1f8a65]">{activeCount}</span>
            </div>
            <div>
              <span className="text-[#807d72] dark:text-[#78756c]">Unlocked: </span>
              <span className="font-semibold text-[#26251e] dark:text-[#edece6]">{unlockedCount} / {policies.length}</span>
            </div>
          </div>
        </div>

        {/* Tier Evolution Progress Bar */}
        <div className="pt-4">
          <div className="flex items-center justify-between text-[12px] text-[#5a5852] dark:text-[#a09c92] mb-1.5">
            <span className="flex items-center gap-1">
              <TrendingUp size={13} className="text-[#f54e00]" />
              <span>
                {currentTier === 'small'
                  ? 'Target: Emerging State (GDP ≥ $60.0B)'
                  : currentTier === 'mid'
                  ? 'Target: Global Powerhouse (GDP ≥ $300.0B)'
                  : 'Apex Economic Status Reached'}
              </span>
            </span>
            <span className="font-mono font-medium text-[#26251e] dark:text-[#edece6]">{tierProgress.toFixed(1)}%</span>
          </div>
          <div className="w-full bg-[#efeee8] dark:bg-[#262520] h-2 rounded-full overflow-hidden">
            <div
              className="bg-[#26251e] dark:bg-[#edece6] h-full rounded-full transition-all duration-500"
              style={{ width: `${Math.min(100, Math.max(5, tierProgress))}%` }}
            />
          </div>
        </div>
      </div>

      {/* Category Filter Pills */}
      <div className="flex items-center gap-1.5 overflow-x-auto pb-1 scrollbar-none">
        {categories.map((cat) => (
          <button
            key={cat.id}
            onClick={() => setActiveCategory(cat.id)}
            className={`px-3 py-1.5 text-[12px] font-medium rounded-[8px] whitespace-nowrap transition-colors border ${
              activeCategory === cat.id
                ? 'bg-[#26251e] dark:bg-[#edece6] text-white dark:text-[#141412] border-[#26251e] dark:border-[#edece6]'
                : 'bg-white dark:bg-[#1c1b18] text-[#5a5852] dark:text-[#a09c92] border-[#e6e5e0] dark:border-[#2c2b26] hover:text-[#26251e] dark:hover:text-[#edece6] hover:bg-[#efeee8] dark:hover:bg-[#262520]'
            }`}
          >
            {cat.label}
          </button>
        ))}
      </div>

      {/* Error / Alert Message */}
      {errorMsg && (
        <div className="p-3 bg-[#cf2d56]/10 border border-[#cf2d56]/30 rounded-[8px] text-[12px] text-[#cf2d56] flex items-center gap-2">
          <AlertCircle size={15} />
          <span>{errorMsg}</span>
        </div>
      )}

      {/* Fiscal Debt & Rollout Reality Banner */}
      <div className="bg-[#fafaf7] dark:bg-[#151412] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[10px] p-4 text-[13px] flex items-start gap-3">
        <AlertTriangle size={18} className="text-[#f54e00] shrink-0 mt-0.5" />
        <div className="space-y-1">
          <div className="font-semibold text-[#26251e] dark:text-[#edece6]">
            Fiscal Mechanics: Inevitable Debt & Policy Rollout Dynamics
          </div>
          <p className="text-[#5a5852] dark:text-[#a09c92] text-[12px] leading-relaxed">
            All public programs draw directly from treasury cash reserves. When annual outlays exceed fiscal receipts and cash reserves fall below $500M, sovereign debt issues automatically to sustain national solvency, pushing borrowing yields higher. Policies require 30 to 120 calendar days to reach 100% full effect.
          </p>
        </div>
      </div>

      {/* Policy Catalog Table */}
      <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse min-w-[700px]">
            <thead>
              <tr className="border-b border-[#e6e5e0] dark:border-[#2c2b26] bg-[#fafaf7] dark:bg-[#151412] text-[11px] font-mono uppercase text-[#807d72] dark:text-[#78756c]">
                <th className="py-3 px-3 lg:px-4 text-center w-12">Status</th>
                <th className="py-3 px-3 lg:px-4">Policy Directive</th>
                <th className="py-3 px-3 lg:px-4">Tier</th>
                <th className="py-3 px-3 lg:px-4">Description</th>
                <th className="py-3 px-3 lg:px-4 hidden lg:table-cell">Favored Sector</th>
                <th className="py-3 px-3 lg:px-4 text-right">Annual Impact</th>
                <th className="py-3 px-3 lg:px-4 text-right">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#efeee8] dark:divide-[#262520] text-[13px]">
              {filteredPolicies.map((p) => {
                const isToggling = togglingId === p.id;
                const isLocked = p.is_locked;
                const isExpanded = expandedId === p.id;
                const daysActive = p.days_active ?? 0;
                const totalDays = p.rollout_days ?? p.rollout_days_total ?? 30;
                const efficacyPct = p.efficacy !== undefined 
                  ? Math.round(p.efficacy > 1 ? p.efficacy : p.efficacy * 100) 
                  : Math.round(p.rollout_progress_pct ?? ((daysActive / totalDays) * 100));

                return (
                  <React.Fragment key={p.id}>
                    <tr
                      className={`hover:bg-[#fafaf7] dark:hover:bg-[#151412] transition-colors cursor-pointer ${
                        p.active ? 'bg-[#fafaf7]/60 dark:bg-[#151412]/60' : isLocked ? 'opacity-65' : ''
                      }`}
                      onClick={() => setExpandedId(isExpanded ? null : p.id)}
                    >
                      <td className="py-3 px-3 lg:px-4 text-center" onClick={(e) => e.stopPropagation()}>
                        {isLocked ? (
                          <Lock size={16} className="text-[#a09c92] mx-auto" />
                        ) : p.active ? (
                          <CheckCircle2 size={18} className="text-[#1f8a65] mx-auto" />
                        ) : (
                          <Circle size={18} className="text-[#a09c92] mx-auto" />
                        )}
                      </td>
                      <td className="py-3 px-3 lg:px-4 font-medium text-[#26251e] dark:text-[#edece6]">
                        <div className="flex flex-col">
                          <div className="flex items-center gap-1.5">
                            <span>{p.name}</span>
                            {isExpanded ? (
                              <ChevronUp size={13} className="text-[#807d72] dark:text-[#a09c92]" />
                            ) : (
                              <ChevronDown size={13} className="text-[#807d72] dark:text-[#a09c92]" />
                            )}
                          </div>
                          <div className="flex items-center gap-2 mt-0.5">
                            <span className="text-[11px] text-[#807d72] dark:text-[#78756c] font-mono uppercase">
                              {p.category}
                            </span>
                            {p.active ? (
                              <span className="text-[10px] font-mono text-[#1f8a65]">
                                {efficacyPct >= 100
                                  ? '100% In Full Effect'
                                  : `Day ${daysActive}/${totalDays} (${efficacyPct}%)`}
                              </span>
                            ) : (
                              <span className="text-[10px] font-mono text-[#807d72] dark:text-[#78756c]">
                                {totalDays} days rollout
                              </span>
                            )}
                          </div>
                          {p.active && (
                            <div className="w-28 bg-[#efeee8] dark:bg-[#262520] h-1.5 rounded-full overflow-hidden mt-1.5">
                              <div
                                className="bg-[#1f8a65] h-full rounded-full transition-all"
                                style={{ width: `${Math.min(100, efficacyPct)}%` }}
                              />
                            </div>
                          )}
                        </div>
                      </td>
                      <td className="py-3 px-3 lg:px-4">
                        <span
                          className={`text-[11px] font-mono uppercase px-2 py-0.5 rounded-[4px] font-medium whitespace-nowrap ${
                            p.min_tier === 'powerhouse'
                              ? 'bg-[#6b46c1]/10 text-[#6b46c1]'
                              : p.min_tier === 'mid'
                              ? 'bg-[#2b6cb0]/10 text-[#2b6cb0]'
                              : 'bg-[#efeee8] dark:bg-[#262520] text-[#5a5852] dark:text-[#a09c92]'
                          }`}
                        >
                          {formatTierName(p.min_tier)}
                        </span>
                      </td>
                      <td className="py-3 px-3 lg:px-4 text-[#5a5852] dark:text-[#a09c92] max-w-[280px] leading-[1.4]">
                        {p.description}
                      </td>
                      <td className="py-3 px-3 lg:px-4 text-[#26251e] dark:text-[#edece6] hidden lg:table-cell whitespace-nowrap">
                        {p.favored_sector}
                      </td>
                      <td
                        className={`py-3 px-4 font-mono text-right whitespace-nowrap ${
                          p.annual_cost < 0 ? 'text-[#1f8a65]' : p.annual_cost > 0 ? 'text-[#cf2d56]' : 'text-[#807d72] dark:text-[#78756c]'
                        }`}
                      >
                        {formatCost(p.annual_cost)}
                      </td>
                      <td className="py-3 px-4 text-right whitespace-nowrap" onClick={(e) => e.stopPropagation()}>
                        <button
                          onClick={() => handleToggle(p)}
                          disabled={isToggling || isLocked}
                          title={
                            isLocked
                              ? `Locked: Requires ${formatTierName(p.min_tier)}`
                              : p.active
                              ? 'Click to repeal policy'
                              : 'Click to enact policy'
                          }
                          className={`px-3 py-1 text-[12px] font-medium rounded-[6px] border transition-colors ${
                            isLocked
                              ? 'bg-[#efeee8] dark:bg-[#262520] text-[#a09c92] border-[#e6e5e0] dark:border-[#2c2b26] cursor-not-allowed'
                              : p.active
                              ? 'bg-[#26251e] dark:bg-[#edece6] text-white dark:text-[#141412] border-[#26251e] hover:bg-[#403e33]'
                              : 'bg-white dark:bg-[#1c1b18] text-[#26251e] dark:text-[#edece6] border-[#e6e5e0] dark:border-[#2c2b26] hover:border-[#807d72]'
                          }`}
                        >
                          {isToggling
                            ? 'Updating...'
                            : isLocked
                            ? 'Locked'
                            : p.active
                            ? 'Active'
                            : 'Enact'}
                        </button>
                      </td>
                    </tr>

                    {/* Expandable Direct / Indirect Impact Pane */}
                    {isExpanded && (
                      <tr className="bg-[#fafaf7] dark:bg-[#151412]">
                        <td colSpan={7} className="py-4 px-6 border-b border-[#efeee8] dark:border-[#262520]">
                          <div className="space-y-3">
                            <div className="flex flex-wrap items-center justify-between gap-2 pb-2 border-b border-[#efeee8] dark:border-[#262520]">
                              <span className="font-mono text-[11px] uppercase tracking-wider text-[#807d72] dark:text-[#78756c]">
                                Policy Impact Matrix & Rollout Timeline
                              </span>
                              <span className="font-mono text-[11px] text-[#26251e] dark:text-[#edece6]">
                                Duration: {p.rollout_days_total || 30} days to 100% full effect
                                {p.active && ` (${Math.max(0, (p.rollout_days_total || 30) - (p.rollout_days || 0))} days remaining)`}
                              </span>
                            </div>

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                              {/* Direct Positive */}
                              <div className="p-3 rounded-[8px] bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] space-y-1">
                                <div className="flex items-center gap-1.5 text-[#1f8a65] text-[11px] font-mono font-semibold uppercase">
                                  <CheckCircle2 size={13} />
                                  <span>Direct Positive Effect</span>
                                </div>
                                <p className="text-[12px] text-[#26251e] dark:text-[#edece6] leading-relaxed">
                                  {p.direct_positive || 'Boosts domestic productivity and strengthens targeted industrial capacity.'}
                                </p>
                              </div>

                              {/* Direct Negative / Sovereign Debt */}
                              <div className="p-3 rounded-[8px] bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] space-y-1">
                                <div className="flex items-center gap-1.5 text-[#cf2d56] text-[11px] font-mono font-semibold uppercase">
                                  <AlertCircle size={13} />
                                  <span>Direct Negative Effect (Debt & Fiscal Drain)</span>
                                </div>
                                <p className="text-[12px] text-[#26251e] dark:text-[#edece6] leading-relaxed">
                                  {p.direct_negative || 'Increases federal budget deficits and forces inevitable sovereign borrowing if treasury drains.'}
                                </p>
                              </div>

                              {/* Indirect Positive */}
                              <div className="p-3 rounded-[8px] bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] space-y-1">
                                <div className="flex items-center gap-1.5 text-[#2b6cb0] text-[11px] font-mono font-semibold uppercase">
                                  <TrendingUp size={13} />
                                  <span>Indirect Positive Effect</span>
                                </div>
                                <p className="text-[12px] text-[#26251e] dark:text-[#edece6] leading-relaxed">
                                  {p.indirect_positive || 'Generates positive economic spillover across regional enterprise supply chains.'}
                                </p>
                              </div>

                              {/* Indirect Negative */}
                              <div className="p-3 rounded-[8px] bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] space-y-1">
                                <div className="flex items-center gap-1.5 text-[#d97706] text-[11px] font-mono font-semibold uppercase">
                                  <AlertTriangle size={13} />
                                  <span>Indirect Negative Effect</span>
                                </div>
                                <p className="text-[12px] text-[#26251e] dark:text-[#edece6] leading-relaxed">
                                  {p.indirect_negative || 'May introduce price rigidities, capacity bottlenecks, or slight consumer headwinds.'}
                                </p>
                              </div>
                            </div>
                          </div>
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
