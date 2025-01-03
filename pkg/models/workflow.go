package models

import (
	"fmt"

	"github.com/mattn/go-redmine"
)

type Settings struct {
	Workflow Workflow `yaml:"workflow"`
	Projects Projects `yaml:"projects"`
}

type StateName string

type IssueTypeName string

type States map[StateName]State

func (s *States) GetFirst() State {
	for _, state := range *s {
		if state.IsFirst {
			return state
		}
	}
	return State{}
}

type IssueTypes map[IssueTypeName]IssueType

type Transitions []Transition

type Transition struct {
	Source StateName `yaml:"source"`
	Target StateName `yaml:"target"`
}

func (t *Transition) GetIDs(statuses []redmine.IssueStatus) (from int, to int) {
	for _, status := range statuses {
		if string(t.Source) == status.Name {
			from = status.Id
		}
		if string(t.Target) == status.Name {
			to = status.Id
		}
		if from != 0 && to != 0 {
			return
		}
	}
	return
}

// Workflow represents
type Workflow struct {
	States      States      `yaml:"states"`
	IssueTypes  IssueTypes  `yaml:"issue_types"`
	Transitions Transitions `yaml:"transitions"`
}

type Job struct {
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Aider  string `yaml:"aider"`
	Prompt string `yaml:"prompt"`
}

type IssueType struct {
	Name        IssueTypeName     `yaml:"-"` // Exclude from YAML unmarshalling
	Jobs        map[StateName]Job `yaml:"jobs"`
	Description string            `yaml:"description"`
}

// State represents a single state within the workflow.
type State struct {
	Name        StateName `yaml:"-"` // Exclude from YAML unmarshalling
	Description string    `yaml:"description"`
	AI          bool      `yaml:"ai"`
	Prompt      string    `yaml:"prompt,omitempty"`
	IsDefault   bool      `yaml:"is_default"`
	IsFirst     bool      `yaml:"is_first"`
	IsClosed    bool      `yaml:"is_closed"`
}

// Get find State by name
func (s *States) Get(name StateName) State {
	return (*s)[name]
}

// UnmarshalYAML implements custom unmarshalling for the Workflow struct.
func (w *Workflow) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawWorkflow struct {
		States      map[StateName]State         `yaml:"states"`
		IssueTypes  map[IssueTypeName]IssueType `yaml:"issue_types"`
		Transitions Transitions                 `yaml:"transitions"`
	}

	var raw rawWorkflow
	if err := unmarshal(&raw); err != nil {
		return err
	}

	w.Transitions = raw.Transitions

	// map States
	cleanStates := make(map[StateName]State)
	for name, state := range raw.States {
		if name == "" {
			continue
		}
		state.Name = name
		cleanStates[name] = state
	}
	w.States = cleanStates

	// map IssueTypes
	cleanIssueTypes := make(map[IssueTypeName]IssueType)
	for name, issueType := range raw.IssueTypes {
		if name == "" {
			continue
		}
		issueType.Name = name
		cleanIssueTypes[name] = issueType
	}
	w.IssueTypes = cleanIssueTypes

	return nil
}

func (s *Settings) Validate() error {
	if len(s.Workflow.States) == 0 {
		return fmt.Errorf("workflow states are required")
	}

	if s.Workflow.States.GetFirst().Name == "" {
		return fmt.Errorf("at least one state must be marked as is_first")
	}

	stateNames := make(map[StateName]bool)

	defaultExists := false
	closedExists := false
	firstExists := false
	for _, state := range s.Workflow.States {
		stateNames[state.Name] = true
		if state.IsFirst {
			firstExists = true
		}
		if state.IsDefault {
			defaultExists = true
		}
		if state.IsClosed {
			closedExists = true
		}
	}

	if !defaultExists {
		return fmt.Errorf("at least one state must be marked as default")
	}
	if !firstExists {
		return fmt.Errorf("at least one state must be marked as is_first")
	}
	if !closedExists {
		return fmt.Errorf("at least one state must be marked as is_closed")
	}

	for issueTypeName, issueType := range s.Workflow.IssueTypes {
		for stateName := range issueType.Jobs {
			if _, ok := stateNames[stateName]; !ok {
				return fmt.Errorf("job %s does not have valid state %s", issueTypeName, stateName)
			}
		}
	}

	// validate transitions
	for _, transition := range s.Workflow.Transitions {
		if _, ok := stateNames[transition.Source]; !ok {
			return fmt.Errorf("transition source %s does not exist", transition.Source)
		}
		if _, ok := stateNames[transition.Target]; !ok {
			return fmt.Errorf("transition target %s does not exist", transition.Target)
		}
	}

	return nil
}
