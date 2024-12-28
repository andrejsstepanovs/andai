package database

import (
	"fmt"

	"github.com/andrejsstepanovs/andai/pkg/redmine"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
)

func NewSettingsCommand(redmine *redmine.Model) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Enable REST API in Redmine",
		RunE: func(cmd *cobra.Command, args []string) error {
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
