package setup

import (
	"github.com/andrejsstepanovs/andai/cmd/setup/database"
	"github.com/andrejsstepanovs/andai/cmd/setup/redmine"
	"github.com/spf13/cobra"
)

func SetupPingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping environment",
		Long:  `Check if all necessary connections work.`,
	}

	cmd.AddCommand(
		database.NewPingCommand(),
		redmine.NewPingCommand(),
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
	)

	return cmd
}
