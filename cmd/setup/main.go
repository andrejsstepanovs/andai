package setup

import (
	"github.com/andrejsstepanovs/andai/pkg/deps"
	"github.com/spf13/cobra"
)

func SetupUpdateCmd(deps *deps.AppDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Changes environment setup",
	}

	cmd.AddCommand(
		newAdminCommand(deps.Model),
		newSettingsCommand(deps.Model),
		newGetTokenCommand(deps.Model),
		newProjectCommand(deps.Model),
	)

	return cmd
}
