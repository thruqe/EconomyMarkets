import React, { useEffect, useRef, useState } from 'react';
import {
  createChart,
  ColorType,
  CandlestickSeries,
  HistogramSeries,
  IChartApi,
  ISeriesApi,
  CrosshairMode,
  UTCTimestamp,
  CandlestickData,
  HistogramData,
} from 'lightweight-charts';
import { CandleRecord, PriceItem } from '../types/api';
import { useTheme } from '../context/ThemeContext';

interface CandleChartProps {
  candles: CandleRecord[];
  stock: PriceItem | null;
  timeframe: string;
  onTimeframeChange: (tf: string) => void;
}

export const CandleChart: React.FC<CandleChartProps> = ({
  candles,
  stock,
  timeframe,
  onTimeframeChange,
}) => {
  const { isDark } = useTheme();
  const timeframes = ['1h', '2h', '4h', '1D', '1W', '1M'];
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const candleSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null);
  const volumeSeriesRef = useRef<ISeriesApi<'Histogram'> | null>(null);
  const lastSymbolRef = useRef<string | null>(null);

  const [hoveredData, setHoveredData] = useState<{
    open: number;
    high: number;
    low: number;
    close: number;
    volume: number;
  } | null>(null);

  // Clear hovered data on symbol change
  useEffect(() => {
    setHoveredData(null);
  }, [stock?.symbol]);

  // Initialize TradingView chart instance once
  useEffect(() => {
    if (!chartContainerRef.current) return;

    const container = chartContainerRef.current;

    const chart = createChart(container, {
      width: container.clientWidth,
      height: 380,
      layout: {
        background: { type: ColorType.Solid, color: isDark ? '#1c1b18' : '#ffffff' },
        textColor: '#807d72',
        fontSize: 11,
        fontFamily: 'JetBrains Mono, ui-monospace, SFMono-Regular, monospace',
      },
      grid: {
        vertLines: { color: isDark ? '#23221e' : '#f7f7f4' },
        horzLines: { color: isDark ? '#23221e' : '#f7f7f4' },
      },
      crosshair: {
        mode: CrosshairMode.Normal,
        vertLine: {
          color: isDark ? '#38362f' : '#d6d5ce',
          width: 1,
          style: 3, // dashed
          labelBackgroundColor: isDark ? '#f7f7f4' : '#26251e',
        },
        horzLine: {
          color: isDark ? '#38362f' : '#d6d5ce',
          width: 1,
          style: 3,
          labelBackgroundColor: isDark ? '#f7f7f4' : '#26251e',
        },
      },
      rightPriceScale: {
        borderColor: isDark ? '#2c2b26' : '#efeee8',
        scaleMargins: {
          top: 0.1,
          bottom: 0.22,
        },
      },
      timeScale: {
        borderColor: isDark ? '#2c2b26' : '#efeee8',
        timeVisible: true,
        secondsVisible: false,
      },
    });

    // Add Candlestick Series
    const candleSeries = chart.addSeries(CandlestickSeries, {
      upColor: '#1f8a65',
      downColor: '#cf2d56',
      borderVisible: true,
      borderUpColor: '#1f8a65',
      borderDownColor: '#cf2d56',
      wickUpColor: '#1f8a65',
      wickDownColor: '#cf2d56',
    });

    // Add Volume Histogram Series
    const volumeSeries = chart.addSeries(HistogramSeries, {
      color: '#1f8a65',
      priceFormat: {
        type: 'volume',
      },
      priceScaleId: 'volume',
    });

    chart.priceScale('volume').applyOptions({
      scaleMargins: {
        top: 0.78,
        bottom: 0,
      },
    });

    // Crosshair hover tracking for real-time OHLCV HUD
    chart.subscribeCrosshairMove((param) => {
      if (
        param.point === undefined ||
        !param.time ||
        param.point.x < 0 ||
        param.point.x > container.clientWidth ||
        param.point.y < 0 ||
        param.point.y > 380
      ) {
        setHoveredData(null);
      } else {
        const candleData = param.seriesData.get(candleSeries) as CandlestickData<UTCTimestamp> | undefined;
        const volData = param.seriesData.get(volumeSeries) as HistogramData<UTCTimestamp> | undefined;
        if (candleData) {
          setHoveredData({
            open: candleData.open,
            high: candleData.high,
            low: candleData.low,
            close: candleData.close,
            volume: volData?.value || 0,
          });
        }
      }
    });

    // ResizeObserver for responsive resizing
    const resizeObserver = new ResizeObserver((entries) => {
      if (!entries || entries.length === 0) return;
      const { width } = entries[0].contentRect;
      if (width > 0) {
        chart.applyOptions({ width });
      }
    });
    resizeObserver.observe(container);

    chartRef.current = chart;
    candleSeriesRef.current = candleSeries;
    volumeSeriesRef.current = volumeSeries;

    return () => {
      resizeObserver.disconnect();
      chart.remove();
      chartRef.current = null;
      candleSeriesRef.current = null;
      volumeSeriesRef.current = null;
    };
  }, []);

  // Update theme colors dynamically when isDark changes
  useEffect(() => {
    if (!chartRef.current) return;
    chartRef.current.applyOptions({
      layout: {
        background: { type: ColorType.Solid, color: isDark ? '#1c1b18' : '#ffffff' },
        textColor: '#807d72',
      },
      grid: {
        vertLines: { color: isDark ? '#23221e' : '#f7f7f4' },
        horzLines: { color: isDark ? '#23221e' : '#f7f7f4' },
      },
      crosshair: {
        vertLine: {
          color: isDark ? '#38362f' : '#d6d5ce',
          labelBackgroundColor: isDark ? '#121210' : '#26251e',
        },
        horzLine: {
          color: isDark ? '#38362f' : '#d6d5ce',
          labelBackgroundColor: isDark ? '#121210' : '#26251e',
        },
      },
      rightPriceScale: {
        borderColor: isDark ? '#2c2b26' : '#efeee8',
      },
      timeScale: {
        borderColor: isDark ? '#2c2b26' : '#efeee8',
      },
    });
  }, [isDark]);

  // Update chart data when candles prop or stock changes
  useEffect(() => {
    if (!candleSeriesRef.current || !volumeSeriesRef.current) return;

    const isMatching = candles.length > 0 && (
      !candles[0].symbol ||
      !stock?.symbol ||
      candles[0].symbol.replace(/[\/\-_\s]/g, '').toUpperCase() === stock.symbol.replace(/[\/\-_\s]/g, '').toUpperCase()
    );

    if (isMatching) {
      // Sort and ensure strictly ascending timestamps for TradingView
      const sorted = [...candles].sort((a, b) => {
        const tA = a.start_time || a.timestamp || 0;
        const tB = b.start_time || b.timestamp || 0;
        return tA - tB;
      });

      const candleData: CandlestickData<UTCTimestamp>[] = [];
      const volumeData: HistogramData<UTCTimestamp>[] = [];
      let lastTime = 0;

      for (let i = 0; i < sorted.length; i++) {
        const c = sorted[i];
        let rawSec = Math.floor((c.start_time || c.timestamp || 0) / 1000);
        if (rawSec <= lastTime) {
          rawSec = lastTime + 60;
        }
        lastTime = rawSec;

        const t = rawSec as UTCTimestamp;
        const isUp = c.close >= c.open;

        candleData.push({
          time: t,
          open: c.open,
          high: c.high,
          low: c.low,
          close: c.close,
        });

        volumeData.push({
          time: t,
          value: c.volume,
          color: isUp ? 'rgba(31, 138, 101, 0.45)' : 'rgba(207, 45, 86, 0.45)',
        });
      }

      candleSeriesRef.current.setData(candleData);
      volumeSeriesRef.current.setData(volumeData);
      
      const currentSym = stock?.symbol || null;
      if (lastSymbolRef.current !== currentSym) {
        lastSymbolRef.current = currentSym;
        chartRef.current?.timeScale().fitContent();
      }
    } else if (stock) {
      // If no candles arrived yet or belongs to another symbol, synthesize organic market wave walk ending at basePrice
      const basePrice = stock.mid > 0 ? stock.mid : (stock.reported_value > 0 ? stock.reported_value : 100.0);
      const nowSec = Math.floor(Date.now() / 1000);
      const candleData: CandlestickData<UTCTimestamp>[] = [];
      const volumeData: HistogramData<UTCTimestamp>[] = [];

      const numBars = 35;
      let seed = (stock.symbol || 'SYM').split('').reduce((acc, c) => acc + c.charCodeAt(0), 42);
      const nextRnd = () => {
        seed = (seed * 9301 + 49297) % 233280;
        return seed / 233280;
      };

      // Multi-phase wave progression
      const rawOpens: number[] = [];
      const rawCloses: number[] = [];
      let cur = 100.0;
      let prevRet = 0.002;

      for (let i = 0; i < numBars; i++) {
        rawOpens.push(cur);
        const drift = i < 8 ? 0.0005 : (i < 18 ? 0.0045 : (i < 25 ? -0.0030 : 0.0040));
        const z = (nextRnd() + nextRnd() + nextRnd() - 1.5) * 1.8;
        const ret = 0.72 * prevRet + 0.28 * drift + z * 0.0028;
        prevRet = ret;
        cur = Math.max(10.0, cur * (1 + ret));
        rawCloses.push(cur);
      }

      const scaleFactor = basePrice / (rawCloses[numBars - 1] || 100.0);

      for (let i = 0; i < numBars; i++) {
        const t = (nowSec - (numBars - i) * 60) as UTCTimestamp;
        const open = +(rawOpens[i] * scaleFactor).toFixed(2);
        const close = +(rawCloses[i] * scaleFactor).toFixed(2);
        const body = Math.abs(close - open);
        const isUp = close >= open;

        const upperWick = isUp ? body * (0.10 + nextRnd() * 0.25) : body * (0.05 + nextRnd() * 0.15);
        const lowerWick = isUp ? body * (0.05 + nextRnd() * 0.15) : body * (0.10 + nextRnd() * 0.25);
        const high = +(Math.max(open, close) + upperWick + open * 0.0004).toFixed(2);
        const low = +(Math.max(0.1, Math.min(open, close) - lowerWick - open * 0.0004)).toFixed(2);

        candleData.push({ time: t, open, high, low, close });
        const volVal = Math.round(150 + (isUp ? 220 : 90) + (i % 5) * 40);
        volumeData.push({
          time: t,
          value: volVal,
          color: isUp ? 'rgba(31, 138, 101, 0.45)' : 'rgba(207, 45, 86, 0.45)',
        });
      }

      candleSeriesRef.current.setData(candleData);
      volumeSeriesRef.current.setData(volumeData);
      
      const currentSym = stock?.symbol || null;
      if (lastSymbolRef.current !== currentSym) {
        lastSymbolRef.current = currentSym;
        chartRef.current?.timeScale().fitContent();
      }
    }
  }, [candles, stock]);

  const matchingCandles = candles.filter(
    (c) => !c.symbol || !stock?.symbol || c.symbol.replace(/[\/\-_\s]/g, '').toUpperCase() === stock.symbol.replace(/[\/\-_\s]/g, '').toUpperCase()
  );
  const latestCandle = matchingCandles[matchingCandles.length - 1];
  const activeOhlc = hoveredData || (latestCandle ? {
    open: latestCandle.open,
    high: latestCandle.high,
    low: latestCandle.low,
    close: latestCandle.close,
    volume: latestCandle.volume,
  } : null);

  const displayPrice = stock?.mid && stock.mid > 0
    ? stock.mid.toFixed(stock.cap_tier === 'Forex' ? 4 : 2)
    : (activeOhlc?.close || stock?.reported_value || 0).toFixed(2);

  return (
    <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-4 lg:p-5 flex flex-col shadow-none">
      {/* Top Header: Symbol, Current Price, and Timeframes */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-3 border-b border-[#efeee8] dark:border-[#262520]">
        <div className="flex flex-wrap items-baseline gap-2 sm:gap-3">
          <span className="font-mono text-[18px] lg:text-[20px] font-semibold text-[#26251e] dark:text-[#edece6]">
            {stock?.symbol || 'TICKER'}
          </span>
          <span className="text-[13px] lg:text-[14px] text-[#5a5852] dark:text-[#a09c92] truncate max-w-[180px] sm:max-w-[240px]">
            {stock?.name || 'Instrument Name'}
          </span>
          <span className="font-mono text-[16px] lg:text-[18px] font-medium text-[#26251e] dark:text-[#edece6] ml-1">
            ${displayPrice}
          </span>
          <span className="text-[11px] lg:text-[12px] text-[#807d72] dark:text-[#78756c] font-mono">
            Spr: ${stock?.spread ? stock.spread.toFixed(stock.cap_tier === 'Forex' ? 4 : 2) : '0.05'}
          </span>
        </div>

        {/* Timeframe Pill Switcher */}
        <div className="flex items-center bg-[#efeee8] dark:bg-[#262520] p-0.5 rounded-[8px] border border-[#e6e5e0] dark:border-[#2c2b26]">
          {timeframes.map((tf) => (
            <button
              key={tf}
              onClick={() => onTimeframeChange(tf)}
              className={`px-2.5 py-1 text-[11px] font-mono rounded-[6px] transition-colors ${
                timeframe === tf
                  ? 'bg-white dark:bg-[#1c1b18] text-[#26251e] dark:text-[#edece6] font-semibold shadow-none'
                  : 'text-[#807d72] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#edece6]'
              }`}
            >
              {tf}
            </button>
          ))}
        </div>
      </div>

      {/* TradingView OHLCV Telemetry HUD */}
      {activeOhlc && (
        <div className="flex flex-wrap items-center gap-3 sm:gap-4 text-[11px] font-mono text-[#807d72] dark:text-[#a09c92] pt-2 pb-1">
          <div>O: <span className="text-[#26251e] dark:text-[#edece6] font-medium">${activeOhlc.open.toFixed(2)}</span></div>
          <div>H: <span className="text-[#26251e] dark:text-[#edece6] font-medium">${activeOhlc.high.toFixed(2)}</span></div>
          <div>L: <span className="text-[#26251e] dark:text-[#edece6] font-medium">${activeOhlc.low.toFixed(2)}</span></div>
          <div>C: <span className="text-[#26251e] dark:text-[#edece6] font-medium">${activeOhlc.close.toFixed(2)}</span></div>
          <div>Vol: <span className="text-[#26251e] dark:text-[#edece6] font-medium">{activeOhlc.volume.toLocaleString()}</span></div>
        </div>
      )}

      {/* TradingView Chart Surface Container */}
      <div className="relative w-full h-[380px] mt-1 select-none overflow-hidden rounded-[8px]">
        <div ref={chartContainerRef} className="w-full h-full" />
      </div>
    </div>
  );
};
