package utils

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/andrejsstepanovs/andai/pkg/models"
	redminemodels "github.com/andrejsstepanovs/andai/pkg/redmine/models"
	"github.com/mattn/go-redmine"
)

const (
	tmpFile = "andai-%d-*.md"
)

func BuildPromptTmpFile(issue redmine.Issue, step models.Step) (string, error) {
	promptTemplate := "{{.Step.Prompt}}"

	data := map[string]interface{}{
		"Step": step,
	}
	tmpl, err := template.New("SimplePrompt").Parse(promptTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	content := buf.String()
	fmt.Println("\n##############\n", content, "\n##############\n")

	tempFile, err := os.CreateTemp("", fmt.Sprintf(tmpFile, issue.Id))
	if err != nil {
		return "", err
	}
	log.Printf("Created temporary file: %q", tempFile.Name())

	_, err = tempFile.WriteString(content)
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	return tempFile.Name(), nil
}

func BuildIssueTmpFile(issue redmine.Issue, comments redminemodels.Comments, step models.Step) (string, error) {
	if !step.Context.Has(models.ContextComments) &&
		!step.Context.Has(models.ContextLastComment) &&
		!step.Context.Has(models.ContextTicket) {
		return "", nil
	}

	var err error
	if step.Context.Has(models.ContextLastComment) {
		if err != nil {
			log.Printf("Failed to get last comment: %v", err)
			return "", err
		}
		if len(comments) > 0 {
			comments = redminemodels.Comments{comments[len(comments)-1]}
			comments[0].Number = 1
		}
	} else if !step.Context.Has(models.ContextComments) {
		comments = nil
	}

	var promptTemplate string
	if step.Context.Has(models.ContextTicket) {
		promptTemplate = "# {{.Issue.Subject}} (ID: {{.Issue.Id}})\n\n" +
			"## Description\n" +
			"{{.Issue.Description}}\n\n" +
			"{{ if .Comments }}" +
			"## Comments:\n" +
			"{{ range .Comments }}\n" +
			"### Comment {{.Number}} (Created: {{.CreatedAt}})\n" +
			"{{.Text}}\n" +
			"{{ end }}" +
			"{{ end }}\n\n" +
			"# Your Task:\n{{.Step.Prompt}}"
	} else {
		promptTemplate = "" +
			"{{ range .Comments }}\n" +
			"### Comment {{.Number}} (Created: {{.CreatedAt}})\n" +
			"{{.Text}}" +
			"{{ end }}"
	}

	data := map[string]interface{}{
		"Issue":       issue,
		"Title":       issue.Subject,
		"Description": issue.Description,
		"Comments":    comments,
		"Step":        step,
	}

	tmpl, err := template.New("JiraIssue").Parse(promptTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	content := buf.String()
	//fmt.Println("\n##############\n", content, "\n##############\n")

	tempFile, err := os.CreateTemp("", fmt.Sprintf(tmpFile, issue.Id))
	if err != nil {
		return "", err
	}
	log.Printf("Created temporary file: %q", tempFile.Name())

	_, err = tempFile.WriteString(content)
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	return tempFile.Name(), nil
}
