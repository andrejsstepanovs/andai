package processor

import (
	"encoding/json"
	"fmt"

	"github.com/andrejsstepanovs/andai/pkg/exec"
	"github.com/andrejsstepanovs/andai/pkg/models"
)

func BobikExecute(promptFile string, step models.Step) (exec.Output, error) {
	format := "Use this file %s as a question and answer!"
	return exec.Exec(step.Command, step.Action, fmt.Sprintf(format, promptFile))
}

func BobikCreateIssue(parentIssueID int, targetIssueTypeName models.IssueTypeName, promptFile string) (exec.Output, error) {
	type answerIssues struct {
		ID          int    `json:"number_int"`
		Subject     string `json:"subject"`
		Description string `json:"description"`
		BlockedBy   []int  `json:"blocked_by_numbers" validate:"omitempty"`
	}
	type answer struct {
		Issues []answerIssues `json:"issues"`
	}
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

	json, err := json.Marshal(example)
	if err != nil {
		return exec.Output{}, err
	}
	exampleJson := string(json)

	format := "Use this file %s as a question and answer!" +
		"You need to answer using raw JSON format \n```json\n%s\n```\n\n" +
		"It is really important that answer contains only raw JSON."

	out, err := exec.Exec("bobik", "zalando", "once", "llm", "quiet", fmt.Sprintf(format, promptFile, exampleJson))
	if err != nil {
		return exec.Output{}, err
	}

	resp := out.Stdout
	fmt.Println(resp)

	panic(1)
	return out, nil
}
