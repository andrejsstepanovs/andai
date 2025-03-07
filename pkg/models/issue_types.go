package models

import "strings"

// ContextTicket is a special context that matches the ticket
const ContextTicket = "ticket"

// ContextComments is a special context that matches comments
const ContextComments = "comments"

// ContextLastComment is a special context that matches the last comment
const ContextLastComment = "last-comment"

// ContextTwoComment is a special context that matches the last two comment
const ContextTwoComment = "last-2-comment"

// ContextThreeComment is a special context that matches the last three comment
const ContextThreeComment = "last-3-comment"

// ContextFourComment is a special context that matches the last four comment
const ContextFourComment = "last-4-comment"

// ContextFifeComment is a special context that matches the last fife comment
const ContextFifeComment = "last-5-comment"

// ContextProject is a special context that matches the project
const ContextProject = "project"

// ContextProjectWiki is a special context that matches the project wiki
const ContextProjectWiki = "wiki"

// ContextChildren is a special context that matches the children
const ContextChildren = "children"

// ContextSiblings is a special context that matches the siblings
const ContextSiblings = "siblings"

// ContextParent is a special context that matches the parent
const ContextParent = "parent"

// ContextParents is a special context that matches the parents
const ContextParents = "parents"

// ContextIssueTypes is a special context that explains what each issue represents
const ContextIssueTypes = "issue_types"

// ContextAffectedFiles provides all files that were touched in children and siblings branches.
const ContextAffectedFiles = "affected-files"

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
	Command        string     `yaml:"command"`
	Action         string     `yaml:"action"`
	Comment        bool       `yaml:"comment"`
	Remember       bool       `yaml:"remember"`
	Context        Contexts   `yaml:"context"`
	Prompt         StepPrompt `yaml:"prompt"`
	Summarize      bool       `yaml:"summarize"`
	CommentSummary bool       `yaml:"comment-summary"`
	History        []string
	ContextFiles   []string
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
	return strings.TrimSpace(prompt)
}

func (c *Contexts) Has(name string) bool {
	for _, context := range *c {
		if context == name {
			return true
		}
	}
	return false
}
