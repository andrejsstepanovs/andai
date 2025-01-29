package workbench

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mattn/go-redmine"
)

type Workbench struct {
	Git        git
	Issue      redmine.Issue
	WorkingDir string
}

type git interface {
	GetAllBranchCommitHashes() ([]string, error)
	GetLastCommitHash() (string, error)
	BranchName(issueID int) string
	CheckoutBranch(name string) error
	GetPath() string
	SetPath(path string)
	Reload()
}

func (i *Workbench) PrepareWorkplace() error {
	err := i.changeDirectory()
	if err != nil {
		log.Printf("Failed to change directory: %v", err)
		return err
	}

	i.Git.Reload()
	err = i.checkoutBranch()
	if err != nil {
		log.Printf("Failed to checkout branch: %v", err)
		return err
	}

	return nil
}

func (i *Workbench) changeDirectory() error {
	targetPath := i.Git.GetPath()
	if filepath.Base(targetPath) == ".git" {
		targetPath = filepath.Dir(targetPath)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory err: %v", err)
	}
	if currentDir != targetPath {
		log.Printf("Changing directory from %s to %s\n", currentDir, targetPath)
	}

	err = os.Chdir(targetPath)
	if err != nil {
		return fmt.Errorf("failed to change directory err: %v", err)
	}

	log.Printf("Active in project directory %s\n", targetPath)
	i.WorkingDir = targetPath

	return nil
}

func (i *Workbench) checkoutBranch() error {
	branchName := i.GetIssueBranchName(i.Issue)
	err := i.Git.CheckoutBranch(branchName)
	if err != nil {
		return fmt.Errorf("failed to checkout branch err: %v", err)
	}
	return nil
}

func (i *Workbench) GetIssueBranchName(issue redmine.Issue) string {
	return i.Git.BranchName(issue.Id)
}

func (i *Workbench) GetBranchCommits() ([]string, error) {
	i.Git.Reload()
	commits, err := i.Git.GetAllBranchCommitHashes()
	if err != nil {
		return nil, fmt.Errorf("failed to get branch commits err: %v", err)
	}

	// reverse the order of commits
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}
	return commits, nil
}
