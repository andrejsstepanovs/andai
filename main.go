package main

import (
	"fmt"
	"os"

	"github.com/andrejsstepanovs/andai/cmd/ping"
	"github.com/andrejsstepanovs/andai/cmd/setup"
	"github.com/andrejsstepanovs/andai/pkg/deps"
	"github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func main() {
	configPath, err := initConfig()
	if err != nil {
		fmt.Println("Error initializing config:", err)
		os.Exit(1)
	}
	workflow, err := loadWorkflow(configPath)
	if err != nil {
		fmt.Println("Error initializing workflow:", err)
		os.Exit(1)
	}

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
		setup.SetupCmd(dependencies, workflow),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() (string, error) {
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
		return "", err
	}
	viper.AddConfigPath(home)

	// Step 3: Read the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found in current directory or home directory")
		} else {
			fmt.Println("Error reading config file:", err)
		}
		return "", err
	}

	filePath := viper.ConfigFileUsed()
	return filePath, nil
}

func loadWorkflow(filePath string) (models.Workflow, error) {
	fmt.Println("Using config file to load workflow:", filePath)
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return models.Workflow{}, err
	}

	var settings models.Settings
	err = yaml.Unmarshal(content, &settings)
	if err != nil {
		fmt.Println("Error unmarshaling YAML:", err)
		return models.Workflow{}, err
	}

	err = settings.Validate()
	if err != nil {
		fmt.Println("Error validating settings:", err)
		return models.Workflow{}, err
	}

	return settings.Workflow, nil
}
