package processor

import (
	"log"
	"strings"

	"github.com/andrejsstepanovs/andai/pkg/ai"
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
	output, err := exec.Exec(step.Command, aiderConfig.Timeout, options)
	if err != nil {
		log.Printf("Failed to execute command: %v", err)
		return output, err
	}
	output = ai.RemoveThinkingFromOutput(output)

	if output.Stdout != "" {
		lines := strings.Split(output.Stdout, "\n")
		lines = aiderRemoveUnnecessaryLines(lines)

		output.Stdout = strings.Join(lines, "\n")
	}

	return output, nil
}

func aiderRemoveUnnecessaryLines(lines []string) []string {
	removeLinesStartingWith := []string{
		"Scanning repo:",
		"Main model:",
		"Weak model:",
		"Summarization failed for model",
		"Initial repo scan can be slow in larger repos, but only happens once",
		"Repo-map can't include",
		"Has it been deleted from the file system but not from git",
		"Analytics have been permanently disabled",
		"Aider v0",
		"Git repo: .git with",
		"Repo-map: using",
		"Warning: it's best to only add files that need changes to the chat",
		"summarizer unexpectedly failed",
		"https://aider.chat/docs/troubleshooting/",
		"architect edit format",
		"aiderignore spec.",
	}

	removeLinesContaining := [][]string{
		{"CONVENTIONS.md"},
		{"Added", "to the chat"},
		{"Tokens:", "sent", "received"},
		{"Model:", "with architect"},
		{"Editor model:", "with diff"},
		{"Skipping", "that matches"},
	}

	removeExactLines := []string{
		"architect edit format",
		"edit format",
	}

	newLines := make([]string, 0)
	for _, line := range lines {
		foundRemove := false
		for _, remove := range removeExactLines {
			if line == remove {
				foundRemove = true
				break
			}
		}
		if foundRemove {
			continue
		}
		includeLine := true
		for _, prefix := range removeLinesStartingWith {
			if strings.HasPrefix(line, prefix) {
				includeLine = false
				break
			}
		}
		if includeLine {
			newLines = append(newLines, line)
		}
	}

	cleanLines := make([]string, 0)
	for _, line := range newLines {
		if strings.Contains(line, "Added") && strings.Contains(line, "to the chat") {
			continue
		}

		includeLine := true
		for _, contains := range removeLinesContaining {
			matches := true
			for _, c := range contains {
				if !strings.Contains(line, c) {
					matches = false
					break
				}
			}
			if matches {
				includeLine = false
				break
			}
		}

		if includeLine {
			cleanLines = append(cleanLines, line)
		}
	}

	return cleanLines
}
