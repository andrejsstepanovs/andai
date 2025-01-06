package models

import (
	"github.com/mattn/go-redmine"
)

type Transitions []Transition

type Transition struct {
	Source StateName `yaml:"source"`
	Target StateName `yaml:"target"`
}

func (t *Transition) GetIDs(statuses []redmine.IssueStatus) (from int, to int) {
	for _, status := range statuses {
		if string(t.Source) == status.Name {
			from = status.Id
		}
		if string(t.Target) == status.Name {
			to = status.Id
		}
		if from != 0 && to != 0 {
			return
		}
	}
	return
}
