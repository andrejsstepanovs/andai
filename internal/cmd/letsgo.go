package cmd

import (
	"github.com/andrejsstepanovs/andai/internal/cmd/setup"
	"github.com/andrejsstepanovs/andai/internal/cmd/work"
	"github.com/spf13/cobra"

	"github.com/andrejsstepanovs/andai/internal"
)

func LetsGo(deps *internal.AppDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lets",
		Short: "Setup and Run the workflow loop",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use: "go",
			RunE: func(_ *cobra.Command, _ []string) error {
				settings, err := deps.Config.Load()
				if err != nil {
					return err
				}

				err = setup.Setup(deps.Model, settings.Projects, settings.Workflow)
				if err != nil {
					return err
				}

				return work.Loop(deps)
			},
		},
	)

	return cmd
}
