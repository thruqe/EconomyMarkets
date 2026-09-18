import React, { useState, useMemo, useRef, useEffect } from 'react';
import {
  Search,
  ChevronDown,
  Layers,
  Globe,
  LineChart,
  DollarSign,
  TrendingUp,
  TrendingDown,
  X,
} from 'lucide-react';
import {
  CandleRecord,
  OrderBookDTO,
  PortfolioDTO,
  PriceItem,
  CountryIndexDTO,
  WorldStockDTO,
  ForexPairDTO,
} from '../types/api';
import { CandleChart } from './CandleChart';
import { OrderBookDOM } from './OrderBookDOM';
import { TradeTicket } from './TradeTicket';
import { PortfolioView } from './PortfolioView';

interface TerminalViewProps {
  selectedStock: PriceItem | null;
  candles: CandleRecord[];
  orderBook: OrderBookDTO | null;
  portfolio: PortfolioDTO | null;
  timeframe: string;
  onTimeframeChange: (tf: string) => void;
  onSubmitOrder: (order: any) => Promise<void>;
  onClosePosition: (symbol: string) => Promise<void>;
  onSelectStock: (symbol: string) => void;
  onUpdateSettings?: (data: { balance?: number; leverage?: number }) => Promise<void>;
  allPrices?: PriceItem[];
  indices?: CountryIndexDTO[];
  worldStocks?: WorldStockDTO[];
  forexPairs?: ForexPairDTO[];
}

type FilterCategory = 'all' | 'domestic' | 'indices' | 'world' | 'forex';

