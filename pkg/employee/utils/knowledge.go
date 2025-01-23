package utils

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
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
	fmt.Println("\n##############\n", content, "\n##############")

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
	parts := make([]string, 0)
	for _, context := range step.Context {
		switch context {
		case models.ContextLastComment:
			if len(comments) > 0 {
				c := comments[len(comments)-1]
				commentsContext, err := getCommentsContext(redminemodels.Comments{c})
				if err != nil {
					log.Printf("Failed to get comments context: %v", err)
					return "", err
				}
				parts = append(parts, commentsContext)
			}
		case models.ContextComments:
			commentsContext, err := getCommentsContext(comments)
			if err != nil {
				log.Printf("Failed to get comments context: %v", err)
				return "", err
			}
			parts = append(parts, commentsContext)
		case models.ContextTicket:
			issueContext, err := getIssueContext(issue)
			if err != nil {
				log.Printf("Failed to get issue context: %v", err)
				return "", err
			}
			parts = append(parts, issueContext)
		}
	}

	content := strings.Join(parts, "\n\n")
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

func getCommentsContext(comments redminemodels.Comments) (string, error) {
	promptTemplate := "{{ range .Comments }}\n" +
		"### Comment {{.Number}} (Created: {{.CreatedAt}})\n" +
		"{{.Text}}" +
		"{{ end }}"

	data := map[string]interface{}{
		"Comments": comments,
	}

	tmpl, err := template.New("Comments").Parse(promptTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	return buf.String(), err
}

func getIssueContext(issue redmine.Issue) (string, error) {
	promptTemplate := "# {{.Issue.Subject}} (ID: {{.Issue.Id}})\n\n" +
		"## Description\n" +
		"{{.Issue.Description}}\n"

	data := map[string]interface{}{
		"Issue": issue,
	}

	tmpl, err := template.New("Issue").Parse(promptTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	return buf.String(), err
}
