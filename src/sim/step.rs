use std::collections::HashMap;
use chrono::Utc;
use rand::Rng;

use crate::company::JumpKind;
use crate::market::{apply_settled_fill, Fill, LiquidationEngine, Order, Side};
use super::events::{Event, EventKind};
use super::report::{AccountSnapshot, PriceQuote, TickReport, Trade};
use super::simulation::Simulation;

const MAX_LIQUIDATION_PASSES: usize = 5;

impl Simulation {
    pub fn step(&mut self) -> TickReport {
        self.tick += 1;
        let events_start = self.event_log.len();
        let mut tick_trades = Vec::new();

        self.advance_fundamentals();
        self.refresh_bank_external_prices();
        self.run_trading_round(&mut tick_trades);
        self.refresh_bank_external_prices();
        self.run_liquidation_cascade(&mut tick_trades);
        self.record_prices();

        let mut prices = HashMap::with_capacity(self.order.len());
        for symbol in &self.order {
            let cm = &self.markets[symbol];
            let mid = cm.book.mid_price().unwrap_or(0.0);
            let has_mid = cm.book.mid_price().is_some();
            let bid = cm.book.best_bid().unwrap_or(0.0);
            let ask = cm.book.best_ask().unwrap_or(0.0);
            let spread = cm.book.spread().unwrap_or(0.0);
            prices.insert(
                symbol.clone(),
                PriceQuote {
                    symbol: symbol.clone(),
                    mid,
                    bid,
                    ask,
                    spread,
                    has_mid,
                },
            );
        }

        let tick_events = self.event_log[events_start..].to_vec();

        let current_prices = self.current_mid_prices();
        let mut accts = Vec::new();
        for a_arc in self.all_accounts() {
            let a = a_arc.read();
            let mr = a.margin_ratio(&current_prices);
            accts.push(AccountSnapshot {
                agent_id: a.agent_id.clone(),
                cash: a.cash,
                equity: a.equity(&current_prices),
                margin_ratio: mr,
            });
        }

        TickReport {
            tick: self.tick,
            timestamp: Utc::now(),
            prices,
            trades: tick_trades,
            events: tick_events,
            accounts: accts,
            citizen: self.citizen_report.clone(),
            national: self.national_report.clone(),
        }
    }

