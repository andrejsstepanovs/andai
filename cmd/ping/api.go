package ping

import (
	"fmt"

	"github.com/andrejsstepanovs/andai/pkg/redmine"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newAPIPingCommand(redmine *redmine.Model) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Ping connection to redmine",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println("Processing Jira issue")

			users, err := redmine.ApiGetUsers()
			if err != nil {
				fmt.Println("Redmine Ping Fail")
				return fmt.Errorf("error redmine ping: %v", err)
			}
			for _, user := range users {
				fmt.Printf("Identifier: %d, Name: %s\n", user.Id, user.Login)
			}
			fmt.Println("Redmine Ping Success")
			return nil
		},
	}
	return cmd
}
