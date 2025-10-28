use crate::config::Config;
use crate::converter::Converter;
use crate::types::{Codec, ConversionMode};
use anyhow::Result;
use clap::{CommandFactory, Parser, Subcommand};
use clap_complete::Shell;
use std::io;
use std::path::PathBuf;

/// Universal converter for DaVinci Resolve and video export
#[derive(Parser, Debug, Clone)]
#[command(name = "davinci-convertor")]
#[command(author, version = env!("CARGO_PKG_VERSION"), about, long_about = None)]
#[command(arg_required_else_help = true)]
pub struct Args {
  #[command(subcommand)]
  pub command: Option<Commands>,

  /// Input file or directory to process
  // The requirement is now handled manually in `handle_cli`.
  #[arg(value_name = "PATH")]
  pub path: Option<PathBuf>,

  /// Output directory for converted files (defaults to input directory)
  #[arg(short = 'o', long, value_name = "DIR")]
  pub output_dir: Option<PathBuf>,

  /// Conversion mode: 'editing' or 'export'
  #[arg(long, default_value = "editing", value_name = "MODE")]
  pub mode: ConversionMode,

  /// Codec for editing mode: 'dnxhr' or 'prores'
  #[arg(long, default_value = "dnxhr", value_name = "CODEC")]
  pub codec: Codec,

  /// Quality profile for editing: 'hq', 'hqx', 'sq', 'lb'
  #[arg(long, default_value = "hq", value_name = "QUALITY")]
  pub quality: String,

  /// Enable verbose output with ffmpeg logs
  #[arg(short, long)]
  pub verbose: bool,

  /// Force overwrite of existing output files
  #[arg(short, long)]
  pub force: bool,

  /// Number of parallel conversion workers
  #[arg(short, long, default_value_t = num_cpus::get(), value_name = "N")]
  pub workers: usize,

  /// Dry run - analyze files without converting
  #[arg(long)]
  pub dry_run: bool,
}

#[derive(Subcommand, Debug, Clone)]
pub enum Commands {
  /// Generate shell completion scripts
  Completion {
    #[arg(value_enum)]
    shell: Shell,
  },
}

/// Parses arguments and runs the appropriate application logic.
pub async fn handle_cli() -> Result<()> {
  Config::check_dependencies()?;

  let args = Args::parse();

  if let Some(Commands::Completion { shell }) = args.command {
    let mut cmd = Args::command();
    let name = cmd.get_name().to_string();
    clap_complete::generate(shell, &mut cmd, name, &mut io::stdout());
    return Ok(());
  }

  if args.path.is_some() {
    let config = Config::from_args(args)?;
    let converter = Converter::new(config);
    converter.run().await?;
  } else {
    Args::command().print_help()?;
  }

  Ok(())
}
