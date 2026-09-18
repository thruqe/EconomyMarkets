use std::sync::Arc;
use parking_lot::RwLock;
use economy::agent::bank::Bank;
use economy::agent::hedgefund::HedgeFund;
use economy::agent::marketmaker::MarketMaker;
use economy::company::company::{generate_universe, GenerationParams};
use economy::market::margin::Account;
use economy::retail::crowd::generate_watchlist_pool;
use economy::sim::simulation::Simulation;

#[test]
fn test_simulation_liveliness() {
    let params = GenerationParams::default();
    let universe = generate_universe(10, 42, &params);
    let mut sim = Simulation::new(10, 100);

    let mut company_arcs = Vec::new();
    for co in universe {
        let sym = co.symbol.clone();
        sim.add_company(co.clone());

        let mut mm = MarketMaker::new(format!("mm_{}", sym), 50_000_000.0, 10.0, 0.05);
        mm.max_inventory = 250_000.0;
        mm.quote_size = 500.0;
        mm.set_company(co.clone());
        sim.add_participant(Arc::new(RwLock::new(mm)), &[&sym]);

        let hf_acct = Arc::new(RwLock::new(Account::new(format!("hf_{}", sym), 20_000_000.0, 5.0, 0.10)));
        let hf = HedgeFund::new(format!("hf_{}", sym), hf_acct, co.clone());
        sim.add_participant(Arc::new(RwLock::new(hf)), &[&sym]);

        let bank_acct = Arc::new(RwLock::new(Account::new(format!("bank_{}", sym), 100_000_000.0, 3.0, 0.15)));
        let bank = Bank::new(format!("bank_{}", sym), bank_acct, vec![co.clone()]);
        sim.add_participant(Arc::new(RwLock::new(bank)), &[&sym]);

        company_arcs.push(co);
    }

    let pool = generate_watchlist_pool(50, 42, &company_arcs, 1, 3, 1000.0, 20000.0);
    for bot in pool {
        let bot_arc = Arc::new(RwLock::new(bot));
        let watched = bot_arc.read().watchlist.clone();
        let watched_refs: Vec<&str> = watched.keys().map(|s| s.as_str()).collect();
        sim.add_participant(bot_arc, &watched_refs);
    }

    let mut total_trades = 0;
    for _ in 0..50 {
        let report = sim.step();
        total_trades += report.trades.len();
    }

    assert!(total_trades > 0, "Simulation must generate active matched fills across agents");
}
