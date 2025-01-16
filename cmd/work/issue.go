package work

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/andrejsstepanovs/andai/pkg/exec"
	"github.com/andrejsstepanovs/andai/pkg/llm"
	"github.com/andrejsstepanovs/andai/pkg/models"
	model "github.com/andrejsstepanovs/andai/pkg/redmine"
	redminemodels "github.com/andrejsstepanovs/andai/pkg/redmine/models"
	"github.com/andrejsstepanovs/andai/pkg/worker"
	"github.com/mattn/go-redmine"
)

const tmpFile = "andai-%d-*.md"

type WorkOnIssue struct {
	model      *model.Model
	llm        *llm.LLM
	issue      redmine.Issue
	project    redmine.Project
	projectCfg models.Project
	git        *worker.Git
	state      models.State
	issueType  models.IssueType
	job        models.Job
	workingDir string
}

func NewWorkOnIssue(
	model *model.Model,
	llm *llm.LLM,
	issue redmine.Issue,
	project redmine.Project,
	projectCfg models.Project,
	git *worker.Git,
	state models.State,
	issueType models.IssueType,
) *WorkOnIssue {
	return &WorkOnIssue{
		model:      model,
		llm:        llm,
		issue:      issue,
		project:    project,
		projectCfg: projectCfg,
		git:        git,
		state:      state,
		issueType:  issueType,
		job:        issueType.Jobs.Get(models.StateName(issue.Status.Name)),
	}
}

func (i *WorkOnIssue) Work() bool {
	log.Printf("Working on %q %q (ID: %d)", i.state.Name, i.issueType.Name, i.issue.Id)

	err := i.PrepareWorkplace()
	if err != nil {
		log.Printf("Failed to prepare workplace: %v", err)
		return false
	}

	fmt.Printf("Total steps: %d\n", len(i.job.Steps))
	for j, step := range i.job.Steps {
		fmt.Printf("Step: %d\n", j+1)
		err = i.action(step)
		if err != nil {
			log.Printf("Failed to action step: %v", err)
			return false
		}
		fmt.Println("Success")
	}

	return true
}

func (i *WorkOnIssue) AddComment(text string) error {
	err := i.model.Comment(i.issue, text)
	if err != nil {
		return fmt.Errorf("failed to comment issue err: %v", err)
	}
	return nil
}

func (i *WorkOnIssue) action(step models.Step) error {
	switch step.Command {
	case "aider", "aid":
		switch step.Action {
		case "architect":
			return i.aiderArchitect(step)
		case "code":
			return i.aiderCode(step)
		default:
			return fmt.Errorf("unknown %q action: %q", step.Command, step.Action)
		}
	default:
		return fmt.Errorf("unknown step command: %q", step.Command)
	}
}

func (i *WorkOnIssue) aiderCode(step models.Step) error {
	fmt.Println(step.Command, step.Action)

	//stdout, stderr, err := exec.Exec("", "once quiet llm", "hi!")
	//stdout, stderr, err := exec.Exec("bobik", "once quiet llm", "hi!")

	return nil
}

func (i *WorkOnIssue) aiderArchitect(step models.Step) error {
	fmt.Println(step.Command, step.Action)

	contextFile, err := i.buildIssueTmpFile()
	if err != nil {
		log.Printf("Failed to build issue tmp file: %v", err)
	}
	defer os.Remove(contextFile)

	args := []string{
		"--architect",
		"--no-pretty",
		"--no-stream",
		"--yes",
		"--no-git",
		"--no-gitignore",
		"--subtree-only",
		"--no-auto-commits",
		"--no-watch-files",
		"--no-auto-lint",
		"--no-auto-test",
		"--no-analytics",
		"--analytics-disable",
		"--yes-always",
		"--no-suggest-shell-commands",
		"--no-fancy-input",
	}
	params := map[string]string{
		"--map-refresh": "auto", // auto,always,files,manual
		"--message":     step.Prompt,
	}
	if contextFile != "" {
		params["--message-file"] = contextFile
	}

	paramsCli := ""
	for k, v := range params {
		paramsCli += fmt.Sprintf("%s=%q", k, v)
	}

	options := fmt.Sprintf("%s %s", strings.Join(args, " "), paramsCli)

	stdout, stderr, err := exec.Exec(step.Command, options)
	if err != nil {
		log.Printf("Failed to execute command: %v", err)
		return err
	}
	fmt.Printf("stdout: %s", stdout)
	fmt.Printf("stdERR: %s", stderr)

	if stdout != "" {
		err = i.AddComment(fmt.Sprintf("<result>%s</result>", stdout))
	}
	if stderr != "" {
		err = i.AddComment(fmt.Sprintf("<error>%s</error>", stderr))
	}

	return nil
}

func (i *WorkOnIssue) buildIssueTmpFile() (string, error) {
	comments, err := i.getComments()
	if err != nil {
		log.Printf("Failed to get comments: %v", err)
		return "", err
	}

	promptTemplate := "# {{.Issue.Subject}} (ID: {{.Issue.Id}})\n\n" +
		"## Description\n" +
		"{{.Issue.Description}}\n\n" +
		"## Comments:\n" +
		"{{ range .Comments }}\n" +
		"### Comment {{.Number}} (Created: {{.CreatedAt}})\n" +
		"{{.Text}}\n" +
		"{{ end }}"

	data := map[string]interface{}{
		"Issue":       i.issue,
		"Title":       i.issue.Subject,
		"Description": i.issue.Description,
		"Comments":    comments,
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

	tempFile, err := os.CreateTemp("", fmt.Sprintf(tmpFile, i.issue.Id))
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

func (i *WorkOnIssue) PrepareWorkplace() error {
	err := i.changeDirectory()
	if err != nil {
		log.Printf("Failed to change directory: %v", err)
		return err
	}
	err = i.checkoutBranch()
	if err != nil {
		log.Printf("Failed to checkout branch: %v", err)
		return err
	}
	return nil
}

func (i *WorkOnIssue) changeDirectory() error {
	targetPath := i.git.Path
	if filepath.Base(targetPath) == ".git" {
		targetPath = filepath.Dir(targetPath)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory err: %v", err)
	}
	if currentDir != targetPath {
		log.Printf("Changing directory from %s to %s\n", currentDir, targetPath)
	}

	err = os.Chdir(targetPath)
	if err != nil {
		return fmt.Errorf("failed to change directory err: %v", err)
	}

	log.Printf("Active in project directory %s\n", targetPath)
	i.workingDir = targetPath

	return nil
}

func (i *WorkOnIssue) checkoutBranch() error {
	err := i.git.CheckoutBranch(strconv.Itoa(i.issue.Id))
	if err != nil {
		return fmt.Errorf("failed to checkout branch err: %v", err)
	}
	return nil
}

func (i *WorkOnIssue) getComments() (redminemodels.Comments, error) {
	comments, err := i.model.DBGetComments(i.issue.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments err: %v", err)
	}

	//log.Printf("Comments: %s", i.issue.Notes)
	//fmt.Printf("Comments: %d\n", len(comments))
	//fmt.Printf("%s\n", strings.Join(comments, "\n"))

	return comments, nil
}