    fn advance_fundamentals(&mut self) {
        let mut total_headcount = 0.0;
        let mut total_capex = 0.0;
        let mut total_corporate_taxes = 0.0;
        let mut sum_returns = 0.0;

        for symbol in &self.order {
            let co = &self.markets[symbol].co;
            total_headcount += co.headcount;
            total_capex += co.capex;
            total_corporate_taxes += co.corporate_tax_paid;
            if co.shares_outstanding > 0.0 && co.true_value > 0.0 {
                sum_returns += (co.reported_value - co.true_value) / co.true_value;
            }
        }

        let scaled_headcount = if total_headcount > 0.0 {
            total_headcount * (4_000_000.0 / 100_000.0)
        } else {
            0.0
        };

        let market_return = if !self.order.is_empty() {
            sum_returns / self.order.len() as f64
        } else {
            0.0
        };

        self.citizen_report = self.citizen.tick(
            self.national.labor.employed_workers,
            self.national.labor.average_hourly_wage,
            self.national.central_bank.cpi_inflation_rate,
            self.national.labor.annual_wage_growth,
            market_return,
            1.0 / 252.0,
        );

        self.national_report = self.national.tick(
            self.citizen_report.consumer_spending,
            total_capex,
            scaled_headcount,
            total_corporate_taxes,
            self.citizen_report.taxes_paid_this_tick,
            1.0 / 252.0,
        );

        if !self.national_report.fomc_announcement.is_empty() {
            self.event_log.push(Event {
                tick: self.tick,
                kind: EventKind::Macro,
                symbol: "FED".to_string(),
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
                macro_headline: Some(self.national_report.fomc_announcement.clone()),
            });
        }

        for symbol in &self.order {
            let cm = self.markets.get_mut(symbol).unwrap();

            let sector_str = cm.co.sector.to_string();
            let demand_factor = self
                .citizen_report
                .sector_demand
                .get(&sector_str)
                .copied()
                .unwrap_or(1.0);

            let composite_demand = self
                .national
                .composite_sector_demand(&sector_str, demand_factor);

            cm.co.update_macro(
                self.national.labor.average_hourly_wage,
                self.national.fiscal.corporate_tax_rate,
                self.national.central_bank.ten_year_yield,
                composite_demand,
                1.0 / 252.0,
            );

            let (fundamental, restatement) = cm.co.tick();

            if fundamental.kind != JumpKind::NoJump {
                self.event_log.push(Event {
                    tick: self.tick,
                    kind: EventKind::Fundamental,
                    symbol: symbol.clone(),
                    fundamental_kind: Some(fundamental.kind),
                    multiplier: Some(fundamental.multiplier),
                    restatement_profile: None,
                    prior_gap_percent: None,
                    severity: None,
                    liquidated_agent_id: None,
                    liquidation_qty: None,
                    ipo_revenue: None,
                    ipo_price: None,
                    ipo_shares: None,
                    distress_details: None,
                    macro_headline: None,
                });
            }

            if let Some(r) = restatement {
                self.event_log.push(Event {
                    tick: self.tick,
                    kind: EventKind::Restatement,
                    symbol: symbol.clone(),
                    fundamental_kind: None,
                    multiplier: None,
                    restatement_profile: Some(r.profile),
                    prior_gap_percent: Some(r.prior_gap_percent),
                    severity: Some(r.severity),
                    liquidated_agent_id: None,
                    liquidation_qty: None,
                    ipo_revenue: None,
                    ipo_price: None,
                    ipo_shares: None,
                    distress_details: None,
                    macro_headline: None,
                });
            }

            if cm.co.reported_value < 1.00 {
                let old_val = cm.co.reported_value;
                cm.co.true_value *= 10.0;
                cm.co.reported_value *= 10.0;
                cm.co.shares_outstanding /= 10.0;
                cm.co.float_shares /= 10.0;
                self.event_log.push(Event {
                    tick: self.tick,
                    kind: EventKind::ReverseSplit,
                    symbol: symbol.clone(),
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
                    distress_details: Some(format!(
                        "1-for-10 reverse split executed to maintain listing standards (${:.2} -> ${:.2})",
                        old_val, cm.co.reported_value
                    )),
                    macro_headline: None,
                });
            } else if cm.co.true_value < 5.00 && self.rng.gen::<f64>() < 0.02 {
                let workout_roll = self.rng.gen::<f64>();
                if workout_roll < 0.30 {
                    let mult = 1.40 + self.rng.gen::<f64>() * 0.35;
                    cm.co.true_value *= mult;
                    cm.co.reported_value *= mult;
                    let facility = 250_000_000.0 + self.rng.gen::<f64>() * 500_000_000.0;
                    self.event_log.push(Event {
                        tick: self.tick,
                        kind: EventKind::Bailout,
                        symbol: symbol.clone(),
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
                        distress_details: Some(format!(
                            "Secured ${:.0}M emergency credit line (+{:.0}% value recovery)",
                            facility / 1e6,
                            (mult - 1.0) * 100.0
                        )),
                        macro_headline: None,
                    });
                } else if workout_roll < 0.60 {
                    let mult = 1.50 + self.rng.gen::<f64>() * 0.30;
                    cm.co.true_value *= mult;
                    cm.co.reported_value *= mult;
                    self.event_log.push(Event {
                        tick: self.tick,
                        kind: EventKind::Acquisition,
                        symbol: symbol.clone(),
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
                        distress_details: Some(format!(
                            "Private equity syndicate submitted buyout offer @ ${:.2} (+{:.0}% premium)",
                            cm.co.reported_value,
                            (mult - 1.0) * 100.0
                        )),
                        macro_headline: None,
                    });
                } else if workout_roll < 0.85 {
                    let mult = 1.30 + self.rng.gen::<f64>() * 0.25;
                    cm.co.true_value *= mult;
                    cm.co.reported_value *= mult;
                    cm.co.net_margin *= 1.25;
                    self.event_log.push(Event {
                        tick: self.tick,
                        kind: EventKind::Restructuring,
                        symbol: symbol.clone(),
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
                        distress_details: Some(format!(
                            "Turnaround plan announced: closed non-core units, improved margins (+{:.0}% valuation)",
                            (mult - 1.0) * 100.0
                        )),
                        macro_headline: None,
                    });
                } else {
                    let mult = 1.35 + self.rng.gen::<f64>() * 0.30;
                    cm.co.true_value *= mult;
                    cm.co.reported_value *= mult;
                    self.event_log.push(Event {
                        tick: self.tick,
                        kind: EventKind::Chapter11,
                        symbol: symbol.clone(),
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
                        distress_details: Some(
                            "Court approved Chapter 11 plan: debt cleared, emerging with clean balance sheet".to_string(),
                        ),
                        macro_headline: None,
                    });
                }
            }
        }

        // Advance private companies & naturally trigger IPOs when reaching Pre-IPO stage
        let grants_level = self.national.fiscal.enterprise_grants_level;
        let mut ready_ipos = Vec::new();
        for co in &mut self.private_companies {
            let grant_growth = grants_level * 0.001;
            let organic_growth = 0.0003 + (self.rng.gen::<f64>() * 0.0006);
            co.annual_revenue *= 1.0 + organic_growth + grant_growth;
            co.private_valuation = co.annual_revenue * co.sector_multiple;

            // Advance enterprise stages
            if co.annual_revenue >= 50_000_000.0 {
                co.stage = "Pre-IPO".to_string();
            } else if co.annual_revenue >= 25_000_000.0 {
                co.stage = "Expansion".to_string();
            } else if co.annual_revenue >= 10_000_000.0 {
                co.stage = "Early Stage".to_string();
            } else {
                co.stage = "Seed".to_string();
            }

            // Natural IPO trigger: private enterprise has reached Pre-IPO scale and files prospectus
            if co.stage == "Pre-IPO" && (self.tick % 150 == 0 || self.rng.gen::<f64>() < 0.015) {
                ready_ipos.push(co.symbol.clone());
            }
        }

        for sym in ready_ipos {
            if let Some(pos) = self.private_companies.iter().position(|c| c.symbol == sym) {
                let mut ipo_co = self.private_companies.remove(pos);
                ipo_co.is_public = true;
                ipo_co.is_ipo = true;
                ipo_co.ipo_tick = self.tick;
                ipo_co.ipo_price = (ipo_co.private_valuation / ipo_co.shares_outstanding).max(5.0);
                ipo_co.true_value = ipo_co.ipo_price;
                ipo_co.reported_value = ipo_co.ipo_price;
                ipo_co.stage = "Public".to_string();

                let co_name = ipo_co.name.clone();
                let co_sym = ipo_co.symbol.clone();
                let offering_val = ipo_co.private_valuation;
                let offering_pr = ipo_co.ipo_price;

                self.list_ipo(ipo_co.clone(), 35_000_000.0);

                let headline = format!(
                    "PRE-IPO LISTING: {} ({}) completes Sovereign Exchange IPO at ${:.2}/share (${:.0}M market valuation)",
                    co_name, co_sym, offering_pr, offering_val / 1e6
                );
                self.event_log.push(Event {
                    tick: self.tick,
                    kind: EventKind::IPO,
                    symbol: co_sym.clone(),
                    fundamental_kind: None,
                    multiplier: None,
                    restatement_profile: None,
                    prior_gap_percent: None,
                    severity: None,
                    liquidated_agent_id: None,
                    liquidation_qty: None,
                    ipo_revenue: Some(ipo_co.annual_revenue),
                    ipo_price: Some(offering_pr),
                    ipo_shares: Some(ipo_co.shares_outstanding),
                    distress_details: Some(headline.clone()),
                    macro_headline: Some(headline),
                });

                // Generate replacement seed startup in incubator
                use crate::company::taxonomy::ALL_SECTORS;
                let sector = ALL_SECTORS[self.rng.gen_range(0..ALL_SECTORS.len())];
                let seed_rev = 4_000_000.0 + (self.rng.gen::<f64>() * 12_000_000.0);
                let mut used_syms: std::collections::HashSet<String> = self.order.iter().cloned().collect();
                for pc in &self.private_companies {
                    used_syms.insert(pc.symbol.clone());
                }
                let mut new_startup = crate::company::company::generate_private_enterprise(
                    sector, seed_rev, self.tick, &mut self.rng, &mut used_syms,
                );
                new_startup.init_runtime(None);
                self.private_companies.push(new_startup);
            }
        }
    }

