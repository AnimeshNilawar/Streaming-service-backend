package upload

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// Upload encoded video to Google Cloud Storage
func UploadToGCS(folderPath, videoID, format string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return err
	}
	defer client.Close()

	//upload each file in the folder
	err = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Destination in GCS
		objectPath := fmt.Sprintf("videos/%s/%s/%s", videoID, format, info.Name())

		// Open file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// upload to GCS
		wc := client.Bucket(bucketName).Object(objectPath).NewWriter(ctx)
		wc.ContentType = getContentType(objectPath)
		if _, err := io.Copy(wc, file); err != nil {
			return err
		}
		wc.Close()

		fmt.Printf("Uploaded %s to GCS\n", objectPath)
		return nil
	})

	os.RemoveAll(folderPath)

	return err
}

func getContentType(filename string) string {
	switch filepath.Ext(filename) {
	case ".mp4":
		return "video/mp4"
	case ".m3u8":
		return "application/x-mpegURL" // Correct MIME type for HLS
	case ".mpd":
		return "application/dash+xml" // Correct MIME type for DASH
	case ".ts":
		return "video/mp2t" // HLS Transport Stream segments
	case ".m4s":
		return "video/iso.segment" // DASH Media Segments
	default:
		return "application/octet-stream" // Default if unknown
	}
}
