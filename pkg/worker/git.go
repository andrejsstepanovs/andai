package worker

import (
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

const BranchPrefix = "AI"

type Git struct {
	Path    string
	repo    *git.Repository
	headRef *plumbing.Reference
	Opened  bool
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

	g.headRef, err = g.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD reference: %v", err)
	}

	g.Opened = true
	return nil
}

func (g *Git) BranchName(name string) string {
	return fmt.Sprintf("%s-%s", BranchPrefix, name)
}

func (g *Git) CheckoutBranch(name string) error {
	branchName := g.BranchName(name)
	branchRefName := plumbing.NewBranchReferenceName(branchName)

	// Check if the branch already exists
	_, err := g.repo.Reference(branchRefName, false)
	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		// Create the new branch reference
		newBranch := plumbing.NewHashReference(branchRefName, g.headRef.Hash())
		err = g.repo.Storer.SetReference(newBranch)
		if err != nil {
			return fmt.Errorf("failed to create branch: %v", err)
		}
		fmt.Printf("Created new branch: %s\n", branchName)
	} else if err != nil {
		return fmt.Errorf("failed to check branch existence: %v", err)
	} else {
		fmt.Printf("Branch already exists: %s\n", branchName)
	}

	// Update HEAD to point to the new branch
	headRef := plumbing.NewSymbolicReference(plumbing.HEAD, branchRefName)
	err = g.repo.Storer.SetReference(headRef)
	if err != nil {
		return fmt.Errorf("failed to update HEAD: %v", err)
	}

	fmt.Printf("Successfully checked out branch: %s\n", branchName)

	return err
}
