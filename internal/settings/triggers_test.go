package settings_test

import (
	"testing"

	"github.com/andrejsstepanovs/andai/internal/settings"
	"github.com/stretchr/testify/assert"
)

func TestTriggers_GetTriggers(t *testing.T) {
	tests := []struct {
		name      string
		triggers  settings.Triggers
		issueType settings.IssueTypeName
		expected  []settings.Trigger
	}{
		{
			name:      "empty triggers",
			triggers:  settings.Triggers{},
			issueType: settings.IssueTypeName(""),
			expected:  []settings.Trigger{},
		},
		{
			name: "single trigger",
			triggers: settings.Triggers{
				{IssueType: settings.IssueTypeName("bug")},
			},
			issueType: settings.IssueTypeName("bug"),
			expected: []settings.Trigger{
				{IssueType: settings.IssueTypeName("bug")},
			},
		},
		{
			name: "multiple triggers",
			triggers: settings.Triggers{
				{IssueType: settings.IssueTypeName("bug")},
				{IssueType: settings.IssueTypeName("feature")},
			},
			issueType: settings.IssueTypeName("feature"),
			expected: []settings.Trigger{
				{IssueType: settings.IssueTypeName("feature")},
			},
		},
		{
			name: "no triggers for issue type",
			triggers: settings.Triggers{
				{IssueType: settings.IssueTypeName("bug")},
				{IssueType: settings.IssueTypeName("feature")},
			},
			issueType: settings.IssueTypeName("task"),
			expected:  []settings.Trigger{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			triggers := tt.triggers.GetTriggers(tt.issueType)
			assert.Equal(t, tt.expected, triggers)
		})
	}
}

func TestTrigger_GetTriggerIf(t *testing.T) {
	tests := []struct {
		name     string
		trigger  settings.Trigger
		movedTo  settings.StateName
		expected *settings.TriggerIf
	}{
		{
			name:     "empty trigger",
			trigger:  settings.Trigger{},
			movedTo:  settings.StateName(""),
			expected: nil,
		},
		{
			name: "single trigger if",
			trigger: settings.Trigger{
				TriggerIf: []settings.TriggerIf{
					{MovedTo: settings.StateName("in_progress")},
				},
			},
			movedTo: settings.StateName("in_progress"),
			expected: &settings.TriggerIf{
				MovedTo: settings.StateName("in_progress"),
			},
		},
		{
			name: "multiple trigger ifs",
			trigger: settings.Trigger{
				TriggerIf: []settings.TriggerIf{
					{MovedTo: settings.StateName("in_progress")},
					{MovedTo: settings.StateName("done")},
				},
			},
			movedTo: settings.StateName("done"),
			expected: &settings.TriggerIf{
				MovedTo: settings.StateName("done"),
			},
		},
		{
			name: "no trigger if for moved to state",
			trigger: settings.Trigger{
				TriggerIf: []settings.TriggerIf{
					{MovedTo: settings.StateName("in_progress")},
					{MovedTo: settings.StateName("done")},
				},
			},
			movedTo:  settings.StateName("todo"),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			triggerIf := tt.trigger.GetTriggerIf(tt.movedTo)
			assert.Equal(t, tt.expected, triggerIf)
		})
	}
}

func TestTriggerIf_AllSiblingsCheck(t *testing.T) {
	tests := []struct {
		name            string
		triggerIf       settings.TriggerIf
		siblingStatuses []settings.StateName
		expected        bool
	}{
		{
			name: "no all siblings status",
			triggerIf: settings.TriggerIf{
				AllSiblingsStatus: "",
			},
			siblingStatuses: []settings.StateName{},
			expected:        true,
		},
		{
			name: "all siblings status match",
			triggerIf: settings.TriggerIf{
				AllSiblingsStatus: settings.StateName("in_progress"),
			},
			siblingStatuses: []settings.StateName{
				settings.StateName("in_progress"),
				settings.StateName("in_progress"),
			},
			expected: true,
		},
		{
			name: "all siblings status mismatch",
			triggerIf: settings.TriggerIf{
				AllSiblingsStatus: settings.StateName("in_progress"),
			},
			siblingStatuses: []settings.StateName{
				settings.StateName("in_progress"),
				settings.StateName("done"),
			},
			expected: false,
		},
		{
			name: "no siblings return true",
			triggerIf: settings.TriggerIf{
				AllSiblingsStatus: settings.StateName("in_progress"),
			},
			siblingStatuses: []settings.StateName{},
			expected:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allSiblingsCheck := tt.triggerIf.AllSiblingsCheck(tt.siblingStatuses)
			assert.Equal(t, tt.expected, allSiblingsCheck)
		})
	}
}
