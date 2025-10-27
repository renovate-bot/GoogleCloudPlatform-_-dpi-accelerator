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
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockCommandRunner is a mock for testing that captures the command.
type MockCommandRunner struct {
	CommandsRun [][]string
	ShouldError error
}

// Run captures the command arguments instead of executing them.
func (m *MockCommandRunner) Run(cmd *exec.Cmd) error {
	m.CommandsRun = append(m.CommandsRun, cmd.Args)
	return m.ShouldError
}

func TestZipAndCopyPlugins(t *testing.T) {
	// 1. Setup
	mockRunner := &MockCommandRunner{}
	wsPath := t.TempDir()
	outputPath := t.TempDir()
	pluginDir := filepath.Join(wsPath, "plugins_out")
	require.NoError(t, os.MkdirAll(pluginDir, 0755))
	_, err := os.Create(filepath.Join(pluginDir, "myplugin.so")) // Need a dummy file to trigger zip
	require.NoError(t, err)

	config := &Config{
		ZipFileName: "plugins.zip",
	}
	builder := &Builder{
		config:        config,
		workspacePath: wsPath,
		outputPath:    outputPath,
		runner:        mockRunner,
	}

	// 2. Execute
	err = builder.zipAndCopyPlugins()

	// 3. Assert
	require.NoError(t, err)
	require.Len(t, mockRunner.CommandsRun, 1, "Expected one command to be run")

	expectedCmd := []string{"zip", "-r", filepath.Join(outputPath, "plugins.zip"), "."}
	assert.Equal(t, expectedCmd, mockRunner.CommandsRun[0])
}

// TestZipAndCopyPlugins_NoPlugins doesn't change as it doesn't run external commands.
func TestZipAndCopyPlugins_NoPlugins(t *testing.T) {
	wsPath := t.TempDir()
	outputPath := t.TempDir()
	pluginDir := filepath.Join(wsPath, "plugins_out")
	require.NoError(t, os.MkdirAll(pluginDir, 0755))

	builder := &Builder{
		config:        &Config{},
		workspacePath: wsPath,
		outputPath:    outputPath,
		runner:        &MockCommandRunner{},
	}

	err := builder.zipAndCopyPlugins()
	assert.NoError(t, err)
}

func TestNewBuilder(t *testing.T) {
	config := &Config{
		Output: "./test-output",
	}
	wsPath := t.TempDir()
	defer func() {
		if err := os.RemoveAll(config.Output); err != nil {
			// Replace 'log.Printf' with your actual project's logger if needed
			log.Printf("Error cleaning up output directory %s: %v", config.Output, err)
		}
	}()

	builder, err := NewBuilder(config, wsPath)
	require.NoError(t, err)
	require.NotNil(t, builder)

	assert.DirExists(t, config.Output, "output directory should be created")
}

func TestBuildPluginsInDocker(t *testing.T) {
	// 1. Setup
	mockRunner := &MockCommandRunner{}
	wsPath := t.TempDir()
	config := &Config{
		GoVersion: "1.24",
		Modules: []Module{
			{Plugins: map[string]string{"myplugin": "cmd/myplugin"}},
		},
	}
	builder := &Builder{
		config:        config,
		workspacePath: wsPath,
		runner:        mockRunner,
	}

	// 2. Execute
	err := builder.buildPluginsInDocker()

	// 3. Assert
	require.NoError(t, err)
	require.Len(t, mockRunner.CommandsRun, 1, "Expected one command to be run")

	expectedCmdPrefix := []string{
		"docker", "run", "--rm", "--platform", "linux/amd64",
		"-v", wsPath + ":/workspace",
		"-w", "/workspace",
		"golang:1.24-bullseye",
		"sh", "./build_plugins.sh",
	}
	assert.Equal(t, expectedCmdPrefix, mockRunner.CommandsRun[0])
}

