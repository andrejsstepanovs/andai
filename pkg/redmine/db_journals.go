package redmine

import (
	"database/sql"
	"errors"

	_ "github.com/go-sql-driver/mysql"
)

const (
	queryGetJournalComments = "SELECT notes FROM journals WHERE journalized_type = ? AND notes != ? AND journalized_id = ? ORDER BY created_on ASC"
)

const JournalIssueType = "Issue"

func (c *Model) DBGetComments(issueId int) ([]string, error) {
	var notes []string
	err := c.queryAndScan(queryGetJournalComments, func(rows *sql.Rows) error {
		var row string
		if err := rows.Scan(&row); err != nil {
			return err
		}
		notes = append(notes, row)
		return nil
	}, JournalIssueType, "", issueId)

	if err != nil && !errors.As(err, &sql.ErrNoRows) {
		return nil, err
	}

	return notes, nil
}
