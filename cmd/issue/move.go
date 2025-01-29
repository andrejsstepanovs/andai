package issue

import (
	"errors"
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

			foundIssue, success, err := findIssue(model, args)
			if err != nil {
				return err
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

func newMoveChildrenCommand(model *model.Model, workflow models.Workflow) *cobra.Command {
	return &cobra.Command{
		Use:   "move-children",
		Short: "Moves All Issue children to next successful or failed step. First param - issue Subject, Second param - success|fail",
		RunE: func(_ *cobra.Command, args []string) error {
			log.Println("Moving issue children")

			foundIssue, success, err := findIssue(model, args)
			if err != nil {
				return err
			}

			children, err := model.APIGetChildren(foundIssue)
			if err != nil {
				return err
			}

			if len(children) == 0 {
				log.Println("No children found")
				return nil
			}

			if success {
				log.Printf("Moving issue %d children %d to success", len(children), foundIssue.Id)
			} else {
				log.Printf("Moving issue %d children %d to fail", len(children), foundIssue.Id)
			}

			for _, child := range children {
				err = actions.TransitionToNextStatus(workflow, model, child, success)
				if err != nil {
					return fmt.Errorf("failed to comment issue err: %v", err)
				}
			}
			return nil
		},
	}
}

func findIssue(model *model.Model, args []string) (redmine.Issue, bool, error) {
	if len(args) < 2 {
		log.Println("Not enough arguments")
		return redmine.Issue{}, false, errors.New("not enough arguments")
	}

	projects, err := model.APIGetProjects()
	if err != nil {
		log.Println("Failed to get projects")
		return redmine.Issue{}, false, err
	}
	if len(projects) == 0 {
		log.Println("No projects found")
		return redmine.Issue{}, false, err
	}
	if len(projects) > 1 {
		log.Println("Too many projects found")
		return redmine.Issue{}, false, err
	}
	project := projects[0]

	issueSubject := args[0]
	success, err := strconv.ParseBool(args[1])
	if err != nil {
		log.Println("Failed to parse success/fail")
		return redmine.Issue{}, false, err
	}

	issues, err := model.APIGetProjectIssues(project)
	if err != nil {
		log.Printf("Failed to get project issues: %v", err)
		return redmine.Issue{}, success, err
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
		return redmine.Issue{}, success, nil
	}

	return foundIssue, success, nil
}
