use colored::Colorize;
use std::sync::Arc;
use std::sync::atomic::{AtomicUsize, Ordering};
use std::time::Instant;

#[derive(Clone)]
pub struct Stats {
  total: Arc<AtomicUsize>,
  success: Arc<AtomicUsize>,
  failed: Arc<AtomicUsize>,
  skipped: Arc<AtomicUsize>,
  start_time: Instant,
}

impl Stats {
  pub fn new(total: usize) -> Self {
    Self {
      total: Arc::new(AtomicUsize::new(total)),
      success: Arc::new(AtomicUsize::new(0)),
      failed: Arc::new(AtomicUsize::new(0)),
      skipped: Arc::new(AtomicUsize::new(0)),
      start_time: Instant::now(),
    }
  }

  pub fn inc_success(&self) {
    self.success.fetch_add(1, Ordering::Relaxed);
  }

  pub fn inc_failed(&self) {
    self.failed.fetch_add(1, Ordering::Relaxed);
  }

  pub fn inc_skipped(&self) {
    self.skipped.fetch_add(1, Ordering::Relaxed);
  }

  pub fn get_counts(&self) -> (usize, usize, usize, usize) {
    (
      self.total.load(Ordering::Relaxed),
      self.success.load(Ordering::Relaxed),
      self.skipped.load(Ordering::Relaxed),
      self.failed.load(Ordering::Relaxed),
    )
  }

  pub fn print_summary(&self) {
    let elapsed = self.start_time.elapsed();
    let (total, success, skipped, failed) = self.get_counts();

    println!("\n{}", "========================================".cyan());
    println!("{} {:.1?}", "✓ Completed in".green().bold(), elapsed);
    println!(
      "{} {} | {} {} | {} {} | {} {}",
      "Total:".cyan(),
      total,
      "Success:".green(),
      success,
      "Skipped:".yellow(),
      skipped,
      "Failed:".red(),
      failed
    );
    println!("{}", "========================================".cyan());
  }
}
