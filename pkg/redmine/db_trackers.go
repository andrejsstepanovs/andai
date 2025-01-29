package redmine

// Trackers are: Epic, Task, Step

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	workflow "github.com/andrejsstepanovs/andai/pkg/models"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/mattn/go-redmine"
)

const (
	queryGetProjectTrackers   = "SELECT tracker_id FROM projects_trackers WHERE project_id = ?"
	queryInsertProjectTracker = "INSERT INTO projects_trackers (project_id, tracker_id) VALUES(?, ?)"
	queryInsertTracker        = "INSERT INTO trackers (name, description, position, default_status_id) VALUES (?, ?, ?, ?)"
)

func (c *Model) DBGetTrackersByName(name string) (int, error) {
	allTrackers, err := c.API().Trackers()
	if err != nil {
		return 0, fmt.Errorf("error redmine trackers: %v", err)
	}
	for _, tracker := range allTrackers {
		if tracker.Name == name {
			return tracker.Id, nil
		}
	}
	return 0, fmt.Errorf("tracker %q not found", name)
}

func (c *Model) DBProjectTrackers(projectID int) ([]int, error) {
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
			log.Printf("Tracker: %s\n", t.Name)
			newTrackers = append(newTrackers, t)
		}
	}

	if len(newTrackers) == 0 {
		log.Println("Trackers OK")
		return nil
	}

	for i, tracker := range newTrackers {
		log.Printf("Creating New Tracker: %s\n", tracker.Name)
		err := c.DBInsertTracker(tracker, i+1, defaultStatus.Id)
		if err != nil {
			return fmt.Errorf("redmine tracker insert err: %v", err)
		}
	}

	return nil
}

func (c *Model) DBSaveProjectTrackers(project redmine.Project, allTrackers []redmine.IdName) error {
	existingTrackerIDs, err := c.DBProjectTrackers(project.Id)
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
		log.Printf("Tracker: %s\n", tracker.Name)
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
