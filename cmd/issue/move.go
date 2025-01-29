package issue

import (
	"fmt"
	"log"
	"strconv"

	"github.com/andrejsstepanovs/andai/pkg/employee/actions"
	"github.com/andrejsstepanovs/andai/pkg/models"
	model "github.com/andrejsstepanovs/andai/pkg/redmine"
	"github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
)

func newMoveCommand(model *model.Model, workflow models.Workflow) *cobra.Command {
	return &cobra.Command{
		Use:   "move",
		Short: "Move Issue to next successful or failed step. First param - issue Subject, Second param - success|fail",
		RunE: func(_ *cobra.Command, args []string) error {
			log.Println("Moving issue")

			if len(args) < 2 {
				log.Println("Not enough arguments")
				return nil
			}

			projects, err := model.APIGetProjects()
			if err != nil {
				log.Println("Failed to get projects")
				return err
			}
			if len(projects) == 0 {
				log.Println("No projects found")
				return err
			}
			if len(projects) > 1 {
				log.Println("Too many projects found")
				return err
			}
			project := projects[0]

			issueSubject := args[0]
			success, err := strconv.ParseBool(args[1])
			if err != nil {
				log.Println("Failed to parse success/fail")
				return err
			}

			issues, err := model.APIGetProjectIssues(project)
			if err != nil {
				log.Printf("Failed to get project issues: %v", err)
				return err
			}

			var foundIssue redmine.Issue
			for _, issue := range issues {
				if issue.Subject == issueSubject {
					foundIssue = issue
					log.Printf("Found issue: %q - ID: %d", issueSubject, foundIssue.Id)
					break
				}
			}
			if foundIssue.Id == 0 {
				log.Printf("Issue not found: %s", issueSubject)
				return nil
			}

			if success {
				log.Printf("Moving issue %d to success", foundIssue.Id)
			} else {
				log.Printf("Moving issue %d to fail", foundIssue.Id)
			}

			err = actions.TransitionToNextStatus(workflow, model, foundIssue, success)
			if err != nil {
				return fmt.Errorf("failed to comment issue err: %v", err)
			}
			return nil
		},
	}
}
