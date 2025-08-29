package onixctl

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
)

// Publisher is responsible for publishing artifacts.
type Publisher struct {
	config *Config
}

// NewPublisher creates a new Publisher.
func NewPublisher(config *Config) *Publisher {
	return &Publisher{config: config}
}

// Publish uploads artifacts to their specified destinations.
func (p *Publisher) Publish() error {
	if p.config.GSPath == "" {
		fmt.Println("No gsPath specified, skipping GCS upload.")
		return nil
	}

	zipFilePath := filepath.Join(p.config.Output, p.config.ZipFileName)
	if _, err := os.Stat(zipFilePath); os.IsNotExist(err) {
		fmt.Printf("Zip file not found at %s, skipping GCS upload.\n", zipFilePath)
		return nil
	}

	return p.uploadToGCS(zipFilePath, p.config.GSPath)
}

// uploadToGCS handles the file upload to Google Cloud Storage.
func (p *Publisher) uploadToGCS(filePath, gsPath string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create GCS client: %w", err)
	}
	defer client.Close()

	// gsPath is expected to be like gs://bucket-name/path/to/object
	if !strings.HasPrefix(gsPath, "gs://") {
		return fmt.Errorf("invalid GCS path: must start with gs://")
	}
	parts := strings.SplitN(strings.TrimPrefix(gsPath, "gs://"), "/", 2)
	if len(parts) < 2 {
		return fmt.Errorf("invalid GCS path: must include bucket and object path")
	}
	bucketName := parts[0]
	objectPath := parts[1]

	// If the object path ends with a '/', treat it as a directory and append the filename.
	if strings.HasSuffix(objectPath, "/") {
		objectPath = objectPath + filepath.Base(filePath)
	}

	fmt.Printf("Uploading %s to gs://%s/%s...\n", filePath, bucketName, objectPath)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for upload: %w", err)
	}
	defer file.Close()

	wc := client.Bucket(bucketName).Object(objectPath).NewWriter(ctx)
	if _, err = io.Copy(wc, file); err != nil {
		return fmt.Errorf("failed to copy file to GCS: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close GCS writer: %w", err)
	}

	fmt.Println("✅ Successfully uploaded to GCS.")
	return nil
}
