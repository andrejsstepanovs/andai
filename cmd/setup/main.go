package setup

import (
	"github.com/spf13/cobra"

	"github.com/andrejsstepanovs/andai/pkg/deps"
	"github.com/andrejsstepanovs/andai/pkg/models"
)

func Cmd(deps *deps.AppDependencies, settings models.Settings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Changes environment setup",
	}

	cmd.AddCommand(
		newAdminCommand(deps.Model),
		newSettingsCommand(deps.Model),
		newGetTokenCommand(deps.Model),
		newProjectsCommand(deps.Model, settings.Projects),
		newWorkflowCommand(deps.Model, settings.Workflow),
		newIDAutoIncrementCommand(deps.Model),
	)

	return cmd
}
