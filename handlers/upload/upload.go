package upload

import (
	"fmt"
	"net/http"
	"packetized-media-streaming/handlers"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	bucketName      = "packetized-media-bucket"
	credentialsFile = "service-account.json"
	localStorage    = "./videos"
)

// Upload video endpoint
func UploadVideo(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Generate a unique filename
	fileExt := filepath.Ext(file.Filename)
	videoID := uuid.New().String()
	newFileName := videoID + fileExt
	localFilePath := filepath.Join(localStorage, newFileName)

	// Save file locally
	if err := c.SaveUploadedFile(file, localFilePath); err != nil {
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
	_, err = handlers.CloudSQLDB.Exec(`INSERT INTO videos (filename, path, duration) VALUES (?, ?, ?)`, newFileName,localFilePath, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save metadata"})
		return
	}

	// Start encoding in the background
	go EncodeVideo(localFilePath, videoID)

	// Respond with immediate playback URL (low quality)
	lowQualityURL := fmt.Sprintf("https://storage.googleapis.com/%s/videos/%s/360p.mp4", bucketName, videoID)

	c.JSON(http.StatusOK, gin.H{
		"message":          "File uploaded successfully",
		"video_id":         videoID,
		"instant_play_url": lowQualityURL,
	})
}
