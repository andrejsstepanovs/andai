package models

const (
	// TriggerTransitionWhoParent is used to indicate that the transition should be applied to the parent issue.
	TriggerTransitionWhoParent = "parent"
	// TriggerTransitionWhoChildren is used to indicate that the transition should be applied to children issues.
	TriggerTransitionWhoChildren = "children"
)

type Triggers []Trigger

type Trigger struct {
	IssueType IssueTypeName `yaml:"issue_type"`
	TriggerIf []TriggerIf   `yaml:"if"`
}

type TriggerIf struct {
	MovedTo           StateName         `yaml:"moved_to"`
	AllSiblingsStatus StateName         `yaml:"all_siblings_status"`
	TriggerTransition TriggerTransition `yaml:"transition"`
}

type TriggerTransition struct {
	Who string    `yaml:"who"`
	To  StateName `yaml:"to"`
}

func (t Triggers) GetTriggers(issueType IssueTypeName) []Trigger {
	triggers := make([]Trigger, 0)
	for _, trigger := range t {
		if trigger.IssueType == issueType {
			triggers = append(triggers, trigger)
		}
	}
	return triggers
}

func (t Trigger) GetTriggerIf(movedTo StateName) *TriggerIf {
	for _, triggerIf := range t.TriggerIf {
		if triggerIf.MovedTo == movedTo {
			return &triggerIf
		}
	}
	return nil
}

func (t TriggerIf) AllSiblingsCheck(siblingStatuses []StateName) bool {
	if t.AllSiblingsStatus == "" {
		return true
	}
	if len(siblingStatuses) == 0 {
		return true
	}
	for _, status := range siblingStatuses {
		if status != t.AllSiblingsStatus {
			return false
		}
	}
	return true
}
