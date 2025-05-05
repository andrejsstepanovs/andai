package knowledge

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/andrejsstepanovs/andai/internal/exec"
	redminemodels "github.com/andrejsstepanovs/andai/internal/redmine/models"
	"github.com/andrejsstepanovs/andai/internal/settings"
	"github.com/mattn/go-redmine"
)

const (
	tmpFile = "andai-%d-*.md"
)

type Knowledge struct {
	Issue             redmine.Issue
	Parent            *redmine.Issue
	Parents           []redmine.Issue
	ClosedChildrenIDs []int
	Children          []redmine.Issue // not closed children
	Siblings          []redmine.Issue
	Workbench         *exec.Workbench
	Project           settings.Project
	IssueTypes        settings.IssueTypes
	Comments          redminemodels.Comments
	Step              settings.Step
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
	//fmt.Println("\n##############\n", content, "\n##############")

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

	if len(k.Step.History) > 0 {
		txt := fmt.Sprintf("# Additional Info:\n%s\n", strings.Join(k.Step.History, "\n"))
		parts = append(parts, txt)
	}

	//if len(k.Step.ContextFiles) > 0 {
	//	for _, file := range k.Step.ContextFiles {
	//		txt := fmt.Sprintf("# File: %s\n", file)
	//		parts = append(parts, txt)
	//	}
	//}

	prompt := k.Step.Prompt.ForCli()
	if prompt != "" {
		txt := fmt.Sprintf("# Your task:\n%s", prompt)
		parts = append(parts, txt)
	}

	if len(parts) == 0 {
		return "", nil
	}

	content := strings.Join(parts, "\n\n")

	tempFile, err := os.CreateTemp("", fmt.Sprintf(tmpFile, k.Issue.Id))
	if err != nil {
		return "", err
	}
	log.Printf("Created temporary file: %q", tempFile.Name())

	_, err = tempFile.WriteString(content)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

func (k Knowledge) getLastNComments(n int, tag string) (string, error) {
	if len(k.Comments) == 0 {
		return "", nil
	}
	count := 1
	for i := 2; i <= n; i++ {
		if len(k.Comments) >= i {
			count = i
		}
	}
	lastComments := k.Comments[len(k.Comments)-count:]
	return k.getComments(lastComments, tag)
}

func (k Knowledge) getCommentContext(context string) (string, error) {
	switch context {
	case settings.ContextLastComment:
		if len(k.Comments) == 0 {
			return "", nil
		}
		return k.getComments(redminemodels.Comments{k.Comments[len(k.Comments)-1]}, settings.ContextLastComment)
	case settings.ContextTwoComment:
		return k.getLastNComments(2, settings.ContextTwoComment)
	case settings.ContextThreeComment:
		return k.getLastNComments(3, settings.ContextThreeComment)
	case settings.ContextFourComment:
		return k.getLastNComments(4, settings.ContextFourComment)
	case settings.ContextFifeComment:
		return k.getLastNComments(5, settings.ContextFifeComment)
	case settings.ContextComments:
		return k.getComments(k.Comments, settings.ContextComments)
	default:
		return "", fmt.Errorf("unknown comment context: %q", context)
	}
}

func (k Knowledge) getContext(context string) (string, error) {
	// Handle comment-related contexts
	if strings.Contains(strings.ToLower(context), "comment") {
		return k.getCommentContext(context)
	}

	// Handle other contexts
	switch context {
	case settings.ContextTicket:
		return k.getIssue()
	case settings.ContextProject:
		return k.getProject()
	case settings.ContextProjectWiki:
		return k.getProjectWiki()
	case settings.ContextParent:
		return k.getParent()
	case settings.ContextParents:
		return k.getParents()
	case settings.ContextSiblings:
		return k.getSiblings()
	case settings.ContextChildren:
		return k.getChildren()
	case settings.ContextAffectedFiles:
		return k.getChangedFiles()
	case settings.ContextIssueTypes:
		return k.getIssueTypes()
	default:
		return "", fmt.Errorf("unknown context: %q", context)
	}
}

func (k Knowledge) getIssue() (string, error) {
	issueContext, err := k.getIssueContext(k.Issue)
	if err != nil {
		log.Printf("Failed to get current issue context: %v", err)
		return "", err
	}
	issueContext = k.TagContent("current_issue", issueContext, 1)
	return issueContext, nil
}

func (k Knowledge) getComments(comments redminemodels.Comments, tag string) (string, error) {
	if len(comments) == 0 {
		return "", nil
	}
	cleanComments := make(redminemodels.Comments, 0)
	kickCommentsIfMatch := [][]string{
		{"Merged", "branch", "into"},
		{"Branch", exec.BranchPrefix, "Commit"},
	}
	for _, c := range comments {
		kickComment := false
		for _, a := range kickCommentsIfMatch {
			allFound := true
			for _, s := range a {
				if !strings.Contains(c.Text, s) {
					allFound = false
					break
				}
			}
			if allFound {
				kickComment = true
				break
			}
		}

		if kickComment {
			continue
		}
		cleanComments = append(cleanComments, c)
	}

	commentsContext, err := k.getCommentsContext(cleanComments)
	if err != nil {
		log.Printf("Failed to get comments context: %v", err)
		return "", err
	}
	commentsContext = k.TagContent(tag, commentsContext, 1)
	return commentsContext, nil
}

func (k Knowledge) getProjectWiki() (string, error) {
	issueContext, err := k.getProjectWikiContext()
	if err != nil {
		log.Printf("Failed to get project wiki context: %v", err)
		return "", err
	}
	issueContext = k.TagContent("project_wiki", issueContext, 1)
	return issueContext, nil
}

func (k Knowledge) getProject() (string, error) {
	issueContext, err := k.getProjectContext()
	if err != nil {
		log.Printf("Failed to get project context: %v", err)
		return "", err
	}
	issueContext = k.TagContent("project", issueContext, 1)
	return issueContext, nil
}

func (k Knowledge) getParent() (string, error) {
	if k.Parent == nil || k.Parent.Id == 0 {
		return "", nil
	}
	issueContext, err := k.getIssueContext(*k.Parent)
	if err != nil {
		log.Printf("Failed to get parent issue context: %v", err)
		return "", err
	}
	issueContext = fmt.Sprintf("<parent_issue>\n%s\n</parent_issue>", issueContext)
	return issueContext, nil
}

func (k Knowledge) getParents() (string, error) {
	if len(k.Parents) == 0 {
		return "", nil
	}
	parentsContext := make([]string, 0)
	for _, p := range k.Parents {
		parentIssueContext, err := k.getIssueContext(p)
		if err != nil {
			log.Printf("Failed to get single parent issue context: %v", err)
			return "", err
		}
		txt := k.TagContent("parent_issue", parentIssueContext, 2)
		parentsContext = append(parentsContext, txt)
	}
	issueContext := k.TagContent("parent_issues", strings.Join(parentsContext, "\n"), 1)
	return issueContext, nil
}

func (k Knowledge) getSiblings() (string, error) {
	if len(k.Siblings) == 0 {
		return "", nil
	}
	siblingsContext := make([]string, 0)
	for _, sibling := range k.Siblings {
		siblingIssueContext, err := k.getIssueContext(sibling)
		if err != nil {
			log.Printf("Failed to get single child issue context: %v", err)
			return "", err
		}
		txt := k.TagContent("sibling_issue", siblingIssueContext, 2)
		siblingsContext = append(siblingsContext, txt)
	}
	issueContext := k.TagContent("sibling_issues", strings.Join(siblingsContext, "\n"), 1)
	return issueContext, nil
}

func (k Knowledge) getChildren() (string, error) {
	if len(k.Children) == 0 {
		return "", nil
	}
	childrenContext := make([]string, 0)
	for _, child := range k.Children {
		childIssueContext, err := k.getIssueContext(child)
		if err != nil {
			log.Printf("Failed to get single child issue context: %v", err)
			return "", err
		}
		txt := k.TagContent("child_issue", childIssueContext, 2)
		childrenContext = append(childrenContext, txt)
	}
	issueContext := k.TagContent("children_issues", strings.Join(childrenContext, "\n"), 1)
	return issueContext, nil
}

// getChangedFiles returns files that were changed in the branch (not perfect yet, relies on count of closed children)
func (k Knowledge) getChangedFiles() (string, error) {
	if len(k.ClosedChildrenIDs) == 0 {
		return "", nil
	}
	commits, err := k.Workbench.GetBranchCommits(len(k.ClosedChildrenIDs))
	if err != nil {
		log.Printf("Failed to get branch commits: %v", err)
		return "", err
	}

	filesMap := make(map[string]struct{})
	for _, commit := range commits {
		files, err := k.Workbench.GetAffectedFiles(commit)
		if err != nil {
			log.Printf("Failed to get affected files for commit %q: %v", commit, err)
			return "", err
		}
		for _, file := range files {
			filesMap[file] = struct{}{}
		}
	}

	files := make([]string, 0)
	for file := range filesMap {
		files = append(files, file)
	}

	if len(files) == 0 {
		return "", nil
	}

	filesContext := strings.Join(files, "\n")

	return k.TagContent("affected_files", filesContext, 1), nil
}

func (k Knowledge) getIssueTypes() (string, error) {
	if len(k.IssueTypes) == 0 {
		return "", nil
	}
	issueTypeContext, err := k.getIssueTypesContext()
	if err != nil {
		log.Printf("Failed to get issue types context: %v", err)
		return "", err
	}
	issueContext := k.TagContent("project_issue_types", issueTypeContext, 1)
	return issueContext, nil
}

func (k Knowledge) getCommentsContext(comments redminemodels.Comments) (string, error) {
	promptTemplate := "{{ range .Comments }}" +
		"\n<comment_{{.Number}} at=\"{{.CreatedAt}}\">\n" +
		"{{.Text}}" +
		"\n</comment_{{.Number}}>" +
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

func (k Knowledge) getIssueContext(issue redmine.Issue) (string, error) {
	promptTemplate := "# Title: {{.Issue.Subject}}\n" +
		"- ID: {{.Issue.Id}}\n" +
		"- Issue Type: {{.IssueType.Name}}\n\n" +
		"# Description\n" +
		"{{.Issue.Description}}"

	data := map[string]interface{}{
		"Issue":     issue,
		"IssueType": k.IssueTypes.Get(settings.IssueTypeName(issue.Tracker.Name)),
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
func (k Knowledge) getProjectWikiContext() (string, error) {
	promptTemplate := "# Project Wiki page:\n" +
		"{{.Project.Wiki}}"

	data := map[string]interface{}{
		"Project": k.Project,
	}

	tmpl, err := template.New("ProjectWiki").Parse(promptTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	return strings.Trim(buf.String(), "\n"), err
}

func (k Knowledge) getIssueTypesContext() (string, error) {
	promptTemplate := "{{ range .IssueTypes }}" +
		"- Issue Type \"{{.Name}}\": {{.Description}}\n" +
		"{{ end }}"

	data := map[string]interface{}{
		"IssueTypes": k.IssueTypes,
	}

	tmpl, err := template.New("IssueTypes").Parse(promptTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	return strings.Trim(buf.String(), "\n"), err
}

func (k Knowledge) TagContent(tagName, content string, tabs int) string {
	tabsStr := strings.Repeat("\t", tabs)
	content = tabsStr + strings.ReplaceAll(content, "\n", "\n"+tabsStr)

	return fmt.Sprintf(
		"<%s>\n%s\n</%s>",
		tagName,
		strings.Trim(content, "\n"),
		tagName,
	)
}
