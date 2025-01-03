package models

import (
	"fmt"

	"github.com/mattn/go-redmine"
)

type Settings struct {
	Workflow  Workflow  `yaml:"workflow"`
	Projects  Projects  `yaml:"projects"`
	LlmModels LlmModels `yaml:"llm_models"`
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

type Priorities []Priority

type Priority struct {
	Type  string `yaml:"type"`
	State string `yaml:"state"`
}

type LlmModels []LlmModel

type LlmModel struct {
	Name     string `yaml:"name"`
	Model    string `yaml:"model"`
	Provider string `yaml:"provider"`
	APIKey   string `yaml:"api_key"`
}

func (m LlmModels) Get(name string) LlmModel {
	for _, model := range m {
		if model.Name == name {
			return model
		}
	}
	panic(fmt.Sprintf("model %s not found", name))
}

// Workflow represents
type Workflow struct {
	States      States      `yaml:"states"`
	IssueTypes  IssueTypes  `yaml:"issue_types"`
	Transitions Transitions `yaml:"transitions"`
	Priorities  Priorities  `yaml:"priorities"`
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
		Priorities  Priorities                  `yaml:"priorities"`
	}

	var raw rawWorkflow
	if err := unmarshal(&raw); err != nil {
		return err
	}

	w.Transitions = raw.Transitions
	w.Priorities = raw.Priorities

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

	issueTypeNames := make(map[StateName]bool)

	for issueTypeName, issueType := range s.Workflow.IssueTypes {
		issueTypeNames[StateName(issueTypeName)] = true
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

	// validate priorities
	for _, priority := range s.Workflow.Priorities {
		if _, ok := stateNames[StateName(priority.State)]; !ok {
			return fmt.Errorf("priority state %s does not exist", priority.State)
		}
		if _, ok := issueTypeNames[StateName(priority.Type)]; !ok {
			return fmt.Errorf("priority issue type %s does not exist", priority.Type)
		}
	}

	return nil
}
