package work

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/internal/deps"
	"github.com/andrejsstepanovs/andai/internal/models"
	model "github.com/andrejsstepanovs/andai/internal/redmine"
	"github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
)

func newTriggersCommand(deps *deps.AppDependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "triggers",
		Short: "Checks last issue status change and applies workflow triggers",
		RunE: func(_ *cobra.Command, _ []string) error {
			settings, err := deps.Config.Load()
			if err != nil {
				return err
			}

			log.Println("Starting triggers check")
			return processTriggers(deps.Model, settings.Workflow)
		},
	}
}

func processTriggers(model *model.Model, workflow models.Workflow) error {
	issueID, statusIDFrom, statusIDTo, err := model.DBGetLastStatusChange()
	if err != nil {
		log.Println("Failed to get last status change")
		return err
	}
	if issueID == 0 {
		log.Println("No status change found")
		return nil
	}
	issue, err := model.API().Issue(issueID)
	if err != nil {
		log.Printf("Failed to get issue: %d", issueID)
		return err
	}
	if issue == nil {
		log.Printf("Issue not found: %d", issueID)
		return nil
	}
	log.Printf("Checking %q id=%d project=%d\n", issue.Tracker.Name, issue.Id, issue.Project.Id)

	statusFrom, err := model.APIGetIssueStatusByID(statusIDFrom)
	if err != nil {
		log.Printf("Failed to get status: %d", statusIDFrom)
		return err
	}
	statusTo, err := model.APIGetIssueStatusByID(statusIDTo)
	if err != nil {
		log.Printf("Failed to get status: %s", statusTo.Name)
		return err
	}

	log.Printf("Last status change for %q %d: %q: %d -> %d (%s -> %s)\n", issue.Tracker.Name, issue.Id, issue.Subject, statusIDFrom, statusIDTo, statusFrom.Name, statusTo.Name)

	triggers := workflow.Triggers.GetTriggers(models.IssueTypeName(issue.Tracker.Name))
	if len(triggers) == 0 {
		log.Println("No triggers found")
		return nil
	}
	log.Printf("Triggers for %q found: %d\n", issue.Tracker.Name, len(triggers))

	for _, trigger := range triggers {
		err = processTrigger(workflow, trigger, model, statusFrom, statusTo, issue)
		if err != nil {
			return fmt.Errorf("failed to process trigger err: %v", err)
		}
	}

	return nil
}

func processTrigger(
	workflow models.Workflow,
	trigger models.Trigger,
	model *model.Model,
	statusFrom redmine.IssueStatus,
	statusTo redmine.IssueStatus,
	issue *redmine.Issue,
) error {
	action := trigger.GetTriggerIf(models.StateName(statusTo.Name))
	if action == nil {
		//log.Printf("No action found for %q %d %s -> %s\n", issue.Tracker.Name, issue.Id, statusFrom.Name, statusTo.Name)
		return nil
	}
	log.Printf("Trigger Action was found for %q %d %s -> %s\n", issue.Tracker.Name, issue.Id, statusFrom.Name, statusTo.Name)

	parent, err := model.APIGetParent(*issue)
	if err != nil {
		return fmt.Errorf("failed to get redmine parent issue err: %v", err)
	}
	if parent != nil {
		log.Printf("Parent found for %q %d: %q\n", issue.Tracker.Name, issue.Id, parent.Subject)
	} else {
		log.Printf("No parent found for %q %d\n", issue.Tracker.Name, issue.Id)
	}
	children, err := model.APIGetChildren(*issue)
	if err != nil {
		return fmt.Errorf("failed to get redmine children issue err: %v", err)
	}
	log.Printf("Children found for %q %d: %d\n", issue.Tracker.Name, issue.Id, len(children))

	siblings, err := model.APIGetIssueSiblings(*issue)
	if err != nil {
		log.Printf("Failed to get siblings for %q %d\n", issue.Tracker.Name, issue.Id)
	}
	log.Printf("Siblings found for %q %d: %d\n", issue.Tracker.Name, issue.Id, len(siblings))

	siblingsStatuses := make([]models.StateName, 0)
	for _, sibling := range siblings {
		status, err := model.APIGetIssueStatusByID(sibling.Status.Id)
		if err != nil {
			log.Printf("Failed to get status for sibling %q %d\n", sibling.Tracker.Name, sibling.Id)
		}
		siblingsStatuses = append(siblingsStatuses, models.StateName(status.Name))
	}
	log.Printf("Siblings statuses for %q %d: %v\n", issue.Tracker.Name, issue.Id, siblingsStatuses)
	siblingStatusOK := action.AllSiblingsCheck(siblingsStatuses)
	log.Printf("Siblings status check for %q %d: %t\n", issue.Tracker.Name, issue.Id, siblingStatusOK)

	if !siblingStatusOK {
		return nil
	}

	nextIssueStatus, err := model.APIGetIssueStatus(string(action.TriggerTransition.To))
	if err != nil {
		return fmt.Errorf("failed to get next issue status err: %v", err)
	}

	return processTriggerWho(action, children, workflow, model, nextIssueStatus, parent, issue)
}

func processTriggerWho(
	action *models.TriggerIf,
	children []redmine.Issue,
	workflow models.Workflow,
	model *model.Model,
	nextIssueStatus redmine.IssueStatus,
	parent *redmine.Issue,
	issue *redmine.Issue,
) error {
	switch action.TriggerTransition.Who {
	case models.TriggerTransitionWhoChildren:
		if len(children) == 0 {
			log.Printf("Should transition children to %q but no children to work with (%q %d)\n", nextIssueStatus.Name, issue.Tracker.Name, issue.Id)
			return nil
		}
		for _, child := range children {
			childState := workflow.States.Get(models.StateName(child.Status.Name))
			log.Printf("Transitioning child %q %d - %q -> %q\n", child.Tracker.Name, child.Id, childState.Name, nextIssueStatus.Name)
		}
		for _, child := range children {
			err := model.Transition(child, nextIssueStatus)
			if err != nil {
				return fmt.Errorf("failed to transition issue err: %v", err)
			}
			fmt.Printf("Successfully moved %d to: %d - %s\n", child.Id, nextIssueStatus.Id, nextIssueStatus.Name)
			// todo, check if this transition triggers something else
		}
	case models.TriggerTransitionWhoParent:
		if parent == nil {
			log.Println("Should have transition parent, but no parent was found")
			return nil
		}
		parentState := workflow.States.Get(models.StateName(parent.Status.Name))
		log.Printf(
			"Transitioning parent %q %d - %q -> %q\n",
			parent.Tracker.Name,
			parent.Id,
			parentState.Name,
			nextIssueStatus.Name,
		)
		err := model.Transition(*parent, nextIssueStatus)
		if err != nil {
			return fmt.Errorf("failed to transition issue err: %v", err)
		}
	}
	return nil
}
