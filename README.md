# 🧠 Smart DaVinci Resolve Converter

[![Go Version](https://img.shields.io/badge/Go-1.25.1%2B-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS-lightgrey.svg)](https://shields.io/)

A smart, high-performance CLI tool for batch-optimizing video files for DaVinci Resolve on Linux and macOS.

This utility automates the tedious process of preparing media for editing. Instead of blindly transcoding everything, it analyzes each file and applies only the necessary changes, ensuring a fast and quality-preserving workflow.

---

### The Problem

DaVinci Resolve, particularly on Linux, performs best with editing-friendly intermediate codecs. Natively working with highly compressed formats like H.264/H.265 often leads to poor timeline performance, stuttering playback, and potential instability. The standard solution is to transcode source footage to a mezzanine codec like DNxHR or ProRes.

### Key Features

- **🧠 Intelligent Analysis:** Uses `ffprobe` to inspect each file's video/audio streams and container.
- ⚡️ **Efficient Operations:** Skips compatible files, remuxes containers (`.mkv` -> `.mov`), or performs partial transcodes (e.g., audio only) when possible.
- 🚀 **Concurrent Processing:** Utilizes all available CPU cores to process multiple files in parallel, drastically reducing wait times.
- 🎬 **Professional Codecs:** Transcodes to **DNxHR** or **Apple ProRes** for a smooth, professional editing experience.
- 📂 **Flexible Workflow:** Supports custom output directories, codec profiles, and quality settings.

---

### Installation

#### Prerequisites

- **Go** (version 1.25.1 or newer)
- **FFmpeg** (which includes `ffprobe`) must be installed and available in your system's PATH.

**Install FFmpeg:**

```bash
# On Debian/Ubuntu
sudo apt update && sudo apt install ffmpeg

# On ArchLinux Based
sudo pacman -Sy ffmpeg

# On macOS (using Homebrew)
brew install ffmpeg
```

### Usage

#### Basic Syntax

```
davinci-convert <file_or_directory> [flags]
```

#### Examples

1.  **Convert a single file using default settings (DNxHR HQ):**

    ```bash
    davinci-convert my_video.mp4
    ```

2.  **Recursively process all videos in a directory:**

    ```bash
    davinci-convert ./path/to/my/footage/
    ```

3.  **Convert to ProRes LT and save to a dedicated proxy directory:**

    ```bash
    davinci-convert ./raw_footage/ -o ./proxies/ --codec prores --quality lt
    ```

4.  **Convert to the highest quality DNxHR using 12 parallel workers:**

    ```bash
    davinci-convert ./source/ --codec dnxhr --quality hqx --workers 12
    ```

5.  **Force overwrite existing files and show detailed FFmpeg logs:**
    ```bash
    davinci-convert video.mkv -f -v
    ```

#### All Flags

```
  --codec string
        Target video codec: 'dnxhr' or 'prores' (default "dnxhr")
  -f
        Force overwrite of existing files
  -o, --output-dir string
        Output directory for converted files (default: same as source)
  --quality string
        Video quality profile
        - for dnxhr: low, medium, hq, hqx
        - for prores: proxy, lt, sq, hq
        (default "hq")
  -v
        Verbose output (show ffmpeg/ffprobe logs)
  --workers int
        Number of parallel conversion jobs (default: number of CPU cores)
```

---

### How It Works

The tool follows this workflow for each file:

1.  **Analyze:** Runs `ffprobe` to get codec and container information.
2.  **Decide:** Based on the analysis, it chooses one of the following actions:
    - `SKIP`: The file is already fully compatible.
    - `REWRAP`: Codecs are compatible, but the container is not (e.g., `.mkv`). The streams are copied into a `.mov` container without re-encoding, which is nearly instantaneous.
    - `CONVERT AUDIO`: The video stream is compatible and is copied, but the audio is transcoded.
    - `CONVERT VIDEO`: The audio stream is compatible and is copied, but the video is transcoded.
    - `FULL CONVERT`: Both video and audio streams require transcoding.
3.  **Execute:** Runs the appropriate `ffmpeg` command to perform the chosen action.

---

### Building from Source

```bash
# Clone the repository
git clone https://github.com/VolodiaKraplich/davinci-convertor.git
cd davinci-convertor

# Build the binary
make build
```

### License

This project is licensed under the MIT License. See the `LICENSE` file for details.
