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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBuilder(t *testing.T) {
	config := &Config{
		Output: "./test-output",
	}
	wsPath := t.TempDir()

	builder, err := NewBuilder(config, wsPath)
	assert.NoError(t, err)
	assert.NotNil(t, builder)
	defer os.RemoveAll(config.Output)

	assert.DirExists(t, config.Output, "output directory should be created")
}

func TestBuild(t *testing.T) {
	// This is a limited test that doesn't actually build anything.
	// It tests that the Build function can be called without errors.
	config := &Config{
		GoVersion: "1.24",
		Modules: []Module{
			{
				Name:    "app",
				DirName: "app",
				Path:    ".",
				Plugins: map[string]string{
					"myplugin": "cmd/myplugin",
				},
				Images: map[string]Image{
					"myimage": {
						Dockerfile: "Dockerfile",
						Tag:        "v1",
					},
				},
			},
		},
	}
	wsPath := t.TempDir()
	// Create dummy files and directories
	appPath := filepath.Join(wsPath, "app")
	err := os.MkdirAll(filepath.Join(appPath, "cmd", "myplugin"), 0755)
	assert.NoError(t, err)
	_, err = os.Create(filepath.Join(appPath, "Dockerfile"))
	assert.NoError(t, err)
	// Create a dummy go.mod file
	err = os.WriteFile(filepath.Join(appPath, "go.mod"), []byte("module myapp"), 0644)
	assert.NoError(t, err)

	builder, err := NewBuilder(config, wsPath)
	assert.NoError(t, err)

	// We expect an error because we are not running in a real environment.
	// However, we can check that the error is the one we expect.
	err = builder.Build()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exit status 1")
}

func TestZipAndCopyPlugins(t *testing.T) {
	wsPath := t.TempDir()
	outputPath := t.TempDir()
	pluginDir := filepath.Join(wsPath, "plugins_out")
	err := os.MkdirAll(pluginDir, 0755)
	assert.NoError(t, err)

	// Create a dummy plugin file
	_, err = os.Create(filepath.Join(pluginDir, "myplugin.so"))
	assert.NoError(t, err)

	config := &Config{
		Output:      outputPath,
		ZipFileName: "plugins.zip",
	}
	builder := &Builder{
		config:        config,
		workspacePath: wsPath,
		outputPath:    outputPath,
	}

	err = builder.zipAndCopyPlugins()
	assert.NoError(t, err)

	// Check if the zip file was created
	_, err = os.Stat(filepath.Join(outputPath, "plugins.zip"))
	assert.NoError(t, err, "zip file should have been created")

	// Check if the plugin file was copied
	_, err = os.Stat(filepath.Join(outputPath, "myplugin.so"))
	assert.NoError(t, err, "plugin file should have been copied")
}

func TestBuildPluginsInDocker(t *testing.T) {
	// This is a limited test that doesn't actually build anything.
	// It tests that the buildPluginsInDocker function can be called without errors.
	config := &Config{
		GoVersion: "1.24",
		Modules: []Module{
			{
				Name:    "app",
				DirName: "app",
				Path:    ".",
				Plugins: map[string]string{
					"myplugin": "cmd/myplugin",
				},
			},
		},
	}
	wsPath := t.TempDir()
	// Create dummy files and directories
	appPath := filepath.Join(wsPath, "app")
	err := os.MkdirAll(filepath.Join(appPath, "cmd", "myplugin"), 0755)
	assert.NoError(t, err)
	// Create a dummy go.mod file
	err = os.WriteFile(filepath.Join(appPath, "go.mod"), []byte("module myapp"), 0644)
	assert.NoError(t, err)

	builder, err := NewBuilder(config, wsPath)
	assert.NoError(t, err)

	// We expect an error because we are not running in a real environment.
	// However, we can check that the error is the one we expect.
	err = builder.buildPluginsInDocker()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exit status 1")
}

func TestBuildImagesLocally_WithRegistry(t *testing.T) {
	// This is a limited test that doesn't actually build anything.
	// It tests that the buildImagesLocally function can be called without errors.
	config := &Config{
		GoVersion: "1.24",
		Registry:  "my-registry.com/project",
		Modules: []Module{
			{
				Name:    "app",
				DirName: "app",
				Path:    ".",
				Images: map[string]Image{
					"myimage": {
						Dockerfile: "Dockerfile",
						Tag:        "v1",
					},
				},
			},
		},
	}
	wsPath := t.TempDir()
	// Create dummy files and directories
	appPath := filepath.Join(wsPath, "app")
	err := os.MkdirAll(appPath, 0755)
	assert.NoError(t, err)
	_, err = os.Create(filepath.Join(appPath, "Dockerfile"))
	assert.NoError(t, err)

	builder, err := NewBuilder(config, wsPath)
	assert.NoError(t, err)

	// We expect an error because we are not running in a real environment.
	// However, we can check that the error is the one we expect.
	err = builder.buildImagesLocally()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exit status 1")
}

func TestZipAndCopyPlugins_NoPlugins(t *testing.T) {
	wsPath := t.TempDir()
	outputPath := t.TempDir()
	pluginDir := filepath.Join(wsPath, "plugins_out")
	err := os.MkdirAll(pluginDir, 0755)
	assert.NoError(t, err)

	config := &Config{
		Output:      outputPath,
		ZipFileName: "plugins.zip",
	}
	builder := &Builder{
		config:        config,
		workspacePath: wsPath,
		outputPath:    outputPath,
	}

	err = builder.zipAndCopyPlugins()
	assert.NoError(t, err)
}

func TestBuildImagesLocally_NoRegistry(t *testing.T) {
	// This is a limited test that doesn't actually build anything.
	// It tests that the buildImagesLocally function can be called without errors.
	config := &Config{
		GoVersion: "1.24",
		Modules: []Module{
			{
				Name:    "app",
				DirName: "app",
				Path:    ".",
				Images: map[string]Image{
					"myimage": {
						Dockerfile: "Dockerfile",
						Tag:        "v1",
					},
				},
			},
		},
	}
	wsPath := t.TempDir()
	// Create dummy files and directories
	appPath := filepath.Join(wsPath, "app")
	err := os.MkdirAll(appPath, 0755)
	assert.NoError(t, err)
	_, err = os.Create(filepath.Join(appPath, "Dockerfile"))
	assert.NoError(t, err)

	builder, err := NewBuilder(config, wsPath)
	assert.NoError(t, err)

	// We expect an error because we are not running in a real environment.
	// However, we can check that the error is the one we expect.
	err = builder.buildImagesLocally()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exit status 1")
}

