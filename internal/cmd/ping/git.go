package ping

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/internal"
	"github.com/andrejsstepanovs/andai/internal/models"
	"github.com/andrejsstepanovs/andai/internal/redmine"
	"github.com/andrejsstepanovs/andai/internal/worker"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newGitPingCommand(deps *internal.AppDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git",
		Short: "Ping (open) repository",
		RunE: func(_ *cobra.Command, _ []string) error {
			settings, err := deps.Config.Load()
			if err != nil {
				return err
			}

			fmt.Println("Pinging Git Repo")
			err = pingGit(deps.Model, settings.Projects)
			if err != nil {
				return err
			}
			fmt.Println("Git Repo Open Success")
			return nil
		},
	}
	return cmd
}

func pingGit(model *redmine.Model, projects models.Projects) error {
	allProjects, err := model.API().Projects()
	if err != nil {
		return fmt.Errorf("failed to get redmine project err: %v", err)
	}

	for _, project := range projects {
		for _, redmineProject := range allProjects {
			if redmineProject.Identifier == project.Identifier {
				log.Printf("Project %d: %s", redmineProject.Id, redmineProject.Name)
				projectRepo, err := model.DBGetRepository(redmineProject)
				if err != nil {
					return fmt.Errorf("failed to get redmine repository err: %v", err)
				}
				log.Printf("Repository %d: %s", projectRepo.ID, projectRepo.RootURL)

				projectConfig := projects.Find(project.Identifier)
				git, err := worker.FindProjectGit(projectConfig, projectRepo)
				if err != nil {
					return fmt.Errorf("failed to find project git err: %v", err)
				}

				log.Printf("Project Repository Opened %s", git.GetPath())
			}
		}
	}

	return nil
}
