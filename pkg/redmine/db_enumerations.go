package redmine

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

const (
	queryGetIssuePriority    = "SELECT id FROM enumerations WHERE type = 'IssuePriority' AND is_default = 1 AND name = ?"
	queryInsertIssuePriority = "INSERT INTO enumerations (name, position, is_default, type, active, project_id, parent_id, position_name) VALUES (?, 1, 1, 'IssuePriority', 1, NULL, NULL, 'default')"
)

const (
	IssuePriority = "Normal"
)

func (c *Model) GetDefaultNormalPriority() (int, error) {
	var ids []int
	err := c.queryAndScan(queryGetIssuePriority, func(rows *sql.Rows) error {
		var row int
		if err := rows.Scan(&row); err != nil {
			return err
		}
		ids = append(ids, row)
		return nil
	}, IssuePriority)

	if err != nil && !errors.As(err, &sql.ErrNoRows) {
		return 0, err
	}
	for _, id := range ids {
		return id, nil
	}
	return 0, nil
}

func (c *Model) DBInsertDefaultNormalPriority() error {
	result, err := c.execDML(queryInsertIssuePriority, IssuePriority)
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