export const TerminalView: React.FC<TerminalViewProps> = ({
  selectedStock,
  candles,
  orderBook,
  portfolio,
  timeframe,
  onTimeframeChange,
  onSubmitOrder,
  onClosePosition,
  onSelectStock,
  onUpdateSettings,
  allPrices = [],
  indices = [],
  worldStocks = [],
  forexPairs = [],
}) => {
  const [ticketPrice, setTicketPrice] = useState<number | undefined>(undefined);
  const [selectorOpen, setSelectorOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [activeCategory, setActiveCategory] = useState<FilterCategory>('all');
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown on click outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setSelectorOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Consolidate all searchable instruments
  const instruments = useMemo(() => {
    const list: Array<{
      symbol: string;
      name: string;
      category: FilterCategory;
      categoryLabel: string;
      price: number;
      change?: number;
      badge?: string;
    }> = [];

    // 1. Domestic public & private equities
    for (const p of allPrices) {
      list.push({
        symbol: p.symbol,
        name: p.name,
        category: 'domestic',
        categoryLabel: 'Domestic Equity',
        price: p.mid > 0 ? p.mid : p.reported_value,
        badge: p.is_private ? p.stage : p.cap_tier,
      });
    }

    // 2. Sovereign Benchmark Indices
    for (const idx of indices) {
      list.push({
        symbol: idx.symbol,
        name: idx.name,
        category: 'indices',
        categoryLabel: 'Sovereign Index',
        price: idx.price,
        change: idx.change_24h,
        badge: idx.country_id,
      });
    }

    // 3. International World Stocks
    for (const ws of worldStocks) {
      list.push({
        symbol: ws.ticker,
        name: ws.name,
        category: 'world',
        categoryLabel: 'World Stock',
        price: ws.price,
        change: ws.change_pct,
        badge: ws.country_id,
      });
    }

    // 4. Forex Pairs
    for (const fx of forexPairs) {
      list.push({
        symbol: fx.symbol,
        name: `${fx.base_currency}/${fx.quote_currency} Exchange`,
        category: 'forex',
        categoryLabel: 'Currency FX',
        price: fx.rate,
        change: fx.change_24h,
        badge: 'FOREX',
      });
    }

    return list;
  }, [allPrices, indices, worldStocks, forexPairs]);

  // Filter instruments based on query and category
  const filteredInstruments = useMemo(() => {
    const q = searchQuery.trim().toLowerCase();
    return instruments.filter((item) => {
      const matchCat = activeCategory === 'all' || item.category === activeCategory;
      if (!matchCat) return false;
      if (!q) return true;
      return (
        item.symbol.toLowerCase().includes(q) ||
        item.name.toLowerCase().includes(q) ||
        item.categoryLabel.toLowerCase().includes(q)
      );
    });
  }, [instruments, searchQuery, activeCategory]);

  return (
    <div className="space-y-6">
      {/* Top Header: Instrument Selector & Market HUD */}
      <div
        ref={dropdownRef}
        className="relative bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-3 sm:p-4 flex flex-col md:flex-row items-start md:items-center justify-between gap-4 select-none shadow-none"
      >
        <div className="flex flex-wrap items-center gap-3 sm:gap-4 w-full md:w-auto">
          {/* Active Instrument Selector Pill */}
          <button
            onClick={() => setSelectorOpen(!selectorOpen)}
            className="flex items-center gap-2.5 px-3 py-2 rounded-[8px] bg-[#fafaf7] dark:bg-[#26251e] hover:bg-[#efeee8] dark:hover:bg-[#2c2b24] border border-[#e6e5e0] dark:border-[#2c2b26] transition-colors group"
            title="Click to switch trading instrument"
          >
            <div className="flex items-baseline gap-2">
              <span className="font-mono text-[16px] font-semibold text-[#26251e] dark:text-[#edece6] group-hover:text-[#f54e00] transition-colors">
                {selectedStock?.symbol || 'SELECT'}
              </span>
              <span className="text-[13px] text-[#5a5852] dark:text-[#a09c92] max-w-[150px] sm:max-w-[220px] truncate">
                {selectedStock?.name || 'Instrument'}
              </span>
            </div>
            <span className="text-[10px] font-mono uppercase px-1.5 py-0.5 rounded-[4px] bg-[#efeee8] dark:bg-[#262520] text-[#807d72] dark:text-[#a09c92]">
              {selectedStock?.cap_tier || selectedStock?.stage || 'Asset'}
            </span>
            <ChevronDown size={14} className="text-[#807d72] dark:text-[#a09c92] ml-1" />
          </button>

          {/* Quick Price & Telemetry */}
          <div className="flex items-baseline gap-3 text-[13px] font-mono">
            <span className="text-[18px] font-semibold text-[#26251e] dark:text-[#edece6]">
              ${selectedStock?.mid ? selectedStock.mid.toFixed(selectedStock.cap_tier === 'Forex' ? 4 : 2) : (selectedStock?.reported_value || 0).toFixed(2)}
            </span>
            <span className="text-[#807d72] dark:text-[#78756c] text-[11px]">
              Spr: ${selectedStock?.spread ? selectedStock.spread.toFixed(selectedStock.cap_tier === 'Forex' ? 4 : 2) : '0.05'}
            </span>
            {selectedStock?.pe_ratio !== undefined && selectedStock.pe_ratio > 0 && (
              <span className="text-[#807d72] dark:text-[#78756c] text-[11px] hidden sm:inline">
                P/E: {selectedStock.pe_ratio.toFixed(1)}
              </span>
            )}
          </div>
        </div>

        {/* Quick search input shortcut */}
        <div className="w-full md:w-auto flex items-center gap-2">
          <div className="relative flex-1 md:w-64">
            <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-[#807d72] dark:text-[#a09c92]" />
            <input
              type="text"
              placeholder="Search ticker, index, FX..."
              value={searchQuery}
              onFocus={() => setSelectorOpen(true)}
              onChange={(e) => {
                setSearchQuery(e.target.value);
                setSelectorOpen(true);
              }}
              className="w-full pl-8 pr-8 py-1.5 bg-[#fafaf7] dark:bg-[#26251e] text-[#26251e] dark:text-[#edece6] text-[12px] font-mono rounded-[8px] border border-[#e6e5e0] dark:border-[#2c2b26] focus:outline-none focus:border-[#f54e00] transition-colors placeholder:text-[#a09c92]"
            />
            {searchQuery && (
              <button
                onClick={() => setSearchQuery('')}
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-[#807d72] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#edece6]"
              >
                <X size={12} />
              </button>
            )}
          </div>
        </div>

        {/* Search & Instrument Switcher Dropdown Modal */}
        {selectorOpen && (
          <div className="absolute top-full left-0 right-0 mt-2 bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-3 z-50 max-h-[460px] flex flex-col shadow-none">
            {/* Category Filter Pills */}
            <div className="flex flex-wrap items-center gap-1.5 pb-2.5 border-b border-[#efeee8] dark:border-[#262520]">
              {[
                { id: 'all', label: 'All Assets', icon: Layers },
                { id: 'domestic', label: 'Domestic Stocks', icon: LineChart },
                { id: 'indices', label: 'Benchmark Indices', icon: TrendingUp },
                { id: 'world', label: 'World Equities', icon: Globe },
                { id: 'forex', label: 'Forex Pairs', icon: DollarSign },
              ].map((cat) => {
                const Icon = cat.icon;
                const isActive = activeCategory === cat.id;
                return (
                  <button
                    key={cat.id}
                    onClick={() => setActiveCategory(cat.id as FilterCategory)}
                    className={`flex items-center gap-1 px-2.5 py-1 rounded-[6px] text-[11px] font-mono transition-colors ${
                      isActive
                        ? 'bg-[#26251e] dark:bg-[#edece6] text-white dark:text-[#121210]'
                        : 'bg-[#fafaf7] dark:bg-[#26251e] text-[#5a5852] dark:text-[#a09c92] hover:bg-[#efeee8] dark:hover:bg-[#262520] hover:text-[#26251e] dark:hover:text-[#edece6] border border-[#e6e5e0] dark:border-[#2c2b26]'
                    }`}
                  >
                    <Icon size={11} />
                    <span>{cat.label}</span>
                  </button>
                );
              })}
            </div>

            {/* Instruments Scroll List */}
            <div className="overflow-y-auto flex-1 mt-2 divide-y divide-[#efeee8] dark:divide-[#262520] max-h-[320px]">
              {filteredInstruments.map((item) => {
                const isSelected = selectedStock?.symbol === item.symbol;
                return (
                  <div
                    key={`${item.category}-${item.symbol}`}
                    onClick={() => {
                      onSelectStock(item.symbol);
                      setSelectorOpen(false);
                      setSearchQuery('');
                    }}
                    className={`p-2.5 flex items-center justify-between gap-3 hover:bg-[#fafaf7] dark:hover:bg-[#262520] cursor-pointer rounded-[6px] transition-colors ${
                      isSelected ? 'bg-[#fafaf7] dark:bg-[#262520] font-semibold' : ''
                    }`}
                  >
                    <div className="flex items-center gap-2.5 min-w-0">
                      <span className="font-mono text-[13px] font-semibold text-[#26251e] dark:text-[#edece6] w-16">
                        {item.symbol}
                      </span>
                      <span className="text-[12px] text-[#5a5852] dark:text-[#a09c92] truncate max-w-[200px] sm:max-w-[280px]">
                        {item.name}
                      </span>
                      <span className="text-[10px] font-mono px-1.5 py-0.5 rounded-[4px] bg-[#efeee8] dark:bg-[#2c2b26] text-[#807d72] dark:text-[#a09c92] hidden sm:inline">
                        {item.badge || item.categoryLabel}
                      </span>
                    </div>

                    <div className="flex items-center gap-3 text-right font-mono text-[12px] flex-shrink-0">
                      <span className="text-[#26251e] dark:text-[#edece6] font-medium">
                        ${item.price.toFixed(item.category === 'forex' ? 4 : 2)}
                      </span>
                      {item.change !== undefined && (
                        <span
                          className={`text-[11px] font-medium ${
                            item.change >= 0 ? 'text-[#1f8a65]' : 'text-[#cf2d56]'
                          }`}
                        >
                          {item.change >= 0 ? `+${item.change.toFixed(2)}%` : `${item.change.toFixed(2)}%`}
                        </span>
                      )}
                    </div>
                  </div>
                );
              })}

              {filteredInstruments.length === 0 && (
                <div className="py-8 text-center text-[#807d72] dark:text-[#a09c92] text-[12px] font-mono">
                  No instruments found matching "{searchQuery}".
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Upper Grid: Chart, DOM, and Ticket */}
      <div className="grid grid-cols-1 md:grid-cols-12 gap-6">
        {/* Left Column: Interactive Candlestick Chart */}
        <div className="md:col-span-12 xl:col-span-6">
          <CandleChart
            candles={candles}
            stock={selectedStock}
            timeframe={timeframe}
            onTimeframeChange={onTimeframeChange}
          />
        </div>

        {/* Center Column: Order Book Depth Ladder */}
        <div className="md:col-span-6 xl:col-span-3">
          <OrderBookDOM
            book={orderBook}
            onSelectPrice={(p) => setTicketPrice(p)}
          />
        </div>

        {/* Right Column: Execution Ticket */}
        <div className="md:col-span-6 xl:col-span-3">
          <TradeTicket
            stock={selectedStock}
            portfolio={portfolio}
            initialPrice={ticketPrice}
            onSubmitOrder={onSubmitOrder}
            onUpdateSettings={onUpdateSettings}
          />
        </div>
      </div>

      {/* Lower Row: Full-width Mark-to-Market Portfolio */}
      <div>
        <PortfolioView
          portfolio={portfolio}
          onClosePosition={onClosePosition}
          onSelectStock={onSelectStock}
        />
      </div>
    </div>
  );
};
