use anyhow::{Context, Result};
use std::path::{Path, PathBuf};
use tokio::fs;
use tracing::{debug, info};

const MEDIA_EXTENSIONS: &[&str] = &[
  ".mov", ".mp4", ".mxf", ".avi", ".mkv", ".wmv", ".flv", ".m4v", ".webm", ".mpg", ".mpeg",
  ".m2ts", ".mts",
];

pub struct FileScanner;

impl FileScanner {
  pub async fn scan(path: &Path) -> Result<Vec<PathBuf>> {
    info!("Scanning for media files in: {}", path.display());

    let metadata = fs::metadata(path)
      .await
      .with_context(|| format!("Failed to access path: {}", path.display()))?;

    let files = if metadata.is_file() {
      if Self::is_media_file(path) {
        vec![path.to_path_buf()]
      } else {
        Vec::new()
      }
    } else {
      Self::scan_directory(path).await?
    };

    info!("Found {} media file(s)", files.len());
    Ok(files)
  }

  async fn scan_directory(path: &Path) -> Result<Vec<PathBuf>> {
    let mut files = Vec::new();
    let mut stack = vec![path.to_path_buf()];

    while let Some(dir) = stack.pop() {
      let mut entries = fs::read_dir(&dir)
        .await
        .with_context(|| format!("Failed to read directory: {}", dir.display()))?;

      while let Some(entry) = entries.next_entry().await? {
        let path = entry.path();
        let metadata = entry.metadata().await?;

        if metadata.is_file() {
          if Self::is_media_file(&path) {
            debug!("Found media file: {}", path.display());
            files.push(path);
          }
        } else if metadata.is_dir() {
          stack.push(path);
        }
      }
    }

    Ok(files)
  }

  fn is_media_file(path: &Path) -> bool {
    path
      .extension()
      .and_then(|ext| ext.to_str())
      .is_some_and(|ext| {
        let ext_lower = format!(".{}", ext.to_lowercase());
        MEDIA_EXTENSIONS.contains(&ext_lower.as_str())
      })
  }
}
