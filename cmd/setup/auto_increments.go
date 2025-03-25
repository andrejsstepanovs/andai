package setup

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/internal"
	"github.com/andrejsstepanovs/andai/internal/redmine"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newIDAutoIncrementCommand(deps *internal.AppDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auto-increments",
		Short: "Changes issue, project, user auto increment number so it's easier to identify and work with in browser",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Update auto increment value")

			err := setupAutoIncrement(deps.Model)
			if err != nil {
				return err
			}
			log.Printf("Auto increments OK\n")

			return nil
		},
	}
	return cmd
}

func setupAutoIncrement(redmine *redmine.Model) error {
	err := redmine.DBSetAutoIncrements()
	if err != nil {
		return fmt.Errorf("error redmine db: %v", err)
	}
	return nil
}
