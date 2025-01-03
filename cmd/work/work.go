package work

import (
	"log"

	"github.com/andrejsstepanovs/andai/pkg/llm"
	"github.com/andrejsstepanovs/andai/pkg/models"
	model "github.com/andrejsstepanovs/andai/pkg/redmine"
	"github.com/spf13/cobra"
)

func newWorkCommand(model *model.Model, llm *llm.LLM, models models.LlmModels) *cobra.Command {
	return &cobra.Command{
		Use:   "once",
		Short: "Work with redmine",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Starting work with redmine")
			log.Printf("Models %d", len(models))

			response, err := llm.Simple("Hi!")
			if err != nil {
				log.Println("Failed to get response from LLM")
			}

			log.Println(response)

			return nil
		},
	}
}

func newNextCommand(model *model.Model, llm *llm.LLM, workflow models.Workflow) *cobra.Command {
	return &cobra.Command{
		Use:   "next",
		Short: "Work with redmine",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Searching for next issue")

			_, err := model.APIGetWorkableIssue(workflow.Priorities)
			if err != nil {
				log.Println("Failed to get workable issue")
				return err
			}

			return nil
		},
	}
}
