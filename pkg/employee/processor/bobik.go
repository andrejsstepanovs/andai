package processor

import (
	"fmt"

	"github.com/andrejsstepanovs/andai/pkg/exec"
	"github.com/andrejsstepanovs/andai/pkg/models"
)

func BobikExecute(promptFile string, step models.Step) (exec.Output, error) {
	format := "Use this file %s as a question and answer!"
	return exec.Exec(step.Command, step.Action, fmt.Sprintf(format, promptFile))
}
