use std::collections::HashSet;
use rand::Rng;
use super::taxonomy::Sector;

struct NameParts {
    prefixes: &'static [&'static str],
    roots: &'static [&'static str],
    suffixes: &'static [&'static str],
}

fn sector_name_parts(sector: Sector) -> NameParts {
    match sector {
        Sector::Energy => NameParts {
            prefixes: &["Petro", "Basin", "Ridge", "Delta", "Summit", "Anchor"],
            roots: &["Field", "Well", "Reserve", "Pipeline", "Crest"],
            suffixes: &["Energy", "Resources", "Petroleum", "Holdings"],
        },
        Sector::Materials => NameParts {
            prefixes: &["Iron", "Granite", "Copper", "Continental", "Bedrock"],
            roots: &["Mining", "Alloy", "Mineral", "Quarry"],
            suffixes: &["Materials", "Industries", "Corp", "Group"],
        },
        Sector::Industrials => NameParts {
            prefixes: &["Union", "Atlas", "Precision", "Ironclad", "Vanguard"],
            roots: &["Machine", "Works", "Fabrication", "Systems"],
            suffixes: &["Industries", "Manufacturing", "Corp", "Holdings"],
        },
        Sector::ConsumerDiscretionary => NameParts {
            prefixes: &["Urban", "Bright", "Luxe", "Metro", "Pacific"],
            roots: &["Retail", "Goods", "Style", "Living"],
            suffixes: &["Brands", "Group", "Co.", "Holdings"],
        },
        Sector::ConsumerStaples => NameParts {
            prefixes: &["Harvest", "Golden", "Homestead", "Heritage", "Clearwater"],
            roots: &["Foods", "Provisions", "Pantry", "Mills"],
            suffixes: &["Foods", "Brands", "Co.", "Group"],
        },
        Sector::HealthCare => NameParts {
            prefixes: &["Vita", "Gene", "Nova", "Helix", "Cortex"],
            roots: &["Health", "Bio", "Therapeutics", "Med"],
            suffixes: &["Health", "Sciences", "Therapeutics", "Labs"],
        },
        Sector::Financials => NameParts {
            prefixes: &["Sterling", "Meridian", "Trust", "Capital", "Ledger"],
            roots: &["Financial", "Trust", "Capital", "Credit"],
            suffixes: &["Financial", "Group", "Holdings", "Partners"],
        },
        Sector::InformationTechnology => NameParts {
            prefixes: &["Nex", "Cyber", "Byte", "Logic", "Quantum"],
            roots: &["Soft", "Data", "Systems", "Cloud"],
            suffixes: &["Technologies", "Systems", "Software", "Inc."],
        },
        Sector::CommunicationServices => NameParts {
            prefixes: &["Signal", "Relay", "Beacon", "Pulse", "Echo"],
            roots: &["Media", "Network", "Stream", "Connect"],
            suffixes: &["Media", "Communications", "Networks", "Group"],
        },
        Sector::Utilities => NameParts {
            prefixes: &["Northern", "Coastal", "Public", "Regional", "Continental"],
            roots: &["Power", "Grid", "Water", "Light"],
            suffixes: &["Utilities", "Power", "Energy", "Co."],
        },
        Sector::RealEstate => NameParts {
            prefixes: &["Cornerstone", "Skyline", "Landmark", "Harborview", "Prairie"],
            roots: &["Realty", "Properties", "Estates", "Trust"],
            suffixes: &["Realty", "Properties", "REIT", "Trust"],
        },
    }
}

pub fn generate_name<R: Rng + ?Sized>(sector: Sector, rng: &mut R) -> String {
    let parts = sector_name_parts(sector);
    let prefix = parts.prefixes[rng.gen_range(0..parts.prefixes.len())];
    let suffix = parts.suffixes[rng.gen_range(0..parts.suffixes.len())];

    if rng.gen::<f64>() < 0.4 {
        let root = parts.roots[rng.gen_range(0..parts.roots.len())];
        format!("{} {} {}", prefix, root, suffix)
    } else {
        format!("{} {}", prefix, suffix)
    }
}

pub fn generate_symbol(name: &str, used_symbols: &mut HashSet<String>) -> String {
    let words: Vec<&str> = name.split_whitespace().collect();
    let mut initials = String::new();
    for w in &words {
        if let Some(c) = w.chars().next() {
            initials.push(c);
        }
    }

    let mut symbol = initials.to_uppercase();
    if symbol.len() > 4 {
        symbol.truncate(4);
    }
    if symbol.len() < 2 && !words.is_empty() {
        let first = words[0].to_uppercase();
        for c in first.chars() {
            if symbol.len() >= 3 {
                break;
            }
            symbol.push(c);
        }
    }

    let base = symbol;
    let mut candidate = base.clone();
    let mut suffix_num = 1;
    while used_symbols.contains(&candidate) {
        suffix_num += 1;
        candidate = format!("{}{}", base, suffix_num);
    }
    used_symbols.insert(candidate.clone());
    candidate
}
