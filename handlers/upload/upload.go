package upload

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// import (
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"os"
// 	// "packetized-media-streaming/handlers"
// 	"path/filepath"

// 	"github.com/gin-gonic/gin"
// 	"github.com/google/uuid"
// )

const (
	bucketName      = "packetized-media-bucket"
	credentialsFile = "service-account.json"
	localStorage    = "./videos"
	maxFileSize     = 2 * 1024 * 1024 * 1024 // 2GB in bytes
)

// // Upload video endpoint
// func UploadVideo(c *gin.Context) {
// 	fileHeader, err := c.FormFile("file")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
// 		return
// 	}

// 	// Check file size (avoid exceeding 2GB)
// 	if fileHeader.Size > maxFileSize {
// 		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "File size exceeds 2GB limit"})
// 		return
// 	}

// 	// Generate a unique filename
// 	fileExt := filepath.Ext(fileHeader.Filename)
// 	videoID := uuid.New().String()
// 	newFileName := videoID + fileExt
// 	localFilePath := filepath.Join(localStorage, newFileName)

// 	// Ensure storage directory exists
// 	if _, err := os.Stat(localStorage); os.IsNotExist(err) {
// 		if err := os.MkdirAll(localStorage, os.ModePerm); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create storage directory"})
// 			return
// 		}
// 	}

// 	// Open file stream
// 	src, err := fileHeader.Open()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
// 		return
// 	}
// 	defer src.Close()

// 	// Create destination file
// 	dst, err := os.Create(localFilePath)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file on disk"})
// 		return
// 	}
// 	defer dst.Close()

// 	// Stream copy file (efficient for large files)
// 	if _, err := io.Copy(dst, src); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
// 		return
// 	}

// 	// Get Duration of video
// 	// duration, err := GetVideoDuration(localFilePath)
// 	// if err != nil {
// 	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video duration"})
// 	// 	return
// 	// }

// 	// // Insert video details into database
// 	// _, err = handlers.CloudSQLDB.Exec(`INSERT INTO videos (filename, path, duration) VALUES (?, ?, ?)`, newFileName, localFilePath, duration)
// 	// if err != nil {
// 	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save metadata"})
// 	// 	return
// 	// }

// 	// Start encoding in the background
// 	go EncodeVideo(localFilePath, videoID)

// 	videoURL := fmt.Sprintf("https://storage.googleapis.com/packetized-media-bucket/videos/%s/DASH/manifest.mpd", videoID)

// 	c.JSON(http.StatusOK, gin.H{
// 		"message":   "File uploaded successfully",
// 		"video_id":  videoID,
// 		"video_url": videoURL,
// 	})
// }

func UploadVideo(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Check file size (avoid exceeding 2GB)
	if fileHeader.Size > maxFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "File size exceeds 2GB limit"})
		return
	}

	// Generate a unique filename
	videoID := uuid.New().String()
	fileExt := filepath.Ext(fileHeader.Filename)
	newFileName := videoID + fileExt

	// Initalize GCS Client
	ctx := c.Request.Context()
	client, err := storage.NewClient(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create GCS client"})
		return
	}
	defer client.Close()

	// Upload the Video to GCS Bucket
	objectPath := fmt.Sprintf("videos/%s/%s", videoID, newFileName)
	object := client.Bucket(bucketName).Object(objectPath)

	// Open the uploaded file
	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer src.Close()

	// Create a Writer to GCS
	wc := object.NewWriter(ctx)
	defer wc.Close()

	wc.ContentType = getContentType(newFileName)

	if _, err := io.Copy(wc, src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file to GCS"})
		return
	}

	// // Wait for the object to be fully uploaded before processing the video
	// if !waitForObject(client, bucketName, objectPath, 10, 10*time.Second) {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "File upload not completed"})
	// 	return
	// }

	// Start encoding in the background
	go processVideoFromGCS(videoID, bucketName, newFileName)

	// Return the video URL
	videoURL := fmt.Sprintf("https://storage.googleapis.com/packetized-media-bucket/videos/%s/DASH/manifest.mpd", videoID)
	c.JSON(http.StatusOK, gin.H{
		"message":   "File uploaded successfully",
		"video_id":  videoID,
		"video_url": videoURL,
	})
}

// waitForObject checks if the object exists in GCS and retries a few times before giving up
func waitForObject(client *storage.Client, bucketName, objectPath string, maxRetries int, delay time.Duration) bool {
	ctx := context.Background()
	obj := client.Bucket(bucketName).Object(objectPath)

	// Retry logic
	for i := 0; i < maxRetries; i++ {
		_, err := obj.Attrs(ctx)
		if err == nil {
			// Object exists, proceed with the process
			return true
		}
		if err == storage.ErrObjectNotExist {
			// Object does not exist, wait and try again
			fmt.Printf("Object %s not found, retrying...\n", objectPath)
			time.Sleep(delay)
		} else {
			// Some other error occurred
			fmt.Printf("Error retrieving object attributes: %v\n", err)
			break
		}
	}
	return false
}
