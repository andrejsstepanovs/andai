package ping

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/andrejsstepanovs/andai/internal"
	"github.com/andrejsstepanovs/andai/internal/ai"
	"github.com/andrejsstepanovs/andai/internal/employee/utils"
	"github.com/andrejsstepanovs/andai/internal/exec"
	"github.com/andrejsstepanovs/andai/internal/redmine"
	"github.com/andrejsstepanovs/andai/internal/settings"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newPingAiderCommand(deps *internal.AppDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aider",
		Short: "Ping aider connection",
		RunE: func(_ *cobra.Command, _ []string) error {
			settings, err := deps.Config.Load()
			if err != nil {
				return err
			}

			fmt.Println("Pinging Aider")
			err = pingAider(deps.Model, settings.Projects, settings.Aider)
			if err != nil {
				return err
			}
			log.Println("Aider OK")
			return nil
		},
	}
	return cmd
}

func pingAider(redmine *redmine.Model, projects settings.Projects, aider settings.Aider) error {
	redmineProjects, err := redmine.API().Projects()
	if err != nil {
		return fmt.Errorf("failed to get redmine project err: %v", err)
	}
	if len(redmineProjects) == 0 {
		return fmt.Errorf("no projects found")
	}
	project := redmineProjects[0]
	log.Printf("Project %s (%s)", project.Name, project.Identifier)

	projectRepo, err := redmine.DBGetRepository(project)
	if err != nil {
		return fmt.Errorf("failed to get redmine repository err: %v", err)
	}

	projectConfig := projects.Find(project.Identifier)
	git, err := exec.FindProjectGit(projectConfig, projectRepo)
	if err != nil {
		return fmt.Errorf("failed to find project git err: %v", err)
	}
	log.Printf("Project Repository Opened %s", git.GetPath())

	wb := &exec.Workbench{Git: git}
	step := settings.Step{
		Command: "aider",
		Action:  "architect",
		Prompt:  "Answer with OK.",
	}
	e := utils.Knowledge{
		Step:      step,
		Workbench: wb,
	}

	err = wb.GoToRepo()
	if err != nil {
		return fmt.Errorf("failed to go to repository: %v", err)
	}

	contextFile, err := e.BuildPromptTmpFile()
	if err != nil {
		log.Println("Failed to create context file")
		return fmt.Errorf("error creating context file: %v", err)
	}

	options := exec.AiderCommand(contextFile, step, aider)
	output, err := exec.Exec(step.Command, time.Minute*5, options)
	if err != nil {
		log.Printf("Failed to execute command: %v", err)
		return fmt.Errorf("error executing command: %v", err)
	}
	output = ai.RemoveThinkingFromOutput(output)

	log.Println("Response from Aider")
	//log.Println(output.Stdout) // "Tokens: 6.2k sent, 4 received."
	//log.Println(output.Stderr)

	text := output.Stdout

	required := []string{"Tokens:", "sent", "received"}
	for _, s := range required {
		if !strings.Contains(text, s) {
			return fmt.Errorf("missing required text: %s", s)
		}
	}

	// Extract values between keywords
	sent := extractBetween(text, "Tokens:", "sent")
	received := extractBetween(text, "sent,", "received")

	// Clean up the values
	sent = strings.TrimSpace(sent)
	received = strings.TrimSpace(received)

	fmt.Printf("Sent: %s\nReceived: %s\n", sent, received)

	return nil
}

func extractBetween(text, start, end string) string {
	startIndex := strings.Index(text, start) + len(start)
	if startIndex < len(start) {
		return ""
	}

	endIndex := strings.Index(text[startIndex:], end)
	if endIndex < 0 {
		return ""
	}

	return text[startIndex : startIndex+endIndex]
}
