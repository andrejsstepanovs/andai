package setup

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/internal"
	model "github.com/andrejsstepanovs/andai/internal/redmine"
	"github.com/andrejsstepanovs/andai/internal/settings"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
)

func newProjectsCommand(deps internal.DependenciesLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Save (Update) projects",
		RunE: func(_ *cobra.Command, _ []string) error {
			d := deps()
			settings, err := d.Config.Load()
			if err != nil {
				return err
			}

			log.Println("Processing Projects sync")

			err = setupProjects(d.Model, settings.Projects)
			if err != nil {
				return err
			}

			log.Println("Project repository OK")
			return nil
		},
	}
	return cmd
}

func setupProjects(model *model.Model, projectsConf settings.Projects) error {
	for _, p := range projectsConf {
		log.Printf("Processing: %s (%s)\n", p.Name, p.Identifier)

		project := redmine.Project{
			Name:        p.Name,
			Identifier:  p.Identifier,
			Description: p.Description,
		}

		project, err := model.APISaveProject(project)
		if err != nil {
			log.Println("Redmine Project Save Fail")
			return fmt.Errorf("error redmine project save: %v", err)
		}
		log.Println("Project OK")

		err = model.APISaveWiki(project, p.Wiki)
		if err != nil {
			log.Println("Redmine Project Wiki Save Fail")
			return fmt.Errorf("error redmine project save: %v", err)
		}
		log.Println("Wiki OK")

		err = model.DBSaveGit(project, p.GitPath)
		if err != nil {
			log.Println("Redmine Git Save Fail")
			return fmt.Errorf("error redmine git save: %v", err)
		}
	}

	return nil
}
