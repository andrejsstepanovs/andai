package processor

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/andrejsstepanovs/andai/pkg/exec"
	"github.com/andrejsstepanovs/andai/pkg/models"
)

func AiderExecute(contextFile string, step models.Step) (exec.Output, error) {
	log.Printf("Context file: %q\n", contextFile)
	defer os.Remove(contextFile)

	options := exec.AiderCommand(contextFile, step)
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
