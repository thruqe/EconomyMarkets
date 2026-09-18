use std::collections::HashMap;
use std::sync::Arc;
use parking_lot::RwLock;
use rand::rngs::StdRng;
use rand::SeedableRng;

use crate::agent::{Bank, HedgeFund, MarketMaker};
use crate::citizen::{CitizenEconomy, CitizenReport};
use crate::company::Company;
use crate::country::{all_known_policies, NationalEconomy, NationalReport, POLICY_CITIZEN_STIMULUS};
use crate::market::{Account, Order, OrderBook, OrderSource, Side};
use super::company_market::CompanyMarket;
use super::events::{Event, EventKind};

pub trait AccountHolder: OrderSource {
    fn account(&self) -> Arc<RwLock<Account>>;
}

impl AccountHolder for MarketMaker {
    fn account(&self) -> Arc<RwLock<Account>> {
        self.account.clone()
    }
}

impl AccountHolder for HedgeFund {
    fn account(&self) -> Arc<RwLock<Account>> {
        self.account.clone()
    }
}

impl AccountHolder for Bank {
    fn account(&self) -> Arc<RwLock<Account>> {
        self.account.clone()
    }
}

impl AccountHolder for crate::retail::SimulatedRetailTrader {
    fn account(&self) -> Arc<RwLock<Account>> {
        self.account.clone()
    }
}

impl AccountHolder for crate::retail::HumanTrader {
    fn account(&self) -> Arc<RwLock<Account>> {
        self.account.clone()
    }
}

pub struct Simulation {
    pub markets: HashMap<String, CompanyMarket>,
    pub order: Vec<String>,
    pub holders: Vec<Arc<RwLock<dyn AccountHolder>>>,
    pub banks: Vec<Arc<RwLock<Bank>>>,

    pub depth_levels: usize,
    pub max_history: usize,
    pub tick: usize,

    pub event_log: Vec<Event>,
    pub rng: StdRng,

    pub citizen: CitizenEconomy,
    pub national: NationalEconomy,
    pub citizen_report: CitizenReport,
    pub national_report: NationalReport,
    pub private_companies: Vec<Company>,
}

impl Simulation {
    pub fn new(depth_levels: usize, max_history: usize) -> Self {
        let mut cit = CitizenEconomy::new();
        let mut nat = NationalEconomy::new();

        let cit_rep = cit.tick(
            nat.labor.employed_workers,
            nat.labor.average_hourly_wage,
            nat.central_bank.cpi_inflation_rate,
            nat.labor.annual_wage_growth,
            0.0,
            1.0 / 252.0,
        );
        let nat_rep = nat.tick(
            cit_rep.consumer_spending,
            3_000_000_000.0,
            4_000_000.0,
            250_000_000.0,
            350_000_000.0,
            1.0 / 252.0,
        );

        Self {
            markets: HashMap::new(),
            order: Vec::new(),
            holders: Vec::new(),
            banks: Vec::new(),
            depth_levels,
            max_history,
            tick: 0,
            event_log: Vec::new(),
            rng: StdRng::seed_from_u64(1337),
            citizen: cit,
            national: nat,
            citizen_report: cit_rep,
            national_report: nat_rep,
            private_companies: Vec::new(),
        }
    }

    pub fn add_company(&mut self, co: Company) {
        if self.markets.contains_key(&co.symbol) {
            return;
        }
        let symbol = co.symbol.clone();
        self.markets.insert(symbol.clone(), CompanyMarket::new(co, self.max_history));
        self.order.push(symbol);
    }

    pub fn add_private_company(&mut self, co: Company) {
        self.private_companies.push(co);
    }

