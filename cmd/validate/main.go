package validate

import (
	"github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/spf13/cobra"
)

func SetupValidateCmd(settings models.Settings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate config files for project",
	}

	cmd.AddCommand(
		newValidateCommand(settings),
	)

	return cmd
}
