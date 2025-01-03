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

func (c *Model) execDML(query string, args ...interface{}) (sql.Result, error) {
	result, err := c.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}
	return result, nil
}

func (c *Model) checkIfExists(query string, args ...interface{}) (bool, error) {
	rows, err := c.queryRows(query, args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

func (c *Model) queryAndScan(query string, scanFunc func(*sql.Rows) error, args ...interface{}) error {
	rows, err := c.queryRows(query, args...)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := scanFunc(rows); err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}
	}
	return rows.Err()
}
