package models

type StateName string

type States map[StateName]State

func (s *States) GetFirst() State {
	for _, state := range *s {
		if state.IsFirst {
			return state
		}
	}
	return State{}
}

func (s *States) GetClosed() State {
	for _, state := range *s {
		if state.IsClosed {
			return state
		}
	}
	return State{}
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
