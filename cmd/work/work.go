package work

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/pkg/ai"
	"github.com/andrejsstepanovs/andai/pkg/employee"
	"github.com/andrejsstepanovs/andai/pkg/employee/actions"
	"github.com/andrejsstepanovs/andai/pkg/models"
	model "github.com/andrejsstepanovs/andai/pkg/redmine"
	"github.com/andrejsstepanovs/andai/pkg/workbench"
	"github.com/andrejsstepanovs/andai/pkg/worker"
	"github.com/spf13/cobra"
)

func newNextCommand(model *model.Model, llmNorm *ai.AI, projects models.Projects, workflow models.Workflow) *cobra.Command {
	return &cobra.Command{
		Use:   "next",
		Short: "Work with redmine",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Searching for next issue")

			issues, err := model.APIGetWorkableIssues(workflow)
			if err != nil {
				log.Println("Failed to get workable issue")
				return err
			}

			if len(issues) == 0 {
				log.Println("No workable issues found")
				return nil
			}

			log.Printf("FOUND WORKABLE ISSUES (%d)", len(issues))
			for _, issue := range issues {
				fmt.Printf("WORKING ON: %q in %q ID=%d: %s\n",
					workflow.IssueTypes.Get(models.IssueTypeName(issue.Tracker.Name)).Name,
					workflow.States.Get(models.StateName(issue.Status.Name)).Name,
					issue.Id,
					issue.Subject,
				)

				parent, err := model.APIGetParent(issue)
				if err != nil {
					return fmt.Errorf("failed to get redmine parent issue err: %v", err)
				}

				parents, err := model.APIGetAllParents(issue)
				if err != nil {
					return fmt.Errorf("failed to get redmine all parent issues err: %v", err)
				}

				closedChildrenIDs, err := model.DBGetClosedChildrenIDs(issue.Id)
				if err != nil {
					return fmt.Errorf("failed to get redmine children ids err: %v", err)
				}

				children, err := model.APIGetChildren(issue)
				if err != nil {
					return fmt.Errorf("failed to get redmine issue relations err: %v", err)
				}

				siblings, err := model.APIGetIssueSiblings(issue)
				if err != nil {
					return fmt.Errorf("failed to get redmine issue siblings err: %v", err)
				}

				log.Printf("Issue %d: %s", issue.Id, issue.Subject)
				project, err := model.API().Project(issue.Project.Id)
				if err != nil {
					return fmt.Errorf("failed to get redmine project err: %v", err)
				}
				log.Printf("Project %d: %s", project.Id, project.Name)

				projectRepo, err := model.DBGetRepository(*project)
				if err != nil {
					return fmt.Errorf("failed to get redmine repository err: %v", err)
				}
				log.Printf("Repository %d: %s", projectRepo.ID, projectRepo.RootURL)

				projectConfig := projects.Find(project.Identifier)
				git, err := worker.FindProjectGit(projectConfig, projectRepo)
				if err != nil {
					return fmt.Errorf("failed to find project git err: %v", err)
				}
				log.Printf("Project Repository Opened %s", git.GetPath())

				wb := &workbench.Workbench{
					Git:   git,
					Issue: issue,
				}

				work := employee.NewEmployee(
					model,
					llmNorm,
					issue,
					parent,
					parents,
					closedChildrenIDs,
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
					return fmt.Errorf("failed to finish work on issue err: %v", err)
				}

				err = actions.TransitionToNextStatus(workflow, model, issue, success)
				if err != nil {
					return fmt.Errorf("failed to comment issue err: %v", err)
				}

				// stop after first issue
				break // nolint:staticcheck
			}

			return nil
		},
	}
}
