package exec_test

import (
	"fmt"
	"testing"

	"github.com/andrejsstepanovs/andai/internal/exec"
	"github.com/stretchr/testify/assert"
)

func TestOutput_AsPrompt(t *testing.T) {
	tests := []struct {
		name     string
		output   exec.Output
		expected string
	}{
		{
			name: "with stdout and stderr",
			output: exec.Output{
				Command: "ls -l",
				Stdout:  "total 0",
				Stderr:  "ls: cannot access 'nonexistent': No such file or directory",
			},
			expected: fmt.Sprintf(
				"Command: %q\nOutput:\n<stdout>\n%s\n</stdout>\n<stderr>\n%s\n</stderr>",
				"ls -l",
				"total 0",
				"ls: cannot access 'nonexistent': No such file or directory",
			),
		},
		{
			name: "only stdout",
			output: exec.Output{
				Command: "echo 'hello'",
				Stdout:  "hello",
				Stderr:  "",
			},
			expected: fmt.Sprintf(
				"Command: %q\nOutput:\n<stdout>\n%s\n</stdout>\n<stderr>\n%s\n</stderr>",
				"echo 'hello'",
				"hello",
				"",
			),
		},
		{
			name: "only stderr",
			output: exec.Output{
				Command: "cat non_existent_file",
				Stdout:  "",
				Stderr:  "cat: non_existent_file: No such file or directory",
			},
			expected: fmt.Sprintf(
				"Command: %q\nOutput:\n<stdout>\n%s\n</stdout>\n<stderr>\n%s\n</stderr>",
				"cat non_existent_file",
				"",
				"cat: non_existent_file: No such file or directory",
			),
		},
		{
			name: "empty stdout and stderr",
			output: exec.Output{
				Command: "touch file.txt",
				Stdout:  "",
				Stderr:  "",
			},
			expected: fmt.Sprintf(
				"Command: %q\nOutput:\n<stdout>\n%s\n</stdout>\n<stderr>\n%s\n</stderr>",
				"touch file.txt",
				"",
				"",
			),
		},
		{
			name: "command with quotes",
			output: exec.Output{
				Command: `git commit -m "Initial commit"`,
				Stdout:  "[main (root-commit) abcdefg] Initial commit",
				Stderr:  "",
			},
			expected: fmt.Sprintf(
				"Command: %q\nOutput:\n<stdout>\n%s\n</stdout>\n<stderr>\n%s\n</stderr>",
				`git commit -m "Initial commit"`,
				"[main (root-commit) abcdefg] Initial commit",
				"",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.output.AsPrompt()
			assert.Equal(t, tt.expected, actual)
		})
	}
}
