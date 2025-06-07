package exec

import (
	"fmt"
	"strings"

	"github.com/andrejsstepanovs/andai/internal/settings"
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
		"--yes-always",
		"--yes",
		"--no-auto-lint",
		"--no-auto-test",
		"--git",
	}

	aiderDefaultParams = map[string]string{
		"--map-refresh": "always", // auto,always,files,manual
	}

	aiderCodeArgs = []string{
		"--auto-commits",
	}

	aiderArchitectCodeArgs = []string{
		"--auto-commits",
	}

	aiderArchitectArgs = []string{
		"--no-auto-commits",
	}

	aiderCommitArgs = []string{
		"--commit",
	}

	aiderCodeParams = map[string]string{
		//	"--chat-mode": "diff-fenced", // code. by default it is code so no need to set anything.
	}
	aiderArchitectParams = map[string]string{
		"--chat-mode": "ask",
	}
	aiderArchitectCodeParams = map[string]string{
		"--chat-mode": "architect",
	}
	aiderCommitParams = map[string]string{
		"--chat-mode": "code", // not sure if this is correct
	}
)

func AiderCommand(contextFile string, step settings.Step, config settings.Aider) string {
	var (
		params map[string]string
		args   []string
	)
	switch step.Action {
	case "architect":
		params = aiderArchitectParams
		args = aiderArchitectArgs
	case "commit":
		params = aiderCommitParams
		args = aiderCommitArgs
	case "code":
		params = aiderCodeParams
		args = aiderCodeArgs
	case "architect-code":
		params = aiderArchitectCodeParams
		args = aiderArchitectCodeArgs
	default:
		panic("unknown step action")
	}

	params["--config"] = config.Config

	if config.MapTokens > 0 {
		params["--map-tokens"] = fmt.Sprintf("%d", config.MapTokens)
	}

	if contextFile != "" {
		params["--message-file"] = contextFile
	} else {
		params["--message"] = step.Prompt.ForCli()
	}

	if config.ModelMetadataFile != "" {
		params["--model-metadata-file"] = config.ModelMetadataFile // https://aider.chat/docs/config/adv-model-settings.html
	}

	paramsCli := make([]string, 0, len(params))
	for k, v := range aiderDefaultParams {
		paramsCli = append(paramsCli, fmt.Sprintf("%s=%q", k, v))
	}
	for k, v := range params {
		paramsCli = append(paramsCli, fmt.Sprintf("%s=%q", k, v))
	}

	if len(step.ContextFiles) > 0 {
		for _, file := range step.ContextFiles {
			//log.Printf("Aider adding file: %q\n", file)
			paramsCli = append(paramsCli, fmt.Sprintf("--file=%q", file))
		}
	}

	args = append(args, aiderArgs...)

	return fmt.Sprintf(
		"%s %s",
		strings.Join(args, " "),
		strings.Join(paramsCli, " "),
	)
}

//#### I’ll need to see the existing tests for the other merger components and the API to know exactly what to change. Could you please add the contents of the following test files:
//####
//#### - tests/utils/campaign_manager/base_merger_component.py
//#### - tests/utils/campaign_manager/brand_name_merger.py
//#### - tests/utils/campaign_manager/campaign_merger_api.py
//#### - tests/utils/campaign_manager/commodity_groups_merger.py
//#### - tests/utils/campaign_manager/target_groups_merger.py
//
//I don’t currently have those test files in the chat. Could you please share the contents of:
//
//- tests/utils/campaign_manager/base_merger_component.py
//- tests/utils/campaign_manager/brand_name_merger.py
//- tests/utils/campaign_manager/campaign_merger_api.py
//- tests/utils/campaign_manager/commodity_groups_merger.py
//- tests/utils/campaign_manager/target_groups_merger.py
//
//Once I have them, I can review and make the appropriate updates.
