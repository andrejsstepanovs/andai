package models

type Settings struct {
	Workflow Workflow `yaml:"workflow"`
}

type States map[string]State

// Workflow represents the entire workflow structure.
type Workflow struct {
	States States `yaml:"states"`
}

// State represents a single state within the workflow.
type State struct {
	Name        string `yaml:"-"` // Exclude from YAML unmarshaling
	Description string `yaml:"description"`
	AI          bool   `yaml:"ai"`
	Prompt      string `yaml:"prompt,omitempty"`
	IsDefault   bool   `yaml:"is_default"`
	IsClosed    bool   `yaml:"is_closed"`
}

// Get find State by name
func (s *States) Get(name string) State {
	return (*s)[name]
}

// UnmarshalYAML implements custom unmarshaling for the Workflow struct.
func (w *Workflow) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawWorkflow struct {
		States map[string]State `yaml:"states"`
	}

	var raw rawWorkflow
	if err := unmarshal(&raw); err != nil {
		return err
	}

	cleanStates := make(map[string]State)
	for name, state := range raw.States {
		if name == "" {
			continue
		}
		state.Name = name
		cleanStates[name] = state
	}

	w.States = cleanStates

	return nil
}