    fn run_trading_round(&mut self, trades: &mut Vec<Trade>) {
        let symbols = self.order.clone();
        for symbol in &symbols {
            let state = {
                let cm = &self.markets[symbol];
                cm.state(self.tick, self.depth_levels)
            };

            let participants = self.markets[symbol].participants.clone();
            let mut all_orders = Vec::new();

            for p_arc in participants {
                let (orders, p_id) = {
                    let mut p = p_arc.write();
                    let id = p.id().to_string();
                    let o = p.next_orders(&state);
                    (o, id)
                };

                if !orders.is_empty() {
                    let has_limit = orders.iter().any(|o| !o.is_market);
                    if has_limit {
                        let cm = self.markets.get_mut(symbol).unwrap();
                        cm.book.cancel_agent_orders(&p_id);
                    }
                    all_orders.extend(orders);
                }
            }

            for o in all_orders {
                let fills = {
                    let cm = self.markets.get_mut(symbol).unwrap();
                    cm.book.submit(o.clone())
                };
                self.settle_fills(symbol, &o, fills, trades);
            }
        }
    }

    fn settle_fills(
        &mut self,
        symbol: &str,
        o: &Order,
        fills: Vec<Fill>,
        trades: &mut Vec<Trade>,
    ) {
        let side_str = if o.side == Side::Buy { "BUY" } else { "SELL" };

        for f in fills {
            trades.push(Trade {
                tick: self.tick,
                symbol: symbol.to_string(),
                taker_id: f.taker_agent_id.clone(),
                maker_id: f.maker_agent_id.clone(),
                side: side_str.to_string(),
                price: f.price,
                quantity: f.quantity,
            });

            if let Some(taker_acct) = self.find_account(&f.taker_agent_id) {
                let mut acct = taker_acct.write();
                apply_settled_fill(&mut acct, symbol, o.side, f.quantity, f.price);
            }
            if let Some(maker_acct) = self.find_account(&f.maker_agent_id) {
                let maker_side = if o.side == Side::Buy { Side::Sell } else { Side::Buy };
                let mut acct = maker_acct.write();
                apply_settled_fill(&mut acct, symbol, maker_side, f.quantity, f.price);
            }
        }
    }

