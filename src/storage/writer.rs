use std::time::Duration;
use tokio::sync::{mpsc, oneshot};
use tokio::task::JoinHandle;

use crate::storage::db::DB;
use crate::storage::models::*;

enum WriterMsg {
    Batch(Batch),
    Flush(oneshot::Sender<Result<(), String>>),
}

/// AsyncWriter provides decoupled, asynchronous batch writing to SQLite.
#[derive(Clone)]
pub struct AsyncWriter {
    sender: mpsc::Sender<WriterMsg>,
}

impl AsyncWriter {
    /// Start a new AsyncWriter with the given DB, channel buffer capacity, and periodic flush interval.
    pub fn new(db: DB, buffer_cap: usize, flush_interval: Duration) -> (Self, JoinHandle<()>) {
        let cap = if buffer_cap == 0 { 1000 } else { buffer_cap };
        let (tx, mut rx) = mpsc::channel::<WriterMsg>(cap);

        let handle = tokio::spawn(async move {
            let mut interval = tokio::time::interval(flush_interval);
            let mut pending_batches: Vec<Batch> = Vec::new();
            let max_batch_size = 500;

            let flush_pending = |db: &DB, pending: &mut Vec<Batch>| -> Result<(), String> {
                if pending.is_empty() {
                    return Ok(());
                }
                let mut merged = Batch::default();
                for b in pending.drain(..) {
                    merged.trades.extend(b.trades);
                    merged.prices.extend(b.prices);
                    merged.candles.extend(b.candles);
                    merged.events.extend(b.events);
                    merged.accounts.extend(b.accounts);
                }
                db.commit_batch(&merged).map_err(|e| e.to_string())
            };

            loop {
                tokio::select! {
                    _ = interval.tick() => {
                        let _ = flush_pending(&db, &mut pending_batches);
                    }
                    msg = rx.recv() => {
                        match msg {
                            Some(WriterMsg::Batch(b)) => {
                                if !b.is_empty() {
                                    pending_batches.push(b);
                                    if pending_batches.len() >= max_batch_size {
                                        let _ = flush_pending(&db, &mut pending_batches);
                                    }
                                }
                            }
                            Some(WriterMsg::Flush(reply)) => {
                                let res = flush_pending(&db, &mut pending_batches);
                                let _ = reply.send(res);
                            }
                            None => {
                                // Channel closed, flush remaining and exit
                                let _ = flush_pending(&db, &mut pending_batches);
                                break;
                            }
                        }
                    }
                }
            }
        });

        (Self { sender: tx }, handle)
    }

    /// Enqueue sends a batch to the asynchronous ingestion channel.
    /// Returns false if the queue is full or closed.
    pub fn enqueue(&self, batch: Batch) -> bool {
        if batch.is_empty() {
            return true;
        }
        self.sender.try_send(WriterMsg::Batch(batch)).is_ok()
    }

    /// Flush triggers an immediate commit of pending batches and waits for completion.
    pub async fn flush(&self) -> Result<(), String> {
        let (tx, rx) = oneshot::channel();
        if self.sender.send(WriterMsg::Flush(tx)).await.is_err() {
            return Ok(());
        }
        rx.await.map_err(|_| "Writer task closed".to_string())?
    }
}
