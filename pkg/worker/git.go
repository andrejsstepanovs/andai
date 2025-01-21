package worker

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/andrejsstepanovs/andai/pkg/models"
	redminemodels "github.com/andrejsstepanovs/andai/pkg/redmine/models"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// BranchPrefix is the prefix for the branch name
const BranchPrefix = "AI"

type Git struct {
	path    string
	repo    *git.Repository
	headRef *plumbing.Reference
	Opened  bool
}

func NewGit(path string) *Git {
	return &Git{
		path: path,
	}
}

func (g *Git) SetPath(path string) {
	g.path = path
}

func (g *Git) GetPath() string {
	return g.path
}

func (g *Git) Open() error {
	var err error

	g.repo, err = git.PlainOpen(g.GetPath())
	if err != nil {
		return fmt.Errorf("failed to open git repository %s: %v", g.GetPath(), err)
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

func FindProjectGit(projectConfig models.Project, projectRepo redminemodels.Repository) (*Git, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	_, mainGoPath, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("Failed to get the current file path")
		return nil, err
	}

	paths := []string{
		projectRepo.RootURL,
		projectConfig.GitPath,
		filepath.Join(currentDir, projectConfig.GitPath),
		filepath.Join(currentDir, "repositories", projectConfig.GitPath),
		filepath.Join(mainGoPath, projectConfig.GitPath),
		filepath.Join(mainGoPath, "repositories", projectConfig.GitPath),
	}
	var gitRet *Git
	for _, path := range paths {
		//log.Printf("Trying to open git repository in %q", path)
		gitRet = NewGit(path)
		err = gitRet.Open()
		if err != nil {
			//log.Printf("failed to open git err: %v", err)
			continue
		}
		gitRet.SetPath(path)
		break
	}

	if !gitRet.Opened {
		log.Printf("failed to open git repository %s", projectRepo.RootURL)
		return nil, errors.New("failed to open git repository")
	}

	return gitRet, nil
}
