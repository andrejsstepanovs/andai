package exec

import (
	"os"
	"path/filepath"
	"testing"

	gitlib "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/mattn/go-redmine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkbench_GetIssueBranchNameOverride(t *testing.T) {
	tests := []struct {
		name     string
		issue    redmine.Issue
		expected string
	}{
		{
			name:     "no custom fields",
			issue:    redmine.Issue{Id: 1},
			expected: "",
		},
		{
			name: "custom fields without Branch field",
			issue: redmine.Issue{
				Id: 2,
				CustomFields: []*redmine.CustomField{
					{Id: 1, Name: "OtherField", Value: "value"},
				},
			},
			expected: "",
		},
		{
			name: "Branch field with nil value",
			issue: redmine.Issue{
				Id: 3,
				CustomFields: []*redmine.CustomField{
					{Id: 1, Name: "Branch", Value: nil},
				},
			},
			expected: "",
		},
		{
			name: "Branch field with empty string value",
			issue: redmine.Issue{
				Id: 4,
				CustomFields: []*redmine.CustomField{
					{Id: 1, Name: "Branch", Value: ""},
				},
			},
			expected: "",
		},
		{
			name: "Branch field with whitespace string value",
			issue: redmine.Issue{
				Id: 5,
				CustomFields: []*redmine.CustomField{
					{Id: 1, Name: "Branch", Value: "   "},
				},
			},
			expected: "",
		},
		{
			name: "Branch field with valid value",
			issue: redmine.Issue{
				Id: 6,
				CustomFields: []*redmine.CustomField{
					{Id: 1, Name: "Branch", Value: "feature/override-branch"},
				},
			},
			expected: "feature/override-branch",
		},
		{
			name: "Branch field with valid value and other fields",
			issue: redmine.Issue{
				Id: 7,
				CustomFields: []*redmine.CustomField{
					{Id: 1, Name: "OtherField", Value: "value"},
					{Id: 2, Name: "Branch", Value: "  another-override  "},
					{Id: 3, Name: "YetAnother", Value: "123"},
				},
			},
			expected: "another-override",
		},
	}

	wb := &Workbench{} // No GitInterface needed for this method

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := wb.GetIssueBranchNameOverride(tt.issue)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestWorkbench_PrepareWorkplace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "workbench-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tmpDir)
		require.NoError(t, err)
	}()

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
				g := NewGit(tmpDir)
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
				g := NewGit(tmpDir)
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

			err = wb.PrepareWorkplace("")
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
	defer func() {
		err := os.RemoveAll(tmpDir)
		require.NoError(t, err)
	}()

	tests := []struct {
		name    string
		setup   func() *Workbench
		wantErr bool
	}{
		{
			name: "change to valid directory",
			setup: func() *Workbench {
				return &Workbench{
					Git: &Git{},
				}
			},
			wantErr: false,
		},
		{
			name: "attempt to change to non-existent directory",
			setup: func() *Workbench {
				return &Workbench{
					Git: &Git{},
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
