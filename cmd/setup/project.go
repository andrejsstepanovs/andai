package setup

import (
	"fmt"

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
			fmt.Println("Processing Projects sync")
			for _, p := range projectsConf {
				fmt.Printf("Processing: %s (%s)\n", p.Name, p.Identifier)

				project := redmine.Project{
					Name:        p.Name,
					Identifier:  p.Identifier,
					Description: p.Description,
				}

				err, project := model.ApiSaveProject(project)
				if err != nil {
					fmt.Println("Redmine Project Save Fail")
					return fmt.Errorf("error redmine project save: %v", err)
				}
				fmt.Println("Project OK")

				err = model.ApiSaveWiki(project, p.Wiki)
				if err != nil {
					fmt.Println("Redmine Project Wiki Save Fail")
					return fmt.Errorf("error redmine project save: %v", err)
				}
				fmt.Println("Wiki OK")

				err = model.DbSaveGit(project, p.GitPath)
				if err != nil {
					fmt.Println("Redmine Git Save Fail")
					return fmt.Errorf("error redmine git save: %v", err)
				}
			}
			fmt.Println("Project repository OK")

			return nil
		},
	}
	return cmd
}
