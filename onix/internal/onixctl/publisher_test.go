package onixctl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublisher_Publish_NoGSPath(t *testing.T) {
	config := &Config{}
	publisher := NewPublisher(config)
	err := publisher.Publish()
	assert.NoError(t, err, "should not return error if gsPath is not set")
}

func TestPublisher_Publish_NoZipFile(t *testing.T) {
	tmpDir := t.TempDir()
	config := &Config{
		GSPath:      "gs://my-bucket/plugins.zip",
		Output:      tmpDir,
		ZipFileName: "plugins.zip",
	}
	publisher := NewPublisher(config)
	err := publisher.Publish()
	assert.NoError(t, err, "should not return error if zip file does not exist")
}

func TestUploadToGCS_PathParsing(t *testing.T) {
	// This is a limited test that doesn't actually upload.
	// It tests the parsing of the GCS path.
	p := &Publisher{}

	// Test case 1: Full path with object name
	err := p.uploadToGCS("dummy.zip", "gs://my-bucket/my-object.zip")
	// We expect an error because we can't connect to GCS, but not a path parsing error.
	assert.NotContains(t, err.Error(), "invalid GCS path")

	// Test case 2: Path ending with a slash
	err = p.uploadToGCS("dummy.zip", "gs://my-bucket/my-folder/")
	assert.NotContains(t, err.Error(), "invalid GCS path")

	// Test case 3: Invalid path (no object)
	err = p.uploadToGCS("dummy.zip", "gs://my-bucket")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid GCS path: must include bucket and object path")

	// Test case 4: Invalid scheme
	err = p.uploadToGCS("dummy.zip", "s3://my-bucket/my-object.zip")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid GCS path: must start with gs://")
}

// Helper function to create a dummy file for testing publish logic
func createDummyFile(t *testing.T, path string) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	assert.NoError(t, err)
	err = os.WriteFile(path, []byte("dummy"), 0644)
	assert.NoError(t, err)
}

func TestPublisher_Publish(t *testing.T) {
	// This is a limited test that doesn't actually upload.
	// It tests the logic of the Publish function.
	tmpDir := t.TempDir()
	config := &Config{
		GSPath:      "gs://my-bucket/plugins.zip",
		Output:      tmpDir,
		ZipFileName: "plugins.zip",
	}
	publisher := NewPublisher(config)

	// Create a dummy zip file
	createDummyFile(t, filepath.Join(tmpDir, "plugins.zip"))

	err := publisher.Publish()
	// We expect an error because we can't connect to GCS, but not a path parsing error.
	assert.NotContains(t, err.Error(), "invalid GCS path")
}
