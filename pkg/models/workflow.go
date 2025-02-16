package models

// Workflow represents
type Workflow struct {
	States      States      `yaml:"states"`
	IssueTypes  IssueTypes  `yaml:"issue_types"`
	Transitions Transitions `yaml:"transitions"`
	Priorities  Priorities  `yaml:"priorities"`
	Triggers    Triggers    `yaml:"triggers"`
	LlmModels   LlmModels   `yaml:"llm_models"`
	Aider       Aider       `yaml:"aider"`
}

// UnmarshalYAML implements custom unmarshalling for the Workflow struct.
func (w *Workflow) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawWorkflow struct {
		States      map[StateName]State         `yaml:"states"`
		IssueTypes  map[IssueTypeName]IssueType `yaml:"issue_types"`
		Transitions Transitions                 `yaml:"transitions"`
		Priorities  Priorities                  `yaml:"priorities"`
		Triggers    Triggers                    `yaml:"triggers"`
		LlmModels   LlmModels                   `yaml:"llm_models"`
		Aider       Aider                       `yaml:"aider"`
	}

	var raw rawWorkflow
	if err := unmarshal(&raw); err != nil {
		return err
	}

	w.Transitions = raw.Transitions
	w.Priorities = raw.Priorities
	w.Triggers = raw.Triggers
	w.LlmModels = raw.LlmModels
	w.Aider = raw.Aider

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
