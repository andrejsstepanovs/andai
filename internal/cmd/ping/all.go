package ping

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/internal"
	"github.com/spf13/cobra"
)

func newPingAllCommand(deps internal.DependenciesLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Ping (open) Git repository",
		RunE: func(_ *cobra.Command, _ []string) error {
			d := deps()
			sett, err := d.Config.Load()
			if err != nil {
				return err
			}
			redmine := d.Model

			err = pingGit(redmine, sett.Projects)
			if err != nil {
				return err
			}
			fmt.Println("Git Repo OK")

			err = pingDB(redmine)
			if err != nil {
				return err
			}
			fmt.Println("DB OK")

			err = pingAPI(redmine)
			if err != nil {
				return err
			}
			fmt.Println("API OK")

			err = pingLLM(d.LlmPool)
			if err != nil {
				return err
			}
			log.Println("LLM OK")

			err = pingAider(redmine, sett.Projects, sett.Aider)
			if err != nil {
				return err
			}
			log.Println("Aider OK")
			return nil
		},
	}
	return cmd
}
