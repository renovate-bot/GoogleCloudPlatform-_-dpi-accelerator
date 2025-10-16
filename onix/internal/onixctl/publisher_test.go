// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package onixctl

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mocks for GCS Client ---

// mockGCSClient is a mock implementation of the gcsClient interface for testing.
type mockGCSClient struct {
	// We can add fields here to control mock behavior, e.g., to return errors.
	bucket *mockGCSBucketHandle
}

func (m *mockGCSClient) Bucket(name string) gcsBucketHandle {
	// In a real test, you might check if the bucket name is correct.
	if m.bucket == nil {
		m.bucket = &mockGCSBucketHandle{
			objData: new(bytes.Buffer), // a buffer to simulate the uploaded object
		}
	}
	return m.bucket
}

// mockGCSBucketHandle is a mock implementation of the gcsBucketHandle interface.
type mockGCSBucketHandle struct {
	objData *bytes.Buffer
}

func (m *mockGCSBucketHandle) Object(name string) *storage.ObjectHandle {
		return &storage.ObjectHandle{}
}

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

func TestUploadToGCSWithClient_PathParsing(t *testing.T) {
	p := NewPublisher(&Config{})
	mockClient := &mockGCSClient{}
	ctx := context.Background()

	// Create a dummy file that can be "uploaded"
	tmpDir := t.TempDir()
	dummyFilePath := filepath.Join(tmpDir, "dummy.zip")
	createDummyFile(t, dummyFilePath)

	testCases := []struct {
		name       string
		gsPath     string
		wantErrMsg string
	}{
		{
			name:       "valid path",
			gsPath:     "gs://my-bucket/my-object.zip",
			wantErrMsg: "", 
		},
		{
			name:       "valid path with folder",
			gsPath:     "gs://my-bucket/my-folder/",
			wantErrMsg: "", 
		},
		{
			name:       "invalid path without object",
			gsPath:     "gs://my-bucket",
			wantErrMsg: "invalid GCS path: must include bucket and object path",
		},
		{
			name:       "invalid path with empty bucket",
			gsPath:     "gs:///my-object.zip",
			wantErrMsg: "invalid GCS path: must include bucket and object path",
		},
		{
			name:       "invalid scheme",
			gsPath:     "s3://my-bucket/my-object.zip",
			wantErrMsg: "invalid GCS path: must start with gs://",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := p.uploadToGCSWithClient(ctx, mockClient, dummyFilePath, tc.gsPath)

			if tc.wantErrMsg == "" {
				if err != nil {
					assert.NotContains(t, err.Error(), "invalid GCS path")
				}
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrMsg)
			}
		})
	}
}

// Helper function to create a dummy file for testing publish logic
func createDummyFile(t *testing.T, path string) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(path, []byte("dummy"), 0644)
	require.NoError(t, err)
}