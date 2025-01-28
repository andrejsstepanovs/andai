package work

import (
	"log"

	"github.com/andrejsstepanovs/andai/pkg/models"
	model "github.com/andrejsstepanovs/andai/pkg/redmine"
	"github.com/spf13/cobra"
)

func newTriggersCommand(model *model.Model, workflow models.Workflow) *cobra.Command {
	return &cobra.Command{
		Use:   "triggers",
		Short: "Checks last issue status change and applies workflow triggers",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Starting triggers check")

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

			statusFrom, err := model.APIGetIssueStatusByID(statusIDFrom)
			if err != nil {
				log.Printf("Failed to get status: %d", statusIDFrom)
				return err
			}
			statusTo, err := model.APIGetIssueStatusByID(statusIDTo)
			if err != nil {
				log.Printf("Failed to get status: %d", statusTo)
				return err
			}

			log.Printf("Last status change for %q %d: %q: %d -> %d (%s -> %s)\n", issue.Tracker.Name, issue.Id, issue.Subject, statusIDFrom, statusIDTo, statusFrom.Name, statusTo.Name)

			triggers := workflow.Triggers.GetTriggers(models.IssueTypeName(issue.Tracker.Name))
			if len(triggers) == 0 {
				log.Println("No triggers found")
				return nil
			}

			for _, trigger := range triggers {
				action := trigger.GetTriggerIf(models.StateName(statusTo.Name))
				if action == nil {
					log.Printf("No action found for %q %d %s -> %s\n", issue.Tracker.Name, issue.Id, statusFrom.Name, statusTo.Name)
					continue
				}
				log.Printf("Trigger Action was found for %q %d %s -> %s\n", issue.Tracker.Name, issue.Id, statusFrom.Name, statusTo.Name)

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
				siblingStatusOK := action.AllSiblingsCheck(siblingsStatuses)
				log.Printf("Siblings status check for %q %d: %t\n", issue.Tracker.Name, issue.Id, siblingStatusOK)

				if !siblingStatusOK {
					return nil
				}

				//switch action.AllSiblingsStatus {
				//case models.TriggerTransitionWhoParent:
				//	parent, err := model.APIGetParent(*issue)
				//	if err != nil {
				//		return fmt.Errorf("failed to get redmine parent issue err: %v", err)
				//	}
				//case models.TriggerTransitionWhoChildren:
				//
				//}
			}

			return nil
		},
	}
}
