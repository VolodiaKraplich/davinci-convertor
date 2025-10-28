use crate::config::Config;
use crate::types::{Action, Codec, ConversionMode, FFProbeOutput};
use anyhow::{Context, Result};
use std::path::Path;
use tokio::process::Command;
use tracing::{debug, trace};

pub struct FFmpeg;

impl FFmpeg {
  pub async fn probe(file: &Path) -> Result<FFProbeOutput> {
    trace!("Probing file: {}", file.display());

    let output = Command::new("ffprobe")
      .args([
        "-v",
        "quiet",
        "-print_format",
        "json",
        "-show_streams",
        &file.to_string_lossy(),
      ])
      .output()
      .await
      .context("Failed to run ffprobe")?;

    if !output.status.success() {
      anyhow::bail!("ffprobe failed with status: {}", output.status);
    }

    serde_json::from_slice(&output.stdout).context("Failed to parse ffprobe output")
  }

  pub async fn analyze(file: &Path, config: &Config) -> Result<Action> {
    let probe = Self::probe(file).await?;

    let video_stream = probe
      .streams
      .iter()
      .find(|s| s.codec_type == "video")
      .ok_or_else(|| anyhow::anyhow!("No video stream found"))?;

    let action = match config.mode {
      ConversionMode::Export => {
        if video_stream.codec_name == "h264"
          && file.extension().and_then(|e| e.to_str()) == Some("mp4")
        {
          Action::Skip
        } else {
          Action::Convert
        }
      }
      ConversionMode::Editing => {
        let target_codec = Self::get_target_codec_name(config.codec);
        if video_stream.codec_name == target_codec {
          if file.extension().and_then(|e| e.to_str()) == Some("mov") {
            Action::Skip
          } else {
            Action::Rewrap
          }
        } else {
          Action::Convert
        }
      }
    };

    debug!("Analysis result for {}: {:?}", file.display(), action);
    Ok(action)
  }

  pub async fn convert(
    file: &Path,
    output_path: &Path,
    action: Action,
    config: &Config,
  ) -> Result<()> {
    let mut cmd = Command::new("ffmpeg");
    cmd.args(["-y", "-i", &file.to_string_lossy()]);

    match action {
      Action::Rewrap => {
        cmd.args([
          "-c",
          "copy",
          "-map_metadata",
          "0",
          &output_path.to_string_lossy(),
        ]);
      }
      Action::Convert => match config.mode {
        ConversionMode::Export => {
          Self::add_export_params(&mut cmd, output_path);
        }
        ConversionMode::Editing => {
          Self::add_editing_params(&mut cmd, output_path, config);
        }
      },
      Action::Skip => unreachable!("Skip action should not reach conversion"),
    }

    if !config.verbose {
      cmd.stdout(std::process::Stdio::null());
      cmd.stderr(std::process::Stdio::null());
    }

    debug!("Running ffmpeg for: {}", file.display());
    let status = cmd.status().await.context("Failed to run ffmpeg")?;

    if !status.success() {
      anyhow::bail!("ffmpeg failed with status: {status}");
    }

    Ok(())
  }

  fn add_export_params(cmd: &mut Command, output_path: &Path) {
    cmd.args([
      "-map",
      "0:v:0",
      "-map",
      "0:a?",
      "-map_metadata",
      "0",
      "-c:v",
      "libx264",
      "-preset",
      "slow",
      "-crf",
      "18",
      "-pix_fmt",
      "yuv420p",
      "-c:a",
      "aac",
      "-b:a",
      "320k",
      "-movflags",
      "+faststart",
      &output_path.to_string_lossy(),
    ]);
  }

  fn add_editing_params(cmd: &mut Command, output_path: &Path, config: &Config) {
    cmd.args(["-map", "0:v:0", "-map", "0:a?", "-map_metadata", "0"]);

    match config.codec {
      Codec::DnxHR => {
        let profile = match config.quality.as_str() {
          "lb" => "dnxhr_lb",
          "sq" => "dnxhr_sq",
          "hqx" => "dnxhr_hqx",
          "444" => "dnxhr_444",
          _ => "dnxhr_hq",
        };

        cmd.args([
          "-c:v",
          "dnxhd",
          "-profile:v",
          profile,
          "-pix_fmt",
          "yuv422p",
          "-c:a",
          "pcm_s16le",
        ]);
      }
      Codec::ProRes => {
        let profile = match config.quality.as_str() {
          "proxy" => "0",
          "lt" => "1",
          "standard" => "2",
          _ => "3",
        };

        cmd.args([
          "-c:v",
          "prores_ks",
          "-profile:v",
          profile,
          "-vendor",
          "ap10",
          "-pix_fmt",
          "yuv422p10le",
          "-c:a",
          "pcm_s16le",
        ]);
      }
    }

    cmd.args(["-movflags", "+faststart", &output_path.to_string_lossy()]);
  }

  const fn get_target_codec_name(codec: Codec) -> &'static str {
    match codec {
      Codec::DnxHR => "dnxhd",
      Codec::ProRes => "prores",
    }
  }
}
