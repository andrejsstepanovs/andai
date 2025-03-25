package models

import (
	"log"

	"github.com/mattn/go-redmine"
)

type Transitions []Transition
type NextTransition struct {
	Valid   bool
	Single  bool
	Success Transition
	Failure Transition
}

type Transition struct {
	Source  StateName `yaml:"source"`  // Initial, Backlog, In Progress, etc
	Target  StateName `yaml:"target"`  // Initial, Testing, etc
	Success bool      `yaml:"success"` // Transition on success if multiple transitions available
	Fail    bool      `yaml:"fail"`    // Transition on fail if multiple transitions available
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

func (n *NextTransition) LogPrint() {
	if !n.Valid {
		log.Println("No transitions available")
		return
	}
	if n.Single {
		log.Printf("Next transition: %s", n.Success.Target)
	} else {
		log.Printf("Next transition for Success: %s", n.Success.Target)
		log.Printf("Next transition for Failure: %s", n.Failure.Target)
	}
}

func (n *NextTransition) GetTarget(success bool) StateName {
	if success {
		return n.Success.Target
	}
	return n.Failure.Target
}

func (t *Transitions) GetNextTransition(source StateName) NextTransition {
	transitions := t.GetTransitions(source)
	if len(transitions) == 0 {
		return NextTransition{Valid: false}
	}
	next := NextTransition{
		Valid:   true,
		Single:  len(transitions) == 1,
		Success: transitions.SuccessTransition(),
		Failure: transitions.FailTransition(),
	}
	if next.Success.Target == next.Failure.Target {
		next.Single = true
	}
	return next
}

func (t *Transitions) GetTransitions(source StateName) (transitions Transitions) {
	for _, transition := range *t {
		if transition.Source == source {
			transitions = append(transitions, transition)
		}
	}
	return
}

func (t *Transitions) SuccessTransition() (transition Transition) {
	if len(*t) == 1 {
		return (*t)[0]
	}
	for _, t := range *t {
		if t.Success {
			return t
		}
	}
	return
}

func (t *Transitions) FailTransition() (transition Transition) {
	if len(*t) == 1 {
		return (*t)[0]
	}
	for _, t := range *t {
		if t.Fail {
			return t
		}
	}
	return
}
