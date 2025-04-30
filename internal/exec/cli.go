package exec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	shell     = "zsh"
	shellPath = "/usr/bin/zsh"
	shellArg  = "-ic"
)

// Exec executes command with timeout.
func Exec(command string, timeout time.Duration, args ...string) (Output, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return WithContext(ctx, command, args...)
}

// WithContext executes command with context for cancellation and timeout.
func WithContext(ctx context.Context, command string, args ...string) (Output, error) {
	// The caller is now responsible for context timeout and cancellation.
	// No need to create a context with timeout here.

	cmdExec := command
	if len(args) > 0 {
		// if one to last is -c then wrap the last argument in quotes
		if len(args) > 1 && args[len(args)-2] == "-c" {
			args[len(args)-1] = fmt.Sprintf("\"%s\"", strings.Replace(args[len(args)-1], "\"", "\\\"", -1))
		}

		cmdExec = fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	}

	fmt.Println(cmdExec)
	output := Output{Command: cmdExec}

	fullCommand := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	log.Printf("EXEC: %s", fullCommand)

	cmd := exec.CommandContext(ctx, shell, shellArg, fullCommand) // nolint:gosec
	cmd.Env = append(os.Environ(), fmt.Sprintf("SHELL=%s", shellPath))

	var stdout, stderr bytes.Buffer
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return output, fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return output, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	var wg sync.WaitGroup

	// Start the command
	if err := cmd.Start(); err != nil {
		return output, fmt.Errorf("failed to start command: %w", err)
	}

	// Read stdout asynchronously
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&stdout, stdoutPipe)
	}()

	// Read stderr asynchronously
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&stderr, stderrPipe)
	}()

	// Wait for I/O goroutines to finish. This ensures we capture all output
	// *before* waiting for the command to exit.
	wg.Wait()

	// Wait for the command to complete and release resources
	err = cmd.Wait()

	// Check if the context timed out or was canceled.
	// This is important because cmd.Wait() might return a generic error (like "signal: killed")
	// when the context is done. Prioritize the context error.
	if ctxErr := ctx.Err(); ctxErr != nil {
		// Check specifically for DeadlineExceeded or Canceled
		if errors.Is(ctxErr, context.DeadlineExceeded) || errors.Is(ctxErr, context.Canceled) {
			err = ctxErr // Prioritize the context error
		}
	}

	retStdOut := stdout.String()
	retStdErr := stderr.String()
	if retStdOut == "<nil>" {
		retStdOut = ""
	}
	if retStdErr == "<nil>" {
		retStdErr = ""
	}

	output.Stdout = strings.TrimSpace(retStdOut)
	output.Stderr = strings.TrimSpace(retStdErr)

	return output, err
}
