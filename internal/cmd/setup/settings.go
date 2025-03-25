package setup

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/internal"
	"github.com/andrejsstepanovs/andai/internal/redmine"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newSettingsCommand(deps *internal.AppDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Enable REST API in Redmine",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Update Redmine settings")
			return setupSettings(deps.Model)
		},
	}
	return cmd
}

func setupSettings(redmine *redmine.Model) error {
	err := redmine.DBSettingsEnableAPI()
	if err != nil {
		log.Println("Redmine Settings Failed to enable API")
		return fmt.Errorf("error redmine: %v", err)
	}

	err = redmine.DBCreateWorkerRole()
	if err != nil {
		log.Println("Role creation failed")
		return fmt.Errorf("error redmine: %v", err)
	}
	roleID, err := redmine.DBGetWorkerRole()
	if err != nil {
		log.Println("Role not found")
		return fmt.Errorf("error redmine: %v", err)
	}

	log.Printf("Worker Role OK. Identifier: %d\n", roleID)

	return nil
}
