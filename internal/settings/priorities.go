package settings

type Priorities []Priority

type Priority struct {
	Type  IssueTypeName `yaml:"type"`
	State StateName     `yaml:"state"`
}
