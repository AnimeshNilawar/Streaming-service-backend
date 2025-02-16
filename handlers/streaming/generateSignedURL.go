package streaming

import (
	"context"
	"time"

	"cloud.google.com/go/storage"
)

const (
	bucketName = "packetized-media-bucket"
)

func GenerateSignedURL(objectPath string) (string, error) {
	ctx := context.Background()
	
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	// Expiry time for signed URL
	expiration := time.Now().Add(1 * time.Hour)

	// Generate signed URL
	url, err := client.Bucket(bucketName).SignedURL(objectPath, &storage.SignedURLOptions{
		Method: "GET",
		Expires: expiration,
	})
	if err != nil {
		return "", err
	}

	return url, nil
}