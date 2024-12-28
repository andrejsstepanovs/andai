package redmine

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewProjectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Create or update project",
		RunE:  runRedmineSaveProject,
	}
	return cmd
}

func runRedmineSaveProject(cmd *cobra.Command, args []string) error {
	fmt.Println("Processing Redmine project save", len(args))

	project := redmine.Project{
		Name:        viper.GetString("project.name"),
		Identifier:  viper.GetString("project.id"),
		Description: viper.GetString("project.description"),
		//CustomFields: nil,
	}

	client := redmine.NewClient(viper.GetString("redmine.url"), viper.GetString("redmine.api_key"))

	err, project := saveProject(client, project)
	if err != nil {
		fmt.Println("Redmine Project Save Fail")
		return fmt.Errorf("error redmine project save: %v", err)
	}

	fmt.Printf("ID: %d, Name: %s\n", project.Id, project.Name)

	return nil
}

func saveProject(client *redmine.Client, project redmine.Project) (error, redmine.Project) {
	projects, err := client.Projects()
	if err != nil {
		fmt.Println("Redmine Projects Fail")
		return fmt.Errorf("error redmine projects: %v", err), redmine.Project{}
	}

	for _, p := range projects {
		fmt.Println(fmt.Sprintf("ID: %d, Name: %s", p.Id, p.Name))
		if p.Identifier == viper.GetString("project.id") {
			fmt.Printf("Project already exists: %s\n", p.Name)
			project.Id = p.Id
			err = client.UpdateProject(project)
			if err != nil && err.Error() != "EOF" {
				fmt.Println("Redmine Update Project Failed")
				return fmt.Errorf("error redmine update project: %v", err.Error()), redmine.Project{}
			}
			fmt.Printf("Project updated: %s\n", project.Name)
			return nil, project
		}
	}

	response, err := client.CreateProject(project)
	if err != nil {
		fmt.Println("Redmine Create Project Failed")
		return fmt.Errorf("error redmine create project: '%s'", err.Error()), redmine.Project{}
	}

	return nil, *response
}
