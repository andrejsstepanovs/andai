package setup

import (
	"github.com/andrejsstepanovs/andai/cmd/setup/api"
	"github.com/andrejsstepanovs/andai/cmd/setup/database"
	"github.com/andrejsstepanovs/andai/pkg/deps"
	"github.com/spf13/cobra"
)

func SetupPingCmd(deps *deps.AppDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping environment",
		Long:  `Check if all necessary connections work.`,
	}

	cmd.AddCommand(
		database.NewPingCommand(deps.Model),
		api.NewPingCommand(deps.Model),
	)

	return cmd
}

func SetupUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Changes environment setup",
	}

	cmd.AddCommand(
		database.NewAdminCommand(),
		database.NewSettingsCommand(),
		database.NewGetTokenCommand(),
		api.NewProjectCommand(),
	)

	return cmd
}
