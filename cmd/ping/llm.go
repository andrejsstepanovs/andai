package ping

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/pkg/ai"
	"github.com/andrejsstepanovs/andai/pkg/deps"
	"github.com/spf13/cobra"
)

func newLLMPingCommand(deps *deps.AppDependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "llm",
		Short: "Test LLM connection",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Ping LLM")
			err := pingLLM(deps.LlmNorm)
			if err != nil {
				return err
			}
			log.Println("LLM OK")
			return nil
		},
	}
	return cmd
}

func pingLLM(llm *ai.AI) error {
	resp, err := llm.Simple("Answer with 1 word")
	if err != nil {
		log.Println("Failed to get response from LlmNorm")
	}

	if resp.Stdout == "" {
		log.Println("LLM failed")
		return fmt.Errorf("LLM failed")
	}

	log.Println("LLM response:", resp.Stdout)

	return nil
}