    pub fn add_participant<P: AccountHolder + 'static>(
        &mut self,
        p: Arc<RwLock<P>>,
        symbols: &[&str],
    ) {
        for &sym in symbols {
            let cm = self.markets.get_mut(sym).expect("sim: AddParticipant called with unregistered symbol");
            cm.participants.push(p.clone());
        }
        self.holders.push(p);
    }

    pub fn add_bank(&mut self, bank: Arc<RwLock<Bank>>, symbols: &[&str]) {
        for &sym in symbols {
            let cm = self.markets.get_mut(sym).expect("sim: AddParticipant called with unregistered symbol");
            cm.participants.push(bank.clone());
        }
        self.holders.push(bank.clone());
        self.banks.push(bank);
    }

    pub fn tick(&self) -> usize {
        self.tick
    }

    pub fn set_tick(&mut self, t: usize) {
        self.tick = t;
    }

    pub fn company(&self, symbol: &str) -> Option<&Company> {
        self.markets.get(symbol).map(|cm| &cm.co)
    }

    pub fn company_mut(&mut self, symbol: &str) -> Option<&mut Company> {
        self.markets.get_mut(symbol).map(|cm| &mut cm.co)
    }

    pub fn book(&self, symbol: &str) -> Option<&OrderBook> {
        self.markets.get(symbol).map(|cm| &cm.book)
    }

    pub fn book_mut(&mut self, symbol: &str) -> Option<&mut OrderBook> {
        self.markets.get_mut(symbol).map(|cm| &mut cm.book)
    }

    pub fn symbols(&self) -> Vec<String> {
        self.order.clone()
    }

    pub fn clear_event_log(&mut self) {
        self.event_log.clear();
    }

    pub fn list_ipo(&mut self, co: Company, underwriter_capital: f64) {
        if self.markets.contains_key(&co.symbol) {
            return;
        }

        let symbol = co.symbol.clone();
        let capital = if underwriter_capital <= 0.0 { 25_000_000.0 } else { underwriter_capital };

        self.add_company(co.clone());

        let mut mm = MarketMaker::new(format!("underwriter_{}", symbol), capital, 10.0, 0.08);
        mm.max_inventory = 500_000.0;
        mm.quote_size = 250.0;
        mm.ladder_levels = 5;
        mm.set_company(co.clone());
        let mm_arc = Arc::new(RwLock::new(mm));
        self.add_participant(mm_arc.clone(), &[&symbol]);

        let hf_acct = Arc::new(RwLock::new(Account::new(
            format!("hf_ipo_{}", symbol),
            15_000_000.0,
            5.0,
            0.10,
        )));
        let mut hf = HedgeFund::new(format!("hf_ipo_{}", symbol), hf_acct, co.clone());
        hf.min_trade_threshold = 0.003;
        hf.full_conviction_threshold = 0.025;
        hf.rebalance_threshold = 0.003;
        hf.execution_rate = 0.06;
        hf.max_order_shares = 200.0;
        let hf_arc = Arc::new(RwLock::new(hf));
        self.add_participant(hf_arc, &[&symbol]);

        let bank_acct = Arc::new(RwLock::new(Account::new(
            format!("bank_ipo_{}", symbol),
            25_000_000.0,
            3.0,
            0.15,
        )));
        let mut bank = Bank::new(format!("bank_ipo_{}", symbol), bank_acct, vec![co.clone()]);
        bank.min_trade_threshold = 0.006;
        bank.full_conviction_threshold = 0.040;
        bank.rebalance_threshold = 0.006;
        bank.execution_rate = 0.05;
        bank.max_order_shares = 250.0;
        let bank_arc = Arc::new(RwLock::new(bank));
        self.add_bank(bank_arc, &[&symbol]);

        let offering_price = if co.true_value <= 0.0 { 100.0 } else { co.true_value };
        let book = self.book_mut(&symbol).unwrap();
        for i in 0..5 {
            let offset = 0.05 * (i as f64 + 1.0);
            let bp = ((offering_price - offset) * 100.0).round() / 100.0;
            let ap = ((offering_price + offset) * 100.0).round() / 100.0;
            book.add_limit_order(Order {
                id: 0,
                agent_id: format!("underwriter_{}", symbol),
                side: Side::Buy,
                price: bp,
                quantity: 250.0,
                is_market: false,
            });
            book.add_limit_order(Order {
                id: 0,
                agent_id: format!("underwriter_{}", symbol),
                side: Side::Sell,
                price: ap,
                quantity: 250.0,
                is_market: false,
            });
        }

        self.event_log.push(Event {
            tick: self.tick,
            kind: EventKind::IPO,
            symbol: symbol.clone(),
            fundamental_kind: None,
            multiplier: None,
            restatement_profile: None,
            prior_gap_percent: None,
            severity: None,
            liquidated_agent_id: None,
            liquidation_qty: None,
            ipo_revenue: Some(co.annual_revenue),
            ipo_price: Some(offering_price),
            ipo_shares: Some(co.shares_outstanding),
            distress_details: None,
            macro_headline: None,
        });
    }

    pub fn bailout_company(&mut self, symbol: &str, mut capital_injection: f64) -> Result<(), &'static str> {
        let cm = self.markets.get_mut(symbol).ok_or("company not registered in simulation")?;
        if capital_injection <= 0.0 {
            capital_injection = 500_000_000.0;
        }

        let debt_relief = cm.co.debt_outstanding * 0.50;
        cm.co.debt_outstanding -= debt_relief;
        cm.co.interest_expense *= 0.50;
        cm.co.true_value *= 1.50;
        cm.co.reported_value *= 1.50;

        self.national.fiscal.national_debt += capital_injection;

        let headline = format!(
            "Federal Government injected ${:.0}M emergency facility (cut debt by ${:.0}M, +50% intrinsic value)",
            capital_injection / 1e6,
            debt_relief / 1e6
        );

        self.event_log.push(Event {
            tick: self.tick,
            kind: EventKind::Bailout,
            symbol: symbol.to_string(),
            fundamental_kind: None,
            multiplier: None,
            restatement_profile: None,
            prior_gap_percent: None,
            severity: None,
            liquidated_agent_id: None,
            liquidation_qty: None,
            ipo_revenue: None,
            ipo_price: None,
            ipo_shares: None,
            distress_details: Some(headline),
            macro_headline: None,
        });

        Ok(())
    }

    pub fn toggle_policy(&mut self, policy_id: &str) -> Result<(bool, String), String> {
        let (tier, _, _) = self.national.current_tier();
        let (active, name) = self.national.fiscal.toggle_policy_checked(policy_id, tier)?;

        if policy_id == POLICY_CITIZEN_STIMULUS && active {
            self.citizen.sentiment.index = (self.citizen.sentiment.index + 15.0).min(100.0);
            self.citizen.sentiment.happiness = (self.citizen.sentiment.happiness + 12.0).min(100.0);
            self.citizen.demographics.aggregate_disposable_income += 12_000_000_000.0;
            self.national.fiscal.national_debt += 12_000_000_000.0;
        }

        // Immediate market shock: apply sector valuation impact on toggle
        let policies = all_known_policies();
        if let Some(pol_info) = policies.iter().find(|p| p.id == policy_id) {
            // The shock is 0.5% of the procurement delta per company in the favored sector
            // Policy active = positive shock, deactivated = negative shock
            let shock_mult = if active { 1.005_f64 } else { 0.995_f64 };
            let favored = &pol_info.favored_sector;
            for symbol in &self.order {
                if let Some(cm) = self.markets.get_mut(symbol) {
                    let sector_name = cm.co.sector.to_string();
                    if favored == "All Sectors" || &sector_name == favored {
                        cm.co.true_value = (cm.co.true_value * shock_mult).max(0.01);
                        cm.co.reported_value = (cm.co.reported_value * shock_mult).max(0.01);
                    }
                }
            }
        }

        let action_str = if active { "ENACTED" } else { "REPEALED" };
        let headline = format!("Executive Policy {}: {}", action_str, name);

        self.event_log.push(Event {
            tick: self.tick,
            kind: EventKind::Macro,
            symbol: "SOVEREIGN_GOV".to_string(),
            fundamental_kind: None,
            multiplier: None,
            restatement_profile: None,
            prior_gap_percent: None,
            severity: None,
            liquidated_agent_id: None,
            liquidation_qty: None,
            ipo_revenue: None,
            ipo_price: None,
            ipo_shares: None,
            distress_details: None,
            macro_headline: Some(headline),
        });

        self.national_report.active_policies = self.national.fiscal.get_active_policies();
        Ok((active, name))
    }
}
