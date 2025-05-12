package validate

import (
	"github.com/andrejsstepanovs/andai/internal"
	"github.com/spf13/cobra"
)

func SetupValidateCmd(deps internal.DependenciesLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate config files for project",
	}

	cmd.AddCommand(
		newValidateCommand(deps),
	)

	return cmd
}
