use serde::{Deserialize, Serialize};

/// Sector follows the standard GICS sector taxonomy.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum Sector {
    Energy,
    Materials,
    Industrials,
    ConsumerDiscretionary,
    ConsumerStaples,
    HealthCare,
    Financials,
    InformationTechnology,
    CommunicationServices,
    Utilities,
    RealEstate,
}

impl Sector {
    pub fn as_str(&self) -> &'static str {
        match self {
            Sector::Energy => "Energy",
            Sector::Materials => "Materials",
            Sector::Industrials => "Industrials",
            Sector::ConsumerDiscretionary => "Consumer Discretionary",
            Sector::ConsumerStaples => "Consumer Staples",
            Sector::HealthCare => "Health Care",
            Sector::Financials => "Financials",
            Sector::InformationTechnology => "Information Technology",
            Sector::CommunicationServices => "Communication Services",
            Sector::Utilities => "Utilities",
            Sector::RealEstate => "Real Estate",
        }
    }
}

impl std::fmt::Display for Sector {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.as_str())
    }
}

pub const ALL_SECTORS: [Sector; 11] = [
    Sector::Energy,
    Sector::Materials,
    Sector::Industrials,
    Sector::ConsumerDiscretionary,
    Sector::ConsumerStaples,
    Sector::HealthCare,
    Sector::Financials,
    Sector::InformationTechnology,
    Sector::CommunicationServices,
    Sector::Utilities,
    Sector::RealEstate,
];

/// CapTier is a market-capitalization tier.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum CapTier {
    MegaCap,
    LargeCap,
    MidCap,
    SmallCap,
}

impl CapTier {
    pub fn as_str(&self) -> &'static str {
        match self {
            CapTier::MegaCap => "Mega Cap",
            CapTier::LargeCap => "Large Cap",
            CapTier::MidCap => "Mid Cap",
            CapTier::SmallCap => "Small Cap",
        }
    }
}

impl std::fmt::Display for CapTier {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.as_str())
    }
}

#[derive(Debug, Clone, Copy)]
pub struct SectorProfile {
    pub drift_min: f64,
    pub drift_max: f64,
    pub vol_min: f64,
    pub vol_max: f64,
    pub jump_risk_multiplier: f64,
    pub high_uncertainty_weight: f64,
    pub ps_multiple_min: f64,
    pub ps_multiple_max: f64,
    pub net_margin_min: f64,
    pub net_margin_max: f64,
}

