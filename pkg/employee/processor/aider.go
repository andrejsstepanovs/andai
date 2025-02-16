package processor

import (
	"fmt"
	"log"
	"strings"

	"github.com/andrejsstepanovs/andai/pkg/exec"
	"github.com/andrejsstepanovs/andai/pkg/models"
)

// AiderExecute executes the command and returns the output.
// If contextFile is provided step.Prompt will be ignored. (don't worry, it should be part of contextFile).
// If you want to use step.Prompt, provide empty string for contextFile.
func AiderExecute(contextFile string, step models.Step, aiderConfig models.Aider) (exec.Output, error) {
	if contextFile != "" {
		log.Printf("Context file: %q\n", contextFile)
	}

	options := exec.AiderCommand(contextFile, step, aiderConfig)
	output, err := exec.Exec(step.Command, options)
	if err != nil {
		log.Printf("Failed to execute command: %v", err)
		return output, err
	}

	if output.Stdout != "" {
		fmt.Printf("stdout: %s\n", output.Stdout)

		lines := strings.Split(output.Stdout, "\n")
		startPos := 0
		lastPos := 0
		for k, line := range lines {
			if startPos == 0 && strings.Contains(line, "CONVENTIONS.md") &&
				strings.Contains(line, "Added") &&
				strings.Contains(line, "to the chat") {
				startPos = k
			}
			if lastPos == 0 && strings.Contains(line, "Tokens:") &&
				strings.Contains(line, "sent") &&
				strings.Contains(line, "received") {
				lastPos = k
			}
		}

		if startPos > 0 && lastPos > 0 {
			output.Stdout = strings.Join(lines[startPos+1:lastPos], "\n")
		}
	}

	return output, nil
}
