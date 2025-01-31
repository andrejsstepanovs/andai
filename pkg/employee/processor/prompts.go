package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/pkg/ai"
	"github.com/andrejsstepanovs/andai/pkg/employee/utils"
	"github.com/andrejsstepanovs/andai/pkg/exec"
	"github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/teilomillet/gollm"
)

func getIdeaResponse(llmNorm *ai.AI, targetIssueTypeName models.IssueTypeName, knowledgeFile string) (exec.Output, error) {
	templatePrompt := gollm.NewPromptTemplate(
		"ThinkIssueSplit",
		"Thinks how to split issue into smaller issues",
		"# Your task is:\n"+
			"Split current_issue into smaller scope {{.TargetIssueType}} issues.\n\n"+
			"## Instructions:\n"+
			"- Your task is to split the current issue into smaller scope issues.\n"+
			"- To do this effectively, you'll need to understand the context provided in the parent task, "+
			"but your focus should remain solely on splitting the current issue.\n"+
			"- Do not create smaller scope issues that address aspects outside the scope of the current issue, "+
			"even if they are mentioned in the parent task.\n"+
			"- Make sure to follow instructions provided in comments.\n\n"+
			"## Answer Instructions:\n"+
			"- Answer with a list of **{{.TargetIssueType}}** issues.\n"+
			"- Each issue should have:\n"+
			"-- Issue Number: integer\n"+
			"-- Subject: Issue title\n"+
			"-- Description: detailed explanation what needs to be achieved.\n"+
			"-- Blocked By: list of integers that blocks this task and need to be done beforehand (avoid circular dependencies).\n",
		gollm.WithPromptOptions(
			gollm.WithContext("You are software engineer working on creating Jira tasks."),
		),
	)

	prompt, err := templatePrompt.Execute(map[string]interface{}{
		"TargetIssueType": targetIssueTypeName,
	})
	if err != nil {
		log.Fatalf("Failed to execute prompt template: %v", err)
	}

	ctx := context.Background()
	return llmNorm.Generate(ctx, prompt, gollm.WithJSONSchemaValidation())
}

func getTaskPrompt(targetIssueTypeName models.IssueTypeName, solution string) (string, error) {
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

	format := "# Your task is:\nConvert solution issues %s into specific format json data.\n\n" +
		"<selected_solutions>\n%s\n</selected_solutions>" +
		"## Instructions:" +
		"- Use selected_solutions tag text and convert it into JSON data.\n" +
		"- I expect your answer to have following JSON structure (example):\n" +
		"```json\n%s\n```\n" +
		"- It is really important that Answer contains only raw JSON with tags: " +
		"issues that contains array of elements.\n" +
		"- Each element should contain: number_int, subject, description, blocked_by_numbers.\n" +
		"- Where blocked_by_numbers is array of integers.\n" +
		"- Do not use any other tags in JSON.\n"

	taskPrompt := fmt.Sprintf(format, targetIssueTypeName, solution, exampleJSON)
	return taskPrompt, err
}

// second string is LlmNorm response json parsing error
func getIssuesFromLLM(taskPrompt string, promptExtension string) (exec.Output, Answer, string, error) {
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
