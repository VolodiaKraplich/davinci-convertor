package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type ConversionMode string

const (
	ModeEditing ConversionMode = "editing"
	ModeExport  ConversionMode = "export"
)

type Action int

const (
	ActionSkip Action = iota
	ActionRewrap
	ActionConvert
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
	StartTime time.Time
}

type Config struct {
	OutputDir string
	Mode      ConversionMode
	Codec     string
	Quality   string
	Verbose   bool
	Force     bool
	Workers   int
}

var config Config

var rootCmd = &cobra.Command{
	Use:   "davinci-convert <file_or_directory>",
	Short: "Universal converter for DaVinci Resolve and video export",
	Long: `Converts video for editing in DaVinci Resolve or exports
finished videos to optimal universal format.

Modes:
  editing - Converts to DNxHR/ProRes for editing
  export  - Exports to H.264 for universal compatibility`,
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
	rootCmd.Flags().StringVarP(&config.OutputDir, "output-dir", "o", "", "Output directory for converted files")
	rootCmd.Flags().StringVar((*string)(&config.Mode), "mode", "editing", "Mode: 'editing' or 'export'")
	rootCmd.Flags().StringVar(&config.Codec, "codec", "dnxhr", "Codec for editing: 'dnxhr' or 'prores'")
	rootCmd.Flags().StringVar(&config.Quality, "quality", "hq", "Quality for editing: hq, hqx")
	rootCmd.Flags().BoolVarP(&config.Verbose, "verbose", "v", false, "Verbose output")
	rootCmd.Flags().BoolVarP(&config.Force, "force", "f", false, "Force overwrite existing files")
	rootCmd.Flags().IntVarP(&config.Workers, "workers", "w", runtime.NumCPU(), "Number of parallel jobs")
}

func checkDependencies() {
	for _, tool := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(tool); err != nil {
			printError(fmt.Sprintf("%s is not installed", tool))
			os.Exit(1)
		}
	}
}

func validateConfig() error {
	if config.Mode != ModeEditing && config.Mode != ModeExport {
		return fmt.Errorf("invalid mode '%s'", config.Mode)
	}

	if config.Mode == ModeEditing {
		if config.Codec != "dnxhr" && config.Codec != "prores" {
			return fmt.Errorf("invalid codec '%s'", config.Codec)
		}
	}

	return nil
}

func runConverter(rootPath string) {
	printHeader()
	if err := validateConfig(); err != nil {
		printError(err.Error())
		os.Exit(1)
	}

	files, err := findMediaFiles(rootPath)
	if err != nil {
		printError(fmt.Sprintf("Error accessing path: %v", err))
		os.Exit(1)
	}
	if len(files) == 0 {
		printWarning("No media files found")
		return
	}

	stats := Stats{StartTime: time.Now(), Total: len(files)}
	jobs := make(chan string, len(files))
	var wg sync.WaitGroup

	color.Magenta("🚀 Starting conversion with %d workers...\n", config.Workers)

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
	mediaExts := []string{".mov", ".mp4", ".mxf", ".avi", ".mkv", ".wmv", ".flv", ".m4v", ".webm"}
	return slices.Contains(mediaExts, ext)
}

func worker(wg *sync.WaitGroup, jobs <-chan string, stats *Stats) {
	defer wg.Done()
	for file := range jobs {
		if err := processFile(file, stats); err != nil {
			stats.Lock()
			stats.Failed++
			stats.Unlock()
			printError(fmt.Sprintf("%s: %v", filepath.Base(file), err))
		}
	}
}

func processFile(file string, stats *Stats) error {
	action, err := analyzeFile(file)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	outputPath := getOutputPath(file)

	if !config.Force {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("file exists (use --force to overwrite)")
		}
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if action == ActionSkip {
		stats.Lock()
		stats.Skipped++
		stats.Unlock()
		color.Green("✓ Skipped: %s (already compatible)", filepath.Base(file))
		return nil
	}

	if err := convert(file, outputPath, action); err != nil {
		return err
	}

	stats.Lock()
	stats.Success++
	stats.Unlock()

	color.Cyan("→ Processed: %s", filepath.Base(file))
	return nil
}

