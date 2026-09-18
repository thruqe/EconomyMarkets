
fn compute_calendar_str(sim_hours: i64) -> (i64, i64, i64, String) {
    let day = (sim_hours / 24) + 1;
    let hour = sim_hours % 24;
    let year = 1 + (day - 1) / 365;
    let day_of_year = (day - 1) % 365 + 1;
    let months = [
        ("Jan", 31), ("Feb", 28), ("Mar", 31), ("Apr", 30),
        ("May", 31), ("Jun", 30), ("Jul", 31), ("Aug", 31),
        ("Sep", 30), ("Oct", 31), ("Nov", 30), ("Dec", 31)
    ];
    let mut rem = day_of_year;
    let mut m_name = "Jan";
    let mut d_month = 1;
    for &(name, days) in months.iter() {
        if rem <= days {
            m_name = name;
            d_month = rem;
            break;
        }
        rem -= days;
    }
    let formatted = format!("Year {}, {} {:02}, {:02}:00 (Day {})", year, m_name, d_month, hour, day);
    (day, hour, year, formatted)
}

fn norm_sym(s: &str) -> String {
    s.chars().filter(|c| c.is_alphanumeric()).collect::<String>().to_uppercase()
}
use axum::{
    extract::{
        ws::{Message, WebSocket, WebSocketUpgrade},
        Query, State,
    },
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use futures_util::{SinkExt, StreamExt};
use parking_lot::RwLock;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use std::time::Instant;
use rand::SeedableRng;
use tokio::sync::broadcast;
use tower_http::cors::{Any, CorsLayer};
use tower_http::services::{ServeDir, ServeFile};

use crate::company::company::Company;
use crate::company::taxonomy::Sector;
use crate::country::country::SovereignBond;
use crate::market::orderbook::{Order, Side};
use crate::retail::human::HumanTrader;
use crate::server::timeframe::{MultiTimeframeManager, TF_1D};
use crate::sim::report::PriceQuote;
use crate::sim::simulation::Simulation;
use crate::storage::db::DB;
use crate::world::world::{BilateralContract, CountryIndex, ForeignLoanRequest, ForeignPolicyAction, ForexPair, Region, World, WorldStock};

#[derive(Clone)]
pub struct AppState {
    pub start_time: Instant,
    pub current_tick: Arc<RwLock<i64>>,
    pub total_trades: Arc<RwLock<i64>>,
    pub total_volume: Arc<RwLock<f64>>,
    pub latest_prices: Arc<RwLock<HashMap<String, PriceQuote>>>,
    pub companies: Arc<RwLock<Vec<Arc<RwLock<Company>>>>>,
    pub sim: Arc<RwLock<Simulation>>,
    pub world: Arc<RwLock<World>>,
    pub human_trader: Arc<RwLock<HumanTrader>>,
    pub tf_manager: Arc<MultiTimeframeManager>,
    pub speed_multiplier: Arc<RwLock<f64>>,
    pub sim_hour: Arc<RwLock<i64>>,
    pub tx_ws: broadcast::Sender<String>,
    pub tx_world_ws: broadcast::Sender<String>,
    pub db: DB,
}

#[derive(Serialize)]
pub struct PriceItem {
    pub symbol: String,
    pub name: String,
    pub sector: String,
    pub cap_tier: String,
    pub true_value: f64,
    pub reported_value: f64,
    pub mid: f64,
    pub bid: f64,
    pub ask: f64,
    pub spread: f64,
    pub annual_revenue: f64,
    pub net_margin: f64,
    pub sector_multiple: f64,
    pub market_cap: f64,
    pub reported_market_cap: f64,
    pub shares_outstanding: f64,
    pub float_shares: f64,
    pub pe_ratio: f64,
    pub ps_ratio: f64,
    pub is_ipo: bool,
    pub ipo_tick: i64,
    pub ipo_price: f64,
    pub is_private: bool,
    pub stage: String,
    pub private_valuation: f64,
}


#[derive(Serialize)]
pub struct PositionDTO {
    pub symbol: String,
    pub side: String,
    pub quantity: f64,
    pub entry_price: f64,
    pub current_price: f64,
    pub value: f64,
    pub unrealized_pnl: f64,
    pub pnl_percent: f64,
}

#[derive(Serialize)]
pub struct PortfolioDTO {
    pub cash: f64,
    pub equity: f64,
    pub buying_power: f64,
    pub margin_ratio: f64,
    pub maintenance_margin_ratio: f64,
    pub max_leverage: f64,
    pub positions: Vec<PositionDTO>,
}

#[derive(Serialize)]
pub struct PolicyItemDTO {
    pub id: String,
    pub category: String,
    pub name: String,
    pub description: String,
    pub favored_sector: String,
    pub annual_cost: f64,
    pub active: bool,
    pub min_tier: String,
    pub is_locked: bool,
    pub rollout_days: u32,
    pub days_active: u32,
    pub efficacy: f64,
    pub direct_positive: String,
    pub direct_negative: String,
    pub indirect_positive: String,
    pub indirect_negative: String,
}

#[derive(Serialize)]
pub struct MacroDTO {
    pub tick: usize,
    pub sim_hour: i64,
    pub sim_day: i64,
    pub sim_year: i64,
    pub sim_date_formatted: String,
    pub gdp: f64,
    pub real_gdp_growth: f64,
    pub central_bank_rate: f64,
    pub ten_year_yield: f64,
    pub inflation_rate: f64,
    pub policy_stance: String,
    pub unemployment_rate: f64,
    pub labor_force: f64,
    pub employed_workers: f64,
    pub average_hourly_wage: f64,
    pub consumer_sentiment: f64,
    pub consumer_spending: f64,
    pub national_debt: f64,
    pub federal_revenue: f64,
    pub federal_outlays: f64,
    pub budget_deficit: f64,
    pub credit_rating: String,
    pub treasury_cash: f64,
    pub borrowing_yield: f64,
    pub dxy_index: f64,
    pub trade_balance: f64,
    pub maintenance_spending_rate: f64,
    pub depreciation_rate: f64,
    pub tourism_revenue: f64,
    pub productivity_index: f64,
    pub bonds: Vec<SovereignBond>,
    pub tier: String,
    pub tier_name: String,
    pub tier_progress: f64,
    pub policies: Vec<PolicyItemDTO>,
}

#[derive(Serialize)]
pub struct NationDTO {
    pub name: String,
    pub currency: String,
    pub flag_code: String,
    pub founded: i32,
    pub configured: bool,
    pub infrastructure_level: f64,
    pub healthcare_level: f64,
    pub education_level: f64,
    pub enterprise_grants_level: f64,
    pub export_capacity_level: f64,
    pub exchange_chartered: bool,
    pub treasury_cash: f64,
    pub reserve_currency_reserves: f64,
    pub national_debt: f64,
    pub credit_rating: String,
    pub borrowing_yield: f64,
    pub maintenance_spending_rate: f64,
    pub depreciation_rate: f64,
    pub tourism_revenue: f64,
    pub productivity_index: f64,
    pub bonds: Vec<SovereignBond>,
    pub loan_requests: Vec<ForeignLoanRequest>,
    pub tier: String,
    pub tier_name: String,
    pub tier_progress: f64,
    pub regions: Vec<Region>,

    // Demographic & Geopolitical
    pub population: f64,
    pub gdp_per_capita: f64,
    pub labor_force: f64,
    pub employed_workers: f64,
    pub average_hourly_wage: f64,
    pub job_openings: f64,
    pub unemployment_rate: f64,
    pub geopolitical_power: f64,
    pub personal_income_tax_rate: f64,
    pub corporate_tax_rate: f64,
    pub sales_tax_rate: f64,
    pub foreign_debt_custody_enabled: bool,
    pub foreign_debt_custody_nation: String,
    pub foreign_debt_custody_amount: f64,
    pub sim_hour: i64,
    pub sim_day: i64,
    pub sim_year: i64,
    pub sim_date_formatted: String,
}

#[derive(Serialize)]
pub struct ForeignCountryDTO {
    pub id: String,
    pub name: String,
    pub currency: String,
    pub flag_code: String,
    pub gdp: f64,
    pub gdp_growth: f64,
    pub inflation: f64,
    pub interest_rate: f64,
    pub unemployment: f64,
    pub population: f64,
    pub trade_balance: f64,
    pub tariff_rate: f64,
    pub stance: String,
    pub trade_volume: f64,
    pub sanctions_level: i32,
    pub aid_flow: f64,

    // Demographics & Geopolitics
    pub currency_strength: f64,
    pub national_debt: f64,
    pub debt_to_gdp: f64,
    pub labor_force: f64,
    pub employed_workers: f64,
    pub gdp_per_capita: f64,
    pub average_hourly_wage: f64,
    pub job_openings: f64,
    pub productivity_index: f64,
    pub geopolitical_power: f64,
    pub sanctioned_by_us: bool,
    pub sanctioned_us: bool,
    pub sanctions_summary: String,
    pub is_reserve_currency: bool,
}

#[derive(Serialize)]
pub struct WorldDTO {
    pub countries: Vec<ForeignCountryDTO>,
    pub forex_pairs: Vec<ForexPair>,
    pub world_stocks: Vec<WorldStock>,
    pub country_indices: Vec<CountryIndex>,
    pub loan_requests: Vec<ForeignLoanRequest>,
    pub bilateral_contracts: Vec<BilateralContract>,
    pub policy_log: Vec<ForeignPolicyAction>,
    pub reserve_currency: String,
    pub reserve_country_name: String,
    pub reserve_currency_reserves: f64,
}

#[derive(Serialize, Clone)]
pub struct EditorialNewsDTO {
    pub id: i64,
    pub tick: i64,
    pub timestamp: i64,
    pub headline: String,
    pub wire_source: String,
    pub category: String,
    pub urgency: String,
    pub sentiment: String,
    pub summary: String,
    pub affected_symbol: Option<String>,
}

#[derive(Deserialize)]
pub struct FoundNationRequest {
    pub name: String,
    pub currency: String,
    pub flag_code: String,
    pub founded: i32,
}

#[derive(Deserialize)]
pub struct InvestRequest {
    pub pillar: String,
    pub amount: f64,
}

#[derive(Deserialize)]
pub struct DiplomacyRequest {
    pub action: String,
    pub country_id: String,
    pub amount: Option<f64>,
    pub level: Option<i32>,
    pub tariff: Option<f64>,
}

#[derive(Deserialize)]
pub struct BondActionRequest {
    pub maturity_years: u32,
    pub amount: f64,
}

#[derive(Deserialize)]
pub struct MaintenanceRequest {
    pub rate: f64,
}

#[derive(Deserialize)]
pub struct LoanRespondRequest {
    pub request_id: String,
    pub action: String,
}

#[derive(Deserialize)]
pub struct ForexSwapRequest {
    pub pair: String,
    pub amount: f64,
    pub side: String,
}

#[derive(Deserialize)]
pub struct BorrowCapitalRequest {
    pub amount: f64,
}

#[derive(Deserialize)]
pub struct RepayDebtRequest {
    pub amount: f64,
}

#[derive(Deserialize)]
pub struct TradeReserveCurrencyRequest {
    pub side: String, // "buy" or "sell"
    pub amount: f64,
}

#[derive(Deserialize)]
pub struct PortfolioSettingsRequest {
    pub balance: Option<f64>,
    pub leverage: Option<f64>,
}

#[derive(Deserialize)]
pub struct ContractRespondRequest {
    pub contract_id: String,
    pub action: String, // "accept" or "decline"
}

#[derive(Deserialize)]
pub struct SanctionToggleRequest {
    pub country_id: String,
    pub action: String, // "sanction" or "lift"
}

pub fn create_router(state: AppState) -> Router {
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let serve_dir = ServeDir::new("frontend/dist")
        .not_found_service(ServeFile::new("frontend/dist/index.html"));

    Router::new()
        .route("/health", get(health_handler))
        .route("/api/status", get(status_handler))
        .route("/api/prices", get(prices_handler))
        .route("/api/trades", get(trades_handler))
        .route("/api/candles", get(candles_handler))
        .route("/api/sim/speed", post(speed_handler))
        .route("/api/events", get(events_handler))
        .route("/api/sim/ipo", post(ipo_handler))
        .route("/api/book", get(book_handler))
        .route("/api/orders", post(orders_handler))
        .route("/api/portfolio", get(portfolio_handler))
        .route("/api/portfolio/close", post(close_position_handler))
        .route("/api/portfolio/settings", post(portfolio_settings_handler))
        .route("/api/macro", get(macro_handler))
        .route("/api/policy/toggle", post(policy_toggle_handler))
        .route("/api/nation", get(get_nation_handler))
        .route("/api/nation/found", post(found_nation_handler))
        .route("/api/nation/invest", post(invest_nation_handler))
        .route("/api/nation/charter", post(charter_exchange_handler))
        .route("/api/nation/maintenance", post(set_maintenance_handler))
        .route("/api/nation/loan/respond", post(loan_respond_handler))
        .route("/api/nation/contract/respond", post(contract_respond_handler))
        .route("/api/nation/world", get(get_world_handler))
        .route("/api/nation/diplomacy", post(diplomacy_handler))
        .route("/api/bonds", get(get_bonds_handler))
        .route("/api/bonds/issue", post(issue_bond_handler))
        .route("/api/bonds/buyback", post(buyback_bond_handler))
        .route("/api/forex", get(get_forex_handler))
        .route("/api/forex/swap", post(forex_swap_handler))
        .route("/api/world/stocks", get(get_world_stocks_handler))
        .route("/api/world/indices", get(get_world_indices_handler))
        .route("/api/nation/borrow", post(borrow_capital_handler))
        .route("/api/nation/debt/repay", post(repay_debt_handler))
        .route("/api/nation/reserve-currency/trade", post(trade_reserve_currency_handler))
        .route("/api/nation/sanctions/toggle", post(sanction_toggle_handler))
        .route("/api/nation/taxes", post(set_taxes_handler))
        .route("/api/nation/borrow/bilateral", post(borrow_bilateral_handler))
        .route("/api/nation/debt-custody/toggle", post(toggle_debt_custody_handler))
        .route("/api/auth/login", post(login_handler))
        .route("/api/auth/signup", post(signup_handler))
        .route("/ws", get(ws_handler))
        .route("/api/ws", get(ws_handler))
        .route("/ws/world", get(world_ws_handler))
        .fallback_service(serve_dir)
        .layer(cors)
        .with_state(state)
}

pub fn build_prices_dto(state: &AppState) -> Vec<PriceItem> {
    let latest_prices = state.latest_prices.read().clone();
    let sim = state.sim.read();

    let mut items = Vec::with_capacity(sim.order.len() + sim.private_companies.len());

    // 1. Live actively traded public equities
    for symbol in &sim.order {
        if let Some(cm) = sim.markets.get(symbol) {
            let co = &cm.co;
            let q = latest_prices.get(symbol).cloned().unwrap_or_default();
            let market_price = if q.mid > 0.0 { q.mid } else { co.true_value };

            items.push(PriceItem {
                symbol: co.symbol.clone(),
                name: co.name.clone(),
                sector: co.sector.to_string(),
                cap_tier: co.cap_tier.to_string(),
                true_value: co.true_value,
                reported_value: co.reported_value,
                mid: q.mid,
                bid: q.bid,
                ask: q.ask,
                spread: q.spread,
                annual_revenue: co.annual_revenue,
                net_margin: co.net_margin,
                sector_multiple: co.sector_multiple,
                market_cap: co.market_cap(),
                reported_market_cap: co.reported_market_cap(),
                shares_outstanding: co.shares_outstanding,
                float_shares: co.float_shares,
                pe_ratio: co.price_to_earnings(market_price),
                ps_ratio: co.price_to_sales(market_price),
                is_ipo: co.is_ipo,
                ipo_tick: co.ipo_tick as i64,
                ipo_price: co.ipo_price,
                is_private: false,
                stage: "Public".to_string(),
                private_valuation: 0.0,
            });
        }
    }

    // 2. Private pre-IPO / growth enterprises in domestic pipeline
    for co in &sim.private_companies {
        items.push(PriceItem {
            symbol: co.symbol.clone(),
            name: co.name.clone(),
            sector: co.sector.to_string(),
            cap_tier: co.cap_tier.to_string(),
            true_value: 0.0,
            reported_value: 0.0,
            mid: 0.0,
            bid: 0.0,
            ask: 0.0,
            spread: 0.0,
            annual_revenue: co.annual_revenue,
            net_margin: co.net_margin,
            sector_multiple: co.sector_multiple,
            market_cap: co.private_valuation,
            reported_market_cap: co.private_valuation,
            shares_outstanding: co.shares_outstanding,
            float_shares: co.float_shares,
            pe_ratio: 0.0,
            ps_ratio: 0.0,
            is_ipo: false,
            ipo_tick: 0,
            ipo_price: 0.0,
            is_private: true,
            stage: co.stage.clone(),
            private_valuation: co.private_valuation,
        });
    }

    items
}


async fn ws_handler(
    ws: WebSocketUpgrade,
    State(state): State<AppState>,
) -> impl IntoResponse {
    ws.on_upgrade(move |socket| handle_socket(socket, state))
}

async fn handle_socket(socket: WebSocket, state: AppState) {
    let (mut sender, mut receiver) = socket.split();
    let mut rx = state.tx_ws.subscribe();

    // Send initial complete snapshot on connect
    let snapshot = {
        let tick = *state.current_tick.read();
        let uptime = format!("{:.1}s", state.start_time.elapsed().as_secs_f64());
        let trades_count = *state.total_trades.read();
        let volume = *state.total_volume.read();
        let speed = *state.speed_multiplier.read();
        let sim_status = serde_json::json!({
            "status": "running",
            "tick": tick,
            "uptime": uptime,
            "total_trades": trades_count,
            "total_volume": volume,
            "companies_count": state.companies.read().len(),
            "speed_multiplier": speed,
        });

        let prices = build_prices_dto(&state);
        let macro_dto = build_macro_dto(&state);
        let nation_dto = build_nation_dto(&state);
        let portfolio_dto = build_portfolio_dto(&state);

        serde_json::json!({
            "type": "snapshot",
            "status": sim_status,
            "prices": prices,
            "macro": macro_dto,
            "nation": nation_dto,
            "portfolio": portfolio_dto,
        })
    };

    if sender.send(Message::Text(snapshot.to_string().into())).await.is_err() {
        return;
    }

    let subscribed_symbol = Arc::new(RwLock::new(None::<String>));
    let sub_sym_for_rx = subscribed_symbol.clone();
    let state_for_rx = state.clone();

    let mut read_task = tokio::spawn(async move {
        while let Some(Ok(msg)) = receiver.next().await {
            match msg {
                Message::Text(text) => {
                    if let Ok(val) = serde_json::from_str::<serde_json::Value>(&text) {
                        if val["action"] == "subscribe_book" {
                            if let Some(s) = val["symbol"].as_str() {
                                *subscribed_symbol.write() = Some(s.to_string());
                            }
                        } else if val["action"] == "unsubscribe_book" {
                            *subscribed_symbol.write() = None;
                        }
                    }
                }
                Message::Close(_) => break,
                _ => {}
            }
        }
    });

    let mut write_task = tokio::spawn(async move {
        while let Ok(msg_str) = rx.recv().await {
            let maybe_book_json = {
                let sym_opt = sub_sym_for_rx.read().clone();
                if let Some(sym) = sym_opt {
                    let sim = state_for_rx.sim.read();
                    if let Some(book) = sim.book(&sym) {
                        let (bids, asks) = book.top_levels(15);
                        let mid = book.mid_price();
                        let spread = book.spread().unwrap_or(0.0);
                        Some(serde_json::json!({
                            "type": "book",
                            "symbol": sym,
                            "mid": mid.unwrap_or(0.0),
                            "spread": spread,
                            "bids": bids,
                            "asks": asks,
                        }))
                    } else {
                        None
                    }
                } else {
                    None
                }
            };

            if sender.send(Message::Text(msg_str.into())).await.is_err() {
                break;
            }

            if let Some(book_json) = maybe_book_json {
                if sender.send(Message::Text(book_json.to_string().into())).await.is_err() {
                    break;
                }
            }
        }
    });

    tokio::select! {
        _ = (&mut read_task) => write_task.abort(),
        _ = (&mut write_task) => read_task.abort(),
    }
}

async fn world_ws_handler(
    ws: WebSocketUpgrade,
    State(state): State<AppState>,
) -> impl IntoResponse {
    ws.on_upgrade(move |socket| handle_world_ws(socket, state))
}

async fn handle_world_ws(socket: WebSocket, state: AppState) {
    let (mut sender, _receiver) = socket.split();
    let mut rx = state.tx_world_ws.subscribe();

    // Send initial world snapshot
    let world_snapshot = {
        let w = state.world.read();
        let countries: Vec<ForeignCountryDTO> = w.foreign.values()
            .map(|c| ForeignCountryDTO {
                id: c.id.clone(),
                name: c.name.clone(),
                currency: c.currency.clone(),
                flag_code: c.flag_code.clone(),
                gdp: c.gdp,
                gdp_growth: c.gdp_growth,
                inflation: c.inflation,
                interest_rate: c.interest_rate,
                unemployment: c.unemployment,
                population: c.population,
                trade_balance: c.trade_balance,
                tariff_rate: c.relation.tariff_rate,
                stance: c.relation.stance.to_string(),
                trade_volume: c.relation.trade_volume,
                sanctions_level: c.relation.sanctions_level,
                aid_flow: c.relation.aid_flow,
                currency_strength: c.currency_strength,
                national_debt: c.national_debt,
                debt_to_gdp: c.debt_to_gdp,
                labor_force: c.labor_force,
                employed_workers: c.employed_workers,
                gdp_per_capita: c.gdp_per_capita,
                average_hourly_wage: c.average_hourly_wage,
                job_openings: c.job_openings,
                productivity_index: c.productivity_index,
                geopolitical_power: c.geopolitical_power,
                sanctioned_by_us: c.sanctioned_by_us,
                sanctioned_us: c.sanctioned_us,
                sanctions_summary: c.sanctions_summary.clone(),
                is_reserve_currency: c.is_reserve_currency,
            })
            .collect();
        serde_json::json!({
            "type": "world_snapshot",
            "countries": countries,
            "forex_pairs": w.forex_pairs,
            "world_stocks": w.world_stocks,
            "country_indices": w.country_indices,
        })
    };

    if sender.send(Message::Text(world_snapshot.to_string().into())).await.is_err() {
        return;
    }

    while let Ok(msg_str) = rx.recv().await {
        if sender.send(Message::Text(msg_str.into())).await.is_err() {
            break;
        }
    }
}

async fn health_handler(State(state): State<AppState>) -> impl IntoResponse {
    let tick = *state.current_tick.read();
    let uptime = format!("{:.1}s", state.start_time.elapsed().as_secs_f64());
    Json(serde_json::json!({
        "status": "healthy",
        "tick": tick,
        "uptime": uptime,
    }))
}

async fn status_handler(State(state): State<AppState>) -> impl IntoResponse {
    let tick = *state.current_tick.read();
    let uptime = format!("{:.0}s", state.start_time.elapsed().as_secs_f64());
    let companies_len = state.companies.read().len();
    let total_trades = *state.total_trades.read();
    let total_volume = *state.total_volume.read();

    Json(serde_json::json!({
        "tick": tick,
        "uptime": uptime,
        "companies": companies_len,
        "total_trades": total_trades,
        "total_volume": total_volume,
    }))
}

async fn prices_handler(State(state): State<AppState>) -> impl IntoResponse {
    Json(build_prices_dto(&state))
}

#[derive(Deserialize)]
struct SymbolQuery {
    symbol: Option<String>,
}

async fn trades_handler(State(state): State<AppState>, Query(q): Query<SymbolQuery>) -> impl IntoResponse {
    let sym = match q.symbol {
        Some(s) if !s.is_empty() => s,
        _ => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": "symbol parameter required"}))).into_response(),
    };

    match state.db.get_recent_trades(&sym, 50) {
        Ok(trades) => Json(trades).into_response(),
        Err(e) => (StatusCode::INTERNAL_SERVER_ERROR, Json(serde_json::json!({"error": e.to_string()}))).into_response(),
    }
}

#[derive(Deserialize)]
struct CandleQuery {
    symbol: Option<String>,
    tf: Option<String>,
    timeframe: Option<String>,
    limit: Option<usize>,
}

async fn candles_handler(State(state): State<AppState>, Query(q): Query<CandleQuery>) -> impl IntoResponse {
    let sym = match q.symbol {
        Some(s) if !s.is_empty() => s,
        _ => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": "symbol parameter required"}))).into_response(),
    };

    let tf = q.tf.or(q.timeframe).unwrap_or_else(|| TF_1D.to_string());
    let limit = q.limit.unwrap_or(3000);

    if let Some(bars) = state.tf_manager.get_bars(&sym, &tf, limit) {
        if !bars.is_empty() {
            return Json(bars).into_response();
        }
    }

    if let Ok(bars) = state.db.get_candles(&sym, &tf, limit) {
        if !bars.is_empty() {
            return Json(bars).into_response();
        }
    }

    // Auto-synthesize baseline candles for any requested symbol from current price if known
    let current_price = {
        let sim = state.sim.read();
        if let Some(cm) = sim.markets.get(&sym) {
            cm.co.true_value
        } else if let Some(pco) = sim.private_companies.iter().find(|c| c.symbol == sym) {
            pco.true_value.max(25.0)
        } else {
            let w = state.world.read();
            if let Some(idx) = w.country_indices.iter().find(|i| i.symbol == sym) {
                idx.price
            } else if let Some(ws) = w.world_stocks.iter().find(|ws| ws.ticker == sym) {
                ws.price
            } else if let Some(fx) = w.forex_pairs.iter().find(|fx| fx.symbol == sym) {
                fx.rate
            } else {
                100.0
            }
        }
    };

    let sim_start_time = (chrono::Utc::now().timestamp_millis() / 60000) * 60000;
    state.tf_manager.register_symbol(&sym);
    state.tf_manager.with_symbol(&sym, |stb| {
        stb.init_live_candles(&sym, current_price, sim_start_time);
    });

    let generated = state.tf_manager.get_bars(&sym, &tf, limit).unwrap_or_default();
    Json(generated).into_response()
}

#[derive(Deserialize)]
struct SpeedRequest {
    multiplier: f64,
}

async fn speed_handler(State(state): State<AppState>, Json(req): Json<SpeedRequest>) -> impl IntoResponse {
    let mut mult = state.speed_multiplier.write();
    if req.multiplier > 0.0 {
        *mult = req.multiplier;
    }
    Json(serde_json::json!({"speed_multiplier": *mult}))
}

fn format_editorial_event(ev: &crate::storage::models::EventRecord) -> (String, String, String, String, String, String) {
    let details_parsed: serde_json::Value = serde_json::from_str(&ev.details).unwrap_or(serde_json::Value::Null);
    let kind_lower = ev.kind.to_lowercase();
    let sym = if ev.symbol.is_empty() { "MARKET" } else { &ev.symbol };

    if kind_lower.contains("macro") {
        let headline = details_parsed["macro_headline"]
            .as_str()
            .unwrap_or("Central Bank Announces Benchmark Monetary Policy Assessment");
        (
            headline.to_string(),
            "Sovereign Central Wire".to_string(),
            "Central Bank".to_string(),
            "BREAKING".to_string(),
            "Neutral".to_string(),
            format!("The Central Bank delivered its statutory macroeconomic assessment regarding inflation rates, sovereign yield dynamics, and economic momentum: {}", headline),
        )
    } else if kind_lower.contains("fundamental") {
        let mult = details_parsed["multiplier"].as_f64().unwrap_or(1.0);
        let f_kind = details_parsed["fundamental_kind"].as_str().unwrap_or("earnings");
        let sentiment = if mult >= 1.0 { "Bullish" } else { "Bearish" };
        let urgency = if (mult - 1.0).abs() > 0.15 { "HIGH IMPACT" } else { "ROUTINE" };
        let headline = match f_kind {
            "GoodNewsJump" | "good_news" => format!("{}: Commercial Breakthrough & Margin Expansion Disclosed", sym),
            "BadNewsJump" | "bad_news" => format!("{}: Supply Disruption & Quarterly Margin Headwinds Announced", sym),
            "MajorPositiveJump" => format!("{}: Landmark Global Contract & Strategic Revenue Surge", sym),
            "MajorNegativeJump" => format!("{}: Regulatory Scrutiny & Asset Impairment Notice Issued", sym),
            _ => format!("{}: Material Intrinsic Valuation Revision Recorded", sym),
        };
        (
            headline,
            "Dow Jones Corporate Wire".to_string(),
            "Corporate".to_string(),
            urgency.to_string(),
            sentiment.to_string(),
            format!("Audited commercial filings triggered institutional re-evaluations for {}, adjusting intrinsic equity valuation by {:.1}%.", sym, (mult - 1.0) * 100.0),
        )
    } else if kind_lower.contains("restatement") {
        (
            format!("{}: Audited Balance Sheet Recalculation & Restatement", sym),
            "Financial Standards Gazette".to_string(),
            "Corporate".to_string(),
            "BULLETIN".to_string(),
            "Bearish".to_string(),
            format!("Financial controllers for {} submitted retroactive accounting adjustments across reported asset values.", sym),
        )
    } else if kind_lower.contains("bailout") {
        let detail = details_parsed["distress_details"].as_str().unwrap_or("Emergency liquidity facility secured");
        (
            format!("{}: Emergency Syndicate Credit Facility Backstop Secured", sym),
            "Restructuring & Recovery Wire".to_string(),
            "Corporate".to_string(),
            "BREAKING".to_string(),
            "Bullish".to_string(),
            detail.to_string(),
        )
    } else if kind_lower.contains("acquisition") {
        let detail = details_parsed["distress_details"].as_str().unwrap_or("Consortium buyout offer submitted");
        (
            format!("{}: Institutional Buyout Syndicate Proposes Premium Acquisition", sym),
            "M&A Syndicate Bulletin".to_string(),
            "Corporate".to_string(),
            "BREAKING".to_string(),
            "Bullish".to_string(),
            detail.to_string(),
        )
    } else if kind_lower.contains("reverse_split") || kind_lower.contains("reversessplit") {
        (
            format!("{}: Capital Restructuring Reverse Split Executed", sym),
            "Exchange Compliance Directorate".to_string(),
            "Corporate".to_string(),
            "ROUTINE".to_string(),
            "Neutral".to_string(),
            format!("Share consolidation completed for {} to maintain exchange listing compliance and optimize float distribution.", sym),
        )
    } else if kind_lower.contains("liquidation") {
        (
            format!("Clearing House Margin Enforcement Executed on {}", sym),
            "Exchange Clearing Monitor".to_string(),
            "Distress".to_string(),
            "HIGH IMPACT".to_string(),
            "Bearish".to_string(),
            format!("Automated risk surveillance mandated liquidation of insolvent participant positions in {} following statutory collateral threshold breaches.", sym),
        )
    } else if kind_lower.contains("ipo") {
        let summary = details_parsed["macro_headline"]
            .as_str()
            .or_else(|| details_parsed["distress_details"].as_str())
            .map(|s| s.to_string())
            .unwrap_or_else(|| format!("Pre-IPO enterprise {} officially chartered on sovereign exchange floor, entering public secondary trading.", sym));
        (
            format!("{}: Initial Public Offering Completed", sym),
            "Sovereign Listing Syndicate".to_string(),
            "Corporate".to_string(),
            "BREAKING".to_string(),
            "Bullish".to_string(),
            summary,
        )
    } else {
        let parsed_msg = details_parsed["macro_headline"]
            .as_str()
            .or_else(|| details_parsed["message"].as_str())
            .or_else(|| details_parsed["text"].as_str())
            .or_else(|| details_parsed["distress_details"].as_str())
            .unwrap_or_else(|| if ev.details.starts_with('{') { "Official sovereign market dispatch issued by regulatory overseers." } else { &ev.details });
        (
            format!("Market Wire: {}", ev.kind),
            "Inter-State Press Syndicate".to_string(),
            "Macro".to_string(),
            "ROUTINE".to_string(),
            "Neutral".to_string(),
            parsed_msg.to_string(),
        )
    }
}

async fn events_handler(State(state): State<AppState>) -> impl IntoResponse {
    let raw_events = state.db.get_recent_events(50).unwrap_or_default();
    let mut editorial_news: Vec<EditorialNewsDTO> = Vec::with_capacity(raw_events.len() + 10);

    for ev in &raw_events {
        let (headline, wire_source, category, urgency, sentiment, summary) = format_editorial_event(ev);
        editorial_news.push(EditorialNewsDTO {
            id: ev.id,
            tick: ev.tick,
            timestamp: ev.timestamp,
            headline,
            wire_source,
            category,
            urgency,
            sentiment,
            summary,
            affected_symbol: if ev.symbol.is_empty() || ev.symbol == "SYSTEM" { None } else { Some(ev.symbol.clone()) },
        });
    }

    let world = state.world.read();
    for pol in world.policy_log.iter().rev().take(10) {
        editorial_news.push(EditorialNewsDTO {
            id: 80000 + pol.tick,
            tick: pol.tick,
            timestamp: chrono::Utc::now().timestamp_millis(),
            headline: format!("Diplomatic Wire: {}", pol.action),
            wire_source: "Inter-State Diplomatic Herald".to_string(),
            category: "Geopolitical".to_string(),
            urgency: "HIGH IMPACT".to_string(),
            sentiment: if pol.action.contains("Sanction") { "Bearish".to_string() } else { "Bullish".to_string() },
            summary: pol.detail.clone(),
            affected_symbol: Some(pol.country_id.clone()),
        });
    }

    editorial_news.sort_by(|a, b| b.tick.cmp(&a.tick));

    Json(editorial_news)
}

#[derive(Deserialize)]
struct IpoRequest {
    sector: Option<String>,
    cap_tier: Option<String>,
    revenue: Option<f64>,
}

async fn ipo_handler(State(state): State<AppState>, Json(req): Json<IpoRequest>) -> impl IntoResponse {
    let mut sim = state.sim.write();
    let tick = *state.current_tick.read();
    let sector = match req.sector.as_deref() {
        Some("Energy") => Sector::Energy,
        Some("Materials") => Sector::Materials,
        Some("Industrials") => Sector::Industrials,
        Some("ConsumerDiscretionary") => Sector::ConsumerDiscretionary,
        Some("ConsumerStaples") => Sector::ConsumerStaples,
        Some("HealthCare") => Sector::HealthCare,
        Some("Financials") => Sector::Financials,
        Some("CommunicationServices") => Sector::CommunicationServices,
        Some("Utilities") => Sector::Utilities,
        Some("RealEstate") => Sector::RealEstate,
        _ => Sector::InformationTechnology,
    };
    let cap_tier = match req.cap_tier.as_deref() {
        Some("SmallCap") => crate::company::taxonomy::CapTier::SmallCap,
        Some("MidCap") => crate::company::taxonomy::CapTier::MidCap,
        Some("LargeCap") => crate::company::taxonomy::CapTier::LargeCap,
        _ => crate::company::taxonomy::CapTier::MegaCap,
    };
    let revenue = req.revenue.unwrap_or(10_000_000_000.0);
    let mut used_symbols = std::collections::HashSet::new();
    for co_lock in state.companies.read().iter() {
        used_symbols.insert(co_lock.read().symbol.clone());
    }

    let mut rng = rand::rngs::StdRng::seed_from_u64(rand::random());
    let ipo_co = crate::company::company::generate_ipo_company(
        sector,
        cap_tier,
        revenue,
        tick as usize,
        &mut rng,
        &mut used_symbols,
    );

    let ipo_symbol = ipo_co.symbol.clone();
    let ipo_price = ipo_co.ipo_price;
    let co_clone = ipo_co.clone();

    sim.list_ipo(co_clone, 50_000_000.0);
    sim.add_participant(state.human_trader.clone(), &[&ipo_symbol]);
    state.companies.write().push(Arc::new(RwLock::new(ipo_co.clone())));
    state.tf_manager.register_symbol(&ipo_symbol);

    Json(serde_json::json!({
        "status": "listed",
        "symbol": ipo_symbol,
        "name": ipo_co.name,
        "sector": ipo_co.sector.to_string(),
        "cap_tier": ipo_co.cap_tier.to_string(),
        "annual_revenue": ipo_co.annual_revenue,
        "net_margin": ipo_co.net_margin,
        "market_cap": ipo_co.market_cap(),
        "shares_outstanding": ipo_co.shares_outstanding,
        "ipo_price": ipo_price,
        "ipo_tick": tick,
    }))
}

async fn book_handler(State(state): State<AppState>, Query(q): Query<SymbolQuery>) -> impl IntoResponse {
    let sym = match q.symbol {
        Some(s) if !s.is_empty() => s,
        _ => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": "symbol parameter required"}))).into_response(),
    };

    let sim = state.sim.read();
    if let Some(book) = sim.book(&sym) {
        let (bids, asks) = book.top_levels(15);
        let mid = book.mid_price();
        let spread = book.spread().unwrap_or(0.0);

        Json(serde_json::json!({
            "symbol": sym,
            "has_mid": mid.is_some(),
            "mid": mid.unwrap_or(0.0),
            "spread": spread,
            "bids": bids,
            "asks": asks,
        })).into_response()
    } else {
        (StatusCode::NOT_FOUND, Json(serde_json::json!({"error": "symbol not found"}))).into_response()
    }
}

#[derive(Deserialize)]
struct OrderRequest {
    symbol: String,
    side: String,
    #[serde(rename = "type")]
    order_type: Option<String>,
    price: f64,
    quantity: f64,
}

async fn orders_handler(State(state): State<AppState>, Json(req): Json<OrderRequest>) -> impl IntoResponse {
    if req.symbol.is_empty() || req.quantity <= 0.0 {
        return (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": "symbol and positive quantity required"}))).into_response();
    }

    let side = if req.side.to_uppercase() == "SELL" { Side::Sell } else { Side::Buy };
    let is_market = req.order_type.as_deref() == Some("MARKET") || req.price <= 0.0;

    // Check if symbol is a domestic equity listed in simulation order book
    let is_domestic = {
        state.sim.read().markets.contains_key(&req.symbol)
    };

    if is_domestic {
        let order = Order {
            id: 0,
            agent_id: "human_trader".to_string(),
            side,
            price: req.price,
            quantity: req.quantity,
            is_market,
        };

        let mut trader = state.human_trader.write();
        trader.submit_order_for_symbol(&req.symbol, order);

        return Json(serde_json::json!({
            "status": "queued",
            "symbol": req.symbol,
            "side": req.side,
            "quantity": req.quantity,
            "price": req.price,
            "type": if is_market { "MARKET" } else { "LIMIT" }
        })).into_response();
    }

    // NON-DOMESTIC ASSET (Forex pair, World stock, Country Index, Bond):
    // Execute immediately against live world quote!
    let exec_price = {
        let w = state.world.read();
        let sym_norm = norm_sym(&req.symbol);
        if let Some(fx) = w.forex_pairs.iter().find(|p| norm_sym(&p.symbol) == sym_norm) {
            Some(fx.rate)
        } else if let Some(ws) = w.world_stocks.iter().find(|s| norm_sym(&s.ticker) == sym_norm) {
            Some(ws.price)
        } else if let Some(idx) = w.country_indices.iter().find(|i| norm_sym(&i.symbol) == sym_norm) {
            Some(idx.price)
        } else {
            let sim = state.sim.read();
            sim.national.bonds.iter().find(|b| norm_sym(&b.id) == sym_norm).map(|b| b.price)
        }
    }.unwrap_or(if req.price > 0.0 { req.price } else { 100.0 });

    let trader = state.human_trader.read();
    let mut acct = trader.account.write();
    crate::market::apply_settled_fill(&mut acct, &req.symbol, side, req.quantity, exec_price);

    Json(serde_json::json!({
        "status": "filled",
        "symbol": req.symbol,
        "side": req.side,
        "quantity": req.quantity,
        "price": exec_price,
        "type": if is_market { "MARKET" } else { "LIMIT" }
    })).into_response()
}

pub fn build_portfolio_dto(state: &AppState) -> PortfolioDTO {
    let trader = state.human_trader.read();
    let acct = trader.account.read();

    let mut current_prices = HashMap::new();
    for (sym, q) in state.latest_prices.read().iter() {
        if q.has_mid {
            current_prices.insert(sym.clone(), q.mid);
        }
    }
    {
        let w = state.world.read();
        for fx in &w.forex_pairs {
            current_prices.insert(fx.symbol.clone(), fx.rate);
            let unslashed = fx.symbol.replace("/", "");
            current_prices.insert(unslashed, fx.rate);
        }
        for ws in &w.world_stocks {
            current_prices.insert(ws.ticker.clone(), ws.price);
        }
        for idx in &w.country_indices {
            current_prices.insert(idx.symbol.clone(), idx.price);
        }
        let sim = state.sim.read();
        for b in &sim.national.bonds {
            current_prices.insert(b.id.clone(), b.price);
        }
    }

    let equity = acct.equity(&current_prices);
    let buying_power = acct.available_buying_power(&current_prices);
    let margin_ratio = acct.margin_ratio(&current_prices).unwrap_or(0.0);

    let mut positions = Vec::new();
    for (sym, pos) in &acct.positions {
        let entry_price = if pos.quantity > 0.0 { pos.entry_cost / pos.quantity } else { 0.0 };
        let cur_price = *current_prices.get(sym).unwrap_or(&entry_price);
        let side_str = match pos.side {
            crate::market::margin::PositionSide::Long => "LONG",
            crate::market::margin::PositionSide::Short => "SHORT",
        };
        let pnl = pos.unrealized_pnl(cur_price);
        let pnl_pct = if pos.entry_cost > 0.0 { (pnl / pos.entry_cost) * 100.0 } else { 0.0 };

        positions.push(PositionDTO {
            symbol: sym.clone(),
            side: side_str.to_string(),
            quantity: pos.quantity,
            entry_price,
            current_price: cur_price,
            value: pos.quantity * cur_price,
            unrealized_pnl: pnl,
            pnl_percent: pnl_pct,
        });
    }

    PortfolioDTO {
        cash: acct.cash,
        equity,
        buying_power,
        margin_ratio,
        maintenance_margin_ratio: acct.maintenance_margin_ratio,
        max_leverage: acct.max_leverage,
        positions,
    }
}

async fn portfolio_handler(State(state): State<AppState>) -> impl IntoResponse {
    Json(build_portfolio_dto(&state))
}

#[derive(Deserialize)]
struct ClosePositionRequest {
    symbol: String,
}

async fn close_position_handler(State(state): State<AppState>, Json(req): Json<ClosePositionRequest>) -> impl IntoResponse {
    let trader = state.human_trader.read();
    let mut acct = trader.account.write();

    if let Some(pos) = acct.positions.get(&req.symbol) {
        let is_domestic = {
            state.sim.read().markets.contains_key(&req.symbol)
        };

        let close_side = match pos.side {
            crate::market::margin::PositionSide::Short => Side::Buy,
            crate::market::margin::PositionSide::Long => Side::Sell,
        };
        let qty = pos.quantity;

        if is_domestic {
            drop(acct);
            drop(trader);
            let mut trader_mut = state.human_trader.write();
            let order = Order {
                id: 0,
                agent_id: "human_trader".to_string(),
                side: close_side,
                price: 0.0,
                quantity: qty,
                is_market: true,
            };
            trader_mut.submit_order_for_symbol(&req.symbol, order);

            return Json(serde_json::json!({
                "status": "close_submitted",
                "symbol": req.symbol,
                "quantity": qty,
            })).into_response();
        }

        // Off-exchange instrument: close immediately at current quote
        let sym_norm = norm_sym(&req.symbol);
        let exec_price = {
            let w = state.world.read();
            if let Some(fx) = w.forex_pairs.iter().find(|p| norm_sym(&p.symbol) == sym_norm) {
                Some(fx.rate)
            } else if let Some(ws) = w.world_stocks.iter().find(|s| norm_sym(&s.ticker) == sym_norm) {
                Some(ws.price)
            } else if let Some(idx) = w.country_indices.iter().find(|i| norm_sym(&i.symbol) == sym_norm) {
                Some(idx.price)
            } else {
                let sim = state.sim.read();
                sim.national.bonds.iter().find(|b| norm_sym(&b.id) == sym_norm).map(|b| b.price)
            }
        }.unwrap_or(if pos.quantity > 0.0 { pos.entry_cost / pos.quantity } else { 100.0 });

        crate::market::apply_settled_fill(&mut acct, &req.symbol, close_side, qty, exec_price);

        Json(serde_json::json!({
            "status": "closed",
            "symbol": req.symbol,
            "quantity": qty,
            "price": exec_price,
        })).into_response()
    } else {
        (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": "no open position for symbol"}))).into_response()
    }
}

async fn portfolio_settings_handler(
    State(state): State<AppState>,
    Json(req): Json<PortfolioSettingsRequest>,
) -> impl IntoResponse {
    let trader = state.human_trader.read();
    let mut acct = trader.account.write();

    if let Some(bal) = req.balance {
        if bal >= 0.0 {
            acct.cash = bal;
        }
    }

    if let Some(lev) = req.leverage {
        if lev >= 1.0 && lev <= 100.0 {
            acct.max_leverage = lev;
        }
    }

    let cash = acct.cash;
    let leverage = acct.max_leverage;
    drop(acct);
    drop(trader);

    Json(serde_json::json!({
        "status": "success",
        "cash": cash,
        "max_leverage": leverage,
        "portfolio": build_portfolio_dto(&state),
    })).into_response()
}

async fn login_handler() -> impl IntoResponse {
    Json(serde_json::json!({
        "user_id": "human_trader",
        "name": "Retail Trader",
        "account_id": "human_trader",
        "starting_capital": 100000,
        "max_leverage": 5
    }))
}

async fn signup_handler() -> impl IntoResponse {
    login_handler().await
}

pub fn build_macro_dto(state: &AppState) -> MacroDTO {
    let sim = state.sim.read();
    let nat = &sim.national;
    let cit_rep = &sim.citizen_report;
    let nat_rep = &sim.national_report;

    let (current_tier, tier_name, tier_progress) = nat.current_tier();
    let tier_rank = |t: &str| match t {
        "powerhouse" => 3,
        "mid" => 2,
        _ => 1,
    };
    let curr_rank = tier_rank(current_tier);

    let all_policies = crate::country::fiscal::all_known_policies();
    let policies = all_policies
        .into_iter()
        .map(|p| {
            let active = nat.fiscal.is_policy_active(&p.id);
            let is_locked = tier_rank(&p.min_tier) > curr_rank;
            let days_active = *sim.national.fiscal.policy_rollouts.get(&p.id).unwrap_or(&0);
            let efficacy = if active { (days_active as f64 / p.rollout_days.max(1) as f64).clamp(0.15, 1.0) } else { 0.0 };
            PolicyItemDTO {
                id: p.id,
                category: p.category,
                name: p.name,
                description: p.description,
                favored_sector: p.favored_sector,
                annual_cost: p.annual_cost,
                active,
                min_tier: p.min_tier,
                is_locked,
                rollout_days: p.rollout_days,
                days_active,
                efficacy,
                direct_positive: p.direct_positive,
                direct_negative: p.direct_negative,
                indirect_positive: p.indirect_positive,
                indirect_negative: p.indirect_negative,
            }
        })
        .collect();

    let (c_day, c_hour, c_year, c_formatted) = compute_calendar_str(*state.sim_hour.read());
    MacroDTO {
        tick: sim.tick,
        sim_hour: c_hour,
        sim_day: c_day,
        sim_year: c_year,
        sim_date_formatted: c_formatted,
        gdp: nat_rep.gdp,
        real_gdp_growth: nat_rep.real_gdp_growth,
        central_bank_rate: nat_rep.fed_funds_rate,
        ten_year_yield: nat_rep.ten_year_yield,
        inflation_rate: nat_rep.cpi_inflation_rate,
        policy_stance: nat_rep.policy_stance.clone(),
        unemployment_rate: nat_rep.unemployment_rate,
        labor_force: nat_rep.labor_force,
        employed_workers: nat_rep.employed_workers,
        average_hourly_wage: nat_rep.average_hourly_wage,
        consumer_sentiment: cit_rep.consumer_confidence,
        consumer_spending: cit_rep.consumer_spending,
        national_debt: nat_rep.national_debt,
        federal_revenue: nat_rep.federal_revenue,
        federal_outlays: nat_rep.federal_outlays,
        budget_deficit: nat_rep.budget_deficit,
        credit_rating: nat_rep.credit_rating.clone(),
        treasury_cash: nat_rep.treasury_cash,
        borrowing_yield: nat_rep.borrowing_yield,
        dxy_index: nat_rep.dollar_index_dxy,
        trade_balance: nat_rep.trade_balance,
        maintenance_spending_rate: nat_rep.maintenance_spending_rate,
        depreciation_rate: nat_rep.depreciation_rate,
        tourism_revenue: nat_rep.tourism_revenue,
        productivity_index: nat_rep.productivity_index,
        bonds: nat_rep.bonds.clone(),
        tier: current_tier.to_string(),
        tier_name: tier_name.to_string(),
        tier_progress,
        policies,
    }
}

async fn macro_handler(State(state): State<AppState>) -> impl IntoResponse {
    Json(build_macro_dto(&state))
}

#[derive(Deserialize)]
struct PolicyToggleRequest {
    policy_id: String,
}

async fn policy_toggle_handler(
    State(state): State<AppState>,
    Json(req): Json<PolicyToggleRequest>,
) -> impl IntoResponse {
    let mut sim = state.sim.write();
    match sim.toggle_policy(&req.policy_id) {
        Ok((new_state, name)) => {
            Json(serde_json::json!({
                "policy_id": req.policy_id,
                "name": name,
                "active": new_state,
            })).into_response()
        }
        Err(err) => {
            (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": err }))).into_response()
        }
    }
}

pub fn build_nation_dto(state: &AppState) -> NationDTO {
    let world = state.world.read();
    let sim = state.sim.read();
    let (c_day, c_hour, c_year, c_formatted) = compute_calendar_str(*state.sim_hour.read());
    let fiscal = &sim.national.fiscal;
    let (current_tier, tier_name, tier_progress) = sim.national.current_tier();

    NationDTO {
        name: if world.home.name.is_empty() { "Republic of Eldoria".to_string() } else { world.home.name.clone() },
        currency: world.home.currency.clone(),
        flag_code: world.home.flag_code.clone(),
        founded: world.home.founded,
        configured: world.home.configured,
        infrastructure_level: fiscal.infrastructure_level,
        healthcare_level: fiscal.healthcare_level,
        education_level: fiscal.education_level,
        enterprise_grants_level: fiscal.enterprise_grants_level,
        export_capacity_level: fiscal.export_capacity_level,
        exchange_chartered: fiscal.exchange_chartered,
        treasury_cash: fiscal.treasury_cash,
        reserve_currency_reserves: fiscal.reserve_currency_reserves,
        national_debt: fiscal.national_debt,
        credit_rating: fiscal.credit_rating.clone(),
        borrowing_yield: fiscal.borrowing_yield,
        maintenance_spending_rate: fiscal.maintenance_spending_rate,
        depreciation_rate: fiscal.depreciation_rate,
        tourism_revenue: fiscal.tourism_revenue,
        productivity_index: fiscal.productivity_index,
        bonds: sim.national.bonds.clone(),
        loan_requests: world.loan_requests.clone(),
        tier: current_tier.to_string(),
        tier_name: tier_name.to_string(),
        tier_progress,
        regions: world.home.regions.clone(),
        population: 18_500_000.0,
        gdp_per_capita: sim.national.nominal_gdp / 18_500_000.0,
        labor_force: sim.national.labor.labor_force,
        employed_workers: sim.national.labor.employed_workers,
        average_hourly_wage: sim.national.labor.average_hourly_wage,
        job_openings: (sim.national.labor.labor_force * 0.048).max(10_000.0),
        unemployment_rate: sim.national.labor.unemployment_rate,
        geopolitical_power: match current_tier {
            "powerhouse" => 88.0,
            "mid" => 56.0,
            _ => 28.0,
        },
        personal_income_tax_rate: fiscal.personal_income_tax_rate,
        corporate_tax_rate: fiscal.corporate_tax_rate,
        sales_tax_rate: fiscal.sales_tax_rate,
        foreign_debt_custody_enabled: fiscal.foreign_debt_custody_enabled,
        foreign_debt_custody_nation: fiscal.foreign_debt_custody_nation.clone(),
        foreign_debt_custody_amount: fiscal.foreign_debt_custody_amount,
        sim_hour: c_hour,
        sim_day: c_day,
        sim_year: c_year,
        sim_date_formatted: c_formatted,
    }
}

async fn borrow_capital_handler(
    State(state): State<AppState>,
    Json(req): Json<BorrowCapitalRequest>,
) -> impl IntoResponse {
    let mut sim = state.sim.write();
    match sim.national.borrow_sovereign_capital(req.amount) {
        Ok(msg) => {
            let fiscal = &sim.national.fiscal;
            Json(serde_json::json!({
                "status": "success",
                "message": msg,
                "amount": req.amount,
                "treasury_cash": fiscal.treasury_cash,
                "national_debt": fiscal.national_debt,
                "credit_rating": fiscal.credit_rating,
                "borrowing_yield": fiscal.borrowing_yield,
            })).into_response()
        }
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": e }))).into_response(),
    }
}

async fn repay_debt_handler(
    State(state): State<AppState>,
    Json(req): Json<RepayDebtRequest>,
) -> impl IntoResponse {
    let mut sim = state.sim.write();
    let gdp = sim.national.nominal_gdp;
    let infl = sim.national.central_bank.cpi_inflation_rate;
    let fiscal = &mut sim.national.fiscal;
    match fiscal.repay_debt(req.amount) {
        Ok(()) => {
            fiscal.update_credit_rating(gdp, infl);
            Json(serde_json::json!({
                "status": "success",
                "message": format!("Successfully paid down ${:.2}M in sovereign debt", req.amount / 1e6),
                "treasury_cash": fiscal.treasury_cash,
                "national_debt": fiscal.national_debt,
                "credit_rating": fiscal.credit_rating,
                "borrowing_yield": fiscal.borrowing_yield,
            })).into_response()
        }
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": e }))).into_response(),
    }
}

async fn trade_reserve_currency_handler(
    State(state): State<AppState>,
    Json(req): Json<TradeReserveCurrencyRequest>,
) -> impl IntoResponse {
    let rate = {
        let w = state.world.read();
        w.forex_pairs.iter()
            .find(|fx| fx.quote_currency == "ZTH")
            .map(|fx| fx.rate)
            .unwrap_or(2.45)
    };

    let mut sim = state.sim.write();
    let gdp = sim.national.nominal_gdp;
    let infl = sim.national.central_bank.cpi_inflation_rate;
    let fiscal = &mut sim.national.fiscal;

    if req.side.to_lowercase() == "buy" {
        match fiscal.buy_reserve_currency(req.amount, rate) {
            Ok(zth_bought) => {
                fiscal.update_credit_rating(gdp, infl);
                Json(serde_json::json!({
                    "status": "success",
                    "side": "buy",
                    "crn_spent": req.amount,
                    "zth_acquired": zth_bought,
                    "rate": rate,
                    "treasury_cash": fiscal.treasury_cash,
                    "reserve_currency_reserves": fiscal.reserve_currency_reserves,
                })).into_response()
            }
            Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": e }))).into_response(),
        }
    } else {
        match fiscal.sell_reserve_currency(req.amount, rate) {
            Ok(crn_proceeds) => {
                fiscal.update_credit_rating(gdp, infl);
                Json(serde_json::json!({
                    "status": "success",
                    "side": "sell",
                    "zth_sold": req.amount,
                    "crn_proceeds": crn_proceeds,
                    "rate": rate,
                    "treasury_cash": fiscal.treasury_cash,
                    "reserve_currency_reserves": fiscal.reserve_currency_reserves,
                })).into_response()
            }
            Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": e }))).into_response(),
        }
    }
}

async fn get_nation_handler(State(state): State<AppState>) -> impl IntoResponse {
    Json(build_nation_dto(&state))
}

async fn found_nation_handler(
    State(state): State<AppState>,
    Json(req): Json<FoundNationRequest>,
) -> impl IntoResponse {
    let mut world = state.world.write();
    world.home.configure(&req.name, &req.currency, &req.flag_code, req.founded);
    Json(serde_json::json!({
        "status": "success",
        "name": world.home.name,
        "currency": world.home.currency,
        "flag_code": world.home.flag_code,
        "founded": world.home.founded,
    }))
}

async fn invest_nation_handler(
    State(state): State<AppState>,
    Json(req): Json<InvestRequest>,
) -> impl IntoResponse {
    let mut sim = state.sim.write();
    let fiscal = &mut sim.national.fiscal;

    let res = match req.pillar.to_lowercase().as_str() {
        "infrastructure" => fiscal.invest_infrastructure(req.amount),
        "healthcare" => fiscal.invest_population(req.amount, "healthcare"),
        "education" => fiscal.invest_population(req.amount, "education"),
        "enterprise" => fiscal.invest_enterprise_grants(req.amount),
        _ => Err("invalid pillar: choose infrastructure, healthcare, education, or enterprise"),
    };

    match res {
        Ok(()) => Json(serde_json::json!({
            "status": "success",
            "pillar": req.pillar,
            "amount": req.amount,
            "infrastructure_level": fiscal.infrastructure_level,
            "healthcare_level": fiscal.healthcare_level,
            "education_level": fiscal.education_level,
            "enterprise_grants_level": fiscal.enterprise_grants_level,
            "treasury_cash": fiscal.treasury_cash,
        })).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({ "error": e })),
        ).into_response(),
    }
}

async fn charter_exchange_handler(State(state): State<AppState>) -> impl IntoResponse {
    let mut sim = state.sim.write();
    match sim.national.fiscal.charter_exchange() {
        Ok(()) => Json(serde_json::json!({
            "status": "chartered",
            "exchange_chartered": true,
        })).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({ "error": e })),
        ).into_response(),
    }
}

