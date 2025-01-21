package redmine

import (
	"errors"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-redmine"
)

const (
	queryInsertIssueStatus = "INSERT INTO issue_statuses (name, is_closed, position) VALUES (?, ?, ?)"
)

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
		log.Printf("Creating New Issue Status: %s\n\n", status.Name)
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
