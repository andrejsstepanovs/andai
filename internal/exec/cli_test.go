package exec_test

import (
	"context"
	"testing"
	"time"

	"github.com/andrejsstepanovs/andai/internal/exec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExec(t *testing.T) {
	t.Run("successful command with stdout", func(t *testing.T) {
		output, err := exec.Exec("echo", 10*time.Second, "hello world")
		require.NoError(t, err)
		assert.Equal(t, "echo hello world", output.Command)
		assert.Equal(t, "hello world", output.Stdout)
		assert.Empty(t, output.Stderr)
	})

	t.Run("command with stderr", func(t *testing.T) {
		// Using a command that reliably produces stderr
		// Note: The exact error message might vary slightly depending on the system/shell
		output, err := exec.Exec("ls", 10*time.Second, "/non_existent_directory_for_test")
		require.Error(t, err) // Expecting an error because the command fails
		assert.Equal(t, "ls /non_existent_directory_for_test", output.Command)
		assert.Empty(t, output.Stdout)
		assert.Contains(t, output.Stderr, "No such file or directory") // Check for the core part of the error
	})

	t.Run("command with both stdout and stderr", func(t *testing.T) {
		// This command prints "hello" to stdout and then tries to list a non-existent file, producing stderr
		output, err := exec.Exec("echo hello && ls /non_existent_directory_for_test_again", 10*time.Second)
		require.Error(t, err)                                                                       // Expecting an error because the ls part fails
		assert.Equal(t, "echo hello && ls /non_existent_directory_for_test_again ", output.Command) // Note space if no args
		assert.Equal(t, "hello", output.Stdout)
		assert.Contains(t, output.Stderr, "No such file or directory")
	})

	t.Run("command times out", func(t *testing.T) {
		// Use a short timeout for a command that takes longer
		output, err := exec.Exec("sleep", 50*time.Millisecond, "0.1") // Sleep for 100ms, timeout at 50ms
		require.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded, "error should be context.DeadlineExceeded")
		assert.Equal(t, "sleep 0.1", output.Command)
		// Stdout/Stderr might be empty or contain partial output depending on timing
	})

	t.Run("command without arguments", func(t *testing.T) {
		output, err := exec.Exec("pwd", 10*time.Second)
		require.NoError(t, err)
		assert.Equal(t, "pwd ", output.Command) // Note the space
		assert.NotEmpty(t, output.Stdout)       // Should print the current directory
		assert.Empty(t, output.Stderr)
	})
}
