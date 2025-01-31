package processor

import (
	"fmt"
)

type AnswerIssues struct {
	ID          int    `json:"number_int"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
	BlockedBy   []int  `json:"blocked_by_numbers" validate:"omitempty"`
}

type Answer struct {
	Issues []AnswerIssues `json:"issues"`
}

func (a Answer) GetDeps() map[int][]int {
	deps := make(map[int][]int)
	for _, issue := range a.Issues {
		if deps[issue.ID] == nil {
			deps[issue.ID] = make([]int, 0)
		}
		deps[issue.ID] = append(deps[issue.ID], issue.BlockedBy...)
	}

	return deps
}

func (a Answer) Validate() error {
	if len(a.Issues) == 0 {
		return fmt.Errorf("no issues provided")
	}

	if err := a.ValidateNoSelfReference(); err != nil {
		return fmt.Errorf("dependent on self validation failed: %v", err)
	}

	if err := a.ValidateDependenciesExist(); err != nil {
		return fmt.Errorf("dependencies validation failed: %v", err)
	}

	if err := a.ValidateCircularDependency(); err != nil {
		return fmt.Errorf("circular dependency validation failed: %v", err)
	}

	return nil
}

func (a *Answer) ValidateNoSelfReference() error {
	for _, issue := range a.Issues {
		for _, blockedByID := range issue.BlockedBy {
			if blockedByID == issue.ID {
				return fmt.Errorf("issue %d has a self-reference in its BlockedBy field", issue.ID)
			}
		}
	}
	return nil
}

func (a Answer) ValidateDependenciesExist() error {
	existingIDs := make(map[int]bool)
	for _, issue := range a.Issues {
		existingIDs[issue.ID] = true
	}

	for _, issue := range a.Issues {
		for _, dependencyID := range issue.BlockedBy {
			if !existingIDs[dependencyID] {
				return fmt.Errorf("issue %d has dependency on non-existent issue %d",
					issue.ID, dependencyID)
			}
		}
	}

	return nil
}

func (a Answer) ValidateCircularDependency() error {
	existingIDs := make(map[int]bool)
	for _, issue := range a.Issues {
		existingIDs[issue.ID] = true
	}

	adj := make(map[int][]int)
	for _, issue := range a.Issues {
		id := issue.ID
		var deps []int
		for _, blockedID := range issue.BlockedBy {
			if existingIDs[blockedID] {
				deps = append(deps, blockedID)
			}
		}
		adj[id] = deps
	}

	visited := make(map[int]bool)
	recStack := make(map[int]bool)

	var detectCycle func(int) bool
	detectCycle = func(id int) bool {
		if recStack[id] {
			return true
		}
		if visited[id] {
			return false
		}
		visited[id] = true
		recStack[id] = true
		for _, neighbor := range adj[id] {
			if detectCycle(neighbor) {
				return true
			}
		}
		recStack[id] = false
		return false
	}

	for id := range adj {
		if !visited[id] {
			if detectCycle(id) {
				return fmt.Errorf("circular dependency detected involving issue %d", id)
			}
		}
	}

	return nil
}
