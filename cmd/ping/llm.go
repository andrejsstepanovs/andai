package ping

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/pkg/ai"
	"github.com/spf13/cobra"
)

func newLLMPingCommand(llm *ai.AI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "llm",
		Short: "Test LLM connection",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Ping LLM")

			response, err := llm.Simple("Answer with 1 word")
			if err != nil {
				log.Println("Failed to get response from LlmNorm")
			}

			if response.Stdout == "" {
				log.Println("LLM failed")
				return fmt.Errorf("LLM failed")
			}

			log.Println("LLM OK")

			return nil
		},
	}
	return cmd
}
