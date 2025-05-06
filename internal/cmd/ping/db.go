package ping

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/internal"
	"github.com/andrejsstepanovs/andai/internal/redmine"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newDBPingCommand(deps internal.DependenciesLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Ping database connection",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Pinging Database")
			err := pingDB(deps().Model)
			if err != nil {
				return err
			}
			log.Println("Database Ping Success")
			return nil
		},
	}
	return cmd
}

func pingDB(redmine *redmine.Model) error {
	users, err := redmine.DBGetAllUsers()
	if err != nil {
		return fmt.Errorf("error getting users: %v", err)
	}

	for _, user := range users {
		log.Printf("Identifier: %d, Login: %q Lastname: %q\n", user.Id, user.Login, user.Lastname)
	}

	if len(users) == 0 {
		return fmt.Errorf("no users found")
	}
	return nil
}
