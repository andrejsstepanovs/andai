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

type Knowledge struct {
	Issue      redmine.Issue
	Parent     *redmine.Issue
	Parents    []redmine.Issue
	Children   []redmine.Issue
	Project    models.Project
	IssueTypes models.IssueTypes
	Comments   redminemodels.Comments
	Step       models.Step
}

func (k Knowledge) BuildPromptTmpFile() (string, error) {
	promptTemplate := "{{.Step.Prompt}}"

	data := map[string]interface{}{
		"Step": k.Step,
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

	tempFile, err := os.CreateTemp("", fmt.Sprintf(tmpFile, k.Issue.Id))
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

func (k Knowledge) BuildIssueKnowledgeTmpFile() (string, error) {
	parts := make([]string, 0)
	for _, context := range k.Step.Context {
		newPart, err := k.getContext(context)
		if err != nil {
			return "", err
		}
		if newPart != "" {
			parts = append(parts, newPart)
		}
	}

	prompt := k.Step.Prompt.ForCli()
	if prompt != "" {
		txt := fmt.Sprintf("# Your task:\n%s", prompt)
		parts = append(parts, txt)
	}

	content := strings.Join(parts, "\n\n")
	//fmt.Println("\n##############\n", content, "\n##############\n")

	//panic(1)

	tempFile, err := os.CreateTemp("", fmt.Sprintf(tmpFile, k.Issue.Id))
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

func (k Knowledge) getContext(context string) (string, error) {
	switch context {
	case models.ContextLastComment:
		if len(k.Comments) == 0 {
			return "", nil
		}
		c := k.Comments[len(k.Comments)-1]
		return k.getComments(redminemodels.Comments{c}, "comment")
	case models.ContextComments:
		return k.getComments(k.Comments, "comments")
	case models.ContextTicket:
		return k.getIssue()
	case models.ContextProject:
		return k.getProject()
	case models.ContextProjectWiki:
		return getProjectWiki(k.Project)
	case models.ContextParent:
		return getParent(k.Parent, k.IssueTypes)
	case models.ContextParents:
		return getParents(k.Parents, k.IssueTypes)
	case models.ContextChildren:
		return getChildren(k.Children, k.IssueTypes)
	case models.ContextIssueTypes:
		return getIssueTypes(k.IssueTypes)
	default:
		return "", fmt.Errorf("unknown context: %q", context)
	}
}

func (k Knowledge) getIssue() (string, error) {
	issueContext, err := getIssueContext(k.Issue, k.IssueTypes)
	if err != nil {
		log.Printf("Failed to get current issue context: %v", err)
		return "", err
	}
	issueContext = tagContent("current_issue", issueContext, 1)
	return issueContext, nil
}

func (k Knowledge) getComments(comments redminemodels.Comments, tag string) (string, error) {
	if len(comments) == 0 {
		return "", nil
	}
	commentsContext, err := k.getCommentsContext(comments)
	if err != nil {
		log.Printf("Failed to get comments context: %v", err)
		return "", err
	}
	commentsContext = tagContent(tag, commentsContext, 1)
	return commentsContext, nil
}

func getProjectWiki(project models.Project) (string, error) {
	issueContext, err := getProjectWikiContext(project)
	if err != nil {
		log.Printf("Failed to get project wiki context: %v", err)
		return "", err
	}
	issueContext = tagContent("project_wiki", issueContext, 1)
	return issueContext, nil
}

func (k Knowledge) getProject() (string, error) {
	issueContext, err := k.getProjectContext()
	if err != nil {
		log.Printf("Failed to get project context: %v", err)
		return "", err
	}
	issueContext = tagContent("project", issueContext, 1)
	return issueContext, nil
}

func getParent(parent *redmine.Issue, issueTypes models.IssueTypes) (string, error) {
	if parent == nil || parent.Id == 0 {
		return "", nil
	}
	issueContext, err := getIssueContext(*parent, issueTypes)
	if err != nil {
		log.Printf("Failed to get parent issue context: %v", err)
		return "", err
	}
	issueContext = fmt.Sprintf("<parent_issue>\n%s\n</parent_issue>", issueContext)
	return issueContext, nil
}

func getParents(parents []redmine.Issue, issueTypes models.IssueTypes) (string, error) {
	if len(parents) == 0 {
		return "", nil
	}
	parentsContext := make([]string, 0)
	for _, p := range parents {
		parentIssueContext, err := getIssueContext(p, issueTypes)
		if err != nil {
			log.Printf("Failed to get single parent issue context: %v", err)
			return "", err
		}
		txt := tagContent("parent_issue", parentIssueContext, 2)
		parentsContext = append(parentsContext, txt)
	}
	issueContext := tagContent("parent_issues", strings.Join(parentsContext, "\n"), 1)
	return issueContext, nil
}

func getChildren(children []redmine.Issue, issueTypes models.IssueTypes) (string, error) {
	if len(children) == 0 {
		return "", nil
	}
	childrenContext := make([]string, 0)
	for _, child := range children {
		childIssueContext, err := getIssueContext(child, issueTypes)
		if err != nil {
			log.Printf("Failed to get single child issue context: %v", err)
			return "", err
		}
		txt := tagContent("child_issue", childIssueContext, 2)
		childrenContext = append(childrenContext, txt)
	}
	issueContext := tagContent("children_issues", strings.Join(childrenContext, "\n"), 1)
	return issueContext, nil
}

func getIssueTypes(issueTypes models.IssueTypes) (string, error) {
	if len(issueTypes) == 0 {
		return "", nil
	}
	issueTypeContext, err := getIssueTypesContext(issueTypes)
	if err != nil {
		log.Printf("Failed to get issue types context: %v", err)
		return "", err
	}
	issueContext := tagContent("project_issue_types", issueTypeContext, 1)
	return issueContext, nil
}

func (k Knowledge) getCommentsContext(comments redminemodels.Comments) (string, error) {
	promptTemplate := "{{ range .Comments }}" +
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

func getIssueContext(issue redmine.Issue, issueTypes models.IssueTypes) (string, error) {
	promptTemplate := "# Title: {{.Issue.Subject}}\n" +
		"- ID: {{.Issue.Id}}\n" +
		"- Issue Type: {{.IssueType.Name}}\n\n" +
		"# Description\n" +
		"{{.Issue.Description}}"

	data := map[string]interface{}{
		"Issue":     issue,
		"IssueType": issueTypes.Get(models.IssueTypeName(issue.Tracker.Name)),
	}

	tmpl, err := template.New("Issue").Parse(promptTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	return buf.String(), err
}

func (k Knowledge) getProjectContext() (string, error) {
	promptTemplate := "# {{.Project.Name}} (Identifier: {{.Project.Identifier}})\n\n" +
		"## Description\n" +
		"{{.Project.Description}}"

	data := map[string]interface{}{
		"Project": k.Project,
	}

	tmpl, err := template.New("Project").Parse(promptTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	return buf.String(), err
}

// todo use wiki content from db or api
func getProjectWikiContext(project models.Project) (string, error) {
	promptTemplate := "# Project Wiki page:\n" +
		"{{.Project.Wiki}}"

	data := map[string]interface{}{
		"Project": project,
	}

	tmpl, err := template.New("ProjectWiki").Parse(promptTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	return strings.Trim(buf.String(), "\n"), err
}

func getIssueTypesContext(issueTypes models.IssueTypes) (string, error) {
	promptTemplate := "{{ range .IssueTypes }}" +
		"- Issue Type \"{{.Name}}\": {{.Description}}\n" +
		"{{ end }}"

	data := map[string]interface{}{
		"IssueTypes": issueTypes,
	}

	tmpl, err := template.New("IssueTypes").Parse(promptTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	return strings.Trim(buf.String(), "\n"), err
}

func tagContent(tagName, content string, tabs int) string {
	return fmt.Sprintf(
		"<%s>\n%s\n</%s>",
		tagName,
		strings.Trim(tabContent(content, tabs), "\n"),
		tagName,
	)
}

func tabContent(content string, tabCount int) string {
	tabs := strings.Repeat("\t", tabCount)
	return tabs + strings.ReplaceAll(content, "\n", "\n"+tabs)
}
