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
	siblings   []redmine.Issue
	project    redmine.Project
	projectCfg models.Project
	workbench  *workbench.Workbench
	state      models.State
	issueType  models.IssueType
	issueTypes models.IssueTypes
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
	siblings []redmine.Issue,
	project redmine.Project,
	projectCfg models.Project,
	workbench *workbench.Workbench,
	state models.State,
	issueType models.IssueType,
	issueTypes models.IssueTypes,
) *Employee {
	return &Employee{
		model:      model,
		llm:        llm,
		issue:      issue,
		parent:     parent,
		parents:    parents,
		children:   children,
		siblings:   siblings,
		project:    project,
		projectCfg: projectCfg,
		workbench:  workbench,
		state:      state,
		issueType:  issueType,
		issueTypes: issueTypes,
		job:        issueType.Jobs.Get(models.StateName(issue.Status.Name)),
	}
}

func (i *Employee) Work() (bool, error) {
	log.Printf("Working on %q %q (ID: %d)", i.state.Name, i.issueType.Name, i.issue.Id)

	err := i.workbench.PrepareWorkplace()
	if err != nil {
		log.Printf("Failed to prepare workplace: %v", err)
		return false, err
	}

	fmt.Printf("Total steps: %d\n", len(i.job.Steps))
	for j, step := range i.job.Steps {
		fmt.Printf("Step: %d\n", j+1)
		step.Prompt = i.addHistory(step)
		output, err := i.processStep(step)
		if err != nil {
			log.Printf("Failed to action step: %v", err)
			return false, err
		}
		i.CommentOutput(step, output)
		if step.Remember {
			i.history = append(i.history, output)
		}
		fmt.Println("Success")
	}

	return true, nil
}

func (i *Employee) processStep(step models.Step) (exec.Output, error) {
	log.Printf("%s - %s", step.Command, step.Action)

	comments, err := i.getComments()
	if err != nil {
		log.Printf("Failed to get comments: %v", err)
		return exec.Output{}, err
	}

	knowledge := utils.Knowledge{
		Issue:      i.issue,
		Parent:     i.parent,
		Parents:    i.parents,
		Children:   i.children,
		Siblings:   i.siblings,
		Project:    i.projectCfg,
		IssueTypes: i.issueTypes,
		Comments:   comments,
		Step:       step,
	}

	contextFile, err := knowledge.BuildIssueKnowledgeTmpFile()
	if err != nil {
		log.Printf("Failed to build issue context tmp file: %v", err)
	}

	if contextFile != "" {
		log.Printf("Context file: %q\n", contextFile)
		//defer os.Remove(contextFile)

		contents, err := utils.GetFileContents(contextFile)
		if err != nil {
			log.Printf("Failed to get file contents: %v", err)
		}
		log.Printf("Context file contents: \n------------------\n%s\n------------------\n", contents)
	}

	switch step.Command {
	case "next":
		return exec.Output{
			Command: "next",
			Stdout:  "Success, moving to next",
		}, nil
	case "git":
		return exec.Exec(step.Command, step.Action)
	case "create-issues":
		trackerID, err := i.model.DBGetTrackersByName(step.Action)
		if err != nil {
			log.Printf("Failed to get tracker by name: %v", err)
			return exec.Output{}, err
		}
		fmt.Printf("Need to create: %q Tracker ID: %d", step.Action, trackerID)

		out, issues, deps, err := processor.BobikCreateIssue(models.IssueTypeName(step.Action), contextFile)
		if err != nil {
			return out, err
		}

		err = i.model.CreateChildIssuesWithDependencies(trackerID, i.issue, issues, deps)
		if err != nil {
			log.Printf("Failed to create new issues: %v", err)
			return exec.Output{}, err
		}
		return exec.Output{}, nil
	case "merge-into-parent":
		currentBranchName := i.workbench.GetIssueBranchName(i.issue)
		parentBranchName := i.workbench.GetIssueBranchName(*i.parent)
		log.Printf("Merging current branch: %q into parent branch: %q", currentBranchName, parentBranchName)

		err = i.workbench.Git.CheckoutBranch(parentBranchName)
		if err != nil {
			log.Printf("Failed to checkout parent branch: %v", err)
			return exec.Output{}, err
		}
		log.Printf("Checked out parent branch: %q", parentBranchName)
		log.Printf("Merging...")

		out, err := exec.Exec("git", "merge", currentBranchName)
		if err != nil {
			log.Printf("Failed to merge current branch: %q into parent branch: %q: %v", currentBranchName, parentBranchName, err)
			return out, err
		}
		log.Printf("Merged current branch: %q into parent branch: %q", currentBranchName, parentBranchName)

		txt := fmt.Sprintf("Merged #%d branch %q into %q", i.issue.Id, currentBranchName, parentBranchName)
		err = i.AddCommentToParent(txt)

		return out, err
	case "bobik":
		promptFile, err := knowledge.BuildPromptTmpFile()
		if err != nil {
			log.Printf("Failed to build prompt tmp file: %v", err)
		}
		return processor.BobikExecute(promptFile, step)
	case "aider":
	case "aid":
		switch step.Action {
		case "architect":
			return processor.AiderExecute(contextFile, step)
		case "code":
			out, err := processor.AiderExecute(contextFile, step)
			if err != nil {
				return out, err
			}
			sha, getShaErr := i.workbench.GetLastCommit()
			if getShaErr != nil {
				log.Printf("Failed to get last commit sha: %v", getShaErr)
				return out, nil
			}
			if sha != "" {
				log.Printf("Last commit sha: %q", sha)
				txt := make([]string, 0)

				branchName := i.workbench.GetIssueBranchName(i.issue)
				format := "- committed changes to branch %s [%s](/projects/lco/repository/%s/revisions/%s/diff)"
				txt = append(txt, fmt.Sprintf(format, branchName, sha, i.project.Identifier, sha))

				format = "- branch [%s](/projects/%s/repository/lco?rev=%s)"
				txt = append(txt, fmt.Sprintf(format, branchName, i.project.Identifier, branchName))

				err = i.AddComment(strings.Join(txt, "\n"))
				if err != nil {
					return out, err
				}
			}
			return out, nil
		default:
			return exec.Output{}, fmt.Errorf("unknown %q action: %q", step.Command, step.Action)
		}
	default:
		return exec.Output{}, fmt.Errorf("unknown step command: %q", step.Command)
	}

	return exec.Output{}, nil
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
