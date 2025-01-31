package processor

import (
	"context"
	"encoding/json"
	"log"

	"github.com/andrejsstepanovs/andai/pkg/ai"
	"github.com/andrejsstepanovs/andai/pkg/employee/utils"
	"github.com/andrejsstepanovs/andai/pkg/exec"
	"github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/mattn/go-redmine"
	"github.com/teilomillet/gollm"
)

func GenerateIssues(llm *ai.AI, targetIssueTypeName models.IssueTypeName, knowledgeFile string) (exec.Output, map[int]redmine.Issue, map[int][]int, error) {
	createIssues, err := getIssues(llm, targetIssueTypeName, knowledgeFile)
	if err != nil {
		return exec.Output{}, nil, nil, err
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

	return exec.Output{}, items, deps, nil
}

func getIssues(llmNorm *ai.AI, targetIssueTypeName models.IssueTypeName, knowledgeFile string) (Answer, error) {
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
	jsonResp, err := json.Marshal(example)
	if err != nil {
		return Answer{}, err
	}

	knowledge, err := utils.GetFileContents(knowledgeFile)
	if err != nil {
		return Answer{}, err
	}

	templatePrompt := gollm.NewPromptTemplate("IssuePlanToJson", "",
		"# Instructions:\n"+
			"- Use Context and specifically comments section to convert proposed issues into JSON data.\n"+
			"- Make sure to not make circular dependencies between issues.\n"+
			"- Convert suggested issues {{.TargetIssueType}} into specific format json data.\n"+
			"- Each element should contain: number_int, subject, description, blocked_by_numbers.\n"+
			"- Where blocked_by_numbers is array of integers.\n"+
			"- Use example data structure for your answer.\n"+
			"- Do not use any other tags in JSON.\n"+
			ai.ForceJson,
		gollm.WithPromptOptions(
			gollm.WithSystemPrompt("You are software engineer working with Jira on single issue breakdown task.", ""),
			gollm.WithDirectives("Convert given context content into issues as JSON structure that be used to create new Jira issues."),
			gollm.WithOutput("JSON"),
			gollm.WithContext(knowledge),
			gollm.WithExamples([]string{string(jsonResp)}...),
		),
	)

	prompt, err := templatePrompt.Execute(map[string]interface{}{
		"TargetIssueType": targetIssueTypeName,
	})
	if err != nil {
		log.Fatalf("Failed to execute prompt template: %v", err)
	}

	log.Println(prompt.String())

	ctx := context.Background()
	templateResponse, err := llmNorm.GenerateJSON(ctx, prompt)
	if err != nil {
		log.Fatalf("Failed to generate template response: %v", err)
	}

	picked := Answer{}
	err = json.Unmarshal([]byte(templateResponse.Stdout), &picked)
	if err != nil {
		return picked, err
	}

	err = picked.Validate()
	if err != nil {
		return picked, err
	}

	return picked, nil
}