    fn find_account(&self, agent_id: &str) -> Option<std::sync::Arc<parking_lot::RwLock<crate::market::Account>>> {
        for h in &self.holders {
            let holder = h.read();
            if holder.id() == agent_id {
                return Some(holder.account());
            }
        }
        None
    }

    fn refresh_bank_external_prices(&mut self) {
        if self.banks.is_empty() {
            return;
        }
        let prices = self.current_mid_prices();
        for b in &self.banks {
            b.write().set_external_prices(prices.clone());
        }
    }

    pub fn current_mid_prices(&self) -> HashMap<String, f64> {
        let mut prices = HashMap::with_capacity(self.order.len());
        for symbol in &self.order {
            if let Some(mid) = self.markets[symbol].book.mid_price() {
                prices.insert(symbol.clone(), mid);
            }
        }
        prices
    }

    fn run_liquidation_cascade(&mut self, trades: &mut Vec<Trade>) {
        for _ in 0..MAX_LIQUIDATION_PASSES {
            let accounts_arc = self.all_accounts();
            let prices = self.current_mid_prices();

            let forced = {
                let accounts_guards: Vec<_> = accounts_arc.iter().map(|a| a.read()).collect();
                let accounts_refs: Vec<&crate::market::Account> = accounts_guards.iter().map(|g| &**g).collect();
                LiquidationEngine::scan_for_liquidations(&accounts_refs, &prices)
            };

            if forced.is_empty() {
                return;
            }

            for fo in forced {
                let fills = match self.markets.get_mut(&fo.symbol) {
                    Some(cm) => cm.book.submit(fo.order.clone()),
                    None => continue,
                };
                self.settle_fills(&fo.symbol, &fo.order, fills, trades);

                self.event_log.push(Event {
                    tick: self.tick,
                    kind: EventKind::Liquidation,
                    symbol: fo.symbol.clone(),
                    fundamental_kind: None,
                    multiplier: None,
                    restatement_profile: None,
                    prior_gap_percent: None,
                    severity: None,
                    liquidated_agent_id: Some(fo.agent_id),
                    liquidation_qty: Some(fo.order.quantity),
                    ipo_revenue: None,
                    ipo_price: None,
                    ipo_shares: None,
                    distress_details: None,
                    macro_headline: None,
                });
            }
        }
    }

    pub fn all_accounts(&self) -> Vec<std::sync::Arc<parking_lot::RwLock<crate::market::Account>>> {
        let mut seen = std::collections::HashSet::new();
        let mut out = Vec::new();
        for h in &self.holders {
            let acct = h.read().account();
            let ptr = std::sync::Arc::as_ptr(&acct);
            if !seen.contains(&ptr) {
                seen.insert(ptr);
                out.push(acct);
            }
        }
        out
    }

    fn record_prices(&mut self) {
        for symbol in &self.order {
            self.markets.get_mut(symbol).unwrap().record_price();
        }
    }
}
