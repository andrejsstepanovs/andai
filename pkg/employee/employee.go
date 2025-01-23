package employee

import (
	"fmt"
	"log"
	"strings"

	"github.com/andrejsstepanovs/andai/pkg/employee/processor"
	"github.com/andrejsstepanovs/andai/pkg/employee/utils"
	"github.com/andrejsstepanovs/andai/pkg/exec"
	"github.com/andrejsstepanovs/andai/pkg/llm"
	"github.com/andrejsstepanovs/andai/pkg/models"
	model "github.com/andrejsstepanovs/andai/pkg/redmine"
	"github.com/andrejsstepanovs/andai/pkg/workbench"
	"github.com/mattn/go-redmine"
)

type Employee struct {
	model      *model.Model
	llm        *llm.LLM
	issue      redmine.Issue
	parent     *redmine.Issue
	parents    []redmine.Issue
	children   []redmine.Issue
	project    redmine.Project
	projectCfg models.Project
	workbench  *workbench.Workbench
	state      models.State
	issueType  models.IssueType
	job        models.Job
	history    []exec.Output
}

func NewWorkOnIssue(
	model *model.Model,
	llm *llm.LLM,
	issue redmine.Issue,
	parent *redmine.Issue,
	parents []redmine.Issue,
	children []redmine.Issue,
	project redmine.Project,
	projectCfg models.Project,
	workbench *workbench.Workbench,
	state models.State,
	issueType models.IssueType,
) *Employee {
	return &Employee{
		model:      model,
		llm:        llm,
		issue:      issue,
		parent:     parent,
		parents:    parents,
		children:   children,
		project:    project,
		projectCfg: projectCfg,
		workbench:  workbench,
		state:      state,
		issueType:  issueType,
		job:        issueType.Jobs.Get(models.StateName(issue.Status.Name)),
	}
}

func (i *Employee) Work() bool {
	log.Printf("Working on %q %q (ID: %d)", i.state.Name, i.issueType.Name, i.issue.Id)

	err := i.workbench.PrepareWorkplace()
	if err != nil {
		log.Printf("Failed to prepare workplace: %v", err)
		return false
	}

	fmt.Printf("Total steps: %d\n", len(i.job.Steps))
	for j, step := range i.job.Steps {
		fmt.Printf("Step: %d\n", j+1)
		step.Prompt = i.addHistory(step)
		output, err := i.processStep(step)
		if err != nil {
			log.Printf("Failed to action step: %v", err)
			return false
		}
		i.CommentOutput(step, output)
		if step.Remember {
			i.history = append(i.history, output)
		}
		fmt.Println("Success")
	}

	return true
}

func (i *Employee) processStep(step models.Step) (exec.Output, error) {
	fmt.Println(step.Command, step.Action)

	comments, err := i.getComments()
	if err != nil {
		log.Printf("Failed to get comments: %v", err)
		return exec.Output{}, err
	}

	contextFile, err := utils.BuildIssueTmpFile(
		i.issue,
		i.parent,
		i.parents,
		i.children,
		i.projectCfg,
		comments,
		step,
	)
	if err != nil {
		log.Printf("Failed to build issue context tmp file: %v", err)
	}

	switch step.Command {
	case "git":
		return exec.Exec(step.Command, step.Action)
	case "aider", "aid":
		switch step.Action {
		case "architect", "code":
			return processor.AiderExecute(contextFile, step)
		default:
			return exec.Output{}, fmt.Errorf("unknown %q action: %q", step.Command, step.Action)
		}
	case "create-issues":
		return processor.BobikCreateIssue(i.issue.Id, models.IssueTypeName(step.Action), contextFile)
	case "bobik":
		promptFile, err := utils.BuildPromptTmpFile(i.issue, step)
		if err != nil {
			log.Printf("Failed to build prompt tmp file: %v", err)
		}
		return processor.BobikExecute(promptFile, step)
	default:
		return exec.Output{}, fmt.Errorf("unknown step command: %q", step.Command)
	}
}

func (i *Employee) addHistory(step models.Step) models.StepPrompt {
	if len(i.history) == 0 {
		return step.Prompt
	}
	hydratedPrompt := make([]string, 0)
	for _, o := range i.history {
		hydratedPrompt = append(hydratedPrompt, o.AsPrompt())
	}
	hydratedPrompt = append(hydratedPrompt, string(step.Prompt))
	return models.StepPrompt(strings.Join(hydratedPrompt, "\n\n----\n\n"))
}
