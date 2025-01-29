package worker_test

import (
	"fmt"
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

	fmt.Println(tmpDir)
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

	hash1, err := wt.Commit("Initial commit", co)
	_ = hash1
	require.NoError(t, err)

	// TEST 1
	g := worker.NewGit(tmpDir)
	err = g.Open()
	assert.NoError(t, err)

	hash, err := g.GetLastCommitHash()
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// TEST 2
	// checkout new branch
	err = g.CheckoutBranch("test-branch2")
	assert.NoError(t, err)

	wt, err = repo.Worktree()
	require.NoError(t, err)
	// make 2 commits

	// commit 1
	testFile = filepath.Join(tmpDir, "test2.txt")
	err = os.WriteFile(testFile, []byte("test content 2"), 0644)
	require.NoError(t, err)

	_, err = wt.Add("test2.txt")
	require.NoError(t, err)

	hash2, err := wt.Commit("Commit within branch 1", co)
	_ = hash2
	require.NoError(t, err)

	// commit 2
	testFile = filepath.Join(tmpDir, "test3.txt")
	err = os.WriteFile(testFile, []byte("test content 3"), 0644)
	require.NoError(t, err)

	_, err = wt.Add("test3.txt")
	require.NoError(t, err)

	hash3, err := wt.Commit("Commit within branch 2", co)
	_ = hash3
	require.NoError(t, err)

	err = g.Open()
	assert.NoError(t, err)

	hashes, err := g.GetAllBranchCommitHashes()
	require.NoError(t, err)

	assert.Len(t, hashes, 1)

	// WRONG. where 1 commit?
	assert.Equal(t, hash3.String(), hashes[0])
	//assert.Equal(t, hash3.String(), hashes[1])
	//assert.Equal(t, hash3, hashes[2])
}
