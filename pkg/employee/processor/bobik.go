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

func BobikExecute(promptFile string, step models.Step) (exec.Output, error) {
	format := "Use this file %s as a question and Answer!"
	return exec.Exec(step.Command, step.Action, fmt.Sprintf(format, promptFile))
}

func BobikCreateIssue(targetIssueTypeName models.IssueTypeName, knowledgeFile string) (exec.Output, map[int]redmine.Issue, map[int][]int, error) {
	taskPrompt, err := getTaskPrompt(targetIssueTypeName)
	if err != nil {
		return exec.Output{}, nil, nil, err
	}

	promptExtension := ""
	var createIssues Answer
	var out exec.Output
	tries := 4
	for {
		tries--
		if tries == 0 {
			return exec.Output{}, nil, nil, fmt.Errorf("failed to get issues from LLM")
		}

		out, createIssues, promptExtension, err = getIssuesFromLLM(
			knowledgeFile,
			taskPrompt,
			promptExtension,
		)
		if promptExtension != "" {
			continue
		}
		if err != nil {
			return exec.Output{}, nil, nil, err
		}
		break
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

func getTaskPrompt(targetIssueTypeName models.IssueTypeName) (string, error) {
	example := Answer{
		Issues: []AnswerIssues{
			{
				ID:          1,
				Subject:     "Issue Title",
				Description: "# Acceptance Criteria:\n- Point 1\n- Point 2:\n-- Subpoint 1\n-- Subpoint 2\n",
			},
			{
				ID:          2,
				Subject:     "Issue Title",
				Description: "# Acceptance Criteria:\n- Point 1\n- Point 2:\n-- Subpoint 1\n-- Subpoint 2\n",
				BlockedBy:   []int{1},
			},
			{
				ID:          3,
				Subject:     "Issue Title",
				Description: "# Acceptance Criteria:\n- Point 1\n- Point 2:\n-- Subpoint 1\n-- Subpoint 2\n",
				BlockedBy:   []int{2},
			},
		},
	}

	jsonTxt, err := json.Marshal(example)
	if err != nil {
		return "", err
	}
	exampleJSON := string(jsonTxt)

	format := "You need to Answer using raw JSON. Expected json format example data:\n" +
		"```json\n%s\n```\n" +
		"Keep track on task dependencies on other tasks you create.\n" +
		"It is super important that tasks do not have cyclomatic dependencies! i.e. no 2 tasks depend on each other.\n" +
		"It is really important that Answer contains only raw JSON with tags: " +
		"issues that contains issue type: **%s** elements.\n" +
		"Each element should contain: number_int, subject, " +
		"description (contains detailed explanation what needs to be achieved.\n" +
		"Use \"# Acceptance Criteria\" to define exact points), blocked_by_numbers (array of integers).\n" +
		"Do not use any other tags in JSON.\n" +
		"Do not create tasks that are out of scope of current issue requirements."

	taskPrompt := fmt.Sprintf(format, exampleJSON, targetIssueTypeName)
	return taskPrompt, err
}

// second string is LLM response json parsing error
func getIssuesFromLLM(knowledgeFile string, taskPrompt string, promptExtension string) (exec.Output, Answer, string, error) {
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

	format := "Use knowledge file %s and your task file %s to form a response. Answer with raws JSON!"
	prompt := fmt.Sprintf(format, knowledgeFile, taskFile)
	//prompt = fmt.Sprintf("\"%s\"", prompt)
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
