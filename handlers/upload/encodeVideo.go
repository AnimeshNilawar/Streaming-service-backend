package upload

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

func processVideoFromGCS(videoId, BucketName, fileName string) {
	// Construct the GCS object path
	objectPath := fmt.Sprintf("videos/%s/%s", videoId, fileName)

	// Initialize GCS Client
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		fmt.Printf("Failed to create GCS client: %v\n", err)
		return
	}
	defer client.Close()

	// Check if the object exists
	obj := client.Bucket(bucketName).Object(objectPath)
	// Retry loop to check if the file exists every 10 seconds
	maxRetries := 5
	retries := 0
	var objAttrs *storage.ObjectAttrs
	for retries < maxRetries {
		// Check if the object exists
		obj := client.Bucket(bucketName).Object(objectPath)
		objAttrs, err = obj.Attrs(ctx)
		if err != nil {
			if err == storage.ErrObjectNotExist {
				// If the object does not exist, retry after 10 seconds
				fmt.Printf("Object %s does not exist in bucket %s. Retrying in 10 seconds...\n", objectPath, bucketName)
			} else {
				// Handle other errors, such as network issues
				fmt.Printf("Error retrieving object attributes: %v\n", err)
				return
			}
		} else {
			// Object exists, log its attributes and break the retry loop
			fmt.Printf("Object %s found in bucket %s\n", objectPath, bucketName)
			fmt.Printf("Object attributes: Name=%s, Size=%d, ContentType=%s\n", objAttrs.Name, objAttrs.Size, objAttrs.ContentType)
			break
		}

		// Increment retry counter and wait for 10 seconds before retrying
		retries++
		time.Sleep(10 * time.Second)
	}

	if retries == maxRetries {
		// If we reached the max retries, exit the function
		fmt.Printf("Exceeded maximum retries (%d). Aborting...\n", maxRetries)
		return
	}

	// Download the file from GCS
	rc, err := obj.NewReader(ctx)
	if err != nil {
		fmt.Printf("Failed to read file from GCS: %v\n", err)
		return
	}
	defer rc.Close()

	// Create a "videos" directory if it doesn't exist
	videoDir := filepath.Join("videos", videoId)
	if err := os.MkdirAll(videoDir, os.ModePerm); err != nil {
		fmt.Printf("Failed to create video directory: %v\n", err)
		return
	}

	// Save the video to the newly created folder
	tempFilePath := filepath.Join(videoDir, fileName)
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		fmt.Printf("Failed to create temp file: %v\n", err)
		return
	}
	defer tempFile.Close()

	// Copy the video content from GCS to the temp file
	if _, err := io.Copy(tempFile, rc); err != nil {
		fmt.Printf("Failed to copy video from GCS to temp file: %v\n", err)
		return
	}

	// Process the video (encoding, etc.) using FFmpeg
	EncodeVideo(tempFilePath, videoId)

	// Delete the temporary file
	if err := os.Remove(tempFilePath); err != nil {
		fmt.Printf("Failed to delete temp file: %v\n", err)
	}
}

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
		"-r", "30", "-vsync", "cfr",
		"-map", "0:v:0", "-map", "0:v:0", "-map", "0:v:0", "-map", "0:a:0",
		"-c:v", "libx264", "-crf", "23", "-profile:v", "main", "-c:a", "aac", "-ar", "48000", "-b:a", "128k",
		"-b:v:0", "800k", "-s:v:0", "640x360",
		"-b:v:1", "1400k", "-s:v:1", "1280x720",
		"-b:v:2", "2800k", "-s:v:2", "1920x1080",
		"-f", "dash",
		"-adaptation_sets", "id=0,streams=v id=1,streams=a",
		"-seg_duration", "10", // 10 second segment duration
		"-use_timeline", "1",
		"-use_template", "1",
		"-init_seg_name", "init-stream$RepresentationID$.m4s",
		"-media_seg_name", "chunk-stream$RepresentationID$-$Number$.m4s",
		filepath.ToSlash(filepath.Join(dashOutput, "manifest.mpd")),
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
