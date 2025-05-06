package settings

import (
	"fmt"
	"log"
)

// LlmModelNormal is the name of the normal model
const LlmModelNormal = "normal"

type LlmModels []LlmModel

type LlmModel struct {
	Name        string   `yaml:"name"`
	Model       string   `yaml:"model"`
	Provider    string   `yaml:"provider"`
	APIKey      string   `yaml:"api_key"`
	Temperature float64  `yaml:"temperature"`
	BaseURL     string   `yaml:"base_url"`
	MaxTokens   int      `yaml:"max_tokens"`
	MaxRetries  int      `yaml:"max_retries"`
	Commands    []string `yaml:"commands"`
}

func (m LlmModels) Get(name string) LlmModel {
	for _, model := range m {
		if model.Name == name {
			return model
		}
	}
	panic(fmt.Sprintf("model %q not found", name))
}

func (m LlmModels) ForCommand(name, command string) LlmModel {
	for _, model := range m {
		for _, cmd := range model.Commands {
			if cmd == command {
				log.Printf("Command %q found in model %q\n", command, model.Name)
				return model
			}
		}
	}

	return m.Get(name)
}
