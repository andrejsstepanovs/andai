package work

import (
	"fmt"
	"log"
	"strconv"

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

	err := i.CheckoutBranch()
	if err != nil {
		log.Printf("Failed to checkout branch: %v", err)
		return false
	}

	fmt.Printf("%v\n", i.issueType)
	fmt.Printf("%d\n", len(i.job.Steps))

	for _, step := range i.job.Steps {
		fmt.Printf("Step: %s\n", step.Prompt)
	}

	return true
}

func (i *WorkOnIssue) CheckoutBranch() error {
	err := i.git.CheckoutBranch(strconv.Itoa(i.issue.Id))
	if err != nil {
		return fmt.Errorf("failed to checkout branch err: %v", err)
	}
	return nil
}

func (i *WorkOnIssue) GetComments() ([]string, error) {
	comments, err := i.model.DBGetComments(i.issue.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments err: %v", err)
	}

	//log.Printf("Comments: %s", i.issue.Notes)
	//fmt.Printf("Comments: %d\n", len(comments))
	//fmt.Printf("%s\n", strings.Join(comments, "\n"))

	return comments, nil
}

func (i *WorkOnIssue) AddComment(text string) error {
	err := i.model.Comment(i.issue, text)
	if err != nil {
		return fmt.Errorf("failed to comment issue err: %v", err)
	}
	return nil
}
