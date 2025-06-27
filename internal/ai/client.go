package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/andrejsstepanovs/andai/internal/exec"
	"github.com/andrejsstepanovs/andai/internal/settings"
	"github.com/teilomillet/gollm"
	"github.com/teilomillet/gollm/config"
	"github.com/teilomillet/gollm/llm"
	"github.com/teilomillet/gollm/providers"
	"github.com/teilomillet/gollm/utils"
)

var ErrTooManyTokens = fmt.Errorf("prompt exceeds max tokens limit")

func NewAI(config settings.LlmModel) (*AI, error) {
	cfg, err := gollm.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	modelsWithNoTemp := []string{"r1", "o1"}
	var temp float64
	useTemperature := true
	for _, model := range modelsWithNoTemp {
		if strings.Contains(strings.ToLower(config.Model), model) {
			useTemperature = false
		}
	}
	if useTemperature {
		temp = config.Temperature
	}

	maxTokens := 128000
	if config.MaxTokens > 0 {
		maxTokens = config.MaxTokens
	}
	maxRetries := 5
	if config.MaxRetries > 0 {
		maxRetries = config.MaxRetries
	}

	cfg.Provider = config.Provider
	cfg.APIKeys = map[string]string{config.Provider: config.APIKey.String()}
	cfg.Model = config.Model
	cfg.MaxTokens = maxTokens
	cfg.MaxRetries = maxRetries
	cfg.Timeout = time.Minute * 10
	cfg.RetryDelay = time.Second * 5
	cfg.LogLevel = gollm.LogLevelInfo

	customProviders := map[string]string{
		"deepseek":   "https://api.deepseek.com/chat/completions",
		"openrouter": "https://openrouter.ai/api/v1/chat/completions",
		"mistral":    "https://api.mistral.ai/v1/chat/completions",
		"google":     "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
		"groq":       "https://api.groq.com/openai/v1/chat/completions",
		"openai":     "https://api.openai.com/v1/chat/completions",
		"litellm":    "http://localhost:4000/v1/chat/completions",
		"custom":     config.BaseURL,
	}

	endpoint, ok := customProviders[config.Provider]
	var conn llm.LLM
	if ok {
		//log.Printf("Using custom OpenAI provider %q", config.Provider)
		registry := providers.NewProviderRegistry()
		registry.Register(config.Provider, func(apiKey, model string, extraHeaders map[string]string) providers.Provider {
			return NewCustomOpenAIProvider(
				config.Provider,
				endpoint,
				apiKey,
				model,
				&temp,
				extraHeaders,
			)
		})
		conn, err = llm.NewLLM(cfg, utils.NewLogger(cfg.LogLevel), registry)
		if err != nil {
			return nil, fmt.Errorf("failed to create custom %q LLM err: %v", config.Provider, err)
		}
	} else {
		//log.Printf("Using provider %q", config.Provider)
		opts := []gollm.ConfigOption{
			gollm.SetProvider(cfg.Provider),
			gollm.SetModel(cfg.Model),
			gollm.SetAPIKey(config.APIKey.String()),
			gollm.SetMaxTokens(cfg.MaxTokens),
			gollm.SetMaxRetries(cfg.MaxRetries),
			gollm.SetRetryDelay(cfg.RetryDelay),
			gollm.SetLogLevel(cfg.LogLevel),
			gollm.SetTimeout(cfg.Timeout),
			gollm.SetTemperature(temp),
		}
		if config.BaseURL != "" {
			opts = append(opts, gollm.SetOllamaEndpoint(config.BaseURL))
		}
		conn, err = gollm.NewLLM(opts...)
		if config.BaseURL != "" {
			conn.SetEndpoint(config.BaseURL)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create %q LLM err: %v", config.Provider, err)
		}
	}

	return &AI{
		client:   conn,
		provider: config.Provider,
		model:    config.Model,
		config:   cfg,
	}, nil
}

type AI struct {
	client   llm.LLM
	provider string
	model    string
	config   *config.Config
}

func (a *AI) Multi(question string, prompts []map[string]string) (exec.Output, error) {
	messages := make([]gollm.PromptMessage, 0)
	for _, conversation := range prompts {
		for role, message := range conversation {
			messages = append(messages, gollm.PromptMessage{Role: role, Content: message})
		}
	}
	prompt := &gollm.Prompt{
		Input:    question,
		Messages: messages,
	}

	//log.Println(prompt) // debug

	return a.Generate(context.Background(), prompt)
}

func (a *AI) Simple(prompt string) (exec.Output, error) {
	resp, err := a.client.Generate(context.Background(), &llm.Prompt{Input: prompt})
	if err != nil {
		return exec.Output{}, err
	}
	resp = RemoveThinkingContent(resp)
	out := exec.Output{
		Stdout: resp,
	}
	return out, nil
}

func (a *AI) Generate(ctx context.Context, prompt *llm.Prompt, opts ...llm.GenerateOption) (exec.Output, error) {
	if a.config != nil && a.estimateTokens(prompt.String()) > a.config.MaxTokens {
		return exec.Output{}, ErrTooManyTokens
	}

	resp, err := a.client.Generate(ctx, prompt, opts...)
	if err != nil {
		return exec.Output{}, err
	}

	log.Printf("%s %s Response OK", a.provider, a.model)
	resp = RemoveThinkingContent(resp)
	//log.Println(resp) // debug

	out := exec.Output{
		Stdout: resp,
	}
	return out, nil
}

// GenerateJSON generates JSON from the given prompt
// Second error is json validation error
// Last error is general LLM call error
func (a *AI) GenerateJSON(ctx context.Context, prompt *llm.Prompt, element any) (exec.Output, error, error) {
	templateResponse, err := a.Generate(ctx, prompt, gollm.WithJSONSchemaValidation())
	if err != nil {
		return exec.Output{}, nil, err
	}

	responseJSON := templateResponse.Stdout
	if responseJSON != "[]" {
		responseJSON = gollm.CleanResponse(responseJSON)
		err = json.Unmarshal([]byte(responseJSON), &element)
		if err != nil {
			log.Println(templateResponse.Stdout)
			return exec.Output{}, err, nil
		}
		templateResponse.Stdout = responseJSON
	}

	return exec.Output{Stdout: responseJSON}, nil, nil
}

func (a *AI) estimateTokens(text string) int {
	// Character-based estimation
	charEstimate := int(math.Ceil(float64(len(text)) / 4))

	// Word-based estimation
	words := strings.Fields(text)
	wordEstimate := int(math.Ceil(float64(len(words)) * 1.33))

	return max(charEstimate, wordEstimate)
}
