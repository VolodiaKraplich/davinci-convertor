# 🧠 Smart DaVinci Resolve Converter

[![Rust Version](https://img.shields.io/badge/Rust-1.90%2B-orange.svg)](https://www.rust-lang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS-lightgrey.svg)](https://shields.io/)

A smart, high-performance CLI tool for batch-optimizing video files for DaVinci Resolve editing and exporting finished projects on Linux and macOS. Built with Rust for maximum performance and safety.

This utility automates the tedious process of preparing media for editing and exporting finished videos. Instead of blindly transcoding everything, it analyzes each file and applies only the necessary changes, ensuring a fast and quality-preserving workflow.

---

## The Problem

DaVinci Resolve, particularly on Linux, performs best with editing-friendly intermediate codecs. Natively working with highly compressed formats like H.264/H.265 often leads to poor timeline performance, stuttering playback, and potential instability. The standard solution is to transcode source footage to a mezzanine codec like DNxHR or ProRes.

Additionally, when exporting finished projects, you often need to convert to universally compatible formats like H.264/MP4 for distribution.

## Key Features

- **🧠 Intelligent Analysis:** Uses `ffprobe` to inspect each file's video/audio streams and container.
- **⚡️ Efficient Operations:** Skips compatible files or performs rewrapping when codecs are compatible but the container is not.
- **🚀 Concurrent Processing:** Leverages Rust's fearless concurrency (via `rayon`) to process multiple files in parallel, drastically reducing wait times.
- **🎬 Professional Codecs:** Transcodes to **DNxHR** or **Apple ProRes** for smooth, professional editing.
- **📤 Export Mode:** Converts finished edits to H.264/MP4 with optimal settings for universal compatibility.
- **📂 Flexible Workflow:** Supports custom output directories, codec profiles, quality settings, and parallel workers, powered by the `clap` argument parser.

---

## Installation

### Prerequisites

- **Rust** (version 1.90 or newer)
- **FFmpeg** (which includes `ffprobe`) must be installed and available in your system's PATH.

**Install Rust:**

```bash
# The recommended way to install Rust is via rustup
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

**Install FFmpeg:**

```bash
# On Debian/Ubuntu
sudo apt update && sudo apt install ffmpeg

# On Arch Linux
sudo pacman -Syu ffmpeg

# On macOS (using Homebrew)
brew install ffmpeg
```

### Building from Source

```bash
# Clone the repository
git clone https://github.com/VolodiaKraplich/davinci-converter.git
cd davinci-converter

# Build the optimized release binary
cargo build --release

# The executable will be available at ./target/release/davinci-converter
```

---

## Usage

### Basic Syntax

```bash
davinci-converter <FILE_OR_DIRECTORY> [FLAGS]
```

### Examples

#### Editing Mode (Default)

1.  **Convert a single file using default settings (DNxHR HQ):**

    ```bash
    davinci-converter my_video.mp4
    ```

2.  **Recursively process all videos in a directory:**

    ```bash
    davinci-converter ./path/to/my/footage/
    ```

3.  **Convert to ProRes and save to a dedicated proxy directory:**

    ```bash
    davinci-converter ./raw_footage/ -o ./proxies/ --codec prores
    ```

4.  **Convert to the highest quality DNxHR using 12 parallel workers:**

    ```bash
    davinci-converter ./source/ --codec dnxhr --quality hqx --workers 12
    ```

5.  **Force overwrite existing files and show detailed FFmpeg logs:**
    ```bash
    davinci-converter video.mkv -f -v
    ```

#### Export Mode

6.  **Export finished edits to H.264/MP4 for distribution:**

    ```bash
    davinci-converter ./finished_edits/ --mode export -o ./exports/
    ```

7.  **Export with verbose output and force overwrite:**
    ```bash
    davinci-converter final_project.mov --mode export -f -v
    ```

### All Flags

```
--mode <MODE>
Conversion mode: 'editing' or 'export' [default: editing]

--codec <CODEC>
Target video codec for editing mode: 'dnxhr' or 'prores' [default: dnxhr]

--quality <QUALITY>
Video quality profile for editing mode. - for dnxhr: hq, hqx [default: hq] - for prores: standard ProRes profile is used

-o, --output-dir <OUTPUT_DIR>
Output directory for converted files [default: same as source]

-f, --force
Force overwrite of existing files

-v, --verbose
Verbose output (show ffmpeg/ffprobe logs)

-w, --workers <WORKERS>
Number of parallel conversion jobs [default: number of CPU cores]

```

---

## How It Works

The tool follows this workflow for each file:

1.  **Analyze:** Runs `ffprobe` using Rust's `std::process::Command` to get codec and container information.
2.  **Decide:** Based on the analysis and mode, it chooses one of the following actions:
    - **SKIP:** The file is already fully compatible with the target format.
    - **REWRAP:** Codecs are compatible, but the container is not (e.g., `.mkv` → `.mov`). The streams are copied into the target container without re-encoding, which is nearly instantaneous.
    - **CONVERT:** The file needs transcoding to the target format.
3.  **Execute:** Spawns the appropriate `ffmpeg` child process to perform the chosen action.

### Editing Mode

Converts videos to editing-friendly formats:

- **DNxHR:** Professional intermediate codec with HQ (high quality) or HQX (highest quality) profiles.
- **ProRes:** Apple's professional codec for high-quality editing.
- **Output:** `.mov` container with PCM audio.

### Export Mode

Converts videos to a universally compatible distribution format:

- **Codec:** H.264 (libx264)
- **Quality:** CRF 18 with slow preset for excellent quality.
- **Audio:** AAC at 320kbps.
- **Container:** MP4 with fast-start for web streaming.
- **Output:** `_export.mp4` files.

---

## Performance

This tool is designed for speed. By default, it leverages all available CPU cores to process files in parallel, providing a significant speedup for batch operations. You can fine-tune the level of parallelism with the `--workers` flag.

- **Single file:** Processes immediately.
- **Multiple files:** Parallel processing based on available cores.
- **Custom worker count:** Use the `--workers` flag to control concurrency.

---

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.

---

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.
