package actions

import (
	"errors"
	"fmt"

	model "github.com/andrejsstepanovs/andai/internal/redmine"
	"github.com/andrejsstepanovs/andai/internal/settings"
	"github.com/mattn/go-redmine"
)

func TransitionToNextStatus(workflow settings.Workflow, model *model.Model, issue redmine.Issue, success bool) error {
	nextTransition := workflow.Transitions.GetNextTransition(settings.StateName(issue.Status.Name))
	nextTransition.LogPrint()
	nextIssueStatus, err := model.APIGetIssueStatus(string(nextTransition.GetTarget(success)))
	if err != nil {
		return fmt.Errorf("failed to get next issue status err: %v", err)
	}
	if nextIssueStatus.Id == 0 {
		return errors.New("next status not found")
	}

	fmt.Printf("Next status: %d - %s\n", nextIssueStatus.Id, nextIssueStatus.Name)

	err = model.Transition(issue, nextIssueStatus)
	if err != nil {
		return fmt.Errorf("failed to transition issue err: %v", err)
	}
	fmt.Printf("Successfully moved to: %d - %s\n", nextIssueStatus.Id, nextIssueStatus.Name)
	return nil
}
