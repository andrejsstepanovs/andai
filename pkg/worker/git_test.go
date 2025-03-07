package worker_test

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrejsstepanovs/andai/pkg/worker"
	gitlib "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

func TestNewGit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	require.NoError(t, err)
	os.RemoveAll(tmpDir)

	t.Run("init git repo", func(t *testing.T) {
		defer os.RemoveAll(tmpDir)

		repo, err := gitlib.PlainInit(tmpDir, false)
		require.NoError(t, err)

		hash1 := commit(t, repo, tmpDir, "Initial commit to main")
		hash2 := commit(t, repo, tmpDir, "Second commit to main")

		// TEST getting last commit hash
		g := worker.NewGit(tmpDir)
		err = g.Open()
		require.NoError(t, err)

		hash, err := g.GetLastCommitHash()
		require.NoError(t, err)
		require.Equal(t, hash2, hash)
		_ = hash1

		// TEST getting all branch commit hashes
		err = g.CheckoutBranch("BRANCH-1")
		require.NoError(t, err)

		hash3 := commit(t, repo, tmpDir, "Branch commit 1")
		hash4 := commit(t, repo, tmpDir, "Branch commit 2")

		// TEST getting all branch commits
		err = g.Open()
		require.NoError(t, err)
		err = g.CheckoutBranch("BRANCH-1")
		require.NoError(t, err)

		hashes, err := g.GetLastCommits(2)
		require.NoError(t, err)

		require.Len(t, hashes, 2)
		require.Equal(t, hash4, hashes[0])
		require.Equal(t, hash3, hashes[1])

		_, err = g.GetLastCommits(200)
		require.NoError(t, err)
	})
}

func commit(t *testing.T, repo *gitlib.Repository, dir string, commitMessage string) string {
	// Create an initial main commit so we have a HEAD reference
	wt, err := repo.Worktree()
	require.NoError(t, err)

	// Create a test file and commit it
	testFileName := randomString("test.txt")
	testFile := filepath.Join(dir, testFileName)
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	_, err = wt.Add(testFileName)
	require.NoError(t, err)

	co := &gitlib.CommitOptions{Author: &object.Signature{}}

	hash1, err := wt.Commit(commitMessage, co)
	require.NoError(t, err)

	return hash1.String()
}

func randomString(postfix string) string {
	b := make([]byte, 12)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)[2:12] + "-" + postfix
}
