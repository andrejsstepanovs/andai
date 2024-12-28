package setup

import (
	"github.com/andrejsstepanovs/andai/cmd/setup/database"
	"github.com/andrejsstepanovs/andai/pkg/deps"
	"github.com/spf13/cobra"
)

func SetupUpdateCmd(deps *deps.AppDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Changes environment setup",
	}

	cmd.AddCommand(
		database.NewAdminCommand(deps.Model),
		database.NewSettingsCommand(deps.Model),
		database.NewGetTokenCommand(),
		//api.NewProjectCommand(),
	)

	return cmd
}
