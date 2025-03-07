package work

import (
	"fmt"
	"log"
	"time"

	"github.com/andrejsstepanovs/andai/pkg/deps"
	"github.com/andrejsstepanovs/andai/pkg/employee"
	"github.com/andrejsstepanovs/andai/pkg/employee/actions"
	"github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/andrejsstepanovs/andai/pkg/workbench"
	"github.com/andrejsstepanovs/andai/pkg/worker"
	"github.com/spf13/cobra"
)

func newLoopCommand(deps *deps.AppDependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "loop",
		Short: "Work forever",
		RunE: func(_ *cobra.Command, _ []string) error {
			return Loop(deps)
		},
	}
}

func Loop(deps *deps.AppDependencies) error {
	for {
		settings, err := deps.Config.Load()
		if err != nil {
			return err
		}

		err = processTriggers(deps.Model, settings.Workflow)
		if err != nil {
			return fmt.Errorf("failed to process triggers err: %v", err)
		}
		wasWorking, err := workNext(deps, settings)
		if err != nil {
			return fmt.Errorf("failed to work next err: %v", err)
		}
		if !wasWorking {
			log.Println("No workable issues found")
			time.Sleep(5 * time.Second)
		}
	}
}

func newNextCommand(deps *deps.AppDependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "next",
		Short: "Work with redmine",
		RunE: func(_ *cobra.Command, _ []string) error {
			settings, err := deps.Config.Load()
			if err != nil {
				return err
			}

			log.Println("Searching for next issue")
			_, err = workNext(deps, settings)
			return err
		},
	}
}

func workNext(deps *deps.AppDependencies, settings *models.Settings) (bool, error) {
	issues, err := deps.Model.APIGetWorkableIssues(settings.Workflow)
	if err != nil {
		log.Println("Failed to get workable issue")
		return false, err
	}

	if len(issues) == 0 {
		log.Println("No workable issues found")
		return false, nil
	}

	log.Printf("FOUND WORKABLE ISSUES (%d)", len(issues))
	for _, issue := range issues {
		fmt.Printf("WORKING ON: %q in %q ID=%d: %s\n",
			settings.Workflow.IssueTypes.Get(models.IssueTypeName(issue.Tracker.Name)).Name,
			settings.Workflow.States.Get(models.StateName(issue.Status.Name)).Name,
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
			return false, fmt.Errorf("failed to get redmine children ids err: %v", err)
		}

		children, err := deps.Model.APIGetChildren(issue)
		if err != nil {
			return false, fmt.Errorf("failed to get redmine issue relations err: %v", err)
		}

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

		projectConfig := settings.Projects.Find(project.Identifier)
		git, err := worker.FindProjectGit(projectConfig, projectRepo)
		if err != nil {
			return false, fmt.Errorf("failed to find project git err: %v", err)
		}
		log.Printf("Project Repository Opened %s", git.GetPath())

		wb := &workbench.Workbench{
			Git:   git,
			Issue: issue,
		}

		work := employee.NewEmployee(
			deps.Model,
			deps.LlmNorm,
			issue,
			parent,
			parents,
			closedChildrenIDs,
			children,
			siblings,
			*project,
			projectConfig,
			wb,
			settings.Aider,
			settings.Workflow.States.Get(models.StateName(issue.Status.Name)),
			settings.Workflow.IssueTypes.Get(models.IssueTypeName(issue.Tracker.Name)),
			settings.Workflow.IssueTypes,
			projectRepo,
		)
		success, err := work.ExecuteWorkflow()
		if err != nil {
			return false, fmt.Errorf("failed to finish work on issue err: %v", err)
		}

		err = actions.TransitionToNextStatus(settings.Workflow, deps.Model, issue, success)
		if err != nil {
			return false, fmt.Errorf("failed to comment issue err: %v", err)
		}

		// true if issue was successfully worked on
		return true, nil // nolint:staticcheck
	}

	return false, nil
}
