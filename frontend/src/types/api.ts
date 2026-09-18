export interface PriceItem {
  symbol: string;
  name: string;
  sector: string;
  cap_tier: string;
  true_value: number;
  reported_value: number;
  mid: number;
  bid: number;
  ask: number;
  spread: number;
  annual_revenue: number;
  net_margin: number;
  sector_multiple: number;
  market_cap: number;
  reported_market_cap: number;
  shares_outstanding: number;
  float_shares: number;
  pe_ratio: number;
  ps_ratio: number;
  is_ipo: boolean;
  ipo_tick: number;
  ipo_price: number;
  is_private?: boolean;
  stage?: string;
  private_valuation?: number;
}

export interface TradeRecord {
  id: number;
  tick: number;
  timestamp: number;
  symbol: string;
  taker_id: string;
  maker_id: string;
  side: string;
  price: number;
  quantity: number;
}

export interface CandleRecord {
  symbol: string;
  timeframe: string;
  start_time?: number;
  timestamp?: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface BookLevel {
  price: number;
  quantity: number;
  orders_count?: number;
}

export interface OrderBookDTO {
  symbol: string;
  bids: BookLevel[];
  asks: BookLevel[];
  mid: number;
  spread: number;
  has_mid?: boolean;
}

export interface PositionDTO {
  symbol: string;
  side: string;
  quantity: number;
  entry_price: number;
  current_price: number;
  value: number;
  unrealized_pnl: number;
  pnl_percent: number;
}

export interface PortfolioDTO {
  cash: number;
  equity: number;
  buying_power: number;
  margin_ratio: number;
  maintenance_margin_ratio: number;
  max_leverage: number;
  positions: PositionDTO[];
}

export interface PolicyItemDTO {
  id: string;
  category: string;
  name: string;
  description: string;
  favored_sector: string;
  annual_cost: number;
  active: boolean;
  min_tier: string;
  is_locked: boolean;
  rollout_days?: number;
  days_active?: number;
  efficacy?: number;
  rollout_days_total?: number;
  rollout_progress_pct?: number;
  direct_positive?: string;
  direct_negative?: string;
  indirect_positive?: string;
  indirect_negative?: string;
}

export interface SovereignBondDTO {
  id: string;
  name: string;
  maturity_years: number;
  coupon_rate: number;
  yield_to_maturity: number;
  price: number;
  outstanding_amount: number;
}

export interface ForexPairDTO {
  symbol: string;
  base_currency: string;
  quote_currency: string;
  rate: number;
  change_24h: number;
  high_24h: number;
  low_24h: number;
  base_rate: number;
}

export interface WorldStockDTO {
  ticker: string;
  name: string;
  country_id: string;
  country_name: string;
  sector: string;
  price: number;
  change_pct: number;
  market_cap: number;
  pe_ratio: number;
  volume: number;
}

export interface ForeignLoanRequestDTO {
  id: string;
  country_id: string;
  country_name: string;
  currency: string;
  amount: number;
  interest_rate: number;
  term_ticks: number;
  purpose: string;
  status: 'pending' | 'accepted' | 'declined' | 'matured_repaid';
  ticks_remaining?: number;
  interest_earned?: number;
  outcome_message?: string;
}

export interface MacroDTO {
  tick: number;
  calendar_year?: number;
  calendar_month?: number;
  calendar_day?: number;
  calendar_formatted?: string;
  sim_day?: number;
  sim_year?: number;
  sim_date_formatted?: string;
  gdp: number;
  real_gdp_growth: number;
  central_bank_rate: number;
  ten_year_yield: number;
  inflation_rate: number;
  policy_stance: string;
  unemployment_rate: number;
  labor_force: number;
  employed_workers: number;
  average_hourly_wage: number;
  consumer_sentiment: number;
  consumer_spending: number;
  national_debt: number;
  federal_revenue: number;
  federal_outlays: number;
  budget_deficit: number;
  credit_rating: string;
  treasury_cash: number;
  borrowing_yield: number;
  dxy_index: number;
  trade_balance: number;
  maintenance_spending_rate: number;
  depreciation_rate: number;
  tourism_revenue: number;
  productivity_index: number;
  bonds: SovereignBondDTO[];
  tier: string;
  tier_name: string;
  tier_progress: number;
  policies: PolicyItemDTO[];
}

export interface SimStatusDTO {
  status: string;
  tick: number;
  uptime: string;
  total_trades: number;
  total_volume: number;
  companies_count: number;
  speed_multiplier: number;
}

export interface EventRecord {
  id: number;
  tick: number;
  timestamp: number;
  kind: string;
  symbol: string;
  details: string;
}

export interface RegionDTO {
  name: string;
  gdp: number;
  tax_revenue: number;
  unemployment_rate: number;
  population: number;
  budget: number;
  debt: number;
  sector_strengths: Record<string, number>;
}

export interface NationDTO {
  name: string;
  currency: string;
  flag_code: string;
  founded: number;
  configured: boolean;
  calendar_formatted?: string;
  sim_day?: number;
  sim_year?: number;
  sim_date_formatted?: string;
  infrastructure_level: number;
  healthcare_level: number;
  education_level: number;
  enterprise_grants_level: number;
  export_capacity_level: number;
  exchange_chartered: boolean;
  treasury_cash: number;
  reserve_currency_reserves?: number;
  national_debt: number;
  personal_income_tax_rate?: number;
  corporate_tax_rate?: number;
  sales_tax_rate?: number;
  foreign_debt_custody_enabled?: boolean;
  custody_nations?: string[];
  credit_rating: string;
  borrowing_yield: number;
  maintenance_spending_rate: number;
  depreciation_rate: number;
  tourism_revenue: number;
  productivity_index: number;
  population: number;
  labor_force?: number;
  employed_workers?: number;
  average_hourly_wage?: number;
  gdp_per_capita: number;
  job_openings: number;
  geopolitical_power: number;
  bonds: SovereignBondDTO[];
  loan_requests: ForeignLoanRequestDTO[];
  tier: string;
  tier_name: string;
  tier_progress: number;
  regions: RegionDTO[];
}

export interface ForeignCountryDTO {
  id: string;
  name: string;
  currency: string;
  flag_code: string;
  gdp: number;
  gdp_growth: number;
  inflation: number;
  interest_rate: number;
  unemployment: number;
  population: number;
  labor_force: number;
  employed_workers: number;
  gdp_per_capita: number;
  average_hourly_wage: number;
  job_openings: number;
  productivity_index: number;
  geopolitical_power: number;
  sanctioned_by_us: boolean;
  sanctioned_us: boolean;
  sanctions_summary: string;
  is_reserve_currency?: boolean;
  trade_balance: number;
  tariff_rate: number;
  stance: string;
  trade_volume: number;
  sanctions_level: number;
  aid_flow: number;
}

export interface CountryIndexDTO {
  symbol: string;
  name: string;
  country_id: string;
  country_name?: string;
  price: number;
  change_24h: number;
  high?: number;
  low?: number;
  high_24h?: number;
  low_24h?: number;
  volume: number;
  constituents_count: number;
}

export interface EditorialNewsDTO {
  id: number;
  tick: number;
  timestamp: number;
  wire: string;
  category: string;
  urgency: 'BREAKING' | 'ALERT' | 'DEVELOPING' | 'ROUTINE';
  sentiment: 'BULLISH' | 'BEARISH' | 'NEUTRAL';
  headline: string;
  summary: string;
  affected_tickers: string[];
}

export interface BorrowCapitalRequest {
  amount: number;
}

export interface RepayDebtRequest {
  amount: number;
}

export interface TradeReserveCurrencyRequest {
  side: 'buy' | 'sell';
  amount: number;
}

export interface SanctionToggleRequest {
  country_id: string;
  action: 'sanction' | 'lift';
}

export interface ForeignPolicyAction {
  tick: number;
  country_id: string;
  action: string;
  detail: string;
}

export interface BilateralContractDTO {
  id: string;
  country_id: string;
  country_name: string;
  contract_type: string;
  title: string;
  terms: string;
  annual_revenue_gain: number;
  export_capacity_boost: number;
  duration_ticks: number;
  status: 'pending' | 'active' | 'declined' | 'expired';
  outcome: string;
}

export interface WorldDTO {
  countries: ForeignCountryDTO[];
  forex_pairs: ForexPairDTO[];
  world_stocks: WorldStockDTO[];
  country_indices: CountryIndexDTO[];
  loan_requests: ForeignLoanRequestDTO[];
  bilateral_contracts?: BilateralContractDTO[];
  policy_log: ForeignPolicyAction[];
  reserve_currency?: string;
  reserve_country_name?: string;
  reserve_currency_reserves?: number;
}
