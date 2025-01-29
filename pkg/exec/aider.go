package exec

import (
	"fmt"
	"strings"

	"github.com/andrejsstepanovs/andai/pkg/models"
)

var (
	aiderArgs = []string{
		"--no-stream",
		"--no-pretty",
		"--no-gitignore",
		"--no-restore-chat-history",
		"--no-analytics",
		"--no-dirty-commits",
		"--no-watch-files",
		"--no-suggest-shell-commands",
		"--no-fancy-input",
		"--no-show-release-notes",
		"--no-check-update",
		"--analytics-disable",
		"--no-detect-urls",
		"--no-show-model-warnings",
		"--no-dry-run",
		"--no-gui",
		"--no-browser",
		"--no-copy-paste",
		"--git",
		"--yes-always",
	}
	aiderCodeArgs = []string{
		"--auto-commits",
		"--no-auto-lint",
		"--no-auto-test",
	}
	aiderArchitectArgs = []string{
		"--architect",
		"--no-auto-commits",
		"--no-auto-lint",
		"--no-auto-test",
	}

	aiderArchitectParams = map[string]string{
		"--map-refresh": "auto", // auto,always,files,manual
	}
	aiderCodeParams = map[string]string{
		"--map-refresh": "auto", // auto,always,files,manual
	}
)

func AiderCommand(contextFile string, step models.Step) string {
	var (
		params map[string]string
		args   []string
	)
	switch step.Action {
	case "architect":
		params = aiderArchitectParams
		args = aiderArchitectArgs
	case "code":
		params = aiderCodeParams
		args = aiderCodeArgs
	default:
		panic("unknown step action")
	}

	if contextFile != "" {
		params["--message-file"] = contextFile
	} else {
		params["--message"] = step.Prompt.ForCli()
	}

	paramsCli := make([]string, 0, len(params))
	for k, v := range params {
		paramsCli = append(paramsCli, fmt.Sprintf("%s=%q", k, v))
	}

	args = append(args, aiderArgs...)

	return fmt.Sprintf(
		"%s %s",
		strings.Join(args, " "),
		strings.Join(paramsCli, " "),
	)
}
