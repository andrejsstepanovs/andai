package models

import "strings"

// ContextTicket is a special context that matches the ticket
const ContextTicket = "ticket"

// ContextComments is a special context that matches comments
const ContextComments = "comments"

// ContextLastComment is a special context that matches the last comment
const ContextLastComment = "last-comment"

// ContextProject is a special context that matches the project
const ContextProject = "project"

// ContextChildren is a special context that matches the children
const ContextChildren = "children"

// ContextParent is a special context that matches the parent
const ContextParent = "parent"

// ContextParents is a special context that matches the parents
const ContextParents = "parents"

// ContextAll is a special context that matches all contexts
const ContextAll = "all"

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

type Contexts []string

type StepPrompt string

type Step struct {
	Command  string     `yaml:"command"`
	Action   string     `yaml:"action"`
	Comment  bool       `yaml:"comment"`
	Remember bool       `yaml:"remember"`
	Context  Contexts   `yaml:"context"`
	Prompt   StepPrompt `yaml:"prompt"`
}

func (j *Jobs) Get(name StateName) Job {
	return (*j)[name]
}

func (s *IssueTypes) Get(name IssueTypeName) IssueType {
	return (*s)[name]
}

func (p *StepPrompt) ForCli() string {
	prompt := string(*p)
	searchReplace := map[string]string{
		`"`:  `\"`,
		`''`: `\'`,
		"\n": ` `,
	}
	for search, replace := range searchReplace {
		prompt = strings.ReplaceAll(prompt, search, replace)
	}
	return prompt
}

func (c *Contexts) Has(name string) bool {
	for _, context := range *c {
		if context == ContextAll {
			return true
		}
		if context == name {
			return true
		}
	}
	return false
}
