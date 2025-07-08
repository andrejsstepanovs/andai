package exec

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	redminemodels "github.com/andrejsstepanovs/andai/internal/redmine/models"
	"github.com/andrejsstepanovs/andai/internal/settings"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// BranchPrefix is the prefix for the branch name
const BranchPrefix = "AI"

type Git struct {
	path     string
	repo     *git.Repository
	worktree *git.Worktree
	Opened   bool
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

func (g *Git) Reload() {
	err := g.Open()
	if err != nil {
		log.Printf("failed to reload git repository: %v", err)
	}
}

func (g *Git) Open() error {
	var err error

	//log.Printf("Opening git repository at path: %s", g.GetPath())
	g.repo, err = git.PlainOpenWithOptions(g.GetPath(), &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return fmt.Errorf("failed to open git repository %s: %v", g.GetPath(), err)
	}

	g.worktree, err = g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %v", err)
	}

	g.Opened = true
	return nil
}

func (g *Git) getHeadRef() (*plumbing.Reference, error) {
	headRef, err := g.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD reference: %v", err)
	}
	return headRef, nil
}

func (g *Git) GetAffectedFiles(sha string) ([]string, error) {
	commit, err := g.repo.CommitObject(plumbing.NewHash(sha))
	if err != nil {
		log.Printf("failed to get commit object: %v", err)
		return nil, err
	}

	files := make(map[string]struct{})

	stats, err := commit.Stats()
	if err != nil {
		log.Printf("failed to get commit stats: %v", err)
		return nil, err
	}
	for _, stat := range stats {
		files[stat.Name] = struct{}{}
	}

	fileNames := make([]string, 0, len(files))
	for file := range files {
		fileNames = append(fileNames, file)
	}

	return fileNames, nil
}

// GetLastCommits returns the last n commit hashes. First commit is the most recent one.
func (g *Git) GetLastCommits(count int) ([]string, error) {
	headRef, err := g.getHeadRef()
	if err != nil {
		log.Printf("failed to get HEAD reference: %v", err)
		return nil, err
	}

	commitIter, err := g.repo.Log(&git.LogOptions{
		From: headRef.Hash(),
	})
	if err != nil {
		log.Printf("failed to get commit iterator: %v", err)
		return nil, err
	}

	var hashes []string
	for i := 0; i < count; i++ {
		commit, err := commitIter.Next()
		if err != nil {
			log.Printf("failed to get commit: %v", err)
			break
		}

		hashes = append(hashes, commit.Hash.String())
	}

	return hashes, nil
}

// GetCurrentBranchName returns the current branch name
func (g *Git) GetCurrentBranchName() (string, error) {
	ref, err := g.repo.Head()
	if err != nil {
		log.Printf("failed to get HEAD reference: %v", err)
		return "", err
	}
	if ref.Name().IsBranch() {
		return ref.Name().Short(), nil
	}

	return "", errors.New("not on a branch")
}

func (g *Git) GetLastCommitHash() (string, error) {
	headRef, err := g.getHeadRef()
	if err != nil {
		log.Printf("failed to get HEAD reference: %v", err)
		return "", err
	}

	commit, err := g.repo.CommitObject(headRef.Hash())
	if err != nil {
		log.Printf("failed to get commit object: %v", err)
		return "", err
	}

	return commit.Hash.String(), nil
}

func (g *Git) Add(path string) error {
	_, err := g.worktree.Add(path)
	return err
}

func (g *Git) Commit(message string) (string, error) {
	opts := &git.CommitOptions{
		Author: &object.Signature{},
	}

	sha, err := g.worktree.Commit(message, opts)
	if err != nil {
		return "", err
	}
	return sha.String(), nil
}

func (g *Git) BranchName(issueID int) string {
	id := strconv.Itoa(issueID)
	return fmt.Sprintf("%s-%s", BranchPrefix, id)
}

func (g *Git) DeleteBranch(branchName string) error {
	branchRefName := plumbing.NewBranchReferenceName(branchName)
	err := g.repo.Storer.RemoveReference(branchRefName)
	if err != nil {
		log.Printf("failed to delete branch %s: %v", branchName, err)
		return fmt.Errorf("failed to delete branch %s: %v", branchName, err)
	}
	return nil
}

func (g *Git) BranchExists(branchName string) (bool, error) {
	branches, err := g.repo.Branches()
	if err != nil {
		return false, fmt.Errorf("failed to get branches: %v", err)
	}

	mappedBranches := make(map[string]string)
	err = branches.ForEach(func(branch *plumbing.Reference) error {
		mappedBranches[branch.Name().Short()] = branch.Hash().String()
		return nil
	})

	_, exists := mappedBranches[branchName]
	return exists, err
}

