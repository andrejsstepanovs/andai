package setup

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/pkg/redmine"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newIDAutoIncrementCommand(redmine *redmine.Model) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auto-increments",
		Short: "Changes issue, project, user auto increment number so its easier to identify and work with.",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Update auto increment value")

			err := redmine.DBSetAutoIncrements()
			if err != nil {
				return fmt.Errorf("error redmine db: %v", err)
			}

			log.Printf("Auto increments OK\n")

			return nil
		},
	}
	return cmd
}
