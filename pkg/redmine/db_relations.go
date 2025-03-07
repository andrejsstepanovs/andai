package redmine

import (
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // mysql driver
)

const (
	queryInsertIssueRelation = "INSERT INTO issue_relations (issue_from_id, issue_to_id, relation_type) VALUES (?, ?, ?)" // nolint:gosec
)

func (c *Model) DBCreateBlockedByIssueRelation(issueID, blockedByID int) error {
	result, err := c.execDML(queryInsertIssueRelation, blockedByID, issueID, RelationBlocks)
	if err != nil {
		return fmt.Errorf("insert issue blocked by relation db err: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("issue blocked by relation not created")
	}
	return nil
}
