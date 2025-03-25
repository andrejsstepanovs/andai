package main

import (
	"log"
	"os"

	"github.com/andrejsstepanovs/andai/cmd"
	"github.com/andrejsstepanovs/andai/cmd/issue"
	"github.com/andrejsstepanovs/andai/cmd/nothing"
	"github.com/andrejsstepanovs/andai/cmd/ping"
	"github.com/andrejsstepanovs/andai/cmd/setup"
	"github.com/andrejsstepanovs/andai/cmd/validate"
	"github.com/andrejsstepanovs/andai/cmd/work"
	"github.com/andrejsstepanovs/andai/internal"
	"github.com/andrejsstepanovs/andai/internal/config"
	"github.com/spf13/cobra"
)

func main() {
	project := os.Getenv("PROJECT")

	projectConfig := config.NewConfig(project, ".")
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
