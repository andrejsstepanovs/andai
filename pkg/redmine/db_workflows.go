package redmine

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/andrejsstepanovs/andai/pkg/redmine/models"
	_ "github.com/go-sql-driver/mysql"
)

const (
	queryInsertWorkflow = "INSERT INTO workflows (tracker_id, old_status_id, new_status_id, role_id, type) VALUES (?, ?, ?, ?, 'WorkflowTransition')"
	queryGetWorkflow    = "SELECT id, tracker_id, old_status_id, new_status_id, role_id, assignee, author, type, field_name, rule FROM workflows WHERE tracker_id = ?"
)

func (c *Model) DBInsertWorkflow(trackerID, oldStatusID, newStatusID, roleID int) error {
	result, err := c.execDML(queryInsertWorkflow, trackerID, oldStatusID, newStatusID, roleID)
	if err != nil {
		return fmt.Errorf("insert workflow err: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("workflow row affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("workflow not updated")
	}
	return nil
}

func (c *Model) DBGetWorkflows(trackerID int) ([]models.Workflow, error) {
	var workflows []models.Workflow
	err := c.queryAndScan(queryGetWorkflow, func(rows *sql.Rows) error {
		var row models.Workflow
		if err := rows.Scan(
			&row.ID,
			&row.TrackerID,
			&row.OldStatusID,
			&row.NewStatusID,
			&row.RoleID,
			&row.Assignee,
			&row.Author,
			&row.Type,
			&row.FieldName,
			&row.Rule,
		); err != nil {
			return err
		}
		workflows = append(workflows, row)
		return nil
	}, trackerID)

	if err != nil {
		return nil, err
	}

	return workflows, nil
}
