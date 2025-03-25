package workbench

import (
	"os"
	"path/filepath"
	"testing"

	gitlib "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/mattn/go-redmine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/andrejsstepanovs/andai/internal/worker"
)

func TestWorkbench_PrepareWorkplace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "workbench-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Initialize git repository
	repo, err := gitlib.PlainInit(tmpDir, false)
	require.NoError(t, err)

	// Create an initial commit so we have a HEAD reference
	wt, err := repo.Worktree()
	require.NoError(t, err)

	// Create a test file and commit it
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	_, err = wt.Add("test.txt")
	require.NoError(t, err)

	co := &gitlib.CommitOptions{Author: &object.Signature{}}

	_, err = wt.Commit("Initial commit", co)
	require.NoError(t, err)

	tests := []struct {
		name    string
		setup   func() (*Workbench, string)
		wantErr bool
	}{
		{
			name: "successful preparation with .git directory",
			setup: func() (*Workbench, string) {
				g := worker.NewGit(tmpDir)
				err := g.Open()
				require.NoError(t, err)

				return &Workbench{
					Git: g,
					Issue: redmine.Issue{
						Id: 123,
					},
				}, tmpDir
			},
			wantErr: false,
		},
		{
			name: "successful preparation with project root directory",
			setup: func() (*Workbench, string) {
				g := worker.NewGit(tmpDir)
				err := g.Open()
				require.NoError(t, err)

				return &Workbench{
					Git: g,
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
			originalWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				err := os.Chdir(originalWd)
				require.NoError(t, err)
			}()

			wb, targetPath := tt.setup()
			wb.Git.SetPath(targetPath)

			err = wb.PrepareWorkplace(nil, "")
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			currentDir, err := os.Getwd()
			require.NoError(t, err)

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
