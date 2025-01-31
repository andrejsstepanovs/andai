package ai

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/andrejsstepanovs/andai/pkg/exec"
	"github.com/teilomillet/gollm"
	"github.com/teilomillet/gollm/llm"
	"github.com/teilomillet/gollm/providers"
	"github.com/teilomillet/gollm/utils"
)

func NewAI(provider, model, apiKey, temperature string) (*AI, error) {
	cfg, err := gollm.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var temp float64
	if temperature != "" {
		// deepseek R1 model does not support temperature
		if provider != "deepseek" && !strings.Contains(strings.ToLower(model), "r1") {
			feetFloat, err := strconv.ParseFloat(temperature, 64)
			if err != nil {
				log.Fatalf("Failed to parse temperature: %v", err)
			}
			temp = feetFloat
		}
	}

	registry := providers.NewProviderRegistry()
	for providerName, endpoint := range map[string]string{
		"zalando":    "https://zllm.data.zalan.do/v1/chat/completions",
		"deepseek":   "https://api.deepseek.com/chat/completions",
		"openrouter": "https://openrouter.ai/api/v1/chat/completions",
		"mistral":    "https://api.mistral.ai/v1/chat/completions",
	} {
		registry.Register(providerName, func(apiKey, model string, extraHeaders map[string]string) providers.Provider {
			return NewCustomOpenAIProvider(
				providerName,
				endpoint,
				apiKey,
				model,
				&temp,
				extraHeaders,
			)
		})
	}

	cfg.Provider = provider
	cfg.APIKeys = map[string]string{provider: apiKey}
	cfg.Model = model
	cfg.MaxTokens = 4096
	cfg.MaxRetries = 30
	cfg.Timeout = time.Minute * 2
	cfg.RetryDelay = time.Second * 5
	cfg.LogLevel = gollm.LogLevelInfo
	conn, err := llm.NewLLM(cfg, utils.NewLogger(cfg.LogLevel), registry)

	if err != nil {
		log.Fatalf("Failed to create LlmNorm: %v", err)
		return nil, err
	}

	return &AI{
		client:   conn,
		provider: provider,
		model:    model,
	}, nil
}

type AI struct {
	client   llm.LLM
	provider string
	model    string
}

func (a *AI) Simple(prompt string) (exec.Output, error) {
	resp, err := a.client.Generate(context.Background(), &llm.Prompt{Input: prompt})
	if err != nil {
		return exec.Output{}, err
	}
	out := exec.Output{
		Stdout: resp,
	}
	return out, nil
}

func (a *AI) Generate(ctx context.Context, prompt *llm.Prompt, opts ...llm.GenerateOption) (exec.Output, error) {
	resp, err := a.client.Generate(ctx, prompt, opts...)
	if err != nil {
		return exec.Output{}, err
	}

	log.Printf("%s %s Response OK", a.provider, a.model)
	//log.Println(resp) // debug

	out := exec.Output{
		Stdout: resp,
	}
	return out, nil
}

// GenerateJSON generates JSON from the given prompt
// Second error is json validation error
// Last error is general LLM call error
func (a *AI) GenerateJSON(ctx context.Context, prompt *llm.Prompt) (exec.Output, error, error) {
	templateResponse, err := a.Generate(ctx, prompt, gollm.WithJSONSchemaValidation())
	if err != nil {
		return exec.Output{}, nil, err
	}

	var picked []string
	responseJson := templateResponse.Stdout
	if responseJson != "[]" {
		responseJson = gollm.CleanResponse(responseJson)
		err = json.Unmarshal([]byte(responseJson), &picked)
		if err != nil {
			log.Println(templateResponse.Stdout)
			return exec.Output{}, err, nil
		}
	}

	return exec.Output{Stdout: responseJson}, nil, nil
}
