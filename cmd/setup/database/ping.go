package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewPingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "database",
		Short: "Ping database connection",
		RunE:  runDatabasePing,
	}
	return cmd
}

func runDatabasePing(cmd *cobra.Command, args []string) error {
	fmt.Println("Processing Jira issue", len(args))

	db, err := sql.Open("mysql", viper.GetString("redmine.db"))
	if err != nil {
		panic(err)
	}

	rows, err := db.Query("SELECT id, login FROM users")
	if err != nil {
		fmt.Println("Redmine Database Ping Fail")
		return fmt.Errorf("error database: %v", err)
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

	fmt.Println("Users from Database")
	for _, user := range users {
		fmt.Println(user.ID, user.Login)
	}
	fmt.Println("Database Ping Success")

	return nil
}