pub fn get_sector_profile(sector: Sector) -> SectorProfile {
    match sector {
        Sector::Energy => SectorProfile {
            drift_min: 0.00001, drift_max: 0.00005, vol_min: 0.0008, vol_max: 0.0018,
            jump_risk_multiplier: 1.1, high_uncertainty_weight: 0.9,
            ps_multiple_min: 1.2, ps_multiple_max: 2.5, net_margin_min: 0.08, net_margin_max: 0.18,
        },
        Sector::Materials => SectorProfile {
            drift_min: 0.00001, drift_max: 0.00005, vol_min: 0.0007, vol_max: 0.0016,
            jump_risk_multiplier: 1.0, high_uncertainty_weight: 0.8,
            ps_multiple_min: 1.2, ps_multiple_max: 2.5, net_margin_min: 0.07, net_margin_max: 0.16,
        },
        Sector::Industrials => SectorProfile {
            drift_min: 0.00002, drift_max: 0.00006, vol_min: 0.0007, vol_max: 0.0015,
            jump_risk_multiplier: 0.9, high_uncertainty_weight: 0.6,
            ps_multiple_min: 1.5, ps_multiple_max: 3.0, net_margin_min: 0.08, net_margin_max: 0.15,
        },
        Sector::ConsumerDiscretionary => SectorProfile {
            drift_min: 0.00002, drift_max: 0.00006, vol_min: 0.0008, vol_max: 0.0017,
            jump_risk_multiplier: 1.1, high_uncertainty_weight: 0.9,
            ps_multiple_min: 1.5, ps_multiple_max: 3.5, net_margin_min: 0.06, net_margin_max: 0.15,
        },
        Sector::ConsumerStaples => SectorProfile {
            drift_min: 0.00002, drift_max: 0.00005, vol_min: 0.0004, vol_max: 0.0009,
            jump_risk_multiplier: 0.6, high_uncertainty_weight: 0.3,
            ps_multiple_min: 1.5, ps_multiple_max: 2.8, net_margin_min: 0.07, net_margin_max: 0.13,
        },
        Sector::HealthCare => SectorProfile {
            drift_min: 0.00002, drift_max: 0.00007, vol_min: 0.0008, vol_max: 0.0018,
            jump_risk_multiplier: 1.2, high_uncertainty_weight: 1.8,
            ps_multiple_min: 3.5, ps_multiple_max: 7.5, net_margin_min: 0.12, net_margin_max: 0.25,
        },
        Sector::Financials => SectorProfile {
            drift_min: 0.00002, drift_max: 0.00005, vol_min: 0.0006, vol_max: 0.0014,
            jump_risk_multiplier: 1.0, high_uncertainty_weight: 0.6,
            ps_multiple_min: 2.0, ps_multiple_max: 4.5, net_margin_min: 0.15, net_margin_max: 0.28,
        },
        Sector::InformationTechnology => SectorProfile {
            drift_min: 0.00003, drift_max: 0.00008, vol_min: 0.0009, vol_max: 0.0020,
            jump_risk_multiplier: 1.3, high_uncertainty_weight: 1.5,
            ps_multiple_min: 4.5, ps_multiple_max: 9.5, net_margin_min: 0.18, net_margin_max: 0.32,
        },
        Sector::CommunicationServices => SectorProfile {
            drift_min: 0.00002, drift_max: 0.00006, vol_min: 0.0007, vol_max: 0.0016,
            jump_risk_multiplier: 1.1, high_uncertainty_weight: 0.9,
            ps_multiple_min: 2.5, ps_multiple_max: 5.0, net_margin_min: 0.10, net_margin_max: 0.22,
        },
        Sector::Utilities => SectorProfile {
            drift_min: 0.00002, drift_max: 0.00006, vol_min: 0.0005, vol_max: 0.0010,
            jump_risk_multiplier: 0.4, high_uncertainty_weight: 0.2,
            ps_multiple_min: 1.5, ps_multiple_max: 2.8, net_margin_min: 0.08, net_margin_max: 0.14,
        },
        Sector::RealEstate => SectorProfile {
            drift_min: 0.00002, drift_max: 0.00005, vol_min: 0.0005, vol_max: 0.0012,
            jump_risk_multiplier: 0.7, high_uncertainty_weight: 0.4,
            ps_multiple_min: 3.0, ps_multiple_max: 6.0, net_margin_min: 0.15, net_margin_max: 0.30,
        },
    }
}

#[derive(Debug, Clone, Copy)]
pub struct CapTierProfile {
    pub vol_multiplier: f64,
    pub jump_multiplier: f64,
    pub shares_outstanding_min: f64,
    pub shares_outstanding_max: f64,
    pub float_fraction_min: f64,
    pub float_fraction_max: f64,
    pub revenue_min: f64,
    pub revenue_max: f64,
}

pub fn get_cap_tier_profile(tier: CapTier) -> CapTierProfile {
    match tier {
        CapTier::MegaCap => CapTierProfile {
            vol_multiplier: 0.75, jump_multiplier: 0.6,
            shares_outstanding_min: 50_000_000.0, shares_outstanding_max: 150_000_000.0,
            float_fraction_min: 0.70, float_fraction_max: 0.90,
            revenue_min: 500_000_000.0, revenue_max: 1_500_000_000.0,
        },
        CapTier::LargeCap => CapTierProfile {
            vol_multiplier: 0.9, jump_multiplier: 0.8,
            shares_outstanding_min: 20_000_000.0, shares_outstanding_max: 60_000_000.0,
            float_fraction_min: 0.65, float_fraction_max: 0.85,
            revenue_min: 150_000_000.0, revenue_max: 500_000_000.0,
        },
        CapTier::MidCap => CapTierProfile {
            vol_multiplier: 1.15, jump_multiplier: 1.2,
            shares_outstanding_min: 5_000_000.0, shares_outstanding_max: 25_000_000.0,
            float_fraction_min: 0.50, float_fraction_max: 0.80,
            revenue_min: 40_000_000.0, revenue_max: 150_000_000.0,
        },
        CapTier::SmallCap => CapTierProfile {
            vol_multiplier: 1.5, jump_multiplier: 1.8,
            shares_outstanding_min: 1_000_000.0, shares_outstanding_max: 8_000_000.0,
            float_fraction_min: 0.35, float_fraction_max: 0.75,
            revenue_min: 5_000_000.0, revenue_max: 40_000_000.0,
        },
    }
}
