package ping

import (
	"github.com/andrejsstepanovs/andai/pkg/deps"
	"github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/spf13/cobra"
)

func SetupPingCmd(deps *deps.AppDependencies, settings models.Settings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping environment",
		Long:  `Check if all necessary connections work.`,
	}

	cmd.AddCommand(
		newAPIPingCommand(deps.Model),
		newDBPingCommand(deps.Model),
		newLLMPingCommand(deps.LlmNorm),
		newPingAiderCommand(deps.Model, settings.Projects, settings.Aider),
	)

	return cmd
}
