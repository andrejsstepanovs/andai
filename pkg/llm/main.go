package llm

import (
	"log"
	"time"

	"github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/teilomillet/gollm"
)

type LLM struct {
	Coder gollm.LLM
}

func NewLLM(config models.LlmModels) *LLM {
	coderCfg := config.Get("coder")

	llm := &LLM{}

	coder, err := llm.getCoder(coderCfg)
	if err != nil {
		log.Fatalf("Failed to create LLM: %v", err)
	}

	llm.Coder = coder

	return llm
}

func (l *LLM) getCoder(config models.LlmModel) (gollm.LLM, error) {
	coder, err := gollm.NewLLM(
		gollm.SetProvider(config.Provider),
		gollm.SetModel(config.Model),
		gollm.SetAPIKey(config.APIKey),
		gollm.SetMaxRetries(3),
		gollm.SetRetryDelay(time.Second*2),
		gollm.SetLogLevel(gollm.LogLevelInfo),
		//gollm.SetMaxTokens(200),
	)
	if err != nil {
		log.Fatalf("Failed to create LLM: %v", err)
		return nil, err
	}
	return coder, nil
}
