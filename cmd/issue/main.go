package issue

import (
	"github.com/spf13/cobra"

	"github.com/andrejsstepanovs/andai/pkg/deps"
	"github.com/andrejsstepanovs/andai/pkg/models"
)

func Cmd(deps *deps.AppDependencies, settings models.Settings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Manages issues",
	}

	cmd.AddCommand(
		newCreateCommand(deps.Model, settings.Workflow),
		newMoveCommand(deps.Model, settings.Workflow),
		newMoveChildrenCommand(deps.Model, settings.Workflow),
	)

	return cmd
}
