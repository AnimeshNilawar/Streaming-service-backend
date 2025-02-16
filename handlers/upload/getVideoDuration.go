package upload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

// Video metadata struct to parse JSON Output
type VideoMetadata struct {
	Format struct {
		Duration string `json:"duration"`
	}	`json:"format"`
}

func GetVideoDuration(filePath string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", filePath)	
	
	// Capture output
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return 0, fmt.Errorf("failed to execute ffprobe: %v", err)
	}

	// Parse JSON output
	var metadata VideoMetadata
	err = json.Unmarshal(out.Bytes(), &metadata)
	if err != nil {
		return 0, fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Convert duration to float
	duration, err := strconv.ParseFloat(metadata.Format.Duration, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert duration to float: %v", err)
	}

	return duration, nil
}