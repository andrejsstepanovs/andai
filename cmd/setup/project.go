package setup

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/pkg/models"
	model "github.com/andrejsstepanovs/andai/pkg/redmine"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
)

func newProjectsCommand(model *model.Model, projectsConf models.Projects) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Save (Update) projects",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Processing Projects sync")
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
			log.Println("Project repository OK")

			return nil
		},
	}
	return cmd
}
