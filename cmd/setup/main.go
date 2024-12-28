package setup

import (
	"github.com/andrejsstepanovs/andai/cmd/setup/redmine"
	"github.com/spf13/cobra"
)

func SetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup environment",
		Long:  `Setup all necessary parts of the project.`,
	}

	cmd.AddCommand(
		redmine.NewPingCommand(),
	)

	return cmd
}
