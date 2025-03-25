package work

import (
	"github.com/spf13/cobra"

	"github.com/andrejsstepanovs/andai/internal"
)

func Cmd(deps *internal.AppDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "work",
		Short: "Changes environment setup",
	}

	cmd.AddCommand(
		newNextCommand(deps),
		newTriggersCommand(deps),
		newLoopCommand(deps),
	)

	return cmd
}
