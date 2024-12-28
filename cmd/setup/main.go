package setup

import (
	"github.com/andrejsstepanovs/andai/cmd/setup/api"
	"github.com/andrejsstepanovs/andai/cmd/setup/database"
	"github.com/spf13/cobra"
)

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
