package setup

import (
	"github.com/andrejsstepanovs/andai/pkg/deps"
	"github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/spf13/cobra"
)

func SetupCmd(deps *deps.AppDependencies, workflowConfig models.Workflow) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Changes environment setup",
	}

	admin := newAdminCommand(deps.Model)
	settings := newSettingsCommand(deps.Model)
	token := newGetTokenCommand(deps.Model)
	project := newProjectCommand(deps.Model)
	workflow := newWorkflowCommand(deps.Model, workflowConfig)

	cmd.AddCommand(
		admin,
		settings,
		token,
		project,
		workflow,
	)

	return cmd
}
