package models_test

import (
	"testing"

	"github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestTriggers_GetTriggers(t *testing.T) {
	tests := []struct {
		name      string
		triggers  models.Triggers
		issueType models.IssueTypeName
		expected  []models.Trigger
	}{
		{
			name:      "empty triggers",
			triggers:  models.Triggers{},
			issueType: models.IssueTypeName(""),
			expected:  []models.Trigger{},
		},
		{
			name: "single trigger",
			triggers: models.Triggers{
				{IssueType: models.IssueTypeName("bug")},
			},
			issueType: models.IssueTypeName("bug"),
			expected: []models.Trigger{
				{IssueType: models.IssueTypeName("bug")},
			},
		},
		{
			name: "multiple triggers",
			triggers: models.Triggers{
				{IssueType: models.IssueTypeName("bug")},
				{IssueType: models.IssueTypeName("feature")},
			},
			issueType: models.IssueTypeName("feature"),
			expected: []models.Trigger{
				{IssueType: models.IssueTypeName("feature")},
			},
		},
		{
			name: "no triggers for issue type",
			triggers: models.Triggers{
				{IssueType: models.IssueTypeName("bug")},
				{IssueType: models.IssueTypeName("feature")},
			},
			issueType: models.IssueTypeName("task"),
			expected:  []models.Trigger{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			triggers := tt.triggers.GetTriggers(tt.issueType)
			assert.Equal(t, tt.expected, triggers)
		})
	}
}
