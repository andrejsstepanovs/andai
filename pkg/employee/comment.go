package employee

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/pkg/exec"
	"github.com/andrejsstepanovs/andai/pkg/models"
	redminemodels "github.com/andrejsstepanovs/andai/pkg/redmine/models"
)

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

func (i *Employee) AddCommentToParent(text string) error {
	if i.parent == nil {
		return fmt.Errorf("no parent issue")
	}
	err := i.model.Comment(*i.parent, text)
	if err != nil {
		return fmt.Errorf("failed to comment issue err: %v", err)
	}
	return nil
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
