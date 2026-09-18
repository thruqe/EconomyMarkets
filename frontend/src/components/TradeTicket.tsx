import React, { useState, useEffect } from 'react';
import { Settings, Sliders, DollarSign, Percent, ShieldCheck, Check, X } from 'lucide-react';
import { PortfolioDTO, PriceItem } from '../types/api';

interface TradeTicketProps {
  stock: PriceItem | null;
  portfolio: PortfolioDTO | null;
  initialPrice?: number;
  onSubmitOrder: (order: {
    symbol: string;
    side: 'Buy' | 'Sell';
    order_type: 'Market' | 'Limit';
    price: number;
    quantity: number;
  }) => Promise<void>;
  onUpdateSettings?: (data: { balance?: number; leverage?: number }) => Promise<void>;
}

export const TradeTicket: React.FC<TradeTicketProps> = ({
  stock,
  portfolio,
  initialPrice,
  onSubmitOrder,
  onUpdateSettings,
}) => {
  const [side, setSide] = useState<'Buy' | 'Sell'>('Buy');
  const [orderType, setOrderType] = useState<'Market' | 'Limit'>('Market');
  const [price, setPrice] = useState<string>('');
  const [quantity, setQuantity] = useState<string>('10');
  const [amountMode, setAmountMode] = useState<'shares' | 'dollars'>('shares');
  const [dollarAmount, setDollarAmount] = useState<string>('1000');
  
  // MT5/TradingView style Take Profit and Stop Loss (optional)
  const [tpEnabled, setTpEnabled] = useState(false);
  const [tpPrice, setTpPrice] = useState<string>('');
  const [slEnabled, setSlEnabled] = useState(false);
  const [slPrice, setSlPrice] = useState<string>('');

  // Settings modal
  const [showSettings, setShowSettings] = useState(false);
  const [settingBalance, setSettingBalance] = useState<string>('100000');
  const [settingLeverage, setSettingLeverage] = useState<number>(5);
  const [savingSettings, setSavingSettings] = useState(false);

  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ text: string; isError: boolean } | null>(null);

  useEffect(() => {
    if (portfolio) {
      setSettingBalance(portfolio.cash.toFixed(0));
      setSettingLeverage(portfolio.max_leverage || 5);
    }
  }, [portfolio?.cash, portfolio?.max_leverage]);

  useEffect(() => {
    if (initialPrice && initialPrice > 0) {
      setPrice(initialPrice.toFixed(2));
    } else if (stock) {
      const defaultP = side === 'Buy' ? stock.ask : stock.bid;
      const effectiveP = defaultP > 0 ? defaultP : (stock.mid || 100);
      setPrice(effectiveP.toFixed(2));
    }
  }, [stock, initialPrice, side]);

  const activePrice = parseFloat(price) || (stock?.mid ?? 100);

  // Calculate actual shares based on amountMode
  const computedShares = amountMode === 'shares'
    ? (parseFloat(quantity) || 0)
    : (activePrice > 0 ? (parseFloat(dollarAmount) || 0) / activePrice : 0);

  const roundedShares = Math.max(1, Math.round(computedShares));
  const totalValue = roundedShares * activePrice;
  const leverage = portfolio?.max_leverage || 5;
  const marginRequired = totalValue / leverage;

  const handlePercentPick = (pct: number) => {
    if (!portfolio || activePrice <= 0) return;
    const maxExpendable = portfolio.buying_power;
    const targetDollars = maxExpendable * pct;
    const targetShares = Math.max(1, Math.floor(targetDollars / activePrice));
    
    if (amountMode === 'shares') {
      setQuantity(String(targetShares));
    } else {
      setDollarAmount((targetShares * activePrice).toFixed(0));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!stock) return;
    if (roundedShares <= 0) {
      setMessage({ text: 'Please specify a valid trade size.', isError: true });
      return;
    }
    if (orderType === 'Limit' && activePrice <= 0) {
      setMessage({ text: 'Please enter a valid limit price.', isError: true });
      return;
    }

    setLoading(true);
    setMessage(null);
    try {
      await onSubmitOrder({
        symbol: stock.symbol,
        side,
        order_type: orderType,
        price: orderType === 'Market' ? 0.0 : activePrice,
        quantity: roundedShares,
      });
      setMessage({
        text: `${side} order executed: ${roundedShares.toLocaleString()} shares of ${stock.symbol} @ $${activePrice.toFixed(2)}`,
        isError: false,
      });
    } catch (err: any) {
      setMessage({ text: err.message || 'Execution failed.', isError: true });
    } finally {
      setLoading(false);
    }
  };

  const handleSaveSettings = async () => {
    if (!onUpdateSettings) return;
    setSavingSettings(true);
    try {
      await onUpdateSettings({
        balance: parseFloat(settingBalance) || 100000,
        leverage: settingLeverage,
      });
      setShowSettings(false);
      setMessage({ text: `Settings saved: $${parseFloat(settingBalance).toLocaleString()} capital @ ${settingLeverage}x leverage`, isError: false });
      setTimeout(() => setMessage(null), 3000);
    } catch (err: any) {
      setMessage({ text: err.message || 'Failed to update settings', isError: true });
    } finally {
      setSavingSettings(false);
    }
  };

  return (
    <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5 flex flex-col justify-between h-full relative shadow-none">
      <div>
        {/* Header & Settings Trigger */}
        <div className="pb-3 border-b border-[#efeee8] dark:border-[#262520] flex items-center justify-between">
          <div className="flex items-center gap-2">
            <h2 className="text-[15px] font-semibold text-[#26251e] dark:text-[#edece6] tracking-tight">Order Ticket</h2>
            <span className="text-[11px] font-mono px-2 py-0.5 rounded-[4px] bg-[#efeee8] dark:bg-[#262520] text-[#5a5852] dark:text-[#a09c92]">
              {leverage}x Leverage
            </span>
          </div>
          <button
            type="button"
            onClick={() => setShowSettings(true)}
            className="flex items-center gap-1.5 text-[12px] font-medium text-[#5a5852] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#edece6] p-1.5 rounded-[6px] hover:bg-[#efeee8] dark:hover:bg-[#262520] transition-colors"
            title="Trading & Account Settings"
          >
            <Settings size={15} />
            <span className="hidden sm:inline">Settings</span>
          </button>
        </div>

        {/* Buy / Sell MT5 Style Selector */}
        <div className="grid grid-cols-2 gap-2 mt-4 bg-[#efeee8] dark:bg-[#262520] p-1 rounded-[8px]">
          <button
            type="button"
            onClick={() => setSide('Buy')}
            className={`py-2 text-[13px] font-semibold rounded-[6px] transition-colors flex items-center justify-center gap-1.5 ${
              side === 'Buy'
                ? 'bg-[#1f8a65] text-white shadow-none'
                : 'text-[#5a5852] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#edece6]'
            }`}
          >
            Buy / Long
          </button>
          <button
            type="button"
            onClick={() => setSide('Sell')}
            className={`py-2 text-[13px] font-semibold rounded-[6px] transition-colors flex items-center justify-center gap-1.5 ${
              side === 'Sell'
                ? 'bg-[#cf2d56] text-white shadow-none'
                : 'text-[#5a5852] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#edece6]'
            }`}
          >
            Sell / Short
          </button>
        </div>

        {/* Order Type Toggle: Market vs Limit */}
        <div className="flex items-center gap-2 mt-3">
          <button
            type="button"
            onClick={() => setOrderType('Market')}
            className={`flex-1 py-1.5 text-[12px] font-mono rounded-[6px] border transition-colors ${
              orderType === 'Market'
                ? 'border-[#26251e] dark:border-[#edece6] bg-[#26251e] dark:bg-[#edece6] text-white dark:text-[#121210]'
                : 'border-[#e6e5e0] dark:border-[#2c2b26] bg-white dark:bg-[#1c1b18] text-[#5a5852] dark:text-[#a09c92] hover:border-[#807d72]'
            }`}
          >
            Market Order
          </button>
          <button
            type="button"
            onClick={() => setOrderType('Limit')}
            className={`flex-1 py-1.5 text-[12px] font-mono rounded-[6px] border transition-colors ${
              orderType === 'Limit'
                ? 'border-[#26251e] dark:border-[#edece6] bg-[#26251e] dark:bg-[#edece6] text-white dark:text-[#121210]'
                : 'border-[#e6e5e0] dark:border-[#2c2b26] bg-white dark:bg-[#1c1b18] text-[#5a5852] dark:text-[#a09c92] hover:border-[#807d72]'
            }`}
          >
            Limit Order
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-3 mt-3">
          {/* Limit Price Input if Limit Order */}
          {orderType === 'Limit' && (
            <div>
              <label className="block text-[11px] font-mono uppercase text-[#807d72] dark:text-[#a09c92] mb-1">
                Limit Price ($)
              </label>
              <input
                type="number"
                step="0.01"
                value={price}
                onChange={(e) => setPrice(e.target.value)}
                className="w-full bg-white dark:bg-[#26251e] border border-[#e6e5e0] dark:border-[#2c2b26] focus:border-[#26251e] dark:focus:border-[#edece6] rounded-[8px] px-3 h-9 font-mono text-[14px] text-[#26251e] dark:text-[#edece6] outline-none"
                placeholder="0.00"
              />
            </div>
          )}

          {/* Amount Mode Toggle: Shares vs Dollar Value */}
          <div>
            <div className="flex items-center justify-between mb-1">
              <label className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#a09c92]">
                Trade Amount
              </label>
              <div className="flex items-center gap-1 bg-[#efeee8] dark:bg-[#262520] p-0.5 rounded-[4px] text-[11px] font-mono">
                <button
                  type="button"
                  onClick={() => setAmountMode('shares')}
                  className={`px-2 py-0.5 rounded-[3px] transition-colors ${
                    amountMode === 'shares' ? 'bg-white dark:bg-[#1c1b18] font-medium text-[#26251e] dark:text-[#edece6]' : 'text-[#807d72] dark:text-[#a09c92]'
                  }`}
                >
                  Shares
                </button>
                <button
                  type="button"
                  onClick={() => setAmountMode('dollars')}
                  className={`px-2 py-0.5 rounded-[3px] transition-colors ${
                    amountMode === 'dollars' ? 'bg-white dark:bg-[#1c1b18] font-medium text-[#26251e] dark:text-[#edece6]' : 'text-[#807d72] dark:text-[#a09c92]'
                  }`}
                >
                  USD ($)
                </button>
              </div>
            </div>

            {amountMode === 'shares' ? (
              <input
                type="number"
                step="1"
                min="1"
                value={quantity}
                onChange={(e) => setQuantity(e.target.value)}
                className="w-full bg-white dark:bg-[#26251e] border border-[#e6e5e0] dark:border-[#2c2b26] focus:border-[#26251e] dark:focus:border-[#edece6] rounded-[8px] px-3 h-9 font-mono text-[14px] text-[#26251e] dark:text-[#edece6] outline-none"
                placeholder="10"
              />
            ) : (
              <input
                type="number"
                step="100"
                min="10"
                value={dollarAmount}
                onChange={(e) => setDollarAmount(e.target.value)}
                className="w-full bg-white dark:bg-[#26251e] border border-[#e6e5e0] dark:border-[#2c2b26] focus:border-[#26251e] dark:focus:border-[#edece6] rounded-[8px] px-3 h-9 font-mono text-[14px] text-[#26251e] dark:text-[#edece6] outline-none"
                placeholder="1000"
              />
            )}
          </div>

          {/* Quick Percentage Buttons */}
          <div className="grid grid-cols-4 gap-1.5">
            {[0.25, 0.50, 0.75, 1.0].map((pct) => (
              <button
                key={pct}
                type="button"
                onClick={() => handlePercentPick(pct)}
                className="py-1 text-[11px] font-mono rounded-[6px] bg-[#fafaf7] dark:bg-[#262520] hover:bg-[#efeee8] dark:hover:bg-[#2f2e26] border border-[#e6e5e0] dark:border-[#2c2b26] text-[#5a5852] dark:text-[#a09c92] font-medium transition-colors"
              >
                {pct === 1.0 ? 'MAX' : `${pct * 100}%`}
              </button>
            ))}
          </div>

          {/* MT5 / TradingView Style TP / SL Toggles */}
          <div className="pt-2 border-t border-[#efeee8] dark:border-[#262520] space-y-2">
            <div className="flex items-center justify-between">
              <label className="flex items-center gap-2 cursor-pointer text-[12px] font-medium text-[#5a5852] dark:text-[#a09c92]">
                <input
                  type="checkbox"
                  checked={tpEnabled}
                  onChange={(e) => setTpEnabled(e.target.checked)}
                  className="rounded border-[#e6e5e0] dark:border-[#2c2b26] text-[#1f8a65] focus:ring-0"
                />
                <span>Take Profit (TP)</span>
              </label>
              {tpEnabled && (
                <input
                  type="number"
                  step="0.01"
                  value={tpPrice}
                  onChange={(e) => setTpPrice(e.target.value)}
                  placeholder={`Target $`}
                  className="w-28 bg-white dark:bg-[#26251e] border border-[#e6e5e0] dark:border-[#2c2b26] focus:border-[#26251e] dark:focus:border-[#edece6] rounded-[6px] px-2 h-7 font-mono text-[12px] text-[#26251e] dark:text-[#edece6] outline-none text-right"
                />
              )}
            </div>

            <div className="flex items-center justify-between">
              <label className="flex items-center gap-2 cursor-pointer text-[12px] font-medium text-[#5a5852] dark:text-[#a09c92]">
                <input
                  type="checkbox"
                  checked={slEnabled}
                  onChange={(e) => setSlEnabled(e.target.checked)}
                  className="rounded border-[#e6e5e0] dark:border-[#2c2b26] text-[#cf2d56] focus:ring-0"
                />
                <span>Stop Loss (SL)</span>
              </label>
              {slEnabled && (
                <input
                  type="number"
                  step="0.01"
                  value={slPrice}
                  onChange={(e) => setSlPrice(e.target.value)}
                  placeholder={`Exit $`}
                  className="w-28 bg-white dark:bg-[#26251e] border border-[#e6e5e0] dark:border-[#2c2b26] focus:border-[#26251e] dark:focus:border-[#edece6] rounded-[6px] px-2 h-7 font-mono text-[12px] text-[#26251e] dark:text-[#edece6] outline-none text-right"
                />
              )}
            </div>
          </div>

          {/* Execution Summary Breakdown */}
          <div className="bg-[#fafaf7] dark:bg-[#262520] p-2.5 rounded-[8px] border border-[#efeee8] dark:border-[#2c2b26] text-[11px] font-mono space-y-1">
            <div className="flex justify-between text-[#807d72] dark:text-[#a09c92]">
              <span>Shares / Units:</span>
              <span className="text-[#26251e] dark:text-[#edece6] font-medium">{roundedShares.toLocaleString()}</span>
            </div>
            <div className="flex justify-between text-[#807d72] dark:text-[#a09c92]">
              <span>Notional Value:</span>
              <span className="text-[#26251e] dark:text-[#edece6]">${totalValue.toFixed(2)}</span>
            </div>
            <div className="flex justify-between text-[#807d72] dark:text-[#a09c92]">
              <span>Margin Required ({leverage}x):</span>
              <span className="text-[#26251e] dark:text-[#edece6]">${marginRequired.toFixed(2)}</span>
            </div>
            <div className="flex justify-between text-[#807d72] dark:text-[#a09c92]">
              <span>Available Cash:</span>
              <span className="text-[#26251e] dark:text-[#edece6]">${portfolio?.cash.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) || '0.00'}</span>
            </div>
          </div>

          {/* Feedback Message */}
          {message && (
            <div
              className={`p-2 rounded-[6px] text-[11px] font-mono ${
                message.isError ? 'bg-[#cf2d56]/10 text-[#cf2d56]' : 'bg-[#1f8a65]/10 text-[#1f8a65]'
              }`}
            >
              {message.text}
            </div>
          )}

          {/* Primary CTA Button */}
          <button
            type="submit"
            disabled={loading}
            className={`w-full font-medium text-[14px] h-10 rounded-[8px] transition-colors disabled:opacity-50 text-white ${
              side === 'Buy'
                ? 'bg-[#1f8a65] hover:bg-[#186f51]'
                : 'bg-[#cf2d56] hover:bg-[#aa2345]'
            }`}
          >
            {loading ? 'Routing Order...' : `${side} ${roundedShares.toLocaleString()} ${stock?.symbol || ''}`}
          </button>
        </form>
      </div>

      {/* Configuration Settings Dialog Modal */}
      {showSettings && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-6 max-w-[420px] w-full space-y-4 shadow-none">
            <div className="flex items-center justify-between pb-3 border-b border-[#efeee8] dark:border-[#262520]">
              <div className="flex items-center gap-2">
                <Sliders size={18} className="text-[#f54e00]" />
                <h3 className="text-[15px] font-semibold text-[#26251e] dark:text-[#edece6]">Trading Settings</h3>
              </div>
              <button
                type="button"
                onClick={() => setShowSettings(false)}
                className="text-[#807d72] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#edece6] p-1 rounded-[6px]"
              >
                <X size={18} />
              </button>
            </div>

            <div className="space-y-4">
              {/* Cash Balance Configuration */}
              <div>
                <label className="block text-[11px] font-mono uppercase text-[#807d72] dark:text-[#a09c92] mb-1">
                  Account Balance / Deposit ($)
                </label>
                <input
                  type="number"
                  step="1000"
                  min="1000"
                  value={settingBalance}
                  onChange={(e) => setSettingBalance(e.target.value)}
                  className="w-full bg-white dark:bg-[#26251e] border border-[#e6e5e0] dark:border-[#2c2b26] focus:border-[#26251e] dark:focus:border-[#edece6] rounded-[8px] px-3 h-10 font-mono text-[14px] text-[#26251e] dark:text-[#edece6] outline-none"
                />
                <div className="flex items-center gap-2 mt-1.5">
                  {[10000, 50000, 100000, 500000].map((b) => (
                    <button
                      key={b}
                      type="button"
                      onClick={() => setSettingBalance(String(b))}
                      className="px-2 py-1 text-[11px] font-mono rounded-[4px] bg-[#fafaf7] dark:bg-[#262520] hover:bg-[#efeee8] dark:hover:bg-[#2f2e26] border border-[#e6e5e0] dark:border-[#2c2b26] text-[#5a5852] dark:text-[#a09c92]"
                    >
                      ${b >= 1000 ? `${b / 1000}k` : b}
                    </button>
                  ))}
                </div>
              </div>

              {/* Leverage Selection */}
              <div>
                <label className="block text-[11px] font-mono uppercase text-[#807d72] dark:text-[#a09c92] mb-1">
                  Account Leverage (1x – 100x)
                </label>
                <div className="grid grid-cols-4 gap-2">
                  {[1, 2, 5, 10, 20, 50, 100].map((lev) => (
                    <button
                      key={lev}
                      type="button"
                      onClick={() => setSettingLeverage(lev)}
                      className={`py-2 text-[12px] font-mono rounded-[6px] border transition-colors ${
                        settingLeverage === lev
                          ? 'border-[#26251e] dark:border-[#edece6] bg-[#26251e] dark:bg-[#edece6] text-white dark:text-[#121210] font-semibold'
                          : 'border-[#e6e5e0] dark:border-[#2c2b26] bg-white dark:bg-[#1c1b18] text-[#5a5852] dark:text-[#a09c92] hover:border-[#807d72]'
                      }`}
                    >
                      {lev}x
                    </button>
                  ))}
                </div>
                <p className="text-[11px] text-[#807d72] dark:text-[#78756c] mt-1.5">
                  Higher leverage amplifies buying power while adjusting margin requirements accordingly.
                </p>
              </div>
            </div>

            <div className="flex gap-2 pt-2">
              <button
                type="button"
                onClick={() => setShowSettings(false)}
                className="flex-1 py-2 rounded-[8px] border border-[#e6e5e0] dark:border-[#2c2b26] text-[#5a5852] dark:text-[#a09c92] font-medium text-[13px] hover:bg-[#fafaf7] dark:hover:bg-[#262520]"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={handleSaveSettings}
                disabled={savingSettings}
                className="flex-1 py-2 rounded-[8px] bg-[#f54e00] hover:bg-[#d04200] text-white font-medium text-[13px] disabled:opacity-50"
              >
                {savingSettings ? 'Applying...' : 'Apply Settings'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
