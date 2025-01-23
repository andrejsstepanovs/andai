package redmine

import (
	"database/sql"
	"errors"

	"github.com/andrejsstepanovs/andai/pkg/redmine/models"
	_ "github.com/go-sql-driver/mysql" // mysql driver
)

const (
	queryGetJournalComments = "SELECT notes, user_id, created_on FROM journals WHERE journalized_type = ? AND notes != ? AND journalized_id = ? ORDER BY created_on ASC"
)

// JournalIssueType is a constant for the journalized type
const JournalIssueType = "Issue"

func (c *Model) DBGetComments(issueID int) (models.Comments, error) {
	var notes []models.Comment
	var i = 1
	err := c.queryAndScan(queryGetJournalComments, func(rows *sql.Rows) error {
		var row models.Comment
		row.Number = i
		if err := rows.Scan(&row.Text, &row.UserID, &row.CreatedAt); err != nil {
			return err
		}
		i++
		notes = append(notes, row)
		return nil
	}, JournalIssueType, "", issueID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	var comments models.Comments = notes
	return comments, nil
}
