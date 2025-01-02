package setup

import (
	"fmt"

	model "github.com/andrejsstepanovs/andai/pkg/redmine"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newProjectCommand(model *model.Model) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Save (Update) project",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println("Processing Redmine project sync")
			project := redmine.Project{
				Name:        viper.GetString("project.name"),
				Identifier:  viper.GetString("project.id"),
				Description: viper.GetString("project.description"),
			}

			err, project := model.SaveProject(project)
			if err != nil {
				fmt.Println("Redmine Project Save Fail")
				return fmt.Errorf("error redmine project save: %v", err)
			}
			fmt.Println("Project OK")

			err = model.DbSaveWiki(project, viper.GetString("project.wiki"))
			if err != nil {
				fmt.Println("Redmine Project Wiki Save Fail")
				return fmt.Errorf("error redmine project save: %v", err)
			}
			fmt.Println("Wiki OK")

			err = model.DbSaveGit(project, viper.GetString("project.git_path"))
			if err != nil {
				fmt.Println("Redmine Git Save Fail")
				return fmt.Errorf("error redmine git save: %v", err)
			}
			fmt.Println("Project repository OK")

			err = model.DbSaveProjectTrackers(project)
			if err != nil {
				fmt.Println("Redmine Project Trackers Save Fail")
				return fmt.Errorf("error redmine project trackers save: %v", err)
			}

			return nil
		},
	}
	return cmd
}
