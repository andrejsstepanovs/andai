package employee

import (
	"fmt"
	"log"
	"strings"

	"github.com/andrejsstepanovs/andai/internal/exec"
	redminemodels "github.com/andrejsstepanovs/andai/internal/redmine/models"
	"github.com/andrejsstepanovs/andai/internal/settings"
	"github.com/mattn/go-redmine"
)

func (i *Routine) getParentComments() (redminemodels.Comments, error) {
	if !i.parentExists() {
		return nil, fmt.Errorf("no parent issue")
	}
	comments, err := i.model.DBGetComments(i.parent.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments err: %v", err)
	}

	return comments, nil
}

func (i *Routine) getComments() (redminemodels.Comments, error) {
	comments, err := i.model.DBGetComments(i.issue.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments err: %v", err)
	}

	return comments, nil
}

func (i *Routine) getSiblingsComments(siblings []redmine.Issue) (redminemodels.Comments, error) {
	var allComments redminemodels.Comments
	for _, sibling := range siblings {
		comments, err := i.model.DBGetComments(sibling.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to get comments for sibling %d err: %v", sibling.Id, err)
		}
		allComments = append(allComments, comments...)
	}
	return allComments, nil
}

func (i *Routine) AddCommentToParent(text string) error {
	if i.parent == nil {
		return fmt.Errorf("no parent issue")
	}
	err := i.model.Comment(*i.parent, text)
	if err != nil {
		return fmt.Errorf("failed to comment issue err: %v", err)
	}
	return nil
}

func (i *Routine) AddComment(text string) error {
	err := i.model.Comment(i.issue, text)
	if err != nil {
		return fmt.Errorf("failed to comment issue err: %v", err)
	}
	return nil
}

func (i *Routine) RememberOutput(step settings.Step, output exec.Output) {
	logCommand := fmt.Sprintf("%s %s", step.Command, step.Action)
	if output.Stdout != "" {
		format := "Command: **%s**\n<result>\n%s\n</result>"
		msg := fmt.Sprintf(format, logCommand, output.Stdout)
		if step.Comment {
			err := i.AddComment(msg)
			if err != nil {
				log.Printf("Failed to add stdout comment: %v", err)
				panic(err)
			}
		}
		if step.Remember {
			i.history = append(i.history, msg)
		}
	}
	if output.Stderr != "" {
		if strings.Contains(output.Stderr, "Scanning repo") {
			log.Println("OK stderr: Aider scanning repo")
			return
		}
		log.Printf("stderr: %s\n", output.Stderr)
		format := "Command: %s\n<error>\n%s\n</error>"
		msg := fmt.Sprintf(format, logCommand, output.Stderr)
		if step.Comment {
			if step.Comment {
				err := i.AddComment(msg)
				if err != nil {
					log.Printf("Failed to add stderr comment: %v", err)
					panic(err)
				}
			}
		}
		if step.Remember {
			i.history = append(i.history, msg)
		}
	}
}
