package models

type Priorities []Priority

type Priority struct {
	Type  IssueTypeName `yaml:"type"`
	State StateName     `yaml:"state"`
}
