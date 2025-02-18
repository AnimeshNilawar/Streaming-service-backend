package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"packetized-media-streaming/handlers"
	"packetized-media-streaming/handlers/streaming"
	"packetized-media-streaming/handlers/upload"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Database
	handlers.InitDB()

	// Ensure proper cleanup on shutdown
	defer func() {
		if handlers.CloudSQLDB != nil {
			handlers.CloudSQLDB.Close()
			fmt.Println("Database connection closed")
		}
	}()

	// Ensure GOOGLE_APPLICATION_CREDENTIALS is set (Optional)
	credentials := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentials == "" {
		fmt.Println("Warning: GOOGLE_APPLICATION_CREDENTIALS is not set")
	}

	// Setup Gin router
	r := gin.Default()

	r.MaxMultipartMemory = 2 << 30 // 500 MB limit  

	// Routes
	r.POST("/upload", upload.UploadVideo)
	r.GET("/stream/:videoID", streaming.GetVideoURL)

	// Get PORT from environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	// Start Server with graceful shutdown handling
	go func() {
		fmt.Printf("Server running on port %s\n", port)
		if err := r.Run(":" + port); err != nil {
			fmt.Printf("Failed to start server: %v\n", err)
			os.Exit(1)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down server...")
	handlers.CloudSQLDB.Close()
	fmt.Println("Server shut down gracefully")
}
