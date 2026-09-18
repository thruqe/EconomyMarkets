import React, { useState, useMemo, useRef, useEffect } from 'react';
import { Search, ArrowUpDown, ChevronRight, ChevronDown, Check, Globe, Landmark, DollarSign, TrendingUp, RefreshCw, PlusCircle, MinusCircle, BarChart3, LineChart, X, ShieldCheck } from 'lucide-react';
import { PriceItem, SovereignBondDTO, ForexPairDTO, WorldStockDTO, CountryIndexDTO } from '../types/api';

interface CustomDropdownProps {
  value: string;
  onChange: (val: string) => void;
  options: { value: string; label: string }[];
  placeholder?: string;
  className?: string;
}

const CustomDropdown: React.FC<CustomDropdownProps> = ({
  value,
  onChange,
  options,
  placeholder,
  className = '',
}) => {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (ref.current && !ref.current.contains(event.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const activeOption = options.find((o) => o.value === value);
  const displayLabel = activeOption ? activeOption.label : (placeholder || value);

  return (
    <div className={`relative ${className}`} ref={ref}>
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="w-full flex items-center justify-between gap-2 px-3 h-9 rounded-[8px] text-[13px] font-medium bg-white dark:bg-[#1c1b18] text-[#26251e] dark:text-[#edece6] border border-[#e6e5e0] dark:border-[#2c2b26] hover:border-[#26251e] dark:hover:border-[#edece6] transition-colors select-none"
      >
        <span className="truncate">{displayLabel}</span>
        <ChevronDown size={13} className={`text-[#807d72] dark:text-[#a09c92] shrink-0 transition-transform ${open ? 'rotate-180' : ''}`} />
      </button>

      {open && (
        <div className="absolute top-full left-0 mt-1 min-w-full sm:min-w-[200px] max-w-[280px] bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[8px] py-1 z-50 max-h-60 overflow-y-auto shadow-none">
          {options.map((opt) => {
            const isSelected = opt.value === value;
            return (
              <button
                key={opt.value}
                type="button"
                onClick={() => {
                  onChange(opt.value);
                  setOpen(false);
                }}
                className={`w-full flex items-center justify-between px-3 py-2 text-[13px] text-left transition-colors ${
                  isSelected
                    ? 'bg-[#fafaf7] dark:bg-[#262520] text-[#f54e00] font-semibold'
                    : 'text-[#5a5852] dark:text-[#a09c92] hover:bg-[#efeee8] dark:hover:bg-[#262520] hover:text-[#26251e] dark:hover:text-[#edece6]'
                }`}
              >
                <span className="truncate">{opt.label}</span>
                {isSelected && <Check size={13} className="text-[#f54e00] shrink-0 ml-2" />}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
};

export type UniverseAssetClass = 'domestic' | 'world' | 'indices' | 'bonds' | 'forex';

interface StockUniverseProps {
  prices: PriceItem[];
  bonds?: SovereignBondDTO[];
  forexPairs?: ForexPairDTO[];
  worldStocks?: WorldStockDTO[];
  countryIndices?: CountryIndexDTO[];
  onSelectStock: (symbol: string) => void;
  onIssueBond?: (maturityYears: number, amount: number) => Promise<void>;
  onBuybackBond?: (maturityYears: number, amount: number) => Promise<void>;
  onSwapForex?: (pair: string, amount: number, side: 'buy' | 'sell') => Promise<void>;
}

export const StockUniverse: React.FC<StockUniverseProps> = ({
  prices,
  bonds = [],
  forexPairs = [],
  worldStocks = [],
  countryIndices = [],
  onSelectStock,
  onIssueBond,
  onBuybackBond,
  onSwapForex,
}) => {
  const [assetClass, setAssetClassState] = useState<UniverseAssetClass>(() => {
    return (localStorage.getItem('em_universe_asset_class') as UniverseAssetClass) || 'domestic';
  });
  const setAssetClass = (val: UniverseAssetClass) => {
    setAssetClassState(val);
    try { localStorage.setItem('em_universe_asset_class', val); } catch (_) {}
  };
  const [search, setSearch] = useState('');
  const [selectedSector, setSelectedSectorState] = useState<string>(() => {
    return localStorage.getItem('em_universe_sector') || 'ALL';
  });
  const setSelectedSector = (val: string) => {
    setSelectedSectorState(val);
    try { localStorage.setItem('em_universe_sector', val); } catch (_) {}
  };
  const [selectedTier, setSelectedTierState] = useState<string>(() => {
    return localStorage.getItem('em_universe_tier') || 'ALL';
  });
  const setSelectedTier = (val: string) => {
    setSelectedTierState(val);
    try { localStorage.setItem('em_universe_tier', val); } catch (_) {}
  };
  const [selectedCountry, setSelectedCountryState] = useState<string>(() => {
    return localStorage.getItem('em_universe_country') || 'ALL';
  });
  const setSelectedCountry = (val: string) => {
    setSelectedCountryState(val);
    try { localStorage.setItem('em_universe_country', val); } catch (_) {}
  };
  const [sortKey, setSortKey] = useState<keyof PriceItem>('market_cap');
  const [sortAsc, setSortAsc] = useState(false);

  // Bond Action Modal state
  const [bondModal, setBondModal] = useState<{ open: boolean; type: 'issue' | 'buyback'; bond: SovereignBondDTO | null }>({
    open: false,
    type: 'issue',
    bond: null,
  });
  const [bondAmount, setBondAmount] = useState('50000000');
  const [bondBusy, setBondBusy] = useState(false);
  const [bondFeedback, setBondFeedback] = useState<string | null>(null);

  // Forex Swap Modal state
  const [forexModal, setForexModal] = useState<{ open: boolean; pair: ForexPairDTO | null }>({
    open: false,
    pair: null,
  });
  const [swapAmount, setSwapAmount] = useState('10000');
  const [swapSide, setSwapSide] = useState<'buy' | 'sell'>('buy');
  const [swapBusy, setSwapBusy] = useState(false);
  const [swapFeedback, setSwapFeedback] = useState<string | null>(null);

  const sectors = useMemo(() => {
    const set = new Set(prices.map((p) => p.sector));
    return ['ALL', ...Array.from(set)];
  }, [prices]);

  const sectorOptions = useMemo(() => {
    return sectors.map((sec) => ({
      value: sec,
      label: sec === 'ALL' ? 'All Sectors' : sec,
    }));
  }, [sectors]);

  const worldCountries = useMemo(() => {
    const set = new Set(worldStocks.map((w) => w.country_name));
    return ['ALL', ...Array.from(set)];
  }, [worldStocks]);

  const countryOptions = useMemo(() => {
    return worldCountries.map((c) => ({
      value: c,
      label: c === 'ALL' ? 'All Foreign Nations' : c,
    }));
  }, [worldCountries]);

  const tierOptions = useMemo(() => [
    { value: 'ALL', label: 'All Tiers & Ventures' },
    { value: 'Public', label: 'Public Listed' },
    { value: 'Private', label: 'Private Ventures' },
    { value: 'Mega Cap', label: 'Mega Cap' },
    { value: 'Large Cap', label: 'Large Cap' },
    { value: 'Mid Cap', label: 'Mid Cap' },
    { value: 'Small Cap', label: 'Small Cap' },
  ], []);

  const filteredStocks = useMemo(() => {
    const clean = (s: string) => s.replace(/\s+/g, '').toLowerCase();
    return prices
      .filter((p) => {
        const matchSearch =
          p.symbol.toLowerCase().includes(search.toLowerCase()) ||
          p.name.toLowerCase().includes(search.toLowerCase());
        const matchSector = selectedSector === 'ALL' || p.sector === selectedSector;
        const matchTier =
          selectedTier === 'ALL' ||
          (selectedTier === 'Private' && p.is_private) ||
          (selectedTier === 'Public' && !p.is_private) ||
          (p.cap_tier && clean(p.cap_tier) === clean(selectedTier));
        return matchSearch && matchSector && matchTier;
      })
      .sort((a, b) => {
        const valA = a[sortKey];
        const valB = b[sortKey];
        if (typeof valA === 'number' && typeof valB === 'number') {
          return sortAsc ? valA - valB : valB - valA;
        }
        return sortAsc
          ? String(valA).localeCompare(String(valB))
          : String(valB).localeCompare(String(valA));
      });
  }, [prices, search, selectedSector, selectedTier, sortKey, sortAsc]);

  const filteredWorldStocks = useMemo(() => {
    return worldStocks.filter((w) => {
      const matchSearch =
        w.ticker.toLowerCase().includes(search.toLowerCase()) ||
        w.name.toLowerCase().includes(search.toLowerCase());
      const matchCountry = selectedCountry === 'ALL' || w.country_name === selectedCountry;
      return matchSearch && matchCountry;
    });
  }, [worldStocks, search, selectedCountry]);

  const filteredForex = useMemo(() => {
    return forexPairs.filter((fx) =>
      fx.symbol.toLowerCase().includes(search.toLowerCase()) ||
      fx.quote_currency.toLowerCase().includes(search.toLowerCase())
    );
  }, [forexPairs, search]);

  const filteredIndices = useMemo(() => {
    return countryIndices.filter((idx) =>
      idx.symbol.toLowerCase().includes(search.toLowerCase()) ||
      idx.name.toLowerCase().includes(search.toLowerCase()) ||
      idx.country_id.toLowerCase().includes(search.toLowerCase())
    );
  }, [countryIndices, search]);

  const handleSort = (key: keyof PriceItem) => {
    if (sortKey === key) {
      setSortAsc(!sortAsc);
    } else {
      setSortKey(key);
      setSortAsc(false);
    }
  };

  const formatB = (val: number) => {
    if (val >= 1e12) return `$${(val / 1e12).toFixed(2)}T`;
    if (val >= 1e9) return `$${(val / 1e9).toFixed(2)}B`;
    if (val >= 1e6) return `$${(val / 1e6).toFixed(2)}M`;
    return `$${val.toFixed(2)}`;
  };

  const handleExecuteBond = async () => {
    if (!bondModal.bond) return;
    const amt = parseFloat(bondAmount);
    if (isNaN(amt) || amt <= 0) return;
    setBondBusy(true);
    setBondFeedback(null);
    try {
      if (bondModal.type === 'issue') {
        if (onIssueBond) await onIssueBond(bondModal.bond.maturity_years, amt);
        setBondFeedback(`Successfully issued ${formatB(amt)} in ${bondModal.bond.name}`);
      } else {
        if (onBuybackBond) await onBuybackBond(bondModal.bond.maturity_years, amt);
        setBondFeedback(`Successfully repurchased ${formatB(amt)} in ${bondModal.bond.name}`);
      }
      setTimeout(() => {
        setBondModal({ open: false, type: 'issue', bond: null });
        setBondFeedback(null);
      }, 1400);
    } catch (err: any) {
      setBondFeedback(`Action failed: ${err.message || err}`);
    } finally {
      setBondBusy(false);
    }
  };

  const handleExecuteSwap = async () => {
    if (!forexModal.pair) return;
    const amt = parseFloat(swapAmount);
    if (isNaN(amt) || amt <= 0) return;
    setSwapBusy(true);
    setSwapFeedback(null);
    try {
      if (onSwapForex) await onSwapForex(forexModal.pair.symbol, amt, swapSide);
      setSwapFeedback(`Swap completed: ${swapSide === 'buy' ? 'Bought' : 'Sold'} ${forexModal.pair.quote_currency} @ ${forexModal.pair.rate.toFixed(4)}`);
      setTimeout(() => {
        setForexModal({ open: false, pair: null });
        setSwapFeedback(null);
      }, 1400);
    } catch (err: any) {
      setSwapFeedback(`Swap error: ${err.message || err}`);
    } finally {
      setSwapBusy(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header & Category Tabs (Mobile scrollable container) */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <div className="text-[11px] font-mono uppercase tracking-[0.08em] text-[#807d72] dark:text-[#a09c92]">
            Asset Universe · Listed Markets
          </div>
          <h1 className="text-[26px] font-normal tracking-[-0.025em] text-[#26251e] dark:text-[#f7f7f4] leading-[1.25]">
            Market Universe
          </h1>
        </div>

        {/* Multi-Asset Filter Tabs (Horizontally scrollable on mobile) */}
        <div className="w-full md:w-auto overflow-x-auto scrollbar-none pb-1 md:pb-0">
          <div className="inline-flex bg-[#efeee8] dark:bg-[#1c1b18] p-1 rounded-[10px] border border-[#e6e5e0] dark:border-[#2c2b26] min-w-max">
            <button
              onClick={() => setAssetClass('domestic')}
              className={`px-3 py-1.5 rounded-[7px] text-[12px] font-medium transition-all flex items-center gap-1.5 shrink-0 ${
                assetClass === 'domestic'
                  ? 'bg-white dark:bg-[#26251e] text-[#26251e] dark:text-[#f7f7f4] border border-[#e6e5e0] dark:border-[#383730]'
                  : 'text-[#5a5852] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#f7f7f4]'
              }`}
            >
              <TrendingUp size={13} className={assetClass === 'domestic' ? 'text-[#f54e00]' : ''} />
              <span>Domestic Stocks ({prices.length})</span>
            </button>
            <button
              onClick={() => setAssetClass('world')}
              className={`px-3 py-1.5 rounded-[7px] text-[12px] font-medium transition-all flex items-center gap-1.5 shrink-0 ${
                assetClass === 'world'
                  ? 'bg-white dark:bg-[#26251e] text-[#26251e] dark:text-[#f7f7f4] border border-[#e6e5e0] dark:border-[#383730]'
                  : 'text-[#5a5852] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#f7f7f4]'
              }`}
            >
              <Globe size={13} className={assetClass === 'world' ? 'text-[#1f8a65]' : ''} />
              <span>World Stocks ({worldStocks.length})</span>
            </button>
            <button
              onClick={() => setAssetClass('indices')}
              className={`px-3 py-1.5 rounded-[7px] text-[12px] font-medium transition-all flex items-center gap-1.5 shrink-0 ${
                assetClass === 'indices'
                  ? 'bg-white dark:bg-[#26251e] text-[#26251e] dark:text-[#f7f7f4] border border-[#e6e5e0] dark:border-[#383730]'
                  : 'text-[#5a5852] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#f7f7f4]'
              }`}
            >
              <BarChart3 size={13} className={assetClass === 'indices' ? 'text-[#c08532]' : ''} />
              <span>Country Indices ({countryIndices.length})</span>
            </button>
            <button
              onClick={() => setAssetClass('bonds')}
              className={`px-3 py-1.5 rounded-[7px] text-[12px] font-medium transition-all flex items-center gap-1.5 shrink-0 ${
                assetClass === 'bonds'
                  ? 'bg-white dark:bg-[#26251e] text-[#26251e] dark:text-[#f7f7f4] border border-[#e6e5e0] dark:border-[#383730]'
                  : 'text-[#5a5852] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#f7f7f4]'
              }`}
            >
              <Landmark size={13} className={assetClass === 'bonds' ? 'text-[#1976d2]' : ''} />
              <span>Bonds ({bonds.length})</span>
            </button>
            <button
              onClick={() => setAssetClass('forex')}
              className={`px-3 py-1.5 rounded-[7px] text-[12px] font-medium transition-all flex items-center gap-1.5 shrink-0 ${
                assetClass === 'forex'
                  ? 'bg-white dark:bg-[#26251e] text-[#26251e] dark:text-[#f7f7f4] border border-[#e6e5e0] dark:border-[#383730]'
                  : 'text-[#5a5852] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#f7f7f4]'
              }`}
            >
              <DollarSign size={13} className={assetClass === 'forex' ? 'text-[#f54e00]' : ''} />
              <span>Currencies & FX ({forexPairs.length})</span>
            </button>
          </div>
        </div>
      </div>

      {/* Filter controls row (responsive stacked on mobile) */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-white dark:bg-[#1c1b18] p-3 rounded-[10px] border border-[#e6e5e0] dark:border-[#2c2b26]">
        <div className="flex flex-col sm:flex-row sm:items-center gap-2.5 w-full sm:w-auto">
          {/* Search Input */}
          <div className="relative w-full sm:min-w-[240px]">
            <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-[#807d72] dark:text-[#a09c92]" />
            <input
              type="text"
              placeholder={
                assetClass === 'domestic'
                  ? 'Search ticker or company...'
                  : assetClass === 'world'
                  ? 'Search world equities...'
                  : assetClass === 'bonds'
                  ? 'Search bonds...'
                  : 'Search currency pairs...'
              }
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full bg-[#fafaf7] dark:bg-[#26251e] border border-[#e6e5e0] dark:border-[#383730] focus:border-[#26251e] dark:focus:border-[#f7f7f4] text-[13px] rounded-[8px] pl-9 pr-3 h-9 text-[#26251e] dark:text-[#f7f7f4] placeholder-[#a09c92] outline-none transition-colors"
            />
          </div>

          {/* Sector and Tier dropdowns for domestic stocks */}
          {assetClass === 'domestic' && (
            <div className="flex flex-wrap items-center gap-2 w-full sm:w-auto">
              <CustomDropdown
                value={selectedSector}
                onChange={setSelectedSector}
                options={sectorOptions}
                placeholder="All Sectors"
                className="flex-1 sm:flex-initial sm:min-w-[150px]"
              />

              <CustomDropdown
                value={selectedTier}
                onChange={setSelectedTier}
                options={tierOptions}
                placeholder="All Tiers & Ventures"
                className="flex-1 sm:flex-initial sm:min-w-[190px]"
              />
            </div>
          )}

          {/* Country select for world stocks */}
          {assetClass === 'world' && (
            <div className="w-full sm:w-auto">
              <CustomDropdown
                value={selectedCountry}
                onChange={setSelectedCountry}
                options={countryOptions}
                placeholder="All Foreign Nations"
                className="w-full sm:min-w-[190px]"
              />
            </div>
          )}
        </div>

        <div className="text-[12px] font-mono text-[#807d72] dark:text-[#a09c92]">
          {assetClass === 'domestic' && `${filteredStocks.length} listed equities`}
          {assetClass === 'world' && `${filteredWorldStocks.length} foreign equities`}
          {assetClass === 'indices' && `${filteredIndices.length} country indices`}
          {assetClass === 'bonds' && `${bonds.length} bonds`}
          {assetClass === 'forex' && `${filteredForex.length} currency pairs`}
        </div>
      </div>

      {/* 1. DOMESTIC STOCKS VIEW */}
      {assetClass === 'domestic' && (
        <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] overflow-hidden">
          {/* Mobile Card List (< 640px) */}
          <div className="block sm:hidden divide-y divide-[#efeee8] dark:divide-[#2c2b26]">
            {filteredStocks.map((stock) => (
              <div
                key={stock.symbol}
                onClick={() => !stock.is_private && onSelectStock(stock.symbol)}
                className={`p-3.5 flex items-center justify-between ${stock.is_private ? 'opacity-75' : 'active:bg-[#fafaf7] dark:active:bg-[#201f1b]'}`}
              >
                <div className="space-y-1 max-w-[60%]">
                  <div className="flex items-center gap-2">
                    <span className="font-mono font-semibold text-[#26251e] dark:text-[#f7f7f4] text-[14px]">
                      {stock.symbol}
                    </span>
                    {stock.is_private ? (
                      <span className="text-[9px] font-mono px-1.5 py-0.5 rounded-[4px] bg-[#fdf3e7] dark:bg-[#382b15] text-[#b45309] dark:text-[#f59e0b] border border-[#fde68a] dark:border-[#78350f]">
                        PRIVATE
                      </span>
                    ) : stock.is_ipo ? (
                      <span className="text-[9px] font-mono px-1.5 py-0.5 rounded-[4px] bg-[#e8f5e9] dark:bg-[#153820] text-[#2e7d32] dark:text-[#4ade80] border border-[#c8e6c9] dark:border-[#166534]">
                        NEW IPO
                      </span>
                    ) : null}
                  </div>
                  <div className="text-[12px] text-[#5a5852] dark:text-[#a09c92] truncate">
                    {stock.name}
                  </div>
                  <div className="text-[11px] font-mono text-[#807d72] dark:text-[#737064]">
                    {stock.sector} · {formatB(stock.is_private && stock.private_valuation ? stock.private_valuation : stock.market_cap)}
                  </div>
                </div>

                <div className="text-right space-y-1">
                  <div className="font-mono font-semibold text-[#26251e] dark:text-[#f7f7f4] text-[14px]">
                    {stock.is_private ? 'Private' : `$${stock.mid > 0 ? stock.mid.toFixed(2) : stock.reported_value.toFixed(2)}`}
                  </div>
                  {!stock.is_private && (
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        onSelectStock(stock.symbol);
                      }}
                      className="text-[11px] font-mono px-2 py-0.5 rounded-[4px] bg-[#26251e] dark:bg-[#f7f7f4] text-white dark:text-[#121210] font-medium"
                    >
                      Trade →
                    </button>
                  )}
                </div>
              </div>
            ))}
            {filteredStocks.length === 0 && (
              <div className="py-8 text-center text-[#807d72] dark:text-[#a09c92] text-[13px]">
                No domestic companies match your search.
              </div>
            )}
          </div>

          {/* Desktop Table (>= 640px) */}
          <div className="hidden sm:block overflow-x-auto">
            <table className="w-full text-left border-collapse min-w-[700px]">
              <thead>
                <tr className="border-b border-[#e6e5e0] dark:border-[#2c2b26] bg-[#fafaf7] dark:bg-[#201f1b] text-[11px] font-mono uppercase text-[#807d72] dark:text-[#a09c92]">
                  <th className="py-3 px-3 lg:px-4 font-medium cursor-pointer" onClick={() => handleSort('symbol')}>
                    <div className="flex items-center gap-1">
                      <span>Symbol</span>
                      <ArrowUpDown size={12} />
                    </div>
                  </th>
                  <th className="py-3 px-3 lg:px-4 font-medium cursor-pointer" onClick={() => handleSort('name')}>
                    <div className="flex items-center gap-1">
                      <span>Company</span>
                      <ArrowUpDown size={12} />
                    </div>
                  </th>
                  <th className="py-3 px-3 lg:px-4 font-medium hidden md:table-cell">Sector</th>
                  <th className="py-3 px-3 lg:px-4 font-medium hidden lg:table-cell">Status</th>
                  <th className="py-3 px-3 lg:px-4 font-medium text-right cursor-pointer" onClick={() => handleSort('mid')}>
                    <div className="flex items-center justify-end gap-1">
                      <span>Price</span>
                      <ArrowUpDown size={12} />
                    </div>
                  </th>
                  <th className="py-3 px-3 lg:px-4 font-medium text-right hidden md:table-cell">Spread</th>
                  <th className="py-3 px-3 lg:px-4 font-medium text-right cursor-pointer" onClick={() => handleSort('market_cap')}>
                    <div className="flex items-center justify-end gap-1">
                      <span>Market Cap</span>
                      <ArrowUpDown size={12} />
                    </div>
                  </th>
                  <th className="py-3 px-3 lg:px-4 font-medium text-right hidden xl:table-cell cursor-pointer" onClick={() => handleSort('pe_ratio')}>
                    <div className="flex items-center justify-end gap-1">
                      <span>P/E</span>
                      <ArrowUpDown size={12} />
                    </div>
                  </th>
                  <th className="py-3 px-3 lg:px-4 font-medium text-right">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#efeee8] dark:divide-[#2c2b26] text-[13px]">
                {filteredStocks.map((stock) => (
                  <tr
                    key={stock.symbol}
                    onClick={() => !stock.is_private && onSelectStock(stock.symbol)}
                    className={`hover:bg-[#fafaf7] dark:hover:bg-[#201f1b] transition-colors group ${stock.is_private ? 'cursor-default opacity-80' : 'cursor-pointer'}`}
                  >
                    <td className="py-3 px-3 lg:px-4 font-mono font-semibold text-[#26251e] dark:text-[#f7f7f4]">
                      {stock.symbol}
                    </td>
                    <td className="py-3 px-3 lg:px-4 text-[#26251e] dark:text-[#f7f7f4] font-medium max-w-[180px] lg:max-w-[220px] truncate">
                      {stock.name}
                    </td>
                    <td className="py-3 px-3 lg:px-4 text-[#5a5852] dark:text-[#a09c92] hidden md:table-cell">{stock.sector}</td>
                    <td className="py-3 px-3 lg:px-4 hidden lg:table-cell">
                      {stock.is_private ? (
                        <span className="text-[10px] font-mono px-2 py-0.5 rounded-[4px] bg-[#fdf3e7] dark:bg-[#382b15] text-[#b45309] dark:text-[#f59e0b] border border-[#fde68a] dark:border-[#78350f]">
                          {stock.stage || 'Pre-IPO'}
                        </span>
                      ) : stock.is_ipo ? (
                        <span className="text-[10px] font-mono px-2 py-0.5 rounded-[4px] bg-[#e8f5e9] dark:bg-[#153820] text-[#2e7d32] dark:text-[#4ade80] border border-[#c8e6c9] dark:border-[#166534]">
                          NEW IPO
                        </span>
                      ) : (
                        <span className="text-[10px] font-mono px-2 py-0.5 rounded-[4px] bg-[#efeee8] dark:bg-[#26251e] text-[#5a5852] dark:text-[#a09c92]">
                          {stock.cap_tier}
                        </span>
                      )}
                    </td>
                    <td className="py-3 px-3 lg:px-4 font-mono text-right text-[#26251e] dark:text-[#f7f7f4] font-medium">
                      {stock.is_private ? (
                        <span className="text-[10px] font-mono px-2 py-0.5 rounded-[4px] bg-[#f3f4f6] dark:bg-[#26251e] text-[#6b7280] dark:text-[#a09c92] border border-[#e5e7eb] dark:border-[#383730] align-middle">
                          PRIVATE
                        </span>
                      ) : (
                        `$${stock.mid > 0 ? stock.mid.toFixed(2) : stock.reported_value.toFixed(2)}`
                      )}
                    </td>
                    <td className="py-3 px-3 lg:px-4 font-mono text-right text-[#807d72] dark:text-[#737064] hidden md:table-cell">
                      {stock.is_private ? '—' : `$${stock.spread.toFixed(2)}`}
                    </td>
                    <td className="py-3 px-3 lg:px-4 font-mono text-right text-[#26251e] dark:text-[#f7f7f4]">
                      {formatB(stock.is_private && stock.private_valuation ? stock.private_valuation : stock.market_cap)}
                    </td>
                    <td className="py-3 px-3 lg:px-4 font-mono text-right text-[#5a5852] dark:text-[#a09c92] hidden xl:table-cell">
                      {stock.is_private ? '—' : (stock.pe_ratio > 0 ? stock.pe_ratio.toFixed(1) : '—')}
                    </td>
                    <td className="py-3 px-3 lg:px-4 text-right">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          if (!stock.is_private) onSelectStock(stock.symbol);
                        }}
                        disabled={stock.is_private}
                        className={`text-[12px] font-medium flex items-center gap-1 ml-auto ${
                          stock.is_private
                            ? 'text-[#c0bdb5] dark:text-[#4a4943] cursor-not-allowed'
                            : 'text-[#26251e] dark:text-[#f7f7f4] group-hover:text-[#f54e00]'
                        }`}
                      >
                        <span>Trade</span>
                        <ChevronRight size={14} />
                      </button>
                    </td>
                  </tr>
                ))}
                {filteredStocks.length === 0 && (
                  <tr>
                    <td colSpan={9} className="py-8 text-center text-[#807d72] dark:text-[#a09c92]">
                      No domestic companies match your search.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* 2. WORLD STOCKS VIEW */}
      {assetClass === 'world' && (
        <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] overflow-hidden">
          {/* Mobile Card List (< 640px) */}
          <div className="block sm:hidden divide-y divide-[#efeee8] dark:divide-[#2c2b26]">
            {filteredWorldStocks.map((ws) => {
              const isUp = ws.change_pct >= 0;
              return (
                <div
                  key={ws.ticker}
                  onClick={() => onSelectStock(ws.ticker)}
                  className="p-3.5 flex items-center justify-between active:bg-[#fafaf7] dark:active:bg-[#201f1b]"
                >
                  <div className="space-y-1 max-w-[60%]">
                    <div className="flex items-center gap-2">
                      <span className="font-mono font-semibold text-[#26251e] dark:text-[#f7f7f4] text-[14px]">
                        {ws.ticker}
                      </span>
                      <span className="text-[10px] font-mono px-1.5 py-0.5 rounded-[4px] bg-[#eef2f6] dark:bg-[#1e293b] text-[#1e3a8a] dark:text-[#93c5fd] border border-[#dbeafe] dark:border-[#1e3a8a]">
                        {ws.country_id}
                      </span>
                    </div>
                    <div className="text-[12px] text-[#5a5852] dark:text-[#a09c92] truncate">
                      {ws.name}
                    </div>
                    <div className="text-[11px] font-mono text-[#807d72] dark:text-[#737064]">
                      {ws.sector} · {formatB(ws.market_cap)}
                    </div>
                  </div>

                  <div className="text-right space-y-1">
                    <div className="font-mono font-semibold text-[#26251e] dark:text-[#f7f7f4] text-[14px]">
                      ${ws.price.toFixed(2)}
                    </div>
                    <div className={`text-[12px] font-mono font-medium ${isUp ? 'text-[#1f8a65]' : 'text-[#cf2d56]'}`}>
                      {isUp ? `+${ws.change_pct.toFixed(2)}%` : `${ws.change_pct.toFixed(2)}%`}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>

          {/* Desktop Table (>= 640px) */}
          <div className="hidden sm:block overflow-x-auto">
            <table className="w-full text-left border-collapse min-w-[700px]">
              <thead>
                <tr className="border-b border-[#e6e5e0] dark:border-[#2c2b26] bg-[#fafaf7] dark:bg-[#201f1b] text-[11px] font-mono uppercase text-[#807d72] dark:text-[#a09c92]">
                  <th className="py-3 px-3 lg:px-4 font-medium">Symbol</th>
                  <th className="py-3 px-3 lg:px-4 font-medium">Company</th>
                  <th className="py-3 px-3 lg:px-4 font-medium">Country</th>
                  <th className="py-3 px-3 lg:px-4 font-medium hidden md:table-cell">Sector</th>
                  <th className="py-3 px-3 lg:px-4 font-medium text-right">Price</th>
                  <th className="py-3 px-3 lg:px-4 font-medium text-right">24h Change</th>
                  <th className="py-3 px-3 lg:px-4 font-medium text-right">Market Cap</th>
                  <th className="py-3 px-3 lg:px-4 font-medium text-right hidden xl:table-cell">P/E</th>
                  <th className="py-3 px-3 lg:px-4 font-medium text-right">Chart</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#efeee8] dark:divide-[#2c2b26] text-[13px]">
                {filteredWorldStocks.map((ws) => {
                  const isUp = ws.change_pct >= 0;
                  return (
                    <tr
                      key={ws.ticker}
                      onClick={() => onSelectStock(ws.ticker)}
                      className="hover:bg-[#fafaf7] dark:hover:bg-[#201f1b] transition-colors cursor-pointer group"
                    >
                      <td className="py-3 px-3 lg:px-4 font-mono font-semibold text-[#26251e] dark:text-[#f7f7f4]">
                        {ws.ticker}
                      </td>
                      <td className="py-3 px-3 lg:px-4 text-[#26251e] dark:text-[#f7f7f4] font-medium max-w-[200px] truncate">
                        {ws.name}
                      </td>
                      <td className="py-3 px-3 lg:px-4">
                        <span className="text-[11px] font-mono px-2 py-0.5 rounded-[4px] bg-[#eef2f6] dark:bg-[#1e293b] text-[#1e3a8a] dark:text-[#93c5fd] border border-[#dbeafe] dark:border-[#1e3a8a]">
                          {ws.country_id} · {ws.country_name}
                        </span>
                      </td>
                      <td className="py-3 px-3 lg:px-4 text-[#5a5852] dark:text-[#a09c92] hidden md:table-cell">{ws.sector}</td>
                      <td className="py-3 px-3 lg:px-4 font-mono text-right text-[#26251e] dark:text-[#f7f7f4] font-medium">
                        ${ws.price.toFixed(2)}
                      </td>
                      <td className={`py-3 px-3 lg:px-4 font-mono text-right font-medium ${isUp ? 'text-[#1f8a65]' : 'text-[#cf2d56]'}`}>
                        {isUp ? `+${ws.change_pct.toFixed(2)}%` : `${ws.change_pct.toFixed(2)}%`}
                      </td>
                      <td className="py-3 px-3 lg:px-4 font-mono text-right text-[#26251e] dark:text-[#f7f7f4]">
                        {formatB(ws.market_cap)}
                      </td>
                      <td className="py-3 px-3 lg:px-4 font-mono text-right text-[#5a5852] dark:text-[#a09c92] hidden xl:table-cell">
                        {ws.pe_ratio.toFixed(1)}
                      </td>
                      <td className="py-3 px-3 lg:px-4 text-right">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            onSelectStock(ws.ticker);
                          }}
                          className="px-2.5 py-1 text-[11px] font-mono rounded-[6px] bg-[#26251e] dark:bg-[#f7f7f4] text-white dark:text-[#121210] hover:bg-[#3d3b32] dark:hover:bg-[#eae8e0] transition-colors inline-flex items-center gap-1.5"
                        >
                          <LineChart size={12} />
                          <span>Chart</span>
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* 3. COUNTRY BENCHMARK INDICES VIEW */}
      {assetClass === 'indices' && (
        <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] overflow-hidden">
          <div className="p-4 border-b border-[#efeee8] dark:border-[#2c2b26] flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <div>
              <h3 className="text-[15px] font-semibold text-[#26251e] dark:text-[#f7f7f4]">Country Benchmark Indices</h3>
              <p className="text-[12px] text-[#807d72] dark:text-[#a09c92] mt-0.5">
                Market-weighted indices tracking overall equity performance. Click any index to view live candlestick charts.
              </p>
            </div>
            <span className="text-[11px] font-mono text-[#807d72] dark:text-[#a09c92] bg-[#fafaf7] dark:bg-[#26251e] px-2.5 py-1 rounded-[6px] border border-[#e6e5e0] dark:border-[#383730] self-start sm:self-auto">
              {filteredIndices.length} Benchmark Indices
            </span>
          </div>

          {/* Mobile Card List for Indices (< 640px) */}
          <div className="block sm:hidden divide-y divide-[#efeee8] dark:divide-[#2c2b26]">
            {filteredIndices.map((idx) => {
              const isUp = idx.change_24h >= 0;
              return (
                <div
                  key={idx.symbol}
                  onClick={() => onSelectStock(idx.symbol)}
                  className="p-4 hover:bg-[#fafaf7] dark:hover:bg-[#201f1b] transition-colors flex items-center justify-between gap-3 cursor-pointer"
                >
                  <div className="space-y-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="font-mono font-semibold text-[#26251e] dark:text-[#f7f7f4] text-[14px]">
                        {idx.symbol}
                      </span>
                      <span className="text-[10px] font-mono px-1.5 py-0.5 rounded-[4px] bg-[#efeee8] dark:bg-[#26251e] text-[#26251e] dark:text-[#f7f7f4] font-semibold border border-[#e6e5e0] dark:border-[#383730]">
                        {idx.country_id}
                      </span>
                    </div>
                    <div className="text-[12px] text-[#5a5852] dark:text-[#a09c92] truncate max-w-[200px]">
                      {idx.name}
                    </div>
                    <div className="text-[11px] font-mono text-[#807d72] dark:text-[#737064]">
                      Vol: {formatB(idx.volume || 45_000_000)} · {idx.constituents_count} Equities
                    </div>
                  </div>
                  <div className="text-right space-y-1 flex-shrink-0">
                    <div className="font-mono font-semibold text-[#26251e] dark:text-[#f7f7f4] text-[14px]">
                      ${idx.price.toFixed(2)}
                    </div>
                    <div className={`text-[11px] font-mono font-medium ${isUp ? 'text-[#1f8a65]' : 'text-[#cf2d56]'}`}>
                      {isUp ? `+${idx.change_24h.toFixed(2)}%` : `${idx.change_24h.toFixed(2)}%`}
                    </div>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        onSelectStock(idx.symbol);
                      }}
                      className="px-2.5 py-0.5 text-[10px] font-mono rounded-[4px] bg-[#26251e] dark:bg-[#f7f7f4] text-white dark:text-[#121210] font-medium"
                    >
                      Chart →
                    </button>
                  </div>
                </div>
              );
            })}
          </div>

          {/* Desktop Table (>= 640px) */}
          <div className="hidden sm:block overflow-x-auto">
            <table className="w-full text-left border-collapse min-w-[700px]">
              <thead>
                <tr className="border-b border-[#e6e5e0] dark:border-[#2c2b26] bg-[#fafaf7] dark:bg-[#201f1b] text-[11px] font-mono uppercase text-[#807d72] dark:text-[#a09c92]">
                  <th className="py-3 px-4 font-medium">Index Symbol</th>
                  <th className="py-3 px-4 font-medium">Benchmark Name</th>
                  <th className="py-3 px-4 font-medium">Country</th>
                  <th className="py-3 px-4 font-medium text-right">Price</th>
                  <th className="py-3 px-4 font-medium text-right">24h Change</th>
                  <th className="py-3 px-4 font-medium text-right">24h Volume</th>
                  <th className="py-3 px-4 font-medium text-right hidden md:table-cell">Day High / Low</th>
                  <th className="py-3 px-4 font-medium text-right hidden lg:table-cell">Constituents</th>
                  <th className="py-3 px-4 font-medium text-right">Chart</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#efeee8] dark:divide-[#2c2b26] text-[13px]">
                {filteredIndices.map((idx) => {
                  const isUp = idx.change_24h >= 0;
                  return (
                    <tr key={idx.symbol} className="hover:bg-[#fafaf7] dark:hover:bg-[#201f1b] transition-colors">
                      <td className="py-3.5 px-4 font-mono font-semibold text-[#26251e] dark:text-[#f7f7f4] text-[14px]">
                        {idx.symbol}
                      </td>
                      <td className="py-3.5 px-4 font-medium text-[#26251e] dark:text-[#f7f7f4]">
                        {idx.name}
                      </td>
                      <td className="py-3.5 px-4">
                        <span className="text-[11px] font-mono px-2 py-0.5 rounded-[4px] bg-[#efeee8] dark:bg-[#26251e] text-[#26251e] dark:text-[#f7f7f4] font-semibold border border-[#e6e5e0] dark:border-[#383730]">
                          {idx.country_id}
                        </span>
                      </td>
                      <td className="py-3.5 px-4 font-mono text-right text-[#26251e] dark:text-[#f7f7f4] font-bold text-[14px]">
                        ${idx.price.toFixed(2)}
                      </td>
                      <td className={`py-3.5 px-4 font-mono text-right font-semibold ${isUp ? 'text-[#1f8a65]' : 'text-[#cf2d56]'}`}>
                        {isUp ? `+${idx.change_24h.toFixed(2)}%` : `${idx.change_24h.toFixed(2)}%`}
                      </td>
                      <td className="py-3.5 px-4 font-mono text-right text-[#26251e] dark:text-[#f7f7f4] font-medium text-[13px]">
                        {formatB(idx.volume || 45_000_000)}
                      </td>
                      <td className="py-3.5 px-4 font-mono text-right text-[#807d72] dark:text-[#737064] text-[12px] hidden md:table-cell">
                        ${(idx.high_24h || idx.price).toFixed(2)} / ${(idx.low_24h || idx.price).toFixed(2)}
                      </td>
                      <td className="py-3.5 px-4 font-mono text-right text-[#5a5852] dark:text-[#a09c92] hidden lg:table-cell">
                        {idx.constituents_count} Equities
                      </td>
                      <td className="py-3.5 px-4 text-right">
                        <button
                          onClick={() => onSelectStock(idx.symbol)}
                          className="px-3 py-1 text-[11px] font-mono rounded-[6px] bg-[#26251e] dark:bg-[#f7f7f4] text-white dark:text-[#121210] hover:bg-[#3d3b32] dark:hover:bg-[#eae8e0] transition-colors inline-flex items-center gap-1.5"
                        >
                          <LineChart size={12} />
                          <span>Chart</span>
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* 4. BONDS VIEW */}
      {assetClass === 'bonds' && (
        <div className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
            {bonds.map((bond) => {
              const ytmPct = (bond.yield_to_maturity * 100).toFixed(2);
              const couponPct = (bond.coupon_rate * 100).toFixed(2);
              return (
                <div
                  key={bond.id}
                  className="bg-white dark:bg-[#1c1b18] p-5 rounded-[12px] border border-[#e6e5e0] dark:border-[#2c2b26] flex flex-col justify-between hover:border-[#26251e] dark:hover:border-[#f7f7f4] transition-all"
                >
                  <div>
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-[11px] font-mono px-2 py-0.5 rounded-[4px] bg-[#eef2f6] dark:bg-[#1e293b] text-[#1976d2] dark:text-[#93c5fd] font-semibold border border-[#dbeafe] dark:border-[#1e3a8a]">
                        {bond.id}
                      </span>
                      <span className="text-[11px] font-mono text-[#807d72] dark:text-[#a09c92]">{bond.maturity_years}Y Tenor</span>
                    </div>
                    <div className="text-[16px] font-medium text-[#26251e] dark:text-[#f7f7f4] leading-snug mb-3">
                      {bond.name}
                    </div>
                    <div className="space-y-2 border-t border-[#f0eee6] dark:border-[#2c2b26] pt-3 text-[13px]">
                      <div className="flex justify-between">
                        <span className="text-[#807d72] dark:text-[#a09c92]">Yield to Maturity (YTM)</span>
                        <span className="font-mono font-semibold text-[#1976d2] dark:text-[#60a5fa]">{ytmPct}%</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-[#807d72] dark:text-[#a09c92]">Coupon Rate</span>
                        <span className="font-mono text-[#26251e] dark:text-[#f7f7f4]">{couponPct}%</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-[#807d72] dark:text-[#a09c92]">Par Value</span>
                        <span className="font-mono text-[#26251e] dark:text-[#f7f7f4]">${bond.price.toFixed(2)}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-[#807d72] dark:text-[#a09c92]">Outstanding Debt</span>
                        <span className="font-mono font-medium text-[#26251e] dark:text-[#f7f7f4]">{formatB(bond.outstanding_amount)}</span>
                      </div>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-2 mt-4 pt-3 border-t border-[#f0eee6] dark:border-[#2c2b26]">
                    <button
                      onClick={() => setBondModal({ open: true, type: 'issue', bond })}
                      className="h-8 rounded-[6px] bg-[#f54e00] hover:bg-[#d44300] text-white text-[12px] font-medium flex items-center justify-center gap-1 transition-colors"
                    >
                      <PlusCircle size={13} />
                      <span>Issue Bonds</span>
                    </button>
                    <button
                      onClick={() => setBondModal({ open: true, type: 'buyback', bond })}
                      className="h-8 rounded-[6px] bg-[#f5f5f2] dark:bg-[#26251e] hover:bg-[#eae8e0] dark:hover:bg-[#383730] text-[#26251e] dark:text-[#f7f7f4] text-[12px] font-medium flex items-center justify-center gap-1 transition-colors border border-[#e6e5e0] dark:border-[#383730]"
                    >
                      <MinusCircle size={13} />
                      <span>Repay</span>
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* 5. FOREX & CURRENCIES VIEW */}
      {assetClass === 'forex' && (
        <div className="space-y-4">
          {/* Global Reserve Currency Banner */}
          <div className="bg-[#f0fdf4] dark:bg-[#152e1e] border border-[#bbf7d0] dark:border-[#166534] rounded-[12px] p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div className="flex items-start sm:items-center gap-3">
              <div className="w-9 h-9 rounded-[8px] bg-white dark:bg-[#1c1b18] border border-[#bbf7d0] dark:border-[#166534] flex items-center justify-center text-[#166534] dark:text-[#4ade80] shrink-0">
                <ShieldCheck size={20} />
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <span className="text-[13px] font-semibold text-[#166534] dark:text-[#4ade80]">
                    Global Reserve Currency: ZTH (Zenthian Federation)
                  </span>
                  <span className="text-[10px] font-mono uppercase px-2 py-0.5 rounded-[4px] bg-[#dcfce7] dark:bg-[#14532d] text-[#15803d] dark:text-[#86efac] border border-[#86efac] dark:border-[#166534] font-medium">
                    World Standard
                  </span>
                </div>
                <p className="text-[12px] text-[#166534]/80 dark:text-[#86efac]/80 mt-0.5">
                  As the world's leading superpower, Zenthian Federation's currency (ZTH) is the global reserve asset. Cross-border trade, bilateral accords, and foreign reserves settle in ZTH.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse min-w-[650px]">
                <thead>
                  <tr className="border-b border-[#e6e5e0] dark:border-[#2c2b26] bg-[#fafaf7] dark:bg-[#201f1b] text-[11px] font-mono uppercase text-[#807d72] dark:text-[#a09c92]">
                    <th className="py-3 px-4 font-medium">Currency Pair</th>
                    <th className="py-3 px-4 font-medium">Base / Quote</th>
                    <th className="py-3 px-4 font-medium text-right">Exchange Rate</th>
                    <th className="py-3 px-4 font-medium text-right">24h Change</th>
                    <th className="py-3 px-4 font-medium text-right hidden sm:table-cell">24h High</th>
                    <th className="py-3 px-4 font-medium text-right hidden sm:table-cell">24h Low</th>
                    <th className="py-3 px-4 font-medium text-right">Action</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-[#efeee8] dark:divide-[#2c2b26] text-[13px]">
                  {filteredForex.map((fx) => {
                    const isUp = fx.change_24h >= 0;
                    const isReserve = fx.quote_currency === 'ZTH';
                    return (
                      <tr key={fx.symbol} className={`hover:bg-[#fafaf7] dark:hover:bg-[#201f1b] transition-colors ${isReserve ? 'bg-[#fafaf7]/50 dark:bg-[#1c1b18]' : ''}`}>
                        <td className="py-3 px-4 font-mono font-semibold text-[#26251e] dark:text-[#f7f7f4] text-[14px]">
                          <div className="flex items-center gap-2">
                            <span>{fx.symbol}</span>
                            {isReserve && (
                              <span className="text-[10px] font-mono px-1.5 py-0.5 rounded-[4px] bg-[#dcfce7] dark:bg-[#14532d] text-[#15803d] dark:text-[#86efac] border border-[#86efac] dark:border-[#166534]">
                                RESERVE
                              </span>
                            )}
                          </div>
                        </td>
                        <td className="py-3 px-4 text-[#5a5852] dark:text-[#a09c92]">
                          {fx.base_currency} against {fx.quote_currency}
                        </td>
                        <td className="py-3 px-4 font-mono text-right text-[#26251e] dark:text-[#f7f7f4] font-semibold text-[14px]">
                          {fx.rate.toFixed(4)}
                        </td>
                        <td className={`py-3 px-4 font-mono text-right font-medium ${isUp ? 'text-[#1f8a65]' : 'text-[#cf2d56]'}`}>
                          {isUp ? `+${fx.change_24h.toFixed(2)}%` : `${fx.change_24h.toFixed(2)}%`}
                        </td>
                        <td className="py-3 px-4 font-mono text-right text-[#807d72] dark:text-[#737064] hidden sm:table-cell">
                          {fx.high_24h.toFixed(4)}
                        </td>
                        <td className="py-3 px-4 font-mono text-right text-[#807d72] dark:text-[#737064] hidden sm:table-cell">
                          {fx.low_24h.toFixed(4)}
                        </td>
                        <td className="py-3 px-4 text-right">
                          <div className="flex items-center justify-end gap-1.5">
                            <button
                              onClick={() => onSelectStock(fx.symbol)}
                              className="h-7 px-2.5 rounded-[6px] bg-[#fafaf7] dark:bg-[#26251e] hover:bg-[#efeee8] dark:hover:bg-[#383730] text-[#26251e] dark:text-[#f7f7f4] border border-[#e6e5e0] dark:border-[#383730] text-[11px] font-mono inline-flex items-center gap-1 transition-colors"
                              title="Open Candlestick Chart in Terminal"
                            >
                              <LineChart size={11} />
                              <span>Chart</span>
                            </button>
                            <button
                              onClick={() => setForexModal({ open: true, pair: fx })}
                              className="h-7 px-3 rounded-[6px] bg-[#26251e] dark:bg-[#f7f7f4] hover:bg-[#3d3b32] dark:hover:bg-[#eae8e0] text-white dark:text-[#121210] text-[11px] font-medium inline-flex items-center gap-1 transition-colors"
                            >
                              <RefreshCw size={11} />
                              <span>Swap</span>
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* BOND ACTION MODAL */}
      {bondModal.open && bondModal.bond && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white dark:bg-[#1c1b18] rounded-[16px] border border-[#e6e5e0] dark:border-[#2c2b26] max-w-[440px] w-full p-6 space-y-4">
            <div className="flex items-center justify-between border-b border-[#efeee8] dark:border-[#2c2b26] pb-3">
              <div>
                <h3 className="text-[17px] font-medium text-[#26251e] dark:text-[#f7f7f4]">
                  {bondModal.type === 'issue' ? 'Issue Treasury Bonds' : 'Repurchase Bonds'}
                </h3>
                <div className="text-[12px] font-mono text-[#807d72] dark:text-[#a09c92]">{bondModal.bond.name}</div>
              </div>
              <button
                onClick={() => setBondModal({ open: false, type: 'issue', bond: null })}
                className="text-[#807d72] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#f7f7f4] p-1 rounded-[6px] hover:bg-[#efeee8] dark:hover:bg-[#26251e] transition-colors"
                aria-label="Close Modal"
              >
                <X size={18} />
              </button>
            </div>

            <div className="space-y-3 text-[13px]">
              <div className="bg-[#fafaf7] dark:bg-[#201f1b] p-3 rounded-[8px] space-y-1 text-[#5a5852] dark:text-[#a09c92] font-mono text-[12px] border border-[#efeee8] dark:border-[#2c2b26]">
                <div className="flex justify-between">
                  <span>Tenor / Maturity:</span>
                  <span className="font-semibold text-[#26251e] dark:text-[#f7f7f4]">{bondModal.bond.maturity_years} Years</span>
                </div>
                <div className="flex justify-between">
                  <span>Borrowing Yield (YTM):</span>
                  <span className="font-semibold text-[#1976d2] dark:text-[#60a5fa]">{(bondModal.bond.yield_to_maturity * 100).toFixed(2)}%</span>
                </div>
                <div className="flex justify-between">
                  <span>Current Outstanding:</span>
                  <span className="font-semibold text-[#26251e] dark:text-[#f7f7f4]">{formatB(bondModal.bond.outstanding_amount)}</span>
                </div>
              </div>

              <div>
                <label className="block text-[12px] font-medium text-[#26251e] dark:text-[#f7f7f4] mb-1">
                  Amount in Home Currency (CRN)
                </label>
                <div className="relative">
                  <input
                    type="number"
                    value={bondAmount}
                    onChange={(e) => setBondAmount(e.target.value)}
                    className="w-full bg-[#fafaf7] dark:bg-[#26251e] border border-[#e6e5e0] dark:border-[#383730] focus:border-[#26251e] dark:focus:border-[#f7f7f4] h-10 px-3 rounded-[8px] text-[14px] font-mono text-[#26251e] dark:text-[#f7f7f4] outline-none"
                  />
                  <div className="text-[11px] text-[#807d72] dark:text-[#a09c92] mt-1 font-mono">
                    Formatted: {formatB(parseFloat(bondAmount) || 0)}
                  </div>
                </div>
              </div>

              {bondFeedback && (
                <div className="p-2.5 rounded-[6px] bg-[#eef2f6] dark:bg-[#1e293b] text-[12px] font-mono text-[#1976d2] dark:text-[#93c5fd] border border-[#dbeafe] dark:border-[#1e3a8a]">
                  {bondFeedback}
                </div>
              )}
            </div>

            <div className="flex items-center justify-end gap-2 pt-2 border-t border-[#efeee8] dark:border-[#2c2b26]">
              <button
                onClick={() => setBondModal({ open: false, type: 'issue', bond: null })}
                className="px-4 py-2 rounded-[8px] text-[13px] font-medium text-[#5a5852] dark:text-[#a09c92] hover:bg-[#fafaf7] dark:hover:bg-[#26251e]"
              >
                Cancel
              </button>
              <button
                onClick={handleExecuteBond}
                disabled={bondBusy}
                className="px-4 py-2 rounded-[8px] text-[13px] font-medium text-white bg-[#f54e00] hover:bg-[#d44300] disabled:opacity-50"
              >
                {bondBusy ? 'Submitting...' : bondModal.type === 'issue' ? 'Issue Bonds' : 'Repurchase & Retire'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* FOREX SWAP MODAL */}
      {forexModal.open && forexModal.pair && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white dark:bg-[#1c1b18] rounded-[16px] border border-[#e6e5e0] dark:border-[#2c2b26] max-w-[420px] w-full p-6 space-y-4">
            <div className="flex items-center justify-between border-b border-[#efeee8] dark:border-[#2c2b26] pb-3">
              <div>
                <h3 className="text-[17px] font-medium text-[#26251e] dark:text-[#f7f7f4]">Currency Swap</h3>
                <div className="text-[12px] font-mono text-[#807d72] dark:text-[#a09c92]">
                  {forexModal.pair.symbol} {forexModal.pair.quote_currency === 'ZTH' ? '(Global Reserve)' : 'Exchange'}
                </div>
              </div>
              <button
                onClick={() => setForexModal({ open: false, pair: null })}
                className="text-[#807d72] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#f7f7f4] p-1 rounded-[6px] hover:bg-[#efeee8] dark:hover:bg-[#26251e] transition-colors"
                aria-label="Close Modal"
              >
                <X size={18} />
              </button>
            </div>

            <div className="space-y-3 text-[13px]">
              <div className="flex gap-2">
                <button
                  onClick={() => setSwapSide('buy')}
                  className={`flex-1 py-2 rounded-[8px] font-medium text-[13px] transition-colors ${
                    swapSide === 'buy'
                      ? 'bg-[#1f8a65] text-white'
                      : 'bg-[#fafaf7] dark:bg-[#26251e] text-[#5a5852] dark:text-[#a09c92] border border-[#e6e5e0] dark:border-[#383730]'
                  }`}
                >
                  Buy {forexModal.pair.quote_currency}
                </button>
                <button
                  onClick={() => setSwapSide('sell')}
                  className={`flex-1 py-2 rounded-[8px] font-medium text-[13px] transition-colors ${
                    swapSide === 'sell'
                      ? 'bg-[#cf2d56] text-white'
                      : 'bg-[#fafaf7] dark:bg-[#26251e] text-[#5a5852] dark:text-[#a09c92] border border-[#e6e5e0] dark:border-[#383730]'
                  }`}
                >
                  Sell {forexModal.pair.quote_currency}
                </button>
              </div>

              <div className="bg-[#fafaf7] dark:bg-[#201f1b] p-3 rounded-[8px] space-y-1 font-mono text-[12px] border border-[#efeee8] dark:border-[#2c2b26]">
                <div className="flex justify-between">
                  <span className="text-[#807d72] dark:text-[#a09c92]">Spot Rate:</span>
                  <span className="font-semibold text-[#26251e] dark:text-[#f7f7f4]">
                    1 CRN = {forexModal.pair.rate.toFixed(4)} {forexModal.pair.quote_currency}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-[#807d72] dark:text-[#a09c92]">24h Change:</span>
                  <span className={forexModal.pair.change_24h >= 0 ? 'text-[#1f8a65]' : 'text-[#cf2d56]'}>
                    {forexModal.pair.change_24h >= 0 ? `+${forexModal.pair.change_24h.toFixed(2)}%` : `${forexModal.pair.change_24h.toFixed(2)}%`}
                  </span>
                </div>
              </div>

              <div>
                <label className="block text-[12px] font-medium text-[#26251e] dark:text-[#f7f7f4] mb-1">
                  Amount in {forexModal.pair.base_currency} (CRN)
                </label>
                <input
                  type="number"
                  value={swapAmount}
                  onChange={(e) => setSwapAmount(e.target.value)}
                  className="w-full bg-[#fafaf7] dark:bg-[#26251e] border border-[#e6e5e0] dark:border-[#383730] focus:border-[#26251e] dark:focus:border-[#f7f7f4] h-10 px-3 rounded-[8px] text-[14px] font-mono text-[#26251e] dark:text-[#f7f7f4] outline-none"
                />
                <div className="text-[11px] text-[#807d72] dark:text-[#a09c92] mt-1 font-mono">
                  Approx. Converted: {((parseFloat(swapAmount) || 0) * (swapSide === 'buy' ? forexModal.pair.rate : 1 / forexModal.pair.rate)).toFixed(2)} {swapSide === 'buy' ? forexModal.pair.quote_currency : forexModal.pair.base_currency}
                </div>
              </div>

              {swapFeedback && (
                <div className="p-2.5 rounded-[6px] bg-[#e8f5e9] dark:bg-[#153820] text-[12px] font-mono text-[#2e7d32] dark:text-[#4ade80] border border-[#c8e6c9] dark:border-[#166534]">
                  {swapFeedback}
                </div>
              )}
            </div>

            <div className="flex items-center justify-end gap-2 pt-2 border-t border-[#efeee8] dark:border-[#2c2b26]">
              <button
                onClick={() => setForexModal({ open: false, pair: null })}
                className="px-4 py-2 rounded-[8px] text-[13px] font-medium text-[#5a5852] dark:text-[#a09c92] hover:bg-[#fafaf7] dark:hover:bg-[#26251e]"
              >
                Cancel
              </button>
              <button
                onClick={handleExecuteSwap}
                disabled={swapBusy}
                className="px-4 py-2 rounded-[8px] text-[13px] font-medium text-white bg-[#26251e] dark:bg-[#f7f7f4] dark:text-[#121210] hover:bg-[#3d3b32] dark:hover:bg-[#eae8e0] disabled:opacity-50"
              >
                {swapBusy ? 'Converting...' : 'Confirm Swap'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
