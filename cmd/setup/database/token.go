package database

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewGetTokenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Set (or get) redmine admin token",
		RunE:  runGetAdminToken,
	}
	return cmd
}

func runGetAdminToken(cmd *cobra.Command, args []string) error {
	fmt.Println("Get redmine admin token or creates it if missing")

	db, err := sql.Open("mysql", viper.GetString("redmine.db"))
	if err != nil {
		panic(err)
	}

	const ADMIN_USER_ID = 1
	results, err := db.Query("SELECT id, value FROM tokens WHERE action = ? AND user_id = ?", "api", ADMIN_USER_ID)
	if err != nil {
		fmt.Println("Redmine Admin Fail")
		return fmt.Errorf("error redmine admin: %v", err)
	}
	defer results.Close()

	type Row struct {
		ID    int
		Value string
	}
	var rows []Row

	// Loop through rows, using Scan to assign column data to struct fields.
	for results.Next() {
		var row Row
		if err := results.Scan(&row.ID, &row.Value); err != nil {
			return err
		}
		rows = append(rows, row)
	}
	if err = results.Err(); err != nil {
		return err
	}

	for _, row := range rows {
		fmt.Println("Token:", row.Value)
	}
	if len(rows) > 0 {
		fmt.Println("Token already exists")
		return nil
	}

	result, err := db.Exec("INSERT INTO tokens (value, action, user_id, created_on, updated_on) VALUES (?, ?, ?, NOW(), NOW())", viper.GetString("redmine.api_key"), "api", ADMIN_USER_ID)
	if err != nil {
		return fmt.Errorf("insert settings token db err: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}

	if affected > 0 {
		fmt.Println("Admin token set")
	} else {
		return errors.New("Admin token not created")
	}

	return nil
}
