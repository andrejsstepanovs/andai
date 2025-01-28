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
		newWorkCommand(deps.Model, deps.LLM, settings.LlmModels),
		newNextCommand(deps.Model, deps.LLM, settings.Projects, settings.Workflow),
		newTriggersCommand(deps.Model, settings.Workflow),
	)

	return cmd
}
