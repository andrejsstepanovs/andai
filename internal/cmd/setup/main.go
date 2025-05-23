package setup

import (
	"github.com/spf13/cobra"

	"github.com/andrejsstepanovs/andai/internal"
)

func Cmd(deps internal.DependenciesLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Changes environment setup",
	}

	cmd.AddCommand(
		newSetupAllCommand(deps),
		newAdminCommand(deps),
		newSettingsCommand(deps),
		newGetTokenCommand(deps),
		newProjectsCommand(deps),
		newWorkflowCommand(deps),
		newCustomFieldsCommand(deps),
		newIDAutoIncrementCommand(deps),
	)

	return cmd
}