func TestBuildImagesLocally_WithRegistry(t *testing.T) {
	// 1. Setup
	mockRunner := &MockCommandRunner{}
	wsPath := t.TempDir()
	dockerfileDir := filepath.Join(wsPath, "app")
	require.NoError(t, os.MkdirAll(dockerfileDir, 0755))
	_, err := os.Create(filepath.Join(dockerfileDir, "Dockerfile"))
	require.NoError(t, err)

	config := &Config{
		Registry: "my-registry.com/project",
		Modules: []Module{
			{
				DirName: "app",
				Images: map[string]Image{
					"myimage": {Dockerfile: "Dockerfile", Tag: "v1"},
				},
			},
		},
	}
	builder := &Builder{
		config:        config,
		workspacePath: wsPath,
		runner:        mockRunner,
	}

	// 2. Execute
	err = builder.buildImagesLocally()

	// 3. Assert
	require.NoError(t, err)
	require.Len(t, mockRunner.CommandsRun, 2, "Expected build and push commands")

	expectedBuildCmd := []string{
		"docker", "buildx", "build", "--platform", "linux/amd64", "--load",
		"-t", "my-registry.com/project/myimage:v1",
		"-f", "Dockerfile", ".",
	}
	expectedPushCmd := []string{"docker", "push", "my-registry.com/project/myimage:v1"}

	assert.Equal(t, expectedBuildCmd, mockRunner.CommandsRun[0])
	assert.Equal(t, expectedPushCmd, mockRunner.CommandsRun[1])
}

func TestBuildImagesLocally_NoRegistry(t *testing.T) {
	// 1. Setup
	mockRunner := &MockCommandRunner{}
	wsPath := t.TempDir()
	dockerfileDir := filepath.Join(wsPath, "app")
	require.NoError(t, os.MkdirAll(dockerfileDir, 0755))
	_, err := os.Create(filepath.Join(dockerfileDir, "Dockerfile"))
	require.NoError(t, err)

	config := &Config{
		Modules: []Module{
			{
				DirName: "app",
				Images: map[string]Image{
					"myimage": {Dockerfile: "Dockerfile", Tag: "v1"},
				},
			},
		},
	}
	builder := &Builder{
		config:        config,
		workspacePath: wsPath,
		runner:        mockRunner,
	}

	// 2. Execute
	err = builder.buildImagesLocally()

	// 3. Assert
	require.NoError(t, err)
	require.Len(t, mockRunner.CommandsRun, 1, "Expected only the build command")

	expectedBuildCmd := []string{
		"docker", "buildx", "build", "--platform", "linux/amd64", "--load",
		"-t", "myimage:v1",
		"-f", "Dockerfile", ".",
	}

	assert.Equal(t, expectedBuildCmd, mockRunner.CommandsRun[0])
}

func TestBuild(t *testing.T) {
	// 1. Setup
	mockRunner := &MockCommandRunner{}
	wsPath := t.TempDir()
	outputPath := t.TempDir()

	// Create dummy files and dirs for build process
	dockerfileDir := filepath.Join(wsPath, "app")
	require.NoError(t, os.MkdirAll(dockerfileDir, 0755))
	_, err := os.Create(filepath.Join(dockerfileDir, "Dockerfile"))
	require.NoError(t, err)
	pluginDir := filepath.Join(wsPath, "plugins_out")
	require.NoError(t, os.MkdirAll(pluginDir, 0755))
	_, err = os.Create(filepath.Join(pluginDir, "myplugin.so"))
	require.NoError(t, err)

	config := &Config{
		GoVersion:   "1.24",
		Registry:    "my-registry.com/project",
		ZipFileName: "plugins.zip",
		Output:      outputPath,
		Modules: []Module{
			{
				DirName: "app",
				Plugins: map[string]string{"myplugin": "cmd/myplugin"},
				Images: map[string]Image{
					"myimage": {Dockerfile: "Dockerfile", Tag: "v1"},
				},
			},
		},
	}
	builder := &Builder{
		config:        config,
		workspacePath: wsPath,
		outputPath:    outputPath,
		runner:        mockRunner,
	}

	// 2. Execute
	err = builder.Build()

	// 3. Assert
	require.NoError(t, err)
	// Expecting 4 commands:
	// 1. docker run (for plugins)
	// 2. docker buildx build (for image)
	// 3. docker push (for image)
	// 4. zip (for plugins)
	require.Len(t, mockRunner.CommandsRun, 4, "Expected four commands to be run")

	// Assert plugin build command
	assert.Contains(t, mockRunner.CommandsRun[0][0], "docker")
	assert.Contains(t, mockRunner.CommandsRun[0][1], "run")

	// Assert image build command
	assert.Contains(t, mockRunner.CommandsRun[1][0], "docker")
	assert.Contains(t, mockRunner.CommandsRun[1][1], "buildx")

	// Assert image push command
	assert.Contains(t, mockRunner.CommandsRun[2][0], "docker")
	assert.Contains(t, mockRunner.CommandsRun[2][1], "push")

	// Assert zip command
	assert.Contains(t, mockRunner.CommandsRun[3][0], "zip")
}

