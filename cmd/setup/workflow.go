package setup

import (
	"fmt"
	"log"
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
			log.Println("Update Redmine workflow")

			log.Println("Issue States:", len(workflowConfig.States))
			statuses := convertToStatuses(workflowConfig.States)
			statuses = sortStatuses(statuses)

			currentIssueStatuses, err := model.Api().IssueStatuses()
			if err != nil {
				return fmt.Errorf("error redmine issue status: %v", err)
			}
			err = model.DBSaveIssueStatuses(statuses, currentIssueStatuses)
			if err != nil {
				log.Println("Redmine Settings Failed to enable API")
				return fmt.Errorf("error redmine: %v", err)
			}

			firstStatus := workflowConfig.States.GetFirst()
			defaultStatus, err := model.APIGetIssueStatusByName(string(firstStatus.Name))
			if err != nil {
				return fmt.Errorf("error redmine: %v", err)
			}
			log.Println("Default Status:", defaultStatus.Name)

			log.Println("Trackers:", len(workflowConfig.IssueTypes))

			currentTrackers, err := model.Api().Trackers()
			if err != nil {
				return fmt.Errorf("error redmine trackers: %v", err)
			}
			err = model.DBSaveTrackers(workflowConfig.IssueTypes, defaultStatus, currentTrackers)
			if err != nil {
				log.Println("Failed to save trackers")
				return fmt.Errorf("redmine err: %v", err)
			}

			err = projectTrackers(model)
			if err != nil {
				return err
			}

			err = transitions(model, workflowConfig, defaultStatus)
			if err != nil {
				return err
			}

			priorityID, err := model.GetDefaultNormalPriority()
			if err != nil {
				return err
			}
			if priorityID == 0 {
				err = model.DBInsertDefaultNormalPriority()
				if err != nil {
					return err
				}
				log.Println("Default Normal Priority inserted")
			} else {
				log.Println("Default Normal Priority already exists")
			}

			return nil
		},
	}
	return cmd
}

func projectTrackers(model *model.Model) error {
	projects, err := model.ApiGetProjects()
	log.Println("Projects:", len(projects))
	if err != nil {
		log.Println("Failed to get projects")
		return fmt.Errorf("redmine err: %v", err)
	}
	allTrackers, err := model.Api().Trackers()
	if err != nil {
		return fmt.Errorf("error redmine trackers: %v", err)
	}
	for _, project := range projects {
		err = model.DBSaveProjectTrackers(project, allTrackers)
		if err != nil {
			log.Printf("Redmine Project %q Trackers Save Fail\n", project.Name)
			return fmt.Errorf("error redmine project trackers save: %v", err)
		}
		log.Printf("Redmine Project %q Trackers OK\n", project.Name)
	}
	return nil
}

func transitions(model *model.Model, workflowConfig models.Workflow, defaultStatus redmine.IssueStatus) error {
	log.Println("Transitions:", len(workflowConfig.Transitions))
	trackers, err := model.Api().Trackers()
	if err != nil {
		log.Println("Failed to get trackers")
		return fmt.Errorf("redmine err: %v", err)
	}
	statuses, err := model.Api().IssueStatuses()
	if err != nil {
		log.Println("Failed to get issue statuses")
		return fmt.Errorf("redmine err: %v", err)
	}
	roleID, err := model.DBGetWorkerRole()
	if err != nil {
		log.Println("Failed to get worker role")
		return fmt.Errorf("redmine err: %v", err)
	}

	for _, tracker := range trackers {
		err = saveTransitions(model, tracker, statuses, defaultStatus, workflowConfig.Transitions, roleID)
		if err != nil {
			log.Println("Failed to save transitions")
			return fmt.Errorf("redmine err: %v", err)
		}
	}

	return nil
}

func saveTransitions(model *model.Model, tracker redmine.IdName, statuses []redmine.IssueStatus, defaultStatus redmine.IssueStatus, transitions models.Transitions, roleID int) error {
	type key struct {
		fromID    int
		toID      int
		trackerID int
		roleID    int
	}
	list := make(map[key]bool)
	list[key{fromID: 0, toID: defaultStatus.Id, trackerID: tracker.Id, roleID: roleID}] = false
	for _, transition := range transitions {
		fromID, toID := transition.GetIDs(statuses)
		log.Printf("Transition: %s -> %s (%d -> %d)\n", transition.Source, transition.Target, fromID, toID)
		list[key{fromID: fromID, toID: toID, trackerID: tracker.Id, roleID: roleID}] = false
	}

	workflows, err := model.DBGetWorkflows(tracker.Id)
	if err != nil {
		log.Printf("Failed to get workflows for Tracker: %s", tracker.Name)
		return fmt.Errorf("redmine err: %v", err)
	}
	for _, workflow := range workflows {
		for k := range list {
			if workflow.TrackerID == k.trackerID &&
				workflow.OldStatusID == k.fromID &&
				workflow.NewStatusID == k.toID &&
				workflow.RoleID == k.roleID {
				list[k] = true
				break
			}
		}
	}

	for l, exists := range list {
		if exists {
			continue
		}

		err = model.DBInsertWorkflow(l.trackerID, l.fromID, l.toID, l.roleID)
		if err != nil {
			log.Printf("Failed to save transitions for Tracker: %s", tracker.Name)
			return fmt.Errorf("redmine err: %v", err)
		}
		log.Printf("Saved transition %d -> %d for Tracker: %s and Role : %d\n", l.fromID, l.toID, tracker.Name, l.roleID)
	}

	return nil
}

func convertToStatuses(workflowStates models.States) []redmine.IssueStatus {
	statuses := make([]redmine.IssueStatus, 0)
	for _, state := range workflowStates {
		statuses = append(statuses, redmine.IssueStatus{
			Name:      string(state.Name),
			IsDefault: state.IsFirst,
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
