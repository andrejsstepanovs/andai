package worker

import (
	"fmt"

	"github.com/go-git/go-git"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

const BranchPrefix = "AI"

type Git struct {
	Path     string
	repo     *git.Repository
	ref      *plumbing.Reference
	worktree *git.Worktree
}

func NewGit(path string) *Git {
	return &Git{
		Path: path,
	}
}

func (g *Git) Open() error {
	var err error

	g.repo, err = git.PlainOpen(g.Path)
	if err != nil {
		return fmt.Errorf("failed to open git repository %s: %v", g.Path, err)
	}

	g.ref, err = g.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD reference: %v", err)
	}

	g.worktree, err = g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %v", err)
	}

	return nil
}

func (g *Git) BranchName(name string) string {
	return fmt.Sprintf("%s-%s", BranchPrefix, name)
}

func (g *Git) CheckoutBranch(name string) error {
	name = g.BranchName(name)
	err := g.worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(name),
		Create: true,
		Force:  false,
	})

	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %v", name, err)
	}

	return err
}
