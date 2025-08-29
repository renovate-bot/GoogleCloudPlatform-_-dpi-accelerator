package onixctl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWorkspace(t *testing.T) {
	ws, err := NewWorkspace()
	assert.NoError(t, err)
	assert.NotNil(t, ws)
	defer ws.Close()

	// Check if the directory was created
	_, err = os.Stat(ws.Path())
	assert.NoError(t, err, "workspace directory should exist")
}

func TestWorkspace_Close(t *testing.T) {
	ws, err := NewWorkspace()
	assert.NoError(t, err)
	assert.NotNil(t, ws)

	path := ws.Path()
	err = ws.Close()
	assert.NoError(t, err)

	// Check if the directory was removed
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err), "workspace directory should not exist after close")
}

func TestWorkspace_Path(t *testing.T) {
	ws, err := NewWorkspace()
	assert.NoError(t, err)
	assert.NotNil(t, ws)
	defer ws.Close()

	// Check that the path is not empty and is absolute
	assert.NotEmpty(t, ws.Path())
	assert.True(t, filepath.IsAbs(ws.Path()), "workspace path should be absolute")
}

func TestRunCommand(t *testing.T) {
	ws, err := NewWorkspace()
	assert.NoError(t, err)
	defer ws.Close()

	err = ws.runCommand(ws.Path(), "ls")
	assert.NoError(t, err)
}
