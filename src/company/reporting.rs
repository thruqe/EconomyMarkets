use rand::Rng;
use serde::{Deserialize, Serialize};
use super::taxonomy::{get_sector_profile, Sector};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum ReportingProfile {
    Honest,
    OptimisticBias,
    ConservativeBias,
    HighUncertainty,
    Fraudulent,
}

impl ReportingProfile {
    pub fn as_str(&self) -> &'static str {
        match self {
            ReportingProfile::Honest => "Honest",
            ReportingProfile::OptimisticBias => "Optimistic Bias",
            ReportingProfile::ConservativeBias => "Conservative Bias",
            ReportingProfile::HighUncertainty => "High Uncertainty",
            ReportingProfile::Fraudulent => "Fraudulent",
        }
    }
}

impl std::fmt::Display for ReportingProfile {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.as_str())
    }
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
pub struct ReportingParams {
    pub profile: ReportingProfile,
    pub noise_std_dev: f64,
    pub persistent_bias: f64,
    pub restatement_probability: f64,
    pub restatement_severity_min: f64,
    pub restatement_severity_max: f64,
}

impl Default for ReportingParams {
    fn default() -> Self {
        Self {
            profile: ReportingProfile::Honest,
            noise_std_dev: 0.02,
            persistent_bias: 0.0,
            restatement_probability: 1.0 / 4000.0,
            restatement_severity_min: 0.2,
            restatement_severity_max: 0.5,
        }
    }
}

pub fn choose_reporting_profile<R: Rng + ?Sized>(sector: Sector, rng: &mut R) -> ReportingProfile {
    let sp = get_sector_profile(sector);

    let weights = [
        (ReportingProfile::Honest, 0.55),
        (ReportingProfile::OptimisticBias, 0.20),
        (ReportingProfile::ConservativeBias, 0.15),
        (ReportingProfile::HighUncertainty, 0.08 * sp.high_uncertainty_weight),
        (ReportingProfile::Fraudulent, 0.02),
    ];

    let total: f64 = weights.iter().map(|(_, w)| *w).sum();
    let r = rng.gen::<f64>() * total;
    let mut cumulative = 0.0;
    for (p, w) in weights {
        cumulative += w;
        if r <= cumulative {
            return p;
        }
    }
    ReportingProfile::Honest
}

pub fn reporting_params_for<R: Rng + ?Sized>(profile: ReportingProfile, rng: &mut R) -> ReportingParams {
    match profile {
        ReportingProfile::Honest => ReportingParams {
            profile: ReportingProfile::Honest,
            noise_std_dev: 0.01 + rng.gen::<f64>() * 0.02,
            persistent_bias: 0.0,
            restatement_probability: 1.0 / 4000.0,
            restatement_severity_min: 0.2,
            restatement_severity_max: 0.5,
        },
        ReportingProfile::OptimisticBias => ReportingParams {
            profile: ReportingProfile::OptimisticBias,
            noise_std_dev: 0.015 + rng.gen::<f64>() * 0.02,
            persistent_bias: 0.00006 + rng.gen::<f64>() * 0.00014,
            restatement_probability: 1.0 / 4000.0,
            restatement_severity_min: 0.2,
            restatement_severity_max: 0.5,
        },
        ReportingProfile::ConservativeBias => ReportingParams {
            profile: ReportingProfile::ConservativeBias,
            noise_std_dev: 0.015 + rng.gen::<f64>() * 0.02,
            persistent_bias: -(0.00006 + rng.gen::<f64>() * 0.00014),
            restatement_probability: 1.0 / 4000.0,
            restatement_severity_min: 0.2,
            restatement_severity_max: 0.5,
        },
        ReportingProfile::HighUncertainty => ReportingParams {
            profile: ReportingProfile::HighUncertainty,
            noise_std_dev: 0.05 + rng.gen::<f64>() * 0.08,
            persistent_bias: 0.0,
            restatement_probability: 1.0 / 2500.0,
            restatement_severity_min: 0.3,
            restatement_severity_max: 0.7,
        },
        ReportingProfile::Fraudulent => ReportingParams {
            profile: ReportingProfile::Fraudulent,
            noise_std_dev: 0.01 + rng.gen::<f64>() * 0.02,
            persistent_bias: 0.00015 + rng.gen::<f64>() * 0.00025,
            restatement_probability: 1.0 / 6000.0,
            restatement_severity_min: 0.85,
            restatement_severity_max: 1.0,
        },
    }
}
