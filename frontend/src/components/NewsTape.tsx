import React, { useState, useMemo } from 'react';
import { EditorialNewsDTO, TradeRecord } from '../types/api';
import { Newspaper, Search, Flame, TrendingUp, TrendingDown, Minus, Clock, Tag } from 'lucide-react';

interface NewsTapeProps {
  events: EditorialNewsDTO[];
  trades: TradeRecord[];
}

export const NewsTape: React.FC<NewsTapeProps> = ({ events, trades }) => {
  const [selectedCategory, setSelectedCategory] = useState<string>('ALL');
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [urgencyFilter, setUrgencyFilter] = useState<string>('ALL');

  const categories = useMemo(() => {
    const cats = new Set<string>();
    cats.add('ALL');
    events.forEach((ev) => {
      if (ev.category) cats.add(ev.category.toUpperCase());
    });
    return Array.from(cats);
  }, [events]);

  const filteredEvents = useMemo(() => {
    return events.filter((ev) => {
      if (selectedCategory !== 'ALL' && ev.category?.toUpperCase() !== selectedCategory) {
        return false;
      }
      if (urgencyFilter !== 'ALL' && ev.urgency !== urgencyFilter) {
        return false;
      }
      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase();
        const inHeadline = ev.headline?.toLowerCase().includes(q);
        const inSummary = ev.summary?.toLowerCase().includes(q);
        const inWire = ev.wire?.toLowerCase().includes(q);
        const inTickers = ev.affected_tickers?.some((t) => t.toLowerCase().includes(q));
        if (!inHeadline && !inSummary && !inWire && !inTickers) {
          return false;
        }
      }
      return true;
    });
  }, [events, selectedCategory, urgencyFilter, searchQuery]);

  const getUrgencyBadge = (urgency: string) => {
    switch (urgency) {
      case 'BREAKING':
        return 'bg-[#cf2d56] text-white font-bold animate-pulse';
      case 'ALERT':
        return 'bg-[#dfa88f] text-[#26251e] font-semibold';
      case 'DEVELOPING':
        return 'bg-[#c0a8dd] text-[#26251e] font-semibold';
      default:
        return 'bg-[#e6e5e0] text-[#55534c]';
    }
  };

  const getSentimentPill = (sentiment: string) => {
    switch (sentiment) {
      case 'BULLISH':
        return {
          icon: <TrendingUp className="w-3 h-3 text-[#1f8a65]" />,
          text: 'BULLISH',
          color: 'text-[#1f8a65] bg-[#9fc9a2]/20 border-[#9fc9a2]/50',
        };
      case 'BEARISH':
        return {
          icon: <TrendingDown className="w-3 h-3 text-[#cf2d56]" />,
          text: 'BEARISH',
          color: 'text-[#cf2d56] bg-[#dfa88f]/20 border-[#dfa88f]/50',
        };
      default:
        return {
          icon: <Minus className="w-3 h-3 text-[#807d72]" />,
          text: 'NEUTRAL',
          color: 'text-[#807d72] bg-[#efeee8] border-[#e6e5e0]',
        };
    }
  };

  return (
    <div className="space-y-6">
      {/* Header Banner */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <div className="text-[11px] font-mono uppercase tracking-[0.08em] text-[#807d72] mb-1 flex items-center gap-1.5">
            <Newspaper className="w-3.5 h-3.5 text-[#c08532]" />
            Editorial Wire Service · Institutional Macro & Corporate Dispatches
          </div>
          <h1 className="text-[26px] font-normal tracking-[-0.025em] text-[#26251e] leading-[1.25]">
            Global Financial News Wire
          </h1>
        </div>

        {/* Search & Urgency Quick Select */}
        <div className="flex flex-wrap items-center gap-2">
          <div className="relative">
            <Search className="w-3.5 h-3.5 absolute left-2.5 top-1/2 -translate-y-1/2 text-[#807d72]" />
            <input
              type="text"
              placeholder="Filter headlines, wires, tickers..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-8 pr-3 py-1.5 text-[12px] bg-white border border-[#e6e5e0] rounded-[6px] text-[#26251e] placeholder-[#a09c92] focus:outline-none focus:border-[#26251e] w-[220px]"
            />
          </div>

          <div className="flex items-center bg-[#efeee8] p-0.5 rounded-[8px] border border-[#e6e5e0]">
            {(['ALL', 'BREAKING', 'ALERT'] as const).map((u) => (
              <button
                key={u}
                onClick={() => setUrgencyFilter(u)}
                className={`px-2.5 py-1 text-[11px] font-mono rounded-[6px] transition-colors ${
                  urgencyFilter === u
                    ? 'bg-white text-[#26251e] font-semibold'
                    : 'text-[#807d72] hover:text-[#26251e]'
                }`}
              >
                {u === 'BREAKING' && <Flame className="w-3 h-3 inline mr-1 text-[#cf2d56]" />}
                {u}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Category Pills Bar */}
      <div className="flex items-center gap-2 overflow-x-auto pb-1 border-b border-[#e6e5e0]">
        {categories.map((cat) => (
          <button
            key={cat}
            onClick={() => setSelectedCategory(cat)}
            className={`px-3 py-1 text-[11px] font-mono whitespace-nowrap rounded-[6px] border transition-colors ${
              selectedCategory === cat
                ? 'bg-[#26251e] text-white border-[#26251e] font-medium'
                : 'bg-white text-[#807d72] border-[#e6e5e0] hover:text-[#26251e]'
            }`}
          >
            {cat}
          </button>
        ))}
      </div>

      {/* Main Grid: Editorial News Dispatches (8 cols) & Order Fills Tape (4 cols) */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* News Dispatches */}
        <div className="lg:col-span-8 bg-white border border-[#e6e5e0] rounded-[12px] p-5 shadow-[0_1px_3px_rgba(0,0,0,0.02)]">
          <div className="flex items-center justify-between pb-3 border-b border-[#efeee8] mb-4">
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-[#1f8a65] animate-ping" />
              <h2 className="text-[15px] font-semibold text-[#26251e] tracking-tight">Editorial Wire Dispatches</h2>
            </div>
            <span className="text-[11px] font-mono text-[#807d72]">
              Showing {filteredEvents.length} of {events.length} reports
            </span>
          </div>

          <div className="divide-y divide-[#efeee8] max-h-[700px] overflow-y-auto pr-2 space-y-4">
            {filteredEvents.map((item) => {
              const sent = getSentimentPill(item.sentiment);
              return (
                <div key={item.id || item.tick + item.timestamp} className="pt-4 first:pt-0 space-y-2 group">
                  {/* Top line metadata: Wire, Category, Urgency, Sentiment, Timestamp */}
                  <div className="flex flex-wrap items-center justify-between gap-2 text-[11px] font-mono">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="font-semibold text-[#26251e] bg-[#efeee8] px-2 py-0.5 rounded-[4px] tracking-wide">
                        {item.wire || 'REUTERS-WIRE'}
                      </span>
                      <span className="text-[#807d72] uppercase">{item.category || 'GENERAL'}</span>
                      <span
                        className={`text-[9px] uppercase tracking-wider px-1.5 py-0.5 rounded-[3px] ${getUrgencyBadge(
                          item.urgency
                        )}`}
                      >
                        {item.urgency}
                      </span>
                      <span
                        className={`inline-flex items-center gap-1 text-[10px] px-2 py-0.5 rounded-[4px] border ${sent.color}`}
                      >
                        {sent.icon}
                        {sent.text}
                      </span>
                    </div>

                    <div className="flex items-center gap-1.5 text-[#a09c92] text-[10px]">
                      <Clock className="w-3 h-3" />
                      <span>Tick #{item.tick}</span>
                      <span>·</span>
                      <span>{new Date(item.timestamp).toLocaleTimeString()}</span>
                    </div>
                  </div>

                  {/* Headline */}
                  <h3 className="text-[15px] font-semibold text-[#26251e] leading-snug group-hover:text-[#c08532] transition-colors">
                    {item.headline}
                  </h3>

                  {/* Summary / Body Paragraph (Clean prose, NEVER raw JSON) */}
                  <p className="text-[13px] text-[#55534c] leading-relaxed font-normal">
                    {item.summary}
                  </p>

                  {/* Affected Tickers Tag Footer */}
                  {item.affected_tickers && item.affected_tickers.length > 0 && (
                    <div className="flex items-center gap-1.5 pt-1">
                      <Tag className="w-3 h-3 text-[#807d72]" />
                      <span className="text-[11px] font-mono text-[#807d72] mr-1">TICKERS:</span>
                      <div className="flex flex-wrap gap-1">
                        {item.affected_tickers.map((t) => (
                          <span
                            key={t}
                            className="text-[11px] font-mono font-medium text-[#26251e] bg-[#f7f7f4] border border-[#e6e5e0] px-1.5 py-0.5 rounded-[4px] hover:bg-[#efeee8] transition-colors"
                          >
                            ${t}
                          </span>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              );
            })}

            {filteredEvents.length === 0 && (
              <div className="py-16 text-center text-[#807d72] text-[13px] space-y-2">
                <Newspaper className="w-8 h-8 mx-auto text-[#d0cec4]" />
                <div className="font-medium text-[#26251e]">No matching editorial dispatches</div>
                <div className="text-[12px]">Try clearing search keywords or selecting 'ALL' categories</div>
              </div>
            )}
          </div>
        </div>

        {/* High-frequency Order Executions Tape */}
        <div className="lg:col-span-4 bg-white border border-[#e6e5e0] rounded-[12px] p-5 shadow-[0_1px_3px_rgba(0,0,0,0.02)]">
          <div className="flex items-center justify-between pb-3 border-b border-[#efeee8] mb-3">
            <h2 className="text-[15px] font-semibold text-[#26251e] tracking-tight">Real-Time Tape Fills</h2>
            <span className="text-[11px] font-mono text-[#807d72]">Order Book Matches</span>
          </div>

          <div className="divide-y divide-[#efeee8] max-h-[700px] overflow-y-auto space-y-1">
            {trades.slice(0, 35).map((tr, i) => {
              const isBuy = tr.side?.toLowerCase() === 'buy';
              return (
                <div key={i} className="py-2 flex items-center justify-between text-[12px] font-mono">
                  <div>
                    <span className="font-semibold text-[#26251e]">{tr.symbol}</span>
                    <span
                      className={`ml-2 text-[10px] uppercase font-bold px-1.5 py-0.5 rounded-[3px] ${
                        isBuy
                          ? 'bg-[#9fc9a2]/30 text-[#1f8a65]'
                          : 'bg-[#dfa88f]/30 text-[#cf2d56]'
                      }`}
                    >
                      {tr.side}
                    </span>
                    <div className="text-[10px] text-[#a09c92] mt-0.5">
                      #{tr.taker_id?.slice(0, 8)} -&gt; #{tr.maker_id?.slice(0, 8)}
                    </div>
                  </div>

                  <div className="text-right">
                    <div className="text-[#26251e] font-semibold">${tr.price.toFixed(2)}</div>
                    <div className="text-[10px] text-[#807d72]">{tr.quantity.toLocaleString()} shs</div>
                  </div>
                </div>
              );
            })}

            {trades.length === 0 && (
              <div className="py-12 text-center text-[#807d72] text-[12px]">
                Awaiting market execution ticks...
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
