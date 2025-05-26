package settings

import (
	"fmt"
	"strings"
)

// ContextTicket matches the ticket
const ContextTicket = "ticket"

// ContextComments matches comments
const ContextComments = "comments"

// ContextParentComments matches the parent comments
const ContextParentComments = "parent-comments"

// ContextLastComment matches the last comment
const ContextLastComment = "last-comment"

// ContextTwoComment matches the last two comment
const ContextTwoComment = "last-2-comments"

// ContextThreeComment matches the last three comment
const ContextThreeComment = "last-3-comments"

// ContextFourComment matches the last four comment
const ContextFourComment = "last-4-comments"

// ContextFifeComment matches the last fife comment
const ContextFifeComment = "last-5-comments"

// ContextProject matches the project
const ContextProject = "project"

// ContextProjectWiki matches the project wiki
const ContextProjectWiki = "wiki"

// ContextChildren matches the children
const ContextChildren = "children"

// ContextSiblings matches the siblings
const ContextSiblings = "siblings"

// ContextParent matches the parent
const ContextParent = "parent"

// ContextParents matches the parents
const ContextParents = "parents"

// ContextIssueTypes explains what each issue represents
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

func (s *Step) String(prefixMessage string) string {
	if s.Action != "" {
		return fmt.Sprintf("%s: %s - %s", prefixMessage, s.Command, s.Action)
	}
	return fmt.Sprintf("%s: %s", prefixMessage, s.Command)
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
