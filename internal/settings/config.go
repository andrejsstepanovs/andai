package settings

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const configFilePrefix = ".andai."
const defaultProjectName = "project"

type Config struct {
	basePath string
	project  string
}

func NewConfig(project, basePath string) *Config {
	conf := &Config{
		basePath: basePath, // Default to current directory
	}

	if project == "" {
		log.Println("PROJECT environment variable not set")
		log.Println("Assuming PROJECT=project")
		project = defaultProjectName
	}

	conf.project = project

	return conf
}

func (c *Config) Load() (*Settings, error) {
	configFile, err := c.findConfigFile()
	if err != nil {
		log.Println("Error finding config file:", err)
		return &Settings{}, err
	}

	settings, err := c.getSettings(configFile)
	if err != nil {
		log.Println("Error getting settings:", err)
		return &Settings{}, err
	}

	err = settings.Validate()
	if err != nil {
		return &Settings{}, fmt.Errorf("settings validation err: %w", err)
	}

	return settings, nil
}

func (c *Config) findConfigFile() (string, error) {
	//Step 1: Set the config file name and type
	configFileBaseName := fmt.Sprintf("%s%s", configFilePrefix, c.project)

	viper.SetConfigName(configFileBaseName)
	viper.SetConfigType("yaml")

	// Step 2: Add search paths for the config file
	// First, look in the current directory
	viper.AddConfigPath(c.basePath)

	// Fallback to the user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		log.Println("Error getting user home directory:", err)
		return "", err
	}
	viper.AddConfigPath(home)

	// Step 3: Read the config file
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Printf("Config file not found in %q", c.basePath)
		}
		return "", err
	}

	return viper.ConfigFileUsed(), nil
}

func (c *Config) getSettings(configFile string) (*Settings, error) {
	log.Println("Using config file to load workflow:", configFile)
	content, err := os.ReadFile(configFile) // nolint:gosec
	if err != nil {
		log.Println("Error reading file:", err)
		return &Settings{}, err
	}

	var settings Settings
	err = yaml.Unmarshal(content, &settings)
	if err != nil {
		log.Printf("error unmarshaling YAML: %v\n", err)
		return &Settings{}, err
	}

	return &settings, nil
}
