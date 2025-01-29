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
	shell           = "zsh"
	shellPath       = "/usr/bin/zsh"
	shellArg        = "-ic"
	allowedCommands = "pwd cat git ls grep aider aid bobik"
)

func Exec(command string, args ...string) (Output, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	allowedCmdList := strings.Split(allowedCommands, " ")
	isAllowed := false
	for _, cmd := range allowedCmdList {
		if cmd == command {
			isAllowed = true
			break
		}
	}

	output := Output{Command: fmt.Sprintf("%s %s", command, strings.Join(args, " "))}
	if !isAllowed {
		return output, fmt.Errorf("command '%s' is not allowed", command)
	}

	fullCommand := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	log.Printf("EXEC: %s", fullCommand)

	cmd := exec.CommandContext(ctx, shell, shellArg, fullCommand) // nolint:gosec
	cmd.Env = append(os.Environ(), fmt.Sprintf("SHELL=%s", shellPath))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return output, fmt.Errorf("exec failed: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	fmt.Printf("stdout: %s", stdout.String())
	fmt.Printf("stdERR: %s", stderr.String())

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

	return output, nil
}
