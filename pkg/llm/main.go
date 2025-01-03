package llm

import (
	"log"
	"time"

	"github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/teilomillet/gollm"
)

type LLM struct {
	clientCoder gollm.LLM
}

func NewLLM(config models.Workflow) *LLM {
	coder, err := gollm.NewLLM(
		gollm.SetProvider("openai"),
		gollm.SetModel("gpt-4o-mini"),
		//gollm.SetAPIKey(apiKey),
		gollm.SetMaxTokens(200),
		gollm.SetMaxRetries(3),
		gollm.SetRetryDelay(time.Second*2),
		gollm.SetLogLevel(gollm.LogLevelInfo),
	)
	if err != nil {
		log.Fatalf("Failed to create LLM: %v", err)
	}

	return &LLM{
		clientCoder: coder,
	}
}
