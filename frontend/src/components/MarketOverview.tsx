import React from 'react';
import { TrendingUp, TrendingDown, ArrowUpRight, Zap, ShieldAlert, Award } from 'lucide-react';
import { MacroDTO, PriceItem } from '../types/api';

interface MarketOverviewProps {
  prices: PriceItem[];
  macro: MacroDTO | null;
  onSelectStock: (symbol: string) => void;
}

export const MarketOverview: React.FC<MarketOverviewProps> = ({
  prices,
  macro,
  onSelectStock,
}) => {
  // Compute Top Gainers & Losers based on mid vs reported_value or arbitrary spread
  const sortedByCap = [...prices].sort((a, b) => b.market_cap - a.market_cap);
  const topStocks = sortedByCap.slice(0, 5);

  const totalMarketCap = prices.reduce((acc, p) => acc + p.market_cap, 0);

  // Group by sector
  const sectorGroups = prices.reduce((acc, p) => {
    if (!acc[p.sector]) acc[p.sector] = [];
    acc[p.sector].push(p);
    return acc;
  }, {} as Record<string, PriceItem[]>);

  const formatB = (val: number) => {
    const isNeg = val < 0;
    const abs = Math.abs(val);
    const sign = isNeg ? '-$' : '$';
    if (abs >= 1e12) return `${sign}${(abs / 1e12).toFixed(2)}T`;
    if (abs >= 1e9) return `${sign}${(abs / 1e9).toFixed(2)}B`;
    if (abs >= 1e6) return `${sign}${(abs / 1e6).toFixed(1)}M`;
    return `${sign}${abs.toFixed(2)}`;
  };

  return (
    <div className="space-y-12">
      {/* Editorial Header */}
      <div>
        <div className="text-[11px] font-mono uppercase tracking-[0.08em] text-[#807d72] dark:text-[#a09c92] mb-1">
          Markets · Live Overview
        </div>
        <h1 className="text-[32px] lg:text-[36px] font-normal tracking-[-0.035em] text-[#26251e] dark:text-[#edece6] leading-[1.2]">
          Financial Markets Overview
        </h1>
        <p className="text-[15px] lg:text-[16px] text-[#5a5852] dark:text-[#a09c92] mt-2 max-w-3xl leading-[1.5]">
          Continuous double-auction exchange integrating institutional liquidity providers, treasury operations, and macroeconomic fundamentals.
        </p>
      </div>

      {/* Top Telemetry Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5">
          <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#78756c]">Total Market Cap</div>
          <div className="text-[26px] font-normal tracking-[-0.025em] text-[#26251e] dark:text-[#edece6] font-mono mt-1">
            {formatB(totalMarketCap)}
          </div>
          <div className="text-[13px] text-[#5a5852] dark:text-[#a09c92] mt-2 flex items-center gap-1">
            <span className="font-mono text-[#26251e] dark:text-[#edece6] font-medium">{prices.length}</span> listed entities across 11 sectors
          </div>
        </div>

        <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5">
          <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#78756c]">Central Bank Rate</div>
          <div className="text-[26px] font-normal tracking-[-0.025em] text-[#26251e] dark:text-[#edece6] font-mono mt-1">
            {macro ? `${(macro.central_bank_rate * 100).toFixed(2)}%` : '5.25%'}
          </div>
          <div className="text-[13px] text-[#5a5852] dark:text-[#a09c92] mt-2">
            Stance: <span className="font-medium text-[#26251e] dark:text-[#edece6]">{macro?.policy_stance || 'Neutral'}</span> (10Y: {macro ? `${(macro.ten_year_yield * 100).toFixed(2)}%` : '4.15%'})
          </div>
        </div>

        <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5">
          <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#78756c]">Consumer Sentiment</div>
          <div className="text-[26px] font-normal tracking-[-0.025em] text-[#26251e] dark:text-[#edece6] font-mono mt-1">
            {macro ? macro.consumer_sentiment.toFixed(1) : '82.4'}
          </div>
          <div className="text-[13px] text-[#5a5852] dark:text-[#a09c92] mt-2">
            Unemployment: <span className="font-mono text-[#26251e] dark:text-[#edece6] font-medium">{macro ? `${(macro.unemployment_rate * 100).toFixed(1)}%` : '3.8%'}</span>
          </div>
        </div>

        <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5">
          <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#78756c]">Credit Rating</div>
          <div className="text-[26px] font-normal tracking-[-0.025em] text-[#26251e] dark:text-[#edece6] font-mono mt-1 flex items-baseline gap-2">
            <span>{macro?.credit_rating || 'AA+'}</span>
            <span className="text-[13px] text-[#807d72] dark:text-[#78756c] font-normal">
              Yield: {macro ? `${(macro.borrowing_yield * 100).toFixed(2)}%` : '3.4%'}
            </span>
          </div>
          <div className="text-[13px] text-[#5a5852] dark:text-[#a09c92] mt-2">
            Budget Deficit: <span className="font-mono text-[#26251e] dark:text-[#edece6] font-medium">{macro ? formatB(macro.budget_deficit) : '$1.2T'}</span>
          </div>
        </div>
      </div>

      {/* Featured Market Leaders & Sector Distribution */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Mega-Caps Table */}
        <div className="lg:col-span-2 bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-6">
          <div className="flex items-center justify-between pb-4 border-b border-[#efeee8] dark:border-[#262520]">
            <div>
              <h2 className="text-[18px] font-semibold text-[#26251e] dark:text-[#edece6] tracking-tight">Market Leaders</h2>
              <p className="text-[13px] text-[#807d72] dark:text-[#78756c]">Highest capitalized equities driving trading volume and index performance.</p>
            </div>
            <button
              onClick={() => onSelectStock(topStocks[0]?.symbol || 'AAPL')}
              className="text-[13px] text-[#f54e00] hover:text-[#d04200] font-medium flex items-center gap-1"
            >
              <span>Trade Leader</span>
              <ArrowUpRight size={14} />
            </button>
          </div>

          <div className="divide-y divide-[#efeee8] dark:divide-[#262520] mt-2">
            {topStocks.map((stock) => (
              <div
                key={stock.symbol}
                onClick={() => onSelectStock(stock.symbol)}
                className="py-3 px-2 flex items-center justify-between hover:bg-[#fafaf7] dark:hover:bg-[#151412] rounded-[8px] cursor-pointer transition-colors"
              >
                <div className="flex items-center gap-3">
                  <div className="w-9 h-9 rounded-[8px] bg-[#efeee8] dark:bg-[#262520] flex items-center justify-center font-mono font-semibold text-[13px] text-[#26251e] dark:text-[#edece6]">
                    {stock.symbol.slice(0, 3)}
                  </div>
                  <div>
                    <div className="font-medium text-[14px] text-[#26251e] dark:text-[#edece6] flex items-center gap-2">
                      <span>{stock.name}</span>
                      <span className="font-mono text-[11px] text-[#807d72] dark:text-[#78756c] uppercase px-1.5 py-0.5 rounded-[4px] bg-[#efeee8] dark:bg-[#262520]">
                        {stock.symbol}
                      </span>
                    </div>
                    <div className="text-[12px] text-[#807d72] dark:text-[#78756c]">
                      {stock.sector} · {stock.cap_tier}
                    </div>
                  </div>
                </div>

                <div className="text-right">
                  <div className="font-mono font-medium text-[15px] text-[#26251e] dark:text-[#edece6]">
                    ${stock.mid > 0 ? stock.mid.toFixed(2) : stock.reported_value.toFixed(2)}
                  </div>
                  <div className="text-[12px] font-mono text-[#807d72] dark:text-[#78756c]">
                    Cap: {formatB(stock.market_cap)} · P/E: {stock.pe_ratio > 0 ? stock.pe_ratio.toFixed(1) : 'N/A'}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Sector Composition */}
        <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-6">
          <h2 className="text-[18px] font-semibold text-[#26251e] dark:text-[#edece6] tracking-tight">Sector Composition</h2>
          <p className="text-[13px] text-[#807d72] dark:text-[#78756c] mb-4">Distribution across industry sectors.</p>

          <div className="space-y-3">
            {Object.entries(sectorGroups).slice(0, 8).map(([sector, items]) => {
              const secCap = items.reduce((sum, item) => sum + item.market_cap, 0);
              const pct = totalMarketCap > 0 ? (secCap / totalMarketCap) * 100 : 0;
              return (
                <div key={sector}>
                  <div className="flex justify-between text-[13px] mb-1">
                    <span className="text-[#26251e] dark:text-[#edece6] font-medium">{sector}</span>
                    <span className="font-mono text-[#807d72] dark:text-[#78756c]">{pct.toFixed(1)}%</span>
                  </div>
                  <div className="w-full bg-[#efeee8] dark:bg-[#262520] h-1.5 rounded-full overflow-hidden">
                    <div
                      className="bg-[#26251e] dark:bg-[#edece6] h-full rounded-full transition-all duration-500"
                      style={{ width: `${Math.min(100, Math.max(2, pct))}%` }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
};
