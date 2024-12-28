package main

import (
	"fmt"
	"os"

	"github.com/andrejsstepanovs/andai/cmd/setup"
	"github.com/andrejsstepanovs/andai/pkg/deps"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func main() {
	initConfig()
	//cobra.OnInitialize(initConfig)

	deps := deps.NewAppDependencies()

	rootCmd := &cobra.Command{
		Use:   "main",
		Short: "A simple CLI application",
	}

	rootCmd.AddCommand(
		setup.SetupPingCmd(deps),
		setup.SetupUpdateCmd(),
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

func CheckRequiredFlags(prefixKey string, requiredFlags []string, cmd *cobra.Command) {
	unsetFlags := make([]string, 0, len(requiredFlags))
	for _, f := range requiredFlags {
		if !viper.GetViper().IsSet(prefixKey + f) {
			unsetFlags = append(unsetFlags, f)
		}
	}

	if len(unsetFlags) > 0 {
		fmt.Fprintln(os.Stderr, "Error: required flags are not set:")
		for _, f := range unsetFlags {
			fmt.Fprintf(os.Stderr, " --%s\n", f)
		}
		cmd.Usage()
		os.Exit(1)
	}
}
