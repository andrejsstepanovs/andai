package models

import (
	"fmt"
)

type Settings struct {
	Workflow  Workflow  `yaml:"workflow"`
	Projects  Projects  `yaml:"projects"`
	LlmModels LlmModels `yaml:"llm_models"`
}

func (s *Settings) validateStates(issueTypeNames map[IssueTypeName]bool) error {
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

		for _, aiState := range state.UseAI {
			if _, ok := issueTypeNames[aiState]; !ok {
				return fmt.Errorf("%q state ai: %q is not a valid issue type", state.Name, aiState)
			}
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

func (s *Settings) validateSteps(issueTypeNames map[IssueTypeName]bool) error {
	for _, types := range s.Workflow.IssueTypes {
		for jobName, job := range types.Jobs {
			for _, step := range job.Steps {
				switch step.Command {
				case "git":
				case "next":
				case "create-issues":
				case "merge-into-parent":
				case "evaluate":
				case "ai":
				case "aid":
				case "aider":
				default:
					return fmt.Errorf("step command %q is not valid", step.Command)
				}

				if step.Command == "aider" || step.Command == "aid" {
					switch step.Action {
					case "commit":
					case "architect":
					case "code":
					default:
						return fmt.Errorf("%q step action %q is not valid for %q in %q", step.Command, step.Action, types.Name, jobName)
					}
				}

				if step.Command == "create-issues" {
					if _, ok := issueTypeNames[IssueTypeName(step.Action)]; !ok {
						return fmt.Errorf("%q step action %q is not a valid issue type for %q in %q", step.Command, step.Action, types.Name, jobName)
					}
				}
			}
		}
	}
	return nil
}

func (s *Settings) validateLlmModels() error {
	for _, model := range s.LlmModels {
		if model.Model == "" {
			return fmt.Errorf("llm model model is required")
		}
		if model.Provider == "" {
			return fmt.Errorf("llm model provider is required")
		}
		if model.APIKey == "" {
			return fmt.Errorf("llm model api_key is required")
		}
		switch model.Name {
		case LlmModelNormal:
		default:
			return fmt.Errorf("llm model %q is not valid", model.Name)
		}
	}
	return nil
}

func (s *Settings) validateTriggers(issueTypeNames map[IssueTypeName]bool, stateNames map[StateName]bool) error {
	for _, trigger := range s.Workflow.Triggers {
		if _, ok := issueTypeNames[trigger.IssueType]; !ok {
			return fmt.Errorf("trigger type %s does not exist", trigger.IssueType)
		}

		for _, triggerIf := range trigger.TriggerIf {
			if _, ok := stateNames[triggerIf.MovedTo]; !ok {
				return fmt.Errorf("trigger state %s does not exist", triggerIf.MovedTo)
			}

			if _, ok := stateNames[triggerIf.TriggerTransition.To]; !ok {
				return fmt.Errorf("trigger transition to state %s does not exist", triggerIf.TriggerTransition.To)
			}

			if triggerIf.AllSiblingsStatus != "" {
				if _, ok := stateNames[triggerIf.AllSiblingsStatus]; !ok {
					return fmt.Errorf("trigger all siblings status %s does not exist", triggerIf.AllSiblingsStatus)
				}
			}

			switch triggerIf.TriggerTransition.Who {
			case TriggerTransitionWhoParent:
			case TriggerTransitionWhoChildren:
				continue
			default:
				return fmt.Errorf("trigger transition 'who': %q is not valid", triggerIf.TriggerTransition.Who)
			}
		}
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

func (s *Settings) validateIssueTypeStates(stateNames map[StateName]bool) (map[IssueTypeName]bool, error) {
	issueTypeNames := make(map[IssueTypeName]bool)
	for issueTypeName, issueType := range s.Workflow.IssueTypes {
		issueTypeNames[issueTypeName] = true
		for stateName := range issueType.Jobs {
			if _, ok := stateNames[stateName]; !ok {
				return nil, fmt.Errorf("job %s does not have valid state %s", issueTypeName, stateName)
			}
		}
	}
	return issueTypeNames, nil
}

func (s *Settings) validateAISteps() error {
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
	return errs
}

func (s *Settings) validateStepContexts() error {
	for issueTypeName, issueType := range s.Workflow.IssueTypes {
		for stateName, job := range issueType.Jobs {
			for k, step := range job.Steps {
				for _, context := range step.Context {
					switch context {
					case ContextTicket:
					case ContextLastComment:
					case ContextComments:
					case ContextProject:
					case ContextProjectWiki:
					case ContextChildren:
					case ContextSiblings:
					case ContextParent:
					case ContextParents:
					case ContextIssueTypes:
					case ContextAffectedFiles:
					default:
						return fmt.Errorf("issue %q state %q job (%d) does not have valid context: %q", issueTypeName, stateName, k, context)
					}
				}
			}
		}
	}
	return nil
}

func (s *Settings) validateIssueTypes(stateNames map[StateName]bool) (map[IssueTypeName]bool, error) {
	issueStateNames, err := s.validateIssueTypeStates(stateNames)
	if err != nil {
		return nil, err
	}

	if err := s.validateAISteps(); err != nil {
		return nil, err
	}

	if err := s.validateStepContexts(); err != nil {
		return nil, err
	}

	return issueStateNames, nil
}

func (s *Settings) Validate() error {

	stateNames := make(map[StateName]bool)
	for _, state := range s.Workflow.States {
		stateNames[state.Name] = true
	}

	issueTypeNames, err := s.validateIssueTypes(stateNames)
	if err != nil {
		return err
	}

	if err := s.validateStates(issueTypeNames); err != nil {
		return err
	}

	if err := s.validateTransitions(stateNames); err != nil {
		return err
	}

	if err := s.validateTriggers(issueTypeNames, stateNames); err != nil {
		return err
	}

	// validate priorities
	for _, priority := range s.Workflow.Priorities {
		if _, ok := stateNames[priority.State]; !ok {
			return fmt.Errorf("priority state %s does not exist", priority.State)
		}
		if _, ok := issueTypeNames[priority.Type]; !ok {
			return fmt.Errorf("priority issue type %s does not exist", priority.Type)
		}
	}

	if err := s.validateLlmModels(); err != nil {
		return err
	}

	if err := s.validateSteps(issueTypeNames); err != nil {
		return err
	}

	return nil
}
