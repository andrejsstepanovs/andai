package redmine

import (
	"database/sql"
	"errors"

	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/mattn/go-redmine"
)

const (
	queryGetClosedChildrenIDs = "SELECT A.id FROM issues A INNER JOIN issue_statuses B ON A.status_id = B.id AND B.is_closed = 1 WHERE A.parent_id = ?" // nolint:gosec
	queryGetChildren          = "SELECT A.id, A.subject, A.project_id, A.parent_id FROM issues A WHERE A.parent_id = ? ORDER BY A.id DESC"              // nolint:gosec
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
