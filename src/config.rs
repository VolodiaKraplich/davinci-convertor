use crate::cli::Args;
use crate::types::{Codec, ConversionMode};
use anyhow::{Context, Result};
use colored::Colorize;
use std::path::PathBuf;

#[derive(Debug, Clone)]
pub struct Config {
  pub path: PathBuf,
  pub output_dir: Option<PathBuf>,
  pub mode: ConversionMode,
  pub codec: Codec,
  pub quality: String,
  pub verbose: bool,
  pub force: bool,
  pub workers: usize,
  pub dry_run: bool,
}

impl Config {
  pub fn from_args(args: Args) -> Result<Self> {
    let path = args
      .path
      .context("Input path is required for the main application logic.")?;

    let config = Self {
      path,
      output_dir: args.output_dir,
      mode: args.mode,
      codec: args.codec,
      quality: args.quality,
      verbose: args.verbose,
      force: args.force,
      workers: args.workers.max(1),
      dry_run: args.dry_run,
    };

    config.validate()?;
    Ok(config)
  }

  fn validate(&self) -> Result<()> {
    // Validate quality profiles
    if self.mode == ConversionMode::Editing {
      let valid_qualities = match self.codec {
        Codec::DnxHR => &["lb", "sq", "hq", "hqx"][..],
        Codec::ProRes => &["proxy", "lt", "standard", "hq"][..],
      };

      if !valid_qualities.contains(&self.quality.as_str()) {
        anyhow::bail!(
          "Invalid quality '{}' for {:?}. Valid options: {}",
          self.quality,
          self.codec,
          valid_qualities.join(", ")
        );
      }
    }

    Ok(())
  }

  pub fn check_dependencies() -> Result<()> {
    for tool in &["ffmpeg", "ffprobe"] {
      which::which(tool).with_context(|| {
        format!(
          "{} {} is not installed. Please install ffmpeg.",
          "✗".red(),
          tool
        )
      })?;
    }
    Ok(())
  }

  pub fn print_header(&self) {
    println!("{}", "========================================".cyan());
    println!("{}", "  DaVinci Converter".cyan().bold());
    println!("{}", "========================================".cyan());

    match self.mode {
      ConversionMode::Export => {
        println!(
          "{} {} | {} {}",
          "Mode:".yellow(),
          "Export".bold(),
          "Workers:".yellow(),
          self.workers
        );
      }
      ConversionMode::Editing => {
        println!(
          "{} {} | {} {:?} | {} {} | {} {}",
          "Mode:".yellow(),
          "Editing".bold(),
          "Codec:".yellow(),
          self.codec,
          "Quality:".yellow(),
          self.quality,
          "Workers:".yellow(),
          self.workers
        );
      }
    }

    if let Some(ref output_dir) = self.output_dir {
      println!("{} {}", "Output directory:".yellow(), output_dir.display());
    }

    if self.dry_run {
      println!(
        "{}",
        "🔍 DRY RUN MODE - No files will be converted"
          .yellow()
          .bold()
      );
    }

    if self.force {
      println!(
        "{}",
        "⚠️  Force mode enabled - existing files will be overwritten".yellow()
      );
    }

    println!("{}", "----------------------------------------".cyan());
  }
}
