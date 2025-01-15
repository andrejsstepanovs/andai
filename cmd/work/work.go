package work

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/pkg/llm"
	"github.com/andrejsstepanovs/andai/pkg/models"
	model "github.com/andrejsstepanovs/andai/pkg/redmine"
	"github.com/andrejsstepanovs/andai/pkg/worker"
	"github.com/mattn/go-redmine"
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
				log.Printf("Issue %d: %s", issue.Id, issue.Subject)
				project, err := model.Api().Project(issue.Project.Id)
				if err != nil {
					return fmt.Errorf("failed to get redmine project err: %v", err)
				}
				log.Printf("Project %d: %s", project.Id, project.Name)

				projectRepo, err := model.DbGetRepository(*project)
				if err != nil {
					return fmt.Errorf("failed to get redmine repository err: %v", err)
				}
				log.Printf("Repository %d: %s", projectRepo.ID, projectRepo.RootUrl)

				projectConfig := projects.Find(project.Identifier)
				git, err := worker.FindProjectGit(projectConfig, projectRepo)
				if err != nil {
					return fmt.Errorf("failed to find project git err: %v", err)
				}

				log.Printf("Project Repository Opened %s", git.Path)

				work := NewWorkOnIssue(
					model,
					llm,
					issue,
					*project,
					projectConfig,
					git,
					workflow.States.Get(models.StateName(issue.Status.Name)),
					workflow.IssueTypes.Get(models.IssueTypeName(issue.Tracker.Name)),
				)
				success := work.Work()

				_ = success
				//err = model.Comment(issue, "AI finished Work - Success: "+strconv.FormatBool(success))
				//if err != nil {
				//	return fmt.Errorf("failed to comment issue err: %v", err)
				//}
				//
				//err = transitionToNextStatus(workflow, model, issue, success)
				//if err != nil {
				//	return fmt.Errorf("failed to comment issue err: %v", err)
				//}
			}

			return nil
		},
	}
}

func transitionToNextStatus(workflow models.Workflow, model *model.Model, issue redmine.Issue, success bool) error {
	nextTransition := workflow.Transitions.GetNextTransition(models.StateName(issue.Status.Name))
	nextTransition.LogPrint()
	nextIssueStatus, err := model.APIGetIssueStatus(string(nextTransition.GetTarget(success)))
	if err != nil {
		return fmt.Errorf("failed to get next issue status err: %v", err)
	}
	fmt.Printf("Next status: %d - %s\n", nextIssueStatus.Id, nextIssueStatus.Name)

	err = model.Transition(issue, nextIssueStatus)
	if err != nil {
		return fmt.Errorf("failed to transition issue err: %v", err)
	}
	fmt.Printf("Successfully moved to: %d - %s\n", nextIssueStatus.Id, nextIssueStatus.Name)
	return nil
}