// ExecCheckoutBranch checks out a branch or creates it if it does not exist.
// Returns true if new branch was created, false if it already existed.
func (g *Git) ExecCheckoutBranch(branchName string) (bool, error) {
	respGit, err := Exec("git", time.Second*10, "branch")
	if err != nil {
		log.Printf("stderr: %s", respGit.Stderr)
		return false, fmt.Errorf("failed to check if branch exists err: %v", err)
	}
	branches := strings.Split(respGit.Stdout, "\n")
	branchExists := false
	for _, branch := range branches {
		if strings.HasPrefix(branch, "*") && strings.TrimSpace(strings.TrimPrefix(branch, "*")) == branchName {
			log.Printf("Already on branch %s\n", branchName)
			return false, nil
		}
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

	branchCreated := !branchExists

	if checkoutErr != nil {
		exec, errDiff := Exec("git", time.Second*10, "diff", "--name-only")
		if errDiff == nil && exec.Stdout != "" {
			files := strings.Split(exec.Stdout, "\n")
			if len(files) > 0 {
				for i, file := range files {
					log.Printf("Unmerged file %d: %s", i+1, file)
				}
				return branchCreated, fmt.Errorf("failed to checkout branch. %d unmerged files detected in branch: %s", len(files), branchName)
			}
		}
		return branchCreated, fmt.Errorf("failed to checkout branch err: %v", checkoutErr)
	}

	if checkoutResp.Stderr != "" {
		log.Printf("git: %s", checkoutResp.Stderr)
	}
	if checkoutResp.Stdout != "" {
		log.Printf("git: %s", checkoutResp.Stdout)
	}

	return branchCreated, nil
}

// CheckoutBranch checks out a branch or creates if missing.
// go-git can fail if there are files that are not in gitignore.
func (g *Git) CheckoutBranch(branchName string) error {
	currentBranch, err := g.GetCurrentBranchName()
	if err != nil {
		return fmt.Errorf("failed to get current branch name: %v", err)
	}
	if currentBranch == branchName {
		log.Printf("Already on branch %s", branchName)
		return nil
	}

	exists, err := g.BranchExists(branchName)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %v", err)
	}
	log.Printf("Branch %s exists: %v", branchName, exists)

	err = g.worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: !exists,
	})

	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %v", branchName, err)
	}

	log.Printf("Checked out branch %s", branchName)
	return nil
}

func GetAllPossiblePaths(projectConfig settings.Project, projectRepo redminemodels.Repository, forGit bool) ([]string, error) {
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
		projectConfig.LocalGitPath,
		projectRepo.RootURL,
		projectConfig.GitPath,
		filepath.Join(currentDir, projectConfig.GitPath),
		filepath.Join(currentDir, "repositories", projectConfig.GitPath),
		filepath.Join(mainGoPath, projectConfig.GitPath),
		filepath.Join(mainGoPath, "repositories", projectConfig.GitPath),
		"/",
	}

	if forGit {
		return paths, nil
	}

	newPaths := make([]string, 0, len(paths))
	for _, path := range paths {
		newPaths = append(newPaths, strings.TrimSuffix(path, ".git"))
	}

	return newPaths, nil
}

func FindProjectGit(projectConfig settings.Project, projectRepo redminemodels.Repository) (*Git, error) {
	paths, err := GetAllPossiblePaths(projectConfig, projectRepo, true)
	if err != nil {
		return nil, err
	}

	var gitRet *Git
	for _, path := range paths {
		if path == "" {
			continue
		}
		//log.Printf("Trying to open git repository in %q", path)
		gitRet = NewGit(path)
		err = gitRet.Open()
		if err != nil {
			log.Printf("failed to open git err: %v", err)
			continue
		}
		gitRet.SetPath(path)
		break
	}

	if gitRet == nil || !gitRet.Opened {
		log.Printf("failed to find git repository location for %q", projectRepo.RootURL)
		return nil, errors.New("failed to open git repository")
	}

	return gitRet, nil
}

func IsGitInstalled() bool {
	out, err := Exec("git", time.Second*10, "--version")
	if err != nil {
		log.Printf("Git is not installed: %v", err)
		return false
	}
	log.Println(out.Stdout)
	return true
}

func IsTreeInstalled() bool {
	out, err := Exec("tree", time.Second*10, "--version")
	if err != nil {
		log.Printf("tree is not installed: %v", err)
		return false
	}
	log.Println(out.Stdout)
	return true
}

func IsAiderInstalled() bool {
	out, err := Exec("aider", time.Second*10, "--version")
	if err != nil {
		log.Printf("Aider is not installed: %v", err)
		return false
	}
	log.Println(out.Stdout)
	return true
}
