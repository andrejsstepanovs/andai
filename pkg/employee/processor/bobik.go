package processor

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/andrejsstepanovs/andai/pkg/exec"
	"github.com/andrejsstepanovs/andai/pkg/models"
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

func BobikCreateIssue(parentIssueID int, targetIssueTypeName models.IssueTypeName, promptFile string) (exec.Output, error) {
	promptExtension := ""
	var createIssues answer
	tries := 4
	for {
		tries--
		if tries == 0 {
			return exec.Output{}, fmt.Errorf("failed to get issues from LLM")
		}
		var err error
		createIssues, promptExtension, err = getIssuesFromLLM(promptFile, targetIssueTypeName, promptExtension)
		if err != nil {
			return exec.Output{}, err
		}
		break
	}

	for _, issue := range createIssues.Issues {

	}

	panic(1)
	return out, nil
}

// getIssuesFromLLM returns issues from LLM
// second string is LLM response json parsing error
func getIssuesFromLLM(promptFile string, targetIssueTypeName models.IssueTypeName, promptExtension string) (answer, string, error) {
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
		return answer{}, "", err
	}
	exampleJson := string(jsonTxt)

	format := "Answer with solution that is stated in the file %s ! " +
		"You need to answer using raw JSON. Expected json format example data: ```%s``` " +
		"Keep track on task dependencies on other tasks you create. " +
		"It is super important that tasks do not have cyclomatic dependencies! i.e. no 2 tasks depend on each other. " +
		//"Do not afraid to extend Description field with other necessary information that will help developer to code this solution." +
		"It is really important that answer contains only raw JSON with tags: issues that contains issue type: %s elements. " +
		"Each element should contain: number_int, subject, " +
		"description (contains detailed explanation what needs to be achieved. " +
		"Use \"# Acceptance Criteria\" to define exact points), blocked_by_numbers (array of integers)." +
		"Do not use any other tags in JSON. " +
		"Do not create tasks that are out of scope of current issue requirements."

	prompt := fmt.Sprintf(format, promptFile, exampleJson, targetIssueTypeName)

	if promptExtension != "" {
		format = "%s You failed last time with error: %s. Try again and be more careful this time!"
		prompt = fmt.Sprintf(format, prompt, promptExtension)
	}

	searchReplace := map[string]string{
		`"`:  `\"`,
		`''`: `\'`,
		"\n": ` `,
	}
	for search, replace := range searchReplace {
		prompt = strings.ReplaceAll(prompt, search, replace)
	}
	prompt = strings.TrimSpace(prompt)
	//fmt.Println(prompt)

	out, err := exec.Exec("bobik", "zalando", "once", "llm", "quiet", "\""+prompt+"\"")
	if err != nil {
		return answer{}, "", err
	}

	result := answer{}
	err = json.Unmarshal([]byte(out.Stdout), &result)
	if err != nil {
		return result, err.Error(), nil
	}
	return result, "", nil
}
