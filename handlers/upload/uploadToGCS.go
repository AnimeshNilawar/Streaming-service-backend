package upload

import (
	"context"
	"fmt"
	"io"
	"os"
	"packetized-media-streaming/handlers"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// Upload encoded video to Google Cloud Storage
func UploadToGCS(filePath, videoID, resolution, format string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return err
	}
	defer client.Close()

	// GCS Path
	gcsPath := fmt.Sprintf("videos/%s/%s.mp4", videoID, resolution)

	// Open file for rendering
	src, err := os.Open(filePath)
	if err != nil {
		return err
	}
	
	defer func() {
		src.Close()
		time.Sleep(500 * time.Millisecond)
		err := os.Remove(filePath)
		if err != nil {
			fmt.Printf("Warning: failed to delete local file %s: %v\n", filePath, err)
		} else {
			fmt.Printf("Deleted local file: %s\n", filePath)
		}
	}()

	// Upload file
	bucket := client.Bucket(bucketName)
	object := bucket.Object(gcsPath)
	writer := object.NewWriter(ctx)
	writer.ContentType = "video/mp4"

	if _, err := io.Copy(writer, src); err != nil {
		return fmt.Errorf("failed to copy file to GCS: %v", err)
	}
	writer.Close()

	// Remove local files after upload
	err = os.Remove(filePath)
	if err != nil {
		fmt.Printf("Warning: failed to delete local file %s: %v\n", filePath, err)
	} else {
		fmt.Printf("Deleted local file: %s\n", filePath)
	}

	// Store video encoding details in database
	_, err = handlers.CloudSQLDB.Exec(`INSERT INTO video_encoding (video_id, resolution, format, path) VALUES (?, ?, ?, ?)`,
		videoID, resolution, format, gcsPath)
	if err != nil {
		return fmt.Errorf("failed to update database: %v", err)
	}

	fmt.Printf("Uploaded %s to GCS at %s \n", filePath, gcsPath)
	return nil
}
