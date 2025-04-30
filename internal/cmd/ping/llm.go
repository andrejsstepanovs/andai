package ping

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/internal"
	"github.com/andrejsstepanovs/andai/internal/ai"
	"github.com/andrejsstepanovs/andai/internal/settings"
	"github.com/spf13/cobra"
)

func newLLMPingCommand(deps *internal.AppDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "llm",
		Short: "Test LLM connection",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Ping LLM")
			err := pingLLM(deps.LlmPool)
			if err != nil {
				return err
			}
			log.Println("LLM OK")
			return nil
		},
	}
	return cmd
}

func pingLLM(llmPool *settings.LlmModels) error {
	for _, m := range *llmPool {
		llm, err := ai.NewAI(m)
		if err != nil {
			return err
		}

		resp, err := llm.Simple("Answer with 1 word: 'Yes'")
		if err != nil {
			log.Println("Failed to get response from LlmNorm")
			return fmt.Errorf("failed to get response from LLM %q: %v", m.Name, err)
		}

		if resp.Stdout == "" {
			return fmt.Errorf("LLM %q failed", m.Name)
		}

		log.Printf("LLM %q response: %s", m.Name, resp.Stdout)
	}

	return nil
}
