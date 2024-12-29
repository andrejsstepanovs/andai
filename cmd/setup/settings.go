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

			return nil
		},
	}
	return cmd
}