async fn set_maintenance_handler(
    State(state): State<AppState>,
    Json(req): Json<MaintenanceRequest>,
) -> impl IntoResponse {
    let mut sim = state.sim.write();
    match sim.national.fiscal.set_maintenance_rate(req.rate) {
        Ok(()) => Json(serde_json::json!({
            "status": "success",
            "maintenance_spending_rate": sim.national.fiscal.maintenance_spending_rate,
        })).into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({ "error": e })),
        ).into_response(),
    }
}

async fn loan_respond_handler(
    State(state): State<AppState>,
    Json(req): Json<LoanRespondRequest>,
) -> impl IntoResponse {
    if req.action.to_lowercase() == "accept" {
        let mut world = state.world.write();
        match world.accept_loan_request(&req.request_id) {
            Ok((amount, country_id, country_name)) => {
                let mut sim = state.sim.write();
                sim.national.fiscal.treasury_cash = (sim.national.fiscal.treasury_cash - amount).max(0.0);
                Json(serde_json::json!({
                    "status": "accepted",
                    "request_id": req.request_id,
                    "country_id": country_id,
                    "country_name": country_name,
                    "amount": amount,
                    "treasury_cash": sim.national.fiscal.treasury_cash,
                })).into_response()
            }
            Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": e }))).into_response(),
        }
    } else {
        let mut world = state.world.write();
        match world.reject_loan_request(&req.request_id) {
            Ok(country_name) => Json(serde_json::json!({
                "status": "declined",
                "request_id": req.request_id,
                "country_name": country_name,
            })).into_response(),
            Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": e }))).into_response(),
        }
    }
}

