package ping

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/pkg/deps"
	"github.com/spf13/cobra"
)

func newPingAllCommand(deps *deps.AppDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Ping (open) Git repository",
		RunE: func(_ *cobra.Command, _ []string) error {
			settings, err := deps.Config.Load()
			if err != nil {
				return err
			}
			redmine := deps.Model
			llm := deps.LlmNorm

			err = pingGit(redmine, settings.Projects)
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

			err = pingLLM(llm)
			if err != nil {
				return err
			}
			log.Println("LLM OK")

			err = pingAider(redmine, settings.Projects, settings.Aider)
			if err != nil {
				return err
			}
			log.Println("Aider OK")
			return nil
		},
	}
	return cmd
}
