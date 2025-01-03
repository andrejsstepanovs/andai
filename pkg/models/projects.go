package models

type Projects []Project

type Project struct {
	Identifier  string `yaml:"identifier"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	GitPath     string `yaml:"git_path"`
	Wiki        string `yaml:"wiki"`
}
