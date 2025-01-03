package redmine

import (
	"database/sql"
	"fmt"
)

func (c *Model) queryRows(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := c.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	return rows, nil
}
