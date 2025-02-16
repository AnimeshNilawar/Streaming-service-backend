package main

import (
	"fmt"
	"net/http"
	"packetized-media-streaming/handlers"
	"packetized-media-streaming/handlers/upload"

	"github.com/gin-gonic/gin"
)

func main() {
	handlers.InitDB()
	defer handlers.CloudSQLDB.Close()

	fmt.Println("Database connection established")

	r := gin.Default()

	// Routes
	r.POST("/upload", upload.UploadVideo)
	r.GET("/stream/:videoID", getVideoURL)

	// Start Server
	port := "8080"
	fmt.Printf("Server running on port %s\n", port)
	err := r.Run(":" + port)
	if err != nil {
		fmt.Println("Failed to start server")
	}
}

func getVideoURL(c *gin.Context) {
	videoID := c.Param("id")

	var videoURL string
	err := handlers.CloudSQLDB.QueryRow("SELECT url FROM videos WHERE id = ?", videoID).Scan(&videoURL)
	if err != nil {
		c.JSON(404, gin.H{"error": "Video not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"video_url": videoURL})
}