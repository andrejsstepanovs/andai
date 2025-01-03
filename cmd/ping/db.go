package ping

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/pkg/redmine"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newDBPingCommand(redmine *redmine.Model) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Ping database connection",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println("Processing Jira issue")

			users, err := redmine.DbGetAllUsers()
			if err != nil {
				log.Println("Failed to load users")
				return fmt.Errorf("error getting users: %v", err)
			}

			log.Println("Users from Database")
			for _, user := range users {
				log.Printf("Identifier: %d, Login: %q Lastname: %q\n", user.Id, user.Login, user.Lastname)
			}
			log.Println("Database Ping Success")

			return nil
		},
	}
	return cmd
}
