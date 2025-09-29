package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type Action int

const (
	ActionSkip Action = iota
	ActionRewrap
	ActionConvertAudio
	ActionConvertVideo
	ActionFullConvert
	ActionUnsupported
)

type Stream struct {
	CodecName string `json:"codec_name"`
	CodecType string `json:"codec_type"`
}

type FFProbeOutput struct {
	Streams []Stream `json:"streams"`
}

type Stats struct {
	sync.Mutex
	Total     int
	Success   int
	Failed    int
	Skipped   int
	Rewrapped int
	StartTime time.Time
}

var config struct {
	OutputDir string
	Codec     string
	Quality   string
	Verbose   bool
	Force     bool
	Workers   int
}

var rootCmd = &cobra.Command{
	Use:   "davinci-convert <file_or_directory>",
	Short: "A smart, high-performance tool to prepare media for DaVinci Resolve.",
	Long: `Smart DaVinci Resolve Converter analyzes media files and intelligently
converts them to editing-friendly formats like DNxHR or ProRes.

It avoids unnecessary transcoding by skipping compatible files or simply
rewrapping containers, saving time and preserving quality.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runConverter(args[0])
	},
}

func Execute() {
	cobra.OnInitialize(checkDependencies)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&config.OutputDir, "output-dir", "o", "", "Output directory for converted files (default: same as source)")
	rootCmd.Flags().StringVar(&config.Codec, "codec", "dnxhr", "Target video codec: 'dnxhr' or 'prores'")
	rootCmd.Flags().StringVar(&config.Quality, "quality", "hq", "Video quality profile (for dnxhr: low, medium, hq, hqx; for prores: proxy, lt, sq, hq)")
	rootCmd.Flags().BoolVarP(&config.Verbose, "verbose", "v", false, "Verbose output (show ffmpeg/ffprobe logs)")
	rootCmd.Flags().BoolVarP(&config.Force, "force", "f", false, "Force overwrite of existing files")
	rootCmd.Flags().IntVarP(&config.Workers, "workers", "w", runtime.NumCPU(), "Number of parallel conversion jobs")
}

func checkDependencies() {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		printError("ffmpeg is not installed or not in PATH")
		os.Exit(1)
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		printError("ffprobe is not installed or not in PATH")
		os.Exit(1)
	}
}

func validateConfig() error {
	if config.Codec != "dnxhr" && config.Codec != "prores" {
		return fmt.Errorf("invalid codec '%s'. Must be 'dnxhr' or 'prores'", config.Codec)
	}

	validQualities := map[string][]string{
		"dnxhr":  {"low", "medium", "hq", "hqx"},
		"prores": {"proxy", "lt", "sq", "hq"},
	}

	qualities := validQualities[config.Codec]
	for _, q := range qualities {
		if config.Quality == q {
			return nil
		}
	}
	return fmt.Errorf("invalid quality '%s' for codec '%s'. Valid options are: %v", config.Quality, config.Codec, qualities)
}

func runConverter(rootPath string) {
	printHeader()
	if err := validateConfig(); err != nil {
		printError(err.Error())
		os.Exit(1)
	}

	files, err := findMediaFiles(rootPath)
	if err != nil {
		printError(fmt.Sprintf("Error accessing path %s: %v", rootPath, err))
		os.Exit(1)
	}
	if len(files) == 0 {
		printWarning("No media files found to process.")
		return
	}

	stats := Stats{StartTime: time.Now(), Total: len(files)}
	jobs := make(chan string, len(files))
	var wg sync.WaitGroup

	color.Magenta("🚀 Starting conversion with %d parallel workers...\n", config.Workers)

	for i := 0; i < config.Workers; i++ {
		wg.Add(1)
		go worker(&wg, jobs, &stats)
	}

	for _, file := range files {
		jobs <- file
	}
	close(jobs)
	wg.Wait()

	printStats(&stats)
}

func findMediaFiles(path string) ([]string, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !fileInfo.IsDir() {
		if isMediaFile(path) {
			return []string{path}, nil
		}
		return nil, nil
	}

	var files []string
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isMediaFile(path) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func isMediaFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	mediaExts := map[string]bool{
		".mov": true, ".mp4": true, ".mxf": true, ".avi": true,
		".mkv": true, ".wmv": true, ".flv": true,
	}
	return mediaExts[ext]
}

func worker(wg *sync.WaitGroup, jobs <-chan string, stats *Stats) {
	defer wg.Done()
	for file := range jobs {
		action, err := analyzeFile(file)
		if err != nil {
			stats.Lock()
			stats.Failed++
			stats.Unlock()
			printError(fmt.Sprintf("Failed to analyze %s: %v", file, err))
			continue
		}

		outputPath := getOutputPath(file)
		if !config.Force {
			if _, err := os.Stat(outputPath); err == nil {
				stats.Lock()
				stats.Failed++
				stats.Unlock()
				printError(fmt.Sprintf("Output file exists: %s (use --force to overwrite)", outputPath))
				continue
			}
		}

		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			stats.Lock()
			stats.Failed++
			stats.Unlock()
			printError(fmt.Sprintf("Failed to create output directory %s: %v", outputDir, err))
			continue
		}

		if action == ActionSkip {
			stats.Lock()
			stats.Skipped++
			stats.Unlock()
			color.Green("✓ Skipping %s (already compatible)", file)
			continue
		}

		cmdArgs := buildFFmpegCommand(file, outputPath, action)
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)

		var stderr, stdout bytes.Buffer
		if config.Verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		} else {
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
		}

		if err := cmd.Run(); err != nil {
			stats.Lock()
			stats.Failed++
			stats.Unlock()
			errMsg := stderr.String()
			if errMsg == "" {
				errMsg = stdout.String()
			}
			printError(fmt.Sprintf("Conversion failed for %s: %v\n%s", file, err, errMsg))
			continue
		}

		stats.Lock()
		if action == ActionRewrap {
			stats.Rewrapped++
		} else {
			stats.Success++
		}
		stats.Unlock()

		color.Cyan("→ Processed %s [%s]", file, actionToString(action))
	}
}

func analyzeFile(file string) (Action, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_streams", file)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ActionUnsupported, fmt.Errorf("ffprobe failed: %w", err)
	}

	var probe FFProbeOutput
	if err := json.Unmarshal(out.Bytes(), &probe); err != nil {
		return ActionUnsupported, fmt.Errorf("invalid ffprobe output: %w", err)
	}

	var videoCodec, audioCodec string
	hasVideo, hasAudio := false, false
	for _, stream := range probe.Streams {
		if stream.CodecType == "video" {
			videoCodec = stream.CodecName
			hasVideo = true
		} else if stream.CodecType == "audio" {
			audioCodec = stream.CodecName
			hasAudio = true
		}
	}

	if !hasVideo {
		return ActionUnsupported, fmt.Errorf("no video stream found")
	}

	targetVideoCodec, acceptableAudioCodecs := getTargetCodecs()
	videoMatch := videoCodec == targetVideoCodec
	audioMatch := !hasAudio || isAcceptableAudio(audioCodec, acceptableAudioCodecs)
	containerMatch := strings.ToLower(filepath.Ext(file)) == ".mov"

	if videoMatch && audioMatch && containerMatch {
		return ActionSkip, nil
	}
	if videoMatch && audioMatch {
		return ActionRewrap, nil
	}
	if videoMatch {
		return ActionConvertAudio, nil
	}
	if audioMatch {
		return ActionConvertVideo, nil
	}
	return ActionFullConvert, nil
}

func getTargetCodecs() (string, []string) {
	if config.Codec == "dnxhr" {
		return "dnxhd", []string{"pcm_s16le", "pcm_s24le", "pcm_s32le"}
	}
	return "prores", []string{"aac"}
}

func isAcceptableAudio(codec string, acceptable []string) bool {
	for _, c := range acceptable {
		if codec == c {
			return true
		}
	}
	return false
}

func buildFFmpegCommand(file, outputPath string, action Action) []string {
	baseArgs := []string{"ffmpeg", "-y", "-i", file, "-map_metadata", "0"}

	switch action {
	case ActionRewrap:
		return append(baseArgs, "-c", "copy", outputPath)

	case ActionConvertAudio:
		audioCodec := getAudioCodec()
		return append(baseArgs, "-c:v", "copy", "-c:a", audioCodec, outputPath)

	case ActionConvertVideo, ActionFullConvert:
		args := append(baseArgs, getVideoCodecParams()...)
		args = append(args,
			"-map", "0:v:0",
			"-map", "0:a?",
			"-sws_flags", "lanczos",
			"-movflags", "+faststart",
		)

		if action == ActionFullConvert {
			audioCodec := getAudioCodec()
			args = append(args, "-c:a", audioCodec)
			if audioCodec == "aac" {
				args = append(args, "-b:a", "320k")
			}
		} else {
			args = append(args, "-c:a", "copy")
		}

		return append(args, outputPath)
	}

	return baseArgs
}

func getAudioCodec() string {
	if config.Codec == "prores" {
		return "aac"
	}
	return "pcm_s16le"
}

func getVideoCodecParams() []string {
	switch config.Codec {
	case "dnxhr":
		return []string{
			"-c:v", "dnxhd",
			"-profile:v", getDNxHRProfile(),
			"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2",
			"-pix_fmt", "yuv422p",
		}
	case "prores":
		return []string{
			"-c:v", "prores_ks",
			"-profile:v", getProResProfile(),
			"-vendor", "ap10",
			"-pix_fmt", "yuv422p10le",
			"-qscale:v", "11",
		}
	}
	return nil
}

func getDNxHRProfile() string {
	profiles := map[string]string{
		"low":    "dnxhr_lb",
		"medium": "dnxhr_sq",
		"hq":     "dnxhr_hq",
		"hqx":    "dnxhr_444",
	}
	return profiles[config.Quality]
}

func getProResProfile() string {
	profiles := map[string]string{
		"proxy": "0",
		"lt":    "1",
		"sq":    "2",
		"hq":    "3",
	}
	return profiles[config.Quality]
}

func getOutputPath(file string) string {
	dir := filepath.Dir(file)
	if config.OutputDir != "" {
		dir = config.OutputDir
	}
	base := filepath.Base(file)
	ext := filepath.Ext(base)
	base = base[:len(base)-len(ext)] + "_converted.mov"
	return filepath.Join(dir, base)
}

func actionToString(action Action) string {
	actions := map[Action]string{
		ActionSkip:         "SKIP",
		ActionRewrap:       "REWRAP",
		ActionConvertAudio: "CONVERT_AUDIO",
		ActionConvertVideo: "CONVERT_VIDEO",
		ActionFullConvert:  "FULL_CONVERT",
	}
	if str, ok := actions[action]; ok {
		return str
	}
	return "UNSUPPORTED"
}

func printHeader() {
	color.Cyan("========================================")
	color.Cyan("  Smart DaVinci Resolve Converter")
	color.Cyan("========================================")
	color.Yellow("Codec: %s | Quality: %s | Workers: %d", config.Codec, config.Quality, config.Workers)
	if config.OutputDir != "" {
		color.Yellow("Output directory: %s", config.OutputDir)
	}
	color.Cyan("----------------------------------------")
}

func printError(msg string) {
	color.Red("❌ Error: %s", msg)
}

func printWarning(msg string) {
	color.Yellow("⚠️ Warning: %s", msg)
}

func printStats(stats *Stats) {
	elapsed := time.Since(stats.StartTime)
	color.Cyan("\n========================================")
	color.Green("✓ Completed in %v", elapsed.Round(time.Second))
	color.Cyan("Total: %d | Success: %d | Skipped: %d | Failed: %d",
		stats.Total, stats.Success+stats.Rewrapped, stats.Skipped, stats.Failed)
	if stats.Rewrapped > 0 {
		color.Cyan("Rewrapped: %d", stats.Rewrapped)
	}
}
