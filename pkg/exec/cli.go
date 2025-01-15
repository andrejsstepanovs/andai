package exec

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	shell           = "zsh"
	shellPath       = "/usr/bin/zsh"
	shellArg        = "-ic"
	allowedCommands = "pwd cat git ls grep aider aid bobik"
)

func Exec(command string, args ...string) (string, string, error) {
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
	if !isAllowed {
		return "", "", fmt.Errorf("command '%s' is not allowed", command)
	}

	fullCommand := fmt.Sprintf("%s %s", command, filepath.Join(args...))

	cmd := exec.CommandContext(ctx, shell, shellArg, fullCommand)
	cmd.Env = append(os.Environ(), fmt.Sprintf("SHELL=%s", shellPath))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", "", fmt.Errorf("exec failed: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	fmt.Printf("stdout: %s", stdout.String())
	fmt.Printf("stdERR: %s", stderr.String())

	return stdout.String(), stderr.String(), nil
}
