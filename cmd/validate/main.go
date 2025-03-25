package validate

import (
	"github.com/andrejsstepanovs/andai/internal/deps"
	"github.com/spf13/cobra"
)

func SetupValidateCmd(deps *deps.AppDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate config files for project",
	}

	cmd.AddCommand(
		newValidateCommand(deps),
	)

	return cmd
}
