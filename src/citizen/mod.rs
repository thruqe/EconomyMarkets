pub mod citizen;
pub mod consumer;
pub mod demographics;
pub mod sentiment;
pub mod taxation;

pub use citizen::{CitizenEconomy, CitizenReport};
pub use consumer::ConsumerEngine;
pub use demographics::Demographics;
pub use sentiment::Sentiment;
pub use taxation::Taxation;
