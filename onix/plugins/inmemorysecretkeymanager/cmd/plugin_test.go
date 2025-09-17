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

package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/beckn/beckn-onix/pkg/model"
)

func TestParseConfig(t *testing.T) {
	testCases := []struct {
		name              string
		config            map[string]string
		wantProjectID     string
		wantPrivateKeyTTL int
		wantPublicKeyTTL  int
	}{
		{
			name: "valid full config",
			config: map[string]string{
				"projectID":                 "test-p",
				"privateKeyCacheTTLSeconds": "120",
				"publicKeyCacheTTLSeconds":  "240",
			},
			wantProjectID:     "test-p",
			wantPrivateKeyTTL: 120,
			wantPublicKeyTTL:  240,
		},
		{
			name:              "valid config with defaults",
			config:            map[string]string{"projectID": "test-p"},
			wantProjectID:     "test-p",
			wantPrivateKeyTTL: 15,   // Default
			wantPublicKeyTTL:  3600, // Default
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseConfig(tc.config)
			if err != nil {
				t.Fatalf("parseConfig() returned unexpected error: %v", err)
			}
			if got.ProjectID != tc.wantProjectID {
				t.Errorf("got ProjectID = %q, want %q", got.ProjectID, tc.wantProjectID)
			}
			if got.CacheTTL.PrivateKeysSeconds != tc.wantPrivateKeyTTL {
				t.Errorf("got PrivateKeysSeconds = %d, want %d", got.CacheTTL.PrivateKeysSeconds, tc.wantPrivateKeyTTL)
			}
			if got.CacheTTL.PublicKeysSeconds != tc.wantPublicKeyTTL {
				t.Errorf("got PublicKeysSeconds = %d, want %d", got.CacheTTL.PublicKeysSeconds, tc.wantPublicKeyTTL)
			}
		})
	}
}

func TestParseConfig_Errors(t *testing.T) {
	testCases := []struct {
		name    string
		config  map[string]string
		wantErr string
	}{
		{
			name:    "missing projectID",
			config:  map[string]string{"privateKeyCacheTTLSeconds": "120"},
			wantErr: "projectID not found or is empty in config",
		},
		{
			name:    "invalid private key TTL value",
			config:  map[string]string{"projectID": "test-p", "privateKeyCacheTTLSeconds": "abc"},
			wantErr: "invalid value for privateKeyCacheTTLSeconds: \"abc\", must be a positive integer",
		},
		{
			name:    "zero public key TTL value",
			config:  map[string]string{"projectID": "test-p", "publicKeyCacheTTLSeconds": "0"},
			wantErr: "must be a positive integer",
		},
		{
			name:    "negative private key TTL value",
			config:  map[string]string{"projectID": "test-p", "privateKeyCacheTTLSeconds": "-10"},
			wantErr: "must be a positive integer",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseConfig(tc.config)
			if err == nil {
				t.Fatal("parseConfig() expected an error, but got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("expected error message to contain %q, but got %q", tc.wantErr, err.Error())
			}
		})
	}
}

// --- Mocks for testing the provider's New method ---

type mockCache struct{}

func (m *mockCache) Get(ctx context.Context, key string) (string, error)                            { return "", nil }
func (m *mockCache) Set(ctx context.Context, key string, value string, expiration time.Duration) error { return nil }
func (m *mockCache) Delete(ctx context.Context, key string) error                                  { return nil }
func (m *mockCache) Clear(ctx context.Context) error                                               { return nil }
func (m *mockCache) Close() error                                                                  { return nil }

type mockRegistry struct{}

func (m *mockRegistry) Lookup(ctx context.Context, req *model.Subscription) ([]model.Subscription, error) {
	return nil, nil
}

// --- Tests for keyMgrProvider ---

func TestKeyMgrProviderNew(t *testing.T) {
	config := map[string]string{
		"projectID":                 "test-project-from-provider",
		"privateKeyCacheTTLSeconds": "60",
		"publicKeyCacheTTLSeconds":  "120",
	}
	kp := keyMgrProvider{}

	// This test is environment-agnostic. It will pass if New() succeeds,
	// or if it fails with the expected error when credentials are not available.
	_, closer, err := kp.New(context.Background(), &mockCache{}, &mockRegistry{}, config)

	if err != nil && !strings.Contains(err.Error(), "failed to create secret manager client") {
		t.Errorf("expected secret manager client creation error, but got: %v", err)
	}

	if err == nil && closer == nil {
		t.Error("expected a non-nil closer function on successful creation")
	}
}

func TestKeyMgrProviderNew_Error(t *testing.T) {
	config := map[string]string{
		// Missing projectID, which will cause parseConfig to fail.
	}
	kp := keyMgrProvider{}
	_, _, err := kp.New(context.Background(), &mockCache{}, &mockRegistry{}, config)

	if err == nil {
		t.Fatal("expected error for invalid config, got nil")
	}
	if !strings.Contains(err.Error(), "projectID not found") {
		t.Errorf("expected error about missing projectID, got: %v", err)
	}
}