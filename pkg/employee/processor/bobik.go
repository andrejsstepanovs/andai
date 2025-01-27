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
	format := "Use this file %s as a question and answer!"
	return exec.Exec(step.Command, step.Action, fmt.Sprintf(format, promptFile))
}

type answerIssues struct {
	ID          int    `json:"number_int"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
	BlockedBy   []int  `json:"blocked_by_numbers" validate:"omitempty"`
}
type answer struct {
	Issues []answerIssues `json:"issues"`
}

func BobikCreateIssue(targetIssueTypeName models.IssueTypeName, knowledgeFile string) (exec.Output, []redmine.Issue, map[int][]int, error) {
	taskPrompt, err := getTaskPrompt(targetIssueTypeName)
	if err != nil {
		return exec.Output{}, nil, nil, err
	}

	promptExtension := ""
	var createIssues answer
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
		if err != nil {
			return exec.Output{}, nil, nil, err
		}
		break
	}

	items := make([]redmine.Issue, 0)
	deps := make(map[int][]int)
	for i, issue := range createIssues.Issues {
		items = append(items, redmine.Issue{
			Subject:     issue.Subject,
			Description: issue.Description,
		})
		if deps[i] == nil {
			deps[i] = make([]int, 0)
		}
		for _, blockedBy := range issue.BlockedBy {
			deps[i] = append(deps[i], blockedBy)
		}
	}

	return out, items, deps, nil
}

// getIssuesFromLLM returns issues from LLM
func getTaskPrompt(targetIssueTypeName models.IssueTypeName) (string, error) {
	example := answer{
		Issues: []answerIssues{
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
	exampleJson := string(jsonTxt)

	format := "You need to answer using raw JSON. Expected json format example data:\n" +
		"```json\n%s\n```\n" +
		"Keep track on task dependencies on other tasks you create.\n" +
		"It is super important that tasks do not have cyclomatic dependencies! i.e. no 2 tasks depend on each other.\n" +
		"It is really important that answer contains only raw JSON with tags: " +
		"issues that contains issue type: **%s** elements.\n" +
		"Each element should contain: number_int, subject, " +
		"description (contains detailed explanation what needs to be achieved.\n" +
		"Use \"# Acceptance Criteria\" to define exact points), blocked_by_numbers (array of integers).\n" +
		"Do not use any other tags in JSON.\n" +
		"Do not create tasks that are out of scope of current issue requirements."

	taskPrompt := fmt.Sprintf(format, exampleJson, targetIssueTypeName)
	return taskPrompt, err
}

// second string is LLM response json parsing error
func getIssuesFromLLM(knowledgeFile string, taskPrompt string, promptExtension string) (exec.Output, answer, string, error) {
	if promptExtension != "" {
		format := "%s\n\n\nYou failed last time with error: %s.\n" +
			"Try again and be more careful this time!"
		taskPrompt = fmt.Sprintf(format, taskPrompt, promptExtension)
	}
	taskFile, err := utils.BuildPromptTextTmpFile(taskPrompt)
	if err != nil {
		log.Printf("Failed to build prompt tmp file: %v", err)
		return exec.Output{}, answer{}, "", err
	}

	format := "Use knowledge file %s and your task file %s to form a response. Answer with raws JSON!"
	prompt := fmt.Sprintf(format, knowledgeFile, taskFile)
	//prompt = fmt.Sprintf("\"%s\"", prompt)
	out, err := exec.Exec("bobik", "zalando", "once", "llm", "quiet", prompt)
	if err != nil {
		return out, answer{}, "", err
	}

	picked := answer{}
	jojo := out.Stdout
	err = json.Unmarshal([]byte(jojo), &picked)
	if err != nil {
		jojo = gollm.CleanResponse(jojo)
		err = json.Unmarshal([]byte(jojo), &picked)
		if err != nil {
			return out, picked, err.Error(), nil
		}
	}
	return out, picked, "", nil
}
