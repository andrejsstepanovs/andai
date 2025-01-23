package validate

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/pkg/models"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newValidateCommand(settings models.Settings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Validates project config file",
		RunE: func(_ *cobra.Command, _ []string) error {
			err := settings.Validate()
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