async fn contract_respond_handler(
    State(state): State<AppState>,
    Json(req): Json<ContractRespondRequest>,
) -> impl IntoResponse {
    let mut world = state.world.write();
    let action_norm = req.action.to_lowercase();
    match world.respond_bilateral_contract(&req.contract_id, &action_norm) {
        Ok((country_id, country_name, rev_gain, export_boost)) => {
            if action_norm == "accept" && export_boost > 0.0 {
                let mut sim = state.sim.write();
                sim.national.fiscal.export_capacity_level = (sim.national.fiscal.export_capacity_level * (1.0 + export_boost)).min(10.0);
            }
            Json(serde_json::json!({
                "status": "success",
                "action": action_norm,
                "contract_id": req.contract_id,
                "country_id": country_id,
                "country_name": country_name,
                "annual_revenue_gain": rev_gain,
                "export_capacity_boost": export_boost,
            })).into_response()
        }
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": e }))).into_response(),
    }
}

async fn get_bonds_handler(State(state): State<AppState>) -> impl IntoResponse {
    let sim = state.sim.read();
    Json(sim.national.bonds.clone())
}

async fn issue_bond_handler(
    State(state): State<AppState>,
    Json(req): Json<BondActionRequest>,
) -> impl IntoResponse {
    let mut sim = state.sim.write();
    match sim.national.issue_bond(req.maturity_years, req.amount) {
        Ok(()) => Json(serde_json::json!({
            "status": "success",
            "maturity_years": req.maturity_years,
            "amount": req.amount,
            "bonds": sim.national.bonds,
            "national_debt": sim.national.fiscal.national_debt,
            "treasury_cash": sim.national.fiscal.treasury_cash,
        })).into_response(),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": e }))).into_response(),
    }
}

