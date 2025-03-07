package redmine

import (
	"database/sql"
	"errors"

	_ "github.com/go-sql-driver/mysql" // mysql driver
)

const (
	queryGetClosedChildrenIDs = "SELECT A.id FROM issues A INNER JOIN issue_statuses B ON A.status_id = B.id AND B.is_closed = 1 WHERE A.parent_id = ?" // nolint:gosec
)

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
