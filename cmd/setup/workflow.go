package setup

import (
	"fmt"
	"sort"

	"github.com/andrejsstepanovs/andai/pkg/models"
	model "github.com/andrejsstepanovs/andai/pkg/redmine"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
)

func newWorkflowCommand(model *model.Model, workflowConfig models.Workflow) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Setup Redmine workflow",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println("Update Redmine workflow")

			fmt.Println("States:", len(workflowConfig.States))
			statuses := convertToStatuses(workflowConfig.States)
			statuses = sortStatuses(statuses)
			err := model.SaveIssueStatuses(statuses)
			if err != nil {
				fmt.Println("Redmine Settings Failed to enable API")
				return fmt.Errorf("error redmine: %v", err)
			}

			defaultStatus, err := model.APIGetDefaultStatus()
			if err != nil {
				return fmt.Errorf("error redmine: %v", err)
			}

			fmt.Println("Trackers:", len(workflowConfig.IssueTypes))
			err = model.SaveTrackers(workflowConfig.IssueTypes, defaultStatus)
			if err != nil {
				fmt.Println("Failed to save trackers")
				return fmt.Errorf("redmine err: %v", err)
			}

			return nil
		},
	}
	return cmd
}

func convertToStatuses(workflowStates models.States) []redmine.IssueStatus {
	statuses := make([]redmine.IssueStatus, 0)
	for _, state := range workflowStates {
		statuses = append(statuses, redmine.IssueStatus{
			Name:      string(state.Name),
			IsDefault: state.IsDefault,
			IsClosed:  state.IsClosed,
		})
	}

	return statuses
}

// sortStatuses Sort statuses: IsDefault first, IsClosed last
func sortStatuses(statuses []redmine.IssueStatus) []redmine.IssueStatus {
	sort.Slice(statuses, func(i, j int) bool {
		// If i is default and j is not, i should come first
		if statuses[i].IsDefault && !statuses[j].IsDefault {
			return true
		}
		// If j is default and i is not, j should come first
		if statuses[j].IsDefault && !statuses[i].IsDefault {
			return false
		}
		// If i is closed and j is not, i should come last
		if statuses[i].IsClosed && !statuses[j].IsClosed {
			return false
		}
		// If j is closed and i is not, j should come last
		if statuses[j].IsClosed && !statuses[i].IsClosed {
			return true
		}
		// Otherwise, maintain the original order
		return i < j
	})

	return statuses
}
