package validate

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/internal/deps"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newValidateCommand(deps *deps.AppDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Validates project config file",
		RunE: func(_ *cobra.Command, _ []string) error {
			settings, err := deps.Config.Load()
			if err != nil {
				return err
			}

			err = settings.Validate()
			if err != nil {
				log.Println("Error validating settings:", err)
				return err
			}
			fmt.Println("Is valid")
			return nil
		},
	}
	return cmd
}
