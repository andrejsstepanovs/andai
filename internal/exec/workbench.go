package exec

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	model "github.com/andrejsstepanovs/andai/internal/redmine"
	"github.com/mattn/go-redmine"
)

type Workbench struct {
	Git        GitInterface
	Issue      redmine.Issue
	WorkingDir string
}

type GitInterface interface {
	GetAffectedFiles(sha string) ([]string, error)
	GetLastCommits(count int) ([]string, error)
	GetLastCommitHash() (string, error)
	BranchName(issueID int) string
	CheckoutBranch(name string) error
	GetPath() string
	SetPath(path string)
	Reload()
	DeleteBranch(string) error
}

func (i *Workbench) GoToRepo() error {
	err := i.changeDirectory()
	if err != nil {
		log.Printf("Failed to change directory: %v", err)
		return err
	}

	i.Git.Reload()
	return nil
}

func (i *Workbench) PrepareWorkplace(parentIssueID *int, finalBranch string) error {
	err := i.GoToRepo()
	if err != nil {
		log.Printf("Failed to change directory: %v", err)
		return err
	}

	if parentIssueID != nil {
		branchName := i.GetIssueBranchName(redmine.Issue{Id: *parentIssueID})
		err = i.CheckoutBranch(branchName)
		if err != nil {
			log.Printf("Prepare workplace: failed to checkout parent branch: %v", err)
			return err
		}
	} else {
		if finalBranch != "" {
			err = i.CheckoutBranch(finalBranch)
			if err != nil {
				log.Printf("Prepare workplace: failed to checkout project final branch: %v", err)
				return err
			}
		}
	}
	i.Git.Reload()

	branchName := i.GetIssueBranchName(i.Issue)
	err = i.CheckoutBranch(branchName)
	if err != nil {
		log.Printf("Prepare workplace: failed to checkout branch: %v", err)
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

func (i *Workbench) CheckoutBranch(branchName string) error {
	resp, err := Exec("git", time.Second*10, "branch")
	if err != nil {
		log.Printf("stderr: %s", resp.Stderr)
		return fmt.Errorf("failed to check if branch exists err: %v", err)
	}
	branches := strings.Split(resp.Stdout, "\n")
	branchExists := false
	for _, branch := range branches {
		if strings.TrimSpace(branch) == branchName {
			branchExists = true
			break
		}
	}

	var checkoutResp Output
	var checkoutErr error

	if branchExists {
		log.Printf("Branch %s already exists\n", branchName)
		checkoutResp, checkoutErr = Exec("git", time.Second*10, "checkout", branchName)
	} else {
		log.Printf("Branch %s does not exist\n", branchName)
		checkoutResp, checkoutErr = Exec("git", time.Second*10, "checkout", "-b", branchName)
	}

	if checkoutErr != nil {
		return fmt.Errorf("failed to checkout branch err: %v", checkoutErr)
	}
	if checkoutResp.Stderr != "" {
		log.Printf("git: %s", resp.Stderr)
	}
	if checkoutResp.Stdout != "" {
		log.Printf("git: %s", resp.Stdout)
	}

	return nil
}

// GetIssueBranchNameOverride in UI user can set branch name override. Use it if set.
func (i *Workbench) GetIssueBranchNameOverride(issue redmine.Issue) string {
	if issue.CustomFields == nil {
		return ""
	}
	for _, field := range issue.CustomFields {
		if field.Name != model.CustomFieldBranch {
			continue
		}
		if field.Value == nil {
			continue
		}
		s := field.Value.(string)
		s = strings.TrimSpace(s)
		if s != "" {
			return s
		}
	}
	return ""
}

// GetIssueSkipMergeOverride in UI user can set flag to not merge into parent. Use it if set.
func (i *Workbench) GetIssueSkipMergeOverride(issue redmine.Issue) bool {
	if issue.CustomFields == nil {
		return false
	}
	for _, field := range issue.CustomFields {
		if field.Name != model.CustomFieldSkipMerge {
			continue
		}
		if field.Value == nil {
			continue
		}
		s := field.Value.(string)
		s = strings.TrimSpace(s)
		if s == "1" {
			return true
		}
	}
	return false
}

func (i *Workbench) GetIssueBranchName(issue redmine.Issue) string {
	overrideBranch := i.GetIssueBranchNameOverride(issue)
	if overrideBranch != "" {
		return overrideBranch
	}
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
	allCommits, err := i.GetBranchCommits(20)
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
	for k, j := 0, len(commits)-1; k < j; k, j = k+1, j-1 {
		commits[k], commits[j] = commits[j], commits[k]
	}

	// returns in order from oldest to newest
	return commits, nil
}

func (i *Workbench) GetAffectedFiles(sha string) ([]string, error) {
	return i.Git.GetAffectedFiles(sha)
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
