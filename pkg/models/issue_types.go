package models

type IssueTypeName string

type IssueTypes map[IssueTypeName]IssueType

type IssueType struct {
	Name        IssueTypeName `yaml:"-"` // Exclude from YAML unmarshalling
	Jobs        Jobs          `yaml:"jobs"`
	Description string        `yaml:"description"`
}

type Jobs map[StateName]Job

type Steps []Step

type Job struct {
	Steps Steps `yaml:"steps"`
}

type Step struct {
	Command string `yaml:"command"`
	Args    string `yaml:"args"`
}

func (j *Jobs) Get(name StateName) Job {
	return (*j)[name]
}

func (s *IssueTypes) Get(name IssueTypeName) IssueType {
	return (*s)[name]
}
