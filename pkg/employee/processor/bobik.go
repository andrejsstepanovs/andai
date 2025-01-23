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

	format := "Answer with solution that is stated in the file %s !" +
		"You need to answer using raw JSON. Expected json format example data: \n```json\n%s\n```\n\n" +
		"You keep track on task dependencies on other tasks you create. " +
		"It is super important that tasks do not have cyclomatic dependencies! i.e. no 2 tasks depend on each other." +
		//"Do not afraid to extend Description field with other necessary information that will help developer to code this solution." +
		"It is really important that answer contains only raw JSON."

	txt := fmt.Sprintf(format, promptFile, exampleJson)
	fmt.Println(txt)
	out, err := exec.Exec("bobik", "zalando", "once", "llm", "quiet", txt)
	if err != nil {
		return exec.Output{}, err
	}

	resp := out.Stdout
	fmt.Println(resp)

	panic(1)
	return out, nil
}
