use crate::config::Config;
use crate::ffmpeg::FFmpeg;
use crate::file_scanner::FileScanner;
use crate::stats::Stats;
use crate::types::Action;
use anyhow::{Context, Result};
use colored::Colorize;
use std::path::{Path, PathBuf};
use std::sync::Arc;
use tokio::fs;
use tokio::sync::Semaphore;
use tracing::error;

pub struct Converter {
  config: Config,
}

impl Converter {
  pub const fn new(config: Config) -> Self {
    Self { config }
  }

  pub async fn run(&self) -> Result<()> {
    self.config.print_header();

    let files = FileScanner::scan(&self.config.path).await?;

    if files.is_empty() {
      println!("{}", "⚠️  Warning: No media files found".yellow());
      return Ok(());
    }

    let stats = Stats::new(files.len());
    let semaphore = Arc::new(Semaphore::new(self.config.workers));

    println!(
      "{} {} {}",
      "🚀 Starting conversion with".magenta(),
      self.config.workers,
      "workers...".magenta()
    );
    println!();

    let mut tasks = Vec::new();

    for file in files {
      let permit = semaphore.clone().acquire_owned().await?;
      let config = self.config.clone();
      let stats = stats.clone();

      let task = tokio::spawn(async move {
        Self::process_file(&file, &config, &stats).await;
        drop(permit);
      });

      tasks.push(task);
    }

    for task in tasks {
      task.await?;
    }

    stats.print_summary();

    Ok(())
  }

  async fn process_file(file: &Path, config: &Config, stats: &Stats) {
    let file_name = file.file_name().unwrap().to_string_lossy().to_string();

    let result: Result<()> = async {
      let action = FFmpeg::analyze(file, config)
        .await
        .context("Analysis failed")?;

      let output_path = Self::get_output_path(file, config);

      if !config.force && fs::metadata(&output_path).await.is_ok() {
        anyhow::bail!("File exists (use --force to overwrite)");
      }

      if action == Action::Skip {
        stats.inc_skipped();
        println!(
          "{} {} {}",
          "✓ Skipped:".green(),
          file_name,
          "(already compatible)".dimmed()
        );
        return Ok(());
      }

      if config.dry_run {
        println!("{} {} {:?}", "🔍 Would process:".cyan(), file_name, action);
        stats.inc_success();
        return Ok(());
      }

      // Create output directory
      if let Some(parent) = output_path.parent() {
        fs::create_dir_all(parent)
          .await
          .context("Failed to create output directory")?;
      }

      FFmpeg::convert(file, &output_path, action, config).await?;

      stats.inc_success();
      println!("{} {}", "→ Processed:".cyan(), file_name);

      Ok(())
    }
    .await;

    if let Err(e) = result {
      stats.inc_failed();
      error!("Failed to process {}: {}", file_name, e);
      eprintln!("{} {}: {}", "✗ Error:".red(), file_name, e);
    }
  }

  fn get_output_path(file: &Path, config: &Config) -> PathBuf {
    let parent = file.parent().unwrap();
    let dir = config.output_dir.as_ref().map_or(parent, |v| v);

    let stem = file.file_stem().unwrap().to_string_lossy();

    let (suffix, ext) = match config.mode {
      crate::types::ConversionMode::Export => ("_export", "mp4"),
      crate::types::ConversionMode::Editing => ("_converted", "mov"),
    };

    dir.join(format!("{stem}{suffix}.{ext}"))
  }
}
