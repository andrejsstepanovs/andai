package setup

import (
	"fmt"

	model "github.com/andrejsstepanovs/andai/pkg/redmine"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newProjectCommand(model *model.Model) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Save (Update) project",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Processing Redmine project save", len(args))
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

			err = model.SaveWiki(project, viper.GetString("project.wiki"))
			if err != nil {
				fmt.Println("Redmine Project Wiki Save Fail")
				return fmt.Errorf("error redmine project save: %v", err)
			}
			fmt.Println("Wiki saved")

			err = model.SaveGit(project, viper.GetString("project.git_path"))
			if err != nil {
				fmt.Println("Redmine Git Save Fail")
				return fmt.Errorf("error redmine git save: %v", err)
			}
			fmt.Println("Git saved")

			return nil
		},
	}
	return cmd
}
