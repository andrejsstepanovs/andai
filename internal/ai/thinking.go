package ai

import (
	"strings"

	"github.com/andrejsstepanovs/andai/internal/exec"
)

// RemoveThinkingContent removes lines starting with <think> and ending with </think>
func RemoveThinkingContent(response string) string {
	lines := strings.Split(response, "\n")
	clean := make([]string, 0)
	inside := false
	for _, line := range lines {
		if strings.Contains(line, "<think>") {
			inside = true
			continue
		}
		if !inside {
			clean = append(clean, line)
		}
		if strings.Contains(line, "</think>") {
			inside = false
		}
	}

	return strings.Join(clean, "\n")
}

func RemoveThinkingFromOutput(output exec.Output) exec.Output {
	output.Stdout = RemoveThinkingContent(output.Stdout)
	return output
}
