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
	"github.com/andrejsstepanovs/andai/internal/cmd/work"
	"github.com/andrejsstepanovs/andai/internal/settings"
	"github.com/spf13/cobra"
)

func main() {
	project := os.Getenv("PROJECT")

	projectConfig := settings.NewConfig(project, ".")
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

	rootCmd := &cobra.Command{
		Use:   "main",
		Short: "A simple CLI application",
	}

	rootCmd.AddCommand(
		cmd.LetsGo(dependencies),
		nothing.Cmd(),
		validate.SetupValidateCmd(dependencies),
		ping.SetupPingCmd(dependencies),
		setup.Cmd(dependencies),
		work.Cmd(dependencies),
		issue.Cmd(dependencies),
	)

	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
