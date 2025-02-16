package work

import (
	"github.com/spf13/cobra"

	"github.com/andrejsstepanovs/andai/pkg/deps"
	"github.com/andrejsstepanovs/andai/pkg/models"
)

func Cmd(deps *deps.AppDependencies, settings models.Settings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "work",
		Short: "Changes environment setup",
	}

	cmd.AddCommand(
		newNextCommand(deps.Model, deps.LlmNorm, settings.Projects, settings.Workflow, settings.Aider),
		newTriggersCommand(deps.Model, settings.Workflow),
	)

	return cmd
}