func analyzeFile(file string) (Action, error) {
	probe, err := runFFProbe(file)
	if err != nil {
		return ActionSkip, err
	}

	var videoStream *Stream
	for i := range probe.Streams {
		if probe.Streams[i].CodecType == "video" {
			videoStream = &probe.Streams[i]
			break
		}
	}

	if videoStream == nil {
		return ActionSkip, fmt.Errorf("no video stream found")
	}

	if config.Mode == ModeExport {
		if videoStream.CodecName == "h264" && strings.ToLower(filepath.Ext(file)) == ".mp4" {
			return ActionSkip, nil
		}
		return ActionConvert, nil
	}

	// Editing mode
	targetCodec := getTargetVideoCodec()
	if videoStream.CodecName == targetCodec {
		if strings.ToLower(filepath.Ext(file)) == ".mov" {
			return ActionSkip, nil
		}
		return ActionRewrap, nil
	}

	return ActionConvert, nil
}

func runFFProbe(file string) (*FFProbeOutput, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_streams", file)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	var probe FFProbeOutput
	if err := json.Unmarshal(out.Bytes(), &probe); err != nil {
		return nil, fmt.Errorf("invalid ffprobe output: %w", err)
	}

	return &probe, nil
}

func convert(file, outputPath string, action Action) error {
	args := buildFFmpegCommand(file, outputPath, action)
	cmd := exec.Command(args[0], args[1:]...)

	if config.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stderr = &bytes.Buffer{}
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	return nil
}

func buildFFmpegCommand(file, outputPath string, action Action) []string {
	args := []string{"ffmpeg", "-y", "-i", file}

	if action == ActionRewrap {
		return append(args, "-c", "copy", "-map_metadata", "0", outputPath)
	}

	if config.Mode == ModeExport {
		return append(args, buildExportParams(outputPath)...)
	}

	return append(args, buildEditingParams(outputPath)...)
}

func buildExportParams(outputPath string) []string {
	return []string{
		"-map", "0:v:0",
		"-map", "0:a?",
		"-map_metadata", "0",
		"-c:v", "libx264",
		"-preset", "slow",
		"-crf", "18",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", "320k",
		"-movflags", "+faststart",
		outputPath,
	}
}

func buildEditingParams(outputPath string) []string {
	params := []string{"-map", "0:v:0", "-map", "0:a?", "-map_metadata", "0"}

	if config.Codec == "dnxhr" {
		profile := map[string]string{
			"hq":  "dnxhr_hq",
			"hqx": "dnxhr_444",
		}[config.Quality]

		params = append(params,
			"-c:v", "dnxhd",
			"-profile:v", profile,
			"-pix_fmt", "yuv422p",
			"-c:a", "pcm_s16le",
		)
	} else {
		params = append(params,
			"-c:v", "prores_ks",
			"-profile:v", "3",
			"-vendor", "ap10",
			"-pix_fmt", "yuv422p10le",
			"-c:a", "pcm_s16le",
		)
	}

	return append(params, "-movflags", "+faststart", outputPath)
}

func getTargetVideoCodec() string {
	if config.Codec == "dnxhr" {
		return "dnxhd"
	}
	return "prores"
}

func getOutputPath(file string) string {
	dir := filepath.Dir(file)
	if config.OutputDir != "" {
		dir = config.OutputDir
	}

	base := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))

	var suffix, ext string
	if config.Mode == ModeExport {
		suffix = "_export"
		ext = ".mp4"
	} else {
		suffix = "_converted"
		ext = ".mov"
	}

	return filepath.Join(dir, base+suffix+ext)
}

func printHeader() {
	color.Cyan("========================================")
	color.Cyan("  DaVinci Converter")
	color.Cyan("========================================")
	if config.Mode == ModeExport {
		color.Yellow("Mode: Export | Workers: %d", config.Workers)
	} else {
		color.Yellow("Mode: Editing | Codec: %s | Quality: %s | Workers: %d",
			config.Codec, config.Quality, config.Workers)
	}
	if config.OutputDir != "" {
		color.Yellow("Output directory: %s", config.OutputDir)
	}
	color.Cyan("----------------------------------------")
}

func printError(msg string) {
	color.Red("✗ Error: %s", msg)
}

func printWarning(msg string) {
	color.Yellow("⚠️  Warning: %s", msg)
}

func printStats(stats *Stats) {
	elapsed := time.Since(stats.StartTime)
	color.Cyan("\n========================================")
	color.Green("✓ Completed in %v", elapsed.Round(time.Second))
	color.Cyan("Total: %d | Success: %d | Skipped: %d | Failed: %d",
		stats.Total, stats.Success, stats.Skipped, stats.Failed)
}
