package work

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/pkg/employee"
	"github.com/andrejsstepanovs/andai/pkg/employee/actions"
	"github.com/andrejsstepanovs/andai/pkg/llm"
	"github.com/andrejsstepanovs/andai/pkg/models"
	model "github.com/andrejsstepanovs/andai/pkg/redmine"
	"github.com/andrejsstepanovs/andai/pkg/workbench"
	"github.com/andrejsstepanovs/andai/pkg/worker"
	"github.com/spf13/cobra"
)

func newWorkCommand(_ *model.Model, llm *llm.LLM, models models.LlmModels) *cobra.Command {
	return &cobra.Command{
		Use:   "once",
		Short: "Work with redmine",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Starting work with redmine")
			log.Printf("Models %d", len(models))

			response, err := llm.Simple("Hi!")
			if err != nil {
				log.Println("Failed to get response from LLM")
			}

			log.Println(response)

			return nil
		},
	}
}

func newNextCommand(model *model.Model, llm *llm.LLM, projects models.Projects, workflow models.Workflow) *cobra.Command {
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

				children, err := model.APIGetChildren(issue)
				if err != nil {
					return fmt.Errorf("failed to get redmine issue relations err: %v", err)
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
				work := employee.NewWorkOnIssue(
					model,
					llm,
					issue,
					parent,
					parents,
					children,
					*project,
					projectConfig,
					wb,
					workflow.States.Get(models.StateName(issue.Status.Name)),
					workflow.IssueTypes.Get(models.IssueTypeName(issue.Tracker.Name)),
					workflow.IssueTypes,
				)
				success := work.Work()

				err = actions.TransitionToNextStatus(workflow, model, issue, success)
				if err != nil {
					return fmt.Errorf("failed to comment issue err: %v", err)
				}
			}

			return nil
		},
	}
}
