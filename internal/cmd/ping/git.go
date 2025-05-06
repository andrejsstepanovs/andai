package ping

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/internal"
	"github.com/andrejsstepanovs/andai/internal/exec"
	"github.com/andrejsstepanovs/andai/internal/redmine"
	"github.com/andrejsstepanovs/andai/internal/settings"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newGitPingCommand(deps internal.DependenciesLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git",
		Short: "Ping (open) repository",
		RunE: func(_ *cobra.Command, _ []string) error {
			d := deps()
			sett, err := d.Config.Load()
			if err != nil {
				return err
			}

			log.Println("Pinging Git Repo")
			err = pingGit(d.Model, sett.Projects)
			if err != nil {
				return err
			}
			log.Println("Git Repo Open Success")
			return nil
		},
	}
	return cmd
}

func pingGit(model *redmine.Model, projects settings.Projects) error {
	gitInstalled := exec.IsGitInstalled()
	if !gitInstalled {
		return fmt.Errorf("git is not installed")
	}

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
				git, err := exec.FindProjectGit(projectConfig, projectRepo)
				if err != nil {
					return fmt.Errorf("failed to find project git err: %v", err)
				}

				log.Printf("Project Repository Opened %s", git.GetPath())
			}
		}
	}

	return nil
}
