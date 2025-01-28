package models

const (
	// TriggerTransitionWhoSelf is used to indicate that the transition should be applied to parent issues.
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
	Who       string        `yaml:"who"`
	IssueType IssueTypeName `yaml:"issue_type"`
	To        StateName     `yaml:"to"`
}
