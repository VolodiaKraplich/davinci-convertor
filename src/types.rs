use serde::Deserialize;
use std::fmt;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ConversionMode {
  Editing,
  Export,
}

impl std::str::FromStr for ConversionMode {
  type Err = anyhow::Error;

  fn from_str(s: &str) -> Result<Self, Self::Err> {
    match s.to_lowercase().as_str() {
      "editing" | "edit" => Ok(Self::Editing),
      "export" => Ok(Self::Export),
      _ => Err(anyhow::anyhow!(
        "Invalid mode '{s}'. Use 'editing' or 'export'"
      )),
    }
  }
}

impl fmt::Display for ConversionMode {
  fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
    match *self {
      Self::Editing => write!(f, "editing"),
      Self::Export => write!(f, "export"),
    }
  }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Action {
  Skip,
  Rewrap,
  Convert,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Codec {
  DnxHR,
  ProRes,
}

impl std::str::FromStr for Codec {
  type Err = anyhow::Error;

  fn from_str(s: &str) -> Result<Self, Self::Err> {
    match s.to_lowercase().as_str() {
      "dnxhr" | "dnxhd" => Ok(Self::DnxHR),
      "prores" => Ok(Self::ProRes),
      _ => Err(anyhow::anyhow!(
        "Invalid codec '{s}'. Use 'dnxhr' or 'prores'"
      )),
    }
  }
}

impl fmt::Display for Codec {
  fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
    match *self {
      Self::DnxHR => write!(f, "DNxHR"),
      Self::ProRes => write!(f, "ProRes"),
    }
  }
}

#[derive(Debug, Deserialize)]
pub struct Stream {
  pub codec_name: String,
  pub codec_type: String,
}

#[derive(Debug, Deserialize)]
pub struct FFProbeOutput {
  pub streams: Vec<Stream>,
}
