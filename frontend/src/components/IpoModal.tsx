import React, { useState } from 'react';
import { X, Sparkles } from 'lucide-react';

interface IpoModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmitIpo: (data: { sector: string; cap_tier: string; revenue: number }) => Promise<void>;
}

export const IpoModal: React.FC<IpoModalProps> = ({ isOpen, onClose, onSubmitIpo }) => {
  const [sector, setSector] = useState('Information Technology');
  const [capTier, setCapTier] = useState('MegaCap');
  const [revenueBillion, setRevenueBillion] = useState('10');
  const [loading, setLoading] = useState(false);

  if (!isOpen) return null;

  const sectors = [
    'Information Technology',
    'Financials',
    'Health Care',
    'Consumer Discretionary',
    'Communication Services',
    'Industrials',
    'Consumer Staples',
    'Energy',
    'Utilities',
    'Real Estate',
    'Materials',
  ];

  const tiers = ['MegaCap', 'LargeCap', 'MidCap', 'SmallCap'];

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const revenue = (parseFloat(revenueBillion) || 10) * 1e9;
      await onSubmitIpo({ sector, cap_tier: capTier, revenue });
      onClose();
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-[#26251e]/40 backdrop-blur-sm p-4">
      <div className="bg-white border border-[#e6e5e0] rounded-[12px] p-6 max-w-md w-full shadow-none space-y-5">
        <div className="flex items-center justify-between pb-3 border-b border-[#efeee8]">
          <div className="flex items-center gap-2">
            <Sparkles size={18} className="text-[#f54e00]" />
            <h2 className="text-[18px] font-semibold text-[#26251e] tracking-tight">
              Launch Corporate IPO
            </h2>
          </div>
          <button
            onClick={onClose}
            className="text-[#807d72] hover:text-[#26251e] p-1 rounded-[6px]"
          >
            <X size={16} />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-[11px] font-mono uppercase text-[#807d72] mb-1">
              GICS Industry Sector
            </label>
            <select
              value={sector}
              onChange={(e) => setSector(e.target.value)}
              className="w-full bg-white border border-[#e6e5e0] rounded-[8px] px-3 h-10 text-[13px] text-[#26251e] outline-none cursor-pointer"
            >
              {sectors.map((s) => (
                <option key={s} value={s}>{s}</option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-[11px] font-mono uppercase text-[#807d72] mb-1">
              Capitalization Tier
            </label>
            <select
              value={capTier}
              onChange={(e) => setCapTier(e.target.value)}
              className="w-full bg-white border border-[#e6e5e0] rounded-[8px] px-3 h-10 text-[13px] text-[#26251e] outline-none cursor-pointer"
            >
              {tiers.map((t) => (
                <option key={t} value={t}>{t}</option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-[11px] font-mono uppercase text-[#807d72] mb-1">
              Annual Revenue ($ Billions)
            </label>
            <input
              type="number"
              step="1"
              value={revenueBillion}
              onChange={(e) => setRevenueBillion(e.target.value)}
              className="w-full bg-white border border-[#e6e5e0] rounded-[8px] px-3 h-10 font-mono text-[14px] text-[#26251e] outline-none"
            />
          </div>

          <div className="pt-2 flex items-center justify-end gap-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-[13px] font-medium text-[#5a5852] hover:text-[#26251e] rounded-[8px]"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="bg-[#f54e00] hover:bg-[#d04200] active:bg-[#b53a00] text-white font-medium text-[13px] px-5 py-2 rounded-[8px] transition-colors disabled:opacity-50"
            >
              {loading ? 'Syndicating IPO...' : 'Float on Exchange'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
