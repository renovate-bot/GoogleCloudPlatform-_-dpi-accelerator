package onixctl

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Workspace manages the temporary directory where modules are checked out and built.
type Workspace struct {
	path string
}

// NewWorkspace creates a new temporary workspace.
func NewWorkspace() (*Workspace, error) {
	tmpDir, err := os.MkdirTemp("", "onixctl-workspace-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp workspace directory: %w", err)
	}
	return &Workspace{path: tmpDir}, nil
}

// Path returns the absolute path of the workspace directory.
func (w *Workspace) Path() string {
	return w.path
}

// Close removes the temporary workspace directory.
func (w *Workspace) Close() error {
	return os.RemoveAll(w.path)
}

// PrepareModules checks out all remote modules into the workspace.
func (w *Workspace) PrepareModules(modules []Module) error {
	for _, module := range modules {
		destPath := filepath.Join(w.path, module.DirName)
		if err := os.MkdirAll(destPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory for module %s: %w", module.Name, err)
		}

		if module.Repo == "" {
			fmt.Printf("Copying local module %s from path %s...\n", module.Name, module.Path)
			sourcePath, err := filepath.Abs(module.Path)
			if err != nil {
				return fmt.Errorf("failed to get absolute path for local module %s: %w", module.Name, err)
			}

			var cmd *exec.Cmd
			if runtime.GOOS == "darwin" {
				cmd = exec.Command("cp", "-r", sourcePath+"/.", destPath)
			} else {
				cmd = exec.Command("cp", "-rT", sourcePath, destPath)
			}
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to copy local module %s: %w", module.Name, err)
			}
		} else {
			fmt.Printf("Cloning module %s from %s...\n", module.Name, module.Repo)
			// Clone into a temporary directory first
			tempClonePath, err := os.MkdirTemp("", "onixctl-clone-*")
			if err != nil {
				return fmt.Errorf("failed to create temp clone directory: %w", err)
			}
			defer os.RemoveAll(tempClonePath)

			cloneOpts := &git.CloneOptions{
				URL:      module.Repo,
				Progress: os.Stdout,
			}
			repo, err := git.PlainClone(tempClonePath, false, cloneOpts)
			if err != nil {
				return fmt.Errorf("failed to clone repo %s: %w", module.Repo, err)
			}

			if module.Version != "" {
				worktree, err := repo.Worktree()
				if err != nil {
					return fmt.Errorf("failed to get worktree for repo %s: %w", module.Repo, err)
				}
				err = worktree.Checkout(&git.CheckoutOptions{
					Hash: plumbing.NewHash(module.Version),
				})
				if err != nil {
					return fmt.Errorf("failed to checkout version %s for repo %s: %w", module.Version, module.Repo, err)
				}
			}

			// Copy the content of the root path from the clone to the workspace
			sourcePath := filepath.Join(tempClonePath, module.Path)
			var cmd *exec.Cmd
			if runtime.GOOS == "darwin" {
				cmd = exec.Command("cp", "-r", sourcePath+"/.", destPath)
			} else {
				cmd = exec.Command("cp", "-rT", sourcePath, destPath)
			}
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to copy module root for %s: %w", module.Name, err)
			}
		}
	}
	return nil
}

// SetupGoWorkspace initializes a go.work file and syncs dependencies.
func (w *Workspace) SetupGoWorkspace(modules []Module, goVersion string) error {
	fmt.Println("Creating go.work file...")
	goWorkContent := fmt.Sprintf("go %s\n\nuse (\n", goVersion)
	for _, module := range modules {
		goWorkContent += fmt.Sprintf("\t\"./%s\"\n", module.DirName)
	}
	goWorkContent += ")\n"

	goWorkPath := filepath.Join(w.path, "go.work")
	if err := os.WriteFile(goWorkPath, []byte(goWorkContent), 0644); err != nil {
		return fmt.Errorf("failed to write go.work file: %w", err)
	}

	fmt.Println("Syncing dependencies with go work sync...")
	if err := w.runCommand(w.path, "go", "work", "sync"); err != nil {
		return fmt.Errorf("failed to sync workspace dependencies: %w", err)
	}

	for _, module := range modules {
		modulePath := filepath.Join(w.path, module.DirName)
		fmt.Printf("Running go mod tidy for %s...\n", module.Name)
		if err := w.runCommand(modulePath, "go", "mod", "tidy"); err != nil {
			return fmt.Errorf("failed to run go mod tidy for module %s: %w", module.Name, err)
		}
	}

	return nil
}

// runCommand is a helper to execute shell commands.
func (w *Workspace) runCommand(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}