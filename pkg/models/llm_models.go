package models

import (
	"fmt"
)

const LlmModelNormal = "normal"

type LlmModels []LlmModel

type LlmModel struct {
	Name        string `yaml:"name"`
	Model       string `yaml:"model"`
	Provider    string `yaml:"provider"`
	APIKey      string `yaml:"api_key"`
	Temperature string `yaml:"temperature"`
}

func (m LlmModels) Get(name string) LlmModel {
	for _, model := range m {
		if model.Name == name {
			return model
		}
	}
	panic(fmt.Sprintf("model %s not found", name))
}
