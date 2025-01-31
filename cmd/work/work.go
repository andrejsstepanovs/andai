// Package work provides functionality for processing Redmine issues through configured workflows.
// It handles the core workflow execution logic for the AndAI framework, including:
//   - Retrieving and processing workable issues from Redmine
//   - Managing issue relationships (parents, children, siblings)
//   - Setting up project contexts and workbenches
//   - Executing configured workflows for issues
//   - Transitioning issues to their next workflow states
//
// The package exposes commands for working with Redmine issues through the cobra CLI framework.
package work

// AI: Helper function to process a single issue
func processIssue(model *model.Model, llmNorm *ai.AI, issue redmine.Issue, projects models.Projects, workflow models.Workflow) (bool, error) {
	fmt.Printf("WORKING ON: %q in %q ID=%d: %s\n",
		workflow.IssueTypes.Get(models.IssueTypeName(issue.Tracker.Name)).Name,
		workflow.States.Get(models.StateName(issue.Status.Name)).Name,
		issue.Id,
		issue.Subject,
	)

	parent, parents, children, siblings, err := getIssueRelations(model, issue)
	if err != nil {
		return false, err
	}

	log.Printf("Issue %d: %s", issue.Id, issue.Subject)
	project, err := model.API().Project(issue.Project.Id)
	if err != nil {
		return false, fmt.Errorf("failed to get redmine project (ID: %d) for issue #%d: %w", 
			issue.Project.Id, issue.Id, err)
	}
	log.Printf("Project %d: %s", project.Id, project.Name)

	wb, err := getProjectContext(model, project, projects)
	if err != nil {
		return false, err
	}
	wb.Issue = issue
	log.Printf("Project Repository Opened %s", wb.Git.GetPath())

	work := employee.NewEmployee(
		model,
		llmNorm,
		issue,
		parent,
		parents,
		children,
		siblings,
		*project,
		projectConfig,
		wb,
		workflow.States.Get(models.StateName(issue.Status.Name)),
		workflow.IssueTypes.Get(models.IssueTypeName(issue.Tracker.Name)),
		workflow.IssueTypes,
	)
	success, err := work.ExecuteWorkflow()
	if err != nil {
		return false, fmt.Errorf("failed to execute workflow for issue #%d (%s): %w",
			issue.Id, issue.Subject, err)
	}

	err = actions.TransitionToNextStatus(workflow, model, issue, success)
	if err != nil {
		return false, fmt.Errorf("failed to transition issue #%d to next status (current: %s): %w",
			issue.Id, issue.Status.Name, err)
	}

	return success, nil
}

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/pkg/ai"

// AI: Helper function to get all issue relations (parent, parents, children, siblings)
func getIssueRelations(model *model.Model, issue redmine.Issue) (
	*redmine.Issue, // parent
	[]redmine.Issue, // parents 
	[]redmine.Issue, // children
	[]redmine.Issue, // siblings
	error,
) {
	parent, err := model.APIGetParent(issue)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to get parent issue for issue #%d: %w", issue.Id, err)
	}

	parents, err := model.APIGetAllParents(issue)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to get all parent issues for issue #%d: %w", issue.Id, err)
	}

	children, err := model.APIGetChildren(issue)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to get child issues for issue #%d: %w", issue.Id, err)
	}

	siblings, err := model.APIGetIssueSiblings(issue)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to get sibling issues for issue #%d: %w", issue.Id, err)
	}

	return parent, parents, children, siblings, nil
}
	"github.com/andrejsstepanovs/andai/pkg/employee"
	"github.com/andrejsstepanovs/andai/pkg/employee/actions"
	"github.com/andrejsstepanovs/andai/pkg/models"
	model "github.com/andrejsstepanovs/andai/pkg/redmine"
	"github.com/andrejsstepanovs/andai/pkg/workbench"
	"github.com/andrejsstepanovs/andai/pkg/worker"
	"github.com/spf13/cobra"
)

// AI: Helper function to get project context and setup workbench
func getProjectContext(model *model.Model, project *redmine.Project, projects models.Projects) (*workbench.Workbench, error) {
	projectRepo, err := model.DBGetRepository(*project)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository for project %s (ID: %d): %w", 
			project.Name, project.Id, err)
	}

	projectConfig := projects.Find(project.Identifier)
	if projectConfig == nil {
		return nil, fmt.Errorf("project configuration not found for identifier %q",
			project.Identifier)
	}

	git, err := worker.FindProjectGit(projectConfig, projectRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize git for project %s (ID: %d) with repo path %s: %w",
			project.Name, project.Id, projectRepo.RootURL, err)
	}

	return &workbench.Workbench{
		Git: git,
	}, nil
}

// newNextCommand creates a cobra command that processes the next available workable Redmine issue.
//
// The command:
//   - Searches for issues that are ready to be worked on based on the workflow configuration
//   - Processes the first workable issue found through the following steps:
//     1. Retrieves all issue relationships (parent, children, siblings)
//     2. Sets up project context and git workbench
//     3. Executes the configured workflow for the issue
//     4. Transitions the issue to its next status based on execution success
//
// Parameters:
//   - model: Redmine model for API and database operations
//   - llmNorm: AI client for LLM operations
//   - projects: Configuration for all available projects
//   - workflow: Workflow configuration defining states and transitions
//
// Returns a cobra.Command configured to execute the next issue processing workflow.
func newNextCommand(model *model.Model, llmNorm *ai.AI, projects models.Projects, workflow models.Workflow) *cobra.Command {
	return &cobra.Command{
		Use:   "next",
		Short: "Work with redmine",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Searching for next issue")

			issues, err := model.APIGetWorkableIssues(workflow)
			if err != nil {
				log.Printf("Failed to retrieve workable issues: %v", err)
				return fmt.Errorf("failed to retrieve workable issues from Redmine: %w", err)
			}

			if len(issues) == 0 {
				log.Println("No workable issues found in current workflow state")
				return nil
			}

			log.Printf("Found %d workable issues to process", len(issues))
			for _, issue := range issues {
				success, err := processIssue(model, llmNorm, issue, projects, workflow)
				if err != nil {
					return fmt.Errorf("failed to process issue #%d (%s): %w", 
						issue.Id, issue.Subject, err)
				}

				if success {
					log.Printf("Successfully processed issue #%d", issue.Id)
				} else {
					log.Printf("Issue #%d processing completed without success", issue.Id)
				}

				// stop after first issue
				break
			}

			return nil
		},
	}
}
