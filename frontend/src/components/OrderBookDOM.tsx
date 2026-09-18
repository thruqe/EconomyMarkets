import React from 'react';
import { OrderBookDTO } from '../types/api';

interface OrderBookDOMProps {
  book: OrderBookDTO | null;
  onSelectPrice: (price: number) => void;
}

export const OrderBookDOM: React.FC<OrderBookDOMProps> = ({ book, onSelectPrice }) => {
  const asks = (book?.asks || []).slice(0, 7).reverse();
  const bids = (book?.bids || []).slice(0, 7);

  const maxQty = Math.max(
    ...asks.map((lvl) => lvl.quantity),
    ...bids.map((lvl) => lvl.quantity),
    1
  );

  return (
    <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5 flex flex-col h-full shadow-none">
      <div className="pb-3 border-b border-[#efeee8] dark:border-[#262520] flex items-center justify-between">
        <div>
          <h2 className="text-[14px] font-semibold text-[#26251e] dark:text-[#edece6] tracking-tight">Depth of Market</h2>
          <span className="text-[11px] font-mono text-[#807d72] dark:text-[#a09c92]">L2 Double Auction</span>
        </div>
        <span className="text-[11px] font-mono text-[#807d72] dark:text-[#a09c92] bg-[#fafaf7] dark:bg-[#262520] px-2 py-0.5 rounded-[4px] border border-[#efeee8] dark:border-[#2c2b26]">FIFO</span>
      </div>

      {/* Column Headers */}
      <div className="grid grid-cols-2 text-[11px] font-mono uppercase text-[#807d72] dark:text-[#a09c92] py-2 border-b border-[#efeee8] dark:border-[#262520]">
        <div>Price ($)</div>
        <div className="text-right">Quantity</div>
      </div>

      {/* Asks (Sells) */}
      <div className="flex-1 flex flex-col justify-end space-y-1 py-1">
        {asks.map((lvl, idx) => {
          const depthPct = Math.min(100, (lvl.quantity / maxQty) * 100);
          return (
            <div
              key={`ask-${idx}`}
              onClick={() => onSelectPrice(lvl.price)}
              className="relative grid grid-cols-2 text-[12px] font-mono py-1 px-1.5 rounded-[4px] cursor-pointer hover:bg-[#fafaf7] dark:hover:bg-[#262520] transition-colors"
            >
              {/* Depth background fill */}
              <div
                className="absolute right-0 top-0 bottom-0 bg-[#cf2d56]/15 rounded-[4px]"
                style={{ width: `${depthPct}%` }}
              />
              <span className="text-[#cf2d56] font-medium z-10">{lvl.price.toFixed(2)}</span>
              <span className="text-right text-[#5a5852] dark:text-[#a09c92] z-10">{lvl.quantity.toLocaleString()}</span>
            </div>
          );
        })}
      </div>

      {/* Spread Indicator Banner */}
      <div className="my-2 py-1.5 px-3 bg-[#fafaf7] dark:bg-[#262520] border-y border-[#efeee8] dark:border-[#2c2b26] flex items-center justify-between text-[11px] font-mono">
        <span className="text-[#807d72] dark:text-[#a09c92]">Mid: <strong className="text-[#26251e] dark:text-[#edece6]">${book?.mid?.toFixed(2) || '0.00'}</strong></span>
        <span className="text-[#807d72] dark:text-[#a09c92]">Spread: <strong className="text-[#26251e] dark:text-[#edece6]">${book?.spread?.toFixed(2) || '0.00'}</strong></span>
      </div>

      {/* Bids (Buys) */}
      <div className="flex-1 flex flex-col space-y-1 py-1">
        {bids.map((lvl, idx) => {
          const depthPct = Math.min(100, (lvl.quantity / maxQty) * 100);
          return (
            <div
              key={`bid-${idx}`}
              onClick={() => onSelectPrice(lvl.price)}
              className="relative grid grid-cols-2 text-[12px] font-mono py-1 px-1.5 rounded-[4px] cursor-pointer hover:bg-[#fafaf7] dark:hover:bg-[#262520] transition-colors"
            >
              {/* Depth background fill */}
              <div
                className="absolute right-0 top-0 bottom-0 bg-[#1f8a65]/15 rounded-[4px]"
                style={{ width: `${depthPct}%` }}
              />
              <span className="text-[#1f8a65] font-medium z-10">{lvl.price.toFixed(2)}</span>
              <span className="text-right text-[#5a5852] dark:text-[#a09c92] z-10">{lvl.quantity.toLocaleString()}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
};
