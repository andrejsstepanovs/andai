package redmine

import (
	"database/sql"
	"errors"

	"github.com/andrejsstepanovs/andai/pkg/redmine/models"
	_ "github.com/go-sql-driver/mysql"
)

const (
	queryGetJournalComments = "SELECT notes, user_id, created_on FROM journals WHERE journalized_type = ? AND notes != ? AND journalized_id = ? ORDER BY created_on ASC"
)

const JournalIssueType = "Issue"

func (c *Model) DBGetComments(issueId int) (models.Comments, error) {
	var notes []models.Comment
	var i = 0
	err := c.queryAndScan(queryGetJournalComments, func(rows *sql.Rows) error {
		var row models.Comment
		row.Number = i
		if err := rows.Scan(&row.Text, &row.UserID, &row.CreatedAt); err != nil {
			return err
		}
		i++
		notes = append(notes, row)
		return nil
	}, JournalIssueType, "", issueId)

	if err != nil && !errors.As(err, &sql.ErrNoRows) {
		return nil, err
	}

	var comments models.Comments
	comments = notes
	return comments, nil
}
