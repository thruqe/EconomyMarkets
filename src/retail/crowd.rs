use std::sync::Arc;
use parking_lot::RwLock;
use rand::rngs::StdRng;
use rand::{Rng, SeedableRng};

use crate::company::{CapTier, Company};
use crate::market::Account;
use super::bot::SimulatedRetailTrader;
use super::taxonomy::{Archetype, InformationStyle, SkillTier};

#[derive(Clone)]
struct WeightedItem<T> {
    item: T,
    weight: f64,
}

fn weighted_pick<T: Clone, R: Rng + ?Sized>(items: &[WeightedItem<T>], rng: &mut R) -> T {
    let total: f64 = items.iter().map(|it| it.weight).sum();
    let r = rng.gen::<f64>() * total;
    let mut cumulative = 0.0;
    for it in items {
        cumulative += it.weight;
        if r <= cumulative {
            return it.item.clone();
        }
    }
    items.last().unwrap().item.clone()
}

fn pick_tier<R: Rng + ?Sized>(rng: &mut R) -> SkillTier {
    let items = [
        WeightedItem { item: SkillTier::Beginner, weight: 0.55 },
        WeightedItem { item: SkillTier::Intermediate, weight: 0.35 },
        WeightedItem { item: SkillTier::Pro, weight: 0.10 },
    ];
    weighted_pick(&items, rng)
}

fn pick_style<R: Rng + ?Sized>(rng: &mut R) -> InformationStyle {
    let items = [
        WeightedItem { item: InformationStyle::TechnicalOnly, weight: 0.45 },
        WeightedItem { item: InformationStyle::Confused, weight: 0.25 },
        WeightedItem { item: InformationStyle::Blended, weight: 0.20 },
        WeightedItem { item: InformationStyle::FundamentalOnly, weight: 0.10 },
    ];
    weighted_pick(&items, rng)
}

fn pick_archetype<R: Rng + ?Sized>(rng: &mut R) -> Archetype {
    let items = [
        WeightedItem { item: Archetype::MomentumChaser, weight: 0.30 },
        WeightedItem { item: Archetype::Contrarian, weight: 0.20 },
        WeightedItem { item: Archetype::PanicProne, weight: 0.20 },
        WeightedItem { item: Archetype::Degenerate, weight: 0.20 },
        WeightedItem { item: Archetype::Disciplined, weight: 0.10 },
    ];
    weighted_pick(&items, rng)
}

pub fn attention_score(c: &Company) -> f64 {
    match c.cap_tier {
        CapTier::MegaCap => 0.3,
        CapTier::LargeCap => 0.6,
        CapTier::MidCap => 1.0,
        CapTier::SmallCap => 1.6,
    }
}

pub fn generate_watchlist_pool(
    n: usize,
    master_seed: i64,
    universe: &[Company],
    min_watchlist: usize,
    max_watchlist: usize,
    starting_cash_min: f64,
    starting_cash_max: f64,
) -> Vec<SimulatedRetailTrader> {
    let mut seeder = StdRng::seed_from_u64(master_seed as u64);
    let mut pool = Vec::with_capacity(n);

    let scored: Vec<WeightedItem<Company>> = universe
        .iter()
        .map(|c| WeightedItem {
            item: c.clone(),
            weight: attention_score(c),
        })
        .collect();

    for i in 0..n {
        let tier = pick_tier(&mut seeder);
        let style = pick_style(&mut seeder);
        let archetype = pick_archetype(&mut seeder);
        let cash = starting_cash_min + seeder.gen::<f64>() * (starting_cash_max - starting_cash_min);

        let watchlist_size = if max_watchlist > min_watchlist {
            min_watchlist + seeder.gen_range(0..=(max_watchlist - min_watchlist))
        } else {
            min_watchlist
        };

        let watchlist = sample_watchlist(&scored, watchlist_size, &mut seeder);

        let sub_seed = seeder.gen::<u64>();
        let bot_rng = StdRng::seed_from_u64(sub_seed);

        let id = format!("retail_{}", i);
        let acct = Arc::new(RwLock::new(Account::new(id.clone(), cash, 2.0, 0.20)));

        let bot = SimulatedRetailTrader::new(id, acct, watchlist, tier, style, archetype, bot_rng);
        pool.push(bot);
    }

    pool
}

fn sample_watchlist<R: Rng + ?Sized>(
    scored: &[WeightedItem<Company>],
    size: usize,
    rng: &mut R,
) -> Vec<Company> {
    let target_size = size.min(scored.len());
    let mut remaining: Vec<WeightedItem<Company>> = scored.to_vec();
    let mut result = Vec::with_capacity(target_size);

    while result.len() < target_size && !remaining.is_empty() {
        let picked = weighted_pick(&remaining, rng);
        result.push(picked.clone());

        if let Some(pos) = remaining.iter().position(|it| it.item.symbol == picked.symbol) {
            remaining.remove(pos);
        }
    }

    result
}
