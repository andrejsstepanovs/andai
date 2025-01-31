package processor

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/pkg/employee/utils"
	"github.com/andrejsstepanovs/andai/pkg/exec"
	"github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/mattn/go-redmine"
	"github.com/teilomillet/gollm"
)

func GollmExecute(promptFile string, step models.Step) (exec.Output, error) {
	format := "Use this file %s as a question and Answer!"

	return exec.Exec(step.Command, step.Action, fmt.Sprintf(format, promptFile))
}

func GollmCreateIssue(targetIssueTypeName models.IssueTypeName, knowledgeFile string) (exec.Output, map[int]redmine.Issue, map[int][]int, error) {
	out, err := getIdeaResponse(targetIssueTypeName, knowledgeFile)
	if err != nil {
		return exec.Output{}, nil, nil, err
	}

	log.Println("######## OUTPUT ##########")
	log.Println(out.Stdout)
	log.Println("##################")

	taskPrompt, err := getTaskPrompt(targetIssueTypeName, out.Stdout)
	if err != nil {
		return exec.Output{}, nil, nil, err
	}

	promptExtension := ""
	var createIssues Answer
	tries := 4
	for {
		tries--
		if tries == 0 {
			return exec.Output{}, nil, nil, fmt.Errorf("failed to get issues from LlmNorm")
		}

		out, createIssues, promptExtension, err = getIssuesFromLLM(taskPrompt, promptExtension)
		if err != nil {
			return exec.Output{}, nil, nil, err
		}
		if promptExtension == "" {
			break
		}
	}

	items := make(map[int]redmine.Issue)
	deps := make(map[int][]int)
	for _, issue := range createIssues.Issues {
		items[issue.ID] = redmine.Issue{
			Subject:     issue.Subject,
			Description: issue.Description,
		}

		if deps[issue.ID] == nil {
			deps[issue.ID] = make([]int, 0)
		}
		deps[issue.ID] = append(deps[issue.ID], issue.BlockedBy...)
	}

	return out, items, deps, nil
}

// second string is LlmNorm response json parsing error
func __getIssuesFromLLM(taskPrompt string, promptExtension string) (exec.Output, Answer, string, error) {
	if promptExtension != "" {
		format := "%s\n\n\nYou failed last time with error: %s.\n" +
			"Try again and be more careful this time!"
		taskPrompt = fmt.Sprintf(format, taskPrompt, promptExtension)
	}
	taskFile, err := utils.BuildPromptTextTmpFile(taskPrompt)
	if err != nil {
		log.Printf("Failed to build prompt tmp file: %v", err)
		return exec.Output{}, Answer{}, "", err
	}
	prompt := fmt.Sprintf("Analyze given file %s and reformat it into JSON. Answer with raw JSON!", taskFile)

	out, err := exec.Exec("bobik", "zalando", "once", "llm", "quiet", prompt)
	if err != nil {
		return out, Answer{}, "", err
	}

	picked := Answer{}
	jojo := out.Stdout
	err = json.Unmarshal([]byte(jojo), &picked)
	if err != nil {
		jojo = gollm.CleanResponse(jojo)
		err = json.Unmarshal([]byte(jojo), &picked)
		if err != nil {
			return out, picked, err.Error(), nil
		}
	}

	err = picked.Validate()
	if err != nil {
		return out, picked, err.Error(), nil
	}

	return out, picked, "", nil
}
