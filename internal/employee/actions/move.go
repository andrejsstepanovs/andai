package actions

import (
	"errors"
	"fmt"
	"log"

	model "github.com/andrejsstepanovs/andai/internal/redmine"
	"github.com/andrejsstepanovs/andai/internal/settings"
	"github.com/mattn/go-redmine"
)

func TransitionToNextStatus(workflow settings.Workflow, model *model.Model, issue redmine.Issue, success bool) error {
	nextTransition := workflow.Transitions.GetNextTransition(settings.StateName(issue.Status.Name))
	//nextTransition.LogPrint()
	nextIssueStatus, err := model.APIGetIssueStatus(string(nextTransition.GetTarget(success)))
	if err != nil {
		return fmt.Errorf("failed to get next issue status err: %v", err)
	}
	if nextIssueStatus.Id == 0 {
		return errors.New("next status not found")
	}

	//log.Printf("Next issue status (%d): %s\n", issue.Id, nextIssueStatus.Name)
	updatedIssue, err := model.API().Issue(issue.Id) // load issue again to get the latest data, because Transition will update
	if err != nil {
		return fmt.Errorf("failed to reload issue after saving custom field: %v", err)
	}

	err = model.Transition(*updatedIssue, nextIssueStatus)
	if err != nil {
		return fmt.Errorf("failed to transition issue err: %v", err)
	}

	log.Printf("Successfully moved issue (%d) to: %q", issue.Id, nextIssueStatus.Name)

	return nil
}
