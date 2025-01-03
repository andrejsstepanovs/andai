package redmine

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	workflow "github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/andrejsstepanovs/andai/pkg/redmine/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-redmine"
	"github.com/spf13/viper"
)

const (
	queryGetProjectTrackers   = "SELECT tracker_id FROM projects_trackers WHERE project_id = ?"
	queryInsertProjectTracker = "INSERT INTO projects_trackers (project_id, tracker_id) VALUES(?, ?)"
	queryInsertTracker        = "INSERT INTO trackers (name, description, position, default_status_id) VALUES (?, ?, ?, ?)"
	queryGetIssuePriority     = "SELECT id FROM enumerations WHERE type = 'IssuePriority' AND is_default = 1 AND name = ?"
	queryInsertIssuePriority  = "INSERT INTO enumerations (name, position, is_default, type, active, project_id, parent_id, position_name) VALUES (?, 1, 1, 'IssuePriority', 1, NULL, NULL, 'default')"
	queryInsertIssueStatus    = "INSERT INTO issue_statuses (name, is_closed, position) VALUES (?, ?, ?)"
)

func (c *Model) DbProjectTrackers(projectID int) ([]int, error) {
	var trackerIDs []int
	err := c.queryAndScan(queryGetProjectTrackers, func(rows *sql.Rows) error {
		var trackerID int
		if err := rows.Scan(&trackerID); err != nil {
			return err
		}
		trackerIDs = append(trackerIDs, trackerID)
		return nil
	}, projectID)

	if err != nil {
		return nil, err
	}

	return trackerIDs, nil
}

func (c *Model) DBSaveTrackers(trackers workflow.IssueTypes, defaultStatus redmine.IssueStatus, current []redmine.IdName) error {
	newTrackers := make([]workflow.IssueType, 0)
	for _, t := range trackers {
		exists := false
		for _, ct := range current {
			if ct.Name == string(t.Name) {
				log.Printf("Tracker %s already exists: %d\n", ct.Name, ct.Id)
				exists = true
				break
			}
		}
		if !exists {
			log.Println(fmt.Sprintf("Tracker: %s", t.Name))
			newTrackers = append(newTrackers, t)
		}
	}

	if len(newTrackers) == 0 {
		log.Println("Trackers OK")
		return nil
	}

	for i, tracker := range newTrackers {
		log.Println(fmt.Sprintf("Creating New Tracker: %s", tracker.Name))
		err := c.DBInsertTracker(tracker, i+1, defaultStatus.Id)
		if err != nil {
			return fmt.Errorf("redmine tracker insert err: %v", err)
		}
	}

	return nil
}

func (c *Model) DbSaveProjectTrackers(project redmine.Project, allTrackers []redmine.IdName) error {
	existingTrackerIDs, err := c.DbProjectTrackers(project.Id)
	if err != nil {
		return fmt.Errorf("get project trackers for project %d err: %v", project.Id, err)
	}

	createTrackers := make([]redmine.IdName, 0)
	for _, tracker := range allTrackers {
		exists := false
		for _, existingTrackerID := range existingTrackerIDs {
			if tracker.Id == existingTrackerID {
				log.Printf("Project %q Tracker for %q already exists Identifier: %d\n", project.Identifier, tracker.Name, tracker.Id)
				exists = true
				break
			}
		}
		if !exists {
			createTrackers = append(createTrackers, tracker)
		}
	}

	for _, tracker := range createTrackers {
		log.Println(fmt.Sprintf("Tracker: %s", tracker.Name))
		err = c.DBInsertProjectTracker(project.Id, tracker.Id)
		if err != nil {
			return fmt.Errorf("redmine project tracker insert err: %v", err)
		}
	}
	return nil
}

func (c *Model) DBInsertProjectTracker(projectID, trackerID int) error {
	result, err := c.execDML(queryInsertProjectTracker, projectID, trackerID)
	if err != nil {
		return fmt.Errorf("error redmine project tracker insert: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("project tracker not saved")
	}

	log.Println("project tracker inserted")
	return nil
}

func (c *Model) DBInsertTracker(issueType workflow.IssueType, position, defaultState int) error {
	result, err := c.execDML(queryInsertTracker, issueType.Name, issueType.Description, position, defaultState)
	if err != nil {
		return fmt.Errorf("redmine tracker err: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("tracker not created")
	}
	return nil
}

func (c *Model) GetDefaultNormalPriority() (int, error) {
	var ids []int
	err := c.queryAndScan(queryGetIssuePriority, func(rows *sql.Rows) error {
		var row int
		if err := rows.Scan(&row); err != nil {
			return err
		}
		ids = append(ids, row)
		return nil
	}, ISSUE_PRIORITY)

	if err != nil && !errors.As(err, &sql.ErrNoRows) {
		return 0, err
	}
	for _, id := range ids {
		return id, nil
	}
	return 0, nil
}

func (c *Model) DBInsertDefaultNormalPriority() error {
	result, err := c.execDML(queryInsertIssuePriority, ISSUE_PRIORITY)
	if err != nil {
		return fmt.Errorf("redmine issue priority err: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("issue priority not created")
	}
	return nil
}

func (c *Model) DBSaveIssueStatuses(statuses []redmine.IssueStatus, current []redmine.IssueStatus) error {
	newStatuses := make([]redmine.IssueStatus, 0)
	for _, status := range statuses {
		exists := false
		for _, s := range current {
			if s.Name == status.Name {
				exists = true
				break
			}
		}
		if !exists {
			newStatuses = append(newStatuses, status)
		}
	}

	if len(newStatuses) == 0 {
		log.Println("Issue Statuses OK")
		return nil
	}

	for i, status := range newStatuses {
		log.Println(fmt.Sprintf("Creating New Issue Status: %s", status.Name))
		err := c.DBInsertIssueStatus(status, i+1)
		if err != nil {
			return fmt.Errorf("redmine issue status insert err: %v", err)
		}
	}

	return nil
}

func (c *Model) DBInsertIssueStatus(status redmine.IssueStatus, position int) error {
	result, err := c.execDML(queryInsertIssueStatus, status.Name, status.IsClosed, position)
	if err != nil {
		return fmt.Errorf("error redmine issue status insert: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("token not created")
	}
	return nil
}
