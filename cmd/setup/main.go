package setup

import (
	"github.com/spf13/cobra"

	"github.com/andrejsstepanovs/andai/pkg/deps"
)

func Cmd(deps *deps.AppDependencies) *cobra.Command {
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
		newIDAutoIncrementCommand(deps),
	)

	return cmd
}
