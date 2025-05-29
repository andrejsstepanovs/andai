package employee

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/andrejsstepanovs/andai/internal/ai"
	"github.com/andrejsstepanovs/andai/internal/employee/actions"
	"github.com/andrejsstepanovs/andai/internal/employee/actions/file"
	"github.com/andrejsstepanovs/andai/internal/employee/actions/models"
	"github.com/andrejsstepanovs/andai/internal/employee/knowledge"
	"github.com/andrejsstepanovs/andai/internal/exec"
	model "github.com/andrejsstepanovs/andai/internal/redmine"
	"github.com/andrejsstepanovs/andai/internal/settings"
)

// ExecuteWorkflow processes all workflow steps defined for the current issue state.
// It prepares the workplace, executes each step sequentially, and maintains execution history.
// Steps are executed in order and their outputs may be preserved in history if marked as "remember".
//
// Returns:
//   - bool: true if all steps completed successfully
//   - error: any error encountered during execution
func (i *Routine) ExecuteWorkflow() (bool, error) {
	log.Printf("Working on %q issue (%d) in state: %q", i.issueType.Name, i.issue.Id, i.state.Name)

	needSetup := true
	if len(i.job.Steps) == 1 && i.job.Steps[0].Command == "next" {
		needSetup = false
	}

	if needSetup {
		parentBranches := i.getTargetBranch()
		err := i.workbench.PrepareWorkplace(parentBranches...)
		if err != nil {
			log.Printf("Failed to prepare workplace: %v", err)
			return false, err
		}
	}

	for stepIndex, step := range i.job.Steps {
		log.Printf("Step: %d / %d", stepIndex+1, len(i.job.Steps))

		step.History = i.history
		if len(i.contextFiles) > 0 {
			step.ContextFiles = i.contextFiles
		}

		executionOutput, err := i.executeWorkflowStep(step)
		if err != nil {
			if errors.Is(err, ErrNegativeOutcome) {
				log.Printf("Negative outcome, skipping remaining steps and moving issue to negative path state.")
				return false, nil
			}
			log.Printf("Failed to action step: %v", err)
			return false, err
		}
		i.RememberOutput(step, executionOutput)

		log.Println("Success")
	}

	if needSetup {
		err := i.model.APISyncRepo(i.project)
		if err != nil {
			log.Printf("Failed to sync repo in redmine: %v", err)
		}
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
func (i *Routine) executeWorkflowStep(workflowStep settings.Step) (exec.Output, error) {
	log.Println(workflowStep.String("Execute Step"))

	comments, err := i.getComments()
	if err != nil {
		log.Printf("Failed to get comments: %v", err)
		return exec.Output{}, err
	}

	parentComments, err := i.getParentComments()
	if err != nil {
		log.Printf("Failed to get parent comments: %v", err)
		return exec.Output{}, err
	}

	understanding := knowledge.Knowledge{
		Issue:             i.issue,
		Parent:            i.parent,
		Parents:           i.parents,
		ClosedChildrenIDs: i.closedChildrenIDs,
		Children:          i.children,
		Siblings:          i.siblings,
		Workbench:         i.workbench,
		Project:           i.projectCfg,
		IssueTypes:        i.issueTypes,
		Comments:          comments,
		ParentComments:    parentComments,
		Step:              workflowStep,
	}

	contextFile, err := understanding.BuildIssueKnowledgeTmpFile()
	if err != nil {
		log.Printf("Failed to build issue context tmp file: %v", err)
	}

	removeContextFile := true
	if contextFile != "" {
		//log.Printf("Context file: %q\n", contextFile)
		defer func(name string) {
			if removeContextFile {
				err := os.Remove(name)
				if err != nil {
					log.Printf("Failed to remove context file: %v", err)
				}
			} else {
				log.Printf("Not removing context file: %q, it may be useful for debugging", name)
			}
		}(contextFile)

		_, err := file.GetContents(contextFile)
		if err != nil {
			log.Printf("Failed to get file contents: %v", err)
		}
		//log.Printf("Context file contents: \n------------------\n%s\n------------------\n", contents)
	}

	out, err := i.executeCommand(workflowStep, contextFile)
	if err != nil {
		removeContextFile = false // Don't remove context file if we have an error, it may be useful for debugging
	}
	return out, err
}

func (i *Routine) executeCommand(workflowStep settings.Step, contextFile string) (exec.Output, error) {
	type commandHandler func(settings.Step, string) (exec.Output, error)

	handlers := map[string]commandHandler{
		"next": func(_ settings.Step, _ string) (exec.Output, error) {
			return exec.Output{
				Command: "next",
				Stdout:  "Success, moving to next",
			}, nil
		},
		"git": func(step settings.Step, _ string) (exec.Output, error) {
			return exec.Exec(step.Command, time.Minute, step.Action)
		},
		"create-issues": func(step settings.Step, contextFile string) (exec.Output, error) {
			return i.createIssueCommand(step, contextFile)
		},
		"evaluate": func(_ settings.Step, contextFile string) (exec.Output, error) {
			m := i.llmPool.ForCommand(settings.LlmModelNormal, "evaluate")
			llmModel, err := ai.NewAI(m)
			if err != nil {
				return exec.Output{}, err
			}

			resp, success, err := actions.EvaluateOutcome(llmModel, contextFile)
			if err != nil {
				log.Printf("Failed to create new issues: %v", err)
				return exec.Output{}, err
			}
			log.Printf("AI evaluation response %q; result is: %t\n", resp.Stdout, success)
			if success {
				return exec.Output{Stdout: "Positive outcome"}, nil
			}
			return exec.Output{Stdout: "Negative"}, ErrNegativeOutcome
		},
		"merge-into-parent": func(_ settings.Step, _ string) (exec.Output, error) {
			return i.mergeIntoParent(i.projectCfg.DeleteBranchAfterMerge)
		},
		"bash": func(step settings.Step, _ string) (exec.Output, error) {
			return i.runBash(step)
		},
		"summarize-task": func(step settings.Step, contextFile string) (exec.Output, error) {
			return i.summarizeTheTask(step, contextFile)
		},
		"project-cmd": func(step settings.Step, _ string) (exec.Output, error) {
			return i.runProjectCmd(step)
		},
		"commit": func(step settings.Step, _ string) (exec.Output, error) {
			return i.commitUncommitted(string(step.Prompt))
		},
		"context-commits": func(_ settings.Step, contextFile string) (exec.Output, error) {
			commits, err := i.findMentionedCommits(contextFile)
			if err != nil {
				log.Printf("Failed to find mentioned commits: %v", err)
				return exec.Output{}, err
			}
			if len(commits) == 0 {
				log.Println("No commits found in context file")
				return exec.Output{Stdout: "No recent commits found"}, nil
			}
			return i.findCommitPatches(commits)
		},
		"context-files": func(_ settings.Step, contextFile string) (exec.Output, error) {
			return i.findMentionedFiles(contextFile)
		},
		"ai": func(_ settings.Step, contextFile string) (exec.Output, error) {
			return i.simpleAI(contextFile)
		},
		"aider": func(step settings.Step, contextFile string) (exec.Output, error) {
			return i.aider(step, contextFile)
		},
	}

	handler, ok := handlers[workflowStep.Command]
	if !ok {
		return exec.Output{}, fmt.Errorf("unknown step command: %q", workflowStep.Command)
	}
	return handler(workflowStep, contextFile)
}

func (i *Routine) findMentionedFiles(contextFile string) (exec.Output, error) {
	content, err := file.GetContents(contextFile)
	if err != nil {
		log.Printf("Failed to get file contents: %v", err)
		return exec.Output{}, err
	}

	allPossiblePaths, err := exec.GetAllPossiblePaths(i.projectCfg, i.projectRepo, false)
	if err != nil {
		log.Printf("Failed to get all possible paths: %v", err)
		return exec.Output{}, err
	}

	foundFiles, err := file.NewFileFinder(allPossiblePaths).FindFilesInText(content)
	if err != nil {
		log.Printf("Failed to find files: %v", err)
		return exec.Output{}, err
	}

	// TODO ignore go.mod and go.sum - use config
	i.contextFiles = foundFiles.GetAbsolutePaths()

	return exec.Output{Stdout: foundFiles.String()}, nil
}

func (i *Routine) findMentionedCommits(contextFile string) ([]string, error) {
	content, err := file.GetContents(contextFile)
	if err != nil {
		log.Printf("Failed to get file contents: %v", err)
		return nil, err
	}

	return knowledge.GitSHA1(content), nil
}

func (i *Routine) findCommitPatches(commits []string) (exec.Output, error) {
	patches := make([]string, 0)
	for n, commit := range commits {
		execOut, err := exec.Exec("git", time.Minute, "format-patch", "-1", "--stdout", "--no-binary", "--no-stat", commit)
		if err != nil {
			log.Printf("Failed to get commit patch for %q: %v", commit, err)
			continue
		}
		patches = append(patches, fmt.Sprintf("Commit %d: ```patch\n%s\n```", n, execOut.Stdout))
	}

	msg := fmt.Sprintf("# Recent project git commits:\n%s", strings.Join(patches, "\n\n"))

	return exec.Output{Stdout: msg}, nil
}

func (i *Routine) runProjectCmd(workflowStep settings.Step) (exec.Output, error) {
	command, err := i.projectCfg.Commands.Find(workflowStep.Action)
	if err != nil {
		return exec.Output{}, err
	}
	parts := command.Command
	if len(parts) == 0 {
		return exec.Output{}, fmt.Errorf("no actual commands provided for %q project %q command", workflowStep.Action, i.projectCfg.Identifier)
	}

	cmd := parts[0]
	arguments := make([]string, 0)
	if len(parts) > 1 {
		arguments = parts[1:]
	}

	ret, err := exec.Exec(cmd, time.Minute*30, arguments...)

	if err != nil {
		hardErr := i.checkCommandHardFailure(err, parts)
		if hardErr != nil {
			return ret, hardErr
		}

		if command.IgnoreError {
			log.Printf("Ignoring error: %v\n", err)
		} else {
			return ret, err
		}
	}

	if command.SuccessIfNoOutput && ret.Stderr == "" && ret.Stdout == "" {
		ret.Stdout = "Success"
	}
	if command.IgnoreStdOutIfNoStdErr && ret.Stderr == "" {
		ret.Stdout = "OK"
	}

	return ret, nil
}

func (i *Routine) checkCommandHardFailure(err error, command []string) error {
	failedIf := [][]string{
		{"make:", "No rule to make target"},
		{"make:", " command not found:"},
	}
	for _, fail := range failedIf {
		realFail := false
		for _, f := range fail {
			if strings.Contains(err.Error(), f) {
				realFail = true
			}
		}
		if realFail {
			return fmt.Errorf("project command %q failed err: %v", strings.Join(command, " "), err)
		}
	}

	return nil
}

func (i *Routine) runBash(workflowStep settings.Step) (exec.Output, error) {
	parts := strings.Split(workflowStep.Action, " ")
	cmd := parts[0]
	arguments := make([]string, 0)
	if len(parts) > 1 {
		arguments = parts[1:]
	}
	ret, err := exec.Exec(cmd, time.Minute*30, arguments...)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func (i *Routine) aider(workflowStep settings.Step, contextFile string) (exec.Output, error) {
	if contextFile == "" {
		return exec.Output{}, fmt.Errorf("no context file provided for aider command")
	}
	switch workflowStep.Action {
	case "commit":
		return i.aiderCommit(workflowStep)
	case "architect":
		return i.aiderArchitect(workflowStep, contextFile)
	case "code", "architect-code":
		return i.aiderCode(workflowStep, contextFile)
	default:
		return exec.Output{}, fmt.Errorf("unknown %q action: %q", workflowStep.Command, workflowStep.Action)
	}
}

func (i *Routine) summarizeTask(workflowStep settings.Step, contextFile string, includeFiles []string) (string, error) {
	contextContent, err := file.GetContents(contextFile)
	if err != nil {
		log.Printf("Failed to get context file contents: %v", err)
		return "", fmt.Errorf("failed to get context file contents: %w", err)
	}

	m := i.llmPool.ForCommand(settings.LlmModelNormal, "summarize-task")
	llmModel, err := ai.NewAI(m)
	if err != nil {
		return "", err
	}

	query := "Perfect, yes. Now do it! Answer only with the reformatted task text."
	remainingFiles := includeFiles
	var ret exec.Output
	var lastError error

	for attempts := 0; attempts < len(remainingFiles); attempts++ {
		history, err := i.buildTaskSummaryAIHistory(contextContent, query, remainingFiles)
		if err != nil {
			log.Printf("Failed to build task summary history: %v", err)
			return "", fmt.Errorf("failed to build task summary history: %w", err)
		}

		ret, err = llmModel.Multi(query, history)
		if err == nil {
			log.Printf("AI response received successfully")
			break // Success, exit the loop
		}

		lastError = err
		log.Printf("AI request attempt %d failed: %v", attempts+1, err)

		// If we hit token limits and have files we can remove, try again with fewer files
		if errors.Is(err, ai.ErrTooManyTokens) && len(remainingFiles) > 0 {
			log.Printf("Too many tokens, reducing included files by one. Consider setting or increasing 'max_tokens' parameter.")
			remainingFiles = remainingFiles[1:]
			continue
		}

		// Any other error or we can't reduce files further
		break
	}

	// If we exhausted all attempts without success
	if lastError != nil {
		log.Printf("All AI requests failed, last error: %v", lastError)
		return "", fmt.Errorf("failed to get AI response: %w", lastError)
	}

	//log.Printf("AI response: %s", ret.Stdout)

	if workflowStep.CommentSummary {
		if err := i.AddComment(ret.Stdout); err != nil {
			log.Printf("Failed to add comment: %v", err)
			return "", fmt.Errorf("failed to add comment: %w", err)
		}
	}

	return file.BuildPromptTextTmpFile(ret.Stdout)
}

func (i *Routine) buildTaskSummaryAIHistory(contextContent, query string, includeFiles []string) ([]map[string]string, error) {
	var summaryPrompt string
	if i.codingAgents.Aider.TaskSummaryPrompt != "" {
		summaryPrompt = i.codingAgents.Aider.TaskSummaryPrompt
	} else {
		summaryPrompt = ai.TaskSummaryPrompt
	}

	history := []map[string]string{
		{"USER": "Help me to reformat this task"},
		{"AI": "OK! Please provide the task contents and all relevant details."},
		{"USER": contextContent},
		{"AI": "Got it! This is a original task text that I should summarize, groom and reformat. How should I reformat this task?"},
		{"USER": summaryPrompt},
		{"AI": "Understood! I will reformat the task using the provided instructions. Is this all?"},
		{"USER": "Yes, remember that you are a senior software engineer with strong system design experience. Avoid situations where we end up with spahghetti code to maintain."},
		{"AI": "Of course! I know how to prepare clean, maintainable tasks. My designs focus on modular structure, clear separation of concerns, and patterns that ensure long-term sustainability of the codebase. Anything else?"},
	}

	if len(includeFiles) > 0 {
		filesContent := make([]string, 0)
		for _, fileName := range includeFiles {
			content, err := file.GetContents(fileName)
			if err != nil {
				return history, err
			}
			filesContent = append(filesContent, fmt.Sprintf("# %s\n%s", fileName, content))
		}

		history = append(history, map[string]string{"USER": "You will probably also need existing code files?"})
		history = append(history, map[string]string{"AI": "Yes!"})
		history = append(history, map[string]string{"USER": strings.Join(filesContent, "\n")})
		history = append(history, map[string]string{"AI": "Thanks! I will use this to understand the task better and improve the task description."})
	}

	history = append(history, map[string]string{"USER": "Yes, but remember, this is super important: do not work on task content, only reformat it."})
	history = append(history, map[string]string{"AI": "Got it! I will reformat the task text and not do what the task is asking because someone else will actually work on it. I just prepare a task description."})
	history = append(history, map[string]string{"USER": query})

	return history, nil
}

func (i *Routine) summarizeTheTask(workflowStep settings.Step, contextFile string) (exec.Output, error) {
	summaryFile, err := i.summarizeTask(workflowStep, contextFile, i.contextFiles)
	if err != nil {
		return exec.Output{}, err
	}

	content, err := file.GetContents(summaryFile)
	if err != nil {
		return exec.Output{}, err
	}

	return exec.Output{Stdout: content}, err
}

func (i *Routine) aiderCode(workflowStep settings.Step, contextFile string) (exec.Output, error) {
	if workflowStep.Summarize {
		var err error
		contextFile, err = i.summarizeTask(workflowStep, contextFile, []string{})
		if err != nil {
			return exec.Output{}, err
		}
	}

	currentCommitSku, err := i.workbench.GetLastCommit()
	if err != nil {
		return exec.Output{}, err
	}

	out, err := actions.AiderExecute(contextFile, workflowStep, i.codingAgents.Aider, true)
	if err != nil {
		return out, err
	}

	commitCount, err := i.commentCommitsSince(currentCommitSku, "code changes")
	if err != nil {
		return out, err
	}

	if commitCount == 0 {
		msg := fmt.Sprintf("No new commits found. Check tool output:\nstdout: %s \n\nstderr:%s", out.Stdout, out.Stderr)
		errComment := i.AddComment(msg)
		if errComment != nil {
			return out, errComment
		}

		prompt := "Evaluate this stdout and stderr. Developer did not made any changes in the code. " +
			"You need to evaluate if this is a positive outcome or not. " +
			"If it is positive, answer with single word: 'Positive' or 'Negative'. " +
			"'Positive' means that the task is already OK and there indeed is nothing to do, i.e it was intentional. " +
			"'Negative' is everything else."

		txt := fmt.Sprintf("stdout:\n%s\n\nstderr:\n%s\n\n# Your task:\n%s", out.Stdout, out.Stderr, prompt)
		promptFile, err := file.BuildPromptTextTmpFile(txt)
		if err != nil {
			log.Printf("Failed to build prompt text tmp file: %v", err)
			return out, fmt.Errorf("failed to build prompt text tmp file: %w", err)
		}
		result, err := i.simpleAI(promptFile)
		if err != nil {
			return out, fmt.Errorf("failed to evaluate no comments message with AI: %w", err)
		}
		if strings.Trim(strings.TrimSpace(result.Stdout), "\"") == "Positive" {
			log.Println("AI evaluation: Positive outcome, no new commits found, task is already OK.")
			return out, nil
		}
		log.Println("AI evaluation: Negative outcome")

		// TODO check if he wants more files and find them.
		// use tree ./ -f
		// or similar

		return out, fmt.Errorf("no new commits found")
	}

	return out, nil
}

func (i *Routine) aiderArchitect(workflowStep settings.Step, contextFile string) (exec.Output, error) {
	if workflowStep.Summarize {
		var err error
		contextFile, err = i.summarizeTask(workflowStep, contextFile, []string{})
		if err != nil {
			return exec.Output{}, err
		}
	}

	architectResult, err := actions.AiderExecute(contextFile, workflowStep, i.codingAgents.Aider, true)
	if err != nil {
		return architectResult, err
	}
	// because architect is running with --yes flag he is proceeding with code changes. We clean it after the run.
	_, err = exec.Exec("git", time.Minute, "reset", "--hard")
	if err != nil {
		return architectResult, err
	}
	_, err = exec.Exec("git", time.Minute, "clean", "-fd", ".aider.tags.cache.v3")
	if err != nil {
		return architectResult, err
	}
	return architectResult, nil
}

// aiderCommit DEPRECATED. not working as expected.
func (i *Routine) aiderCommit(workflowStep settings.Step) (exec.Output, error) {
	// TODO check if there is anything uncommitted.

	lastSha, err := i.workbench.GetLastCommit()
	if err != nil {
		return exec.Output{}, err
	}

	commitResult, err := actions.AiderExecute(
		"Commit any uncommitted changes. Do nothing if no uncommitted changes are present.",
		workflowStep,
		i.codingAgents.Aider,
		true,
	)
	if err != nil {
		return commitResult, err
	}

	_, err = i.commentCommitsSince(lastSha, "code changes")
	if err != nil {
		return commitResult, fmt.Errorf("failed to get commits since %q: %v", lastSha, err)
	}

	return commitResult, nil
}

func (i *Routine) simpleAI(promptFile string) (exec.Output, error) {
	prompt, err := file.GetContents(promptFile)
	if err != nil {
		log.Printf("Failed to get file contents: %v", err)
	}

	m := i.llmPool.ForCommand(settings.LlmModelNormal, "ai")
	llmModel, err := ai.NewAI(m)
	if err != nil {
		return exec.Output{}, err
	}

	ret, err := llmModel.Simple(prompt)
	if err != nil {
		log.Printf("Failed to run AI: %v", err)
	}

	return ret, err
}

func (i *Routine) mergeIntoParent(deleteBranchAfterMerge bool) (exec.Output, error) {
	currentBranchName := i.workbench.GetIssueBranchName(i.issue)
	parentBranches := i.getTargetBranch()
	parentBranchName := parentBranches[len(parentBranches)-1] // last is our first parent

	skipMerge := i.workbench.GetIssueSkipMergeOverride(i.issue)
	if skipMerge {
		msg := fmt.Sprintf("Skipped merge (because %q = 1) of current branch: %q into parent branch: %q", model.CustomFieldSkipMerge, currentBranchName, parentBranchName)
		log.Println(msg)

		i.AddCommentToParent(msg)
		i.AddComment(msg)

		return exec.Output{Stdout: "Skipping merge"}, nil
	}

	log.Printf("Merging current branch: %q into parent branch: %q", currentBranchName, parentBranchName)

	err := i.workbench.CheckoutBranch(parentBranchName)
	if err != nil {
		log.Printf("Failed to checkout parent branch: %v", err)
		return exec.Output{}, err
	}
	log.Printf("Checked out parent branch: %q", parentBranchName)
	log.Printf("Merging...")

	out, err := exec.Exec("git", time.Minute, "merge", currentBranchName, "--no-edit")
	if err != nil {
		log.Printf("Failed to merge current branch: %q into parent branch: %q err: %v, stderr: %s", currentBranchName, parentBranchName, err, out.Stderr)
		return out, err
	}
	log.Printf("Merged current branch: %q into parent branch: %q", currentBranchName, parentBranchName)

	err = i.commentParentBranchDiff()
	if err != nil {
		return out, err
	}

	err = i.commentParentAboutMerge()
	if err != nil {
		return out, err
	}

	err = i.commentCurrentAboutMerge()
	if err != nil {
		return out, err
	}

	// delete the branch after merge
	if deleteBranchAfterMerge {
		err = i.workbench.Git.DeleteBranch(currentBranchName)
		if err != nil {
			log.Printf("Failed to delete branch: %s err: %v", currentBranchName, err)
			return out, err
		}
	}

	return exec.Output{Stdout: "Merged"}, nil
}

func (i *Routine) createIssueCommand(workflowStep settings.Step, contextFile string) (exec.Output, error) {
	if contextFile == "" {
		return exec.Output{}, fmt.Errorf("no context file provided for create-issues command")
	}
	trackerID, err := i.model.DBGetTrackersByName(workflowStep.Action)
	if err != nil {
		log.Printf("Failed to get tracker by name: %v", err)
		return exec.Output{}, err
	}
	log.Printf("Need to create: %q Tracker ID: %d\n", workflowStep.Action, trackerID)

	m := i.llmPool.ForCommand(settings.LlmModelNormal, "create-issues")
	llmModel, err := ai.NewAI(m)
	if err != nil {
		return exec.Output{}, err
	}

	executionOutput, issues, deps, err := actions.GenerateIssues(
		llmModel,
		settings.IssueTypeName(workflowStep.Action),
		contextFile,
	)
	if err != nil {
		if errors.Is(err, models.ErrNoIssues) {
			msg := "No new issues are needed here"
			log.Println(msg)
			return exec.Output{Stdout: msg}, nil
		}
		return executionOutput, err
	}

	err = i.model.CreateChildIssuesWithDependencies(trackerID, i.issue, issues, deps)
	if err != nil {
		log.Printf("Failed to create new issues: %v", err)
		return exec.Output{}, err
	}
	msg := fmt.Sprintf("Created %d Issues", len(issues))
	log.Println(msg)

	return exec.Output{Stdout: msg}, nil
}
