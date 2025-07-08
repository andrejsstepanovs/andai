package redmine

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/mattn/go-redmine"
)

const (
	queryGetClosedChildrenIDs   = "SELECT A.id FROM issues A INNER JOIN issue_statuses B ON A.status_id = B.id AND B.is_closed = 1 WHERE A.parent_id = ?" // nolint:gosec
	queryGetChildren            = "SELECT A.id, A.subject, A.project_id, A.parent_id FROM issues A WHERE A.parent_id = ? ORDER BY A.id DESC"              // nolint:gosec
	queryInsertCustomFieldValue = "INSERT INTO custom_values (customized_type, customized_id, custom_field_id, value) VALUES ('Issue', ?, ?, ?)"          // nolint:gosec
	queryUpdateCustomFieldValue = "UPDATE custom_values SET value = ? WHERE id = ?"                                                                       // nolint:gosec
	querySelectCustomFieldValue = "SELECT id FROM custom_values WHERE customized_type = 'Issue' AND customized_id = ? AND custom_field_id = ?"            // nolint:gosec
)

func (c *Model) DBGetChildren(parent redmine.Issue) ([]redmine.Issue, error) {
	var children []redmine.Issue
	err := c.queryAndScan(queryGetChildren, func(rows *sql.Rows) error {
		var child redmine.Issue
		if err := rows.Scan(&child.Id, &child.Subject, &child.ProjectId, &child.ParentId); err != nil {
			return err
		}
		child.Project = &redmine.IdName{
			Id: child.ProjectId,
		}
		child.Parent = &redmine.Id{
			Id: child.ParentId,
		}

		children = append(children, child)
		return nil
	}, parent.Id)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return children, nil
}

// DBGetClosedChildrenIDs returns IDs of closed children of the parent issue.
func (c *Model) DBGetClosedChildrenIDs(parentID int) ([]int, error) {
	var ids []int
	err := c.queryAndScan(queryGetClosedChildrenIDs, func(rows *sql.Rows) error {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
		return nil
	}, parentID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return ids, nil
}

func (c *Model) DBInsertCustomFieldValue(issueID, customFieldID int, value string) error {
	_, err := c.execDML(queryInsertCustomFieldValue, issueID, customFieldID, value)
	if err != nil {
		return fmt.Errorf("insert custom field value err: %v", err)
	}
	return nil
}

func (c *Model) DBUpdateCustomFieldValue(customValueID int, value string) error {
	_, err := c.execDML(queryUpdateCustomFieldValue, value, customValueID)
	if err != nil {
		return fmt.Errorf("update custom field value db err: %v", err)
	}

	return nil
}

func (c *Model) DBFindCustomFieldValueID(issueID, customFieldID int) (int, error) {
	var id int
	err := c.queryAndScan(querySelectCustomFieldValue, func(rows *sql.Rows) error {
		if err := rows.Scan(&id); err != nil {
			return err
		}
		return nil
	}, issueID, customFieldID)
	if err != nil {
		return 0, fmt.Errorf("update custom field value db err: %v", err)
	}

	return id, nil
}
