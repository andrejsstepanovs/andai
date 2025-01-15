package work

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/andrejsstepanovs/andai/pkg/exec"
	"github.com/andrejsstepanovs/andai/pkg/llm"
	"github.com/andrejsstepanovs/andai/pkg/models"
	model "github.com/andrejsstepanovs/andai/pkg/redmine"
	"github.com/andrejsstepanovs/andai/pkg/worker"
	"github.com/mattn/go-redmine"
)

type WorkOnIssue struct {
	model      *model.Model
	llm        *llm.LLM
	issue      redmine.Issue
	project    redmine.Project
	projectCfg models.Project
	git        *worker.Git
	state      models.State
	issueType  models.IssueType
	job        models.Job
	workingDir string
}

func NewWorkOnIssue(
	model *model.Model,
	llm *llm.LLM,
	issue redmine.Issue,
	project redmine.Project,
	projectCfg models.Project,
	git *worker.Git,
	state models.State,
	issueType models.IssueType,
) *WorkOnIssue {
	return &WorkOnIssue{
		model:      model,
		llm:        llm,
		issue:      issue,
		project:    project,
		projectCfg: projectCfg,
		git:        git,
		state:      state,
		issueType:  issueType,
		job:        issueType.Jobs.Get(models.StateName(issue.Status.Name)),
	}
}

func (i *WorkOnIssue) Work() bool {
	log.Printf("Working on %q %q (ID: %d)", i.state.Name, i.issueType.Name, i.issue.Id)

	err := i.PrepareWorkplace()
	if err != nil {
		log.Printf("Failed to prepare workplace: %v", err)
		return false
	}

	fmt.Printf("Total steps: %d\n", len(i.job.Steps))
	for j, step := range i.job.Steps {
		fmt.Printf("Step: %d\n", j+1)
		err = i.action(step)
		if err != nil {
			log.Printf("Failed to action step: %v", err)
			return false
		}
		fmt.Println("Success")
	}

	stdout, stderr, err := exec.Exec("bobik", "once quiet llm", "hi!")
	//stdout, stderr, err := exec.Exec("bobik", "once quiet llm", "hi!")
	if err != nil {
		log.Printf("Failed to execute command: %v", err)
		return false
	}
	fmt.Printf("stdout: %s", stdout)
	fmt.Printf("stdERR: %s", stderr)

	return true
}

func (i *WorkOnIssue) AddComment(text string) error {
	err := i.model.Comment(i.issue, text)
	if err != nil {
		return fmt.Errorf("failed to comment issue err: %v", err)
	}
	return nil
}

func (i *WorkOnIssue) action(step models.Step) error {
	switch step.Command {
	case "architect":
		return i.architectAction(step.Args)
	case "coder":
		return i.coderAction(step.Args)
	case "aid":
		return i.coderAction(step.Args)
	case "aider":
		return i.coderAction(step.Args)
	default:
		return fmt.Errorf("unknown aider: %q", step.Command)
	}
}

func (i *WorkOnIssue) architectAction(prompt string) error {
	fmt.Println(prompt)
	return nil
}

func (i *WorkOnIssue) coderAction(prompt string) error {
	fmt.Println(prompt)
	return nil
}

func (i *WorkOnIssue) PrepareWorkplace() error {
	err := i.changeDirectory()
	if err != nil {
		log.Printf("Failed to change directory: %v", err)
		return err
	}
	err = i.checkoutBranch()
	if err != nil {
		log.Printf("Failed to checkout branch: %v", err)
		return err
	}
	return nil
}

func (i *WorkOnIssue) changeDirectory() error {
	targetPath := i.git.Path
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
	i.workingDir = targetPath

	return nil
}

func (i *WorkOnIssue) checkoutBranch() error {
	err := i.git.CheckoutBranch(strconv.Itoa(i.issue.Id))
	if err != nil {
		return fmt.Errorf("failed to checkout branch err: %v", err)
	}
	return nil
}

func (i *WorkOnIssue) getComments() ([]string, error) {
	comments, err := i.model.DBGetComments(i.issue.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments err: %v", err)
	}

	//log.Printf("Comments: %s", i.issue.Notes)
	//fmt.Printf("Comments: %d\n", len(comments))
	//fmt.Printf("%s\n", strings.Join(comments, "\n"))

	return comments, nil
}
