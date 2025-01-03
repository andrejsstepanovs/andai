package setup

import (
	"fmt"

	"github.com/andrejsstepanovs/andai/pkg/redmine"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newSettingsCommand(redmine *redmine.Model) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Enable REST API in Redmine",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println("Update Redmine settings")

			err := redmine.DbSettingsEnableAPI()
			if err != nil {
				fmt.Println("Redmine Settings Failed to enable API")
				return fmt.Errorf("error redmine: %v", err)
			}

			roleID, err := redmine.DBGetWorkerRole()
			if err != nil {
				fmt.Println("Role not found")
				return fmt.Errorf("error redmine: %v", err)
			}
			if roleID == 0 {
				err = redmine.DBCreateWorkerRole()
				if err != nil {
					fmt.Println("Role creation failed")
					return fmt.Errorf("error redmine: %v", err)
				}
			}
			roleID, err = redmine.DBGetWorkerRole()
			if err != nil {
				fmt.Println("Role not found")
				return fmt.Errorf("error redmine: %v", err)
			}
			fmt.Printf("Worker Role OK. Identifier: %d\n", roleID)

			return nil
		},
	}
	return cmd
}
