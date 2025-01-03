package work

import (
	"github.com/andrejsstepanovs/andai/pkg/deps"
	"github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/spf13/cobra"
)

func Cmd(deps *deps.AppDependencies, settings models.Settings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "work",
		Short: "Changes environment setup",
	}

	cmd.AddCommand(
		newWorkCommand(deps.Model, settings.LlmModels),
	)

	return cmd
}
