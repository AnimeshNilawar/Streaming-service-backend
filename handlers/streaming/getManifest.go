package streaming

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetVideoURL(c *gin.Context) {
	videoID := c.Param("videoID")
	format := c.Query("format") // Either "DASH" or "HLS"

	if format != "DASH" && format != "HLS" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format. Use 'DASH' or 'HLS'."})
		return
	}

	// Determine manifest file path
	objectPath := fmt.Sprintf("videos/%s/%s/manifest.mpd", videoID, format)
	if format == "HLS" {
		objectPath = fmt.Sprintf("videos/%s/%s/playlist.m3u8", videoID, format)
	}

	// Generate signed URL
	url, err := GenerateSignedURL(objectPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate signed URL"})
		return
	}

	// Return the signed URL
	c.JSON(http.StatusOK, gin.H{"signed_url": url})
}