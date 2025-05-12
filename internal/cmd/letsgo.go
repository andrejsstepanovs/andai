package cmd

import (
	"context"

	"github.com/andrejsstepanovs/andai/internal/cmd/setup"
	"github.com/andrejsstepanovs/andai/internal/cmd/work"
	"github.com/spf13/cobra"

	"github.com/andrejsstepanovs/andai/internal"
)

func LetsGo(deps internal.DependenciesLoader) *cobra.Command {
	var project string
	cmd := &cobra.Command{
		Use:   "go",
		Short: "Setup and Run the workflow loop. [OPTIONAL...] --project <identifier>",
		RunE: func(_ *cobra.Command, _ []string) error {
			d := deps()
			settings, err := d.Config.Load()
			if err != nil {
				return err
			}

			err = setup.Setup(d.Model, settings.Projects, settings.Workflow)
			if err != nil {
				return err
			}
			ctx := context.Background()

			return work.Loop(ctx, deps, project)
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "Project identifier (optional)")

	return cmd
}
