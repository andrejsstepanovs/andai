package settings

import (
	"fmt"
)

// LlmModelNormal is the name of the normal model
const LlmModelNormal = "normal"

type LlmModels []LlmModel

type LlmModel struct {
	Name        string  `yaml:"name"`
	Model       string  `yaml:"model"`
	Provider    string  `yaml:"provider"`
	APIKey      string  `yaml:"api_key"`
	Temperature float64 `yaml:"temperature"`
	BaseURL     string  `yaml:"base_url"`
	MaxTokens   int     `yaml:"max_tokens"`
	MaxRetries  int     `yaml:"max_retries"`
}

func (m LlmModels) Get(name string) LlmModel {
	for _, model := range m {
		if model.Name == name {
			return model
		}
	}
	panic(fmt.Sprintf("model %s not found", name))
}
