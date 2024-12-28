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

	admin := newAdminCommand(deps.Model)
	settings := newSettingsCommand(deps.Model)
	token := newGetTokenCommand(deps.Model)
	project := newProjectCommand(deps.Model)

	cmd.AddCommand(
		admin,
		settings,
		token,
		project,
	)

	return cmd
}
