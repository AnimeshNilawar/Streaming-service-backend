package upload

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"packetized-media-streaming/handlers"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	bucketName      = "packetized-media-bucket"
	credentialsFile = "service-account.json"
	localStorage    = "./videos"
	maxFileSize     = 2 * 1024 * 1024 * 1024 // 2GB in bytes
)

// Upload video endpoint
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
	fileExt := filepath.Ext(fileHeader.Filename)
	videoID := uuid.New().String()
	newFileName := videoID + fileExt
	localFilePath := filepath.Join(localStorage, newFileName)

	// Ensure storage directory exists
	if _, err := os.Stat(localStorage); os.IsNotExist(err) {
		if err := os.MkdirAll(localStorage, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create storage directory"})
			return
		}
	}

	// Open file stream
	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(localFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file on disk"})
		return
	}
	defer dst.Close()

	// Stream copy file (efficient for large files)
	if _, err := io.Copy(dst, src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Get Duration of video
	duration, err := GetVideoDuration(localFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video duration"})
		return
	}

	// Insert video details into database
	_, err = handlers.CloudSQLDB.Exec(`INSERT INTO videos (filename, path, duration) VALUES (?, ?, ?)`, newFileName, localFilePath, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save metadata"})
		return
	}

	// Start encoding in the background
	go EncodeVideo(localFilePath, videoID)

	videoURL := fmt.Sprintf("https://storage.googleapis.com/packetized-media-bucket/videos/%s/DASH/manifest.mpd", videoID)

	c.JSON(http.StatusOK, gin.H{
		"message":   "File uploaded successfully",
		"video_id":  videoID,
		"video_url": videoURL,
	})
}
