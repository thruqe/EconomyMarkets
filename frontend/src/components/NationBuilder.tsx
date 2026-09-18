import React, { useState, useEffect } from 'react';
import {
  Building2,
  GraduationCap,
  HeartPulse,
  Coins,
  ShieldCheck,
  Globe2,
  AlertTriangle,
  MapPin,
  Award,
  TrendingUp,
  Wrench,
  Palmtree,
  Gauge,
  HandCoins,
  Check,
  X,
  Users,
  Briefcase,
  Landmark,
  ShieldAlert,
  Ban,
  Scale,
  Unlock,
  FileText,
  ScrollText,
  RefreshCw,
} from 'lucide-react';
import { NationDTO, WorldDTO, ForeignLoanRequestDTO, BilateralContractDTO } from '../types/api';
import { api } from '../services/api';

interface NationBuilderProps {
  nation: NationDTO | null;
  world: WorldDTO | null;
  onInvest: (pillar: string, amount: number) => Promise<void>;
  onCharter: () => Promise<void>;
  onConfigure: (data: { name: string; currency: string; flag_code: string; founded: number }) => Promise<void>;
  onDiplomacy: (data: { action: string; country_id: string; amount?: number; level?: number; tariff?: number }) => Promise<void>;
  onSetMaintenance?: (rate: number) => Promise<void>;
  onRespondLoan?: (requestId: string, action: 'accept' | 'decline') => Promise<void>;
  onRespondContract?: (contractId: string, action: 'accept' | 'decline') => Promise<void>;
  onBorrowCapital?: (amount: number) => Promise<void>;
  onRepayDebt?: (amount: number) => Promise<void>;
  onTradeReserveCurrency?: (data: { side: 'buy' | 'sell'; amount: number }) => Promise<void>;
  onToggleSanction?: (countryId: string, action: 'sanction' | 'lift') => Promise<void>;
}

