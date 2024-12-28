package ping

import (
	"fmt"

	"github.com/andrejsstepanovs/andai/pkg/redmine"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
)

func newApiPingCommand(redmine *redmine.Model) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Ping connection to redmine",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Processing Jira issue", len(args))

			users, err := redmine.ApiGetUsers()
			if err != nil {
				fmt.Println("Redmine Ping Fail")
				return fmt.Errorf("error redmine ping: %v", err)
			}
			for _, user := range users {
				fmt.Println(fmt.Sprintf("ID: %d, Name: %s", user.Id, user.Login))
			}
			fmt.Println("Redmine Ping Success")
			return nil
		},
	}
	return cmd
}
