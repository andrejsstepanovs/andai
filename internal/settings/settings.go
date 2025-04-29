package settings

import (
	"fmt"
	"strings"
)

type Settings struct {
	Workflow  Workflow  `yaml:"workflow"`
	Projects  Projects  `yaml:"projects"`
	LlmModels LlmModels `yaml:"llm_models"`
	Aider     Aider     `yaml:"aider"`
}

func (s *Settings) getAllIssueTypesAndStates() map[IssueTypeName]map[StateName]State {
	issueTypesAndStates := make(map[IssueTypeName]map[StateName]State)
	for issueTypeName := range s.Workflow.IssueTypes {
		states := make(map[StateName]State)
		for stateName, state := range s.Workflow.States {
			states[stateName] = state
		}
		issueTypesAndStates[issueTypeName] = states
	}
	return issueTypesAndStates
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
		for stateName, job := range types.Jobs {
			for _, step := range job.Steps {
				err := s.validateStep(step, issueTypeNames, stateName, types)
				if err != nil {
					return err
				}

				if step.Command == "next" && len(job.Steps) > 1 {
					return fmt.Errorf("step %q in %q in %q must be the only step (dont chain `next` with anything else)", step.Command, stateName, types.Name)
				}
			}
		}
	}
	return nil
}

// nolint: cyclop
func (s *Settings) validateStep(
	step Step,
	issueTypeNames map[IssueTypeName]bool,
	stateName StateName,
	types IssueType,
) error {
	switch step.Command {
	case "git":
	case "next":
	case "create-issues":
	case "merge-into-parent":
	case "project-cmd":
	case "summarize-task":
	case "commit": //nolint:goconst
	case "evaluate":
	case "ai":
	case "bash":
	case "context-files":
	case "aider":
	default:
		return fmt.Errorf("step command %q is not valid", step.Command)
	}

	if step.Command == "aider" {
		switch step.Action {
		case "commit":
		case "architect":
		case "code":
		case "architect-code":
		default:
			return fmt.Errorf("%q step action %q is not valid for %q in %q", step.Command, step.Action, types.Name, stateName)
		}
	} else {
		if step.Summarize {
			return fmt.Errorf("%q step %q in %q cannot have summarize (only `aider` can have `summarize`)", step.Command, step.Action, stateName)
		}
		if step.CommentSummary {
			return fmt.Errorf("%q step %q in %q cannot have summarize (only `aider` can have `comment-summary`)", step.Command, step.Action, stateName)
		}
	}

	if step.Command == "create-issues" {
		if _, ok := issueTypeNames[IssueTypeName(step.Action)]; !ok {
			return fmt.Errorf("%q step action %q is not a valid issue type for %q in %q", step.Command, step.Action, types.Name, stateName)
		}
	}

	if step.Command == "commit" {
		if step.Prompt == "" {
			return fmt.Errorf("%q step prompt is required for %q in %q", step.Command, types.Name, stateName)
		}
	}

	if step.Command == "project-cmd" {
		if step.Action == "" {
			return fmt.Errorf("%q step action is required for %q in %q", step.Command, types.Name, stateName)
		}
		if step.Prompt != "" {
			return fmt.Errorf("%q step %q in %q cannot have `prompt`", step.Command, step.Action, stateName)
		}
		if len(step.Context) > 0 {
			return fmt.Errorf("%q step %q in %q cannot have `context`", step.Command, step.Action, stateName)
		}
		if step.Summarize {
			return fmt.Errorf("%q step %q cannot have `summarize`", step.Command, step.Action)
		}
		for _, projectCfg := range s.Projects {
			found := false
			for _, cmd := range projectCfg.Commands {
				if cmd.Name == step.Action {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("%q step action %q missing for %q in %q in project %q", step.Command, step.Action, types.Name, stateName, projectCfg.Name)
			}
		}
	}

	if step.Command == "context-files" {
		if len(step.Context) == 0 {
			return fmt.Errorf("%q step %q must have at least one context file", step.Command, step.Action)
		}
		if !step.Remember {
			// this is just for the sake of reminding the user what the command is doing. It is not really used in code. Will be passed to history anyway.
			return fmt.Errorf("%q step %q must have remember set to true (mandatory)", step.Command, step.Action)
		}
	}

	if step.Command == "evaluate" {
		if len(step.Context) == 0 {
			return fmt.Errorf("%q step %q must have at least one context", step.Command, step.Action)
		}

		for _, state := range s.Workflow.States {
			if state.Name != stateName {
				continue
			}
			transitions := s.Workflow.Transitions.GetTransitions(state.Name)
			if len(transitions) <= 1 {
				return fmt.Errorf("command %q for %q in state %s must have more than one transition", step.Command, stateName, state.Name)
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
			if success != 1 && fail != 1 {
				return fmt.Errorf("command %q for %q in state %s must have at least one success and one fail transition", step.Command, stateName, state.Name)
			}
		}
	}

	return nil
}

func (s *Settings) validateProjects() error {
	if len(s.Projects) == 0 {
		return fmt.Errorf("projects are required")
	}
	uniqueIdentifiers := make(map[string]bool)
	for _, project := range s.Projects {
		if project.Name == "" {
			return fmt.Errorf("project name is required")
		}
		if project.Identifier == "" {
			return fmt.Errorf("project %q identifier is required", project.Name)
		}
		if strings.Contains(project.Identifier, " ") {
			return fmt.Errorf("project %q identifier cannot contain spaces", project.Name)
		}
		if strings.ToLower(project.Identifier) != project.Identifier {
			return fmt.Errorf("project %q identifier must be lowercase", project.Name)
		}
		if project.Wiki == "" {
			return fmt.Errorf("project %q wiki is required", project.Identifier)
		}
		if project.FinalBranch == "" {
			return fmt.Errorf("project %q final_branch is required", project.Identifier)
		}
		if project.LocalGitPath == "" {
			return fmt.Errorf("project %q git_local_dir is required. If you run in container use '/var/repositories/project/.git'", project.Identifier)
		}
		if project.GitPath == "" {
			return fmt.Errorf("project %q git_path is required. Try using: '/project/.git'", project.Identifier)
		}
		projCommands := make(map[string]bool)
		for _, cmd := range project.Commands {
			projCommands[cmd.Name] = true
			if cmd.Name == "" {
				return fmt.Errorf("project %q command name is required", project.Identifier)
			}
			if len(cmd.Command) == 0 {
				return fmt.Errorf("project %q command %q is missing command", cmd.Name, project.Identifier)
			}
		}
		if len(projCommands) != len(project.Commands) {
			return fmt.Errorf("project %q has duplicate commands", project.Identifier)
		}
		uniqueIdentifiers[project.Identifier] = true
	}

	if len(uniqueIdentifiers) != len(s.Projects) {
		return fmt.Errorf("projects have duplicate identifiers")
	}

	return nil
}

func (s *Settings) validateAider() error {
	//if s.Aider.MapTokens == 0 {
	//	return fmt.Errorf("aider map_tokens is required")
	//}
	if s.Aider.Config == "" {
		return fmt.Errorf("aider config is required")
	}
	if s.Aider.Timeout == 0 {
		return fmt.Errorf("aider timeout (duration) is required. Example: 5m")
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

func (s *Settings) validatePriorities(issueTypeNames map[IssueTypeName]bool, stateNames map[StateName]bool) error {
	for _, priority := range s.Workflow.Priorities {
		if _, ok := stateNames[priority.State]; !ok {
			return fmt.Errorf("priority state %s does not exist", priority.State)
		}
		if _, ok := issueTypeNames[priority.Type]; !ok {
			return fmt.Errorf("priority issue type %s does not exist", priority.Type)
		}
	}

	// check for duplicates
	priorityMap := make(map[string]bool)
	for _, priority := range s.Workflow.Priorities {
		key := fmt.Sprintf("%s-%s", priority.Type, priority.State)
		if _, ok := priorityMap[key]; ok {
			return fmt.Errorf("priority %q is duplicated", key)
		}
		priorityMap[key] = true
	}

	// check that all are defined
	for typeName, typeStates := range s.getAllIssueTypesAndStates() {
		for stateName, state := range typeStates {
			if state.IsClosed || state.IsDefault {
				continue
			}

			exists := false
			for _, priority := range s.Workflow.Priorities {
				if priority.Type == typeName && priority.State == stateName {
					exists = true
					break
				}
			}

			if !exists {
				return fmt.Errorf("issue type %q state %q does not have a priority defined", typeName, stateName)
			}
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

// nolint: cyclop
func (s *Settings) validateStepContexts() error {
	for issueTypeName, issueType := range s.Workflow.IssueTypes {
		for stateName, job := range issueType.Jobs {
			for k, step := range job.Steps {
				for _, context := range step.Context {
					switch context {
					case ContextTicket:
					case ContextLastComment:
					case ContextTwoComment:
					case ContextThreeComment:
					case ContextFourComment:
					case ContextFifeComment:
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
	for issueTypeName, issueType := range s.Workflow.IssueTypes {
		if len(issueType.Name) > 30 {
			return nil, fmt.Errorf("issue type %q name is too long (by %d) (max 30)", issueTypeName, len(issueType.Name)-30)
		}
		if len(issueType.Description) > 255 {
			return nil, fmt.Errorf("issue type %q description is too long (by %d) (max 255)", issueTypeName, len(issueType.Description)-255)
		}
	}

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

	if err := s.validatePriorities(issueTypeNames, stateNames); err != nil {
		return err
	}

	if err := s.validateLlmModels(); err != nil {
		return err
	}

	if err := s.validateSteps(issueTypeNames); err != nil {
		return err
	}

	if err := s.validateAider(); err != nil {
		return err
	}

	if err := s.validateProjects(); err != nil {
		return err
	}

	return nil
}
