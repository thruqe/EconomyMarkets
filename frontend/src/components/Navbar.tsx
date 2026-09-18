import React, { useState, useRef, useEffect } from 'react';
import {
  Activity,
  Globe,
  BarChart2,
  BookOpen,
  Layers,
  Newspaper,
  Landmark,
  Menu,
  X,
  ChevronDown,
  Check,
  Sun,
  Moon,
  Laptop,
} from 'lucide-react';
import { SimStatusDTO } from '../types/api';
import { useTheme } from '../context/ThemeContext';

export type ActiveTab = 'overview' | 'terminal' | 'universe' | 'nation' | 'macro' | 'policy' | 'tape';

interface NavbarProps {
  status: SimStatusDTO | null;
  activeTab: ActiveTab;
  onTabChange: (tab: ActiveTab) => void;
  speed: number;
  onSpeedChange: (speed: number) => void;
  tradesThisTick?: number;
  calendarFormatted?: string;
}

export const Navbar: React.FC<NavbarProps> = ({
  status,
  activeTab,
  onTabChange,
  speed,
  onSpeedChange,
  tradesThisTick,
  calendarFormatted,
}) => {
  const { theme, isDark, toggleTheme } = useTheme();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [moreDropdownOpen, setMoreDropdownOpen] = useState(false);
  const moreRef = useRef<HTMLDivElement>(null);

  const primaryTabs = [
    { id: 'overview', label: 'Overview', icon: BarChart2 },
    { id: 'terminal', label: 'Terminal', icon: Activity },
    { id: 'universe', label: 'Universe', icon: Layers },
    { id: 'nation', label: 'Nation', icon: Landmark },
  ] as const;

  const secondaryTabs = [
    { id: 'macro', label: 'Macro', shortLabel: 'Macro', icon: Globe },
    { id: 'policy', label: 'Policy Studio', shortLabel: 'Policy', icon: BookOpen },
    { id: 'tape', label: 'News & Events', shortLabel: 'News', icon: Newspaper },
  ] as const;

  const allTabs = [...primaryTabs, ...secondaryTabs];
  const speeds = [0.5, 1.0, 2.0, 5.0, 10.0];

  // Close dropdown on click outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (moreRef.current && !moreRef.current.contains(event.target as Node)) {
        setMoreDropdownOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const formatCurrency = (val?: number) => {
    if (val === undefined) return '$0.00';
    if (val >= 1e12) return `$${(val / 1e12).toFixed(2)}T`;
    if (val >= 1e9) return `$${(val / 1e9).toFixed(2)}B`;
    if (val >= 1e6) return `$${(val / 1e6).toFixed(2)}M`;
    return `$${val.toLocaleString(undefined, { maximumFractionDigits: 2 })}`;
  };

  const handleTabClick = (tabId: ActiveTab) => {
    onTabChange(tabId);
    setMobileOpen(false);
    setMoreDropdownOpen(false);
  };

  const isSecondaryActive = secondaryTabs.some((t) => t.id === activeTab);
  const activeSecondaryTab = secondaryTabs.find((t) => t.id === activeTab);
  const ActiveSecondaryIcon = activeSecondaryTab ? activeSecondaryTab.icon : null;

  return (
    <header className="sticky top-0 z-40 bg-[#f7f7f4]/95 dark:bg-[#121210]/95 backdrop-blur-md border-b border-[#e6e5e0] dark:border-[#2c2b26] px-3 sm:px-4 lg:px-6 h-16 flex items-center justify-between gap-2 lg:gap-4 select-none">
      {/* Left: Brand & Responsive Navigation Bar */}
      <div className="flex items-center gap-2 sm:gap-3 lg:gap-4 flex-shrink-0">
        <div
          className="flex items-center gap-2 cursor-pointer flex-shrink-0 group"
          onClick={() => handleTabClick('overview')}
        >
          {/* Financial Charts Brand SVG Icon */}
          <div className="w-5 h-5 rounded-[5px] flex items-center justify-center bg-[#181714] border border-[#2c2b26] p-0.5 shadow-none group-hover:border-[#f54e00]/60 transition-colors">
            <svg viewBox="0 0 32 32" className="w-full h-full" fill="none">
              {/* Candlestick 1: Up green */}
              <line x1="8.5" y1="12" x2="8.5" y2="24" stroke="#1f8a65" strokeWidth="2" strokeLinecap="round"/>
              <rect x="7" y="15" width="3" height="6" rx="0.75" fill="#1f8a65"/>

              {/* Candlestick 2: Pullback */}
              <line x1="16" y1="10" x2="16" y2="22" stroke="#cf2d56" strokeWidth="2" strokeLinecap="round"/>
              <rect x="14.5" y="13" width="3" height="5" rx="0.75" fill="#cf2d56"/>

              {/* Candlestick 3: Signature orange breakout */}
              <line x1="23.5" y1="6" x2="23.5" y2="20" stroke="#f54e00" strokeWidth="2" strokeLinecap="round"/>
              <rect x="22" y="8" width="3" height="8" rx="0.75" fill="#f54e00"/>

              {/* Trend Vector */}
              <path d="M5 21 C10 19, 13 14, 18 13 C21 12.5, 24 8, 27 6" stroke="#f54e00" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round"/>

              {/* Peak indicator dot */}
              <circle cx="27" cy="6" r="2" fill="#ffffff"/>
            </svg>
          </div>
          <span className="font-medium text-[15px] sm:text-[16px] lg:text-[17px] tracking-[-0.03em] text-[#26251e] dark:text-[#edece6] whitespace-nowrap">
            Economy<span className="text-[#807d72] dark:text-[#a09c92]">Markets</span>
          </span>
        </div>

        {/* Desktop navigation for wide screens (2xl: 1536px+): All 7 tabs visible */}
        <nav className="hidden 2xl:flex items-center gap-1">
          {allTabs.map((tab) => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                onClick={() => handleTabClick(tab.id as ActiveTab)}
                className={`flex items-center gap-1.5 px-2.5 py-1.5 rounded-[8px] text-[13px] font-medium whitespace-nowrap transition-colors flex-shrink-0 ${
                  isActive
                    ? 'bg-white dark:bg-[#1c1b18] text-[#26251e] dark:text-[#edece6] border border-[#e6e5e0] dark:border-[#2c2b26] shadow-none'
                    : 'text-[#5a5852] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#edece6] hover:bg-[#efeee8] dark:hover:bg-[#262520]'
                }`}
              >
                <Icon size={14} className={isActive ? 'text-[#f54e00]' : 'text-[#807d72] dark:text-[#a09c92]'} />
                <span>{tab.label}</span>
              </button>
            );
          })}
        </nav>

        {/* Adaptive navigation for standard desktop screens (lg: 1024px to 1535px):
            Shows 4 core tabs + "More ▾" dropdown with consistent dropdown styling */}
        <nav className="hidden lg:flex 2xl:hidden items-center gap-1">
          {primaryTabs.map((tab) => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                onClick={() => handleTabClick(tab.id as ActiveTab)}
                className={`flex items-center gap-1.5 px-2.5 py-1.5 rounded-[8px] text-[12px] font-medium whitespace-nowrap transition-colors flex-shrink-0 ${
                  isActive
                    ? 'bg-white dark:bg-[#1c1b18] text-[#26251e] dark:text-[#edece6] border border-[#e6e5e0] dark:border-[#2c2b26]'
                    : 'text-[#5a5852] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#edece6] hover:bg-[#efeee8] dark:hover:bg-[#262520]'
                }`}
              >
                <Icon size={13} className={isActive ? 'text-[#f54e00]' : 'text-[#807d72] dark:text-[#a09c92]'} />
                <span>{tab.label}</span>
              </button>
            );
          })}

          {/* More ▾ Dropdown */}
          <div className="relative" ref={moreRef}>
            <button
              onClick={() => setMoreDropdownOpen(!moreDropdownOpen)}
              className={`flex items-center gap-1.5 px-2.5 py-1.5 rounded-[8px] text-[12px] font-medium whitespace-nowrap transition-colors ${
                isSecondaryActive
                  ? 'bg-white dark:bg-[#1c1b18] text-[#26251e] dark:text-[#edece6] border border-[#e6e5e0] dark:border-[#2c2b26]'
                  : 'text-[#5a5852] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#edece6] hover:bg-[#efeee8] dark:hover:bg-[#262520]'
              }`}
            >
              {ActiveSecondaryIcon && <ActiveSecondaryIcon size={13} className="text-[#f54e00]" />}
              <span>{isSecondaryActive && activeSecondaryTab ? activeSecondaryTab.shortLabel : 'More'}</span>
              <ChevronDown size={12} className={`text-[#807d72] dark:text-[#a09c92] transition-transform duration-150 ${moreDropdownOpen ? 'rotate-180' : ''}`} />
            </button>

            {moreDropdownOpen && (
              <div className="absolute top-full left-0 mt-1 w-48 bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[8px] py-1 z-50 shadow-none">
                {secondaryTabs.map((tab) => {
                  const Icon = tab.icon;
                  const isActive = activeTab === tab.id;
                  return (
                    <button
                      key={tab.id}
                      onClick={() => handleTabClick(tab.id as ActiveTab)}
                      className={`w-full flex items-center justify-between px-3 py-2 text-[13px] text-left transition-colors ${
                        isActive
                          ? 'bg-[#fafaf7] dark:bg-[#262520] text-[#f54e00] font-semibold'
                          : 'text-[#5a5852] dark:text-[#a09c92] hover:bg-[#efeee8] dark:hover:bg-[#262520] hover:text-[#26251e] dark:hover:text-[#edece6]'
                      }`}
                    >
                      <div className="flex items-center gap-2">
                        <Icon size={14} className={isActive ? 'text-[#f54e00]' : 'text-[#807d72] dark:text-[#a09c92]'} />
                        <span>{tab.label}</span>
                      </div>
                      {isActive && <Check size={13} className="text-[#f54e00] shrink-0 ml-2" />}
                    </button>
                  );
                })}
              </div>
            )}
          </div>
        </nav>
      </div>

      {/* Right: Telemetry, Speed Controls & Primary CTA */}
      <div className="flex items-center gap-1.5 sm:gap-2 lg:gap-3 flex-shrink-0">
        {status && (
          <div className="flex items-center gap-2 xl:gap-3 text-[11px] sm:text-[12px] font-mono border-r border-[#e6e5e0] dark:border-[#2c2b26] pr-2 sm:pr-3">
            <div className="flex items-center gap-1.5 whitespace-nowrap">
              <span className="w-2 h-2 rounded-full bg-[#1f8a65] animate-pulse flex-shrink-0" />
              <span className="text-[#807d72] dark:text-[#a09c92] hidden sm:inline">Date:</span>
              <span className="text-[#26251e] dark:text-[#edece6] font-semibold">
                {calendarFormatted || `Day ${status.tick}`}
              </span>
            </div>
            <div className="hidden 2xl:block whitespace-nowrap">
              <span className="text-[#807d72] dark:text-[#a09c92]">Trades: </span>
              <span className="text-[#26251e] dark:text-[#edece6]">{status.total_trades.toLocaleString()}</span>
              {tradesThisTick !== undefined && tradesThisTick > 0 && (
                <span className="text-[#807d72] dark:text-[#a09c92] ml-1">(+{tradesThisTick}/step)</span>
              )}
            </div>
            <div className="hidden 2xl:block whitespace-nowrap">
              <span className="text-[#807d72] dark:text-[#a09c92]">Volume: </span>
              <span className="text-[#26251e] dark:text-[#edece6]">{formatCurrency(status.total_volume)}</span>
            </div>
          </div>
        )}

        {/* Speed Controls: full pills on desktop screens (xl: 1200px+), compact cycle button on < 1200px */}
        <div className="hidden xl:flex items-center bg-[#efeee8] dark:bg-[#262520] p-0.5 rounded-[8px] border border-[#e6e5e0] dark:border-[#2c2b26] flex-shrink-0">
          {speeds.map((s) => (
            <button
              key={s}
              onClick={() => onSpeedChange(s)}
              data-tooltip={`Speed: ${s === 10 ? '10x (1 full day per tick)' : `${s}x (${s === 5 ? '6h' : s === 2 ? '2h' : s === 1 ? '1h' : '30m'}/tick)`}`}
              data-tooltip-pos="bottom"
              className={`px-2 py-1 text-[11px] font-mono rounded-[6px] transition-colors flex-shrink-0 ${
                speed === s
                  ? 'bg-white dark:bg-[#1c1b18] text-[#26251e] dark:text-[#edece6] font-semibold shadow-none'
                  : 'text-[#807d72] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#edece6]'
              }`}
            >
              {s}x
            </button>
          ))}
        </div>

        {/* Compact cycle button for all screens < 1200px (including mobile & tablet) */}
        <button
          onClick={() => {
            const nextIdx = (speeds.indexOf(speed) + 1) % speeds.length;
            onSpeedChange(speeds[nextIdx]);
          }}
          className="flex xl:hidden items-center gap-1 px-2 sm:px-2.5 h-8 text-[11px] font-mono rounded-[6px] bg-[#efeee8] dark:bg-[#262520] hover:bg-[#e6e5e0] dark:hover:bg-[#2f2e26] text-[#26251e] dark:text-[#edece6] border border-[#e6e5e0] dark:border-[#2c2b26] flex-shrink-0 transition-colors"
          data-tooltip={`Current speed: ${speed}x (Click to cycle 0.5x, 1x, 2x, 5x, 10x)`}
          data-tooltip-pos="bottom"
          aria-label={`Simulation speed: ${speed}x`}
        >
          <span className="text-[#f54e00] font-bold text-[10px]">⚡</span>
          <span>{speed}x</span>
        </button>

        {/* Theme Toggle Button (Light/Dark/Auto Mode with custom tooltip) */}
        <button
          onClick={toggleTheme}
          className="flex items-center gap-1.5 px-2 sm:px-2.5 h-8 text-[11px] font-mono rounded-[8px] bg-white dark:bg-[#1c1b18] text-[#5a5852] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#edece6] border border-[#e6e5e0] dark:border-[#2c2b26] transition-colors flex-shrink-0"
          data-tooltip={`Theme: ${theme.toUpperCase()} (Click to toggle Light / Dark / Auto)`}
          data-tooltip-pos="bottom"
          aria-label={`Active theme: ${theme.toUpperCase()}`}
        >
          {theme === 'auto' ? (
            <Laptop size={13} className="text-[#f54e00]" />
          ) : isDark ? (
            <Moon size={13} className="text-[#807d72] dark:text-[#a09c92]" />
          ) : (
            <Sun size={13} className="text-[#f54e00]" />
          )}
          <span className="hidden sm:inline">{theme.toUpperCase()}</span>
        </button>

        {/* Mobile & Tablet menu trigger (below lg: 1024px) */}
        <button
          onClick={() => setMobileOpen(!mobileOpen)}
          className="lg:hidden text-[#26251e] dark:text-[#edece6] p-1.5 rounded-[6px] hover:bg-[#efeee8] dark:hover:bg-[#262520] flex-shrink-0"
          aria-label="Toggle Navigation Menu"
        >
          {mobileOpen ? <X size={20} /> : <Menu size={20} />}
        </button>
      </div>

      {/* Mobile and Tablet drawer */}
      {mobileOpen && (
        <div className="absolute top-16 left-0 right-0 bg-[#f7f7f4] dark:bg-[#181714] border-b border-[#e6e5e0] dark:border-[#2c2b26] p-4 flex flex-col gap-3 lg:hidden z-50 shadow-none">
          <div className="flex items-center justify-between pb-2 border-b border-[#efeee8] dark:border-[#262520]">
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-[#1f8a65] animate-pulse" />
              <span className="text-[12px] font-mono text-[#26251e] dark:text-[#edece6] font-semibold">
                {calendarFormatted || `Day ${status?.tick}`}
              </span>
            </div>
            <button
              onClick={toggleTheme}
              className="flex items-center gap-1.5 px-2.5 py-1 text-[11px] font-mono rounded-[6px] bg-white dark:bg-[#1c1b18] text-[#26251e] dark:text-[#edece6] border border-[#e6e5e0] dark:border-[#2c2b26]"
            >
              {theme === 'auto' ? (
                <Laptop size={12} className="text-[#f54e00]" />
              ) : isDark ? (
                <Moon size={12} className="text-[#807d72] dark:text-[#a09c92]" />
              ) : (
                <Sun size={12} className="text-[#f54e00]" />
              )}
              <span>{theme === 'auto' ? 'Auto Canvas' : isDark ? 'Dark Canvas' : 'Light Canvas'}</span>
            </button>
          </div>

          {/* Dedicated full 5-speed picker in mobile drawer */}
          <div className="space-y-1.5">
            <div className="text-[10px] font-mono uppercase tracking-[0.08em] text-[#807d72] dark:text-[#a09c92]">
              Simulation Velocity
            </div>
            <div className="grid grid-cols-5 gap-1 bg-[#efeee8] dark:bg-[#262520] p-1 rounded-[8px] border border-[#e6e5e0] dark:border-[#2c2b26]">
              {speeds.map((s) => (
                <button
                  key={s}
                  onClick={() => onSpeedChange(s)}
                  className={`py-1.5 text-[12px] font-mono rounded-[6px] text-center transition-colors ${
                    speed === s
                      ? 'bg-white dark:bg-[#1c1b18] text-[#26251e] dark:text-[#edece6] font-bold shadow-none'
                      : 'text-[#807d72] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#edece6]'
                  }`}
                >
                  {s}x
                </button>
              ))}
            </div>
          </div>

          {/* Navigation Items */}
          <div className="space-y-1 pt-1">
            <div className="text-[10px] font-mono uppercase tracking-[0.08em] text-[#807d72] dark:text-[#a09c92]">
              Navigation
            </div>
            {allTabs.map((tab) => {
              const Icon = tab.icon;
              const isActive = activeTab === tab.id;
              return (
                <button
                  key={tab.id}
                  onClick={() => handleTabClick(tab.id as ActiveTab)}
                  className={`w-full flex items-center justify-between px-3 py-2.5 rounded-[8px] text-[14px] font-medium text-left transition-colors ${
                    isActive
                      ? 'bg-white dark:bg-[#1c1b18] text-[#26251e] dark:text-[#edece6] border border-[#e6e5e0] dark:border-[#2c2b26]'
                      : 'text-[#5a5852] dark:text-[#a09c92] hover:bg-[#efeee8] dark:hover:bg-[#262520]'
                  }`}
                >
                  <div className="flex items-center gap-2.5">
                    <Icon size={16} className={isActive ? 'text-[#f54e00]' : 'text-[#807d72] dark:text-[#a09c92]'} />
                    <span>{tab.label}</span>
                  </div>
                  {isActive && <Check size={14} className="text-[#f54e00]" />}
                </button>
              );
            })}
          </div>
        </div>
      )}
    </header>
  );
};
