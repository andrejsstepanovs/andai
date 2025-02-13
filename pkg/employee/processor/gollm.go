package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/andrejsstepanovs/andai/pkg/ai"
	"github.com/andrejsstepanovs/andai/pkg/employee/utils"
	"github.com/andrejsstepanovs/andai/pkg/exec"
	"github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/mattn/go-redmine"
	"github.com/teilomillet/gollm"
)

func EvaluateOutcome(llm *ai.AI, knowledgeFile string) (exec.Output, bool, error) {
	knowledge, err := utils.GetFileContents(knowledgeFile)
	if err != nil {
		return exec.Output{}, false, err
	}

	templatePrompt := gollm.NewPromptTemplate("EvaluateOutcome", "",
		"Your task is to evaluate final outcome of the conversation. "+
			"It is either positive or negative. There is no in between.\n\n"+
			"# Instructions:\n"+
			"- Use Context and specifically last comment section to evaluate final outcome of the topic.\n"+
			"- It can be either positive or negative.\n"+
			"- In case of positive outcome, answer with 1 word \"Positive\".\n"+
			"- In case of negative outcome, answer with 1 word \"Negative\".\n"+
			"- Do not explain why you came to this conclusion or any other information about your thinking process.\n"+
			"- Answer with 1 word (\"Positive\" or \"Negative\")!\n",
		gollm.WithPromptOptions(
			gollm.WithOutput("1 word"),
			gollm.WithContext(knowledge),
		),
	)

	prompt, err := templatePrompt.Execute(map[string]interface{}{})
	if err != nil {
		return exec.Output{}, false, err
	}

	ctx := context.Background()

	out, err := llm.Generate(ctx, prompt)
	if err != nil {
		return exec.Output{}, false, err
	}

	if strings.Trim(strings.TrimSpace(out.Stdout), "\"") == "Positive" {
		return out, true, nil
	}
	return out, false, nil
}

func GenerateIssues(llm *ai.AI, targetIssueTypeName models.IssueTypeName, knowledgeFile string) (exec.Output, map[int]redmine.Issue, map[int][]int, error) {
	var (
		err              error
		query            string
		validationPrompt string
	)
	var createIssues Answer
	for i := 0; i < 5; i++ {
		createIssues, query, err = getIssues(llm, targetIssueTypeName, knowledgeFile, validationPrompt)
		if err != nil {
			return exec.Output{}, nil, nil, err
		}
		if query == "" {
			break
		}
		log.Println("Failed to validate. Trying again.")
		validationPrompt = fmt.Sprintf("Your last answer was not good: ----\n\n%s\n\n----. Try again and this time make sure your answer (JSON) is valid!", query)
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

func getIssues(llmNorm *ai.AI, targetIssueTypeName models.IssueTypeName, knowledgeFile, promptExend string) (Answer, string, error) {
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
		return Answer{}, "", err
	}

	knowledge, err := utils.GetFileContents(knowledgeFile)
	if err != nil {
		return Answer{}, "", err
	}

	templatePrompt := gollm.NewPromptTemplate("IssuePlanToJson", "",
		"You are software engineer working with Jira on single issue breakdown task. "+
			"Someone already thought about how to split current issue, use that info.\n\n"+
			"# Instructions:\n"+
			"- Use Context and specifically comments section to convert proposed issues into JSON data.\n"+
			"- Make sure to not make circular dependencies between issues.\n"+
			"- Convert suggested issues {{.TargetIssueType}} into specific format json data.\n"+
			"- Each element should contain: number_int (int), subject (text), description (text), blocked_by_numbers (array of integers).\n"+
			"- Where blocked_by_numbers is array of integers.\n"+
			"- Use example data structure for your answer.\n"+
			"- Do not use any other tags in JSON.\n\n"+
			ai.ForceJson+"\n"+promptExend,
		gollm.WithPromptOptions(
			gollm.WithDirectives("Convert given context content into issues as JSON structure that be used to create new Jira issues."),
			gollm.WithOutput("JSON"),
			gollm.WithContext(knowledge),
			gollm.WithExamples([]string{"\n```\n" + string(jsonResp) + "\n```\n"}...),
		),
	)

	prompt, err := templatePrompt.Execute(map[string]interface{}{
		"TargetIssueType": targetIssueTypeName,
	})
	if err != nil {
		return Answer{}, "", err
	}

	ctx := context.Background()

	picked := Answer{}
	_, validationErr, err := llmNorm.GenerateJSON(ctx, prompt, &picked)
	if err != nil {
		return Answer{}, "", err
	}
	if validationErr != nil {
		return Answer{}, validationErr.Error(), nil
	}

	err = picked.Validate()
	if err != nil {
		return picked, err.Error(), err
	}

	return picked, "", nil
}
