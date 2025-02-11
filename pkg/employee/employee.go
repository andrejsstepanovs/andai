package employee

import (
	"fmt"
	"log"
	"strings"

	"github.com/andrejsstepanovs/andai/pkg/ai"
	"github.com/andrejsstepanovs/andai/pkg/employee/processor"
	"github.com/andrejsstepanovs/andai/pkg/employee/utils"
	"github.com/andrejsstepanovs/andai/pkg/exec"
	"github.com/andrejsstepanovs/andai/pkg/models"
	model "github.com/andrejsstepanovs/andai/pkg/redmine"
	"github.com/andrejsstepanovs/andai/pkg/workbench"
	"github.com/mattn/go-redmine"
)

type Employee struct {
	model      *model.Model
	llmNorm    *ai.AI
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

// AI: NewEmployee creates an Employee instance configured to work on a specific Redmine issue.
// It initializes the employee with all necessary context including issue relationships,
// project details, and workflow configuration.
func NewEmployee(
	model *model.Model,
	llm *ai.AI,
	issue redmine.Issue,
	parentIssue *redmine.Issue,
	parentIssues []redmine.Issue,
	childIssues []redmine.Issue,
	siblingIssues []redmine.Issue,
	project redmine.Project,
	projectConfig models.Project,
	workbench *workbench.Workbench,
	state models.State,
	issueType models.IssueType,
	issueTypes models.IssueTypes,
) *Employee {
	return &Employee{
		model:      model,
		llmNorm:    llm,
		issue:      issue,
		parent:     parentIssue,
		parents:    parentIssues,
		children:   childIssues,
		siblings:   siblingIssues,
		project:    project,
		projectCfg: projectConfig,
		workbench:  workbench,
		state:      state,
		issueType:  issueType,
		issueTypes: issueTypes,
		job:        issueType.Jobs.Get(models.StateName(issue.Status.Name)),
	}
}

// ExecuteWorkflow processes all workflow steps defined for the current issue state.
// It prepares the workplace, executes each step sequentially, and maintains execution history.
// Steps are executed in order and their outputs may be preserved in history if marked as "remember".
//
// Returns:
//   - bool: true if all steps completed successfully
//   - error: any error encountered during execution
func (i *Employee) ExecuteWorkflow() (bool, error) {
	log.Printf("Working on %q %q (ID: %d)", i.state.Name, i.issueType.Name, i.issue.Id)

	var parentIssueID *int
	if i.parent != nil {
		parentIssueID = &i.parent.Id
	}
	err := i.workbench.PrepareWorkplace(parentIssueID)
	if err != nil {
		log.Printf("Failed to prepare workplace: %v", err)
		return false, err
	}

	fmt.Printf("Total steps: %d\n", len(i.job.Steps))
	for stepIndex, step := range i.job.Steps {
		fmt.Printf("Step: %d\n", stepIndex+1)
		step.Prompt = i.appendHistoryToPrompt(step)
		executionOutput, err := i.executeWorkflowStep(step)
		if err != nil {
			log.Printf("Failed to action step: %v", err)
			return false, err
		}
		i.CommentOutput(step, executionOutput)
		if step.Remember {
			i.history = append(i.history, executionOutput)
		}
		fmt.Println("Success")
	}

	return true, nil
}

// executeWorkflowStep processes a single workflow step for an employee's task.
// It handles different types of commands (next, git, create-issues, merge-into-parent, ai, aider)
// and manages the execution context and results. The method maintains state through the employee's
// history for steps marked as "remember".
//
// Parameters:
//   - workflowStep: The workflow step configuration to execute
//
// Returns:
//   - Output: The execution results including command, stdout and stderr
//   - error: Any error that occurred during execution
func (i *Employee) executeWorkflowStep(workflowStep models.Step) (exec.Output, error) {
	log.Printf("%s - %s", workflowStep.Command, workflowStep.Action)

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
		Step:       workflowStep,
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

	switch workflowStep.Command {
	case "next":
		return exec.Output{
			Command: "next",
			Stdout:  "Success, moving to next",
		}, nil
	case "git":
		return exec.Exec(workflowStep.Command, workflowStep.Action)
	case "create-issues":
		trackerID, err := i.model.DBGetTrackersByName(workflowStep.Action)
		if err != nil {
			log.Printf("Failed to get tracker by name: %v", err)
			return exec.Output{}, err
		}
		fmt.Printf("Need to create: %q Tracker ID: %d", workflowStep.Action, trackerID)

		executionOutput, issues, deps, err := processor.GenerateIssues(i.llmNorm, models.IssueTypeName(workflowStep.Action), contextFile)
		if err != nil {
			return executionOutput, err
		}

		err = i.model.CreateChildIssuesWithDependencies(trackerID, i.issue, issues, deps)
		if err != nil {
			log.Printf("Failed to create new issues: %v", err)
			return exec.Output{}, err
		}
		return exec.Output{Stdout: fmt.Sprintf("Created %d Issues", len(issues))}, nil
	case "merge-into-parent":
		currentBranchName := i.workbench.GetIssueBranchName(i.issue)
		if i.parent == nil {
			return exec.Output{}, fmt.Errorf("parent issue is not set for %d, branch: %s", i.issue.Id, currentBranchName)
		}
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

		commentText := fmt.Sprintf("Merged #%d branch %q into %q", i.issue.Id, currentBranchName, parentBranchName)
		err = i.AddCommentToParent(commentText)
		if err != nil {
			log.Printf("Failed to add comment to parent: %v", err)
			return out, err
		}

		commentText = fmt.Sprintf("Merged #%d branch %q into parent %q", i.issue.Id, currentBranchName, parentBranchName)
		err = i.AddComment(commentText)
		if err != nil {
			log.Printf("Failed to add comment: %v", err)
			return out, err
		}

		// delete the branch after merge
		err = i.workbench.Git.DeleteBranch(currentBranchName)
		if err != nil {
			log.Printf("Failed to delete branch: %s err: %v", currentBranchName, err)
			return out, err
		}

		return exec.Output{Stdout: "Merged"}, nil
	case "ai":
		promptFile, err := knowledge.BuildPromptTmpFile()
		if err != nil {
			log.Printf("Failed to build prompt tmp file: %v", err)
		}

		prompt, err := utils.GetFileContents(promptFile)
		if err != nil {
			log.Printf("Failed to get file contents: %v", err)
		}
		return i.llmNorm.Simple(prompt)
	case "aid", "aider":
		switch workflowStep.Action {
		case "commit":
			lastSha, err := i.workbench.GetLastCommit()
			if err != nil {
				return exec.Output{}, err
			}

			commitResult, err := processor.AiderExecute("Commit any uncommitted changes", workflowStep)
			if err != nil {
				return commitResult, err
			}

			commits, err := i.workbench.GetCommitsSinceInReverseOrder(lastSha)
			if err != nil {
				return commitResult, fmt.Errorf("failed to get commits since %q: %v", lastSha, err)
			}
			for _, sha := range commits {
				format := "%d. Commit [%s](/projects/%s/repository/%s/revisions/%s/diff)"
				txt := fmt.Sprintf(format, 1, sha, i.project.Identifier, i.project.Identifier, sha)
				err = i.AddComment(txt)
				if err != nil {
					return commitResult, err
				}
			}

			return commitResult, nil
		case "architect":
			architectResult, err := processor.AiderExecute(contextFile, workflowStep)
			if err != nil {
				return architectResult, err
			}
			// because architect is running with --yes flag he is proceeding with code changes. We clean it after the run.
			_, err = exec.Exec("git", "reset", "--hard")
			if err != nil {
				return architectResult, err
			}
			_, err = exec.Exec("git", "clean", "-fd", ".aider.tags.cache.v3")
			if err != nil {
				return architectResult, err
			}
			return architectResult, nil
		case "code":
			out, err := processor.AiderExecute(contextFile, workflowStep)
			if err != nil {
				return out, err
			}
			commits, getShaErr := i.workbench.GetBranchCommits(1)
			if getShaErr != nil {
				log.Printf("Failed to get last commit sha: %v", getShaErr)
				return out, nil
			}

			txt := make([]string, 0)
			branchName := i.workbench.GetIssueBranchName(i.issue)
			format := "### Branch [%s](/projects/%s/repository/%s?rev=%s)"
			txt = append(txt, fmt.Sprintf(format, branchName, i.project.Identifier, i.project.Identifier, branchName))
			if len(commits) > 0 {
				for n, sha := range commits {
					format = "%d. Commit [%s](/projects/%s/repository/%s/revisions/%s/diff)"
					txt = append(txt, fmt.Sprintf(format, n+1, sha, i.project.Identifier, i.project.Identifier, sha))
				}

				err = i.AddComment(strings.Join(txt, "\n"))
				if err != nil {
					return out, err
				}
			}
			return out, nil
		default:
			return exec.Output{}, fmt.Errorf("unknown %q action: %q", workflowStep.Command, workflowStep.Action)
		}
	default:
		return exec.Output{}, fmt.Errorf("unknown step command: %q", workflowStep.Command)
	}

	return exec.Output{}, nil
}

// appendHistoryToPrompt combines previous execution outputs with the current step prompt.
// It prepends historical execution outputs to the current step's prompt, separated by
// delimiters, creating a context-aware prompt for the next operation.
// If no history exists, returns the original prompt unchanged.
func (i *Employee) appendHistoryToPrompt(step models.Step) models.StepPrompt {
	if len(i.history) == 0 {
		return step.Prompt
	}
	promptParts := make([]string, 0)
	for _, output := range i.history {
		promptParts = append(promptParts, output.AsPrompt())
	}
	promptParts = append(promptParts, string(step.Prompt))
	return models.StepPrompt(strings.Join(promptParts, "\n\n----\n\n"))
}
