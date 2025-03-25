package ping

import (
	"github.com/andrejsstepanovs/andai/internal"
	"github.com/spf13/cobra"
)

func SetupPingCmd(deps *internal.AppDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping environment",
		Long:  `Check if all necessary connections work.`,
	}

	cmd.AddCommand(
		newPingAllCommand(deps),
		newAPIPingCommand(deps),
		newDBPingCommand(deps),
		newLLMPingCommand(deps),
		newPingAiderCommand(deps),
		newGitPingCommand(deps),
	)

	return cmd
}
