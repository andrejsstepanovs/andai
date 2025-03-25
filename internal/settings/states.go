package settings

// StateName : Initial, Backlog, In Progress, etc
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

type UseAI []IssueTypeName

// State represents a single state within the workflow.
type State struct {
	Name        StateName `yaml:"-"` // Exclude from YAML unmarshalling
	Description string    `yaml:"description"`
	UseAI       UseAI     `yaml:"ai"`
	Prompt      string    `yaml:"prompt,omitempty"`
	IsDefault   bool      `yaml:"is_default"`
	IsFirst     bool      `yaml:"is_first"`
	IsClosed    bool      `yaml:"is_closed"`
}

func (s *States) Get(name StateName) State {
	return (*s)[name]
}

func (a *UseAI) Yes(name IssueTypeName) bool {
	for _, issueTypeName := range *a {
		if issueTypeName == name {
			return true
		}
	}
	return false
}
