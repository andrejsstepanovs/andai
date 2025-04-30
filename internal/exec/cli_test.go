package exec_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/andrejsstepanovs/andai/internal/exec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ExecWithContext(t *testing.T) {
	t.Run("successful command with stdout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		output, err := exec.WithContext(ctx, "echo", "hello world")
		require.NoError(t, err)
		assert.Equal(t, "echo hello world", output.Command) // Updated assertion
		assert.Equal(t, "hello world", output.Stdout)
		assert.Empty(t, output.Stderr)
	})

	t.Run("command with stderr", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		output, err := exec.WithContext(ctx, "ls", "/non_existent_directory_for_test")
		require.Error(t, err)
		assert.Equal(t, "ls /non_existent_directory_for_test", output.Command) // Updated assertion
		assert.Empty(t, output.Stdout)
		assert.Contains(t, output.Stderr, "No such file or directory")
	})

	t.Run("command with both stdout and stderr", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		// Use sh -c for compound commands
		cmdStr := "echo hello && ls /non_existent_directory_for_test_again"
		output, err := exec.WithContext(ctx, "sh", "-c", cmdStr)
		require.Error(t, err)
		assert.Equal(t, fmt.Sprintf("sh -c \"%s\"", cmdStr), output.Command) // Updated assertion
		assert.Contains(t, output.Stdout, "hello")                           // Still check contains due to potential shell variations
		assert.Contains(t, output.Stderr, "No such file or directory")
	})

	t.Run("command times out", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		output, err := exec.WithContext(ctx, "sleep", "0.1")
		require.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded, "error should be context.DeadlineExceeded")
		assert.Equal(t, "sleep 0.1", output.Command) // Updated assertion
	})

	t.Run("command is canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var output exec.Output
		var err error
		done := make(chan struct{})

		go func() {
			output, err = exec.WithContext(ctx, "sleep", "1")
			close(done)
		}()

		time.Sleep(50 * time.Millisecond)
		cancel()
		<-done

		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled, "error should be context.Canceled")
		assert.Equal(t, "sleep 1", output.Command) // Updated assertion
	})

	t.Run("command without arguments", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		output, err := exec.WithContext(ctx, "pwd")
		require.NoError(t, err)
		assert.Equal(t, "pwd", output.Command) // Updated assertion (no trailing space)
		assert.NotEmpty(t, output.Stdout)
		assert.Empty(t, output.Stderr)
	})

	t.Run("command with streaming output", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Command for bash -c
		streamingCmd := `for i in {1..10}; do echo -n "Output chunk \$i... "; sleep 0.1; done; echo "done"`
		output, err := exec.WithContext(ctx, "bash", "-c", streamingCmd)
		require.NoError(t, err)

		// Verify the output contains all the expected chunks
		for i := 1; i <= 10; i++ {
			assert.Contains(t, output.Stdout, fmt.Sprintf("Output chunk %d", i))
		}
		assert.Contains(t, output.Stdout, "done")

		// Command should match what we passed
		assert.Equal(t, `bash -c "for i in {1..10}; do echo -n \"Output chunk \$i... \"; sleep 0.1; done; echo \"done\""`, output.Command) // Updated assertion
		assert.Empty(t, output.Stderr)
	})
}
