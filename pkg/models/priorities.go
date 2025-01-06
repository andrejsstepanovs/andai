package models

type Priorities []Priority

type Priority struct {
	Type  string    `yaml:"type"`
	State StateName `yaml:"state"`
}
