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

const (
	tmpFile = "andai-%d-*.md"
)

var (
	aiderArgs = []string{
		"--no-stream",
		"--yes",
		"--no-pretty",
		"--yes-always",
		"--no-gitignore",
		"--no-analytics",
		"--no-watch-files",
		"--no-suggest-shell-commands",
		"--no-fancy-input",
		"--no-show-release-notes",
		"--no-check-update",
		"--analytics-disable",
		"--no-detect-urls",
		"--no-show-model-warnings",
	}
	aiderCodeArgs = []string{
		"--git",
		"--no-auto-lint",
		"--no-auto-test",
	}
	aiderArchitectArgs = []string{
		"--architect",
		"--no-git",
		"--no-auto-commits",
		"--no-auto-lint",
		"--no-auto-test",
	}

	aiderArchitectParams = map[string]string{
		"--map-refresh": "auto", // auto,always,files,manual
	}
	aiderCodeParams = map[string]string{
		"--map-refresh": "auto", // auto,always,files,manual
	}
)

type Employee struct {
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
) *Employee {
	return &Employee{
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

func (i *Employee) Work() bool {
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

func (i *Employee) AddComment(text string) error {
	err := i.model.Comment(i.issue, text)
	if err != nil {
		return fmt.Errorf("failed to comment issue err: %v", err)
	}
	return nil
}

func (i *Employee) action(step models.Step) error {
	fmt.Println(step.Command, step.Action)
	switch step.Command {
	case "aider", "aid":
		switch step.Action {
		case "architect", "code":
			return i.aiderExecute(step)
		default:
			return fmt.Errorf("unknown %q action: %q", step.Command, step.Action)
		}
	default:
		return fmt.Errorf("unknown step command: %q", step.Command)
	}
}

func (i *Employee) aiderExecute(step models.Step) error {
	contextFile, err := i.buildIssueTmpFile(step)
	if err != nil {
		log.Printf("Failed to build issue tmp file: %v", err)
	}
	log.Printf("Context file: %q\n", contextFile)
	//defer os.Remove(contextFile)

	options := i.aiderCommand(contextFile, step)
	stdout, stderr, err := exec.Exec(step.Command, options)
	if err != nil {
		log.Printf("Failed to execute command: %v", err)
		return err
	}

	if stdout != "" {
		fmt.Printf("stdout: %s\n", stdout)
		err = i.AddComment(fmt.Sprintf("<result>%s</result>", stdout))
		if err != nil {
			log.Printf("Failed to add stdout comment: %v", err)
			return err
		}
	}
	if stderr != "" {
		log.Printf("stderr: %s\n", stderr)
		err = i.AddComment(fmt.Sprintf("<error>%s</error>", stderr))
		if err != nil {
			log.Printf("Failed to add stderr comment: %v", err)
			return err
		}
	}

	return nil
}

func (i *Employee) aiderCommand(contextFile string, step models.Step) string {
	var (
		params map[string]string
		args   []string
	)
	switch step.Action {
	case "architect":
		params = aiderArchitectParams
		args = aiderArchitectArgs
	case "code":
		params = aiderCodeParams
		args = aiderCodeArgs
	default:
		panic("unknown step action")
	}
	//params["--message"] = step.Prompt
	if contextFile != "" {
		params["--message-file"] = contextFile
	}

	paramsCli := make([]string, 0, len(params))
	for k, v := range params {
		paramsCli = append(paramsCli, fmt.Sprintf("%s=%q", k, v))
	}

	args = append(args, aiderArgs...)

	return fmt.Sprintf(
		"%s %s",
		strings.Join(args, " "),
		strings.Join(paramsCli, " "),
	)
}

func (i *Employee) buildIssueTmpFile(step models.Step) (string, error) {
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
		"{{ end }}\n\n" +
		"# Your Task:\n{{.Step.Prompt}}"

	data := map[string]interface{}{
		"Issue":       i.issue,
		"Title":       i.issue.Subject,
		"Description": i.issue.Description,
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

func (i *Employee) PrepareWorkplace() error {
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

func (i *Employee) changeDirectory() error {
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

func (i *Employee) checkoutBranch() error {
	err := i.git.CheckoutBranch(strconv.Itoa(i.issue.Id))
	if err != nil {
		return fmt.Errorf("failed to checkout branch err: %v", err)
	}
	return nil
}

func (i *Employee) getComments() (redminemodels.Comments, error) {
	comments, err := i.model.DBGetComments(i.issue.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments err: %v", err)
	}

	//log.Printf("Comments: %s", i.issue.Notes)
	//fmt.Printf("Comments: %d\n", len(comments))
	//fmt.Printf("%s\n", strings.Join(comments, "\n"))

	return comments, nil
}
