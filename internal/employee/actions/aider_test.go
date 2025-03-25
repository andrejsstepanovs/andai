package actions

import (
	"strings"
	"testing"

	"github.com/andrejsstepanovs/andai/internal/employee/utils"
	"github.com/stretchr/testify/assert"
)

func TestAider_aiderRemoveUnnecessaryLines(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "remove lines starting with specified prefixes",
			input: []string{
				"Initial repo scan can be slow in larger repos, but only happens once.",
				"Normal line",
				"Repo-map can't include something",
				"Has it been deleted from the file system but not from git?",
			},
			expected: []string{
				"Normal line",
			},
		},
		{
			name: "remove lines containing CONVENTIONS.md",
			input: []string{
				"Some text with CONVENTIONS.md in middle",
				"Normal line",
				"CONVENTIONS.md at start",
			},
			expected: []string{
				"Normal line",
			},
		},
		{
			name: "skip Added to chat lines",
			input: []string{
				"Added file.txt to the chat",
				"Normal line",
			},
			expected: []string{
				"Normal line",
			},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: []string{},
		},
		{
			name: "no lines to remove",
			input: []string{
				"Normal line 1",
				"Normal line 2",
			},
			expected: []string{
				"Normal line 1",
				"Normal line 2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := aiderRemoveUnnecessaryLines(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLongTextRealWorldExample(t *testing.T) {
	content, err := utils.GetFileContents("testdata/output.txt")
	assert.NoError(t, err)
	expectedOutput, err := utils.GetFileContents("testdata/expected.txt")
	assert.NoError(t, err)

	lines := strings.Split(content, "\n")
	lines = aiderRemoveUnnecessaryLines(lines)

	newContent := strings.Join(lines, "\n")
	assert.Equal(t, expectedOutput, newContent)
}