async fn buyback_bond_handler(
    State(state): State<AppState>,
    Json(req): Json<BondActionRequest>,
) -> impl IntoResponse {
    let mut sim = state.sim.write();
    match sim.national.buyback_bond(req.maturity_years, req.amount) {
        Ok(()) => Json(serde_json::json!({
            "status": "success",
            "maturity_years": req.maturity_years,
            "amount": req.amount,
            "bonds": sim.national.bonds,
            "national_debt": sim.national.fiscal.national_debt,
            "treasury_cash": sim.national.fiscal.treasury_cash,
        })).into_response(),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": e }))).into_response(),
    }
}

async fn get_forex_handler(State(state): State<AppState>) -> impl IntoResponse {
    let world = state.world.read();
    Json(world.forex_pairs.clone())
}

async fn forex_swap_handler(
    State(state): State<AppState>,
    Json(req): Json<ForexSwapRequest>,
) -> impl IntoResponse {
    let world = state.world.read();
    if let Some(fx) = world.forex_pairs.iter().find(|p| p.symbol == req.pair) {
        let rate = fx.rate;
        let converted = if req.side.to_lowercase() == "buy" {
            req.amount * rate
        } else {
            req.amount / rate
        };
        Json(serde_json::json!({
            "status": "success",
            "pair": req.pair,
            "side": req.side,
            "amount": req.amount,
            "rate": rate,
            "converted": converted,
        })).into_response()
    } else {
        (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": "forex pair not found" }))).into_response()
    }
}

async fn get_world_stocks_handler(State(state): State<AppState>) -> impl IntoResponse {
    let world = state.world.read();
    Json(world.world_stocks.clone())
}

async fn get_world_handler(State(state): State<AppState>) -> impl IntoResponse {
    let world = state.world.read();
    let countries: Vec<ForeignCountryDTO> = world
        .foreign_countries_sorted()
        .into_iter()
        .map(|c| ForeignCountryDTO {
            id: c.id.clone(),
            name: c.name.clone(),
            currency: c.currency.clone(),
            flag_code: c.flag_code.clone(),
            gdp: c.gdp,
            gdp_growth: c.gdp_growth,
            inflation: c.inflation,
            interest_rate: c.interest_rate,
            unemployment: c.unemployment,
            population: c.population,
            trade_balance: c.trade_balance,
            tariff_rate: c.relation.tariff_rate,
            stance: c.relation.stance.to_string(),
            trade_volume: c.relation.trade_volume,
            sanctions_level: c.relation.sanctions_level,
            aid_flow: c.relation.aid_flow,
            currency_strength: c.currency_strength,
            national_debt: c.national_debt,
            debt_to_gdp: c.debt_to_gdp,
            labor_force: c.labor_force,
            employed_workers: c.employed_workers,
            gdp_per_capita: c.gdp_per_capita,
            average_hourly_wage: c.average_hourly_wage,
            job_openings: c.job_openings,
            productivity_index: c.productivity_index,
            geopolitical_power: c.geopolitical_power,
            sanctioned_by_us: c.sanctioned_by_us,
            sanctioned_us: c.sanctioned_us,
            sanctions_summary: c.sanctions_summary.clone(),
            is_reserve_currency: c.is_reserve_currency,
        })
        .collect();

    let sim = state.sim.read();
    let world_dto = WorldDTO {
        countries,
        forex_pairs: world.forex_pairs.clone(),
        world_stocks: world.world_stocks.clone(),
        country_indices: world.country_indices.clone(),
        loan_requests: world.loan_requests.clone(),
        bilateral_contracts: world.bilateral_contracts.clone(),
        policy_log: world.policy_log.clone(),
        reserve_currency: world.reserve_currency.clone(),
        reserve_country_name: world.reserve_country_name.clone(),
        reserve_currency_reserves: sim.national.fiscal.reserve_currency_reserves,
    };

    Json(world_dto)
}

async fn get_world_indices_handler(State(state): State<AppState>) -> impl IntoResponse {
    let world = state.world.read();
    Json(world.country_indices.clone())
}

async fn sanction_toggle_handler(
    State(state): State<AppState>,
    Json(req): Json<SanctionToggleRequest>,
) -> impl IntoResponse {
    let mut world = state.world.write();
    let sim = state.sim.read();
    let (tier, _, _) = sim.national.current_tier();
    let home_power = match tier {
        "powerhouse" => 88.0,
        "mid" => 56.0,
        _ => 28.0,
    };

    let result = if req.action.to_lowercase() == "lift" {
        world.lift_sanction(&req.country_id)
    } else {
        world.sanction_country(&req.country_id, home_power)
    };

    match result {
        Ok((action, detail)) => Json(serde_json::json!({
            "status": "success",
            "action": action,
            "detail": detail,
            "country_id": req.country_id,
        })).into_response(),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": e }))).into_response(),
    }
}

async fn diplomacy_handler(
    State(state): State<AppState>,
    Json(req): Json<DiplomacyRequest>,
) -> impl IntoResponse {
    let mut world = state.world.write();
    let tick = *state.current_tick.read();

    let ok = match req.action.to_lowercase().as_str() {
        "sanctions" => {
            let level = req.level.unwrap_or(1);
            world.impose_sanctions(&req.country_id, level, tick)
        }
        "deal" => {
            let new_tariff = req.tariff.unwrap_or(0.02);
            world.negotiate_deal(&req.country_id, new_tariff, tick)
        }
        _ => false,
    };

    if ok {
        Json(serde_json::json!({ "status": "success", "action": req.action, "country_id": req.country_id })).into_response()
    } else {
        (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({ "error": "unknown action or country" })),
        ).into_response()
    }
}

