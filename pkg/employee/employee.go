package employee

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
	history    []exec.Output
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
		step.Prompt = i.addHistory(step)
		output, err := i.processStep(step)
		if err != nil {
			log.Printf("Failed to action step: %v", err)
			return false
		}
		i.CommentOutput(step, output)
		if step.Remember {
			i.history = append(i.history, output)
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

func (i *Employee) CommentOutput(step models.Step, output exec.Output) {
	if !step.Comment {
		return
	}
	logCommand := fmt.Sprintf("%s %s", step.Command, step.Action)
	if output.Stdout != "" {
		format := "Command: **%s**\n<result>\n%s\n</result>"
		msg := fmt.Sprintf(format, logCommand, output.Stdout)
		err := i.AddComment(msg)
		if err != nil {
			log.Printf("Failed to add stdout comment: %v", err)
			panic(err)
		}
	}
	if output.Stderr != "" {
		log.Printf("stderr: %s\n", output.Stderr)
		if step.Comment {
			format := "Command: %s\n<error>\n%s\n</error>"
			msg := fmt.Sprintf(format, logCommand, output.Stderr)
			err := i.AddComment(msg)
			if err != nil {
				log.Printf("Failed to add stderr comment: %v", err)
				panic(err)
			}
		}
	}
}

func (i *Employee) processStep(step models.Step) (exec.Output, error) {
	fmt.Println(step.Command, step.Action)
	switch step.Command {
	case "git":
		return exec.Exec(step.Command, step.Action)
	case "aider", "aid":
		switch step.Action {
		case "architect", "code":
			return i.aiderExecute(step)
		default:
			return exec.Output{}, fmt.Errorf("unknown %q action: %q", step.Command, step.Action)
		}
	case "bobik":
		return i.bobikExecute(step)
	default:
		return exec.Output{}, fmt.Errorf("unknown step command: %q", step.Command)
	}
}

func (i *Employee) addHistory(step models.Step) models.StepPrompt {
	if len(i.history) == 0 {
		return step.Prompt
	}
	hydratedPrompt := make([]string, 0)
	for _, o := range i.history {
		hydratedPrompt = append(hydratedPrompt, o.AsPrompt())
	}
	hydratedPrompt = append(hydratedPrompt, string(step.Prompt))
	return models.StepPrompt(strings.Join(hydratedPrompt, "\n\n----\n\n"))
}

func (i *Employee) bobikExecute(step models.Step) (exec.Output, error) {
	promptFile, err := i.buildPromptTmpFile(step)
	if err != nil {
		log.Printf("Failed to build prompt tmp file: %v", err)
	}

	format := "Use this file %s as a question and answer!"
	return exec.Exec(step.Command, step.Action, fmt.Sprintf(format, promptFile))
}

func (i *Employee) aiderExecute(step models.Step) (exec.Output, error) {
	contextFile, err := i.buildIssueTmpFile(step)
	if err != nil {
		log.Printf("Failed to build issue tmp file: %v", err)
	}
	log.Printf("Context file: %q\n", contextFile)
	defer os.Remove(contextFile)

	options := exec.AiderCommand(contextFile, step)
	output, err := exec.Exec(step.Command, options)
	if err != nil {
		log.Printf("Failed to execute command: %v", err)
		return output, err
	}

	if output.Stdout != "" {
		fmt.Printf("stdout: %s\n", output.Stdout)

		lines := strings.Split(output.Stdout, "\n")
		startPos := 0
		lastPos := 0
		for k, line := range lines {
			if startPos == 0 && strings.Contains(line, "CONVENTIONS.md") &&
				strings.Contains(line, "Added") &&
				strings.Contains(line, "to the chat") {
				startPos = k
			}
			if lastPos == 0 && strings.Contains(line, "Tokens:") &&
				strings.Contains(line, "sent") &&
				strings.Contains(line, "received") {
				lastPos = k
			}
		}

		if startPos > 0 && lastPos > 0 {
			output.Stdout = strings.Join(lines[startPos+1:lastPos], "\n")
		}
	}

	return output, nil
}

func (i *Employee) buildPromptTmpFile(step models.Step) (string, error) {
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

func (i *Employee) buildIssueTmpFile(step models.Step) (string, error) {
	if !step.Context.Has(models.ContextComments) &&
		!step.Context.Has(models.ContextLastComment) &&
		!step.Context.Has(models.ContextTicket) {
		return "", nil
	}

	comments := redminemodels.Comments{}
	var err error
	if step.Context.Has(models.ContextComments) {
		comments, err = i.getComments()
		if err != nil {
			log.Printf("Failed to get comments: %v", err)
			return "", err
		}
	} else if step.Context.Has(models.ContextLastComment) {
		comments, err = i.getComments()
		if err != nil {
			log.Printf("Failed to get last comment: %v", err)
			return "", err
		}
		if len(comments) > 0 {
			comments = redminemodels.Comments{comments[len(comments)-1]}
			comments[0].Number = 1
		}
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
	//fmt.Println("\n##############\n", content, "\n##############\n")

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
