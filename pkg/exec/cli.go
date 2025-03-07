package exec

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	shell     = "zsh"
	shellPath = "/usr/bin/zsh"
	shellArg  = "-ic"
)

// Exec executes command with timeout.
// TODO need context cancelling. Now Ctr+C is stopping exec bash but not the loop.
func Exec(command string, timeout time.Duration, args ...string) (Output, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	output := Output{Command: fmt.Sprintf("%s %s", command, strings.Join(args, " "))}

	fullCommand := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	log.Printf("EXEC: %s", fullCommand)

	cmd := exec.CommandContext(ctx, shell, shellArg, fullCommand) // nolint:gosec
	cmd.Env = append(os.Environ(), fmt.Sprintf("SHELL=%s", shellPath))

	var stdout, stderr bytes.Buffer
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	// Start the command
	if err := cmd.Start(); err != nil {
		return output, err
	}

	// Read stdout asynchronously
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				stdout.Write(buf[:n])
				fmt.Printf("stdout: %s", string(buf[:n]))
			}
			if err != nil {
				break
			}
		}
	}()

	// Read stderr asynchronously
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 {
				stderr.Write(buf[:n])
				fmt.Printf("stdERR: %s", string(buf[:n]))
			}
			if err != nil {
				break
			}
		}
	}()

	// Wait for the command to complete
	err := cmd.Wait()

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
