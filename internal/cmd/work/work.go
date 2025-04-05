package work

import (
	"fmt"
	"log"
	"time"

	"github.com/andrejsstepanovs/andai/internal"
	"github.com/andrejsstepanovs/andai/internal/employee"
	"github.com/andrejsstepanovs/andai/internal/employee/actions"
	"github.com/andrejsstepanovs/andai/internal/exec"
	"github.com/andrejsstepanovs/andai/internal/settings"
	"github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
)

func newLoopCommand(deps *internal.AppDependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "loop",
		Short: "Work forever",
		RunE: func(_ *cobra.Command, _ []string) error {
			return Loop(deps)
		},
	}
}

func Loop(deps *internal.AppDependencies) error {
	// AI: Initialize tracking variables for incremental sleep
	lastSuccessfulTask := time.Now()
	consecutiveEmptyChecks := 0
	currentSleepDuration := time.Duration(0) // Start with no sleep

	for {
		params, err := deps.Config.Load()
		if err != nil {
			return err
		}

		err = processTriggers(deps.Model, params.Workflow)
		if err != nil {
			return fmt.Errorf("failed to process triggers err: %v", err)
		}
		wasWorking, err := workNext(deps, params)
		if err != nil {
			return fmt.Errorf("failed to work next err: %v", err)
		}

		if wasWorking {
			lastSuccessfulTask = time.Now()
			consecutiveEmptyChecks = 0
			if currentSleepDuration > 0 {
				log.Printf("Resetting sleep duration to 0 after completing work")
			}
			currentSleepDuration = time.Duration(0) // Reset sleep duration after work
		} else {
			consecutiveEmptyChecks++
			timeSinceLastSuccess := time.Since(lastSuccessfulTask)

			switch {
			case consecutiveEmptyChecks < 2:
				if currentSleepDuration > 0 {
					log.Printf("First empty check - resetting sleep duration to 0")
				}
				currentSleepDuration = 0 // No sleep for first empty check
			case timeSinceLastSuccess > 5*time.Minute:
				newDuration := 30 * time.Second
				if currentSleepDuration != newDuration {
					log.Printf("Increasing sleep to %v (idle > 5min, %d consecutive empty checks)",
						newDuration, consecutiveEmptyChecks)
				}
				currentSleepDuration = newDuration
			case timeSinceLastSuccess > 1*time.Minute:
				newDuration := 15 * time.Second
				if currentSleepDuration != newDuration {
					log.Printf("Increasing sleep to %v (idle > 1min, %d consecutive empty checks)",
						newDuration, consecutiveEmptyChecks)
				}
				currentSleepDuration = newDuration
			default:
				newDuration := 5 * time.Second
				if currentSleepDuration != newDuration {
					log.Printf("Setting default sleep to %v (%d consecutive empty checks)",
						newDuration, consecutiveEmptyChecks)
				}
				currentSleepDuration = newDuration
			}

			if currentSleepDuration > 0 {
				log.Printf("No workable issues found. Sleeping for %v (idle for %v, %d consecutive empty checks)",
					currentSleepDuration,
					time.Since(lastSuccessfulTask).Round(time.Second),
					consecutiveEmptyChecks)
				time.Sleep(currentSleepDuration)
				continue // Skip the rest of the loop after sleeping
			}
		}
	}
}

func newNextCommand(deps *internal.AppDependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "next",
		Short: "Work with redmine",
		RunE: func(_ *cobra.Command, _ []string) error {
			params, err := deps.Config.Load()
			if err != nil {
				return err
			}

			log.Println("Searching for workable issues")
			_, err = workNext(deps, params)
			return err
		},
	}
}

func getFirstWorkableIssuePerProjects(issues []redmine.Issue) []redmine.Issue {
	issuesPerProject := make(map[int][]redmine.Issue)
	for _, issue := range issues {
		if len(issuesPerProject[issue.Project.Id]) == 0 {
			issuesPerProject[issue.Project.Id] = append(issuesPerProject[issue.Project.Id], issue)
		}
	}
	workableProjectIssues := make([]redmine.Issue, 0)
	for _, projectIssues := range issuesPerProject {
		workableProjectIssues = append(workableProjectIssues, projectIssues...)
	}
	return workableProjectIssues
}

