import React, { useState, useEffect, useMemo } from 'react';
import { api } from './services/api';
import {
  CandleRecord,
  EditorialNewsDTO,
  MacroDTO,
  NationDTO,
  OrderBookDTO,
  PortfolioDTO,
  PriceItem,
  SimStatusDTO,
  TradeRecord,
  WorldDTO,
} from './types/api';
import { Navbar, ActiveTab } from './components/Navbar';
import { MarketOverview } from './components/MarketOverview';
import { StockUniverse } from './components/StockUniverse';
import { TerminalView } from './components/TerminalView';
import { NationBuilder } from './components/NationBuilder';
import { MacroDashboard } from './components/MacroDashboard';
import { PolicyStudio } from './components/PolicyStudio';
import { NewsTape } from './components/NewsTape';

import { socketService } from './services/websocket';
import { worldSocketService } from './services/worldSocket';


export const App: React.FC = () => {
  const [activeTab, setActiveTabState] = useState<ActiveTab>(() => {
    return (localStorage.getItem('em_active_tab') as ActiveTab) || 'overview';
  });
  const setActiveTab = (tab: ActiveTab) => {
    setActiveTabState(tab);
    try { localStorage.setItem('em_active_tab', tab); } catch (_) {}
  };

  const [status, setStatus] = useState<SimStatusDTO | null>(null);
  const [prices, setPrices] = useState<PriceItem[]>([]);
  const [selectedSymbol, setSelectedSymbolState] = useState<string>(() => {
    return localStorage.getItem('em_selected_symbol') || '';
  });
  const setSelectedSymbol = (sym: string | ((prev: string) => string)) => {
    setSelectedSymbolState((prev) => {
      const next = typeof sym === 'function' ? sym(prev) : sym;
      try { if (next) localStorage.setItem('em_selected_symbol', next); } catch (_) {}
      return next;
    });
  };

  const [candles, setCandles] = useState<CandleRecord[]>([]);
  const [orderBook, setOrderBook] = useState<OrderBookDTO | null>(null);
  const [portfolio, setPortfolio] = useState<PortfolioDTO | null>(null);
  const [nation, setNation] = useState<NationDTO | null>(null);
  const [world, setWorld] = useState<WorldDTO | null>(null);
  const [macro, setMacro] = useState<MacroDTO | null>(null);
  const [trades, setTrades] = useState<TradeRecord[]>([]);
  const [events, setEvents] = useState<EditorialNewsDTO[]>([]);

  const [speed, setSpeedState] = useState<number>(() => {
    const saved = localStorage.getItem('em_sim_speed');
    return saved ? parseFloat(saved) : 1.0;
  });
  const setSpeed = (s: number) => {
    setSpeedState(s);
    try { localStorage.setItem('em_sim_speed', s.toString()); } catch (_) {}
  };

  const [timeframe, setTimeframeState] = useState<string>(() => {
    return localStorage.getItem('em_chart_tf') || '1D';
  });
  const setTimeframe = (tf: string) => {
    setTimeframeState(tf);
    try { localStorage.setItem('em_chart_tf', tf); } catch (_) {}
  };

  const [tradesThisTick, setTradesThisTick] = useState<number>(0);
  const [calendarFormatted, setCalendarFormatted] = useState<string>('');
  const [backendOnline, setBackendOnline] = useState<boolean | null>(null);

  // Sync initial persisted speed to backend
  useEffect(() => {
    const saved = localStorage.getItem('em_sim_speed');
    if (saved) {
      const s = parseFloat(saved);
      if (s > 0) {
        api.setSimSpeed(s).catch(() => {});
      }
    }
  }, []);

  // WebSocket Live Streaming & Lifecycle
  useEffect(() => {
    socketService.connect();
    worldSocketService.connect();

    const unsubscribe = socketService.addListener((msg) => {
      setBackendOnline(true);
      if (msg.type === 'snapshot') {
        setStatus(msg.status);
        setPrices(msg.prices);
        setMacro(msg.macro);
        setNation(msg.nation);
        setPortfolio(msg.portfolio);
        if (msg.macro?.sim_date_formatted) {
          setCalendarFormatted(msg.macro.sim_date_formatted);
        }
        setSelectedSymbol((prev) => (prev ? prev : (msg.prices[0]?.symbol || '')));
      } else if (msg.type === 'tick') {
        setStatus((prev) =>
          prev
            ? {
                ...prev,
                tick: msg.tick,
                total_trades: msg.trades_count,
                total_volume: msg.volume,
                speed_multiplier: msg.speed,
              }
            : null
        );
        setPrices(msg.prices);
        if (msg.speed !== undefined) {
          setSpeedState(msg.speed);
        }
        if (msg.trades_this_tick !== undefined) setTradesThisTick(msg.trades_this_tick);
        if (msg.calendar_formatted) {
          setCalendarFormatted(msg.calendar_formatted);
          setMacro((prev) => prev ? { ...prev, sim_date_formatted: msg.calendar_formatted! } : prev);
        }
      } else if (msg.type === 'macro_update') {
        setMacro(msg.macro);
        setNation(msg.nation);
        setPortfolio(msg.portfolio);
        if (msg.macro?.sim_date_formatted) {
          setCalendarFormatted(msg.macro.sim_date_formatted);
        }
      } else if (msg.type === 'book') {
        setOrderBook(msg);
      }
    });

    const unsubscribeWorld = worldSocketService.addListener((msg) => {
      if (msg.type === 'world_snapshot') {
        setWorld({
          countries: msg.countries,
          forex_pairs: msg.forex_pairs,
          world_stocks: msg.world_stocks,
          country_indices: msg.country_indices,
          loan_requests: [],
          policy_log: [],
        });
      } else if (msg.type === 'world_tick') {
        setWorld((prev) =>
          prev
            ? { ...prev, forex_pairs: msg.forex_pairs, world_stocks: msg.world_stocks, country_indices: msg.country_indices }
            : prev
        );
      }
    });

    // Initial background data fetch for static/macro world
    api.getWorld().then(setWorld).catch(() => {});
    api.getEvents(50).then(setEvents).catch(() => {});

    // Low frequency background refresher for events only
    const bgTimer = setInterval(() => {
      api.getEvents(50).then(setEvents).catch(() => {});
    }, 6000);

    return () => {
      unsubscribe();
      unsubscribeWorld();
      clearInterval(bgTimer);
      worldSocketService.disconnect();
    };
  }, []);


  // Subscribe to live Order Book streaming whenever symbol changes
  useEffect(() => {
    if (selectedSymbol) {
      setCandles([]);
      setTrades([]);
      setOrderBook(null);
      socketService.subscribeBook(selectedSymbol);
      api.getCandles(selectedSymbol, timeframe).then(setCandles).catch(() => {});
      api.getTrades(selectedSymbol, 30).then(setTrades).catch(() => {});

      // Refresh candles periodically
      const candleTimer = setInterval(() => {
        api.getCandles(selectedSymbol, timeframe).then(setCandles).catch(() => {});
        api.getTrades(selectedSymbol, 30).then(setTrades).catch(() => {});
      }, 3000);

      return () => {
        clearInterval(candleTimer);
      };
    }
  }, [selectedSymbol, timeframe]);

  const handleSpeedChange = async (newSpeed: number) => {
    setSpeed(newSpeed);
    try {
      await api.setSimSpeed(newSpeed);
    } catch (_) {}
  };

  const handleSelectStock = (sym: string) => {
    setSelectedSymbol(sym);
    setActiveTab('terminal');
  };

  const handleSubmitOrder = async (order: any) => {
    await api.submitOrder(order);
  };

  const handleClosePosition = async (sym: string) => {
    await api.closePosition(sym);
  };

  const handleTogglePolicy = async (id: string) => {
    await api.togglePolicy(id);
  };


  const handleInvestNation = async (pillar: string, amount: number) => {
    await api.investNation({ pillar, amount });
  };

  const handleCharterNation = async () => {
    await api.charterExchange();
  };

  const handleConfigureNation = async (data: any) => {
    await api.foundNation(data);
  };

  const handleDiplomacy = async (data: any) => {
    await api.sendDiplomacy(data);
  };

  const handleIssueBond = async (maturityYears: number, amount: number) => {
    await api.issueBond({ maturity_years: maturityYears, amount });
    const [updatedNat, updatedMacro] = await Promise.all([api.getNation(), api.getMacro()]);
    setNation(updatedNat);
    setMacro(updatedMacro);
  };

  const handleBuybackBond = async (maturityYears: number, amount: number) => {
    await api.buybackBond({ maturity_years: maturityYears, amount });
    const [updatedNat, updatedMacro] = await Promise.all([api.getNation(), api.getMacro()]);
    setNation(updatedNat);
    setMacro(updatedMacro);
  };

  const handleSwapForex = async (pair: string, amount: number, side: 'buy' | 'sell') => {
    await api.swapForex({ pair, amount, side });
    const updatedPort = await api.getPortfolio();
    setPortfolio(updatedPort);
  };

  const handleSetMaintenance = async (rate: number) => {
    await api.setMaintenance(rate);
    const updatedNat = await api.getNation();
    setNation(updatedNat);
  };

  const handleRespondLoan = async (requestId: string, action: 'accept' | 'decline') => {
    await api.respondLoan({ request_id: requestId, action });
    const [updatedWorld, updatedNat] = await Promise.all([api.getWorld(), api.getNation()]);
    setWorld(updatedWorld);
    setNation(updatedNat);
  };

  const handleRespondContract = async (contractId: string, action: 'accept' | 'decline') => {
    await api.respondContract({ contract_id: contractId, action });
    const [updatedWorld, updatedNat] = await Promise.all([api.getWorld(), api.getNation()]);
    setWorld(updatedWorld);
    setNation(updatedNat);
  };

  const handleUpdatePortfolioSettings = async (data: { balance?: number; leverage?: number }) => {
    await api.updatePortfolioSettings(data);
    const updatedPort = await api.getPortfolio();
    setPortfolio(updatedPort);
  };

  const handleRepayDebt = async (amount: number) => {
    await api.repayDebt(amount);
    const updatedNat = await api.getNation();
    setNation(updatedNat);
  };

  const handleTradeReserveCurrency = async (data: { side: 'buy' | 'sell'; amount: number }) => {
    await api.tradeReserveCurrency(data);
    const [updatedNat, updatedWorld] = await Promise.all([api.getNation(), api.getWorld()]);
    setNation(updatedNat);
    setWorld(updatedWorld);
  };

  // Resolve the active instrument into a unified PriceItem across all asset classes
  const selectedStock: PriceItem | null = useMemo(() => {
    if (!selectedSymbol) {
      return prices[0] || null;
    }

    const symUpper = selectedSymbol.trim().toUpperCase();
    const cleanSym = symUpper.replace(/[\/\-_\s]/g, '');

    // 1. Check domestic public & private prices
    const foundPrice = prices.find((p) => {
      const pSym = p.symbol.toUpperCase();
      return pSym === symUpper || pSym.replace(/[\/\-_\s]/g, '') === cleanSym;
    });
    if (foundPrice) return foundPrice;

    // 2. Check sovereign benchmark indices
    const foundIndex = world?.country_indices?.find((i) => {
      const iSym = i.symbol.toUpperCase();
      return iSym === symUpper || iSym.replace(/[\/\-_\s]/g, '') === cleanSym;
    });
    if (foundIndex) {
      return {
        symbol: foundIndex.symbol,
        name: foundIndex.name,
        sector: `${foundIndex.country_id} Sovereign Benchmark`,
        cap_tier: 'Index',
        true_value: foundIndex.price,
        reported_value: foundIndex.price,
        mid: foundIndex.price,
        bid: +(foundIndex.price * 0.9995).toFixed(2),
        ask: +(foundIndex.price * 1.0005).toFixed(2),
        spread: +(foundIndex.price * 0.001).toFixed(2),
        annual_revenue: 0,
        net_margin: 0,
        sector_multiple: 0,
        market_cap: 0,
        reported_market_cap: 0,
        shares_outstanding: 0,
        float_shares: 0,
        pe_ratio: 0,
        ps_ratio: 0,
        is_ipo: false,
        ipo_tick: 0,
        ipo_price: 0,
        is_private: false,
        stage: 'Benchmark Index',
        private_valuation: 0,
      };
    }

    // 3. Check international world stocks
    const foundWorldStock = world?.world_stocks?.find((w) => {
      const wTicker = w.ticker.toUpperCase();
      return wTicker === symUpper || wTicker.replace(/[\/\-_\s]/g, '') === cleanSym;
    });
    if (foundWorldStock) {
      return {
        symbol: foundWorldStock.ticker,
        name: foundWorldStock.name,
        sector: `${foundWorldStock.country_name} · ${foundWorldStock.sector}`,
        cap_tier: 'World Equities',
        true_value: foundWorldStock.price,
        reported_value: foundWorldStock.price,
        mid: foundWorldStock.price,
        bid: +(foundWorldStock.price * 0.999).toFixed(2),
        ask: +(foundWorldStock.price * 1.001).toFixed(2),
        spread: +(foundWorldStock.price * 0.002).toFixed(2),
        annual_revenue: 0,
        net_margin: 0,
        sector_multiple: 0,
        market_cap: foundWorldStock.market_cap,
        reported_market_cap: foundWorldStock.market_cap,
        shares_outstanding: 0,
        float_shares: 0,
        pe_ratio: foundWorldStock.pe_ratio,
        ps_ratio: 0,
        is_ipo: false,
        ipo_tick: 0,
        ipo_price: 0,
        is_private: false,
        stage: 'International',
        private_valuation: 0,
      };
    }

    // 4. Check forex pairs
    const foundForex = world?.forex_pairs?.find((f) => {
      const fSym = f.symbol.toUpperCase();
      return fSym === symUpper || fSym.replace(/[\/\-_\s]/g, '') === cleanSym;
    });
    if (foundForex) {
      return {
        symbol: foundForex.symbol,
        name: `${foundForex.base_currency}/${foundForex.quote_currency} Foreign Exchange Rate`,
        sector: 'Currency FX',
        cap_tier: 'Forex',
        true_value: foundForex.rate,
        reported_value: foundForex.rate,
        mid: foundForex.rate,
        bid: +(foundForex.rate * 0.9998).toFixed(4),
        ask: +(foundForex.rate * 1.0002).toFixed(4),
        spread: +(foundForex.rate * 0.0004).toFixed(4),
        annual_revenue: 0,
        net_margin: 0,
        sector_multiple: 0,
        market_cap: 0,
        reported_market_cap: 0,
        shares_outstanding: 0,
        float_shares: 0,
        pe_ratio: 0,
        ps_ratio: 0,
        is_ipo: false,
        ipo_tick: 0,
        ipo_price: 0,
        is_private: false,
        stage: 'Forex Pair',
        private_valuation: 0,
      };
    }

    // Explicit fallback for selected symbol so UI never gets stuck on PR (prices[0])
    return {
      symbol: selectedSymbol,
      name: selectedSymbol,
      sector: 'Market Asset',
      cap_tier: 'Equities',
      true_value: 100.0,
      reported_value: 100.0,
      mid: 100.0,
      bid: 99.95,
      ask: 100.05,
      spread: 0.1,
      annual_revenue: 0,
      net_margin: 0,
      sector_multiple: 0,
      market_cap: 0,
      reported_market_cap: 0,
      shares_outstanding: 0,
      float_shares: 0,
      pe_ratio: 0,
      ps_ratio: 0,
      is_ipo: false,
      ipo_tick: 0,
      ipo_price: 0,
      is_private: false,
      stage: 'Listed',
      private_valuation: 0,
    };
  }, [selectedSymbol, prices, world]);

  return (
    <div className="min-h-screen bg-[#f7f7f4] dark:bg-[#121210] text-[#26251e] dark:text-[#edece6] flex flex-col selection:bg-[#f54e00]/20 selection:text-[#f54e00]">
      {/* Top Navigation */}
      <Navbar
        status={status}
        activeTab={activeTab}
        onTabChange={setActiveTab}
        speed={speed}
        onSpeedChange={handleSpeedChange}
        tradesThisTick={tradesThisTick}
        calendarFormatted={calendarFormatted || macro?.sim_date_formatted || nation?.sim_date_formatted}
      />

      {/* Connectivity status banner */}
      {backendOnline === false && (
        <div className="bg-[#cf2d56] text-white py-2 px-4 text-center text-[12px] font-mono">
          Connecting to simulation service...
        </div>
      )}

      {/* Main Content Floor */}
      <main className="flex-1 max-w-[1360px] w-full mx-auto p-4 md:p-8">
        {activeTab === 'overview' && (
          <MarketOverview
            prices={prices}
            macro={macro}
            onSelectStock={handleSelectStock}
          />
        )}

        {activeTab === 'terminal' && (
          <TerminalView
            selectedStock={selectedStock}
            candles={candles}
            orderBook={orderBook}
            portfolio={portfolio}
            timeframe={timeframe}
            onTimeframeChange={setTimeframe}
            onSubmitOrder={handleSubmitOrder}
            onClosePosition={handleClosePosition}
            onSelectStock={handleSelectStock}
            allPrices={prices}
            indices={world?.country_indices || []}
            worldStocks={world?.world_stocks || []}
            forexPairs={world?.forex_pairs || []}
            onUpdateSettings={handleUpdatePortfolioSettings}
          />
        )}

        {activeTab === 'universe' && (
          <StockUniverse
            prices={prices}
            bonds={nation?.bonds || macro?.bonds || []}
            forexPairs={world?.forex_pairs || []}
            worldStocks={world?.world_stocks || []}
            countryIndices={world?.country_indices || []}
            onSelectStock={handleSelectStock}
            onIssueBond={handleIssueBond}
            onBuybackBond={handleBuybackBond}
            onSwapForex={handleSwapForex}
          />
        )}

        {activeTab === 'nation' && (
          <NationBuilder
            nation={nation}
            world={world}
            onInvest={handleInvestNation}
            onCharter={handleCharterNation}
            onConfigure={handleConfigureNation}
            onDiplomacy={handleDiplomacy}
            onSetMaintenance={handleSetMaintenance}
            onRespondLoan={handleRespondLoan}
            onRespondContract={handleRespondContract}
            onRepayDebt={handleRepayDebt}
            onTradeReserveCurrency={handleTradeReserveCurrency}
          />
        )}

        {activeTab === 'macro' && (
          <MacroDashboard macro={macro} />
        )}

        {activeTab === 'policy' && (
          <PolicyStudio
            policies={macro?.policies || []}
            macro={macro}
            onTogglePolicy={handleTogglePolicy}
          />
        )}

        {activeTab === 'tape' && (
          <NewsTape
            events={events}
            trades={trades}
          />
        )}
      </main>
    </div>
  );
};

export default App;
