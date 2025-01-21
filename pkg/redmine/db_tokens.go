package redmine

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/pkg/redmine/models"
	_ "github.com/go-sql-driver/mysql" // mysql driver
)

const (
	queryUpdateTokens = "UPDATE tokens SET value = ?, updated_on = NOW() WHERE action = ? AND user_id = ?"                   // nolint:gosec
	queryInsertTokens = "INSERT INTO tokens (value, action, user_id, created_on, updated_on) VALUES (?, ?, ?, NOW(), NOW())" // nolint:gosec
	queryGetToken     = "SELECT id, action, value FROM tokens WHERE action = ? AND user_id = ?"                              // nolint:gosec
)

func (c *Model) DBUpdateAPIToken(userID int, tokenValue string) error {
	result, err := c.execDML(queryUpdateTokens, tokenValue, TokenActionAPI, userID)
	if err != nil {
		return fmt.Errorf("failed to update API token for user %d: %w", userID, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("no rows affected for user %d", userID)
	}
	log.Printf("API token updated for user %d\n", userID)
	return nil
}

func (c *Model) DBCreateAPIToken(userID int, tokenValue string) error {
	result, err := c.execDML(queryInsertTokens, tokenValue, TokenActionAPI, userID)
	if err != nil {
		return fmt.Errorf("insert settings token db err: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("token not created")
	}
	return nil
}

func (c *Model) DBGetToken(userID int) (models.Token, error) {
	var tokens []models.Token
	err := c.queryAndScan(queryGetToken, func(rows *sql.Rows) error {
		var token models.Token
		if err := rows.Scan(&token.ID, &token.Action, &token.Value); err != nil {
			return err
		}
		tokens = append(tokens, token)
		return nil
	}, TokenActionAPI, userID)

	if err != nil {
		return models.Token{}, err
	}

	if len(tokens) > 0 {
		return tokens[0], nil
	}
	return models.Token{}, nil
}
