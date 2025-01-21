package models

import (
	"fmt"
)

type Settings struct {
	Workflow  Workflow  `yaml:"workflow"`
	Projects  Projects  `yaml:"projects"`
	LlmModels LlmModels `yaml:"llm_models"`
}

func (s *Settings) validateStates() error {
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

	return nil
}

func (s *Settings) validateTransitions(stateNames map[StateName]bool) error {
	// validate transitions existence
	for _, transition := range s.Workflow.Transitions {
		if _, ok := stateNames[transition.Source]; !ok {
			return fmt.Errorf("transition source %s does not exist", transition.Source)
		}
		if _, ok := stateNames[transition.Target]; !ok {
			return fmt.Errorf("transition target %s does not exist", transition.Target)
		}
	}

	// check for multiple transitions if there are no more than 1 Success or Fail transitions
	for _, state := range s.Workflow.States {
		transitions := s.Workflow.Transitions.GetTransitions(state.Name)
		if len(transitions) <= 1 {
			continue
		}
		success := 0
		fail := 0
		for _, transition := range transitions {
			if transition.Success {
				success++
			}
			if transition.Fail {
				fail++
			}
		}
		if success > 1 {
			return fmt.Errorf("state %s has more than one success transition", state.Name)
		}
		if fail > 1 {
			return fmt.Errorf("state %s has more than one fail transition", state.Name)
		}
	}
	return nil
}

func (s *Settings) validateIssueTypes(stateNames map[StateName]bool) (map[StateName]bool, error) {
	issueTypeNames := make(map[StateName]bool)
	for issueTypeName, issueType := range s.Workflow.IssueTypes {
		issueTypeNames[StateName(issueTypeName)] = true
		for stateName := range issueType.Jobs {
			if _, ok := stateNames[stateName]; !ok {
				return nil, fmt.Errorf("job %s does not have valid state %s", issueTypeName, stateName)
			}
		}
	}

	var errs error
	for _, state := range s.Workflow.States {
		for issueTypeName, issueType := range s.Workflow.IssueTypes {
			if !state.UseAI.Yes(issueTypeName) {
				continue
			}
			haveSteps := false
			for stateName, job := range issueType.Jobs {
				if state.Name != stateName {
					continue
				}
				haveSteps = len(job.Steps) > 0
			}
			if !haveSteps {
				if errs == nil {
					errs = fmt.Errorf("issue type %q does not have steps defined for state %q (in issue_types)", issueType.Name, state.Name)
				} else {
					errs = fmt.Errorf("%v\nissue type %q does not have steps defined for state %q (in issue_types)", errs, issueType.Name, state.Name)
				}
			}
		}
	}

	for issueTypeName, issueType := range s.Workflow.IssueTypes {
		for stateName, job := range issueType.Jobs {
			for k, step := range job.Steps {
				for _, context := range step.Context {
					switch context {
					case ContextTicket:
					case ContextAll:
					case ContextLastComment:
					case ContextComments:
					default:
						return nil, fmt.Errorf("issue %q state %q job (%d) does not have valid context %s", issueTypeName, stateName, k, context)
					}
				}
			}
		}
	}

	if errs != nil {
		return nil, errs
	}
	return issueTypeNames, nil
}

func (s *Settings) Validate() error {
	if err := s.validateStates(); err != nil {
		return err
	}

	stateNames := make(map[StateName]bool)
	for _, state := range s.Workflow.States {
		stateNames[state.Name] = true
	}

	issueTypeNames, err := s.validateIssueTypes(stateNames)
	if err != nil {
		return err
	}

	if err := s.validateTransitions(stateNames); err != nil {
		return err
	}

	// validate priorities
	for _, priority := range s.Workflow.Priorities {
		if _, ok := stateNames[priority.State]; !ok {
			return fmt.Errorf("priority state %s does not exist", priority.State)
		}
		if _, ok := issueTypeNames[StateName(priority.Type)]; !ok {
			return fmt.Errorf("priority issue type %s does not exist", priority.Type)
		}
	}

	return nil
}
