import React, { useState } from 'react';
import { PortfolioDTO } from '../types/api';
import { XCircle } from 'lucide-react';

interface PortfolioViewProps {
  portfolio: PortfolioDTO | null;
  onClosePosition: (symbol: string) => Promise<void>;
  onSelectStock: (symbol: string) => void;
}

export const PortfolioView: React.FC<PortfolioViewProps> = ({
  portfolio,
  onClosePosition,
  onSelectStock,
}) => {
  const [closing, setClosing] = useState<string | null>(null);

  const formatB = (val: number) => {
    return `$${val.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
  };

  const handleClose = async (symbol: string) => {
    setClosing(symbol);
    try {
      await onClosePosition(symbol);
    } finally {
      setClosing(null);
    }
  };

  if (!portfolio) {
    return (
      <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-6 text-center text-[#807d72] dark:text-[#a09c92] shadow-none">
        Loading trader portfolio telemetry...
      </div>
    );
  }

  const isMarginWarning = portfolio.margin_ratio > 0.70;

  return (
    <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5 space-y-4 shadow-none">
      {/* Balances Telemetry Strip */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-3 pb-4 border-b border-[#efeee8] dark:border-[#262520]">
        <div>
          <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#a09c92]">Net Equity</div>
          <div className="text-[18px] font-mono font-semibold text-[#26251e] dark:text-[#edece6] mt-0.5">
            {formatB(portfolio.equity)}
          </div>
        </div>
        <div>
          <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#a09c92]">Cash Balance</div>
          <div className="text-[18px] font-mono text-[#26251e] dark:text-[#edece6] mt-0.5">
            {formatB(portfolio.cash)}
          </div>
        </div>
        <div>
          <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#a09c92]">Buying Power</div>
          <div className="text-[18px] font-mono text-[#26251e] dark:text-[#edece6] mt-0.5">
            {formatB(portfolio.buying_power)}
          </div>
        </div>
        <div>
          <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#a09c92]">Margin Utilized</div>
          <div className={`text-[18px] font-mono font-medium mt-0.5 ${isMarginWarning ? 'text-[#cf2d56]' : 'text-[#26251e] dark:text-[#edece6]'}`}>
            {(portfolio.margin_ratio * 100).toFixed(1)}%
          </div>
        </div>
        <div>
          <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#a09c92]">Max Leverage</div>
          <div className="text-[18px] font-mono text-[#807d72] dark:text-[#a09c92] mt-0.5">
            {portfolio.max_leverage.toFixed(0)}x (Reg T)
          </div>
        </div>
      </div>

      {/* Positions Table */}
      <div>
        <div className="text-[13px] font-semibold text-[#26251e] dark:text-[#edece6] mb-2 flex items-center justify-between">
          <span>Active Open Positions ({portfolio.positions.length})</span>
          <span className="text-[11px] font-mono text-[#807d72] dark:text-[#a09c92]">Real-time Mark-to-Market</span>
        </div>

        {portfolio.positions.length === 0 ? (
          <div className="py-8 text-center text-[#807d72] dark:text-[#a09c92] text-[13px] bg-[#fafaf7] dark:bg-[#262520] rounded-[8px] border border-dashed border-[#e6e5e0] dark:border-[#2c2b26]">
            No open market positions. Use the order ticket above to open long or short exposure.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-[12px] border-collapse">
              <thead>
                <tr className="border-b border-[#efeee8] dark:border-[#262520] text-[10px] font-mono uppercase text-[#807d72] dark:text-[#a09c92]">
                  <th className="py-2 px-2">Symbol</th>
                  <th className="py-2 px-2">Side</th>
                  <th className="py-2 px-2 text-right">Shares</th>
                  <th className="py-2 px-2 text-right">Entry</th>
                  <th className="py-2 px-2 text-right">Current</th>
                  <th className="py-2 px-2 text-right">Value</th>
                  <th className="py-2 px-2 text-right">Unrealized PnL</th>
                  <th className="py-2 px-2 text-right">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#efeee8] dark:divide-[#262520]">
                {portfolio.positions.map((pos) => {
                  const isProfit = pos.unrealized_pnl >= 0;
                  return (
                    <tr
                      key={pos.symbol}
                      onClick={() => onSelectStock(pos.symbol)}
                      className="hover:bg-[#fafaf7] dark:hover:bg-[#262520] cursor-pointer transition-colors"
                    >
                      <td className="py-2 px-2 font-mono font-semibold text-[#26251e] dark:text-[#edece6]">
                        {pos.symbol}
                      </td>
                      <td className="py-2 px-2">
                        <span
                          className={`text-[10px] font-mono uppercase px-1.5 py-0.5 rounded-[4px] ${
                            pos.side === 'LONG'
                              ? 'bg-[#1f8a65]/15 text-[#1f8a65]'
                              : 'bg-[#cf2d56]/15 text-[#cf2d56]'
                          }`}
                        >
                          {pos.side}
                        </span>
                      </td>
                      <td className="py-2 px-2 font-mono text-right text-[#26251e] dark:text-[#edece6]">
                        {pos.quantity.toLocaleString()}
                      </td>
                      <td className="py-2 px-2 font-mono text-right text-[#5a5852] dark:text-[#a09c92]">
                        ${pos.entry_price.toFixed(2)}
                      </td>
                      <td className="py-2 px-2 font-mono text-right text-[#26251e] dark:text-[#edece6] font-medium">
                        ${pos.current_price.toFixed(2)}
                      </td>
                      <td className="py-2 px-2 font-mono text-right text-[#26251e] dark:text-[#edece6]">
                        {formatB(pos.value)}
                      </td>
                      <td className={`py-2 px-2 font-mono text-right font-medium ${isProfit ? 'text-[#1f8a65]' : 'text-[#cf2d56]'}`}>
                        {isProfit ? '+' : ''}{formatB(pos.unrealized_pnl)} ({pos.pnl_percent >= 0 ? '+' : ''}{pos.pnl_percent.toFixed(2)}%)
                      </td>
                      <td className="py-2 px-2 text-right">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleClose(pos.symbol);
                          }}
                          disabled={closing === pos.symbol}
                          className="px-2 py-1 text-[11px] font-medium rounded-[6px] border border-[#e6e5e0] dark:border-[#2c2b26] hover:border-[#cf2d56] hover:text-[#cf2d56] text-[#5a5852] dark:text-[#a09c92] transition-colors"
                        >
                          {closing === pos.symbol ? 'Closing...' : 'Close'}
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};