export const NationBuilder: React.FC<NationBuilderProps> = ({
  nation,
  world,
  onInvest,
  onCharter,
  onConfigure,
  onDiplomacy,
  onSetMaintenance,
  onRespondLoan,
  onRespondContract,
  onBorrowCapital,
  onRepayDebt,
  onTradeReserveCurrency,
  onToggleSanction,
}) => {
  const [investing, setInvesting] = useState<string | null>(null);
  const [isConfiguring, setIsConfiguring] = useState(false);
  const [name, setName] = useState(nation?.name || 'Republic of Eldoria');
  const [currency, setCurrency] = useState(nation?.currency || 'CRN');
  const [flag, setFlag] = useState(nation?.flag_code || 'EL');
  const [founded, setFounded] = useState(nation?.founded || 2024);
  const [dipMsg, setDipMsg] = useState<string | null>(null);

  // Sovereign Borrowing Modal state
  const [borrowModalOpen, setBorrowModalOpen] = useState(false);
  const [borrowAmount, setBorrowAmount] = useState('1000000000');
  const [borrowBusy, setBorrowBusy] = useState(false);

  // Sovereign Debt Repayment Modal state
  const [repayModalOpen, setRepayModalOpen] = useState(false);
  const [repayAmount, setRepayAmount] = useState('1000000000');
  const [repayBusy, setRepayBusy] = useState(false);

  // FX Reserve Currency (ZTH) Trade Modal state
  const [reserveModalOpen, setReserveModalOpen] = useState(false);
  const [reserveSide, setReserveSide] = useState<'buy' | 'sell'>('buy');
  const [reserveAmount, setReserveAmount] = useState('500000000');
  const [reserveBusy, setReserveBusy] = useState(false);

  // Bilateral Loans Modal state
  const [loanModalOpen, setLoanModalOpen] = useState(false);
  const [loanModalTab, setLoanModalTab] = useState<'pending' | 'approved'>('pending');

  // Sanctions action state
  const [sanctionBusyId, setSanctionBusyId] = useState<string | null>(null);

  // Maintenance slider state
  const [maintenanceRate, setMaintenanceRate] = useState<number>(
    nation?.maintenance_spending_rate || 120_000_000
  );
  const [maintenanceSaving, setMaintenanceSaving] = useState(false);

  // Citizen & Corporate Taxation state
  const [personalTax, setPersonalTax] = useState<number>((nation?.personal_income_tax_rate ?? 0.15) * 100);
  const [corporateTax, setCorporateTax] = useState<number>((nation?.corporate_tax_rate ?? 0.21) * 100);
  const [salesTax, setSalesTax] = useState<number>((nation?.sales_tax_rate ?? 0.10) * 100);
  const [taxSaving, setTaxSaving] = useState(false);

  // Bilateral Borrowing Lender selection
  const [borrowLenderId, setBorrowLenderId] = useState<string>('ZTH');
  const [custodyBusy, setCustodyBusy] = useState(false);

  // Loan & Contract action state
  const [loanBusyId, setLoanBusyId] = useState<string | null>(null);
  const [contractBusyId, setContractBusyId] = useState<string | null>(null);

  useEffect(() => {
    if (nation) {
      if (nation.name) setName(nation.name);
      if (nation.currency) setCurrency(nation.currency);
      if (nation.flag_code) setFlag(nation.flag_code);
      if (nation.founded) setFounded(nation.founded);
      if (nation.maintenance_spending_rate !== undefined) {
        setMaintenanceRate(nation.maintenance_spending_rate);
      }
      if (nation.personal_income_tax_rate !== undefined) setPersonalTax(nation.personal_income_tax_rate * 100);
      if (nation.corporate_tax_rate !== undefined) setCorporateTax(nation.corporate_tax_rate * 100);
      if (nation.sales_tax_rate !== undefined) setSalesTax(nation.sales_tax_rate * 100);
    }
  }, [nation]);

  if (!nation) {
    return (
      <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-8 text-center text-[#807d72] dark:text-[#78756c]">
        Loading sovereign nation registry...
      </div>
    );
  }

  const formatB = (val: number) => {
    if (val >= 1e12) return `$${(val / 1e12).toFixed(2)}T`;
    if (val >= 1e9) return `$${(val / 1e9).toFixed(2)}B`;
    if (val >= 1e6) return `$${(val / 1e6).toFixed(1)}M`;
    return `$${val.toFixed(2)}`;
  };

  const handleInvest = async (pillar: string, amount: number) => {
    setInvesting(pillar);
    try {
      await onInvest(pillar, amount);
    } finally {
      setInvesting(null);
    }
  };

  const handleSaveConfig = async (e: React.FormEvent) => {
    e.preventDefault();
    await onConfigure({ name, currency, flag_code: flag, founded: Number(founded) });
    setIsConfiguring(false);
  };

  const handleSaveMaintenance = async () => {
    if (!onSetMaintenance) return;
    setMaintenanceSaving(true);
    try {
      await onSetMaintenance(maintenanceRate);
      setDipMsg(`Maintenance budget adjusted to ${formatB(maintenanceRate)}/year.`);
      setTimeout(() => setDipMsg(null), 3000);
    } catch (err: any) {
      setDipMsg(err.message || 'Failed to adjust maintenance budget');
    } finally {
      setMaintenanceSaving(false);
    }
  };

  const handleRespondLoanAction = async (reqId: string, action: 'accept' | 'decline') => {
    if (!onRespondLoan) return;
    setLoanBusyId(reqId);
    try {
      await onRespondLoan(reqId, action);
      setDipMsg(`Bilateral loan ${action === 'accept' ? 'approved and disbursed' : 'declined'}.`);
      setTimeout(() => setDipMsg(null), 3500);
    } catch (err: any) {
      setDipMsg(err.message || 'Loan action failed');
    } finally {
      setLoanBusyId(null);
    }
  };

  const handleRespondContractAction = async (conId: string, action: 'accept' | 'decline') => {
    if (!onRespondContract) return;
    setContractBusyId(conId);
    try {
      await onRespondContract(conId, action);
      setDipMsg(`Trade accord ${action === 'accept' ? 'ratified and active' : 'declined'}.`);
      setTimeout(() => setDipMsg(null), 3500);
    } catch (err: any) {
      setDipMsg(err.message || 'Accord action failed');
    } finally {
      setContractBusyId(null);
    }
  };

  const handleBorrowCapital = async () => {
    const amt = parseFloat(borrowAmount);
    if (isNaN(amt) || amt <= 0) return;
    setBorrowBusy(true);
    try {
      const lender = world?.countries.find((c) => c.id === borrowLenderId);
      const rate = lender?.interest_rate || 0.045;
      await api.borrowBilateral({
        lender_id: borrowLenderId,
        amount: amt,
        interest_rate: rate,
      });
      setDipMsg(`Successfully borrowed ${formatB(amt)} sovereign credit from ${lender?.name || borrowLenderId} at ${(rate * 100).toFixed(2)}% annual interest.`);
      setBorrowModalOpen(false);
      setTimeout(() => setDipMsg(null), 4000);
    } catch (err: any) {
      setDipMsg(`Borrowing failed: ${err.message || err}`);
    } finally {
      setBorrowBusy(false);
    }
  };

  const handleSaveTaxes = async () => {
    setTaxSaving(true);
    try {
      await api.setTaxes({
        personal_income_tax_rate: personalTax / 100,
        corporate_tax_rate: corporateTax / 100,
        sales_tax_rate: salesTax / 100,
      });
      setDipMsg(`Tax schedule enacted: PIT ${personalTax.toFixed(0)}%, Corporate ${corporateTax.toFixed(0)}%, Sales/VAT ${salesTax.toFixed(0)}%.`);
      setTimeout(() => setDipMsg(null), 3500);
    } catch (err: any) {
      setDipMsg(err.message || 'Failed to update taxes');
    } finally {
      setTaxSaving(false);
    }
  };

  const handleToggleDebtCustody = async (nationId: string, enable: boolean) => {
    setCustodyBusy(true);
    try {
      await api.toggleDebtCustody({ nation_id: nationId, enable });
      setDipMsg(`Sovereign debt custody ${enable ? 'authorized' : 'revoked'} for ${nationId}.`);
      setTimeout(() => setDipMsg(null), 3500);
    } catch (err: any) {
      setDipMsg(err.message || 'Failed to toggle debt custody');
    } finally {
      setCustodyBusy(false);
    }
  };

  const handleRepayDebt = async () => {
    const amt = parseFloat(repayAmount);
    if (isNaN(amt) || amt <= 0) return;
    setRepayBusy(true);
    try {
      if (onRepayDebt) {
        await onRepayDebt(amt);
      } else {
        await api.repayDebt(amt);
      }
      setDipMsg(`Successfully repaid ${formatB(amt)} of debt from treasury.`);
      setRepayModalOpen(false);
      setTimeout(() => setDipMsg(null), 4000);
    } catch (err: any) {
      setDipMsg(`Debt repayment failed: ${err.message || err}`);
    } finally {
      setRepayBusy(false);
    }
  };

  const handleTradeReserveCurrency = async () => {
    const amt = parseFloat(reserveAmount);
    if (isNaN(amt) || amt <= 0) return;
    setReserveBusy(true);
    try {
      if (onTradeReserveCurrency) {
        await onTradeReserveCurrency({ side: reserveSide, amount: amt });
      } else {
        await api.tradeReserveCurrency({ side: reserveSide, amount: amt });
      }
      setDipMsg(`Reserve currency order executed: ${reserveSide.toUpperCase()} ${formatB(amt)} ZTH.`);
      setReserveModalOpen(false);
      setTimeout(() => setDipMsg(null), 4000);
    } catch (err: any) {
      setDipMsg(`Reserve trade failed: ${err.message || err}`);
    } finally {
      setReserveBusy(false);
    }
  };

  const handleToggleSanction = async (countryId: string, currentSanctioned: boolean) => {
    setSanctionBusyId(countryId);
    const action = currentSanctioned ? 'lift' : 'sanction';
    try {
      if (onToggleSanction) {
        await onToggleSanction(countryId, action);
      } else {
        await api.toggleSanctions(countryId, action);
      }
      setDipMsg(`Sanctions status with ${countryId} successfully updated to: ${action.toUpperCase()}.`);
      setTimeout(() => setDipMsg(null), 4000);
    } catch (err: any) {
      setDipMsg(`Sanctions action failed: ${err.message || err}`);
    } finally {
      setSanctionBusyId(null);
    }
  };

  const canCharter = nation.infrastructure_level >= 20 && nation.education_level >= 20;

  // Active loan requests
  const loanRequests: ForeignLoanRequestDTO[] =
    world?.loan_requests && world.loan_requests.length > 0
      ? world.loan_requests
      : nation.loan_requests || [];

  return (
    <div className="space-y-8">
      {/* Editorial Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <div className="text-[11px] font-mono uppercase tracking-[0.08em] text-[#807d72] dark:text-[#a09c92] mb-1">
            Economic Governance · Sovereign Balance Sheet & Reserves
          </div>
          <h1 className="text-[28px] lg:text-[32px] font-normal tracking-[-0.03em] text-[#26251e] dark:text-[#edece6] leading-[1.2]">
            National Economy & Development
          </h1>
          <p className="text-[14px] lg:text-[15px] text-[#5a5852] dark:text-[#a09c92] mt-1 max-w-2xl leading-[1.5]">
            Manage public infrastructure investments, sovereign debt, global foreign exchange reserves, demographics, and international credit agreements.
          </p>
        </div>

        <button
          onClick={() => setIsConfiguring(!isConfiguring)}
          className="bg-white dark:bg-[#1c1b18] hover:bg-[#fafaf7] dark:hover:bg-[#262520] text-[#26251e] dark:text-[#edece6] border border-[#e6e5e0] dark:border-[#2c2b26] px-4 h-9 rounded-[8px] text-[13px] font-medium transition-colors self-start md:self-auto"
        >
          {isConfiguring ? 'Cancel Setup' : 'Edit Country Profile'}
        </button>
      </div>

      {/* Nation Configuration Form */}
      {isConfiguring && (
        <form
          onSubmit={handleSaveConfig}
          className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-6 space-y-4"
        >
          <h2 className="text-[16px] font-semibold text-[#26251e] dark:text-[#edece6]">Country Profile Settings</h2>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div>
              <label className="block text-[11px] font-mono uppercase text-[#807d72] dark:text-[#a09c92] mb-1">Nation Name</label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Republic of Eldoria"
                className="w-full bg-white dark:bg-[#151412] border border-[#e6e5e0] dark:border-[#2c2b26] focus:border-[#26251e] dark:focus:border-[#edece6] rounded-[8px] px-3 h-9 text-[13px] text-[#26251e] dark:text-[#edece6] outline-none"
              />
            </div>
            <div>
              <label className="block text-[11px] font-mono uppercase text-[#807d72] dark:text-[#a09c92] mb-1">National Currency</label>
              <input
                type="text"
                value={currency}
                onChange={(e) => setCurrency(e.target.value)}
                placeholder="e.g. CRN"
                className="w-full bg-white dark:bg-[#151412] border border-[#e6e5e0] dark:border-[#2c2b26] focus:border-[#26251e] dark:focus:border-[#edece6] rounded-[8px] px-3 h-9 font-mono text-[13px] text-[#26251e] dark:text-[#edece6] outline-none uppercase"
              />
            </div>
            <div>
              <label className="block text-[11px] font-mono uppercase text-[#807d72] dark:text-[#a09c92] mb-1">Flag Code (ISO)</label>
              <input
                type="text"
                value={flag}
                onChange={(e) => setFlag(e.target.value)}
                placeholder="e.g. EL"
                className="w-full bg-white dark:bg-[#151412] border border-[#e6e5e0] dark:border-[#2c2b26] focus:border-[#26251e] dark:focus:border-[#edece6] rounded-[8px] px-3 h-9 font-mono text-[13px] text-[#26251e] dark:text-[#edece6] outline-none uppercase"
              />
            </div>
            <div>
              <label className="block text-[11px] font-mono uppercase text-[#807d72] dark:text-[#a09c92] mb-1">Founding Year</label>
              <input
                type="number"
                value={founded}
                onChange={(e) => setFounded(Number(e.target.value))}
                className="w-full bg-white dark:bg-[#151412] border border-[#e6e5e0] dark:border-[#2c2b26] focus:border-[#26251e] dark:focus:border-[#edece6] rounded-[8px] px-3 h-9 font-mono text-[13px] text-[#26251e] dark:text-[#edece6] outline-none"
              />
            </div>
          </div>
          <button
            type="submit"
            className="bg-[#f54e00] hover:bg-[#d04200] text-white px-4 h-9 rounded-[8px] text-[13px] font-medium transition-colors"
          >
            Save Profile
          </button>
        </form>
      )}

      {/* Identity, Treasury, FX Reserves & Status Tier Banner */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
        {/* Status Tier */}
        <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5">
          <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#78756c] flex items-center gap-1">
            <Award size={13} className="text-[#f54e00]" />
            <span>Development Tier</span>
          </div>
          <div className="text-[20px] font-semibold tracking-[-0.02em] text-[#26251e] dark:text-[#edece6] mt-1 truncate">
            {nation.tier_name || 'Developing Nation'}
          </div>
          <div className="text-[12px] text-[#5a5852] dark:text-[#a09c92] mt-2 flex items-center justify-between">
            <span>Progress:</span>
            <span className="font-mono font-medium text-[#26251e] dark:text-[#edece6]">{(nation.tier_progress || 25).toFixed(1)}%</span>
          </div>
          <div className="w-full bg-[#efeee8] dark:bg-[#262520] h-1.5 rounded-full overflow-hidden mt-1.5">
            <div
              className="bg-[#26251e] dark:bg-[#edece6] h-full rounded-full transition-all"
              style={{ width: `${Math.min(100, Math.max(5, nation.tier_progress || 25))}%` }}
            />
          </div>
        </div>

        {/* Treasury Cash & Debt */}
        <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5 flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between gap-1">
              <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#78756c]">Treasury Cash</div>
              <div className="flex items-center gap-1">
                <button
                  onClick={() => setBorrowModalOpen(true)}
                  className="px-2 py-0.5 rounded-[5px] bg-[#26251e] hover:bg-[#3d3b32] dark:bg-[#edece6] dark:hover:bg-[#d8d6ce] text-white dark:text-[#141412] text-[10px] font-mono font-medium transition-colors"
                  title="Borrow capital"
                >
                  + Borrow
                </button>
                <button
                  onClick={() => setRepayModalOpen(true)}
                  className="px-2 py-0.5 rounded-[5px] bg-[#f54e00] hover:bg-[#d04200] text-white text-[10px] font-mono font-medium transition-colors"
                  title="Repay debt"
                >
                  - Repay
                </button>
              </div>
            </div>
            <div className="text-[22px] font-normal tracking-[-0.025em] text-[#26251e] dark:text-[#edece6] font-mono mt-1">
              {formatB(nation.treasury_cash)}
            </div>
          </div>
          <div className="text-[12px] text-[#5a5852] dark:text-[#a09c92] mt-2 pt-2 border-t border-[#efeee8] dark:border-[#262520] flex justify-between items-center">
            <span>Debt:</span>
            <span className="font-mono text-[#26251e] dark:text-[#edece6] font-semibold">{formatB(nation.national_debt)}</span>
          </div>
        </div>

        {/* Foreign FX Reserves (ZTH Global Reserve) */}
        <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5 flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between">
              <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#78756c] flex items-center gap-1">
                <Globe2 size={12} className="text-[#f54e00]" />
                <span>FX Reserves</span>
              </div>
              <button
                onClick={() => setReserveModalOpen(true)}
                className="px-2 py-0.5 rounded-[5px] border border-[#e6e5e0] dark:border-[#2c2b26] hover:border-[#26251e] dark:hover:border-[#edece6] text-[#26251e] dark:text-[#edece6] text-[10px] font-mono font-medium transition-colors"
              >
                Trade ZTH
              </button>
            </div>
            <div className="text-[22px] font-normal tracking-[-0.025em] text-[#26251e] dark:text-[#edece6] font-mono mt-1">
              {formatB(nation.reserve_currency_reserves || 0)}
            </div>
          </div>
          <div className="text-[12px] text-[#5a5852] dark:text-[#a09c92] mt-2 pt-2 border-t border-[#efeee8] dark:border-[#262520] flex justify-between items-center">
            <span className="flex items-center gap-1">
              <span className="inline-block w-1.5 h-1.5 rounded-full bg-[#1f8a65]" />
              Reserve Peg:
            </span>
            <span className="font-mono text-[#26251e] dark:text-[#edece6] font-medium">ZTH (Global)</span>
          </div>
        </div>

        {/* Sovereign Credit Rating */}
        <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5">
          <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#78756c]">Credit Rating</div>
          <div className="text-[22px] font-normal tracking-[-0.025em] text-[#26251e] dark:text-[#edece6] font-mono mt-1 flex items-baseline gap-2">
            <span>{nation.credit_rating}</span>
            <span className="text-[11px] text-[#807d72] dark:text-[#78756c] font-normal">
              Yield: {(nation.borrowing_yield * 100).toFixed(2)}%
            </span>
          </div>
          <div className="text-[12px] text-[#5a5852] dark:text-[#a09c92] mt-2">
            Flag: <span className="font-mono text-[#26251e] dark:text-[#edece6] font-medium">{nation.flag_code}</span> · Currency: <span className="font-mono font-medium text-[#26251e] dark:text-[#edece6]">{nation.currency}</span>
          </div>
        </div>

        {/* Securities Exchange */}
        <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5 flex flex-col justify-between">
          <div>
            <div className="text-[11px] font-mono uppercase text-[#807d72] dark:text-[#78756c]">Securities Exchange</div>
            <div className="text-[15px] font-semibold text-[#26251e] dark:text-[#edece6] mt-1 flex items-center gap-2">
              {nation.exchange_chartered ? (
                <span className="flex items-center gap-1.5 text-[#1f8a65]">
                  <ShieldCheck size={16} />
                  <span>Chartered & Active</span>
                </span>
              ) : (
                <span className="flex items-center gap-1.5 text-[#cf2d56]">
                  <AlertTriangle size={16} />
                  <span>Charter Pending</span>
                </span>
              )}
            </div>
            <div className="text-[11px] text-[#807d72] dark:text-[#78756c] mt-1">
              Req: Infra ≥ 20% & Edu ≥ 20%
            </div>
          </div>

          {!nation.exchange_chartered && (
            <button
              onClick={onCharter}
              disabled={!canCharter}
              className="mt-2 w-full bg-[#f54e00] hover:bg-[#d04200] disabled:bg-[#efeee8] dark:disabled:bg-[#262520] disabled:text-[#a09c92] text-white text-[11px] font-medium h-7 rounded-[6px] transition-colors"
            >
              {canCharter ? 'Charter Exchange' : 'Prerequisites Unmet'}
            </button>
          )}
        </div>
      </div>

      {/* MAINTENANCE & PRODUCTIVITY */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Maintenance Manager */}
        <div className="lg:col-span-2 bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-6 space-y-4">
          <div className="flex items-start justify-between">
            <div>
              <div className="flex items-center gap-2 text-[#26251e] dark:text-[#edece6]">
                <Wrench size={18} className="text-[#f54e00]" />
                <h2 className="text-[17px] font-semibold tracking-tight">
                  Infrastructure Maintenance
                </h2>
              </div>
              <p className="text-[13px] text-[#5a5852] dark:text-[#a09c92] mt-1">
                Public capital assets experience depreciation (~2.5%/yr). Budgeted maintenance spending counters wear-and-tear and maintains domestic output capacity.
              </p>
            </div>
            <span className="text-[11px] font-mono px-2 py-0.5 rounded-[4px] bg-[#fff3e0] dark:bg-[#332210] text-[#e65100] dark:text-[#ff9800] border border-[#ffe0b2] dark:border-[#553310] whitespace-nowrap">
              Depreciation: 2.5%/yr
            </span>
          </div>

          <div className="bg-[#fafaf7] dark:bg-[#151412] p-4 rounded-[10px] border border-[#efeee8] dark:border-[#262520] space-y-3">
            <div className="flex items-center justify-between text-[13px]">
              <span className="text-[#5a5852] dark:text-[#a09c92] font-medium">Annual Maintenance Outlay:</span>
              <span className="font-mono font-semibold text-[#26251e] dark:text-[#edece6] text-[16px]">
                {formatB(maintenanceRate)} / year
              </span>
            </div>

            <input
              type="range"
              min="0"
              max="400000000"
              step="10000000"
              value={maintenanceRate}
              onChange={(e) => setMaintenanceRate(parseFloat(e.target.value))}
              className="w-full accent-[#f54e00] cursor-pointer"
            />

            <div className="flex items-center justify-between text-[11px] font-mono text-[#807d72] dark:text-[#78756c]">
              <span>$0M (Neglect)</span>
              <span>$120M (Breakeven)</span>
              <span>$400M (Modernizing)</span>
            </div>
          </div>

          <div className="flex items-center justify-between pt-2">
            <div className="text-[12px] text-[#5a5852] dark:text-[#a09c92]">
              Status:{' '}
              {maintenanceRate >= 180_000_000 ? (
                <span className="text-[#1f8a65] font-medium">Actively Modernizing Capital</span>
              ) : maintenanceRate >= 100_000_000 ? (
                <span className="text-[#2c5282] dark:text-[#63b3ed] font-medium">Stable (Counters Natural Wear)</span>
              ) : (
                <span className="text-[#cf2d56] font-medium">Underfunded (Assets Degrading)</span>
              )}
            </div>
            <button
              onClick={handleSaveMaintenance}
              disabled={maintenanceSaving}
              className="px-4 py-2 rounded-[8px] bg-[#26251e] hover:bg-[#3d3b32] dark:bg-[#edece6] dark:hover:bg-[#d8d6ce] text-white dark:text-[#141412] text-[12px] font-medium transition-colors disabled:opacity-50"
            >
              {maintenanceSaving ? 'Saving...' : 'Update Maintenance Budget'}
            </button>
          </div>
        </div>

        {/* Tourism & Productivity */}
        <div className="space-y-4">
          {/* Tourism */}
          <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5 space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-1.5 text-[#26251e] dark:text-[#edece6]">
                <Palmtree size={16} className="text-[#1f8a65]" />
                <span className="font-semibold text-[14px]">Tourism Inflows</span>
              </div>
              <span className="font-mono text-[11px] text-[#1f8a65] font-medium">
                Treasury Receipt
              </span>
            </div>
            <div className="text-[24px] font-mono text-[#26251e] dark:text-[#edece6] font-normal">
              {formatB(nation.tourism_revenue || 65_000_000)} <span className="text-[12px] text-[#807d72] dark:text-[#78756c]">/ yr</span>
            </div>
            <p className="text-[12px] text-[#807d72] dark:text-[#78756c] leading-relaxed">
              Receipts correlate with healthcare coverage ({(nation.healthcare_level).toFixed(0)}%), transport connectivity ({(nation.infrastructure_level).toFixed(0)}%), and public security.
            </p>
          </div>

          {/* Productivity (TFP) Index */}
          <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5 space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-1.5 text-[#26251e] dark:text-[#edece6]">
                <Gauge size={16} className="text-[#f54e00]" />
                <span className="font-semibold text-[14px]">Labor Productivity (TFP)</span>
              </div>
              <span className="font-mono text-[11px] text-[#f54e00] font-medium">
                Output Multiplier
              </span>
            </div>
            <div className="text-[24px] font-mono text-[#26251e] dark:text-[#edece6] font-normal">
              {(nation.productivity_index || 100.0).toFixed(1)} <span className="text-[12px] text-[#807d72] dark:text-[#78756c]">pts</span>
            </div>
            <p className="text-[12px] text-[#807d72] dark:text-[#78756c] leading-relaxed">
              Driven by STEM workforce skills, infrastructure quality, and private enterprise commercialization grants.
            </p>
          </div>
        </div>
      </div>

      {/* 4 Public Development Pillars */}
      <div>
        <div className="mb-3">
          <h2 className="text-[18px] font-semibold text-[#26251e] dark:text-[#edece6] tracking-tight">
            National Capital Investment
          </h2>
          <p className="text-[13px] text-[#807d72] dark:text-[#78756c]">
            Allocate treasury capital across key public sectors to expand productivity and domestic economic capacity.
          </p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
          {/* Infrastructure */}
          <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5 space-y-3 flex flex-col justify-between">
            <div>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 text-[#26251e] dark:text-[#edece6]">
                  <Building2 size={16} className="text-[#f54e00]" />
                  <span className="font-semibold text-[14px]">Infrastructure</span>
                </div>
                <span className="font-mono font-medium text-[14px] text-[#26251e] dark:text-[#edece6]">
                  {nation.infrastructure_level.toFixed(1)}%
                </span>
              </div>
              <div className="w-full bg-[#efeee8] dark:bg-[#262520] h-1.5 rounded-full overflow-hidden mt-2">
                <div
                  className="bg-[#26251e] dark:bg-[#edece6] h-full rounded-full transition-all"
                  style={{ width: `${Math.min(100, nation.infrastructure_level)}%` }}
                />
              </div>
              <p className="text-[12px] text-[#807d72] dark:text-[#78756c] mt-2">
                Logistics, deepwater ports, power grids & transit. Expands export capacity and tourism transit.
              </p>
            </div>

            <div className="flex items-center gap-1.5 pt-3 border-t border-[#efeee8] dark:border-[#262520]">
              <button
                onClick={() => handleInvest('infrastructure', 100_000_000)}
                disabled={investing !== null || nation.treasury_cash < 100_000_000}
                className="flex-1 py-1.5 text-[11px] font-mono rounded-[6px] border border-[#e6e5e0] dark:border-[#2c2b26] hover:border-[#26251e] dark:hover:border-[#edece6] text-[#26251e] dark:text-[#edece6] transition-colors disabled:opacity-40"
              >
                + $100M
              </button>
              <button
                onClick={() => handleInvest('infrastructure', 250_000_000)}
                disabled={investing !== null || nation.treasury_cash < 250_000_000}
                className="flex-1 py-1.5 text-[11px] font-mono rounded-[6px] border border-[#e6e5e0] dark:border-[#2c2b26] hover:border-[#26251e] dark:hover:border-[#edece6] text-[#26251e] dark:text-[#edece6] transition-colors disabled:opacity-40"
              >
                + $250M
              </button>
            </div>
          </div>

          {/* Healthcare */}
          <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5 space-y-3 flex flex-col justify-between">
            <div>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 text-[#26251e] dark:text-[#edece6]">
                  <HeartPulse size={16} className="text-[#f54e00]" />
                  <span className="font-semibold text-[14px]">Healthcare</span>
                </div>
                <span className="font-mono font-medium text-[14px] text-[#26251e] dark:text-[#edece6]">
                  {nation.healthcare_level.toFixed(1)}%
                </span>
              </div>
              <div className="w-full bg-[#efeee8] dark:bg-[#262520] h-1.5 rounded-full overflow-hidden mt-2">
                <div
                  className="bg-[#26251e] dark:bg-[#edece6] h-full rounded-full transition-all"
                  style={{ width: `${Math.min(100, nation.healthcare_level)}%` }}
                />
              </div>
              <p className="text-[12px] text-[#807d72] dark:text-[#78756c] mt-2">
                Public health, clinic networks & wellness. Drives labor participation and tourism safety.
              </p>
            </div>

            <div className="flex items-center gap-1.5 pt-3 border-t border-[#efeee8] dark:border-[#262520]">
              <button
                onClick={() => handleInvest('healthcare', 100_000_000)}
                disabled={investing !== null || nation.treasury_cash < 100_000_000}
                className="flex-1 py-1.5 text-[11px] font-mono rounded-[6px] border border-[#e6e5e0] dark:border-[#2c2b26] hover:border-[#26251e] dark:hover:border-[#edece6] text-[#26251e] dark:text-[#edece6] transition-colors disabled:opacity-40"
              >
                + $100M
              </button>
              <button
                onClick={() => handleInvest('healthcare', 250_000_000)}
                disabled={investing !== null || nation.treasury_cash < 250_000_000}
                className="flex-1 py-1.5 text-[11px] font-mono rounded-[6px] border border-[#e6e5e0] dark:border-[#2c2b26] hover:border-[#26251e] dark:hover:border-[#edece6] text-[#26251e] dark:text-[#edece6] transition-colors disabled:opacity-40"
              >
                + $250M
              </button>
            </div>
          </div>

          {/* Education & R&D */}
          <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5 space-y-3 flex flex-col justify-between">
            <div>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 text-[#26251e] dark:text-[#edece6]">
                  <GraduationCap size={16} className="text-[#f54e00]" />
                  <span className="font-semibold text-[14px]">Education & R&D</span>
                </div>
                <span className="font-mono font-medium text-[14px] text-[#26251e] dark:text-[#edece6]">
                  {nation.education_level.toFixed(1)}%
                </span>
              </div>
              <div className="w-full bg-[#efeee8] dark:bg-[#262520] h-1.5 rounded-full overflow-hidden mt-2">
                <div
                  className="bg-[#26251e] dark:bg-[#edece6] h-full rounded-full transition-all"
                  style={{ width: `${Math.min(100, nation.education_level)}%` }}
                />
              </div>
              <p className="text-[12px] text-[#807d72] dark:text-[#78756c] mt-2">
                Universities, research labs & STEM training. Powers Total Factor Productivity and market charter.
              </p>
            </div>

            <div className="flex items-center gap-1.5 pt-3 border-t border-[#efeee8] dark:border-[#262520]">
              <button
                onClick={() => handleInvest('education', 100_000_000)}
                disabled={investing !== null || nation.treasury_cash < 100_000_000}
                className="flex-1 py-1.5 text-[11px] font-mono rounded-[6px] border border-[#e6e5e0] dark:border-[#2c2b26] hover:border-[#26251e] dark:hover:border-[#edece6] text-[#26251e] dark:text-[#edece6] transition-colors disabled:opacity-40"
              >
                + $100M
              </button>
              <button
                onClick={() => handleInvest('education', 250_000_000)}
                disabled={investing !== null || nation.treasury_cash < 250_000_000}
                className="flex-1 py-1.5 text-[11px] font-mono rounded-[6px] border border-[#e6e5e0] dark:border-[#2c2b26] hover:border-[#26251e] dark:hover:border-[#edece6] text-[#26251e] dark:text-[#edece6] transition-colors disabled:opacity-40"
              >
                + $250M
              </button>
            </div>
          </div>

          {/* Enterprise Grants */}
          <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-5 space-y-3 flex flex-col justify-between">
            <div>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 text-[#26251e] dark:text-[#edece6]">
                  <Coins size={16} className="text-[#f54e00]" />
                  <span className="font-semibold text-[14px]">Enterprise Grants</span>
                </div>
                <span className="font-mono font-medium text-[14px] text-[#26251e] dark:text-[#edece6]">
                  {nation.enterprise_grants_level.toFixed(1)}%
                </span>
              </div>
              <div className="w-full bg-[#efeee8] dark:bg-[#262520] h-1.5 rounded-full overflow-hidden mt-2">
                <div
                  className="bg-[#26251e] dark:bg-[#edece6] h-full rounded-full transition-all"
                  style={{ width: `${Math.min(100, nation.enterprise_grants_level)}%` }}
                />
              </div>
              <p className="text-[12px] text-[#807d72] dark:text-[#78756c] mt-2">
                Seed grants, incubator funding & commercial subsidies. Nurtures private ventures into IPO listings.
              </p>
            </div>

            <div className="flex items-center gap-1.5 pt-3 border-t border-[#efeee8] dark:border-[#262520]">
              <button
                onClick={() => handleInvest('enterprise', 100_000_000)}
                disabled={investing !== null || nation.treasury_cash < 100_000_000}
                className="flex-1 py-1.5 text-[11px] font-mono rounded-[6px] border border-[#e6e5e0] dark:border-[#2c2b26] hover:border-[#26251e] dark:hover:border-[#edece6] text-[#26251e] dark:text-[#edece6] transition-colors disabled:opacity-40"
              >
                + $100M
              </button>
              <button
                onClick={() => handleInvest('enterprise', 250_000_000)}
                disabled={investing !== null || nation.treasury_cash < 250_000_000}
                className="flex-1 py-1.5 text-[11px] font-mono rounded-[6px] border border-[#e6e5e0] dark:border-[#2c2b26] hover:border-[#26251e] dark:hover:border-[#edece6] text-[#26251e] dark:text-[#edece6] transition-colors disabled:opacity-40"
              >
                + $250M
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* BILATERAL LOANS OVERVIEW CARD */}
      <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-6 space-y-4">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-3 pb-3 border-b border-[#efeee8] dark:border-[#262520]">
          <div>
            <div className="flex items-center gap-2">
              <HandCoins size={18} className="text-[#f54e00]" />
              <h2 className="text-[18px] font-semibold text-[#26251e] dark:text-[#edece6] tracking-tight">
                Bilateral Loans & Credit
              </h2>
            </div>
            <p className="text-[13px] text-[#807d72] dark:text-[#78756c] mt-1">
              Foreign sovereign states issue financial loan requests for infrastructure, reserves, and development. Review incoming petitions, disburse credit, and track ongoing interest yield payments.
            </p>
          </div>
          <div className="flex items-center gap-2 self-start md:self-auto">
            <button
              onClick={() => { setLoanModalTab('pending'); setLoanModalOpen(true); }}
              className="px-3 py-1.5 rounded-[8px] bg-[#f54e00] hover:bg-[#d04200] text-white font-medium text-[12px] transition-colors flex items-center gap-1.5"
            >
              <HandCoins size={14} />
              <span>Review Requests ({loanRequests.filter((r) => r.status === 'pending').length} Pending)</span>
            </button>
            <button
              onClick={() => { setLoanModalTab('approved'); setLoanModalOpen(true); }}
              className="px-3 py-1.5 rounded-[8px] bg-[#fafaf7] dark:bg-[#262520] hover:bg-[#efeee8] dark:hover:bg-[#33322b] border border-[#e6e5e0] dark:border-[#2c2b26] text-[#26251e] dark:text-[#edece6] font-medium text-[12px] transition-colors"
            >
              Approved Loans ({loanRequests.filter((r) => r.status === 'accepted' || r.status === 'matured_repaid').length})
            </button>
          </div>
        </div>

        {loanRequests.length === 0 ? (
          <div className="py-8 text-center text-[#807d72] dark:text-[#78756c] text-[13px]">
            No outstanding foreign loan requests at this time.
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
            {loanRequests.slice(0, 3).map((req) => {
              const isPending = req.status === 'pending';
              const isAccepted = req.status === 'accepted';
              const isMatured = req.status === 'matured_repaid';
              return (
                <div
                  key={req.id}
                  className={`p-4 rounded-[10px] border transition-all flex flex-col justify-between ${
                    isAccepted || isMatured
                      ? 'bg-[#1f8a65]/5 border-[#1f8a65]/30'
                      : req.status === 'declined'
                      ? 'bg-[#fafaf7] dark:bg-[#151412] border-[#efeee8] dark:border-[#262520] opacity-60'
                      : 'bg-white dark:bg-[#1c1b18] border-[#e6e5e0] dark:border-[#2c2b26] hover:border-[#26251e] dark:hover:border-[#edece6]'
                  }`}
                >
                  <div>
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-[11px] font-mono px-2 py-0.5 rounded-[4px] bg-[#efeee8] dark:bg-[#262520] text-[#26251e] dark:text-[#edece6] font-semibold border border-[#e6e5e0] dark:border-[#2c2b26]">
                        {req.country_id} · {req.country_name}
                      </span>
                      <span
                        className={`text-[10px] font-mono uppercase px-2 py-0.5 rounded-full ${
                          isAccepted || isMatured
                            ? 'bg-[#1f8a65] text-white'
                            : isPending
                            ? 'bg-[#f54e00]/15 text-[#f54e00] font-semibold'
                            : 'bg-[#efeee8] dark:bg-[#262520] text-[#807d72] dark:text-[#78756c]'
                        }`}
                      >
                        {isMatured ? 'Matured & Repaid' : req.status}
                      </span>
                    </div>

                    <div className="text-[14px] font-medium text-[#26251e] dark:text-[#edece6] mb-2 leading-snug">
                      {req.purpose}
                    </div>

                    <div className="space-y-1.5 text-[12px] font-mono text-[#5a5852] dark:text-[#a09c92] bg-[#fafaf7] dark:bg-[#151412] p-2.5 rounded-[6px] border border-[#efeee8] dark:border-[#262520]">
                      <div className="flex justify-between">
                        <span>Principal:</span>
                        <span className="font-semibold text-[#26251e] dark:text-[#edece6]">{formatB(req.amount)}</span>
                      </div>
                      <div className="flex justify-between">
                        <span>Interest Yield:</span>
                        <span className="font-semibold text-[#1f8a65]">{(req.interest_rate * 100).toFixed(2)}% / yr</span>
                      </div>
                      {req.outcome_message && (
                        <div className="text-[11px] text-[#807d72] dark:text-[#78756c] pt-1 border-t border-[#efeee8] dark:border-[#262520]">
                          {req.outcome_message}
                        </div>
                      )}
                    </div>
                  </div>

                  {isPending ? (
                    <div className="grid grid-cols-2 gap-2 mt-4 pt-3 border-t border-[#efeee8] dark:border-[#262520]">
                      <button
                        onClick={() => handleRespondLoanAction(req.id, 'accept')}
                        disabled={loanBusyId === req.id || nation.treasury_cash < req.amount}
                        className="py-1.5 rounded-[6px] bg-[#1f8a65] hover:bg-[#186e50] text-white text-[11px] font-medium flex items-center justify-center gap-1 transition-colors disabled:opacity-40"
                      >
                        <Check size={12} />
                        <span>Fund Loan</span>
                      </button>
                      <button
                        onClick={() => handleRespondLoanAction(req.id, 'decline')}
                        disabled={loanBusyId === req.id}
                        className="py-1.5 rounded-[6px] bg-[#fafaf7] dark:bg-[#262520] hover:bg-[#efeee8] dark:hover:bg-[#33322b] text-[#5a5852] dark:text-[#a09c92] border border-[#e6e5e0] dark:border-[#2c2b26] text-[11px] font-medium flex items-center justify-center gap-1 transition-colors"
                      >
                        <X size={12} />
                        <span>Decline</span>
                      </button>
                    </div>
                  ) : (
                    <div className="mt-3 pt-2 text-right">
                      <button
                        onClick={() => { setLoanModalTab('approved'); setLoanModalOpen(true); }}
                        className="text-[11px] font-mono text-[#f54e00] hover:underline"
                      >
                        View Approved Loans →
                      </button>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* CITIZEN METRICS & DEMOGRAPHICS */}
      <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-6 space-y-6">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-2 pb-3 border-b border-[#efeee8] dark:border-[#262520]">
          <div>
            <div className="flex items-center gap-2">
              <Users size={18} className="text-[#f54e00]" />
              <h2 className="text-[18px] font-semibold text-[#26251e] dark:text-[#edece6] tracking-tight">
                Demographics & Labor Force
              </h2>
            </div>
            <p className="text-[13px] text-[#807d72] dark:text-[#78756c] mt-0.5">
              Key indicators on populace demographics, workforce participation, median wages, and policy impact on living standards.
            </p>
          </div>
          <div className="flex items-center gap-2 font-mono text-[11px] text-[#5a5852] dark:text-[#a09c92] bg-[#fafaf7] dark:bg-[#151412] px-2.5 py-1 rounded-[6px] border border-[#e6e5e0] dark:border-[#2c2b26] self-start md:self-auto">
            <span>Population:</span>
            <span className="font-semibold text-[#26251e] dark:text-[#edece6]">{((nation.population || 18_500_000) / 1e6).toFixed(2)}M</span>
          </div>
        </div>

        {/* 4 Citizen Telemetry Metric Cards */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="p-4 rounded-[8px] bg-[#fafaf7] dark:bg-[#151412] border border-[#efeee8] dark:border-[#262520] space-y-2">
            <div className="flex items-center justify-between text-[#807d72] dark:text-[#78756c] text-[11px] font-mono uppercase">
              <span>Labor Participation</span>
              <Briefcase size={14} className="text-[#1f8a65]" />
            </div>
            <div className="text-[20px] font-mono font-semibold text-[#26251e] dark:text-[#edece6]">
              47.5%
            </div>
            <div className="text-[12px] text-[#5a5852] dark:text-[#a09c92] space-y-1 pt-1 border-t border-[#efeee8] dark:border-[#262520]">
              <div className="flex justify-between">
                <span>Labor Force:</span>
                <span className="font-mono text-[#26251e] dark:text-[#edece6] font-medium">{(((nation.population || 18_500_000) * 0.475) / 1e6).toFixed(2)}M</span>
              </div>
              <div className="flex justify-between">
                <span>Unemployment:</span>
                <span className="font-mono text-[#1f8a65] font-medium">5.3%</span>
              </div>
              <div className="flex justify-between">
                <span>Job Openings:</span>
                <span className="font-mono text-[#26251e] dark:text-[#edece6]">{Math.round(nation.job_openings || 142000).toLocaleString()}</span>
              </div>
            </div>
          </div>

          <div className="p-4 rounded-[8px] bg-[#fafaf7] dark:bg-[#151412] border border-[#efeee8] dark:border-[#262520] space-y-2">
            <div className="flex items-center justify-between text-[#807d72] dark:text-[#78756c] text-[11px] font-mono uppercase">
              <span>Wages & Income</span>
              <Coins size={14} className="text-[#f54e00]" />
            </div>
            <div className="text-[20px] font-mono font-semibold text-[#26251e] dark:text-[#edece6]">
              ${(nation.average_hourly_wage || 3.25).toFixed(2)} <span className="text-[12px] text-[#807d72] dark:text-[#78756c] font-normal">/ hr</span>
            </div>
            <div className="text-[12px] text-[#5a5852] dark:text-[#a09c92] space-y-1 pt-1 border-t border-[#efeee8] dark:border-[#262520]">
              <div className="flex justify-between">
                <span>Median Annual Pay:</span>
                <span className="font-mono text-[#26251e] dark:text-[#edece6] font-medium">${Math.round(((nation.average_hourly_wage || 3.25) * 2080)).toLocaleString()}</span>
              </div>
              <div className="flex justify-between">
                <span>GDP per Capita:</span>
                <span className="font-mono text-[#26251e] dark:text-[#edece6] font-medium">${Math.round(nation.gdp_per_capita || 2432).toLocaleString()}</span>
              </div>
              <div className="flex justify-between">
                <span>Wage Growth:</span>
                <span className="font-mono text-[#1f8a65]">+3.2%</span>
              </div>
            </div>
          </div>

          <div className="p-4 rounded-[8px] bg-[#fafaf7] dark:bg-[#151412] border border-[#efeee8] dark:border-[#262520] space-y-2">
            <div className="flex items-center justify-between text-[#807d72] dark:text-[#78756c] text-[11px] font-mono uppercase">
              <span>Living Standard Index</span>
              <HeartPulse size={14} className="text-[#cf2d56]" />
            </div>
            <div className="text-[20px] font-mono font-semibold text-[#26251e] dark:text-[#edece6]">
              {(((nation.healthcare_level + nation.education_level + nation.infrastructure_level) / 3) || 50).toFixed(1)} <span className="text-[12px] text-[#807d72] dark:text-[#78756c] font-normal">/ 100</span>
            </div>
            <div className="text-[12px] text-[#5a5852] dark:text-[#a09c92] space-y-1 pt-1 border-t border-[#efeee8] dark:border-[#262520]">
              <div className="flex justify-between">
                <span>Healthcare Access:</span>
                <span className="font-mono text-[#26251e] dark:text-[#edece6] font-medium">{nation.healthcare_level.toFixed(1)}%</span>
              </div>
              <div className="flex justify-between">
                <span>Education & Skills:</span>
                <span className="font-mono text-[#26251e] dark:text-[#edece6] font-medium">{nation.education_level.toFixed(1)}%</span>
              </div>
              <div className="flex justify-between">
                <span>Public Infrastructure:</span>
                <span className="font-mono text-[#26251e] dark:text-[#edece6] font-medium">{nation.infrastructure_level.toFixed(1)}%</span>
              </div>
            </div>
          </div>

          <div className="p-4 rounded-[8px] bg-[#fafaf7] dark:bg-[#151412] border border-[#efeee8] dark:border-[#262520] space-y-2">
            <div className="flex items-center justify-between text-[#807d72] dark:text-[#78756c] text-[11px] font-mono uppercase">
              <span>Policy Impact</span>
              <Scale size={14} className="text-[#26251e] dark:text-[#edece6]" />
            </div>
            <div className="text-[20px] font-mono font-semibold text-[#1f8a65]">
              Net Positive
            </div>
            <div className="text-[11px] text-[#5a5852] dark:text-[#a09c92] space-y-1 pt-1 border-t border-[#efeee8] dark:border-[#262520] font-mono">
              <div className="flex items-center gap-1 text-[#1f8a65]">
                <Check size={12} />
                <span>Enterprise grants boosting hiring</span>
              </div>
              <div className="flex items-center gap-1 text-[#1f8a65]">
                <Check size={12} />
                <span>Port investments expanding wages</span>
              </div>
              <div className="flex items-center gap-1 text-[#807d72] dark:text-[#78756c]">
                <span>· Maintenance: ${formatB(nation.maintenance_spending_rate)}/yr</span>
              </div>
            </div>
          </div>
        </div>

        {/* Global Demographics Comparison Table */}
        <div className="pt-2">
          <div className="flex items-center justify-between mb-2">
            <h3 className="text-[14px] font-semibold text-[#26251e] dark:text-[#edece6]">
              International Demographics Comparison
            </h3>
            <span className="text-[11px] font-mono text-[#807d72] dark:text-[#78756c]">Foreign Peers & Domestic</span>
          </div>
        </div>

        <div className="overflow-x-auto -mx-6 px-6">
          <table className="w-full text-left border-collapse min-w-[850px]">
            <thead>
              <tr className="border-b border-[#e6e5e0] dark:border-[#2c2b26] bg-[#fafaf7] dark:bg-[#151412] text-[11px] font-mono uppercase text-[#807d72] dark:text-[#78756c]">
                <th className="py-3 px-3.5 font-medium">Nation</th>
                <th className="py-3 px-3.5 font-medium text-right">Population</th>
                <th className="py-3 px-3.5 font-medium text-right">Labor Force</th>
                <th className="py-3 px-3.5 font-medium text-right">Employed</th>
                <th className="py-3 px-3.5 font-medium text-right">Unemployment</th>
                <th className="py-3 px-3.5 font-medium text-right">Hourly Wage</th>
                <th className="py-3 px-3.5 font-medium text-right">GDP per Capita</th>
                <th className="py-3 px-3.5 font-medium text-right">Job Openings</th>
                <th className="py-3 px-3.5 font-medium text-right">Productivity (TFP)</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#efeee8] dark:divide-[#262520] text-[13px]">
              {/* Domestic Row */}
              <tr className="bg-[#f54e00]/5 hover:bg-[#f54e00]/10 transition-colors font-medium">
                <td className="py-3 px-3.5 flex items-center gap-2">
                  <span className="font-mono text-[11px] px-1.5 py-0.5 rounded-[4px] bg-[#f54e00] text-white font-semibold">
                    {nation.flag_code}
                  </span>
                  <div>
                    <div className="font-bold text-[#26251e] dark:text-[#edece6]">{nation.name} (Domestic)</div>
                    <div className="text-[11px] font-mono text-[#f54e00]">Our Sovereign State</div>
                  </div>
                </td>
                <td className="py-3 px-3.5 text-right font-mono text-[#26251e] dark:text-[#edece6]">
                  {((nation.population || 18_500_000) / 1e6).toFixed(1)}M
                </td>
                <td className="py-3 px-3.5 text-right font-mono text-[#5a5852] dark:text-[#a09c92]">
                  {(((nation.population || 18_500_000) * 0.475) / 1e6).toFixed(1)}M
                </td>
                <td className="py-3 px-3.5 text-right font-mono text-[#1f8a65]">
                  {(((nation.population || 18_500_000) * 0.475 * 0.947) / 1e6).toFixed(1)}M
                </td>
                <td className="py-3 px-3.5 text-right font-mono text-[#1f8a65] font-semibold">
                  5.3%
                </td>
                <td className="py-3 px-3.5 text-right font-mono text-[#26251e] dark:text-[#edece6] font-semibold">
                  ${(nation.average_hourly_wage || 3.25).toFixed(2)}
                </td>
                <td className="py-3 px-3.5 text-right font-mono text-[#26251e] dark:text-[#edece6]">
                  ${Math.round(nation.gdp_per_capita || 2432).toLocaleString()}
                </td>
                <td className="py-3 px-3.5 text-right font-mono text-[#5a5852] dark:text-[#a09c92]">
                  {Math.round(nation.job_openings || 142000).toLocaleString()}
                </td>
                <td className="py-3 px-3.5 text-right font-mono text-[#f54e00] font-semibold">
                  {(nation.productivity_index || 100.0).toFixed(1)}
                </td>
              </tr>

              {/* Foreign Countries Rows */}
              {world?.countries.map((c) => {
                const popM = (c.population / 1e6).toFixed(1);
                const laborM = c.labor_force ? (c.labor_force / 1e6).toFixed(1) : ((c.population * 0.62) / 1e6).toFixed(1);
                const empM = c.employed_workers ? (c.employed_workers / 1e6).toFixed(1) : ((c.population * 0.62 * (1 - c.unemployment)) / 1e6).toFixed(1);
                const unempPct = (c.unemployment * 100).toFixed(1);
                const wage = c.average_hourly_wage ? c.average_hourly_wage.toFixed(2) : (c.gdp_per_capita ? (c.gdp_per_capita / 2080).toFixed(2) : '28.50');
                const gdpCap = Math.round(c.gdp_per_capita || (c.gdp / c.population));
                const jobs = Math.round(c.job_openings || (c.population * 0.007));
                const tfp = (c.productivity_index || (90 + (c.geopolitical_power * 0.3))).toFixed(1);

                return (
                  <tr key={c.id} className="hover:bg-[#fafaf7] dark:hover:bg-[#1c1b18] transition-colors">
                    <td className="py-3 px-3.5 flex items-center gap-2">
                      <span className="font-mono text-[11px] px-1.5 py-0.5 rounded-[4px] bg-[#efeee8] dark:bg-[#262520] text-[#5a5852] dark:text-[#a09c92] font-semibold">
                        {c.flag_code}
                      </span>
                      <div>
                        <div className="font-semibold text-[#26251e] dark:text-[#edece6] flex items-center gap-1.5">
                          <span>{c.name}</span>
                          {(c.is_reserve_currency || c.id === 'ZTH') && (
                            <span className="px-1.5 py-0.2 rounded-[4px] bg-[#f54e00]/10 text-[#f54e00] font-mono text-[9px] font-semibold">
                              Reserve
                            </span>
                          )}
                        </div>
                        <div className="text-[11px] font-mono text-[#807d72] dark:text-[#78756c]">{c.currency} · Power: {(c.geopolitical_power || 50).toFixed(0)}</div>
                      </div>
                    </td>
                    <td className="py-3 px-3.5 text-right font-mono text-[#26251e] dark:text-[#edece6]">
                      {popM}M
                    </td>
                    <td className="py-3 px-3.5 text-right font-mono text-[#5a5852] dark:text-[#a09c92]">
                      {laborM}M
                    </td>
                    <td className="py-3 px-3.5 text-right font-mono text-[#5a5852] dark:text-[#a09c92]">
                      {empM}M
                    </td>
                    <td className="py-3 px-3.5 text-right font-mono font-medium">
                      <span className={c.unemployment > 0.06 ? 'text-[#cf2d56]' : 'text-[#1f8a65]'}>
                        {unempPct}%
                      </span>
                    </td>
                    <td className="py-3 px-3.5 text-right font-mono text-[#26251e] dark:text-[#edece6]">
                      ${wage}
                    </td>
                    <td className="py-3 px-3.5 text-right font-mono text-[#26251e] dark:text-[#edece6]">
                      ${gdpCap.toLocaleString()}
                    </td>
                    <td className="py-3 px-3.5 text-right font-mono text-[#5a5852] dark:text-[#a09c92]">
                      {jobs.toLocaleString()}
                    </td>
                    <td className="py-3 px-3.5 text-right font-mono text-[#26251e] dark:text-[#edece6]">
                      {tfp}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>

        {/* Domestic Regions breakdown */}
        <div className="pt-2">
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-[14px] font-semibold text-[#26251e] dark:text-[#edece6]">
              Domestic Provinces & Demographic Centers
            </h3>
            <span className="text-[11px] font-mono text-[#807d72] dark:text-[#78756c]">4 Administrative Regions</span>
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {(nation.regions || []).map((reg) => (
            <div
              key={reg.name}
              className="p-4 rounded-[10px] bg-[#fafaf7] dark:bg-[#151412] border border-[#efeee8] dark:border-[#262520] space-y-2.5"
            >
              <div className="flex items-center justify-between">
                <div className="font-semibold text-[14px] text-[#26251e] dark:text-[#edece6]">{reg.name}</div>
                <span className="text-[10px] font-mono px-2 py-0.5 rounded-[4px] bg-[#efeee8] dark:bg-[#262520] text-[#5a5852] dark:text-[#a09c92]">
                  {Object.keys(reg.sector_strengths || {})[0] || 'Provincial'}
                </span>
              </div>
              <div className="space-y-1.5 text-[12px] text-[#5a5852] dark:text-[#a09c92]">
                <div className="flex justify-between">
                  <span>Population:</span>
                  <span className="font-mono text-[#26251e] dark:text-[#edece6]">{(reg.population / 1e6).toFixed(2)}M</span>
                </div>
                <div className="flex justify-between">
                  <span>Gross Product:</span>
                  <span className="font-mono text-[#26251e] dark:text-[#edece6] font-medium">{formatB(reg.gdp)}</span>
                </div>
                <div className="flex justify-between">
                  <span>Annual Tax Revenue:</span>
                  <span className="font-mono text-[#26251e] dark:text-[#edece6]">{formatB(reg.tax_revenue)}</span>
                </div>
                <div className="flex justify-between">
                  <span>Regional Unemployment:</span>
                  <span className="font-mono text-[#26251e] dark:text-[#edece6]">{(reg.unemployment_rate * 100).toFixed(1)}%</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* CITIZEN & CORPORATE TAX ARCHITECTURE */}
      <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-6 space-y-5">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-3 pb-3 border-b border-[#efeee8] dark:border-[#262520]">
          <div>
            <div className="flex items-center gap-2">
              <Coins size={18} className="text-[#f54e00]" />
              <h2 className="text-[18px] font-semibold text-[#26251e] dark:text-[#edece6] tracking-tight">
                Sovereign Tax Architecture & Fiscal Schedule
              </h2>
            </div>
            <p className="text-[13px] text-[#807d72] dark:text-[#78756c] mt-0.5">
              Calibrate federal tax collection rates. Taxes directly fund public operations and prevent runaway debt, with direct trade-offs on citizen purchasing power and corporate reinvestment.
            </p>
          </div>
          <button
            onClick={handleSaveTaxes}
            disabled={taxSaving}
            className="px-4 py-2 rounded-[8px] bg-[#26251e] dark:bg-[#edece6] text-white dark:text-[#141412] text-[12px] font-medium hover:bg-[#403e33] transition-colors disabled:opacity-50 self-start md:self-auto flex items-center gap-1.5"
          >
            <Check size={14} />
            <span>{taxSaving ? 'Ratifying...' : 'Enact Tax Reforms'}</span>
          </button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
          {/* Personal Income Tax */}
          <div className="p-4 rounded-[10px] bg-[#fafaf7] dark:bg-[#151412] border border-[#efeee8] dark:border-[#262520] space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-[12px] font-semibold text-[#26251e] dark:text-[#edece6]">Personal Income Tax (PIT)</span>
              <span className="font-mono text-[14px] font-bold text-[#f54e00]">{personalTax.toFixed(0)}%</span>
            </div>
            <input
              type="range"
              min="0"
              max="50"
              step="1"
              value={personalTax}
              onChange={(e) => setPersonalTax(parseFloat(e.target.value))}
              className="w-full accent-[#f54e00] cursor-pointer"
            />
            <p className="text-[11px] text-[#807d72] dark:text-[#78756c] leading-relaxed">
              Deducted from citizen wages. Higher rates increase annual fiscal revenue, but reduce citizen disposable income and consumer sentiment.
            </p>
          </div>

          {/* Corporate Profit Tax */}
          <div className="p-4 rounded-[10px] bg-[#fafaf7] dark:bg-[#151412] border border-[#efeee8] dark:border-[#262520] space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-[12px] font-semibold text-[#26251e] dark:text-[#edece6]">Corporate Profit Tax</span>
              <span className="font-mono text-[14px] font-bold text-[#f54e00]">{corporateTax.toFixed(0)}%</span>
            </div>
            <input
              type="range"
              min="0"
              max="50"
              step="1"
              value={corporateTax}
              onChange={(e) => setCorporateTax(parseFloat(e.target.value))}
              className="w-full accent-[#f54e00] cursor-pointer"
            />
            <p className="text-[11px] text-[#807d72] dark:text-[#78756c] leading-relaxed">
              Levied on private enterprise earnings. Generates corporate tax receipts; excessively high rates lower retained capital and corporate stock valuations.
            </p>
          </div>

          {/* Sales Tax / VAT */}
          <div className="p-4 rounded-[10px] bg-[#fafaf7] dark:bg-[#151412] border border-[#efeee8] dark:border-[#262520] space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-[12px] font-semibold text-[#26251e] dark:text-[#edece6]">Sales & Consumption Tax (VAT)</span>
              <span className="font-mono text-[14px] font-bold text-[#f54e00]">{salesTax.toFixed(0)}%</span>
            </div>
            <input
              type="range"
              min="0"
              max="30"
              step="1"
              value={salesTax}
              onChange={(e) => setSalesTax(parseFloat(e.target.value))}
              className="w-full accent-[#f54e00] cursor-pointer"
            />
            <p className="text-[11px] text-[#807d72] dark:text-[#78756c] leading-relaxed">
              Applied on domestic transactions and consumer spending. Provides steady budget cash flow without dampening corporate capital expenditures.
            </p>
          </div>
        </div>
      </div>

      {/* FOREIGN SOVEREIGN DEBT CUSTODY DIRECTORATE */}
      <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-6 space-y-4">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-3 pb-3 border-b border-[#efeee8] dark:border-[#262520]">
          <div>
            <div className="flex items-center gap-2">
              <ShieldCheck size={18} className="text-[#1f8a65]" />
              <h2 className="text-[18px] font-semibold text-[#26251e] dark:text-[#edece6] tracking-tight">
                Foreign Sovereign Debt Custody & Reserve Anchor
              </h2>
            </div>
            <p className="text-[13px] text-[#807d72] dark:text-[#78756c] mt-0.5">
              Allow bigger or higher-power foreign nations to custody their sovereign debt tranches within our national depository. Holding foreign debt significantly strengthens CRN exchange rate confidence and narrows sovereign borrowing spreads.
            </p>
          </div>
          <div className="font-mono text-[12px] px-2.5 py-1 rounded-[6px] bg-[#fafaf7] dark:bg-[#151412] border border-[#e6e5e0] dark:border-[#2c2b26] text-[#1f8a65]">
            Active Custody Accords: {nation.custody_nations?.length || 0}
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3">
          {world?.countries.filter((c) => (c.geopolitical_power || 50) >= 60 || c.id === 'ZTH' || c.id === 'VAL').slice(0, 4).map((c) => {
            const isCustodied = nation.custody_nations?.includes(c.id) ?? false;
            return (
              <div
                key={c.id}
                className={`p-3.5 rounded-[10px] border transition-all flex flex-col justify-between space-y-3 ${
                  isCustodied
                    ? 'bg-[#1f8a65]/5 border-[#1f8a65]/30'
                    : 'bg-[#fafaf7] dark:bg-[#151412] border-[#efeee8] dark:border-[#262520]'
                }`}
              >
                <div>
                  <div className="flex items-center justify-between mb-1.5">
                    <span className="font-mono text-[11px] px-1.5 py-0.5 rounded-[4px] bg-[#efeee8] dark:bg-[#262520] font-semibold text-[#26251e] dark:text-[#edece6]">
                      {c.flag_code} · {c.id}
                    </span>
                    <span
                      className={`text-[10px] font-mono px-2 py-0.5 rounded-full ${
                        isCustodied ? 'bg-[#1f8a65] text-white' : 'bg-[#efeee8] dark:bg-[#262520] text-[#807d72] dark:text-[#78756c]'
                      }`}
                    >
                      {isCustodied ? 'Custody Active' : 'Available'}
                    </span>
                  </div>
                  <div className="font-semibold text-[13px] text-[#26251e] dark:text-[#edece6]">
                    {c.name}
                  </div>
                  <div className="text-[11px] text-[#807d72] dark:text-[#78756c] font-mono mt-0.5">
                    GDP: {formatB(c.gdp)} · Power: {(c.geopolitical_power || 50).toFixed(0)}
                  </div>
                </div>

                <button
                  onClick={() => handleToggleDebtCustody(c.id, !isCustodied)}
                  disabled={custodyBusy}
                  className={`w-full py-1.5 text-[11px] font-medium rounded-[6px] border transition-colors ${
                    isCustodied
                      ? 'bg-transparent text-[#cf2d56] border-[#cf2d56]/30 hover:bg-[#cf2d56]/10'
                      : 'bg-[#26251e] dark:bg-[#edece6] text-white dark:text-[#141412] border-[#26251e] dark:border-[#edece6] hover:bg-[#403e33]'
                  }`}
                >
                  {isCustodied ? 'Revoke Custody' : 'Grant Debt Custody'}
                </button>
              </div>
            );
          })}
        </div>
      </div>

      {/* FOREIGN RELATIONS & SANCTIONS */}
      <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-6 space-y-4">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-3 pb-3 border-b border-[#efeee8] dark:border-[#262520]">
          <div>
            <h2 className="text-[18px] font-semibold text-[#26251e] dark:text-[#edece6] tracking-tight flex items-center gap-2">
              <Globe2 size={20} className="text-[#f54e00]" />
              <span>Foreign Relations & Sanctions</span>
            </h2>
            <p className="text-[13px] text-[#807d72] dark:text-[#78756c] mt-0.5">
              Review foreign states, trade agreements, diplomatic stances, and sanctions. High-power nations possess greater leverage in diplomatic and economic disputes.
            </p>
          </div>
          <div className="flex items-center gap-2 bg-[#efeee8] dark:bg-[#262520] px-3 py-1.5 rounded-[8px] border border-[#e6e5e0] dark:border-[#2c2b26] self-start md:self-auto font-mono text-[12px]">
            <Scale size={14} className="text-[#26251e] dark:text-[#edece6]" />
            <span className="text-[#807d72] dark:text-[#78756c]">Geopolitical Power:</span>
            <span className="font-bold text-[#26251e] dark:text-[#edece6]">{(nation.geopolitical_power || 45.0).toFixed(1)} / 100</span>
          </div>
        </div>

        {dipMsg && (
          <div className="p-3 bg-[#1f8a65]/10 border border-[#1f8a65]/30 rounded-[8px] text-[12px] text-[#1f8a65] font-mono">
            {dipMsg}
          </div>
        )}

        <div className="overflow-x-auto -mx-6 px-6">
          <table className="w-full text-left border-collapse min-w-[850px]">
            <thead>
              <tr className="border-b border-[#e6e5e0] dark:border-[#2c2b26] bg-[#fafaf7] dark:bg-[#151412] text-[11px] font-mono uppercase text-[#807d72] dark:text-[#78756c]">
                <th className="py-3 px-4">Nation & Power</th>
                <th className="py-3 px-4">Currency</th>
                <th className="py-3 px-4">Diplomatic Stance</th>
                <th className="py-3 px-4 text-right">Foreign GDP</th>
                <th className="py-3 px-4 text-right">Tariff Rate</th>
                <th className="py-3 px-4 text-right">Bilateral Trade</th>
                <th className="py-3 px-4">Sanctions Status</th>
                <th className="py-3 px-4 text-right">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#efeee8] dark:divide-[#262520] text-[13px]">
              {world?.countries.map((c) => {
                const isHostile = c.stance === 'Hostile' || c.stance === 'Cold';
                const isAllied = c.stance === 'Allied' || c.stance === 'Friendly';
                const ourPower = nation.geopolitical_power || 45.0;
                const isStronger = (c.geopolitical_power || 50.0) > ourPower;
                const isReserve = c.is_reserve_currency || c.id === 'ZTH';

                return (
                  <tr key={c.id} className="hover:bg-[#fafaf7] dark:hover:bg-[#151412] transition-colors">
                    <td className="py-3.5 px-4 font-medium text-[#26251e] dark:text-[#edece6]">
                      <div className="flex items-center gap-2">
                        <span className="font-mono text-[11px] px-1.5 py-0.5 rounded-[4px] bg-[#efeee8] dark:bg-[#262520] text-[#5a5852] dark:text-[#a09c92] font-semibold">
                          {c.flag_code}
                        </span>
                        <div>
                          <div className="font-semibold text-[14px] flex items-center gap-1.5">
                            <span>{c.name}</span>
                            {isReserve && (
                              <span className="px-1.5 py-0.2 rounded-[4px] bg-[#f54e00]/10 text-[#f54e00] font-mono text-[10px] font-semibold">
                                Reserve Currency
                              </span>
                            )}
                          </div>
                          <div className="text-[11px] font-mono text-[#807d72] dark:text-[#78756c] flex items-center gap-1.5 mt-0.5">
                            <span>Power: {(c.geopolitical_power || 50.0).toFixed(1)}</span>
                            {isStronger && (
                              <span className="text-[#cf2d56] text-[10px]">(Greater Leverage)</span>
                            )}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="py-3.5 px-4 font-mono font-medium text-[#26251e] dark:text-[#edece6]">
                      {c.currency}
                    </td>
                    <td className="py-3.5 px-4">
                      <span
                        className={`inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium ${
                          isAllied
                            ? 'bg-[#1f8a65]/10 text-[#1f8a65]'
                            : isHostile
                            ? 'bg-[#cf2d56]/10 text-[#cf2d56]'
                            : 'bg-[#efeee8] dark:bg-[#262520] text-[#5a5852] dark:text-[#a09c92]'
                        }`}
                      >
                        {c.stance}
                      </span>
                    </td>
                    <td className="py-3.5 px-4 text-right font-mono text-[#26251e] dark:text-[#edece6]">
                      {formatB(c.gdp)}
                    </td>
                    <td className="py-3.5 px-4 text-right font-mono text-[#5a5852] dark:text-[#a09c92]">
                      5.0%
                    </td>
                    <td className="py-3.5 px-4 text-right font-mono font-medium text-[#26251e] dark:text-[#edece6]">
                      {formatB(c.trade_balance)}
                    </td>
                    <td className="py-3.5 px-4">
                      {c.sanctioned_by_us ? (
                        <span className="inline-flex items-center gap-1 text-[11px] font-mono text-[#cf2d56] bg-[#cf2d56]/10 px-2 py-0.5 rounded-[4px]">
                          <ShieldAlert size={12} />
                          <span>Sanctioned</span>
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1 text-[11px] font-mono text-[#1f8a65] bg-[#1f8a65]/10 px-2 py-0.5 rounded-[4px]">
                          <Check size={12} />
                          <span>Normal Trade</span>
                        </span>
                      )}
                    </td>
                    <td className="py-3.5 px-4 text-right">
                      <div className="flex items-center justify-end gap-1.5">
                        <button
                          onClick={() => handleToggleSanction(c.id, c.sanctioned_by_us)}
                          disabled={sanctionBusyId === c.id}
                          className={`px-2.5 py-1 text-[11px] font-mono font-medium rounded-[6px] border transition-colors flex items-center gap-1 ${
                            c.sanctioned_by_us
                              ? 'border-[#1f8a65] text-[#1f8a65] hover:bg-[#1f8a65] hover:text-white'
                              : 'border-[#cf2d56] text-[#cf2d56] hover:bg-[#cf2d56] hover:text-white'
                          }`}
                        >
                          {c.sanctioned_by_us ? (
                            <>
                              <Unlock size={11} />
                              <span>Lift Sanctions</span>
                            </>
                          ) : (
                            <>
                              <Ban size={11} />
                              <span>Sanction</span>
                            </>
                          )}
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

      {/* BILATERAL TRADE ACCORDS & CONTRACTS */}
      <div className="bg-white dark:bg-[#1c1b18] border border-[#e6e5e0] dark:border-[#2c2b26] rounded-[12px] p-6 space-y-4">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-3 pb-3 border-b border-[#efeee8] dark:border-[#262520]">
          <div>
            <div className="flex items-center gap-2">
              <ScrollText size={18} className="text-[#f54e00]" />
              <h2 className="text-[18px] font-semibold text-[#26251e] dark:text-[#edece6] tracking-tight">
                Trade Accords & Commercial Contracts
              </h2>
            </div>
            <p className="text-[13px] text-[#807d72] dark:text-[#78756c] mt-1">
              Foreign sovereign states propose trade pacts, raw material supply contracts, and freight protocols. Accords settle in ZTH global reserve currency and augment national exports.
            </p>
          </div>
          <span className="text-[11px] font-mono px-2.5 py-1 rounded-[6px] bg-[#fafaf7] dark:bg-[#151412] border border-[#e6e5e0] dark:border-[#2c2b26] text-[#5a5852] dark:text-[#a09c92] self-start md:self-auto">
            {(world?.bilateral_contracts || []).filter((c) => c.status === 'active').length} Accords Active
          </span>
        </div>

        {(!world?.bilateral_contracts || world.bilateral_contracts.length === 0) ? (
          <div className="py-8 text-center text-[#807d72] dark:text-[#78756c] text-[13px]">
            No foreign diplomatic contract proposals pending at this time.
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
            {world.bilateral_contracts.map((con) => {
              const isPending = con.status === 'pending';
              const isActive = con.status === 'active';
              return (
                <div
                  key={con.id}
                  className={`p-4 rounded-[10px] border transition-all flex flex-col justify-between ${
                    isActive
                      ? 'bg-[#1f8a65]/5 border-[#1f8a65]/30'
                      : con.status === 'declined'
                      ? 'bg-[#fafaf7] dark:bg-[#151412] border-[#efeee8] dark:border-[#262520] opacity-60'
                      : 'bg-white dark:bg-[#1c1b18] border-[#e6e5e0] dark:border-[#2c2b26] hover:border-[#26251e] dark:hover:border-[#edece6]'
                  }`}
                >
                  <div>
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-[11px] font-mono px-2 py-0.5 rounded-[4px] bg-[#efeee8] dark:bg-[#262520] text-[#26251e] dark:text-[#edece6] font-semibold">
                        {con.country_id} · {con.country_name}
                      </span>
                      <span
                        className={`text-[10px] font-mono uppercase px-2 py-0.5 rounded-full ${
                          isActive
                            ? 'bg-[#1f8a65] text-white'
                            : isPending
                            ? 'bg-[#f54e00]/15 text-[#f54e00] font-semibold'
                            : 'bg-[#efeee8] dark:bg-[#262520] text-[#807d72] dark:text-[#78756c]'
                        }`}
                      >
                        {con.status}
                      </span>
                    </div>

                    <div className="text-[12px] font-mono text-[#f54e00] font-medium mb-1">
                      {con.contract_type}
                    </div>

                    <div className="text-[14px] font-semibold text-[#26251e] dark:text-[#edece6] mb-1 leading-snug">
                      {con.title}
                    </div>

                    <p className="text-[12px] text-[#5a5852] dark:text-[#a09c92] mb-3 leading-relaxed">
                      {con.terms}
                    </p>

                    <div className="space-y-1.5 text-[12px] font-mono text-[#5a5852] dark:text-[#a09c92] bg-[#fafaf7] dark:bg-[#151412] p-2.5 rounded-[6px] border border-[#efeee8] dark:border-[#262520]">
                      <div className="flex justify-between">
                        <span>Revenue Yield:</span>
                        <span className="font-semibold text-[#1f8a65]">+{formatB(con.annual_revenue_gain)}</span>
                      </div>
                      <div className="flex justify-between">
                        <span>Export Boost:</span>
                        <span className="font-semibold text-[#26251e] dark:text-[#edece6]">+{Math.round(con.export_capacity_boost * 100)}%</span>
                      </div>
                      <div className="flex justify-between">
                        <span>Duration:</span>
                        <span>{con.duration_ticks} days</span>
                      </div>
                      <div className="text-[11px] text-[#807d72] dark:text-[#78756c] pt-1 border-t border-[#efeee8] dark:border-[#262520]">
                        Status: {con.outcome}
                      </div>
                    </div>
                  </div>

                  {isPending && (
                    <div className="grid grid-cols-2 gap-2 mt-4 pt-3 border-t border-[#efeee8] dark:border-[#262520]">
                      <button
                        onClick={() => handleRespondContractAction(con.id, 'accept')}
                        disabled={contractBusyId === con.id}
                        className="py-1.5 rounded-[6px] bg-[#1f8a65] hover:bg-[#186e50] text-white text-[11px] font-medium flex items-center justify-center gap-1 transition-colors"
                      >
                        <Check size={12} />
                        <span>Accept Accord</span>
                      </button>
                      <button
                        onClick={() => handleRespondContractAction(con.id, 'decline')}
                        disabled={contractBusyId === con.id}
                        className="py-1.5 rounded-[6px] bg-[#fafaf7] dark:bg-[#262520] hover:bg-[#efeee8] dark:hover:bg-[#33322b] text-[#5a5852] dark:text-[#a09c92] border border-[#e6e5e0] dark:border-[#2c2b26] text-[11px] font-medium flex items-center justify-center gap-1 transition-colors"
                      >
                        <X size={12} />
                        <span>Decline</span>
                      </button>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* BORROW CAPITAL MODAL */}
      {borrowModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white dark:bg-[#1c1b18] rounded-[16px] border border-[#e6e5e0] dark:border-[#2c2b26] max-w-[460px] w-full p-6 space-y-4">
            <div className="flex items-center justify-between border-b border-[#efeee8] dark:border-[#262520] pb-3">
              <div>
                <h3 className="text-[17px] font-semibold text-[#26251e] dark:text-[#edece6] flex items-center gap-2">
                  <Landmark className="w-5 h-5 text-[#f54e00]" />
                  <span>Borrow Capital</span>
                </h3>
                <div className="text-[12px] font-mono text-[#807d72] dark:text-[#78756c]">
                  Expand Treasury liquidity by issuing sovereign debt
                </div>
              </div>
              <button
                onClick={() => setBorrowModalOpen(false)}
                className="text-[#807d72] dark:text-[#78756c] hover:text-[#26251e] dark:hover:text-[#edece6] p-1 rounded-[6px] hover:bg-[#efeee8] dark:hover:bg-[#262520] transition-colors"
                aria-label="Close Modal"
              >
                <X size={18} />
              </button>
            </div>

            <div className="space-y-3 text-[13px]">
              <div className="bg-[#fafaf7] dark:bg-[#151412] p-3 rounded-[8px] space-y-1.5 font-mono text-[12px]">
                <div className="flex justify-between">
                  <span className="text-[#807d72] dark:text-[#78756c]">Treasury Cash:</span>
                  <span className="font-semibold text-[#1f8a65]">{formatB(nation.treasury_cash)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-[#807d72] dark:text-[#78756c]">National Debt:</span>
                  <span className="font-semibold text-[#cf2d56]">{formatB(nation.national_debt)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-[#807d72] dark:text-[#78756c]">Credit Rating / Borrow Yield:</span>
                  <span className="font-semibold text-[#26251e] dark:text-[#edece6]">{nation.credit_rating} ({(nation.borrowing_yield * 100).toFixed(2)}%)</span>
                </div>
              </div>

              <div>
                <label className="block text-[12px] font-medium text-[#26251e] dark:text-[#edece6] mb-1">
                  Foreign Lender Nation
                </label>
                <select
                  value={borrowLenderId}
                  onChange={(e) => setBorrowLenderId(e.target.value)}
                  className="w-full bg-[#fafaf7] dark:bg-[#151412] border border-[#e6e5e0] dark:border-[#2c2b26] focus:border-[#26251e] dark:focus:border-[#edece6] h-10 px-3 rounded-[8px] text-[13px] font-mono text-[#26251e] dark:text-[#edece6] outline-none mb-3"
                >
                  {world?.countries.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name} ({c.id}) — Rate: {(c.interest_rate * 100).toFixed(2)}% · Stance: {c.stance}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-[12px] font-medium text-[#26251e] dark:text-[#edece6] mb-1">
                  Principal Amount ({nation.currency})
                </label>
                <div className="grid grid-cols-3 gap-2 mb-2">
                  {['250000000', '500000000', '1000000000'].map((amt) => (
                    <button
                      key={amt}
                      type="button"
                      onClick={() => setBorrowAmount(amt)}
                      className={`py-1 text-[11px] font-mono rounded-[6px] border transition-colors ${
                        borrowAmount === amt
                          ? 'bg-[#26251e] dark:bg-[#edece6] text-white dark:text-[#141412] border-[#26251e] dark:border-[#edece6]'
                          : 'bg-white dark:bg-[#1c1b18] text-[#5a5852] dark:text-[#a09c92] border-[#e6e5e0] dark:border-[#2c2b26] hover:border-[#26251e]'
                      }`}
                    >
                      {formatB(parseFloat(amt))}
                    </button>
                  ))}
                </div>
                <input
                  type="number"
                  value={borrowAmount}
                  onChange={(e) => setBorrowAmount(e.target.value)}
                  className="w-full bg-[#fafaf7] dark:bg-[#151412] border border-[#e6e5e0] dark:border-[#2c2b26] focus:border-[#26251e] dark:focus:border-[#edece6] h-10 px-3 rounded-[8px] text-[14px] font-mono text-[#26251e] dark:text-[#edece6] outline-none"
                />
                <div className="text-[11px] text-[#807d72] dark:text-[#78756c] mt-1 font-mono">
                  Principal: {formatB(parseFloat(borrowAmount) || 0)} · Increases Debt & Cash
                </div>
              </div>

              <div className="p-3 bg-[#1976d2]/10 border border-[#1976d2]/20 rounded-[8px] text-[12px] text-[#1976d2] dark:text-[#63b3ed] space-y-1">
                <div className="font-semibold">Fiscal Notice:</div>
                <p className="text-[11px] leading-relaxed">
                  Borrowing immediately adds liquid funds into the Treasury to finance infrastructure, education, or loans. It increases national debt principal and annual interest carrying costs.
                </p>
              </div>
            </div>

            <div className="flex items-center justify-end gap-2 pt-2 border-t border-[#efeee8] dark:border-[#262520]">
              <button
                onClick={() => setBorrowModalOpen(false)}
                className="px-4 py-2 rounded-[8px] text-[13px] font-medium text-[#5a5852] dark:text-[#a09c92] hover:bg-[#fafaf7] dark:hover:bg-[#262520]"
              >
                Cancel
              </button>
              <button
                onClick={handleBorrowCapital}
                disabled={borrowBusy || !parseFloat(borrowAmount) || parseFloat(borrowAmount) <= 0}
                className="px-4 py-2 rounded-[8px] text-[13px] font-medium text-white bg-[#f54e00] hover:bg-[#d04200] disabled:opacity-50"
              >
                {borrowBusy ? 'Processing...' : 'Confirm Borrowing'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* REPAY DEBT MODAL */}
      {repayModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white dark:bg-[#1c1b18] rounded-[16px] border border-[#e6e5e0] dark:border-[#2c2b26] max-w-[460px] w-full p-6 space-y-4">
            <div className="flex items-center justify-between border-b border-[#efeee8] dark:border-[#262520] pb-3">
              <div>
                <h3 className="text-[17px] font-semibold text-[#26251e] dark:text-[#edece6] flex items-center gap-2">
                  <Landmark className="w-5 h-5 text-[#1f8a65]" />
                  <span>Repay National Debt</span>
                </h3>
                <div className="text-[12px] font-mono text-[#807d72] dark:text-[#78756c]">
                  Retire outstanding debt obligations using Treasury cash
                </div>
              </div>
              <button
                onClick={() => setRepayModalOpen(false)}
                className="text-[#807d72] dark:text-[#78756c] hover:text-[#26251e] dark:hover:text-[#edece6] p-1 rounded-[6px] hover:bg-[#efeee8] dark:hover:bg-[#262520] transition-colors"
                aria-label="Close Modal"
              >
                <X size={18} />
              </button>
            </div>

            <div className="space-y-3 text-[13px]">
              <div className="bg-[#fafaf7] dark:bg-[#151412] p-3 rounded-[8px] space-y-1.5 font-mono text-[12px]">
                <div className="flex justify-between">
                  <span className="text-[#807d72] dark:text-[#78756c]">Available Treasury Cash:</span>
                  <span className="font-semibold text-[#1f8a65]">{formatB(nation.treasury_cash)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-[#807d72] dark:text-[#78756c]">Current Outstanding Debt:</span>
                  <span className="font-semibold text-[#cf2d56]">{formatB(nation.national_debt)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-[#807d72] dark:text-[#78756c]">Current Borrow Yield:</span>
                  <span className="font-semibold text-[#26251e] dark:text-[#edece6]">{(nation.borrowing_yield * 100).toFixed(2)}%</span>
                </div>
              </div>

              <div>
                <label className="block text-[12px] font-medium text-[#26251e] dark:text-[#edece6] mb-1">
                  Repayment Amount ({nation.currency})
                </label>
                <div className="grid grid-cols-3 gap-2 mb-2">
                  {[
                    { label: '25% Debt', val: (nation.national_debt * 0.25).toFixed(0) },
                    { label: '50% Debt', val: (nation.national_debt * 0.5).toFixed(0) },
                    { label: '100% Debt', val: Math.min(nation.national_debt, nation.treasury_cash).toFixed(0) },
                  ].map((btn) => (
                    <button
                      key={btn.label}
                      type="button"
                      onClick={() => setRepayAmount(btn.val)}
                      className="py-1 text-[11px] font-mono rounded-[6px] border border-[#e6e5e0] dark:border-[#2c2b26] bg-white dark:bg-[#1c1b18] text-[#5a5852] dark:text-[#a09c92] hover:border-[#1f8a65] transition-colors"
                    >
                      {btn.label}
                    </button>
                  ))}
                </div>
                <input
                  type="number"
                  value={repayAmount}
                  onChange={(e) => setRepayAmount(e.target.value)}
                  className="w-full bg-[#fafaf7] dark:bg-[#151412] border border-[#e6e5e0] dark:border-[#2c2b26] focus:border-[#1f8a65] h-10 px-3 rounded-[8px] text-[14px] font-mono text-[#26251e] dark:text-[#edece6] outline-none"
                />
                <div className="text-[11px] text-[#807d72] dark:text-[#78756c] mt-1 font-mono">
                  Repayment: {formatB(parseFloat(repayAmount) || 0)} · Consumes Treasury Cash
                </div>
              </div>

              <div className="p-3 bg-[#1f8a65]/10 border border-[#1f8a65]/20 rounded-[8px] text-[12px] text-[#1f8a65] space-y-1">
                <div className="font-semibold">Economic Consequence:</div>
                <p className="text-[11px] leading-relaxed">
                  Repaying debt reduces debt-to-GDP, easing depreciation pressure on your national currency ({nation.currency}), boosting credit ratings, and lowering sovereign bond borrowing yields.
                </p>
              </div>
            </div>

            <div className="flex items-center justify-end gap-2 pt-2 border-t border-[#efeee8] dark:border-[#262520]">
              <button
                onClick={() => setRepayModalOpen(false)}
                className="px-4 py-2 rounded-[8px] text-[13px] font-medium text-[#5a5852] dark:text-[#a09c92] hover:bg-[#fafaf7] dark:hover:bg-[#262520]"
              >
                Cancel
              </button>
              <button
                onClick={handleRepayDebt}
                disabled={
                  repayBusy ||
                  !parseFloat(repayAmount) ||
                  parseFloat(repayAmount) <= 0 ||
                  parseFloat(repayAmount) > nation.treasury_cash
                }
                className="px-4 py-2 rounded-[8px] text-[13px] font-medium text-white bg-[#1f8a65] hover:bg-[#186e50] disabled:opacity-50"
              >
                {repayBusy ? 'Repaying...' : 'Confirm Debt Repayment'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* RESERVE CURRENCY (ZTH) TRADE MODAL */}
      {reserveModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white dark:bg-[#1c1b18] rounded-[16px] border border-[#e6e5e0] dark:border-[#2c2b26] max-w-[480px] w-full p-6 space-y-4">
            <div className="flex items-center justify-between border-b border-[#efeee8] dark:border-[#262520] pb-3">
              <div>
                <h3 className="text-[17px] font-semibold text-[#26251e] dark:text-[#edece6] flex items-center gap-2">
                  <Globe2 className="w-5 h-5 text-[#f54e00]" />
                  <span>Foreign FX Reserves (ZTH)</span>
                </h3>
                <div className="text-[12px] font-mono text-[#807d72] dark:text-[#78756c]">
                  Trade Global Reserve Currency (Zenthian Federation · ZTH)
                </div>
              </div>
              <button
                onClick={() => setReserveModalOpen(false)}
                className="text-[#807d72] dark:text-[#78756c] hover:text-[#26251e] dark:hover:text-[#edece6] p-1 rounded-[6px] hover:bg-[#efeee8] dark:hover:bg-[#262520] transition-colors"
                aria-label="Close Modal"
              >
                <X size={18} />
              </button>
            </div>

            {/* Side Tabs */}
            <div className="grid grid-cols-2 gap-2 bg-[#efeee8] dark:bg-[#262520] p-1 rounded-[8px]">
              <button
                type="button"
                onClick={() => setReserveSide('buy')}
                className={`py-1.5 text-[12px] font-mono font-medium rounded-[6px] transition-colors ${
                  reserveSide === 'buy'
                    ? 'bg-[#1f8a65] text-white'
                    : 'text-[#5a5852] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#edece6]'
                }`}
              >
                Buy ZTH (Spend {nation.currency})
              </button>
              <button
                type="button"
                onClick={() => setReserveSide('sell')}
                className={`py-1.5 text-[12px] font-mono font-medium rounded-[6px] transition-colors ${
                  reserveSide === 'sell'
                    ? 'bg-[#cf2d56] text-white'
                    : 'text-[#5a5852] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#edece6]'
                }`}
              >
                Sell ZTH (Receive {nation.currency})
              </button>
            </div>

            <div className="space-y-3 text-[13px]">
              <div className="bg-[#fafaf7] dark:bg-[#151412] p-3 rounded-[8px] space-y-1.5 font-mono text-[12px]">
                <div className="flex justify-between">
                  <span className="text-[#807d72] dark:text-[#78756c]">ZTH Reserve Balance:</span>
                  <span className="font-semibold text-[#f54e00]">{formatB(nation.reserve_currency_reserves || 0)} ZTH</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-[#807d72] dark:text-[#78756c]">Treasury Cash:</span>
                  <span className="font-semibold text-[#26251e] dark:text-[#edece6]">{formatB(nation.treasury_cash)} {nation.currency}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-[#807d72] dark:text-[#78756c]">Global Status:</span>
                  <span className="font-semibold text-[#1f8a65]">Preeminent World Reserve</span>
                </div>
              </div>

              <div>
                <label className="block text-[12px] font-medium text-[#26251e] dark:text-[#edece6] mb-1">
                  Order Amount ({reserveSide === 'buy' ? nation.currency : 'ZTH'})
                </label>
                <div className="grid grid-cols-4 gap-2 mb-2">
                  {['100000000', '500000000', '1000000000', '5000000000'].map((amt) => (
                    <button
                      key={amt}
                      type="button"
                      onClick={() => setReserveAmount(amt)}
                      className={`py-1 text-[11px] font-mono rounded-[6px] border transition-colors ${
                        reserveAmount === amt
                          ? 'bg-[#26251e] dark:bg-[#edece6] text-white dark:text-[#141412] border-[#26251e]'
                          : 'bg-white dark:bg-[#1c1b18] text-[#5a5852] dark:text-[#a09c92] border-[#e6e5e0] dark:border-[#2c2b26] hover:border-[#26251e]'
                      }`}
                    >
                      {formatB(parseFloat(amt))}
                    </button>
                  ))}
                </div>
                <input
                  type="number"
                  value={reserveAmount}
                  onChange={(e) => setReserveAmount(e.target.value)}
                  className="w-full bg-[#fafaf7] dark:bg-[#151412] border border-[#e6e5e0] dark:border-[#2c2b26] focus:border-[#f54e00] h-10 px-3 rounded-[8px] text-[14px] font-mono text-[#26251e] dark:text-[#edece6] outline-none"
                />
                <div className="text-[11px] text-[#807d72] dark:text-[#78756c] mt-1 font-mono">
                  Volume: {formatB(parseFloat(reserveAmount) || 0)} {reserveSide === 'buy' ? nation.currency : 'ZTH'}
                </div>
              </div>

              <div className="p-3 bg-[#fafaf7] dark:bg-[#151412] border border-[#efeee8] dark:border-[#262520] rounded-[8px] text-[12px] space-y-1">
                <div className="font-semibold text-[#26251e] dark:text-[#edece6]">Reserve Currency Role:</div>
                <p className="text-[11px] text-[#5a5852] dark:text-[#a09c92] leading-relaxed">
                  ZTH is issued by the Zenthian Federation, the world’s most powerful economy. Accumulating ZTH reserves provides hard liquidity for international trade settlements, enhances foreign debt collateral, and elevates sovereign creditworthiness.
                </p>
              </div>
            </div>

            <div className="flex items-center justify-end gap-2 pt-2 border-t border-[#efeee8] dark:border-[#262520]">
              <button
                onClick={() => setReserveModalOpen(false)}
                className="px-4 py-2 rounded-[8px] text-[13px] font-medium text-[#5a5852] dark:text-[#a09c92] hover:bg-[#fafaf7] dark:hover:bg-[#262520]"
              >
                Cancel
              </button>
              <button
                onClick={handleTradeReserveCurrency}
                disabled={reserveBusy || !parseFloat(reserveAmount) || parseFloat(reserveAmount) <= 0}
                className={`px-4 py-2 rounded-[8px] text-[13px] font-medium text-white transition-colors disabled:opacity-50 ${
                  reserveSide === 'buy' ? 'bg-[#1f8a65] hover:bg-[#186e50]' : 'bg-[#cf2d56] hover:bg-[#a82244]'
                }`}
              >
                {reserveBusy ? 'Executing...' : `Execute ${reserveSide.toUpperCase()}`}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* DEDICATED BILATERAL LOANS MODAL */}
      {loanModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white dark:bg-[#1c1b18] rounded-[16px] border border-[#e6e5e0] dark:border-[#2c2b26] max-w-[760px] w-full p-6 space-y-5">
            <div className="flex items-center justify-between border-b border-[#efeee8] dark:border-[#262520] pb-3">
              <div>
                <h3 className="text-[18px] font-semibold text-[#26251e] dark:text-[#edece6] flex items-center gap-2">
                  <HandCoins className="w-5 h-5 text-[#f54e00]" />
                  <span>Bilateral Loans Management</span>
                </h3>
                <div className="text-[12px] font-mono text-[#807d72] dark:text-[#78756c]">
                  Manage Foreign Capital Petitions & Track Approved Debt Service Yields
                </div>
              </div>
              <button
                onClick={() => setLoanModalOpen(false)}
                className="text-[#807d72] dark:text-[#78756c] hover:text-[#26251e] dark:hover:text-[#edece6] p-1 rounded-[6px] hover:bg-[#efeee8] dark:hover:bg-[#262520] transition-colors"
              >
                <X size={18} />
              </button>
            </div>

            {/* Modal Tabs */}
            <div className="flex items-center gap-2 bg-[#efeee8] dark:bg-[#262520] p-1 rounded-[8px]">
              <button
                type="button"
                onClick={() => setLoanModalTab('pending')}
                className={`flex-1 py-1.5 text-[13px] font-medium rounded-[6px] transition-colors ${
                  loanModalTab === 'pending'
                    ? 'bg-white dark:bg-[#1c1b18] text-[#26251e] dark:text-[#edece6]'
                    : 'text-[#5a5852] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#edece6]'
                }`}
              >
                Pending Requests ({loanRequests.filter((r) => r.status === 'pending').length})
              </button>
              <button
                type="button"
                onClick={() => setLoanModalTab('approved')}
                className={`flex-1 py-1.5 text-[13px] font-medium rounded-[6px] transition-colors ${
                  loanModalTab === 'approved'
                    ? 'bg-white dark:bg-[#1c1b18] text-[#26251e] dark:text-[#edece6]'
                    : 'text-[#5a5852] dark:text-[#a09c92] hover:text-[#26251e] dark:hover:text-[#edece6]'
                }`}
              >
                Approved Loans ({loanRequests.filter((r) => r.status === 'accepted' || r.status === 'matured_repaid').length})
              </button>
            </div>

            {/* Modal Tab Content */}
            {loanModalTab === 'pending' ? (
              <div className="space-y-3 max-h-[420px] overflow-y-auto pr-1">
                {loanRequests.filter((r) => r.status === 'pending').length === 0 ? (
                  <div className="py-12 text-center text-[#807d72] dark:text-[#78756c] text-[13px] font-mono bg-[#fafaf7] dark:bg-[#151412] rounded-[8px] border border-dashed border-[#e6e5e0] dark:border-[#2c2b26]">
                    No incoming loan petitions awaiting review. Foreign nations dispatch requests periodically.
                  </div>
                ) : (
                  loanRequests.filter((r) => r.status === 'pending').map((req) => (
                    <div key={req.id} className="p-4 rounded-[10px] border border-[#e6e5e0] dark:border-[#2c2b26] bg-[#fafaf7] dark:bg-[#151412] space-y-3">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <span className="text-[11px] font-mono px-2 py-0.5 rounded-[4px] bg-[#efeee8] dark:bg-[#262520] font-semibold text-[#26251e] dark:text-[#edece6]">
                            {req.country_id} · {req.country_name}
                          </span>
                          <span className="text-[12px] font-mono text-[#807d72] dark:text-[#78756c]">ID: {req.id}</span>
                        </div>
                        <span className="text-[11px] font-mono text-[#f54e00] font-semibold bg-[#f54e00]/10 px-2 py-0.5 rounded-[4px]">
                          Pending Review
                        </span>
                      </div>

                      <div className="text-[14px] font-medium text-[#26251e] dark:text-[#edece6]">
                        {req.purpose}
                      </div>

                      <div className="grid grid-cols-3 gap-2 bg-white dark:bg-[#1c1b18] p-2.5 rounded-[6px] border border-[#efeee8] dark:border-[#262520] text-[12px] font-mono">
                        <div>
                          <div className="text-[#807d72] dark:text-[#78756c] text-[10px] uppercase">Principal</div>
                          <div className="font-semibold text-[#26251e] dark:text-[#edece6]">{formatB(req.amount)}</div>
                        </div>
                        <div>
                          <div className="text-[#807d72] dark:text-[#78756c] text-[10px] uppercase">Interest Yield</div>
                          <div className="font-semibold text-[#1f8a65]">{(req.interest_rate * 100).toFixed(2)}% APR</div>
                        </div>
                        <div>
                          <div className="text-[#807d72] dark:text-[#78756c] text-[10px] uppercase">Term</div>
                          <div className="text-[#26251e] dark:text-[#edece6]">{req.term_ticks} Days</div>
                        </div>
                      </div>

                      <div className="flex gap-2 pt-1">
                        <button
                          onClick={() => handleRespondLoanAction(req.id, 'accept')}
                          disabled={loanBusyId === req.id || nation.treasury_cash < req.amount}
                          className="flex-1 py-2 rounded-[6px] bg-[#1f8a65] hover:bg-[#186e50] text-white text-[12px] font-medium flex items-center justify-center gap-1.5 transition-colors disabled:opacity-40"
                        >
                          <Check size={14} />
                          <span>Approve & Disburse ({formatB(req.amount)})</span>
                        </button>
                        <button
                          onClick={() => handleRespondLoanAction(req.id, 'decline')}
                          disabled={loanBusyId === req.id}
                          className="px-4 py-2 rounded-[6px] bg-white dark:bg-[#1c1b18] hover:bg-[#efeee8] dark:hover:bg-[#262520] text-[#5a5852] dark:text-[#a09c92] border border-[#e6e5e0] dark:border-[#2c2b26] text-[12px] font-medium flex items-center justify-center gap-1.5 transition-colors"
                        >
                          <X size={14} />
                          <span>Decline</span>
                        </button>
                      </div>
                    </div>
                  ))
                )}
              </div>
            ) : (
              <div className="space-y-3 max-h-[420px] overflow-y-auto pr-1">
                {loanRequests.filter((r) => r.status === 'accepted' || r.status === 'matured_repaid').length === 0 ? (
                  <div className="py-12 text-center text-[#807d72] dark:text-[#78756c] text-[13px] font-mono bg-[#fafaf7] dark:bg-[#151412] rounded-[8px] border border-dashed border-[#e6e5e0] dark:border-[#2c2b26]">
                    No active or approved loans. Approve incoming bilateral requests to generate interest income for the treasury.
                  </div>
                ) : (
                  loanRequests.filter((r) => r.status === 'accepted' || r.status === 'matured_repaid').map((req) => {
                    const isMatured = req.status === 'matured_repaid';
                    const ticksLeft = req.ticks_remaining !== undefined ? req.ticks_remaining : 0;
                    const progress = req.term_ticks > 0 ? Math.max(0, Math.min(100, Math.round(((req.term_ticks - ticksLeft) / req.term_ticks) * 100))) : 100;
                    const interest = req.interest_earned || 0;

                    return (
                      <div key={req.id} className="p-4 rounded-[10px] border border-[#e6e5e0] dark:border-[#2c2b26] bg-white dark:bg-[#1c1b18] space-y-3">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <span className="text-[11px] font-mono px-2 py-0.5 rounded-[4px] bg-[#1f8a65]/15 text-[#1f8a65] font-semibold">
                              {req.country_id} · {req.country_name}
                            </span>
                            <span className="text-[12px] font-mono text-[#807d72] dark:text-[#78756c]">ID: {req.id}</span>
                          </div>
                          <span
                            className={`text-[11px] font-mono uppercase px-2 py-0.5 rounded-full ${
                              isMatured
                                ? 'bg-[#26251e] dark:bg-[#edece6] text-white dark:text-[#141412]'
                                : 'bg-[#1f8a65] text-white font-medium'
                            }`}
                          >
                            {isMatured ? 'Matured & Repaid' : 'Active Servicing'}
                          </span>
                        </div>

                        <div className="text-[14px] font-medium text-[#26251e] dark:text-[#edece6]">
                          {req.purpose}
                        </div>

                        {/* Progress Bar */}
                        <div>
                          <div className="flex justify-between text-[11px] font-mono text-[#807d72] dark:text-[#78756c] mb-1">
                            <span>Repayment Progress: {progress}%</span>
                            <span>{isMatured ? '0 days remaining' : `${ticksLeft} days remaining`}</span>
                          </div>
                          <div className="w-full bg-[#efeee8] dark:bg-[#262520] h-2 rounded-full overflow-hidden">
                            <div
                              className={`h-full rounded-full transition-all ${isMatured ? 'bg-[#26251e] dark:bg-[#edece6]' : 'bg-[#1f8a65]'}`}
                              style={{ width: `${progress}%` }}
                            />
                          </div>
                        </div>

                        {/* Financial Telemetry Strip */}
                        <div className="grid grid-cols-3 gap-2 bg-[#fafaf7] dark:bg-[#151412] p-2.5 rounded-[6px] border border-[#efeee8] dark:border-[#262520] text-[12px] font-mono">
                          <div>
                            <div className="text-[#807d72] dark:text-[#78756c] text-[10px] uppercase">Principal Disbursed</div>
                            <div className="font-semibold text-[#26251e] dark:text-[#edece6]">{formatB(req.amount)}</div>
                          </div>
                          <div>
                            <div className="text-[#807d72] dark:text-[#78756c] text-[10px] uppercase">Coupon Yield (APR)</div>
                            <div className="font-semibold text-[#1f8a65]">{(req.interest_rate * 100).toFixed(2)}%</div>
                          </div>
                          <div>
                            <div className="text-[#807d72] dark:text-[#78756c] text-[10px] uppercase">Interest Earned</div>
                            <div className="font-semibold text-[#1f8a65]">+{formatB(interest)}</div>
                          </div>
                        </div>

                        <div className="text-[11px] font-mono text-[#807d72] dark:text-[#78756c] pt-1 border-t border-[#efeee8] dark:border-[#262520] flex justify-between">
                          <span>Outcome: {req.outcome_message || 'Servicing coupon payments into treasury'}</span>
                          <span className="text-[#1f8a65] font-semibold">Alliance Stance: Elevated</span>
                        </div>
                      </div>
                    );
                  })
                )}
              </div>
            )}

            <div className="flex items-center justify-end pt-2 border-t border-[#efeee8] dark:border-[#262520]">
              <button
                onClick={() => setLoanModalOpen(false)}
                className="px-4 py-2 rounded-[8px] text-[13px] font-medium text-[#26251e] dark:text-[#edece6] bg-[#fafaf7] dark:bg-[#262520] hover:bg-[#efeee8] dark:hover:bg-[#33322b] border border-[#e6e5e0] dark:border-[#2c2b26]"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
