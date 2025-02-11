package workbench

import (
	"errors"
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
	GetLastCommits(count int) ([]string, error)
	GetLastCommitHash() (string, error)
	BranchName(issueID int) string
	CheckoutBranch(name string) error
	GetPath() string
	SetPath(path string)
	Reload()
	DeleteBranch(string) error
}

func (i *Workbench) PrepareWorkplace(parentIssueID *int) error {
	err := i.changeDirectory()
	if err != nil {
		log.Printf("Failed to change directory: %v", err)
		return err
	}

	i.Git.Reload()

	if parentIssueID != nil {
		branchName := i.GetIssueBranchName(redmine.Issue{Id: *parentIssueID})
		err = i.checkoutBranch(branchName)
		if err != nil {
			log.Printf("Failed to checkout parent branch: %v", err)
		}
		i.Git.Reload()
	}

	branchName := i.GetIssueBranchName(i.Issue)
	err = i.checkoutBranch(branchName)
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

func (i *Workbench) checkoutBranch(branchName string) error {
	err := i.Git.CheckoutBranch(branchName)
	if err != nil {
		return fmt.Errorf("failed to checkout branch err: %v", err)
	}
	return nil
}

func (i *Workbench) GetIssueBranchName(issue redmine.Issue) string {
	return i.Git.BranchName(issue.Id)
}

func (i *Workbench) DeleteBranch(branch string) error {
	return i.Git.DeleteBranch(branch)
}

// GetLastCommit returns the last commit hash
func (i *Workbench) GetLastCommit() (string, error) {
	return i.Git.GetLastCommitHash()
}

func (i *Workbench) GetCommitsSinceInReverseOrder(sinceSha string) ([]string, error) {
	// from newest to oldest
	allCommits, err := i.GetBranchCommits(100)
	if err != nil {
		return nil, errors.New("failed to get last commits")
	}

	commits := make([]string, 0)
	for _, sha := range allCommits {
		if sha == sinceSha { // until we find the last commit
			break
		}
		commits = append(commits, sha)
	}

	// reverse the order of commits
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}

	return commits, nil
}

// GetBranchCommits last is newest. First commit is the newest one.
func (i *Workbench) GetBranchCommits(count int) ([]string, error) {
	i.Git.Reload()
	commits, err := i.Git.GetLastCommits(count)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch commits err: %v", err)
	}
	return commits, nil
}
