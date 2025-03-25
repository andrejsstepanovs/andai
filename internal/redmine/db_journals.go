package redmine

import (
	"database/sql"
	"errors"

	"github.com/andrejsstepanovs/andai/internal/redmine/models"
	_ "github.com/go-sql-driver/mysql" // mysql driver
)

const (
	queryGetJournalComments  = "SELECT notes, user_id, created_on FROM journals WHERE journalized_type = ? AND notes != ? AND journalized_id = ? ORDER BY created_on ASC"                          // nolint:gosec
	queryGetLastStatusChange = "SELECT B.journalized_id, A.old_value, A.value FROM journal_details A INNER JOIN journals B ON A.id = B.id WHERE A.prop_key = ? ORDER BY B.updated_on DESC LIMIT 1" // nolint:gosec
)

// JournalIssueType is a constant for the journalized type
const JournalIssueType = "Issue"

// JournalStatusID is a constant for the status_id
const JournalStatusID = "status_id"

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

func (c *Model) DBGetLastStatusChange() (issueID int, statusIDFrom int, statusIDTo int, err error) {
	err = c.queryAndScan(queryGetLastStatusChange, func(rows *sql.Rows) error {
		if err := rows.Scan(&issueID, &statusIDFrom, &statusIDTo); err != nil {
			return err
		}
		return nil
	}, JournalStatusID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return
	}

	return
}
