package database

import (
	"fmt"

	"github.com/andrejsstepanovs/andai/pkg/redmine"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
)

func NewPingCommand(redmine *redmine.Model) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "database",
		Short: "Ping database connection",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Processing Jira issue", len(args))

			users, err := redmine.DbGetAllUsers()
			if err != nil {
				fmt.Println("Failed to load users")
				return fmt.Errorf("error getting users: %v", err)
			}

			fmt.Println("Users from Database")
			for _, user := range users {
				fmt.Printf("ID: %d, Login: %q Lastname: %q\n", user.Id, user.Login, user.Lastname)
			}
			fmt.Println("Database Ping Success")

			return nil
		},
	}
	return cmd
}
