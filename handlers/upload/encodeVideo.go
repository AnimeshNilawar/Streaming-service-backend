package upload

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Encode video into different qualities using FFmpeg
func EncodeVideo(inputPath, videoID string) {
	output360p := fmt.Sprintf("%s/%s_360p.mp4", localStorage, videoID)
	output720p := fmt.Sprintf("%s/%s_720p.mp4", localStorage, videoID)
	output1080p := fmt.Sprintf("%s/%s_1080p.mp4", localStorage, videoID)

	// FFmpeg commands
	commands := [][]string{
		{"-i", inputPath, "-vf", "scale=640:360", "-c:v", "libx264", "-preset", "slow", "-crf", "23", "-c:a", "aac", "-b:a", "128k", output360p},
		{"-i", inputPath, "-vf", "scale=1280:720", "-c:v", "libx264", "-preset", "slow", "-crf", "20", "-c:a", "aac", "-b:a", "192k", output720p},
		{"-i", inputPath, "-vf", "scale=1920:1080", "-c:v", "libx264", "-preset", "slow", "-crf", "18", "-c:a", "aac", "-b:a", "256k", output1080p},
	}	

	// Run encoding commands
	for _, cmdArgs := range commands {
		cmd := exec.Command("ffmpeg", cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Printf("FFmpeg failed: %v\n", err)
		}
	}

	time.Sleep(1 * time.Second)

	// Delete the original file after encoding and uploading
	os.Remove(inputPath)
	fmt.Println("Original file deleted: ", inputPath)

	// Upload encoded videos to GCS
	UploadToGCS(output360p, videoID, "360p", "HLS")
	UploadToGCS(output720p, videoID, "720p", "HLS")
	UploadToGCS(output1080p, videoID, "1080p", "HLS")
}