#[derive(Deserialize)]
struct TaxesRequest {
    personal_income_tax_rate: Option<f64>,
    corporate_tax_rate: Option<f64>,
    sales_tax_rate: Option<f64>,
}

async fn set_taxes_handler(
    State(state): State<AppState>,
    Json(req): Json<TaxesRequest>,
) -> impl IntoResponse {
    let mut sim = state.sim.write();
    sim.national.fiscal.set_tax_rates(req.personal_income_tax_rate, req.corporate_tax_rate, req.sales_tax_rate);
    sim.citizen.taxation.effective_income_tax_rate = sim.national.fiscal.personal_income_tax_rate;
    sim.citizen.taxation.sales_tax_rate = sim.national.fiscal.sales_tax_rate;

    Json(serde_json::json!({
        "status": "success",
        "personal_income_tax_rate": sim.national.fiscal.personal_income_tax_rate,
        "corporate_tax_rate": sim.national.fiscal.corporate_tax_rate,
        "sales_tax_rate": sim.national.fiscal.sales_tax_rate,
    }))
}

#[derive(Deserialize)]
struct BilateralBorrowRequest {
    #[serde(alias = "lender_id")]
    country_id: String,
    amount: f64,
    interest_rate: Option<f64>,
}

async fn borrow_bilateral_handler(
    State(state): State<AppState>,
    Json(req): Json<BilateralBorrowRequest>,
) -> impl IntoResponse {
    let mut sim = state.sim.write();
    let rate = req.interest_rate.unwrap_or(0.045);
    match sim.national.fiscal.borrow_bilateral(&req.country_id, req.amount, rate) {
        Ok(msg) => {
            Json(serde_json::json!({
                "status": "success",
                "message": msg,
                "treasury_cash": sim.national.fiscal.treasury_cash,
                "national_debt": sim.national.fiscal.national_debt,
                "borrowing_yield": sim.national.fiscal.borrowing_yield,
            })).into_response()
        }
        Err(err) => {
            (StatusCode::BAD_REQUEST, Json(serde_json::json!({ "error": err }))).into_response()
        }
    }
}

#[derive(Deserialize)]
struct DebtCustodyRequest {
    #[serde(alias = "country_id")]
    nation_id: String,
    #[serde(alias = "enable")]
    enabled: bool,
}

async fn toggle_debt_custody_handler(
    State(state): State<AppState>,
    Json(req): Json<DebtCustodyRequest>,
) -> impl IntoResponse {
    let mut sim = state.sim.write();
    let (enabled, amount) = sim.national.fiscal.toggle_foreign_debt_custody(&req.nation_id, req.enabled);
    Json(serde_json::json!({
        "status": "success",
        "enabled": enabled,
        "nation_id": req.nation_id,
        "custody_amount": amount,
    }))
}
