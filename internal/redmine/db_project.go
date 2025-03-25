package redmine

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql" // mysql driver
)

const (
	queryGetProjectIsPublic = "SELECT is_public FROM projects WHERE id = ?"
)

func (c *Model) DBProjectPublic(projectID int) (bool, error) {
	var isPublic bool
	err := c.queryAndScan(queryGetProjectIsPublic, func(rows *sql.Rows) error {
		if err := rows.Scan(&isPublic); err != nil {
			return err
		}
		return nil
	}, projectID)

	if err != nil {
		return false, err
	}
	return isPublic, nil
}
