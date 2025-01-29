package worker_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andrejsstepanovs/andai/pkg/worker"
	gitlib "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-test-*")
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

	g := worker.NewGit(tmpDir)
	err = g.Open()
	assert.NoError(t, err)

	hash, err := g.GetLastCommitHash()
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}
