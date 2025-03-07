package issue

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/andrejsstepanovs/andai/pkg/deps"
	"github.com/andrejsstepanovs/andai/pkg/employee/actions"
	"github.com/andrejsstepanovs/andai/pkg/models"
	model "github.com/andrejsstepanovs/andai/pkg/redmine"
	"github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
)

func newMoveCommand(deps *deps.AppDependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "move",
		Short: "Move Issue to next successful or failed step. First param - issue Subject, Second param - success|fail",
		RunE: func(_ *cobra.Command, args []string) error {
			log.Println("Moving issue")
			settings, err := deps.Config.Load()
			if err != nil {
				return err
			}

			foundIssue, success, nestStatus, err := findIssue(deps.Model, args, settings.Workflow)
			if err != nil {
				return err
			}

			if nestStatus != "" {
				log.Printf("Moving issue %d to %s", foundIssue.Id, nestStatus)
				nextIssueStatus, err := deps.Model.APIGetIssueStatus(string(nestStatus))
				if err != nil {
					return fmt.Errorf("failed to get next issue status err: %v", err)
				}
				if nextIssueStatus.Id == 0 {
					return errors.New("next status not found")
				}

				err = deps.Model.Transition(foundIssue, nextIssueStatus)
				if err != nil {
					return fmt.Errorf("failed to comment issue err: %v", err)
				}
			} else {
				if success {
					log.Printf("Moving issue %d to success", foundIssue.Id)
				} else {
					log.Printf("Moving issue %d to fail", foundIssue.Id)
				}

				err = actions.TransitionToNextStatus(settings.Workflow, deps.Model, foundIssue, success)
				if err != nil {
					return fmt.Errorf("failed to comment issue err: %v", err)
				}
			}
			return nil
		},
	}
}

func newMoveChildrenCommand(deps *deps.AppDependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "move-children",
		Short: "Moves All Issue children to next successful or failed step. First param - issue Subject, Second param - success|fail",
		RunE: func(_ *cobra.Command, args []string) error {
			log.Println("Moving issue children")

			settings, err := deps.Config.Load()
			if err != nil {
				return err
			}

			foundIssue, success, useStatus, err := findIssue(deps.Model, args, settings.Workflow)
			if err != nil {
				return err
			}

			children, err := deps.Model.APIGetChildren(foundIssue)
			if err != nil {
				return err
			}

			if len(children) == 0 {
				log.Println("No children found")
				return nil
			}

			if useStatus != "" {
				log.Printf("Moving issue %d children %d to %s", len(children), foundIssue.Id, useStatus)
				log.Println("not implemented")
			} else {
				if success {
					log.Printf("Moving issue %d children %d to success", len(children), foundIssue.Id)
				} else {
					log.Printf("Moving issue %d children %d to fail", len(children), foundIssue.Id)
				}

				for _, child := range children {
					err = actions.TransitionToNextStatus(settings.Workflow, deps.Model, child, success)
					if err != nil {
						return fmt.Errorf("failed to comment issue err: %v", err)
					}
				}
			}
			return nil
		},
	}
}

func findIssue(model *model.Model, args []string, workflow models.Workflow) (redmine.Issue, bool, models.StateName, error) {
	if len(args) < 2 {
		log.Println("Not enough arguments")
		return redmine.Issue{}, false, "", errors.New("not enough arguments")
	}

	projects, err := model.APIGetProjects()
	if err != nil {
		log.Println("Failed to get projects")
		return redmine.Issue{}, false, "", err
	}
	if len(projects) == 0 {
		log.Println("No projects found")
		return redmine.Issue{}, false, "", err
	}
	if len(projects) > 1 {
		log.Println("Too many projects found")
		return redmine.Issue{}, false, "", err
	}
	project := projects[0]

	issueSubject := args[0]

	issues, err := model.APIGetProjectIssues(project)
	if err != nil {
		log.Printf("Failed to get project issues: %v", err)
		return redmine.Issue{}, false, "", err
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
		return redmine.Issue{}, false, "", nil
	}

	toTarget := models.StateName("")
	success, err := strconv.ParseBool(args[1])
	if err != nil {
		toTarget = models.StateName(args[1])
		transitions := workflow.Transitions.GetTransitions(models.StateName(foundIssue.Status.Name))
		found := false
		for _, transition := range transitions {
			if transition.Target == toTarget {
				log.Printf("Transition found: %s -> %s", foundIssue.Status.Name, toTarget)
				found = true
			}
		}
		if !found {
			log.Printf("Transition %q not found", toTarget)
			return redmine.Issue{}, false, "", err
		}
	}

	return foundIssue, success, toTarget, nil
}
