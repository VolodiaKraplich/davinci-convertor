# 🧠 Smart DaVinci Resolve Converter

[![Go Version](https://img.shields.io/badge/Go-1.25.1%2B-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS-lightgrey.svg)](https://shields.io/)

A smart, high-performance CLI tool for batch-optimizing video files for DaVinci Resolve editing and exporting finished projects on Linux and macOS.

This utility automates the tedious process of preparing media for editing and exporting finished videos. Instead of blindly transcoding everything, it analyzes each file and applies only the necessary changes, ensuring a fast and quality-preserving workflow.

---

## The Problem

DaVinci Resolve, particularly on Linux, performs best with editing-friendly intermediate codecs. Natively working with highly compressed formats like H.264/H.265 often leads to poor timeline performance, stuttering playback, and potential instability. The standard solution is to transcode source footage to a mezzanine codec like DNxHR or ProRes.

Additionally, when exporting finished projects, you often need to convert to universally compatible formats like H.264/MP4 for distribution.

## Key Features

- **🧠 Intelligent Analysis:** Uses `ffprobe` to inspect each file's video/audio streams and container
- **⚡️ Efficient Operations:** Skips compatible files or performs rewrapping when codecs are compatible but container is not
- **🚀 Concurrent Processing:** Utilizes all available CPU cores to process multiple files in parallel, drastically reducing wait times
- **🎬 Professional Codecs:** Transcodes to **DNxHR** or **Apple ProRes** for smooth, professional editing
- **📤 Export Mode:** Converts finished edits to H.264/MP4 with optimal settings for universal compatibility
- **📂 Flexible Workflow:** Supports custom output directories, codec profiles, quality settings, and parallel workers

---

## Installation

### Prerequisites

- **Go** (version 1.25.1 or newer)
- **FFmpeg** (which includes `ffprobe`) must be installed and available in your system's PATH

**Install FFmpeg:**

```bash
# On Debian/Ubuntu
sudo apt update && sudo apt install ffmpeg

# On Arch Linux Based
sudo pacman -Sy ffmpeg

# On macOS (using Homebrew)
brew install ffmpeg
```

### Building from Source

```bash
# Clone the repository
git clone https://github.com/VolodiaKraplich/davinci-convertor.git
cd davinci-convertor

# Build the binary
make build
```

---

## Usage

### Basic Syntax

```
davinci-convert <file_or_directory> [flags]
```

### Examples

#### Editing Mode (Default)

1.  **Convert a single file using default settings (DNxHR HQ):**

    ```bash
    davinci-convert my_video.mp4
    ```

2.  **Recursively process all videos in a directory:**

    ```bash
    davinci-convert ./path/to/my/footage/
    ```

3.  **Convert to ProRes and save to a dedicated proxy directory:**

    ```bash
    davinci-convert ./raw_footage/ -o ./proxies/ --codec prores
    ```

4.  **Convert to the highest quality DNxHR using 12 parallel workers:**

    ```bash
    davinci-convert ./source/ --codec dnxhr --quality hqx --workers 12
    ```

5.  **Force overwrite existing files and show detailed FFmpeg logs:**
    ```bash
    davinci-convert video.mkv -f -v
    ```

#### Export Mode

6.  **Export finished edits to H.264/MP4 for distribution:**

    ```bash
    davinci-convert ./finished_edits/ --mode export -o ./exports/
    ```

7.  **Export with verbose output and force overwrite:**
    ```bash
    davinci-convert final_project.mov --mode export -f -v
    ```

### All Flags

```
  --mode string
        Conversion mode: 'editing' or 'export' (default "editing")
  --codec string
        Target video codec for editing mode: 'dnxhr' or 'prores' (default "dnxhr")
  --quality string
        Video quality profile for editing mode:
        - for dnxhr: hq, hqx (default "hq")
        - for prores: standard ProRes profile is used
  -o, --output-dir string
        Output directory for converted files (default: same as source)
  -f, --force
        Force overwrite of existing files
  -v, --verbose
        Verbose output (show ffmpeg/ffprobe logs)
  -w, --workers int
        Number of parallel conversion jobs (default: number of CPU cores)
```

---

## How It Works

The tool follows this workflow for each file:

1.  **Analyze:** Runs `ffprobe` to get codec and container information
2.  **Decide:** Based on the analysis and mode, it chooses one of the following actions:
    - **SKIP:** The file is already fully compatible with the target format
    - **REWRAP:** Codecs are compatible, but the container is not (e.g., `.mkv` → `.mov`). The streams are copied into the target container without re-encoding, which is nearly instantaneous
    - **CONVERT:** The file needs transcoding to the target format
3.  **Execute:** Runs the appropriate `ffmpeg` command to perform the chosen action

### Editing Mode

Converts videos to editing-friendly formats:

- **DNxHR:** Professional intermediate codec with HQ (high quality) or HQX (highest quality) profiles
- **ProRes:** Apple's professional codec for high-quality editing
- Output: `.mov` container with PCM audio

### Export Mode

Converts videos to universally compatible distribution format:

- **Codec:** H.264 (libx264)
- **Quality:** CRF 18 with slow preset for excellent quality
- **Audio:** AAC at 320kbps
- **Container:** MP4 with fast-start for web streaming
- Output: `_export.mp4` files

---

## Performance

The tool automatically uses all available CPU cores by default, providing significant speedup for batch operations:

- Single file: Processes immediately
- Multiple files: Parallel processing based on available cores
- Custom worker count: Use `--workers` flag to control parallelism

---

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.

---

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.
