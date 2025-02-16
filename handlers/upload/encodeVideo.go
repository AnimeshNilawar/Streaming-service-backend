package upload

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Encode video into different qualities using FFmpeg
func EncodeVideo(inputPath, videoID string) {
	// Convert to absolute path
	absInputPath, err := filepath.Abs(inputPath)
	if err != nil {
		fmt.Printf("Error getting absolute path: %v\n", err)
		return
	}
	inputPath = filepath.ToSlash(absInputPath)

	hlsOutput := filepath.ToSlash(filepath.Join(localStorage, videoID+"_hls"))
	dashOutput := filepath.ToSlash(filepath.Join(localStorage, videoID+"_dash"))

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		fmt.Printf("Error: Input file does not exist: %s\n", inputPath)
		return
	}

	fmt.Println("HLS Output Path:", hlsOutput)
	fmt.Println("DASH Output Path:", dashOutput)

	// Create Output Directories
	os.MkdirAll(hlsOutput, os.ModePerm)
	os.MkdirAll(dashOutput, os.ModePerm)

	// FFmpeg command for HLS
	hlsCmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-preset", "fast", "-g", "48", "-sc_threshold", "0",
		"-map", "0:v:0", "-map", "0:a:0",
		"-c:v", "libx264", "-crf", "23", "-profile:v", "main", "-c:a", "aac", "-ar", "48000", "-b:a", "128k",
		"-b:v:0", "800k", "-s:v:0", "640x360",
		"-hls_time", "10", "-hls_playlist_type", "vod",
		"-hls_flags", "independent_segments",
		"-hls_segment_filename", filepath.Join(hlsOutput, "segment_%03d.ts"),
		filepath.Join(hlsOutput, "playlist.m3u8"),
	)

	// FFmpeg command for DASH
	dashCmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-preset", "fast", "-g", "48", "-sc_threshold", "0",
		"-map", "0:v:0", "-map", "0:a:0",
		"-c:v", "libx264", "-crf", "23", "-profile:v", "main", "-c:a", "aac", "-ar", "48000", "-b:a", "128k",
		"-b:v:0", "800k", "-s:v:0", "640x360",
		"-b:v:1", "1400k", "-s:v:1", "1280x720",
		"-b:v:2", "2800k", "-s:v:2", "1920x1080",
		"-f", "dash",
		"-adaptation_sets", "id=0,streams=v id=1,streams=a",
		"-seg_duration", "10",
		"-use_timeline", "1",
		"-use_template", "1",
		"-init_seg_name", filepath.Join(dashOutput, "init-stream$RepresentationID$.m4s"), // Define init segment path
		"-media_seg_name", filepath.Join(dashOutput, "chunk-stream$RepresentationID$-$Number$.m4s"), // Define segment names
		filepath.Join(dashOutput, "manifest.mpd"), // DASH Manifest file
	)

	// Capture output for debugging	hlsCmd.Stderr = os.Stderr
	hlsCmd.Stdout = os.Stdout
	dashCmd.Stderr = os.Stderr
	dashCmd.Stdout = os.Stdout

	// Run FFmpeg process
	fmt.Println("Executing HLS Command:", hlsCmd.String())
	if err := hlsCmd.Run(); err != nil {
		fmt.Printf("HLS encoding failed: %v\n", err)
		return
	}

	fmt.Println("Executing DASH Command:", dashCmd.String())
	if err := dashCmd.Run(); err != nil {
		fmt.Printf("DASH encoding failed: %v\n", err)
		return
	}

	fmt.Println("Encoding completed for HLS & DASH")

	// Upload HLS & DASH segment to GCS
	UploadToGCS(hlsOutput, videoID, "HLS")
	UploadToGCS(dashOutput, videoID, "DASH")

	// Delete original file
	if err := os.Remove(inputPath); err != nil {
		fmt.Printf("Warning: failed to delete local file %s: %v\n", inputPath, err)
	} else {
		fmt.Printf("Deleted local file: %s\n", inputPath)
	}
}
