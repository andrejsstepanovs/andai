package models

type IssueTypeName string

type IssueTypes map[IssueTypeName]IssueType

type IssueType struct {
	Name        IssueTypeName     `yaml:"-"` // Exclude from YAML unmarshalling
	Jobs        map[StateName]Job `yaml:"jobs"`
	Description string            `yaml:"description"`
}

type Job struct {
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Aider  string `yaml:"aider"`
	Prompt string `yaml:"prompt"`
}

func (s *IssueTypes) Get(name IssueTypeName) IssueType {
	return (*s)[name]
}