func workNext(deps *internal.AppDependencies, params *settings.Settings) (bool, error) {
	issues, err := deps.Model.APIGetWorkableIssues(params.Workflow)
	if err != nil {
		log.Println("Failed to get workable issue")
		return false, err
	}

	for _, issue := range getFirstWorkableIssuePerProjects(issues) {
		// check if AI is allowed to work on it
		currentIssueState := params.Workflow.States.Get(settings.StateName(issue.Status.Name))
		currentIssueType := params.Workflow.IssueTypes.Get(settings.IssueTypeName(issue.Tracker.Name))

		if !currentIssueState.UseAI.Yes(currentIssueType.Name) {
			f := "Waiting on USER to finish work on %q (ID: %d) in %q - %q\n"
			log.Printf(f, currentIssueType.Name, issue.Id, currentIssueState.Name, issue.Subject)
			continue
		}

		fmt.Printf("WORKING ON: %q in %q ID=%d: %s\n",
			params.Workflow.IssueTypes.Get(settings.IssueTypeName(issue.Tracker.Name)).Name,
			params.Workflow.States.Get(settings.StateName(issue.Status.Name)).Name,
			issue.Id,
			issue.Subject,
		)

		parent, err := deps.Model.APIGetParent(issue)
		if err != nil {
			return false, fmt.Errorf("failed to get redmine parent issue err: %v", err)
		}

		parents, err := deps.Model.APIGetAllParents(issue)
		if err != nil {
			return false, fmt.Errorf("failed to get redmine all parent issues err: %v", err)
		}

		closedChildrenIDs, err := deps.Model.DBGetClosedChildrenIDs(issue.Id)
		if err != nil {
			return false, fmt.Errorf("failed to get redmine closed children ids err: %v", err)
		}

		closedChildren := make([]redmine.Issue, 0, len(closedChildrenIDs))
		for _, childID := range closedChildrenIDs {
			childIssue, err := deps.Model.API().Issue(childID)
			if err != nil {
				log.Printf("WARN: Failed to get details for closed child issue %d: %v", childID, err)
				continue
			}
			closedChildren = append(closedChildren, *childIssue)
		}

		openChildren, err := deps.Model.APIGetChildren(issue)
		if err != nil {
			return false, fmt.Errorf("failed to get redmine open children err: %v", err)
		}

		allChildren := make([]redmine.Issue, 0, len(openChildren)+len(closedChildren))
		allChildren = append(allChildren, openChildren...)
		allChildren = append(allChildren, closedChildren...)

		siblings, err := deps.Model.APIGetIssueSiblings(issue)
		if err != nil {
			return false, fmt.Errorf("failed to get redmine issue siblings err: %v", err)
		}

		log.Printf("Issue %d: %s", issue.Id, issue.Subject)
		project, err := deps.Model.API().Project(issue.Project.Id)
		if err != nil {
			return false, fmt.Errorf("failed to get redmine project err: %v", err)
		}
		log.Printf("Project %d: %s", project.Id, project.Name)

		projectRepo, err := deps.Model.DBGetRepository(*project)
		if err != nil {
			return false, fmt.Errorf("failed to get redmine repository err: %v", err)
		}
		log.Printf("Repository %d: %s", projectRepo.ID, projectRepo.RootURL)

		projectConfig := params.Projects.Find(project.Identifier)
		git, err := exec.FindProjectGit(projectConfig, projectRepo)
		if err != nil {
			return false, fmt.Errorf("failed to find project git err: %v", err)
		}
		log.Printf("Project Repository Opened %s", git.GetPath())

		wb := &exec.Workbench{
			Git:   git,
			Issue: issue,
		}

		work := employee.NewRoutine(
			deps.Model,
			deps.LlmNorm,
			issue,
			parent,
			parents,
			closedChildrenIDs,
			allChildren,
			siblings,
			*project,
			projectConfig,
			wb,
			params.Aider,
			params.Workflow.States.Get(settings.StateName(issue.Status.Name)),
			params.Workflow.IssueTypes.Get(settings.IssueTypeName(issue.Tracker.Name)),
			params.Workflow.IssueTypes,
			projectRepo,
		)
		success, err := work.ExecuteWorkflow()
		if err != nil {
			return false, fmt.Errorf("failed to finish work on issue err: %v", err)
		}

		err = actions.TransitionToNextStatus(params.Workflow, deps.Model, issue, success)
		if err != nil {
			return false, fmt.Errorf("failed to comment issue err: %v", err)
		}

		// true if issue was successfully worked on
		return true, nil // nolint:staticcheck
	}

	return false, nil
}
