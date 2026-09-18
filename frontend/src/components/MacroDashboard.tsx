import React from 'react';
import { Landmark, TrendingUp, DollarSign, Users, Award, LineChart, Palmtree, Wrench } from 'lucide-react';
import { MacroDTO } from '../types/api';

interface MacroDashboardProps {
  macro: MacroDTO | null;
}

export const MacroDashboard: React.FC<MacroDashboardProps> = ({ macro }) => {
  if (!macro) {
    return (
      <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-8 text-center text-[#807d72] dark:text-[#78756c]">
        Loading macroeconomic data...
      </div>
    );
  }

  const formatCurrency = (val: number) => {
    const abs = Math.abs(val);
    const sign = val < 0 ? '-' : '';
    if (abs >= 1e12) return `${sign}$${(abs / 1e12).toFixed(2)}T`;
    if (abs >= 1e9) return `${sign}$${(abs / 1e9).toFixed(2)}B`;
    if (abs >= 1e6) return `${sign}$${(abs / 1e6).toFixed(1)}M`;
    return `${sign}$${abs.toFixed(2)}`;
  };

  const bonds = macro.bonds || [];
  const y1 = bonds.find((b) => b.maturity_years === 1)?.yield_to_maturity ?? (macro.borrowing_yield - 0.005);
  const y5 = bonds.find((b) => b.maturity_years === 5)?.yield_to_maturity ?? macro.borrowing_yield;
  const y10 = bonds.find((b) => b.maturity_years === 10)?.yield_to_maturity ?? (macro.borrowing_yield + 0.008);
  const y30 = bonds.find((b) => b.maturity_years === 30)?.yield_to_maturity ?? (macro.borrowing_yield + 0.018);

  const yieldPoints = [
    { tenor: '1Y T-Bill', y: y1 * 100, x: 10 },
    { tenor: '5Y Note', y: y5 * 100, x: 35 },
    { tenor: '10Y Benchmark', y: y10 * 100, x: 65 },
    { tenor: '30Y Bond', y: y30 * 100, x: 95 },
  ];

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <div className="text-[11px] font-mono uppercase tracking-[0.08em] text-[#807d72] dark:text-[#a09c92] mb-1">
            Monetary & Fiscal Indicators
          </div>
          <h1 className="text-[28px] lg:text-[32px] font-normal tracking-[-0.03em] text-[#26251e] dark:text-[#edece6] leading-[1.2]">
            Macroeconomic Dashboard
          </h1>
          <p className="text-[14px] lg:text-[15px] text-[#5a5852] dark:text-[#a09c92] mt-1 max-w-3xl leading-[1.5]">
            Monetary policy rate setting, sovereign bond yield curve, inflation trends, and labor market indicators.
          </p>
        </div>

        {/* Status Tier Badge */}
        <div className="flex items-center gap-2 bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] px-4 py-2 rounded-[10px] self-start md:self-auto">
          <Award size={18} className="text-[#f54e00]" />
          <div>
            <div className="text-[10px] font-mono uppercase text-[#807d72] dark:text-[#78756c]">Development Tier</div>
            <div className="text-[13px] font-semibold text-[#26251e] dark:text-[#edece6]">{macro.tier_name || 'Developing Nation'}</div>
          </div>
        </div>
      </div>

      {/* 4 Primary Macro Pillar Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
        {/* Central Bank Pillar */}
        <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5 space-y-3">
          <div className="flex items-center gap-2 text-[#26251e] dark:text-[#edece6]">
            <Landmark size={18} className="text-[#f54e00]" />
            <h2 className="text-[15px] font-semibold tracking-tight">Central Bank</h2>
          </div>
          <div>
            <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#78756c]">Target Rate</div>
            <div className="text-[24px] font-mono text-[#26251e] dark:text-[#edece6] font-semibold mt-0.5">
              {(macro.central_bank_rate * 100).toFixed(2)}%
            </div>
          </div>
          <div className="text-[12px] text-[#5a5852] dark:text-[#a09c92] space-y-1 pt-2 border-t border-[#efeee8] dark:border-[#262520]">
            <div className="flex justify-between">
              <span>Stance:</span>
              <span className="font-medium text-[#26251e] dark:text-[#edece6]">{macro.policy_stance}</span>
            </div>
            <div className="flex justify-between">
              <span>CPI Inflation:</span>
              <span className="font-mono text-[#26251e] dark:text-[#edece6]">{(macro.inflation_rate * 100).toFixed(2)}%</span>
            </div>
            <div className="flex justify-between">
              <span>10Y Bond Yield:</span>
              <span className="font-mono text-[#26251e] dark:text-[#edece6]">{(macro.ten_year_yield * 100).toFixed(2)}%</span>
            </div>
          </div>
        </div>

        {/* Fiscal Pillar */}
        <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5 space-y-3">
          <div className="flex items-center gap-2 text-[#26251e] dark:text-[#edece6]">
            <DollarSign size={18} className="text-[#f54e00]" />
            <h2 className="text-[15px] font-semibold tracking-tight">Treasury & Debt</h2>
          </div>
          <div>
            <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#78756c]">National Debt</div>
            <div className="text-[24px] font-mono text-[#26251e] dark:text-[#edece6] font-semibold mt-0.5">
              {formatCurrency(macro.national_debt)}
            </div>
          </div>
          <div className="text-[12px] text-[#5a5852] dark:text-[#a09c92] space-y-1 pt-2 border-t border-[#efeee8] dark:border-[#262520]">
            <div className="flex justify-between">
              <span>Credit Rating:</span>
              <span className="font-mono font-semibold text-[#26251e] dark:text-[#edece6]">{macro.credit_rating}</span>
            </div>
            <div className="flex justify-between">
              <span>Budget Balance:</span>
              <span className={`font-mono ${macro.budget_deficit > 0 ? 'text-[#cf2d56]' : 'text-[#1f8a65]'}`}>
                {macro.budget_deficit > 0 ? `-${formatCurrency(macro.budget_deficit)}` : `+${formatCurrency(-macro.budget_deficit)}`}
              </span>
            </div>
            <div className="flex justify-between">
              <span>Treasury Cash:</span>
              <span className="font-mono text-[#26251e] dark:text-[#edece6]">{formatCurrency(macro.treasury_cash)}</span>
            </div>
          </div>
        </div>

        {/* Labor Market Pillar */}
        <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5 space-y-3">
          <div className="flex items-center gap-2 text-[#26251e] dark:text-[#edece6]">
            <Users size={18} className="text-[#f54e00]" />
            <h2 className="text-[15px] font-semibold tracking-tight">Labor Market</h2>
          </div>
          <div>
            <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#78756c]">Unemployment Rate</div>
            <div className="text-[24px] font-mono text-[#26251e] dark:text-[#edece6] font-semibold mt-0.5">
              {(macro.unemployment_rate * 100).toFixed(1)}%
            </div>
          </div>
          <div className="text-[12px] text-[#5a5852] dark:text-[#a09c92] space-y-1 pt-2 border-t border-[#efeee8] dark:border-[#262520]">
            <div className="flex justify-between">
              <span>Employed Workers:</span>
              <span className="font-mono text-[#26251e] dark:text-[#edece6]">{(macro.employed_workers / 1e6).toFixed(1)}M</span>
            </div>
            <div className="flex justify-between">
              <span>Avg Hourly Wage:</span>
              <span className="font-mono text-[#26251e] dark:text-[#edece6]">${macro.average_hourly_wage.toFixed(2)}/hr</span>
            </div>
            <div className="flex justify-between">
              <span>Labor Force:</span>
              <span className="font-mono text-[#26251e] dark:text-[#edece6]">{(macro.labor_force / 1e6).toFixed(1)}M</span>
            </div>
          </div>
        </div>

        {/* Consumer Sentiment Pillar */}
        <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5 space-y-3">
          <div className="flex items-center gap-2 text-[#26251e] dark:text-[#edece6]">
            <TrendingUp size={18} className="text-[#f54e00]" />
            <h2 className="text-[15px] font-semibold tracking-tight">Consumer Activity</h2>
          </div>
          <div>
            <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#78756c]">Confidence Index</div>
            <div className="text-[24px] font-mono text-[#26251e] dark:text-[#edece6] font-semibold mt-0.5">
              {macro.consumer_sentiment.toFixed(1)}
            </div>
          </div>
          <div className="text-[12px] text-[#5a5852] dark:text-[#a09c92] space-y-1 pt-2 border-t border-[#efeee8] dark:border-[#262520]">
            <div className="flex justify-between">
              <span>Consumer Spending:</span>
              <span className="font-mono text-[#26251e] dark:text-[#edece6]">{formatCurrency(macro.consumer_spending)}</span>
            </div>
            <div className="flex justify-between">
              <span>Currency Index:</span>
              <span className="font-mono text-[#26251e] dark:text-[#edece6]">{macro.dxy_index.toFixed(2)}</span>
            </div>
            <div className="flex justify-between">
              <span>Trade Balance:</span>
              <span className="font-mono text-[#26251e] dark:text-[#edece6]">{formatCurrency(macro.trade_balance)}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Yield Curve */}
      <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-6 space-y-4">
        <div className="flex items-center justify-between pb-3 border-b border-[#efeee8] dark:border-[#262520]">
          <div className="flex items-center gap-2">
            <LineChart size={18} className="text-[#f54e00]" />
            <h2 className="text-[18px] font-semibold text-[#26251e] dark:text-[#edece6] tracking-tight">
              Sovereign Bond Yield Curve
            </h2>
          </div>
          <span className="text-[11px] font-mono px-2 py-0.5 rounded-[4px] bg-[#eef2f6] dark:bg-[#152438] text-[#1976d2] dark:text-[#63b3ed] border border-[#dbeafe] dark:border-[#1e3a5f]">
            Credit Spread: {macro.credit_rating}
          </span>
        </div>

        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          {yieldPoints.map((pt) => (
            <div key={pt.tenor} className="bg-[#fafaf7] dark:bg-[#151412] p-4 rounded-[8px] border border-[#efeee8] dark:border-[#262520] space-y-1 text-center">
              <div className="text-[11px] font-mono text-[#807d72] dark:text-[#78756c] uppercase">{pt.tenor}</div>
              <div className="text-[22px] font-mono font-semibold text-[#1976d2] dark:text-[#63b3ed]">{pt.y.toFixed(2)}%</div>
              <div className="text-[11px] text-[#5a5852] dark:text-[#a09c92]">Annual Yield (YTM)</div>
            </div>
          ))}
        </div>
      </div>

      {/* National Budget Breakdown */}
      <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-6">
        <h2 className="text-[18px] font-semibold text-[#26251e] dark:text-[#edece6] tracking-tight mb-1">
          National Budget & Receipts Statement
        </h2>
        <p className="text-[13px] text-[#807d72] dark:text-[#78756c] mb-4">
          Fiscal receipts versus public services, capital maintenance, and debt servicing outlays.
        </p>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          <div>
            <h3 className="text-[13px] font-mono uppercase text-[#807d72] dark:text-[#78756c] pb-2 border-b border-[#efeee8] dark:border-[#262520]">
              Revenue Receipts
            </h3>
            <div className="divide-y divide-[#efeee8] dark:divide-[#262520] text-[13px]">
              <div className="py-2.5 flex justify-between">
                <span className="text-[#5a5852] dark:text-[#a09c92]">Total Revenue</span>
                <span className="font-mono font-medium text-[#26251e] dark:text-[#edece6]">{formatCurrency(macro.federal_revenue)}</span>
              </div>
              <div className="py-2.5 flex justify-between">
                <span className="text-[#5a5852] dark:text-[#a09c92]">Corporate Income Taxes (18%)</span>
                <span className="font-mono text-[#5a5852] dark:text-[#a09c92]">{formatCurrency(macro.federal_revenue * 0.26)}</span>
              </div>
              <div className="py-2.5 flex justify-between">
                <span className="text-[#5a5852] dark:text-[#a09c92]">Personal Income & Payroll</span>
                <span className="font-mono text-[#5a5852] dark:text-[#a09c92]">{formatCurrency(macro.federal_revenue * 0.54)}</span>
              </div>
              <div className="py-2.5 flex justify-between">
                <span className="text-[#5a5852] dark:text-[#a09c92]">Tariffs & Customs</span>
                <span className="font-mono text-[#5a5852] dark:text-[#a09c92]">{formatCurrency(macro.federal_revenue * 0.10)}</span>
              </div>
              <div className="py-2.5 flex justify-between items-center bg-[#f1f8e9]/50 dark:bg-[#1a2e16]/50 -mx-2 px-2 rounded-[6px]">
                <span className="text-[#1f8a65] font-medium flex items-center gap-1.5">
                  <Palmtree size={14} />
                  <span>Tourism Inflows</span>
                </span>
                <span className="font-mono font-semibold text-[#1f8a65]">
                  {formatCurrency(macro.tourism_revenue || 65_000_000)}
                </span>
              </div>
            </div>
          </div>

          <div>
            <h3 className="text-[13px] font-mono uppercase text-[#807d72] dark:text-[#78756c] pb-2 border-b border-[#efeee8] dark:border-[#262520]">
              Budget Outlays (Expenditures)
            </h3>
            <div className="divide-y divide-[#efeee8] dark:divide-[#262520] text-[13px]">
              <div className="py-2.5 flex justify-between">
                <span className="text-[#5a5852] dark:text-[#a09c92]">Total Expenditures</span>
                <span className="font-mono font-medium text-[#26251e] dark:text-[#edece6]">{formatCurrency(macro.federal_outlays)}</span>
              </div>
              <div className="py-2.5 flex justify-between">
                <span className="text-[#5a5852] dark:text-[#a09c92]">Public Services & Healthcare</span>
                <span className="font-mono text-[#5a5852] dark:text-[#a09c92]">{formatCurrency(macro.federal_outlays * 0.50)}</span>
              </div>
              <div className="py-2.5 flex justify-between">
                <span className="text-[#5a5852] dark:text-[#a09c92]">Procurement & Capital Works</span>
                <span className="font-mono text-[#5a5852] dark:text-[#a09c92]">{formatCurrency(macro.federal_outlays * 0.25)}</span>
              </div>
              <div className="py-2.5 flex justify-between items-center bg-[#fff8e1]/50 dark:bg-[#332210]/50 -mx-2 px-2 rounded-[6px]">
                <span className="text-[#e65100] dark:text-[#ff9800] font-medium flex items-center gap-1.5">
                  <Wrench size={14} />
                  <span>Infrastructure Maintenance</span>
                </span>
                <span className="font-mono font-semibold text-[#e65100] dark:text-[#ff9800]">
                  {formatCurrency(macro.maintenance_spending_rate || 120_000_000)}
                </span>
              </div>
              <div className="py-2.5 flex justify-between">
                <span className="text-[#5a5852] dark:text-[#a09c92]">Net Interest on Debt</span>
                <span className="font-mono text-[#cf2d56]">{formatCurrency(macro.federal_outlays * 0.12)}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
