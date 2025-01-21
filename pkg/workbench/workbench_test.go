package workbench

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andrejsstepanovs/andai/pkg/worker"
	"github.com/mattn/go-redmine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkbench_PrepareWorkplace(t *testing.T) {
	// AI: Create temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "workbench-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// AI: Create a fake git repository structure
	gitDir := filepath.Join(tmpDir, ".git")
	err = os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	tests := []struct {
		name    string
		setup   func() (*Workbench, string)
		wantErr bool
	}{
		{
			name: "successful preparation with .git directory",
			setup: func() (*Workbench, string) {
				return &Workbench{
					Git: &worker.Git{
						Opened: true,
					},
					Issue: redmine.Issue{
						Id: 123,
					},
				}, gitDir
			},
			wantErr: false,
		},
		{
			name: "successful preparation with project root directory",
			setup: func() (*Workbench, string) {
				return &Workbench{
					Git: &worker.Git{
						Opened: true,
					},
					Issue: redmine.Issue{
						Id: 123,
					},
				}, tmpDir
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// AI: Save current working directory
			originalWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				err := os.Chdir(originalWd)
				require.NoError(t, err)
			}()

			wb, targetPath := tt.setup()
			wb.Git.SetPath(targetPath)

			err = wb.PrepareWorkplace()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			currentDir, err := os.Getwd()
			require.NoError(t, err)

			// AI: If path ends with .git, compare with parent directory
			expectedDir := targetPath
			if filepath.Base(targetPath) == ".git" {
				expectedDir = filepath.Dir(targetPath)
			}
			assert.Equal(t, expectedDir, currentDir)
			assert.Equal(t, expectedDir, wb.WorkingDir)
		})
	}
}

func TestWorkbench_changeDirectory(t *testing.T) {
	// AI: Create temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "workbench-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		setup   func() *Workbench
		wantErr bool
	}{
		{
			name: "change to valid directory",
			setup: func() *Workbench {
				return &Workbench{
					Git: &worker.Git{},
				}
			},
			wantErr: false,
		},
		{
			name: "attempt to change to non-existent directory",
			setup: func() *Workbench {
				return &Workbench{
					Git: &worker.Git{},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// AI: Save current working directory
			originalWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				err := os.Chdir(originalWd)
				require.NoError(t, err)
			}()

			wb := tt.setup()
			if tt.wantErr {
				wb.Git.SetPath(filepath.Join(tmpDir, "nonexistent"))
			} else {
				wb.Git.SetPath(tmpDir)
			}

			err = wb.changeDirectory()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			currentDir, err := os.Getwd()
			require.NoError(t, err)
			assert.Equal(t, tmpDir, currentDir)
			assert.Equal(t, tmpDir, wb.WorkingDir)
		})
	}
}
