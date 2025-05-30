package work

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andrejsstepanovs/andai/internal"
	"github.com/andrejsstepanovs/andai/internal/employee"
	"github.com/andrejsstepanovs/andai/internal/employee/actions"
	"github.com/andrejsstepanovs/andai/internal/exec"
	"github.com/andrejsstepanovs/andai/internal/settings"
	"github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
)

func newLoopCommand(deps internal.DependenciesLoader) *cobra.Command {
	var project string
	cmd := &cobra.Command{
		Use:   "loop",
		Short: "Work forever. [OPTIONAL...] --project <identifier>",
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return Loop(ctx, deps, project)
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "Project identifier (optional)")
	return cmd
}

// Loop runs work next loop forever
// TODO fix this
//
//nolint:cyclop
func Loop(ctx context.Context, deps internal.DependenciesLoader, project string) error {
	lastSuccessfulTask := time.Now()
	consecutiveEmptyChecks := 0
	currentSleepDuration := time.Duration(0)

	for {
		select {
		case <-ctx.Done():
			log.Println("Received interrupt signal, exiting work loop.")
			return nil
		default:
		}

		d := deps()
		params, err := d.Config.Load()
		if err != nil {
			return err
		}

		err = processTriggers(d.Model, params.Workflow)
		if err != nil {
			return fmt.Errorf("failed to process triggers err: %v", err)
		}
		projects, err := getProjects(d, project)
		if err != nil {
			return fmt.Errorf("failed to get projects: %v", err)
		}
		log.Printf("Searching workable issues (in %d projects)", len(projects))

		wasWorking, err := workNext(d, params, projects)
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

				select {
				case <-ctx.Done():
					log.Println("Received interrupt signal during sleep, exiting work loop.")
					return nil
				case <-time.After(currentSleepDuration):
				}
				continue
			}
		}
	}
}

func newNextCommand(deps internal.DependenciesLoader) *cobra.Command {
	var project string
	cmd := &cobra.Command{
		Use:   "next",
		Short: "Works on single next available task. [OPTIONAL...] --project <identifier>",
		RunE: func(_ *cobra.Command, _ []string) error {
			d := deps()
			params, err := d.Config.Load()
			if err != nil {
				return err
			}

			projects, err := getProjects(d, project)
			if err != nil {
				return fmt.Errorf("failed to get projects: %v", err)
			}
			log.Printf("Searching workable issues (in %d projects)", len(projects))

			_, err = workNext(d, params, projects)
			return err
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "Project identifier (optional)")
	return cmd
}

func getProjects(deps *internal.AppDependencies, project string) ([]redmine.Project, error) {
	projects, err := deps.Model.GetValidProjects()
	if project != "" {
		if err != nil {
			return nil, fmt.Errorf("failed to get projects: %v", err)
		}
		var forProject *redmine.Project
		for _, p := range projects {
			if p.Identifier == project {
				forProject = &p
				break
			}
		}
		if forProject == nil {
			return nil, fmt.Errorf("project %q not found", project)
		}
		projects = []redmine.Project{*forProject}
	}

	return projects, nil
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

func workNext(deps *internal.AppDependencies, params *settings.Settings, projects []redmine.Project) (bool, error) {
	issues, err := deps.Model.APIGetWorkableIssues(params.Workflow, projects)
	if err != nil {
		log.Println("Failed to get workable issue")
		return false, err
	}

	for _, issue := range getFirstWorkableIssuePerProjects(issues) {
		// check if AI is allowed to work on it
		currentIssueState := params.Workflow.States.Get(settings.StateName(issue.Status.Name))
		currentIssueType := params.Workflow.IssueTypes.Get(settings.IssueTypeName(issue.Tracker.Name))

		if !currentIssueState.UseAI.Yes(currentIssueType.Name) {
			f := "Project: %q - Waiting on USER to finish work on %q (ID: %d) in %q - %q\n"
			log.Printf(f, issue.Project.Name, currentIssueType.Name, issue.Id, currentIssueState.Name, issue.Subject)
			continue
		}

		//log.Printf("WORKING ON: %q in %q ID=%d: %s\n",
		//	params.Workflow.IssueTypes.Get(settings.IssueTypeName(issue.Tracker.Name)).Name,
		//	params.Workflow.States.Get(settings.StateName(issue.Status.Name)).Name,
		//	issue.Id,
		//	issue.Subject,
		//)

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

		//log.Printf("Issue %d: %s", issue.Id, issue.Subject)
		project, err := deps.Model.API().Project(issue.Project.Id)
		if err != nil {
			return false, fmt.Errorf("failed to get redmine project err: %v", err)
		}
		log.Printf("Project (%d) %q - Issue (%d): %q", project.Id, project.Name, issue.Id, issue.Subject)

		projectRepo, err := deps.Model.DBGetRepository(*project)
		if err != nil {
			return false, fmt.Errorf("failed to get redmine repository err: %v", err)
		}
		//log.Printf("Repository %d: %s", projectRepo.ID, projectRepo.RootURL)

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
			deps.LlmPool,
			issue,
			parent,
			parents,
			closedChildrenIDs,
			allChildren,
			siblings,
			*project,
			projectConfig,
			wb,
			params.CodingAgents,
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
