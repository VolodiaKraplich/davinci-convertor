mod cli;
mod config;
mod converter;
mod ffmpeg;
mod file_scanner;
mod stats;
mod types;

use anyhow::Result;
use tracing_subscriber::{EnvFilter, fmt, prelude::*};

#[tokio::main]
async fn main() -> Result<()> {
  // Initialize logging
  tracing_subscriber::registry()
    .with(fmt::layer())
    .with(EnvFilter::from_default_env().add_directive(tracing::Level::INFO.into()))
    .init();

  cli::handle_cli().await?;

  Ok(())
}