func TestBuild_PluginBuildFail(t *testing.T) {
	// 1. Setup
	mockRunner := &MockCommandRunner{
		ShouldError: assert.AnError,
	}
	wsPath := t.TempDir()
	outputPath := t.TempDir()

	config := &Config{
		GoVersion: "1.24",
		Modules: []Module{
			{Plugins: map[string]string{"myplugin": "cmd/myplugin"}},
		},
		Output: outputPath,
	}
	builder := &Builder{
		config:        config,
		workspacePath: wsPath,
		outputPath:    outputPath,
		runner:        mockRunner,
	}

	// 2. Execute
	err := builder.Build()

	// 3. Assert
	require.Error(t, err)
	assert.Equal(t, mockRunner.ShouldError, err)
	require.Len(t, mockRunner.CommandsRun, 1, "Expected only the plugin build command to be run")
	assert.Contains(t, mockRunner.CommandsRun[0][0], "docker")
}

func TestBuild_ImageBuildFail(t *testing.T) {
	// 1. Setup
	mockRunner := &MockCommandRunner{}
	wsPath := t.TempDir()
	outputPath := t.TempDir()
	dockerfileDir := filepath.Join(wsPath, "app")
	require.NoError(t, os.MkdirAll(dockerfileDir, 0755))
	_, err := os.Create(filepath.Join(dockerfileDir, "Dockerfile"))
	require.NoError(t, err)

	config := &Config{
		GoVersion: "1.24",
		Modules: []Module{
			{
				DirName: "app",
				Images:  map[string]Image{"myimage": {Dockerfile: "Dockerfile", Tag: "v1"}},
			},
		},
		Output: outputPath,
	}
	builder := &Builder{
		config:        config,
		workspacePath: wsPath,
		outputPath:    outputPath,
		runner:        mockRunner,
	}

	// 2. Execute
	// Simulate error only on the second command (image build)
	mockRunner.ShouldError = assert.AnError
	err = builder.buildImagesLocally()

	// 3. Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to build image")
	require.Len(t, mockRunner.CommandsRun, 1, "Expected only the image build command to be run")
	assert.Contains(t, mockRunner.CommandsRun[0][0], "docker")
}

func TestBuild_ZipFail(t *testing.T) {
	// 1. Setup
	mockRunner := &MockCommandRunner{
		ShouldError: assert.AnError,
	}
	wsPath := t.TempDir()
	outputPath := t.TempDir()
	pluginDir := filepath.Join(wsPath, "plugins_out")
	require.NoError(t, os.MkdirAll(pluginDir, 0755))
	_, err := os.Create(filepath.Join(pluginDir, "myplugin.so"))
	require.NoError(t, err)

	config := &Config{
		ZipFileName: "plugins.zip",
		Output:      outputPath,
	}
	builder := &Builder{
		config:        config,
		workspacePath: wsPath,
		outputPath:    outputPath,
		runner:        mockRunner,
	}

	// 2. Execute
	err = builder.zipAndCopyPlugins()

	// 3. Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create zip archive")
	require.Len(t, mockRunner.CommandsRun, 1, "Expected only the zip command to be run")
	assert.Contains(t, mockRunner.CommandsRun[0][0], "zip")
}

func TestNewBuilder_OutputDirCreationFail(t *testing.T) {
	// 1. Setup
	outPath := "./test-output"
	// Create a file where the directory should be
	_, err := os.Create(outPath)
	require.NoError(t, err)

	defer func() {
		if err := os.RemoveAll(outPath); err != nil {
			// Replace 'log.Printf' with your actual project's logger if needed
			log.Printf("Error cleaning up output directory %s: %v", outPath, err)
		}
	}()

	config := &Config{
		Output: outPath,
	}
	wsPath := t.TempDir()

	// 2. Execute
	_, err = NewBuilder(config, wsPath)

	// 3. Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create output directory")
}
