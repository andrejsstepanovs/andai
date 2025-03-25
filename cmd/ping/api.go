package ping

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/internal/deps"
	"github.com/andrejsstepanovs/andai/internal/redmine"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newAPIPingCommand(deps *deps.AppDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Ping connection to redmine",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Ping API")
			err := pingAPI(deps.Model)
			if err != nil {
				return err
			}
			log.Println("Redmine Ping Success")
			return nil
		},
	}
	return cmd
}

func pingAPI(redmine *redmine.Model) error {
	users, err := redmine.APIGetUsers()
	if err != nil {
		log.Println("Redmine Ping Fail")
		return fmt.Errorf("error redmine ping: %v", err)
	}
	for _, user := range users {
		log.Printf("Identifier: %d, Name: %s\n", user.Id, user.Login)
	}
	return nil
}
