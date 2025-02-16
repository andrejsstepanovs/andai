package ping

import (
	"fmt"
	"log"
	"strings"

	"github.com/andrejsstepanovs/andai/pkg/employee/utils"
	"github.com/andrejsstepanovs/andai/pkg/exec"
	"github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/andrejsstepanovs/andai/pkg/redmine"
	"github.com/andrejsstepanovs/andai/pkg/workbench"
	"github.com/andrejsstepanovs/andai/pkg/worker"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newPingAiderCommand(redmine *redmine.Model, projects models.Projects, aider models.Aider) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aider",
		Short: "Ping aider connection",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println("Pinging Aider")

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
			git, err := worker.FindProjectGit(projectConfig, projectRepo)
			if err != nil {
				return fmt.Errorf("failed to find project git err: %v", err)
			}
			log.Printf("Project Repository Opened %s", git.GetPath())

			wb := &workbench.Workbench{Git: git}
			step := models.Step{
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
			output, err := exec.Exec(step.Command, options)
			if err != nil {
				log.Printf("Failed to execute command: %v", err)
				return fmt.Errorf("error executing command: %v", err)
			}

			log.Println("Response from Aider")
			//log.Println(output.Stdout) // "Tokens: 6.2k sent, 4 received."
			//log.Println(output.Stderr)

			text := output.Stdout

			required := []string{"Tokens:", "sent", "received"}
			isValid := true
			for _, s := range required {
				if !strings.Contains(text, s) {
					isValid = false
					return fmt.Errorf("Missing required text: %s\n", s)
				}
			}

			if !isValid {
				log.Printf("stdout: %s", output.Stdout)
				log.Printf("sterr: %s", output.Stderr)
				return fmt.Errorf("invalid aider response")
			}

			if isValid {
				// Extract values between keywords
				sent := extractBetween(text, "Tokens:", "sent")
				received := extractBetween(text, "sent,", "received")

				// Clean up the values
				sent = strings.TrimSpace(sent)
				received = strings.TrimSpace(received)

				fmt.Printf("Sent: %s\nReceived: %s\n", sent, received)
			}

			log.Println("Aider Ping Success")
			return nil
		},
	}
	return cmd
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
