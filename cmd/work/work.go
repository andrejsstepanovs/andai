package work

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/andrejsstepanovs/andai/pkg/llm"
	"github.com/andrejsstepanovs/andai/pkg/models"
	model "github.com/andrejsstepanovs/andai/pkg/redmine"
	"github.com/andrejsstepanovs/andai/pkg/worker"
	"github.com/spf13/cobra"
)

func newWorkCommand(model *model.Model, llm *llm.LLM, models models.LlmModels) *cobra.Command {
	return &cobra.Command{
		Use:   "once",
		Short: "Work with redmine",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Starting work with redmine")
			log.Printf("Models %d", len(models))

			response, err := llm.Simple("Hi!")
			if err != nil {
				log.Println("Failed to get response from LLM")
			}

			log.Println(response)

			return nil
		},
	}
}

func newNextCommand(model *model.Model, llm *llm.LLM, projects models.Projects, workflow models.Workflow) *cobra.Command {
	return &cobra.Command{
		Use:   "next",
		Short: "Work with redmine",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Searching for next issue")

			issues, err := model.APIGetWorkableIssues(workflow)
			if err != nil {
				log.Println("Failed to get workable issue")
				return err
			}

			if len(issues) == 0 {
				log.Println("No workable issues found")
				return nil
			}

			log.Printf("FOUND WORKABLE ISSUES (%d)", len(issues))
			for _, issue := range issues {
				log.Printf("Issue %d: %s", issue.Id, issue.Subject)
				project, err := model.Api().Project(issue.Project.Id)
				if err != nil {
					return fmt.Errorf("failed to get redmine project err: %v", err)
				}
				log.Printf("Project %d: %s", project.Id, project.Name)

				projectRepo, err := model.DbGetRepository(*project)
				if err != nil {
					return fmt.Errorf("failed to get redmine repository err: %v", err)
				}
				log.Printf("Repository %d: %s", projectRepo.ID, projectRepo.RootUrl)

				var projectConfig models.Project
				for _, p := range projects {
					if p.Identifier == project.Identifier {
						projectConfig = p
						log.Printf("Project %d: %s", p.Identifier, p.Name)
						break
					}
				}

				var git *worker.Git
				currentDir, err := os.Getwd()
				if err != nil {
					return err
				}

				_, mainGoPath, _, ok := runtime.Caller(0)
				if !ok {
					fmt.Println("Failed to get the current file path")
					return err
				}

				paths := []string{
					projectRepo.RootUrl,
					projectConfig.GitPath,
					filepath.Join(currentDir, projectConfig.GitPath),
					filepath.Join(currentDir, "repositories", projectConfig.GitPath),
					filepath.Join(mainGoPath, projectConfig.GitPath),
					filepath.Join(mainGoPath, "repositories", projectConfig.GitPath),
				}
				fmt.Println(paths)
				for _, path := range paths {
					log.Printf("Trying to open git repository in %q", path)
					git = worker.NewGit(path)
					err = git.Open()
					if err != nil {
						log.Printf("failed to open git err: %v", err)
						continue
					}
					break
				}

				if !git.Opened {
					log.Printf("failed to open git repository %s", projectRepo.RootUrl)
					return nil
				}

				log.Println("Project Repository Opened")

				err = git.CheckoutBranch(strconv.Itoa(issue.Id))
				if err != nil {
					log.Printf("failed to checkout branch err: %v", err)
					return nil
				}
			}

			return nil
		},
	}
}
