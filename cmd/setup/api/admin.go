package api

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewAdminCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Admin setup for redmine (dont change password)",
		RunE:  runRedmineAdmin,
	}
	return cmd
}

func runRedmineAdmin(cmd *cobra.Command, args []string) error {
	fmt.Println("Processing Jira issue", len(args))

	db, err := sql.Open("mysql", viper.GetString("redmine.db"))
	if err != nil {
		panic(err)
	}

	rows, err := db.Query("SELECT id, login FROM users")
	if err != nil {
		fmt.Println("Redmine Admin Fail")
		return fmt.Errorf("error redmine admin: %v", err)
	}
	defer rows.Close()

	type User struct {
		ID    int
		Login string
	}
	var users []User

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Login); err != nil {
			return err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return err
	}

	for _, user := range users {
		fmt.Println(user.ID, user.Login)
	}

	return nil
}
