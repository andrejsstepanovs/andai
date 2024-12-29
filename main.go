package main

import (
	"fmt"
	"os"

	"github.com/andrejsstepanovs/andai/cmd/ping"
	"github.com/andrejsstepanovs/andai/cmd/setup"
	"github.com/andrejsstepanovs/andai/pkg/deps"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	initConfig()
	//cobra.OnInitialize(initConfig)

	dependencies, err := deps.NewAppDependencies()
	if err != nil {
		fmt.Println("Error creating dependencies:", err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:   "main",
		Short: "A simple CLI application",
	}

	rootCmd.AddCommand(
		ping.SetupPingCmd(dependencies),
		setup.UpdateCmd(dependencies),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	//Step 1: Set the config file name and type
	viper.SetConfigName(".andai.yaml") // Name of the config file (without extension)
	viper.SetConfigType("yaml")        // Type of the config file

	// Step 2: Add search paths for the config file
	// First, look in the current directory
	viper.AddConfigPath(".")

	// Fallback to the user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
		return
	}
	viper.AddConfigPath(home)

	// Step 3: Read the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found in any of the specified paths
			fmt.Println("Config file not found in current directory or home directory")
		} else {
			// Config file was found but another error occurred
			fmt.Println("Error reading config file:", err)
		}
		return
	}
}
