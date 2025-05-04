package main

import (
	"log"
	"os"

	"github.com/andrejsstepanovs/andai/internal"
	"github.com/andrejsstepanovs/andai/internal/cmd"
	"github.com/andrejsstepanovs/andai/internal/cmd/issue"
	"github.com/andrejsstepanovs/andai/internal/cmd/nothing"
	"github.com/andrejsstepanovs/andai/internal/cmd/ping"
	"github.com/andrejsstepanovs/andai/internal/cmd/setup"
	"github.com/andrejsstepanovs/andai/internal/cmd/validate"
	"github.com/andrejsstepanovs/andai/internal/cmd/version"
	"github.com/andrejsstepanovs/andai/internal/cmd/work"
	"github.com/andrejsstepanovs/andai/internal/settings"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "main",
		Short: "A simple CLI application",
	}

	// DependenciesLoader use callback so we dont block the app if not used
	var dependenciesLoader internal.DependenciesLoader = func() *internal.AppDependencies {
		projectConfig := settings.NewConfig(".")
		_, err := projectConfig.Load()
		if err != nil {
			log.Println("Error loading config:", err)
			os.Exit(1)
		}

		dependencies, err := internal.NewAppDependencies(projectConfig)
		if err != nil {
			log.Println("Error creating dependencies:", err)
			os.Exit(1)
		}

		return dependencies
	}

	rootCmd.AddCommand(
		version.Cmd(),
		cmd.LetsGo(dependenciesLoader),
		nothing.Cmd(),
		validate.SetupValidateCmd(dependenciesLoader),
		ping.SetupPingCmd(dependenciesLoader),
		setup.Cmd(dependenciesLoader),
		work.Cmd(dependenciesLoader),
		issue.Cmd(dependenciesLoader),
	)

	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
