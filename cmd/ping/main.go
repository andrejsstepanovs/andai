package ping

import (
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
		newPingAllCommand(deps),
		newAPIPingCommand(deps),
		newDBPingCommand(deps),
		newLLMPingCommand(deps),
		newPingAiderCommand(deps),
		newGitPingCommand(deps),
	)

	return cmd
}
