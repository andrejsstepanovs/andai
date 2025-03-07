package models

type Workflow struct {
	ID          int
	TrackerID   int
	OldStatusID int
	NewStatusID int
	RoleID      int
	Assignee    int
	Author      int
	Type        string
	FieldName   *string
	Rule        *string
}
