import {
  CandleRecord,
  CountryIndexDTO,
  EditorialNewsDTO,
  EventRecord,
  ForexPairDTO,
  MacroDTO,
  NationDTO,
  OrderBookDTO,
  PortfolioDTO,
  PriceItem,
  SimStatusDTO,
  SovereignBondDTO,
  TradeRecord,
  WorldDTO,
  WorldStockDTO,
} from '../types/api';

const BASE_URL = '';

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const errorBody = await res.text();
    throw new Error(`API Error ${res.status}: ${errorBody || res.statusText}`);
  }
  return res.json();
}

export const api = {
  async getStatus(): Promise<SimStatusDTO> {
    const res = await fetch(`${BASE_URL}/api/status`);
    return handleResponse<SimStatusDTO>(res);
  },

  async getPrices(): Promise<PriceItem[]> {
    const res = await fetch(`${BASE_URL}/api/prices`);
    return handleResponse<PriceItem[]>(res);
  },

  async getTrades(symbol?: string, limit = 50): Promise<TradeRecord[]> {
    const params = new URLSearchParams();
    if (symbol) params.append('symbol', symbol);
    params.append('limit', limit.toString());
    const res = await fetch(`${BASE_URL}/api/trades?${params.toString()}`);
    return handleResponse<TradeRecord[]>(res);
  },

  async getCandles(symbol: string, timeframe = '1D'): Promise<CandleRecord[]> {
    const params = new URLSearchParams({ symbol, tf: timeframe });
    const res = await fetch(`${BASE_URL}/api/candles?${params.toString()}`);
    return handleResponse<CandleRecord[]>(res);
  },

  async getOrderBook(symbol: string): Promise<OrderBookDTO> {
    const res = await fetch(`${BASE_URL}/api/book?symbol=${encodeURIComponent(symbol)}`);
    return handleResponse<OrderBookDTO>(res);
  },

  async getPortfolio(): Promise<PortfolioDTO> {
    const res = await fetch(`${BASE_URL}/api/portfolio`);
    return handleResponse<PortfolioDTO>(res);
  },

  async submitOrder(order: {
    symbol: string;
    side: 'Buy' | 'Sell';
    order_type: 'Market' | 'Limit';
    price: number;
    quantity: number;
  }): Promise<{ status: string; order_id: number }> {
    const res = await fetch(`${BASE_URL}/api/orders`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(order),
    });
    return handleResponse<{ status: string; order_id: number }>(res);
  },

  async closePosition(symbol: string): Promise<{ status: string; symbol: string; quantity: number }> {
    const res = await fetch(`${BASE_URL}/api/portfolio/close`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ symbol }),
    });
    return handleResponse<{ status: string; symbol: string; quantity: number }>(res);
  },

  async getMacro(): Promise<MacroDTO> {
    const res = await fetch(`${BASE_URL}/api/macro`);
    return handleResponse<MacroDTO>(res);
  },

  async togglePolicy(policy_id: string): Promise<{ policy_id: string; name: string; active: boolean }> {
    const res = await fetch(`${BASE_URL}/api/policy/toggle`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ policy_id }),
    });
    return handleResponse<{ policy_id: string; name: string; active: boolean }>(res);
  },

  async getNation(): Promise<NationDTO> {
    const res = await fetch(`${BASE_URL}/api/nation`);
    return handleResponse<NationDTO>(res);
  },

  async foundNation(data: { name: string; currency: string; flag_code: string; founded: number }): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/nation/found`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  async investNation(data: { pillar: string; amount: number }): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/nation/invest`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  async charterExchange(): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/nation/charter`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    });
    return handleResponse(res);
  },

  async getWorld(): Promise<WorldDTO> {
    const res = await fetch(`${BASE_URL}/api/nation/world`);
    return handleResponse<WorldDTO>(res);
  },

  async sendDiplomacy(data: {
    action: string;
    country_id: string;
    amount?: number;
    level?: number;
    tariff?: number;
  }): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/nation/diplomacy`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  async getEvents(limit = 100): Promise<EditorialNewsDTO[]> {
    const res = await fetch(`${BASE_URL}/api/events?limit=${limit}`);
    return handleResponse<EditorialNewsDTO[]>(res);
  },

  async borrowCapital(amount: number): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/nation/borrow`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ amount }),
    });
    return handleResponse(res);
  },

  async repayDebt(amount: number): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/nation/debt/repay`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ amount }),
    });
    return handleResponse(res);
  },

  async tradeReserveCurrency(data: { side: 'buy' | 'sell'; amount: number }): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/nation/reserve-currency/trade`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  async toggleSanctions(country_id: string, action: 'sanction' | 'lift'): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/nation/sanctions/toggle`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ country_id, action }),
    });
    return handleResponse(res);
  },

  async getCountryIndices(): Promise<CountryIndexDTO[]> {
    const res = await fetch(`${BASE_URL}/api/world/indices`);
    return handleResponse<CountryIndexDTO[]>(res);
  },

  async setSimSpeed(multiplier: number): Promise<{ multiplier: number }> {
    const res = await fetch(`${BASE_URL}/api/sim/speed`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ multiplier }),
    });
    return handleResponse<{ multiplier: number }>(res);
  },

  async submitIpo(data: { sector?: string; cap_tier?: string; revenue?: number }): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/sim/ipo`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  async getBonds(): Promise<SovereignBondDTO[]> {
    const res = await fetch(`${BASE_URL}/api/bonds`);
    return handleResponse<SovereignBondDTO[]>(res);
  },

  async issueBond(data: { maturity_years: number; amount: number }): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/bonds/issue`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  async buybackBond(data: { maturity_years: number; amount: number }): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/bonds/buyback`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  async getForex(): Promise<ForexPairDTO[]> {
    const res = await fetch(`${BASE_URL}/api/forex`);
    return handleResponse<ForexPairDTO[]>(res);
  },

  async swapForex(data: { pair: string; amount: number; side: 'buy' | 'sell' }): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/forex/swap`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  async getWorldStocks(): Promise<WorldStockDTO[]> {
    const res = await fetch(`${BASE_URL}/api/world/stocks`);
    return handleResponse<WorldStockDTO[]>(res);
  },

  async setMaintenance(rate: number): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/nation/maintenance`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ rate }),
    });
    return handleResponse(res);
  },

  async respondLoan(data: { request_id: string; action: 'accept' | 'decline' }): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/nation/loan/respond`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  async respondContract(data: { contract_id: string; action: 'accept' | 'decline' }): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/nation/contract/respond`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  async updatePortfolioSettings(data: { balance?: number; leverage?: number }): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/portfolio/settings`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  async setTaxes(data: { personal_income_tax_rate: number; corporate_tax_rate: number; sales_tax_rate: number }): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/nation/taxes`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  async borrowBilateral(data: { lender_id: string; amount: number; interest_rate: number }): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/nation/borrow/bilateral`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  async toggleDebtCustody(data: { nation_id: string; enable: boolean }): Promise<any> {
    const res = await fetch(`${BASE_URL}/api/nation/debt-custody/toggle`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },
};
