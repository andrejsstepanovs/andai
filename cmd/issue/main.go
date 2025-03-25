package issue

import (
	"github.com/spf13/cobra"

	"github.com/andrejsstepanovs/andai/internal/deps"
)

func Cmd(deps *deps.AppDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Manages issues",
	}

	cmd.AddCommand(
		newCreateCommand(deps),
		newMoveCommand(deps),
		newMoveChildrenCommand(deps),
	)

	return cmd
}
