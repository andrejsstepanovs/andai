package setup

import (
	"log"
	"time"

	"github.com/andrejsstepanovs/andai/internal"
	model "github.com/andrejsstepanovs/andai/internal/redmine"
	"github.com/andrejsstepanovs/andai/internal/settings"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newSetupAllCommand(deps internal.DependenciesLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Setup everything",
		RunE: func(_ *cobra.Command, _ []string) error {
			d := deps()
			s, err := d.Config.Load()
			if err != nil {
				return err
			}

			log.Println("Setup ALL")
			return Setup(d.Model, s.Projects, s.Workflow)
		},
	}
	return cmd
}

func Setup(model *model.Model, projectsConf settings.Projects, workflowConfig settings.Workflow) error {
	maxTries := 7
	for {
		err := setupAll(model, projectsConf, workflowConfig)
		if err == nil {
			break
		}
		log.Printf("Error: %v\n", err)
		time.Sleep(3 * time.Second)
		maxTries--
		if maxTries == 0 {
			return err
		}
	}
	return nil
}

func setupAll(model *model.Model, projectsConf settings.Projects, workflowConfig settings.Workflow) error {
	err := setupAutoIncrement(model)
	if err != nil {
		return err
	}
	log.Println("Auto increments OK")

	err = setupAdmin(model)
	if err != nil {
		return err
	}
	log.Println("Admin OK")

	err = setupSettings(model)
	if err != nil {
		return err
	}
	log.Println("Settings OK")

	err = setupToken(model)
	if err != nil {
		return err
	}
	log.Println("Token OK")

	// could ping now

	err = setupProjects(model, projectsConf)
	if err != nil {
		return err
	}
	log.Println("Projects OK")

	err = setupWorkflow(model, workflowConfig)
	if err != nil {
		return err
	}
	log.Println("Workflow OK")

	err = setupCustomFields(model)
	if err != nil {
		return err
	}
	log.Println("Workflow OK")

	return nil
}
